package network

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// useTempConfig points the saved configuration at a directory the test may
// write, and puts the real path back afterwards.
func useTempConfig(t *testing.T) string {
	t.Helper()

	original := ethConfigFile
	path := filepath.Join(t.TempDir(), "eth.nodhcp")
	ethConfigFile = path

	t.Cleanup(func() {
		ethConfigFile = original
	})

	return path
}

// recordCommands replaces the command runner with one that only remembers what
// it was asked to do.
func recordCommands(t *testing.T) *commandLog {
	t.Helper()

	original := runCommand
	recorder := &commandLog{}
	runCommand = recorder.record

	t.Cleanup(func() {
		runCommand = original
	})

	return recorder
}

type commandLog struct {
	mutex sync.Mutex
	lines []string
	fail  bool
}

func (l *commandLog) record(name string, args ...string) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.lines = append(l.lines, strings.TrimSpace(name+" "+strings.Join(args, " ")))
	if l.fail {
		return os.ErrPermission
	}

	return nil
}

func (l *commandLog) all() []string {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return append([]string(nil), l.lines...)
}

func (l *commandLog) contains(fragment string) bool {
	for _, line := range l.all() {
		if strings.Contains(line, fragment) {
			return true
		}
	}

	return false
}

// clearTrial makes each test start with no change waiting, because the trial
// is package state.
//
// It also points the record the detached revert reads at a directory the test
// may write, and replaces the detached revert with a recorder. Without the
// second part a test run would write /run on the machine it runs on and leave
// a sleeping shell behind for every trial it starts.
func clearTrial(t *testing.T) *revertLog {
	t.Helper()

	originalFile := trialStateFile
	trialStateFile = filepath.Join(t.TempDir(), "trial.json")

	originalRevert := startExternalRevert
	recorder := &revertLog{}
	startExternalRevert = recorder.record

	stop := func() {
		trialMutex.Lock()
		defer trialMutex.Unlock()

		if pendingTrial != nil {
			pendingTrial.timer.Stop()
			pendingTrial = nil
		}

		// The reachability check in ethernet_reach.go reads these before it
		// reads anything else, and both are package state that would otherwise
		// carry one test's trial into the next.
		trialPending.Store(false)
		adoptOnce = new(sync.Once)
	}

	stop()
	t.Cleanup(func() {
		stop()
		startExternalRevert = originalRevert
		trialStateFile = originalFile
	})

	return recorder
}

type revertLog struct {
	mutex sync.Mutex
	calls []revertCall
	fail  bool
}

type revertCall struct {
	token string
	after time.Duration
}

func (l *revertLog) record(token string, after time.Duration) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	l.calls = append(l.calls, revertCall{token: token, after: after})
	if l.fail {
		return os.ErrPermission
	}

	return nil
}

func (l *revertLog) all() []revertCall {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return append([]revertCall(nil), l.calls...)
}

func TestParseEthernetLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		want   ethernetConfig
		wantOK bool
	}{
		{
			name:   "an address with a prefix and a router",
			line:   "10.0.0.222/24 10.0.0.1",
			want:   ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "10.0.0.1"},
			wantOK: true,
		},
		{
			name:   "an address with no prefix takes the sixteen the script assumes",
			line:   "10.0.0.222",
			want:   ethernetConfig{Address: "10.0.0.222", Prefix: 16},
			wantOK: true,
		},
		{
			name:   "an unusable prefix falls back the same way the script does",
			line:   "10.0.0.222/99 10.0.0.1",
			want:   ethernetConfig{Address: "10.0.0.222", Prefix: 16, Gateway: "10.0.0.1"},
			wantOK: true,
		},
		{
			name:   "trailing carriage returns survive an edit on a windows host",
			line:   "10.0.0.222/24 10.0.0.1\r",
			want:   ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "10.0.0.1"},
			wantOK: true,
		},
		{
			name:   "a router that is not an address is dropped rather than kept",
			line:   "10.0.0.222/24 not-an-address",
			want:   ethernetConfig{Address: "10.0.0.222", Prefix: 24},
			wantOK: true,
		},
		{name: "a comment is not a configuration", line: "# 10.0.0.222/24"},
		{name: "an empty line is not a configuration", line: "   "},
		{name: "a line that is not an address is not a configuration", line: "hello/24"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, ok := parseEthernetLine(test.line)
			if ok != test.wantOK {
				t.Fatalf("ok = %v, want %v", ok, test.wantOK)
			}
			if ok && got != test.want {
				t.Errorf("config = %+v, want %+v", got, test.want)
			}
		})
	}
}

func TestReadEthernetConfigTakesTheFirstUsableLine(t *testing.T) {
	path := useTempConfig(t)

	body := "# saved by the web ui\n\n10.0.0.222/24 10.0.0.1\n10.0.0.223/24 10.0.0.1\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("failed to write the config: %s", err)
	}

	config, isStatic := readEthernetConfig()
	if !isStatic {
		t.Fatal("a file with a usable line should read as static")
	}

	want := ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "10.0.0.1"}
	if config != want {
		t.Errorf("config = %+v, want %+v", config, want)
	}
}

