package network

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// A trial ended two ways before this file: somebody pressed the button, or the
// deadline passed and the board put the old settings back. Neither of them is
// the question the trial actually asks.
//
// The question is "can a client still reach the board". A person who opens the
// new address and signs in has answered it. Asking that person to then find
// the settings page and press a further button adds a step that can only be
// missed, and it is missed most easily in the case the trial exists for: the
// address moved, the page the operator was reading is dead, the session cookie
// belongs to a host that no longer answers, and there are three minutes to
// notice all of that. A change that worked gets rolled back because nobody
// pressed anything, and the operator watches the board return to an address
// they had deliberately left.
//
// So a signed-in request that arrives over the address the trial applied
// confirms the trial. Four things have to hold before one counts:
//
//   - A session authenticated it, not an api key. A key belongs to a script,
//     and a script that can reach the board says nothing about whether the
//     operator can. A key holder that wants to keep a change can still call
//     the confirm route, which is an explicit act rather than a side effect of
//     polling.
//   - That session belongs to an administrator. Every ethernet route is behind
//     RequireRole(RoleAdmin), so ratifying a change must not be reachable
//     through a session that could not have made it. An operator whose browser
//     reaches the board keeps nothing.
//   - The request arrived on the address the trial applied. That is the whole
//     proof, and it is what excludes a request that came in over Wi-Fi, over
//     the USB network gadget, or over a connection older than the change.
//   - The peer is not loopback. The on-device runtime calls back over loopback
//     with credentials that are valid, and the supervisor probes 127.0.0.1.
//     Neither of them has crossed the network that the change could break.
//   - The interface already carries the change. Until it does, a request is a
//     request over the address the board is leaving.
//
// What this does not prove is that the operator can reach the board from
// everywhere they need to. A wrong gateway is invisible to a client on the
// same subnet, and it was invisible to the button as well: somebody who could
// press it could reach the board too. The trial tests reachability from
// whoever answers, and that has not changed.

// trialPending is read on every signed-in request, so it is an atomic and not
// the mutex. While no trial runs, which is nearly always, the check costs this
// load and the one inside adoptOnce and touches nothing else.
//
// It does not make the whole path lock free. While a trial does run, every
// signed-in request takes trialMutex, and a request that confirms holds it
// across the /boot write and the sync. Measured on the device on 2026-09-07,
// that write and sync together take under a millisecond, so the others queue
// for no time worth naming. Do not read the flag as a promise about the
// confirming request itself.
var trialPending atomic.Bool

// adoptOnce is the single look at /run for a trial that a previous process
// started. After it, a trial this process did not start either raised the flag
// or does not exist, and the hot path is two atomic loads.
//
// It is a pointer so a test can hand the process a fresh one.
var adoptOnce = new(sync.Once)

func adoptPendingTrial() {
	if _, ok := readTrialState(); ok {
		trialPending.Store(true)
	}
}

// NoteSignedInAdminRequest records that an administrator's client reached the
// board. The authentication middleware calls it, so it runs on every request
// an administrator's session authenticates and does nothing at all while no
// trial is running.
func NoteSignedInAdminRequest(c *gin.Context) {
	if c == nil || c.Request == nil {
		return
	}

	noteReachable(localAddressOf(c.Request), c.Request.RemoteAddr)
}

func noteReachable(local string, remote string) {
	adoptOnce.Do(adoptPendingTrial)
	if !trialPending.Load() {
		return
	}

	if local == "" || isLoopbackAddress(local) || isLoopbackAddress(remote) {
		return
	}

	trialMutex.Lock()
	defer trialMutex.Unlock()

	token, mode, ok := trialReachedLocked(local)
	if !ok {
		return
	}

	if err := confirmTrialLocked(token); err != nil {
		// The deadline keeps running. A board that kept an address it could
		// not save would work until the next reboot and lose it then.
		log.Errorf("a signed-in client reached the board at %s but the trial could not be saved: %s", local, err)
		return
	}

	log.Infof("ethernet trial %s confirmed: a signed-in client reached the board at %s in %s mode",
		token, local, mode)
}

