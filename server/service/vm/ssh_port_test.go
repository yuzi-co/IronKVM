package vm

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPortConfigMatchesTheBootScript(t *testing.T) {
	want := "# Written by IronKVM: Settings > SSH > Port.\nPort 2222\n"
	if got := portConfig(2222); got != want {
		t.Errorf("portConfig = %q, want %q", got, want)
	}

	// S50sshd writes the same two lines; a change to one must reach the other.
	script, err := os.ReadFile("../../../kvmapp/system/init.d/S50sshd")
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range []string{"'# Written by IronKVM: Settings > SSH > Port.'", `"Port $port"`} {
		if !strings.Contains(string(script), line) {
			t.Errorf("S50sshd does not write %s", line)
		}
	}
}

func TestValidSSHPort(t *testing.T) {
	for port, want := range map[int]bool{0: false, -1: false, 1: true, 22: true, 2222: true, 65535: true, 65536: false} {
		if got := validSSHPort(port); got != want {
			t.Errorf("validSSHPort(%d) = %t", port, got)
		}
	}
}

func TestParseTCPListeners(t *testing.T) {
	tcp := `  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000:0016 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 1 1 0000000000000000 100 0 0 10 0
   1: 00000000:0050 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 2 1 0000000000000000 100 0 0 10 0
   2: DE00000A:0016 0100000A:D431 01 00000000:00000000 02:0004A9C1 00000000     0        0 3 4 0000000000000000 20 4 29 10 -1
`
	tcp6 := `  sl  local_address                         remote_address                        st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
   0: 00000000000000000000000000000000:1F95 00000000000000000000000000000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 4 1 0000000000000000 100 0 0 10 0
`
	ports := map[int]bool{}
	parseTCPListeners([]byte(tcp), ports)
	parseTCPListeners([]byte(tcp6), ports)

	// 22 and 80 listen, 8085 listens on IPv6, and the established
	// connection from port 54321 is not a listener.
	if len(ports) != 3 || !ports[22] || !ports[80] || !ports[8085] || ports[54321] {
		t.Errorf("listeners = %v", ports)
	}
}

func TestListenPortsReadsBothFamilies(t *testing.T) {
	config := parseSSHDConfig([]byte("port 2222\nlistenaddress [::]:2222\nlistenaddress 0.0.0.0:2222 rdomain x\n"))
	if got := listenPorts(config); len(got) != 2 || got[0] != 2222 || got[1] != 2222 {
		t.Errorf("listenPorts = %v", got)
	}
	if !honoursPort(config, 2222) || honoursPort(config, 22) {
		t.Error("honoursPort misread one listener")
	}
	if honoursPort(parseSSHDConfig([]byte("port 2222\nlistenaddress 0.0.0.0:22\n")), 2222) {
		t.Error("a ListenAddress with its own port overrides Port")
	}
}

func TestSetSSHPortWritesTheDropInAndTheSetting(t *testing.T) {
	f := useSSHFixture(t)

	if err := setSSHPort(2222); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, sshPortDropIn); got != portConfig(2222) {
		t.Errorf("drop-in = %q", got)
	}
	if got := readFile(t, sshPortFile); got != "2222\n" {
		t.Errorf("setting = %q", got)
	}
	if f.reloads != 1 {
		t.Errorf("reloads = %d", f.reloads)
	}

	// Back to the default: nothing is left behind.
	f.listening[2222] = true
	if err := setSSHPort(22); err != nil {
		t.Fatal(err)
	}
	if exists(sshPortDropIn) || exists(sshPortFile) {
		t.Error("port 22 left a drop-in or a setting")
	}
}

func TestSetSSHPortLeavesKeysOnlyAlone(t *testing.T) {
	f := useSSHFixture(t)
	line, _ := newKey(t, "")
	f.writeKeys(t, line+"\n")
	if err := setKeysOnly(true); err != nil {
		t.Fatal(err)
	}

	if err := setSSHPort(2222); err != nil {
		t.Fatal(err)
	}
	if !exists(sshKeysOnlyDropIn) || !keysOnlyEnabled() {
		t.Error("changing the port turned keys-only off")
	}
	if readFile(t, sshKeysOnlyDropIn) != keysOnlyConfig {
		t.Error("changing the port rewrote the keys-only drop-in")
	}
}