func TestAMissingFileIsTheDHCPCase(t *testing.T) {
	useTempConfig(t)

	if _, isStatic := readEthernetConfig(); isStatic {
		t.Error("a board with no config file should read as dhcp")
	}
}

func TestTheRenderedLineReadsBack(t *testing.T) {
	path := useTempConfig(t)

	want := ethernetConfig{Address: "192.168.4.9", Prefix: 22, Gateway: "192.168.4.1"}
	if err := writeEthernetConfig(want); err != nil {
		t.Fatalf("failed to write the config: %s", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read the config: %s", err)
	}
	if string(data) != "192.168.4.9/22 192.168.4.1\n" {
		t.Errorf("file = %q", string(data))
	}

	got, isStatic := readEthernetConfig()
	if !isStatic || got != want {
		t.Errorf("config = %+v (static %v), want %+v", got, isStatic, want)
	}
}

func TestValidateEthernetConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  ethernetConfig
		wantErr bool
	}{
		{name: "a normal address", config: ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "10.0.0.1"}},
		{name: "no router is allowed", config: ethernetConfig{Address: "10.0.0.222", Prefix: 24}},
		{
			name:    "the network address of the subnet",
			config:  ethernetConfig{Address: "10.0.0.0", Prefix: 24, Gateway: "10.0.0.1"},
			wantErr: true,
		},
		{
			name:    "the broadcast address of the subnet",
			config:  ethernetConfig{Address: "10.0.0.255", Prefix: 24, Gateway: "10.0.0.1"},
			wantErr: true,
		},
		{
			name:    "a router in another subnet",
			config:  ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "10.9.9.1"},
			wantErr: true,
		},
		{
			name:    "a router that is this device",
			config:  ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "10.0.0.222"},
			wantErr: true,
		},
		{
			name:    "a prefix that leaves no room for a router",
			config:  ethernetConfig{Address: "10.0.0.222", Prefix: 31},
			wantErr: true,
		},
		{name: "a prefix of zero", config: ethernetConfig{Address: "10.0.0.222", Prefix: 0}, wantErr: true},
		{name: "a loopback address", config: ethernetConfig{Address: "127.0.0.1", Prefix: 8}, wantErr: true},
		{name: "a multicast address", config: ethernetConfig{Address: "224.0.0.9", Prefix: 24}, wantErr: true},
		{name: "an unset address", config: ethernetConfig{Address: "0.0.0.0", Prefix: 24}, wantErr: true},
		{name: "an ipv6 address", config: ethernetConfig{Address: "fe80::1", Prefix: 24}, wantErr: true},
		{name: "text", config: ethernetConfig{Address: "the router", Prefix: 24}, wantErr: true},
		{
			name:    "a router that is not an address",
			config:  ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "the router"},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateEthernetConfig(test.config)
			if test.wantErr && err == nil {
				t.Error("the configuration was accepted and should not have been")
			}
			if !test.wantErr && err != nil {
				t.Errorf("the configuration was rejected: %s", err)
			}
		})
	}
}

func TestClampTrialSeconds(t *testing.T) {
	tests := []struct {
		in   int
		want int
	}{
		{in: 0, want: defaultTrialSeconds},
		{in: 1, want: minTrialSeconds},
		{in: minTrialSeconds, want: minTrialSeconds},
		{in: 120, want: 120},
		{in: 10000, want: maxTrialSeconds},
		{in: -5, want: minTrialSeconds},
	}

	for _, test := range tests {
		if got := clampTrialSeconds(test.in); got != test.want {
			t.Errorf("clampTrialSeconds(%d) = %d, want %d", test.in, got, test.want)
		}
	}
}

func TestAStaticTrialSetsTheAddressWithoutSavingIt(t *testing.T) {
	path := useTempConfig(t)
	commands := recordCommands(t)
	clearTrial(t)

	config := ethernetConfig{Address: "10.0.0.99", Prefix: 24, Gateway: "10.0.0.1"}
	if err := applyStatic(config); err != nil {
		t.Fatalf("failed to apply: %s", err)
	}

	if !commands.contains("ip -4 addr flush dev eth0") {
		t.Error("the old address was not removed")
	}
	if !commands.contains("ip addr add 10.0.0.99/24 brd + dev eth0") {
		t.Errorf("the new address was not added: %v", commands.all())
	}
	if !commands.contains("ip route add default via 10.0.0.1 dev eth0") {
		t.Error("the default route was not added")
	}

	// The whole point of the trial: a board that loses power here comes back
	// on whatever it had before.
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("applying wrote the configuration file, which only a confirmation may do")
	}
}

func TestAConfirmationSavesTheConfiguration(t *testing.T) {
	path := useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	config := ethernetConfig{Address: "10.0.0.99", Prefix: 24, Gateway: "10.0.0.1"}
	startTrial("token-a", ethModeStatic, config, 60)

	if err := confirmTrial("token-a"); err != nil {
		t.Fatalf("failed to confirm: %s", err)
	}

	saved, isStatic := readEthernetConfig()
	if !isStatic || saved != config {
		t.Errorf("saved = %+v (static %v), want %+v", saved, isStatic, config)
	}
	if _, err := os.ReadFile(path); err != nil {
		t.Errorf("the configuration file was not written: %s", err)
	}
	if describeTrial() != nil {
		t.Error("the trial is still pending after it was confirmed")
	}
}

