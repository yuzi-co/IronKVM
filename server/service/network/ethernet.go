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

	go applyAfterResponse(req.Mode, config)

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
func applyAfterResponse(mode string, config ethernetConfig) {
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
	}
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

	pendingTrial = &trial{
		token:    token,
		mode:     mode,
		config:   config,
		deadline: time.Now().Add(time.Duration(seconds) * time.Second),
	}
	pendingTrial.timer = time.AfterFunc(time.Duration(seconds)*time.Second, func() {
		log.Warnf("ethernet trial %s was not confirmed, putting the saved configuration back", token)
		revertNow()
	})
}

func describeTrial() *proto.EthernetTrial {
	trialMutex.Lock()
	defer trialMutex.Unlock()

	if pendingTrial == nil {
		return nil
	}

	remaining := int(time.Until(pendingTrial.deadline).Seconds())
	if remaining < 0 {
		remaining = 0
	}

	return &proto.EthernetTrial{
		Token:            pendingTrial.token,
		Mode:             pendingTrial.mode,
		Address:          pendingTrial.config.Address,
		Prefix:           pendingTrial.config.Prefix,
		Gateway:          pendingTrial.config.Gateway,
		RemainingSeconds: remaining,
	}
}

func confirmTrial(token string) error {
	trialMutex.Lock()

	if pendingTrial == nil {
		trialMutex.Unlock()
		return fmt.Errorf("there is no change waiting to be confirmed")
	}
	if pendingTrial.token != token {
		trialMutex.Unlock()
		// A confirmation for a trial that a later request replaced must not
		// save the later one.
		return fmt.Errorf("this change is no longer the one waiting to be confirmed")
	}

	confirmed := pendingTrial
	confirmed.timer.Stop()
	pendingTrial = nil
	trialMutex.Unlock()

	var err error
	if confirmed.mode == ethModeStatic {
		err = writeEthernetConfig(confirmed.config)
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
	trialMutex.Unlock()

	stopDHCPClient()

	if err := runCommand(ethInitScript, "start"); err != nil {
		log.Errorf("failed to restore the ethernet configuration: %s", err)
	}
}

func applyStatic(config ethernetConfig) error {
	stopDHCPClient()

	if err := runCommand("ip", "addr", "flush", "dev", ethInterface); err != nil {
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

	if err := runCommand("ip", "addr", "flush", "dev", ethInterface); err != nil {
		return err
	}

	// The same arguments S30eth uses, so a trial and a boot ask for the same
	// things. -b puts udhcpc in the background once it has a lease.
	return runCommand("udhcpc", "-i", ethInterface, "-t", "10", "-T", "1", "-A", "5",
		"-b", "-p", udhcpcPidFile)
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
