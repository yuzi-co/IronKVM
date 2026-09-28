package netboot

import (
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestLeasesAreReadFromDnsmasqsFile(t *testing.T) {
	content := "1790000000 02:11:22:33:44:55 172.31.255.2 host-pc 01:02:11:22:33:44:55\n" +
		"0 02:aa:bb:cc:dd:ee 172.31.255.2 * *\n" +
		"garbage\n" +
		"notanumber 02:aa:bb:cc:dd:ee 172.31.255.2 * *\n"

	got := parseLeases([]byte(content))
	want := []Lease{
		{MAC: "02:11:22:33:44:55", IP: "172.31.255.2", Hostname: "host-pc", Expires: time.Unix(1790000000, 0).UTC()},
		{MAC: "02:aa:bb:cc:dd:ee", IP: "172.31.255.2"},
	}
	if !slices.Equal(got, want) {
		t.Fatalf("leases %+v, want %+v", got, want)
	}
}

func TestTheLogTailIsTheLastLines(t *testing.T) {
	dir := t.TempDir()
	name := filepath.Join(dir, "usb.log")

	if got := tailLines(name, 10); got != nil {
		t.Fatalf("a missing log: %q", got)
	}

	var b strings.Builder
	for i := 0; i < 3000; i++ {
		b.WriteString("line " + strconv.Itoa(i) + "\n")
	}
	if err := os.WriteFile(name, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	got := tailLines(name, 3)
	if !slices.Equal(got, []string{"line 2997", "line 2998", "line 2999"}) {
		t.Fatalf("tail %q", got)
	}

	if err := os.WriteFile(name, []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := tailLines(name, 10); !slices.Equal(got, []string{"one", "two"}) {
		t.Fatalf("a short log: %q", got)
	}

	if err := os.WriteFile(name, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got := tailLines(name, 10); got != nil {
		t.Fatalf("an empty log: %q", got)
	}
}

func withRoutes(t *testing.T, content string) {
	t.Helper()
	name := filepath.Join(t.TempDir(), "route")
	if err := os.WriteFile(name, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	saved := procNetRoute
	t.Cleanup(func() { procNetRoute = saved })
	procNetRoute = name
}

const routeHeader = "Iface\tDestination\tGateway \tFlags\tRefCnt\tUse\tMetric\tMask\t\tMTU\tWindow\tIRTT\n"

// The LAN is where the default route goes. The USB link is never it, even on
// a board where something put a default route there.
func TestTheLANIsTheDefaultRoutesInterface(t *testing.T) {
	withRoutes(t, routeHeader+
		"usb0\t00000000\t02FF1FAC\t0003\t0\t0\t0\t00000000\t0\t0\t0\n"+
		"eth0\t0001A8C0\t00000000\t0001\t0\t0\t0\t00FFFFFF\t0\t0\t0\n"+
		"wlan0\t00000000\t0101A8C0\t0002\t0\t0\t600\t00000000\t0\t0\t0\n"+
		"eth0\t00000000\t0101A8C0\t0003\t0\t0\t0\t00000000\t0\t0\t0\n")

	got, err := defaultRouteInterface()
	if err != nil || got != "eth0" {
		t.Fatalf("got %q %v, want eth0 (a route that is not up is skipped)", got, err)
	}

	withRoutes(t, routeHeader+"eth0\t0001A8C0\t00000000\t0001\t0\t0\t0\t00FFFFFF\t0\t0\t0\n")
	if got, err := defaultRouteInterface(); err == nil {
		t.Fatalf("no default route, got %q", got)
	}
}
