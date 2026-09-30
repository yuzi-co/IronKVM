package vpn

import (
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"

	log "github.com/sirupsen/logrus"
)

// Steps are what Connect, Disconnect and ConnectAtBoot need from an add-on's
// CLI.
type Steps struct {
	Start func() error
	Stop  func() error
	Up    func() error
	Down  func() error
	// State is the daemon's state as the page shows it. An error means the
	// daemon did not answer.
	State func() (proto.VpnState, error)
}

// How long Connect waits for a daemon it has just started to say where it
// stands. Variables for the tests.
var (
	settleWait = 15 * time.Second
	settlePoll = time.Second
)

// Connect is the page's one switch turned on: it starts the daemon when it is
// not running, then brings the network up. A daemon that is not logged in is
// only started, and the page shows its login; one that is already connected,
// or connecting on its own, is left alone. The caller holds the VPN lock after
// the exclusivity check, as Guard does.
func Connect(d addon.Daemon, s Steps) error {
	if !addon.Running(d) {
		if err := s.Start(); err != nil {
			return err
		}
	}
	if settle(s.State, settleWait, settlePoll) != proto.VpnStopped {
		return nil
	}
	if err := s.Up(); err != nil {
		// NetBird reconnects on its own after a start, and refuses an up
		// that meets one in progress. Connected is what was asked for.
		if st, e := s.State(); e == nil && st == proto.VpnRunning {
			return nil
		}
		return err
	}
	return nil
}

// Disconnect is the switch turned off: the network goes down, then the
// daemon stops, which frees its memory. A down that fails does not keep the
// daemon running: stopping it disconnects as well.
func Disconnect(d addon.Daemon, s Steps) error {
	if addon.Running(d) {
		if err := s.Down(); err != nil {
			log.Warnf("%s down before stop failed: %s", d.Name, err)
		}
	}
	return s.Stop()
}

// settle asks for the state until the daemon answers with one other than not
// running (Tailscale's Starting and NoState), and returns the last answer.
func settle(state func() (proto.VpnState, error), wait, poll time.Duration) proto.VpnState {
	deadline := time.Now().Add(wait)
	last := proto.VpnNotRunning
	for {
		if st, err := state(); err == nil {
			last = st
			if st != proto.VpnNotRunning {
				return st
			}
		}
		if !time.Now().Add(poll).Before(deadline) {
			return last
		}
		time.Sleep(poll)
	}
}

// BootWindow is how soon after the board boots a server start counts as part
// of that boot. A server restarted later, by an update or the supervisor,
// leaves the VPN as the operator left it.
var BootWindow = 10 * time.Minute

// How long ConnectAtBoot waits for the boot script's daemon, and how often it
// looks. Variables for the tests.
var (
	bootWait = 3 * time.Minute
	bootPoll = 5 * time.Second
)

// ConnectAtBoot makes connect at boot mean connected. The boot script only
// starts the daemon, and a daemon that was last disconnected with the page's
// switch remembers that: Tailscale keeps its down in its state file. So once
// the daemon answers after a boot, a logged-in daemon that is not connected is
// brought up. It asks twice before it acts, as NetBird is idle for a moment
// after its start before it reconnects on its own. It returns when it is done
// and is meant to run in its own goroutine.
func ConnectAtBoot(d addon.Daemon, s Steps) {
	if !addon.BootEnabled(d) {
		return
	}
	if up := SystemUptimeSec(); up == 0 || time.Duration(up)*time.Second > BootWindow {
		return
	}

	deadline := time.Now().Add(bootWait)
	stopped := 0
	for {
		if addon.Running(d) {
			st, err := s.State()
			switch {
			case err != nil:
				stopped = 0
			case st == proto.VpnStopped:
				stopped++
				if stopped >= 2 {
					if err := addon.Exclusive(d.Name, s.Up); err != nil {
						log.Errorf("%s: connect at boot failed: %s", d.Name, err)
						return
					}
					log.Infof("%s: connected at boot", d.Name)
					return
				}
			case st == proto.VpnRunning, st == proto.VpnNotLogin:
				return
			default:
				stopped = 0
			}
		}
		if !time.Now().Add(bootPoll).Before(deadline) {
			log.Warnf("%s: not connected at boot: the daemon did not answer in %s", d.Name, bootWait)
			return
		}
		time.Sleep(bootPoll)
	}
}
