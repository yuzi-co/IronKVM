package network

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
)

// stubEthernetAddresses makes the wired interface answer with a fixed set, so
// a test does not depend on the addresses of the machine it runs on.
func stubEthernetAddresses(t *testing.T, addresses ...string) {
	t.Helper()

	original := ethernetCarries
	held := append([]string(nil), addresses...)
	ethernetCarries = func(address string) bool {
		for _, entry := range held {
			if entry == address {
				return true
			}
		}

		return false
	}

	t.Cleanup(func() {
		ethernetCarries = original
	})
}

// startAppliedTrial starts a trial and says the interface carries it, which is
// the state every case below is about.
func startAppliedTrial(token string, mode string, config ethernetConfig) {
	startTrial(token, mode, config, 60)
	markTrialApplied(token)
}

func TestASignedInRequestOnTheNewAddressKeepsTheChange(t *testing.T) {
	path := useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startAppliedTrial("token-reach", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24, Gateway: "10.0.0.1"})

	noteReachable("10.0.0.222", "10.0.0.5:51314")

	if describeTrial() != nil {
		t.Fatal("a client reached the board and the trial is still waiting")
	}

	saved, isStatic := readEthernetConfig()
	if !isStatic || saved.Address != "10.0.0.222" || saved.Prefix != 24 || saved.Gateway != "10.0.0.1" {
		t.Errorf("the confirmed trial was not saved: %+v static=%v", saved, isStatic)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the configuration file was not written: %s", err)
	}
	if _, ok := readTrialState(); ok {
		t.Error("the record the detached revert reads was left behind")
	}
}

func TestOnlyTheAppliedAddressKeepsTheChange(t *testing.T) {
	tests := []struct {
		name   string
		local  string
		remote string
	}{
		{
			name:   "an address the trial did not apply",
			local:  "10.0.0.7",
			remote: "10.0.0.5:51314",
		},
		{
			name:   "the loopback the supervisor probes",
			local:  "127.0.0.1",
			remote: "127.0.0.1:44120",
		},
		{
			name:   "a caller on loopback holding real credentials",
			local:  "10.0.0.222",
			remote: "127.0.0.1:44120",
		},
		{
			name:   "a request with no local address to read",
			local:  "",
			remote: "10.0.0.5:51314",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useTempConfig(t)
			recordCommands(t)
			clearTrial(t)

			startAppliedTrial("token-reach", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24})

			noteReachable(test.local, test.remote)

			if describeTrial() == nil {
				t.Fatal("the trial was confirmed by a request that proves nothing")
			}
			if _, isStatic := readEthernetConfig(); isStatic {
				t.Error("the configuration was saved by a request that proves nothing")
			}
		})
	}
}

// The apply runs half a second after the response leaves. A request that
// arrives in that window came in over the address the board is leaving.
func TestARequestBeforeTheApplyProvesNothing(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startTrial("token-early", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24}, 60)

	noteReachable("10.0.0.222", "10.0.0.5:51314")

	if describeTrial() == nil {
		t.Fatal("a request that arrived before the change was applied confirmed it")
	}

	markTrialApplied("token-early")
	noteReachable("10.0.0.222", "10.0.0.5:51314")

	if describeTrial() != nil {
		t.Error("the same request after the apply did not confirm the trial")
	}
}

func TestMarkTrialAppliedIgnoresAnotherToken(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startTrial("token-current", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24}, 60)
	markTrialApplied("token-replaced")

	noteReachable("10.0.0.222", "10.0.0.5:51314")

	if describeTrial() == nil {
		t.Error("an apply reported against a replaced token armed the trial")
	}
}

// A DHCP trial does not know its own address, so the interface is what says
// whether the lease arrived.
func TestADhcpTrialTakesTheAddressFromTheInterface(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)
	stubEthernetAddresses(t, "10.0.0.150")

	if err := writeEthernetConfig(ethernetConfig{Address: "10.0.0.222", Prefix: 24}); err != nil {
		t.Fatalf("failed to write the starting configuration: %s", err)
	}

	startAppliedTrial("token-dhcp", ethModeDHCP, ethernetConfig{})

	noteReachable("10.0.0.222", "10.0.0.5:51314")
	if describeTrial() == nil {
		t.Fatal("an address the interface does not carry confirmed the lease")
	}

	noteReachable("10.0.0.150", "10.0.0.5:51314")
	if describeTrial() != nil {
		t.Fatal("the leased address did not confirm the trial")
	}

	if _, isStatic := readEthernetConfig(); isStatic {
		t.Error("a confirmed dhcp trial left the static configuration in place")
	}
}

// A server that restarted during a trial holds no trial of its own. The
// detached revert is still counting, and the person who just signed in is the
// one who can end it.
func TestAnAdoptedTrialIsConfirmedByAReachingClient(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	state := trialState{
		Token:   "token-adopted",
		Mode:    ethModeStatic,
		Address: "10.0.0.222",
		Prefix:  24,
		Gateway: "10.0.0.1",
		Applied: true,
	}
	if err := writeTrialState(state); err != nil {
		t.Fatalf("failed to write the trial record: %s", err)
	}

	noteReachable("10.0.0.222", "10.0.0.5:51314")

	saved, isStatic := readEthernetConfig()
	if !isStatic || saved.Address != "10.0.0.222" {
		t.Errorf("the adopted trial was not saved: %+v static=%v", saved, isStatic)
	}
	if _, ok := readTrialState(); ok {
		t.Error("the record the detached revert reads was left behind")
	}
}

