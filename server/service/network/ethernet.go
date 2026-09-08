package network

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

const (
	// defaultTrialSeconds is how long the board waits for a confirmation.
	//
	// The session cookie belongs to the old address, so the browser cannot
	// carry it to the new one. The window has to cover a person noticing that
	// the page went away, opening the new address, logging in again, and
	// pressing the button. Three minutes is generous for that, and the cost of
	// being generous is only paid when the change was wrong: it is how long
	// the board stays unreachable before it puts itself back.
	defaultTrialSeconds = 180
	minTrialSeconds     = 30
	maxTrialSeconds     = 600

	// applyDelay lets the response reach the browser before the interface goes
	// down. The address change breaks every open connection, this one
	// included.
	applyDelay = 500 * time.Millisecond

	udhcpcPidFile = "/run/udhcpc.eth0.pid"
	ethInitScript = "/etc/init.d/S30eth"
)

// trial is an applied change that nobody has confirmed yet. Only one exists at
// a time: a second request replaces the first, because the interface can only
// carry one configuration.
type trial struct {
	token    string
	mode     string
	config   ethernetConfig
	deadline time.Time
	timer    *time.Timer
	// applied says the interface carries this trial. A request that arrives
	// before it does came in over the address the board is leaving, and
	// ethernet_reach.go must not read one as proof that the change works.
	applied bool
}

var (
	trialMutex   sync.Mutex
	pendingTrial *trial
)

func (s *Service) GetEthernet(c *gin.Context) {
	var rsp proto.Response

	saved, isStatic := readEthernetConfig()

	mode := ethModeDHCP
	if isStatic {
		mode = ethModeStatic
	}

	data := &proto.GetEthernetRsp{
		Mode:    mode,
		Address: saved.Address,
		Prefix:  saved.Prefix,
		Gateway: saved.Gateway,
		Live:    liveEthernet(),
		Trial:   describeTrial(),
	}

	// A board that has always been on DHCP has no saved static settings, so
	// offer the current lease as the starting point of the form.
	if data.Address == "" {
		data.Address = data.Live.Address
		data.Prefix = data.Live.Prefix
		data.Gateway = data.Live.Gateway
	}

	rsp.OkRspWithData(c, data)
	log.Debugf("get ethernet config: mode=%s address=%s/%d gateway=%s live=%+v",
		data.Mode, data.Address, data.Prefix, data.Gateway, data.Live)
}

func (s *Service) SetEthernet(c *gin.Context) {
	var req proto.SetEthernetReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	config := ethernetConfig{
		Address: strings.TrimSpace(req.Address),
		Prefix:  req.Prefix,
		Gateway: strings.TrimSpace(req.Gateway),
	}

	if req.Mode == ethModeStatic {
		if err := validateEthernetConfig(config); err != nil {
			rsp.ErrRsp(c, -2, err.Error())
			return
		}
	} else {
		config = ethernetConfig{}
	}

	// Nothing to do is worth detecting: the apply below drops every open
	// connection, and paying that for a request that changes nothing would
	// look like a fault.
	if isAlreadyPersisted(req.Mode, config) && describeTrial() == nil {
		rsp.OkRspWithData(c, &proto.SetEthernetRsp{
			Address: config.Address,
			Prefix:  config.Prefix,
		})
		log.Debugf("ethernet config unchanged, nothing applied")
		return
	}

	seconds := clampTrialSeconds(req.TrialSeconds)
	token, err := newTrialToken()
	if err != nil {
		rsp.ErrRsp(c, -3, "failed to start the trial")
		return
	}

	startTrial(token, req.Mode, config, seconds)

	live := liveEthernet()
	data := &proto.SetEthernetRsp{
		Token:        token,
		TrialSeconds: seconds,
		Address:      config.Address,
		Prefix:       config.Prefix,
	}
	if req.Mode == ethModeDHCP {
		// The lease decides the address, so the caller cannot be told where to
		// look. Report where the board is now, which is where a DHCP server
		// that remembers the board will most likely put it again.
		data.Address = live.Address
		data.Prefix = live.Prefix
	}

	rsp.OkRspWithData(c, data)

	go applyAfterResponse(token, req.Mode, config)

	log.Infof("ethernet trial started: mode=%s address=%s/%d gateway=%s seconds=%d",
		req.Mode, config.Address, config.Prefix, config.Gateway, seconds)
}

func (s *Service) ConfirmEthernet(c *gin.Context) {
	var req proto.ConfirmEthernetReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	if err := confirmTrial(req.Token); err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	rsp.OkRsp(c)
	log.Infof("ethernet trial confirmed and saved")
}

// applyAfterResponse waits for the response to leave, then changes the
// interface. It runs in its own goroutine because the request that started it
// is one of the connections the change breaks.
func applyAfterResponse(token string, mode string, config ethernetConfig) {
	time.Sleep(applyDelay)

	var err error
	if mode == ethModeStatic {
		err = applyStatic(config)
	} else {
		err = applyDHCP()
	}

	if err != nil {
		log.Errorf("failed to apply the ethernet trial: %s", err)
		// The interface is now in whatever state the failure left it, which
		// may be no address at all. Put the saved configuration back rather
		// than wait out the trial.
		revertNow()
		return
	}

	markTrialApplied(token)
}

