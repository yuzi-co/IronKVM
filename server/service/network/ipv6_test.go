package network

import (
	"errors"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// useTempIPv6 points every IPv6 path at a scratch tree and records the sysctl
// writes. The kernel has IPv6 and the image has a DHCPv6 client, unless the
// test removes either.
func useTempIPv6(t *testing.T) (dir string, sysctls *[]string) {
	t.Helper()

	dir = t.TempDir()
	oldConfig, oldInet, oldRoute, oldMark, oldLook, oldWrite := ipv6ConfigFile, ipv6InetFile, ipv6RouteFile, ipv6RestoredMark, lookPath, writeIPv6Sysctl
	ipv6ConfigFile = filepath.Join(dir, "etc-kvm", "ipv6")
	ipv6InetFile = filepath.Join(dir, "if_inet6")
	ipv6RouteFile = filepath.Join(dir, "ipv6_route")
	ipv6RestoredMark = filepath.Join(dir, "restored")
	lookPath = func(name string) (string, error) { return "/sbin/" + name, nil }

	if err := os.WriteFile(ipv6InetFile, []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	log := &[]string{}
	writeIPv6Sysctl = func(key, value string) error {
		*log = append(*log, key+"="+value)
		return nil
	}

	t.Cleanup(func() {
		ipv6ConfigFile, ipv6InetFile, ipv6RouteFile, ipv6RestoredMark, lookPath, writeIPv6Sysctl = oldConfig, oldInet, oldRoute, oldMark, oldLook, oldWrite
	})

	return dir, log
}

func staticConfig() ipv6Config {
	return ipv6Config{Mode: ipv6ModeStatic, Address: "2001:db8:1::50", Prefix: 64, Gateway: "fe80::1"}
}

func TestValidateIPv6Static(t *testing.T) {
	good := []ipv6Config{
		staticConfig(),
		{Address: "fd00::10", Prefix: 64},
		{Address: "2001:db8:1::50", Prefix: 64, Gateway: "2001:db8:1::1"},
		{Address: "2001:db8:1::51", Prefix: 126, Gateway: "2001:db8:1::52"},
	}
	for _, c := range good {
		if err := validateIPv6Static(c); err != nil {
			t.Errorf("%+v rejected: %v", c, err)
		}
	}

	bad := map[string]ipv6Config{
		"empty":             {},
		"ipv4":              {Address: "192.168.1.5", Prefix: 64},
		"mapped":            {Address: "::ffff:192.168.1.5", Prefix: 64},
		"zone":              {Address: "fe80::1%eth0", Prefix: 64},
		"loopback":          {Address: "::1", Prefix: 64},
		"unspecified":       {Address: "::", Prefix: 64},
		"multicast":         {Address: "ff02::1", Prefix: 64},
		"link-local":        {Address: "fe80::5", Prefix: 64},
		"prefix zero":       {Address: "2001:db8::5", Prefix: 0},
		"prefix too short":  {Address: "2001:db8::5", Prefix: 15},
		"prefix too long":   {Address: "2001:db8::5", Prefix: 127},
		"network address":   {Address: "2001:db8:1::", Prefix: 64},
		"router is ipv4":    {Address: "2001:db8:1::5", Prefix: 64, Gateway: "192.168.1.1"},
		"router is device":  {Address: "2001:db8:1::5", Prefix: 64, Gateway: "2001:db8:1::5"},
		"router off subnet": {Address: "2001:db8:1::5", Prefix: 64, Gateway: "2001:db8:2::1"},
		"router multicast":  {Address: "2001:db8:1::5", Prefix: 64, Gateway: "ff02::2"},
		"router unspec":     {Address: "2001:db8:1::5", Prefix: 64, Gateway: "::"},
	}
	for name, c := range bad {
		if err := validateIPv6Static(c); err == nil {
			t.Errorf("%s: %+v accepted", name, c)
		}
	}
}

func TestIPv6ChangeAllowedFromIPv6Client(t *testing.T) {
	v4 := netip.MustParseAddr("192.168.1.10")
	v6 := netip.MustParseAddr("2001:db8::10")
	mapped := netip.MustParseAddr("::ffff:192.168.1.10")
	slaac := defaultIPv6Config()
	dhcp := ipv6Config{Mode: ipv6ModeDHCPv6}

	for _, client := range []netip.Addr{v4, mapped, {}} {
		for _, next := range []ipv6Config{dhcp, staticConfig(), {Mode: ipv6ModeOff}} {
			if !ipv6ChangeAllowed(client, slaac, next) {
				t.Errorf("a change from %v refused", client)
			}
		}
	}

	if !ipv6ChangeAllowed(v6, slaac, slaac) {
		t.Error("a request that changes nothing was refused")
	}
	if !ipv6ChangeAllowed(v6, slaac, dhcp) {
		t.Error("DHCPv6 on top of SLAAC adds addresses and was refused")
	}
	for name, c := range map[string][2]ipv6Config{
		"slaac to off":      {slaac, {Mode: ipv6ModeOff}},
		"slaac to static":   {slaac, staticConfig()},
		"static to slaac":   {staticConfig(), slaac},
		"dhcpv6 to slaac":   {dhcp, slaac},
		"static to another": {staticConfig(), {Mode: ipv6ModeStatic, Address: "2001:db8:1::51", Prefix: 64}},
	} {
		if ipv6ChangeAllowed(v6, c[0], c[1]) {
			t.Errorf("%s over IPv6 was allowed", name)
		}
	}
}

func TestParseIPv6DefaultRoute(t *testing.T) {
	route := strings.Join([]string{
		"00000000000000000000000000000001 80 00000000000000000000000000000000 00 00000000000000000000000000000000 00000000 00000003 00000000 80200001 lo",
		"fe800000000000000000000000000000 40 00000000000000000000000000000000 00 00000000000000000000000000000000 00000100 00000001 00000000 00000001 eth0",
		"00000000000000000000000000000000 00 00000000000000000000000000000000 00 fe800000000000000000000000000001 00000400 00000001 00000000 00000003 eth0",
		"00000000000000000000000000000000 00 00000000000000000000000000000000 00 fe800000000000000000000000000009 00000400 00000001 00000000 00000003 wlan0",
		"",
	}, "\n")

	if got := parseIPv6DefaultRoute(route, "eth0"); got != "fe80::1" {
		t.Errorf("eth0 gateway = %q", got)
	}
	if got := parseIPv6DefaultRoute(route, "wlan0"); got != "fe80::9" {
		t.Errorf("wlan0 gateway = %q", got)
	}
	if got := parseIPv6DefaultRoute(route, "usb0"); got != "" {
		t.Errorf("usb0 gateway = %q", got)
	}
	// A default route with no gateway flag is a device route, not a router.
	noGateway := "00000000000000000000000000000000 00 00000000000000000000000000000000 00 00000000000000000000000000000000 00000400 00000001 00000000 00000001 eth0"
	if got := parseIPv6DefaultRoute(noGateway, "eth0"); got != "" {
		t.Errorf("gateway of a device route = %q", got)
	}
}

// Nothing the IPv6 path runs may touch IPv4, or the process that holds the
// IPv4 lease.
func TestApplyIPv6NeverTouchesIPv4(t *testing.T) {
	_, sysctls := useTempIPv6(t)

	for _, config := range []ipv6Config{
		{Mode: ipv6ModeSLAAC}, {Mode: ipv6ModeDHCPv6}, staticConfig(), {Mode: ipv6ModeOff},
	} {
		for _, prev := range []string{ipv6ModeSLAAC, ipv6ModeDHCPv6, ipv6ModeStatic, ipv6ModeOff} {
			commands := recordCommands(t)
			if err := applyIPv6(prev, config); err != nil {
				t.Fatalf("%s -> %s: %v", prev, config.Mode, err)
			}
			for _, line := range commands.lines {
				fields := strings.Fields(line)
				switch fields[0] {
				case "ip":
					if len(fields) < 2 || fields[1] != "-6" {
						t.Errorf("%s -> %s ran %q", prev, config.Mode, line)
					}
				case dhcpv6Client:
				default:
					t.Errorf("%s -> %s ran %q", prev, config.Mode, line)
				}
				if strings.Contains(line, "udhcpc ") || strings.Contains(line, " -4 ") || strings.Contains(line, ethInitScript) {
					t.Errorf("%s -> %s touched IPv4: %q", prev, config.Mode, line)
				}
			}
		}
	}
	if len(*sysctls) == 0 {
		t.Error("no sysctl was written")
	}
}

func TestApplyStaticCommandsAndSysctls(t *testing.T) {
	_, sysctls := useTempIPv6(t)
	commands := recordCommands(t)

	if err := applyIPv6(ipv6ModeSLAAC, staticConfig()); err != nil {
		t.Fatal(err)
	}

	want := []string{
		"ip -6 addr flush dev eth0 scope global",
		"ip -6 route del default dev eth0",
		"ip -6 addr add 2001:db8:1::50/64 dev eth0",
		"ip -6 route replace default via fe80::1 dev eth0",
	}
	if strings.Join(commands.lines, "\n") != strings.Join(want, "\n") {
		t.Errorf("commands:\n%s", strings.Join(commands.lines, "\n"))
	}
	if strings.Join(*sysctls, " ") != "accept_ra=0 autoconf=0 disable_ipv6=0" {
		t.Errorf("sysctls = %v", *sysctls)
	}
}

func TestApplySLAACRestartsTheLinkOnlyWhenNeeded(t *testing.T) {
	_, sysctls := useTempIPv6(t)
	recordCommands(t)

	if err := applyIPv6(ipv6ModeStatic, defaultIPv6Config()); err != nil {
		t.Fatal(err)
	}
	if strings.Join(*sysctls, " ") != "disable_ipv6=1 accept_ra=1 autoconf=1 disable_ipv6=0" {
		t.Errorf("static to slaac sysctls = %v", *sysctls)
	}

	*sysctls = nil
	if err := applyIPv6(ipv6ModeSLAAC, ipv6Config{Mode: ipv6ModeDHCPv6}); err != nil {
		t.Fatal(err)
	}
	if strings.Join(*sysctls, " ") != "accept_ra=1 autoconf=1" {
		t.Errorf("slaac to dhcpv6 restarted the link: %v", *sysctls)
	}
}

func TestChangeSavesOnlyAfterApply(t *testing.T) {
	useTempIPv6(t)
	recordCommands(t)

	if err := changeIPv6(staticConfig()); err != nil {
		t.Fatal(err)
	}
	if got := readIPv6Config(); got != staticConfig() {
		t.Fatalf("saved = %+v", got)
	}
}

// A failed apply leaves the saved file alone and applies the saved setting
// again, which is the rollback.
func TestFailedChangeKeepsTheOldSettingAndReappliesIt(t *testing.T) {
	useTempIPv6(t)

	saved := staticConfig()
	if err := writeIPv6Config(saved); err != nil {
		t.Fatal(err)
	}

	var ran []string
	oldRun := runCommand
	t.Cleanup(func() { runCommand = oldRun })
	runCommand = func(name string, args ...string) error {
		line := name + " " + strings.Join(args, " ")
		ran = append(ran, line)
		if strings.Contains(line, "addr add 2001:db8:9::5/64") {
			return errors.New("RTNETLINK answers: File exists")
		}
		return nil
	}

	next := ipv6Config{Mode: ipv6ModeStatic, Address: "2001:db8:9::5", Prefix: 64, Gateway: "fe80::9"}
	if err := changeIPv6(next); err == nil {
		t.Fatal("the failing change succeeded")
	}

	if got := readIPv6Config(); got != saved {
		t.Fatalf("saved setting after the failure = %+v, want %+v", got, saved)
	}
	// The failing add, then the old address put back.
	last := ran[len(ran)-2:]
	if !strings.Contains(last[0], "addr add 2001:db8:1::50/64") || !strings.Contains(last[1], "route replace default via fe80::1") {
		t.Fatalf("the old setting was not applied again:\n%s", strings.Join(ran, "\n"))
	}
}

// When even the way back fails, the interface is handed to the kernel's
// default instead of being left half configured.
func TestFailedRollbackFallsBackToSLAAC(t *testing.T) {
	_, sysctls := useTempIPv6(t)
	if err := writeIPv6Config(staticConfig()); err != nil {
		t.Fatal(err)
	}
	commands := recordCommands(t)
	commands.fail = true

	if err := changeIPv6(ipv6Config{Mode: ipv6ModeStatic, Address: "2001:db8:9::5", Prefix: 64}); err == nil {
		t.Fatal("the failing change succeeded")
	}

	got := strings.Join(*sysctls, " ")
	if !strings.HasSuffix(got, "disable_ipv6=1 accept_ra=1 autoconf=1 disable_ipv6=0") {
		t.Errorf("no fall back to SLAAC at the end: %s", got)
	}
}

func TestDHCPv6NeedsAClient(t *testing.T) {
	useTempIPv6(t)
	recordCommands(t)
	lookPath = func(string) (string, error) { return "", os.ErrNotExist }

	if err := changeIPv6(ipv6Config{Mode: ipv6ModeDHCPv6}); err == nil {
		t.Fatal("DHCPv6 applied without a client")
	}
	if got := readIPv6Config(); got.Mode != ipv6ModeSLAAC {
		t.Fatalf("saved mode = %q", got.Mode)
	}
}

func TestNoIPv6InTheKernel(t *testing.T) {
	useTempIPv6(t)
	commands := recordCommands(t)
	if err := os.Remove(ipv6InetFile); err != nil {
		t.Fatal(err)
	}

	if err := changeIPv6(staticConfig()); err == nil {
		t.Fatal("applied without IPv6 in the kernel")
	}
	if len(commands.lines) != 0 {
		t.Errorf("ran %v", commands.lines)
	}
	if _, err := os.Stat(ipv6ConfigFile); !os.IsNotExist(err) {
		t.Errorf("a setting was saved: %v", err)
	}
}

func TestReadIPv6ConfigDefaults(t *testing.T) {
	useTempIPv6(t)

	if got := readIPv6Config(); got != defaultIPv6Config() {
		t.Errorf("no file: %+v", got)
	}

	for _, content := range []string{
		"not json",
		`{"mode":"banana"}`,
		`{"mode":"static","address":"192.168.1.1","prefix":24}`,
	} {
		if err := os.MkdirAll(filepath.Dir(ipv6ConfigFile), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ipv6ConfigFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
		if got := readIPv6Config(); got != defaultIPv6Config() {
			t.Errorf("%q read as %+v", content, got)
		}
	}
}

func TestRestoreRunsOncePerBootAndSkipsTheDefault(t *testing.T) {
	useTempIPv6(t)
	commands := recordCommands(t)

	// The default needs nothing.
	RestoreIPv6()
	if len(commands.lines) != 0 {
		t.Fatalf("restored the default: %v", commands.lines)
	}

	if err := os.Remove(ipv6RestoredMark); err != nil {
		t.Fatal(err)
	}
	if err := writeIPv6Config(staticConfig()); err != nil {
		t.Fatal(err)
	}
	RestoreIPv6()
	if len(commands.lines) == 0 {
		t.Fatal("the saved static setting was not applied")
	}

	n := len(commands.lines)
	RestoreIPv6()
	if len(commands.lines) != n {
		t.Fatal("restored twice in one boot")
	}
}
