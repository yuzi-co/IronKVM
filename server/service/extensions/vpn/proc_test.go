package vpn

import (
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/extensions/addon"
)

// scratchProc is a /proc with one netbird at pid 4242 that started 500 s after
// boot on a system up 1700 s, and an addons group at 61.6 MB of 64M.
func scratchProc(t *testing.T) {
	t.Helper()
	base := t.TempDir()
	saved := struct {
		proc, cg, marker, initd string
		ts, nb                  addon.Daemon
	}{addon.ProcDir, CgroupDir, addon.DistroMarker, addon.InitdDir, addon.Tailscale, addon.NetBird}
	t.Cleanup(func() {
		addon.ProcDir, CgroupDir, addon.DistroMarker, addon.InitdDir = saved.proc, saved.cg, saved.marker, saved.initd
		addon.Tailscale, addon.NetBird = saved.ts, saved.nb
	})
	addon.ProcDir = filepath.Join(base, "proc")
	CgroupDir = filepath.Join(base, "addons")
	addon.DistroMarker = filepath.Join(base, "absent")
	addon.InitdDir = filepath.Join(base, "init.d")
	addon.Tailscale.PidFile = filepath.Join(base, "tailscaled.pid")
	addon.NetBird.PidFile = filepath.Join(base, "netbird.pid")

	files := map[string]string{
		"proc/4242/stat":        "4242 (net bird) S 1 4242 4242 0 -1 4194560 100 0 0 0 5 3 0 0 20 0 9 0 50000 44000000 7950\n",
		"proc/4242/status":      "Name:\tnetbird\nVmPeak:\t   50000 kB\nVmRSS:\t   31800 kB\n",
		"proc/4242/cmdline":     "/usr/bin/netbird\x00service\x00run\x00",
		"proc/uptime":           "1700.25 3000.00\n",
		"netbird.pid":           "4242\n",
		"addons/memory.current": "64592691\n",
		"addons/memory.high":    "67108864\n",
		"addons/memory.max":     "max\n",
	}
	for name, body := range files {
		path := filepath.Join(base, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestUptimeSec(t *testing.T) {
	scratchProc(t)
	if got := UptimeSec(4242); got != 1200 {
		t.Fatalf("got %d, want 1200", got)
	}
	if got := UptimeSec(9999); got != 0 {
		t.Fatalf("a missing process has no uptime, got %d", got)
	}
}

func TestRSS(t *testing.T) {
	scratchProc(t)
	if got := RSS(4242); got != 31800*1024 {
		t.Fatalf("got %d", got)
	}
}

func TestGroupMemoryReadsMaxAsNoLimit(t *testing.T) {
	scratchProc(t)
	current, high, max := GroupMemory()
	if current != 64592691 || high != 67108864 || max != 0 {
		t.Fatalf("got %d %d %d", current, high, max)
	}
	CgroupDir = filepath.Join(t.TempDir(), "absent")
	if c, h, m := GroupMemory(); c != 0 || h != 0 || m != 0 {
		t.Fatalf("no group reads as zeros, got %d %d %d", c, h, m)
	}
}

func TestFill(t *testing.T) {
	scratchProc(t)
	st := proto.VpnStatus{State: proto.VpnRunning}
	Fill(&st, addon.NetBird)
	if st.UptimeSec != 1200 || st.Memory.DaemonRSS != 31800*1024 {
		t.Fatalf("daemon figures: %+v", st)
	}
	if st.Memory.GroupCurrent != 64592691 || st.Memory.GroupHigh != 67108864 || st.Memory.GroupMax != 0 {
		t.Fatalf("group figures: %+v", st.Memory)
	}
	if st.BootEnabled || st.BlockedBy != "" {
		t.Fatalf("nothing enabled and nothing blocking: %+v", st)
	}
	if st.Peers == nil {
		t.Fatal("peers must be an empty list, not null, for the page")
	}
}