// A server that died between writing the record and applying the change leaves
// a trial the interface never took, and neither shape of it may be confirmed.
// A DHCP trial has no address of its own, so the address the board was leaving
// answers for it. A static trial that names the address the board already had
// answers for itself.
func TestAnAdoptedTrialThatWasNeverAppliedProvesNothing(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		address string
	}{
		{name: "a static trial", mode: ethModeStatic, address: "10.0.0.222"},
		{name: "a dhcp trial", mode: ethModeDHCP, address: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			useTempConfig(t)
			recordCommands(t)
			clearTrial(t)
			stubEthernetAddresses(t, "10.0.0.222")

			state := trialState{
				Token:   "token-unapplied",
				Mode:    test.mode,
				Address: test.address,
				Prefix:  24,
			}
			if err := writeTrialState(state); err != nil {
				t.Fatalf("failed to write the trial record: %s", err)
			}

			noteReachable("10.0.0.222", "10.0.0.5:51314")

			if _, isStatic := readEthernetConfig(); isStatic {
				t.Error("a trial that was never applied saved a static configuration")
			}
			if _, ok := readTrialState(); !ok {
				t.Error("a trial that was never applied was called off anyway")
			}
		})
	}
}

// The apply is recorded in the file as well as in memory, so the trial
// survives a restart as an applied one rather than as an unproven one.
func TestApplyingATrialIsRecordedOutsideTheProcess(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startTrial("token-record", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24}, 60)

	state, ok := readTrialState()
	if !ok {
		t.Fatal("the trial was not recorded at all")
	}
	if state.Applied {
		t.Error("a trial is recorded as applied before the interface takes it")
	}

	markTrialApplied("token-record")

	state, ok = readTrialState()
	if !ok {
		t.Fatal("the record went away when the trial was applied")
	}
	if !state.Applied {
		t.Error("the apply was not recorded outside the process")
	}
}

func TestNoTrialMeansNoWork(t *testing.T) {
	useTempConfig(t)
	commands := recordCommands(t)
	clearTrial(t)

	noteReachable("10.0.0.222", "10.0.0.5:51314")

	if _, isStatic := readEthernetConfig(); isStatic {
		t.Error("a request with no trial running saved a configuration")
	}
	if len(commands.all()) != 0 {
		t.Errorf("a request with no trial running ran commands: %v", commands.all())
	}
}

// A confirmation that cannot write the file has to leave the deadline running.
// A board that kept an address it failed to save would work until the next
// reboot and lose it then.
func TestAReachThatCannotSaveLeavesTheTrialRunning(t *testing.T) {
	recordCommands(t)
	clearTrial(t)

	original := ethConfigFile
	ethConfigFile = filepath.Join(t.TempDir(), "absent", "eth.nodhcp")
	t.Cleanup(func() { ethConfigFile = original })

	startAppliedTrial("token-unwritable", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24})

	noteReachable("10.0.0.222", "10.0.0.5:51314")

	if describeTrial() == nil {
		t.Error("a trial that could not be saved stopped waiting anyway")
	}
	if _, ok := readTrialState(); !ok {
		t.Error("a trial that could not be saved called off the detached revert")
	}
}

func TestNoteSignedInAdminRequestReadsTheAcceptingSocket(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startAppliedTrial("token-context", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24})

	request := httptest.NewRequest(http.MethodGet, "/api/network/ethernet", nil)
	request.RemoteAddr = "10.0.0.5:51314"
	request = request.WithContext(context.WithValue(
		request.Context(),
		http.LocalAddrContextKey,
		&net.TCPAddr{IP: net.ParseIP("10.0.0.222"), Port: 443},
	))

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = request

	NoteSignedInAdminRequest(c)

	if describeTrial() != nil {
		t.Error("a signed-in request over the new address did not confirm the trial")
	}
}

// A request that carries no accepting socket, which is every request a test
// builds by hand, must not confirm anything.
func TestNoteSignedInAdminRequestWithoutALocalAddress(t *testing.T) {
	useTempConfig(t)
	recordCommands(t)
	clearTrial(t)

	startAppliedTrial("token-nolocal", ethModeStatic, ethernetConfig{Address: "10.0.0.222", Prefix: 24})

	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/api/network/ethernet", nil)

	NoteSignedInAdminRequest(c)
	NoteSignedInAdminRequest(&gin.Context{})

	if describeTrial() == nil {
		t.Error("a request with no accepting socket confirmed the trial")
	}
}

func TestHostOf(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "an address and a port", value: "10.0.0.222:443", want: "10.0.0.222"},
		{name: "an address on its own", value: "10.0.0.222", want: "10.0.0.222"},
		{
			name:  "the form a dual stack listener writes",
			value: "[::ffff:10.0.0.222]:443",
			want:  "10.0.0.222",
		},
		{name: "an address with a zone", value: "[fe80::1%eth0]:443", want: ""},
		{name: "a name rather than an address", value: "nanokvm.local:443", want: ""},
		{name: "nothing at all", value: "", want: ""},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := hostOf(test.value); got != test.want {
				t.Errorf("hostOf(%q) = %q, want %q", test.value, got, test.want)
			}
		})
	}
}