func TestSetSSHPortRefuses(t *testing.T) {
	f := useSSHFixture(t)
	f.listening[3000] = true

	for _, test := range []struct {
		port int
		want error
	}{
		{0, errPortInvalid},
		{65536, errPortInvalid},
		{80, errPortReserved},
		{443, errPortReserved},
		{8069, errPortReserved},
		{3000, errPortInUse},
	} {
		if err := setSSHPort(test.port); !errors.Is(err, test.want) {
			t.Errorf("port %d: %v, want %v", test.port, err, test.want)
		}
	}
	if exists(sshPortDropIn) || exists(sshPortFile) || f.reloads != 0 {
		t.Error("a refused port changed something")
	}
}

func TestSetSSHPortAllowsSSHDsOwnPort(t *testing.T) {
	f := useSSHFixture(t)
	f.sshdSays["port"] = "2222"
	f.listening[2222] = true

	// sshd itself listens on 2222, so asking for it again is not a clash.
	if err := setSSHPort(2222); err != nil {
		t.Fatal(err)
	}
}

func TestSetSSHPortRollsBackWhatSSHDDoesNotTake(t *testing.T) {
	f := useSSHFixture(t)
	if err := setSSHPort(2222); err != nil {
		t.Fatal(err)
	}
	f.listening[2222] = true

	// A main sshd_config with its own Port line: sshd would listen on both,
	// and the change is taken back, leaving the earlier drop-in in place.
	f.sshdSays["mainPort"] = "22"
	if err := setSSHPort(2200); !errors.Is(err, errPortNotHonoured) {
		t.Fatalf("second Port line: %v", err)
	}
	if readFile(t, sshPortDropIn) != portConfig(2222) || readFile(t, sshPortFile) != "2222\n" {
		t.Error("the earlier port was not restored")
	}

	// An sshd that cannot parse its configuration.
	delete(f.sshdSays, "mainPort")
	if err := os.Remove(sshPortDropIn); err != nil {
		t.Fatal(err)
	}
	f.sshdErr = errors.New("bad configuration")
	if err := setSSHPort(2200); !errors.Is(err, errPortNotHonoured) {
		t.Fatalf("sshd -T failing: %v", err)
	}
	if exists(sshPortDropIn) {
		t.Error("a drop-in sshd rejected was left in place")
	}
	if f.reloads != 1 {
		t.Errorf("reloads = %d, want only the first change", f.reloads)
	}
}

func TestSetSSHPortReportsAFailedReload(t *testing.T) {
	f := useSSHFixture(t)
	f.reloadFn = func() error { return errors.New("no such process") }

	if err := setSSHPort(2222); !errors.Is(err, errReloadFailed) {
		t.Fatalf("got %v", err)
	}
	// Saved all the same: the next start of sshd reads it.
	if readFile(t, sshPortFile) != "2222\n" {
		t.Error("the setting was not kept")
	}
}

func TestSSHPortEndpoint(t *testing.T) {
	f := useSSHFixture(t)
	r := sshEngine()

	for body, code := range map[string]float64{
		`{"port":"x"}`:   -1,
		`{"port":70000}`: -2,
		`{"port":443}`:   -3,
		`{"port":2222}`:  0,
	} {
		if rsp := serve(r, http.MethodPost, "/api/vm/ssh/port", body); rsp["code"] != code {
			t.Errorf("%s answered %v, want code %v", body, rsp, code)
		}
	}
	if f.reloads != 1 {
		t.Errorf("reloads = %d", f.reloads)
	}

	data := serve(r, http.MethodGet, "/api/vm/ssh", "")["data"].(map[string]any)
	if data["port"] != float64(2222) {
		t.Errorf("state port = %v", data["port"])
	}
}

func TestSetSSHPortMovesTheMDNSAnnouncement(t *testing.T) {
	f := useSSHFixture(t)
	service := `<service>
    <type>_ssh._tcp</type>
    <port>22</port>
  </service>
`
	if err := os.MkdirAll(filepath.Dir(avahiSSHServices[0]), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(avahiSSHServices[0], []byte(service), 0o644); err != nil {
		t.Fatal(err)
	}
	// The sftp file is missing, as on an image without it: skipped, no error.

	if err := setSSHPort(2222); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, avahiSSHServices[0]); !strings.Contains(got, "<port>2222</port>") || strings.Contains(got, "<port>22</port>") {
		t.Errorf("service = %q", got)
	}
	if f.avahiReloads != 1 {
		t.Errorf("avahi reloads = %d", f.avahiReloads)
	}

	f.listening[2222] = true
	if err := setSSHPort(22); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, avahiSSHServices[0]); got != service {
		t.Errorf("back to 22, service = %q", got)
	}
}
