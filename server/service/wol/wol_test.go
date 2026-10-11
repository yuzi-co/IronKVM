package wol

import (
	"context"
	"errors"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func testStore(t *testing.T) (*Store, string, string) {
	t.Helper()
	dir := t.TempDir()
	mac := filepath.Join(dir, "cache", "wol")
	opt := filepath.Join(dir, "wol-devices.json")

	return NewStore(mac, opt), mac, opt
}

func TestParseMAC(t *testing.T) {
	for in, want := range map[string]string{
		"aa:bb:cc:dd:ee:ff": "AA:BB:CC:DD:EE:FF",
		"AA-BB-CC-DD-EE-FF": "AA:BB:CC:DD:EE:FF",
		"aabb.ccdd.eeff":    "AA:BB:CC:DD:EE:FF",
		" aabbccddeeff ":    "AA:BB:CC:DD:EE:FF",
	} {
		got, err := ParseMAC(in)
		if err != nil || got != want {
			t.Errorf("ParseMAC(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "aa:bb", "gg:bb:cc:dd:ee:ff", "aabbccddeeff00"} {
		if _, err := ParseMAC(in); err == nil {
			t.Errorf("ParseMAC(%q) accepted", in)
		}
	}
}

// The file an older application wrote must read as it always did, with names
// that hold spaces, and must stay in that format after a rename.
func TestLegacyFileReadAndKept(t *testing.T) {
	s, macFile, optFile := testStore(t)
	if err := os.MkdirAll(filepath.Dir(macFile), 0o755); err != nil {
		t.Fatal(err)
	}
	legacy := "aa:bb:cc:dd:ee:01\nAA:BB:CC:DD:EE:02 Living room PC\n\n  AA-BB-CC-DD-EE-03   nas  \n"
	if err := os.WriteFile(macFile, []byte(legacy), 0o644); err != nil {
		t.Fatal(err)
	}

	list, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, d := range list {
		got = append(got, d.Mac+"|"+d.Name)
	}
	want := []string{"AA:BB:CC:DD:EE:01|", "AA:BB:CC:DD:EE:02|Living room PC", "AA:BB:CC:DD:EE:03|nas"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("list = %v, want %v", got, want)
	}

	if err := s.Rename("AA:BB:CC:DD:EE:01", "Desk"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(macFile)
	if string(data) != "AA:BB:CC:DD:EE:01 Desk\nAA:BB:CC:DD:EE:02 Living room PC\nAA:BB:CC:DD:EE:03 nas\n" {
		t.Fatalf("file after rename:\n%s", data)
	}
	// No options yet, so no options file.
	if _, err := os.Stat(optFile); !os.IsNotExist(err) {
		t.Fatalf("options file exists after a rename: %v", err)
	}
}

func TestSaveAddEditDuplicate(t *testing.T) {
	s, macFile, optFile := testStore(t)

	d, err := s.Save("", Device{Mac: "aa:bb:cc:dd:ee:01", Name: "  NAS  ", Options: Options{
		Broadcast: "192.168.1.0/24", Port: 7, SecureOn: "11:22:33:44:55:66", IP: "192.168.1.10",
	}}, false)
	if err != nil {
		t.Fatal(err)
	}
	if d.Mac != "AA:BB:CC:DD:EE:01" || d.Name != "NAS" {
		t.Fatalf("saved %+v", d)
	}

	// The legacy file carries only the MAC and the name.
	data, _ := os.ReadFile(macFile)
	if string(data) != "AA:BB:CC:DD:EE:01 NAS\n" {
		t.Fatalf("legacy file = %q", data)
	}
	// The password file is private.
	info, err := os.Stat(optFile)
	if err != nil {
		t.Fatal(err)
	}
	if runtimeHasUnixModes() && info.Mode().Perm() != 0o600 {
		t.Fatalf("options mode = %v", info.Mode().Perm())
	}

	for _, tc := range []struct {
		name string
		dev  Device
	}{
		{"same MAC", Device{Mac: "AA-BB-CC-DD-EE-01", Name: "other"}},
		{"same name, other case", Device{Mac: "AA:BB:CC:DD:EE:02", Name: "nas"}},
	} {
		if _, err := s.Save("", tc.dev, false); !errors.Is(err, ErrDuplicate) {
			t.Errorf("%s: err = %v, want ErrDuplicate", tc.name, err)
		}
	}

	// Edit that keeps the stored password and changes the MAC.
	d, err = s.Save("AA:BB:CC:DD:EE:01", Device{Mac: "AA:BB:CC:DD:EE:09", Name: "NAS", Options: Options{Port: 9}}, true)
	if err != nil {
		t.Fatal(err)
	}
	if d.SecureOn != "11:22:33:44:55:66" || d.Port != 9 {
		t.Fatalf("edited %+v", d)
	}
	list, _ := s.List()
	if len(list) != 1 || list[0].Mac != "AA:BB:CC:DD:EE:09" || list[0].SecureOn == "" {
		t.Fatalf("list = %+v", list)
	}
	if opts, _ := s.readOptions(); len(opts) != 1 {
		t.Fatalf("options of the old MAC were kept: %v", opts)
	}

	// Edit that removes the password.
	d, err = s.Save("AA:BB:CC:DD:EE:09", Device{Mac: "AA:BB:CC:DD:EE:09", Name: "NAS"}, false)
	if err != nil || d.SecureOn != "" {
		t.Fatalf("edit without keep: %+v, %v", d, err)
	}

	if _, err := s.Save("AA:BB:CC:DD:EE:77", Device{Mac: "AA:BB:CC:DD:EE:78"}, false); !errors.Is(err, ErrNotFound) {
		t.Fatalf("edit of a missing device: %v", err)
	}

	if err := s.Delete("AA:BB:CC:DD:EE:09"); err != nil {
		t.Fatal(err)
	}
	if list, _ := s.List(); len(list) != 0 {
		t.Fatalf("list after delete = %+v", list)
	}
	if opts, _ := s.readOptions(); len(opts) != 0 {
		t.Fatalf("options after delete = %v", opts)
	}
}

func TestValidateOptions(t *testing.T) {
	bad := []Options{
		{Broadcast: "not-an-ip"},
		{Broadcast: "224.0.0.1"},
		{Broadcast: "10.0.0.0/31"},
		{Broadcast: "fe80::1"},
		{Port: 80},
		{SecureOn: "123"},
		{IP: "::1"},
		{IP: "127.0.0.1"},
		{IP: "example.com"},
	}
	for _, o := range bad {
		if _, err := ValidateOptions(o); err == nil {
			t.Errorf("%+v accepted", o)
		} else {
			var ie *InputError
			if !errors.As(err, &ie) {
				t.Errorf("%+v: error is not an InputError: %v", o, err)
			}
		}
	}
	if _, err := ValidateOptions(Options{Broadcast: "192.168.1.255", Port: 9, SecureOn: "aa-bb-cc-dd-ee-ff", IP: "10.0.0.5"}); err != nil {
		t.Errorf("good options: %v", err)
	}
}

func TestBroadcastTarget(t *testing.T) {
	for in, want := range map[string]string{
		"192.168.1.0/24":  "192.168.1.255",
		"192.168.1.77/24": "192.168.1.255",
		"10.1.0.0/16":     "10.1.255.255",
		"192.168.1.255":   "192.168.1.255",
	} {
		got, err := BroadcastTarget(in)
		if err != nil || got.String() != want {
			t.Errorf("BroadcastTarget(%q) = %v, %v; want %s", in, got, err, want)
		}
	}
}

func TestMagicPacket(t *testing.T) {
	p, err := MagicPacket("AA:BB:CC:DD:EE:FF", "")
	if err != nil || len(p) != 102 {
		t.Fatalf("len = %d, %v", len(p), err)
	}
	for i := 0; i < 6; i++ {
		if p[i] != 0xff {
			t.Fatalf("sync byte %d = %x", i, p[i])
		}
	}
	if p[6] != 0xaa || p[101] != 0xff {
		t.Fatalf("payload = %x", p)
	}

	p, err = MagicPacket("AA:BB:CC:DD:EE:FF", "11:22:33:44:55:66")
	if err != nil || len(p) != 108 || p[102] != 0x11 || p[107] != 0x66 {
		t.Fatalf("with SecureOn: len = %d, %v", len(p), err)
	}
}

func stubSend(t *testing.T) (frames *[][]string, datagrams *[]string) {
	t.Helper()
	etherWake, udp := runEtherWake, sendUDP
	t.Cleanup(func() { runEtherWake, sendUDP = etherWake, udp })

	frames = &[][]string{}
	datagrams = &[]string{}
	runEtherWake = func(args ...string) ([]byte, error) {
		*frames = append(*frames, args)
		return nil, nil
	}
	sendUDP = func(target net.IP, port int, payload []byte) error {
		*datagrams = append(*datagrams, target.String()+":"+string(rune('0'+port))+":"+string(rune('0'+len(payload)/100)))
		return nil
	}

	return frames, datagrams
}

func TestSendChoosesTransport(t *testing.T) {
	frames, datagrams := stubSend(t)

	// No options: the frame ether-wake has always sent.
	if _, err := Send(Device{Mac: "AA:BB:CC:DD:EE:01"}); err != nil {
		t.Fatal(err)
	}
	want := [][]string{{"-b", "AA:BB:CC:DD:EE:01"}}
	if !reflect.DeepEqual(*frames, want) || len(*datagrams) != 0 {
		t.Fatalf("frames = %v, datagrams = %v", *frames, *datagrams)
	}

	// SecureOn alone stays on ether-wake.
	*frames = nil
	if _, err := Send(Device{Mac: "AA:BB:CC:DD:EE:01", Options: Options{SecureOn: "11:22:33:44:55:66"}}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*frames, [][]string{{"-b", "-p", "11:22:33:44:55:66", "AA:BB:CC:DD:EE:01"}}) {
		t.Fatalf("frames = %v", *frames)
	}

	// A subnet or a port selects UDP.
	*frames = nil
	if _, err := Send(Device{Mac: "AA:BB:CC:DD:EE:01", Options: Options{Broadcast: "192.168.1.0/24"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := Send(Device{Mac: "AA:BB:CC:DD:EE:01", Options: Options{Port: 7}}); err != nil {
		t.Fatal(err)
	}
	if len(*frames) != 0 || len(*datagrams) != 2 ||
		!strings.HasPrefix((*datagrams)[0], "192.168.1.255:9") || !strings.HasPrefix((*datagrams)[1], "255.255.255.255:7") {
		t.Fatalf("frames = %v, datagrams = %v", *frames, *datagrams)
	}
}

func TestSendFailureHidesPassword(t *testing.T) {
	etherWake := runEtherWake
	t.Cleanup(func() { runEtherWake = etherWake })
	runEtherWake = func(args ...string) ([]byte, error) {
		return []byte("ether-wake: no such interface\n"), errors.New("exit status 1")
	}

	out, err := Send(Device{Mac: "AA:BB:CC:DD:EE:01", Options: Options{SecureOn: "11:22:33:44:55:66"}})
	if err == nil || out != "ether-wake: no such interface" {
		t.Fatalf("out = %q, err = %v", out, err)
	}
	if strings.Contains(err.Error(), "11:22") {
		t.Fatalf("error leaks the password: %v", err)
	}
}

func TestWakeByNameAndTouch(t *testing.T) {
	s, macFile, _ := testStore(t)
	frames, _ := stubSend(t)

	if _, err := s.Save("", Device{Mac: "AA:BB:CC:DD:EE:01", Name: "NAS"}, false); err != nil {
		t.Fatal(err)
	}

	d, _, err := s.Wake("", "nas")
	if err != nil || d.Mac != "AA:BB:CC:DD:EE:01" {
		t.Fatalf("wake by name: %+v, %v", d, err)
	}
	if _, _, err := s.Wake("", "tv"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("unknown name: %v", err)
	}
	if _, _, err := s.Wake("nonsense", ""); !errors.Is(err, ErrInvalidMAC) {
		t.Fatalf("bad MAC: %v", err)
	}

	// An unlisted MAC joins the list, as it always did; a listed one does not repeat.
	if _, _, err := s.Wake("aa-bb-cc-dd-ee-02", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Wake("aa-bb-cc-dd-ee-02", ""); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(macFile)
	if string(data) != "AA:BB:CC:DD:EE:01 NAS\nAA:BB:CC:DD:EE:02\n" {
		t.Fatalf("file = %q", data)
	}
	if len(*frames) != 3 {
		t.Fatalf("frames = %v", *frames)
	}
}

func TestParseARP(t *testing.T) {
	arp := `IP address       HW type     Flags       HW address            Mask     Device
192.168.1.1      0x1         0x2         aa:bb:cc:00:00:01     *        eth0
192.168.1.20     0x1         0x0         00:00:00:00:00:00     *        eth0
192.168.1.30     0x1         0x2         aa:bb:cc:00:00:03     *        eth0
`
	got := ParseARP(arp)
	want := []Neighbor{{Mac: "AA:BB:CC:00:00:01", IP: "192.168.1.1"}, {Mac: "AA:BB:CC:00:00:03", IP: "192.168.1.30"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseARP = %v", got)
	}
}

func TestScanNamesFromMDNS(t *testing.T) {
	readARPOld, browseOld := readARP, browseMDNS
	t.Cleanup(func() { readARP, browseMDNS = readARPOld, browseOld })
	readARP = func() string {
		return "IP address HW type Flags HW address Mask Device\n192.168.1.5 0x1 0x2 aa:bb:cc:00:00:05 * eth0\n192.168.1.6 0x1 0x2 aa:bb:cc:00:00:06 * eth0\n"
	}
	browseMDNS = func() string {
		return "+;eth0;IPv4;nas;_smb._tcp;local\n=;eth0;IPv4;nas;_smb._tcp;local;nas.local;192.168.1.5;445;\"\"\n=;eth0;IPv6;nas;_smb._tcp;local;nas.local;fe80::1;445;\"\"\n"
	}

	got := Scan()
	if len(got) != 2 || got[0].Name != "nas" || got[1].Name != "" {
		t.Fatalf("Scan = %+v", got)
	}
}

func TestIsUpPingThenARP(t *testing.T) {
	old := runProbe
	t.Cleanup(func() { runProbe = old })

	var calls []string
	runProbe = func(_ context.Context, name string, args ...string) bool {
		calls = append(calls, name)
		return name == "arping"
	}
	if !IsUp("192.168.1.9") || !reflect.DeepEqual(calls, []string{"ping", "arping"}) {
		t.Fatalf("calls = %v", calls)
	}

	calls = nil
	if IsUp("not-an-ip") || len(calls) != 0 {
		t.Fatalf("a hostname was probed: %v", calls)
	}
}

func runtimeHasUnixModes() bool { return os.PathSeparator == '/' }