func clampTrialSeconds(seconds int) int {
	if seconds == 0 {
		return defaultTrialSeconds
	}
	if seconds < minTrialSeconds {
		return minTrialSeconds
	}
	if seconds > maxTrialSeconds {
		return maxTrialSeconds
	}

	return seconds
}

func isAlreadyPersisted(mode string, config ethernetConfig) bool {
	saved, isStatic := readEthernetConfig()

	if mode == ethModeDHCP {
		return !isStatic
	}

	return isStatic && saved == config
}

func newTrialToken() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return hex.EncodeToString(buffer), nil
}

func startTrial(token string, mode string, config ethernetConfig, seconds int) {
	trialMutex.Lock()
	defer trialMutex.Unlock()

	if pendingTrial != nil {
		pendingTrial.timer.Stop()
	}

	window := time.Duration(seconds) * time.Second
	deadline := time.Now().Add(window)

	pendingTrial = &trial{
		token:    token,
		mode:     mode,
		config:   config,
		deadline: deadline,
	}
	// Read by every signed-in request. See ethernet_reach.go: a request that
	// arrives over the new address confirms the trial, and the flag is what
	// keeps that check off the mutex while no trial is running.
	trialPending.Store(true)

	pendingTrial.timer = time.AfterFunc(window, func() {
		log.Warnf("ethernet trial %s was not confirmed, putting the saved configuration back", token)
		revertNow()
	})

	// The timer above is in this process, and this process is not the thing
	// the trial protects against. See ethernet_guard.go: the same deadline is
	// written to /run and handed to a detached shell, so a server that dies
	// between the apply and the confirmation still leaves something behind
	// that puts the address back.
	state := trialState{
		Token:    token,
		Mode:     mode,
		Address:  config.Address,
		Prefix:   config.Prefix,
		Gateway:  config.Gateway,
		Deadline: deadline,
	}
	if err := writeTrialState(state); err != nil {
		log.Errorf("failed to record the ethernet trial: %s", err)
		return
	}
	if err := startExternalRevert(token, window+externalRevertGrace); err != nil {
		// The in-process timer still reverts, so the trial goes ahead. Clear
		// the record rather than leave one that claims a revert nothing is
		// counting down to.
		log.Errorf("failed to start the detached ethernet revert: %s", err)
		clearTrialState()
	}
}

func describeTrial() *proto.EthernetTrial {
	trialMutex.Lock()
	defer trialMutex.Unlock()

	if pendingTrial == nil {
		return describeAdoptedTrialLocked()
	}

	return &proto.EthernetTrial{
		Token:            pendingTrial.token,
		Mode:             pendingTrial.mode,
		Address:          pendingTrial.config.Address,
		Prefix:           pendingTrial.config.Prefix,
		Gateway:          pendingTrial.config.Gateway,
		RemainingSeconds: remainingSeconds(pendingTrial.deadline),
	}
}

// describeAdoptedTrialLocked reports a trial that a previous server process
// started. The detached revert outlived that process and is still counting, so
// reporting nothing here would tell the caller the address is settled while it
// is about to change back underneath them.
func describeAdoptedTrialLocked() *proto.EthernetTrial {
	state, ok := readTrialState()
	if !ok {
		return nil
	}

	return &proto.EthernetTrial{
		Token:            state.Token,
		Mode:             state.Mode,
		Address:          state.Address,
		Prefix:           state.Prefix,
		Gateway:          state.Gateway,
		RemainingSeconds: remainingSeconds(state.Deadline),
	}
}

func remainingSeconds(deadline time.Time) int {
	remaining := int(time.Until(deadline).Seconds())
	if remaining < 0 {
		return 0
	}

	return remaining
}

func confirmTrial(token string) error {
	trialMutex.Lock()
	defer trialMutex.Unlock()

	return confirmTrialLocked(token)
}

func confirmTrialLocked(token string) error {
	mode, config, err := awaitingConfirmationLocked(token)
	if err != nil {
		return err
	}

	// Saved before the revert is called off, and not after. A confirmation
	// that cannot write the file has to leave the deadline running: a board
	// that kept an address it failed to save would work until the next reboot
	// and lose it then, which is the one outcome nobody could have predicted
	// from the page that said the change was kept.
	if err := persistEthernet(mode, config); err != nil {
		return err
	}

	if pendingTrial != nil {
		pendingTrial.timer.Stop()
		pendingTrial = nil
	}

	// The detached revert wakes at its own deadline, does not find this token,
	// and exits without touching the interface.
	clearTrialState()
	trialPending.Store(false)

	return nil
}