func TestConfirmingDHCPRemovesTheSavedConfiguration(t *testing.T) {
	path := useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	if err := writeEthernetConfig(ethernetConfig{Address: "10.0.0.99", Prefix: 24}); err != nil {
		t.Fatalf("failed to write the config: %s", err)
	}

	startTrial("token-b", ethModeDHCP, ethernetConfig{}, 60)
	if err := confirmTrial("token-b"); err != nil {
		t.Fatalf("failed to confirm: %s", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("the configuration file survived a switch to dhcp")
	}
}

func TestAConfirmationNeedsTheRightToken(t *testing.T) {
	path := useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startTrial("token-c", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)

	if err := confirmTrial("token-wrong"); err == nil {
		t.Error("a confirmation with the wrong token was accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("a confirmation with the wrong token saved the configuration")
	}
	if describeTrial() == nil {
		t.Error("a confirmation with the wrong token cancelled the trial")
	}
}

// A second request replaces the first, so the token of the first must stop
// working. Otherwise a browser tab left open on the old change could save a
// configuration the board is not running.
func TestASecondTrialRetiresTheFirstToken(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startTrial("token-first", ethModeStatic, ethernetConfig{Address: "10.0.0.98", Prefix: 24}, 60)
	startTrial("token-second", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)

	if err := confirmTrial("token-first"); err == nil {
		t.Error("the retired token still confirmed")
	}

	if err := confirmTrial("token-second"); err != nil {
		t.Fatalf("the current token failed to confirm: %s", err)
	}

	saved, _ := readEthernetConfig()
	if saved.Address != "10.0.0.99" {
		t.Errorf("saved address = %q, want 10.0.0.99", saved.Address)
	}
}

func TestConfirmingWithNoTrialFails(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	if err := confirmTrial("token-none"); err == nil {
		t.Error("a confirmation was accepted with nothing to confirm")
	}
}

// The revert runs the boot script, so a board that reverts and a board that
// reboots come back the same way.
func TestAnUnconfirmedTrialReverts(t *testing.T) {
	path := useTempConfig(t)
	commands := recordCommands(t)
	clearTrial(t)

	startTrial("token-d", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, minTrialSeconds)

	// Reach past the clamp rather than wait out a real trial.
	trialMutex.Lock()
	pendingTrial.timer.Stop()
	pendingTrial.timer = time.AfterFunc(20*time.Millisecond, revertNow)
	trialMutex.Unlock()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if describeTrial() == nil {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if describeTrial() != nil {
		t.Fatal("the trial did not revert")
	}
	if !commands.contains(ethInitScript + " start") {
		t.Errorf("the revert did not run the boot script: %v", commands.all())
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Error("a trial that was never confirmed saved the configuration")
	}
}

// An apply that fails leaves the interface with no address at all, which is
// the one state nothing recovers from on its own.
func TestAFailedApplyRevertsImmediately(t *testing.T) {
	useTempConfig(t)
	commands := recordCommands(t)
	clearTrial(t)

	commands.fail = true
	startTrial("token-e", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24}, 60)

	applyAfterResponse("token-e", ethModeStatic, ethernetConfig{Address: "10.0.0.99", Prefix: 24})

	if describeTrial() != nil {
		t.Error("the trial survived an apply that failed")
	}
	if !commands.contains(ethInitScript + " start") {
		t.Errorf("the failure did not restore the saved configuration: %v", commands.all())
	}
}

func TestIsAlreadyPersisted(t *testing.T) {
	useTempConfig(t)

	config := ethernetConfig{Address: "10.0.0.99", Prefix: 24, Gateway: "10.0.0.1"}

	if !isAlreadyPersisted(ethModeDHCP, ethernetConfig{}) {
		t.Error("a board with no file should read as already on dhcp")
	}
	if isAlreadyPersisted(ethModeStatic, config) {
		t.Error("a board with no file should not read as already static")
	}

	if err := writeEthernetConfig(config); err != nil {
		t.Fatalf("failed to write the config: %s", err)
	}

	if !isAlreadyPersisted(ethModeStatic, config) {
		t.Error("the saved configuration should read as already applied")
	}
	if isAlreadyPersisted(ethModeStatic, ethernetConfig{Address: "10.0.0.98", Prefix: 24, Gateway: "10.0.0.1"}) {
		t.Error("a different address should not read as already applied")
	}
	if isAlreadyPersisted(ethModeDHCP, ethernetConfig{}) {
		t.Error("a static board should not read as already on dhcp")
	}
}

func TestTrialTokensDiffer(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 64; i++ {
		token, err := newTrialToken()
		if err != nil {
			t.Fatalf("failed to make a token: %s", err)
		}
		if len(token) != 32 {
			t.Fatalf("token = %q, want 32 hex characters", token)
		}
		if seen[token] {
			t.Fatalf("token %q was issued twice", token)
		}
		seen[token] = true
	}
}