// trialReachedLocked answers whether local is the address the pending trial
// applied, and names the trial if it is.
func trialReachedLocked(local string) (token string, mode string, ok bool) {
	var applied string

	if pendingTrial != nil {
		if !pendingTrial.applied {
			return "", "", false
		}
		token, mode, applied = pendingTrial.token, pendingTrial.mode, pendingTrial.config.Address
	} else {
		// A trial that outlived the process that started it. The detached
		// revert is still counting, and the person who just signed in is the
		// one who can end it.
		state, found := readTrialState()
		if !found {
			trialPending.Store(false)
			return "", "", false
		}
		if !state.Applied {
			// The process that started this trial did not live to apply it.
			// The interface still carries the address the board was leaving,
			// so a request that arrives on it proves nothing.
			return "", "", false
		}
		token, mode, applied = state.Token, state.Mode, state.Address
	}

	if mode == ethModeDHCP {
		// The lease picks the address, so the trial does not know it. Any
		// address the interface carries is one the lease put there: the apply
		// flushes the interface first, so the address the board is leaving is
		// gone before a request can arrive on it.
		return token, mode, ethernetCarries(local)
	}

	return token, mode, applied != "" && applied == local
}

// markTrialApplied says the interface now carries the trial. It records the
// fact in both places the trial lives, because a trial adopted from /run by a
// later process has no other way to tell an applied change from one that was
// only recorded.
func markTrialApplied(token string) {
	trialMutex.Lock()
	defer trialMutex.Unlock()

	if pendingTrial == nil || pendingTrial.token != token {
		return
	}
	pendingTrial.applied = true

	state, ok := readTrialState()
	if !ok || state.Token != token {
		return
	}

	state.Applied = true
	if err := writeTrialState(state); err != nil {
		// The in-process trial is still marked, so this server confirms on
		// reach as usual. Only a server that restarts from here loses the
		// fact, and it then waits out the deadline instead of confirming.
		log.Errorf("failed to record that the ethernet trial was applied: %s", err)
	}
}

// localAddressOf reports the address the request arrived on. net/http puts the
// accepting socket's local address in the connection context, which is the
// only place a handler can read it: the Host header is what the client typed
// and a client can type anything.
func localAddressOf(r *http.Request) string {
	address, ok := r.Context().Value(http.LocalAddrContextKey).(net.Addr)
	if !ok || address == nil {
		return ""
	}

	return hostOf(address.String())
}

// hostOf takes the address out of a "host:port" pair and normalises it. A
// dual-stack listener writes an IPv4 peer as ::ffff:10.0.0.222, and the
// interface reports the same address as 10.0.0.222.
func hostOf(value string) string {
	host, _, err := net.SplitHostPort(value)
	if err != nil {
		host = value
	}

	ip := net.ParseIP(strings.Trim(host, "[]"))
	if ip == nil {
		return ""
	}
	if v4 := ip.To4(); v4 != nil {
		return v4.String()
	}

	return ip.String()
}

func isLoopbackAddress(value string) bool {
	ip := net.ParseIP(hostOf(value))

	return ip == nil || ip.IsLoopback()
}

// ethernetCarries reports whether the wired interface holds this address. It
// is a var so a test does not depend on the addresses of the machine it runs
// on.
var ethernetCarries = func(address string) bool {
	iface, err := net.InterfaceByName(ethInterface)
	if err != nil {
		return false
	}

	addresses, err := iface.Addrs()
	if err != nil {
		return false
	}

	for _, entry := range addresses {
		ipNet, ok := entry.(*net.IPNet)
		if !ok {
			continue
		}
		if hostOf(ipNet.IP.String()) == address {
			return true
		}
	}

	return false
}