// awaitingConfirmationLocked returns the change this token confirms. It reads
// the file when this process holds no trial, which is a trial that a previous
// process started: the address is live, the detached revert is still counting,
// and the person holding the token is the one who can end it.
func awaitingConfirmationLocked(token string) (string, ethernetConfig, error) {
	if pendingTrial != nil {
		if pendingTrial.token != token {
			// A confirmation for a trial that a later request replaced must
			// not save the later one.
			return "", ethernetConfig{}, fmt.Errorf("this change is no longer the one waiting to be confirmed")
		}

		return pendingTrial.mode, pendingTrial.config, nil
	}

	state, ok := readTrialState()
	if !ok {
		return "", ethernetConfig{}, fmt.Errorf("there is no change waiting to be confirmed")
	}
	if state.Token != token {
		return "", ethernetConfig{}, fmt.Errorf("this change is no longer the one waiting to be confirmed")
	}

	return state.Mode, ethernetConfig{
		Address: state.Address,
		Prefix:  state.Prefix,
		Gateway: state.Gateway,
	}, nil
}

func persistEthernet(mode string, config ethernetConfig) error {
	var err error
	if mode == ethModeStatic {
		err = writeEthernetConfig(config)
	} else {
		err = removeEthernetConfig()
	}
	if err != nil {
		return err
	}

	_ = exec.Command("sync").Run()

	return nil
}

// revertNow puts the interface back on the saved configuration. It runs the
// boot script rather than repeat what the script does, so a reverted board and
// a rebooted board end up in the same state.
func revertNow() {
	trialMutex.Lock()
	if pendingTrial != nil {
		pendingTrial.timer.Stop()
		pendingTrial = nil
	}
	// Cleared here too, so the detached revert finds no token and does not run
	// the boot script a second time behind this one.
	clearTrialState()
	trialPending.Store(false)
	trialMutex.Unlock()

	stopDHCPClient()

	if err := runCommand(ethInitScript, "start"); err != nil {
		log.Errorf("failed to restore the ethernet configuration: %s", err)
	}
}

func applyStatic(config ethernetConfig) error {
	stopDHCPClient()

	// -4, because this panel manages IPv4 and nothing else. A bare flush takes
	// the IPv6 addresses with it, the link-local included.
	//
	// Measured on the device on 2026-09-07. Right after a revert eth0 had no
	// IPv6 at all, and it still had none several minutes later. Some time after
	// that the global address came back on its own and the link-local did not,
	// which is not a state the interface reaches by itself. Disabling and
	// re-enabling IPv6 on the interface restored both. disable_ipv6 was 0
	// throughout, so nothing had turned IPv6 off.
	//
	// So the cost is an address family that goes away for minutes at a time and
	// comes back incompletely, on hardware whose whole job is being reachable
	// when other things are not. It is not a permanent loss and it does not
	// need a link bounce, which is what the first reading of it suggested.
	//
	// S30eth carries the same flag for the same reason. Its own flush runs
	// before the link is up at boot, where the loss does not show, and it runs
	// again on every revert, where it does.
	if err := runCommand("ip", "-4", "addr", "flush", "dev", ethInterface); err != nil {
		return err
	}

	address := fmt.Sprintf("%s/%d", config.Address, config.Prefix)
	if err := runCommand("ip", "addr", "add", address, "brd", "+", "dev", ethInterface); err != nil {
		return err
	}

	if config.Gateway == "" {
		return nil
	}

	return runCommand("ip", "route", "add", "default", "via", config.Gateway, "dev", ethInterface)
}

func applyDHCP() error {
	stopDHCPClient()

	if err := runCommand("ip", "-4", "addr", "flush", "dev", ethInterface); err != nil {
		return err
	}

	// The same arguments S30eth uses, so a trial and a boot ask for the same
	// things. -O 121 requests the classless static routes: without it a board
	// that took its lease from a trial would be missing routes that the same
	// board has after a reboot, and the difference would show up later as a
	// subnet that stopped being reachable. -B asks the server to broadcast its
	// reply, which is the only delivery a relay agent can perform for a client
	// that holds no address yet; busybox sets the flag only while ciaddr is
	// zero, so a renewal stays unicast. -b puts udhcpc in the background once
	// it has a lease.
	return runCommand("udhcpc", "-i", ethInterface, "-t", "10", "-T", "1", "-A", "5",
		"-O", "121", "-B", "-b", "-p", udhcpcPidFile)
}

// stopDHCPClient ends the lease loop. A client left running would put its own
// address back over a static one at the next renewal.
func stopDHCPClient() {
	data, err := os.ReadFile(udhcpcPidFile)
	if err != nil {
		return
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || pid <= 1 {
		return
	}

	process, err := os.FindProcess(pid)
	if err == nil {
		_ = process.Kill()
	}

	_ = os.Remove(udhcpcPidFile)
}

// runCommand is a var so the tests can watch what would run without changing
// the interface of the machine they run on.
var runCommand = func(name string, args ...string) error {
	output, err := exec.Command(name, args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %s: %w: %s", name, strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}

	return nil
}
