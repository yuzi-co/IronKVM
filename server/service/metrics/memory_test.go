package metrics

import (
	"path/filepath"
	"testing"
)

// useMemoryRoots points every memory source at an empty temporary tree and
// returns the three roots.
func useMemoryRoots(t *testing.T) (proc, zram, cgroup string) {
	t.Helper()

	root := t.TempDir()
	proc = filepath.Join(root, "proc")
	zram = filepath.Join(root, "zram0")
	cgroup = filepath.Join(root, "cgroup")

	setVar(t, &procDir, proc)
	setVar(t, &zramDir, zram)
	setVar(t, &cgroupDir, cgroup)

	return proc, zram, cgroup
}

const meminfoFixture = `MemTotal:         229376 kB
MemFree:           20480 kB
MemAvailable:     112640 kB
Buffers:            1024 kB
SwapTotal:        131072 kB
SwapFree:         130048 kB
CmaTotal:          16384 kB
CmaFree:            8192 kB
`

func TestMemoryReadsEverySource(t *testing.T) {
	proc, zram, cgroup := useMemoryRoots(t)
	writeFixture(t, proc, "meminfo", meminfoFixture)
	writeFixture(t, proc, "vmstat", "nr_free_pages 5000\npswpin 12\npswpout 34\n")
	writeFixture(t, proc, "pressure/memory",
		"some avg10=0.00 avg60=0.00 avg300=0.00 total=1500000\n"+
			"full avg10=0.00 avg60=0.00 avg300=0.00 total=250000\n")
	writeFixture(t, zram, "mm_stat", "8192000 2048000 2621440 0 2621440 10 0 0 0\n")
	writeFixture(t, cgroup, "kvm/memory.current", "83886080\n")
	writeFixture(t, cgroup, "kvm/memory.events", "low 0\nhigh 3\nmax 1\noom 0\noom_kill 0\noom_group_kill 0\n")
	writeFixture(t, cgroup, "addons/memory.current", "4194304\n")
	writeFixture(t, cgroup, "addons/memory.events", "low 0\nhigh 0\nmax 0\noom 2\noom_kill 1\n")

	want := `# HELP ironkvm_memory_bytes Memory from /proc/meminfo, by field.
# TYPE ironkvm_memory_bytes gauge
ironkvm_memory_bytes{kind="MemTotal"} 234881024
ironkvm_memory_bytes{kind="MemAvailable"} 115343360
ironkvm_memory_bytes{kind="MemFree"} 20971520
ironkvm_memory_bytes{kind="SwapTotal"} 134217728
ironkvm_memory_bytes{kind="SwapFree"} 133169152
ironkvm_memory_bytes{kind="CmaTotal"} 16777216
ironkvm_memory_bytes{kind="CmaFree"} 8388608
# HELP ironkvm_swap_pages_total Pages swapped in and out, from /proc/vmstat.
# TYPE ironkvm_swap_pages_total counter
ironkvm_swap_pages_total{direction="in"} 12
ironkvm_swap_pages_total{direction="out"} 34
# HELP ironkvm_zram_bytes zram0 sizes, from mm_stat.
# TYPE ironkvm_zram_bytes gauge
ironkvm_zram_bytes{kind="orig_data_size"} 8192000
ironkvm_zram_bytes{kind="compr_data_size"} 2048000
ironkvm_zram_bytes{kind="mem_used_total"} 2621440
# HELP ironkvm_pressure_stall_seconds_total Time tasks stalled waiting for a resource, from /proc/pressure.
# TYPE ironkvm_pressure_stall_seconds_total counter
ironkvm_pressure_stall_seconds_total{resource="memory",kind="some"} 1.5
ironkvm_pressure_stall_seconds_total{resource="memory",kind="full"} 0.25
# HELP ironkvm_cgroup_memory_bytes Memory charged to a cgroup, from memory.current.
# TYPE ironkvm_cgroup_memory_bytes gauge
ironkvm_cgroup_memory_bytes{group="kvm"} 83886080
ironkvm_cgroup_memory_bytes{group="addons"} 4194304
# HELP ironkvm_cgroup_memory_events_total Memory events in a cgroup, from memory.events.
# TYPE ironkvm_cgroup_memory_events_total counter
ironkvm_cgroup_memory_events_total{group="kvm",event="high"} 3
ironkvm_cgroup_memory_events_total{group="kvm",event="max"} 1
ironkvm_cgroup_memory_events_total{group="kvm",event="oom"} 0
ironkvm_cgroup_memory_events_total{group="kvm",event="oom_kill"} 0
ironkvm_cgroup_memory_events_total{group="addons",event="high"} 0
ironkvm_cgroup_memory_events_total{group="addons",event="max"} 0
ironkvm_cgroup_memory_events_total{group="addons",event="oom"} 2
ironkvm_cgroup_memory_events_total{group="addons",event="oom_kill"} 1
`
	assertText(t, render(t, collectMemory), want)
}

// A kernel without CMA, PSI switched off, no zram, no vmstat, a board with no
// add-ons, and a kvm group whose events file cannot be read. Each missing
// source drops its own samples and nothing else.
func TestMemoryLeavesOutWhatIsMissing(t *testing.T) {
	proc, _, cgroup := useMemoryRoots(t)
	writeFixture(t, proc, "meminfo", `MemTotal:         229376 kB
MemFree:           20480 kB
MemAvailable:     112640 kB
SwapTotal:        131072 kB
SwapFree:         130048 kB
`)
	writeFixture(t, cgroup, "kvm/memory.current", "83886080\n")

	want := `# HELP ironkvm_memory_bytes Memory from /proc/meminfo, by field.
# TYPE ironkvm_memory_bytes gauge
ironkvm_memory_bytes{kind="MemTotal"} 234881024
ironkvm_memory_bytes{kind="MemAvailable"} 115343360
ironkvm_memory_bytes{kind="MemFree"} 20971520
ironkvm_memory_bytes{kind="SwapTotal"} 134217728
ironkvm_memory_bytes{kind="SwapFree"} 133169152
# HELP ironkvm_cgroup_memory_bytes Memory charged to a cgroup, from memory.current.
# TYPE ironkvm_cgroup_memory_bytes gauge
ironkvm_cgroup_memory_bytes{group="kvm"} 83886080
`
	assertText(t, render(t, collectMemory), want)
}

func TestMemoryWritesNothingOnABareHost(t *testing.T) {
	useMemoryRoots(t)

	assertText(t, render(t, collectMemory), "")
}

func TestParseKBFieldsConvertsKilobytesToBytes(t *testing.T) {
	got := parseKBFields("Name:\tNanoKVM-Server\nVmRSS:\t   40960 kB\nThreads:\t12\n")

	if got["VmRSS"] != 41943040 {
		t.Errorf("VmRSS = %d, want 41943040", got["VmRSS"])
	}
	if got["Threads"] != 12 {
		t.Errorf("Threads = %d, want 12", got["Threads"])
	}
	if _, ok := got["Name"]; ok {
		t.Error("a line whose value is not a number must be skipped")
	}
}
