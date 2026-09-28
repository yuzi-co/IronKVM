package metrics

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// useNodeRoots points /proc and /sys at an empty temporary tree and returns
// the two roots.
func useNodeRoots(t *testing.T) (proc, sys string) {
	t.Helper()

	root := t.TempDir()
	proc = filepath.Join(root, "proc")
	sys = filepath.Join(root, "sys")
	setVar(t, &procDir, proc)
	setVar(t, &sysDir, sys)

	return proc, sys
}

// assertLines fails for each line of want that got does not hold, and for each
// fragment of absent that it does.
func assertLines(t *testing.T, got string, want, absent []string) {
	t.Helper()

	lines := make(map[string]bool)
	for _, line := range strings.Split(got, "\n") {
		lines[line] = true
	}
	for _, line := range want {
		if !lines[line] {
			t.Errorf("missing line %q", line)
		}
	}
	for _, fragment := range absent {
		if strings.Contains(got, fragment) {
			t.Errorf("unexpected %q", fragment)
		}
	}
	if t.Failed() {
		t.Logf("exposition text:\n%s", got)
	}
}

func TestNodeSystemReadsUnameStatAndLoad(t *testing.T) {
	proc, _ := useNodeRoots(t)
	setVar(t, &nodeNow, func() time.Time { return time.Unix(1790000000, 500000000) })
	writeFixture(t, proc, "sys/kernel/ostype", "Linux\n")
	writeFixture(t, proc, "sys/kernel/osrelease", "5.10.4-tag-\n")
	writeFixture(t, proc, "sys/kernel/version", "#1 PREEMPT Mon Sep 1 00:00:00 UTC 2026\n")
	writeFixture(t, proc, "sys/kernel/hostname", "kvm\n")
	writeFixture(t, proc, "sys/kernel/domainname", "(none)\n")
	writeFixture(t, proc, "sys/kernel/arch", "riscv64\n")
	writeFixture(t, proc, "stat", `cpu  1150 20 330 90000 45 0 5 0 0 0
cpu0 1150 20 330 90000 45 0 5 0 0 0
intr 123456 0 0 0 44 0
ctxt 987654
btime 1789990000
processes 4321
procs_running 2
procs_blocked 1
softirq 5555 0 1 2 3
`)
	writeFixture(t, proc, "loadavg", "0.52 0.48 0.40 2/97 4321\n")

	want := `# HELP node_uname_info Labeled system information as provided by the uname system call.
# TYPE node_uname_info gauge
node_uname_info{domainname="(none)",machine="riscv64",nodename="kvm",release="5.10.4-tag-",sysname="Linux",version="#1 PREEMPT Mon Sep 1 00:00:00 UTC 2026"} 1
# HELP node_time_seconds System time in seconds since epoch (1970).
# TYPE node_time_seconds gauge
node_time_seconds 1790000000.5
# HELP node_cpu_seconds_total Seconds the CPUs spent in each mode.
# TYPE node_cpu_seconds_total counter
node_cpu_seconds_total{cpu="0",mode="user"} 11.5
node_cpu_seconds_total{cpu="0",mode="nice"} 0.2
node_cpu_seconds_total{cpu="0",mode="system"} 3.3
node_cpu_seconds_total{cpu="0",mode="idle"} 900
node_cpu_seconds_total{cpu="0",mode="iowait"} 0.45
node_cpu_seconds_total{cpu="0",mode="irq"} 0
node_cpu_seconds_total{cpu="0",mode="softirq"} 0.05
node_cpu_seconds_total{cpu="0",mode="steal"} 0
# HELP node_boot_time_seconds Node boot time, in unixtime.
# TYPE node_boot_time_seconds gauge
node_boot_time_seconds 1789990000
# HELP node_context_switches_total Total number of context switches.
# TYPE node_context_switches_total counter
node_context_switches_total 987654
# HELP node_intr_total Total number of interrupts serviced.
# TYPE node_intr_total counter
node_intr_total 123456
# HELP node_forks_total Total number of forks.
# TYPE node_forks_total counter
node_forks_total 4321
# HELP node_procs_running Number of processes in runnable state.
# TYPE node_procs_running gauge
node_procs_running 2
# HELP node_procs_blocked Number of processes blocked waiting for I/O to complete.
# TYPE node_procs_blocked gauge
node_procs_blocked 1
# HELP node_load1 1m load average.
# TYPE node_load1 gauge
node_load1 0.52
# HELP node_load5 5m load average.
# TYPE node_load5 gauge
node_load5 0.48
# HELP node_load15 15m load average.
# TYPE node_load15 gauge
node_load15 0.4
`
	assertText(t, render(t, collectNodeSystem), want)
}

// A kernel older than 6.1 has no /proc/sys/kernel/arch.
func TestNodeUnameMachineWithoutArchFile(t *testing.T) {
	proc, _ := useNodeRoots(t)
	writeFixture(t, proc, "sys/kernel/ostype", "Linux\n")

	got := render(t, collectNodeSystem)
	if !strings.Contains(got, `machine="`+unameMachine()+`"`) || unameMachine() == "" {
		t.Fatalf("machine label is missing or empty:\n%s", got)
	}
}

func TestNodeMemoryNamesEveryField(t *testing.T) {
	proc, _ := useNodeRoots(t)
	writeFixture(t, proc, "meminfo", `MemTotal:         160000 kB
MemFree:           20000 kB
MemAvailable:      90000 kB
Active(anon):       1000 kB
Shmem:               512 kB
HugePages_Total:       0
MemFree:           99999 kB
`)

	want := `# HELP node_memory_MemTotal_bytes Memory information field MemTotal_bytes.
# TYPE node_memory_MemTotal_bytes gauge
node_memory_MemTotal_bytes 163840000
# HELP node_memory_MemFree_bytes Memory information field MemFree_bytes.
# TYPE node_memory_MemFree_bytes gauge
node_memory_MemFree_bytes 20480000
# HELP node_memory_MemAvailable_bytes Memory information field MemAvailable_bytes.
# TYPE node_memory_MemAvailable_bytes gauge
node_memory_MemAvailable_bytes 92160000
# HELP node_memory_Active_anon_bytes Memory information field Active_anon_bytes.
# TYPE node_memory_Active_anon_bytes gauge
node_memory_Active_anon_bytes 1024000
# HELP node_memory_Shmem_bytes Memory information field Shmem_bytes.
# TYPE node_memory_Shmem_bytes gauge
node_memory_Shmem_bytes 524288
# HELP node_memory_HugePages_Total Memory information field HugePages_Total.
# TYPE node_memory_HugePages_Total gauge
node_memory_HugePages_Total 0
`
	assertText(t, render(t, collectNodeMemory), want)
}

func TestNodePressureWaitingAndStalled(t *testing.T) {
	proc, _ := useNodeRoots(t)
	// A kernel before 5.13 writes no full line for cpu.
	writeFixture(t, proc, "pressure/cpu", "some avg10=0.00 avg60=0.00 avg300=0.00 total=2000000\n")
	writeFixture(t, proc, "pressure/memory",
		"some avg10=0.00 avg60=0.00 avg300=0.00 total=1500000\n"+
			"full avg10=0.00 avg60=0.00 avg300=0.00 total=250000\n")

	want := `# HELP node_pressure_cpu_waiting_seconds_total Total time in seconds that processes have waited for CPU time
# TYPE node_pressure_cpu_waiting_seconds_total counter
node_pressure_cpu_waiting_seconds_total 2
# HELP node_pressure_memory_waiting_seconds_total Total time in seconds that processes have waited for memory
# TYPE node_pressure_memory_waiting_seconds_total counter
node_pressure_memory_waiting_seconds_total 1.5
# HELP node_pressure_memory_stalled_seconds_total Total time in seconds no process could make progress due to memory congestion
# TYPE node_pressure_memory_stalled_seconds_total counter
node_pressure_memory_stalled_seconds_total 0.25
`
	assertText(t, render(t, collectNodePressure), want)
}

func TestNodeFilesystemReportsRealMounts(t *testing.T) {
	proc, _ := useNodeRoots(t)
	writeFixture(t, proc, "1/mounts", `/dev/root / ext4 rw,relatime 0 0
devtmpfs /dev devtmpfs rw,relatime,size=78000k 0 0
proc /proc proc rw,relatime 0 0
sysfs /sys sysfs rw,relatime 0 0
cgroup2 /sys/fs/cgroup cgroup2 rw,relatime 0 0
tmpfs /tmp tmpfs rw,relatime 0 0
tmpfs /run tmpfs rw,nosuid,nodev 0 0
/dev/mmcblk0p1 /boot vfat ro,relatime 0 0
/dev/mmcblk0p3 /data ext4 rw,noatime 0 0
/dev/sda1 /mnt/usb\040disk exfat rw 0 0
/dev/sdb1 /mnt/gone ext4 rw 0 0
/dev/mmcblk0p4 /data ext4 rw,noatime 0 0
`)
	stats := map[string]fsStat{
		"/":             {blockSize: 4096, blocks: 262144, free: 131072, avail: 120000, files: 65536, filesFree: 60000},
		"/boot":         {blockSize: 512, blocks: 32768, free: 16384, avail: 16384, files: 0, filesFree: 0},
		"/data":         {blockSize: 4096, blocks: 1000000, free: 900000, avail: 850000, files: 250000, filesFree: 249000},
		"/mnt/usb disk": {blockSize: 4096, blocks: 10, free: 5, avail: 5, files: 0, filesFree: 0},
	}
	setVar(t, &statFS, func(path string) (fsStat, error) {
		s, ok := stats[path]
		if !ok {
			return fsStat{}, errors.New("no such mount")
		}
		return s, nil
	})

	root := `{device="/dev/root",fstype="ext4",mountpoint="/"}`
	boot := `{device="/dev/mmcblk0p1",fstype="vfat",mountpoint="/boot"}`
	data := `{device="/dev/mmcblk0p4",fstype="ext4",mountpoint="/data"}`
	usb := `{device="/dev/sda1",fstype="exfat",mountpoint="/mnt/usb disk"}`
	want := "# HELP node_filesystem_size_bytes Filesystem size in bytes.\n" +
		"# TYPE node_filesystem_size_bytes gauge\n" +
		"node_filesystem_size_bytes" + root + " 1073741824\n" +
		"node_filesystem_size_bytes" + boot + " 16777216\n" +
		"node_filesystem_size_bytes" + data + " 4096000000\n" +
		"node_filesystem_size_bytes" + usb + " 40960\n" +
		"# HELP node_filesystem_free_bytes Filesystem free space in bytes.\n" +
		"# TYPE node_filesystem_free_bytes gauge\n" +
		"node_filesystem_free_bytes" + root + " 536870912\n" +
		"node_filesystem_free_bytes" + boot + " 8388608\n" +
		"node_filesystem_free_bytes" + data + " 3686400000\n" +
		"node_filesystem_free_bytes" + usb + " 20480\n" +
		"# HELP node_filesystem_avail_bytes Filesystem space available to non-root users in bytes.\n" +
		"# TYPE node_filesystem_avail_bytes gauge\n" +
		"node_filesystem_avail_bytes" + root + " 491520000\n" +
		"node_filesystem_avail_bytes" + boot + " 8388608\n" +
		"node_filesystem_avail_bytes" + data + " 3481600000\n" +
		"node_filesystem_avail_bytes" + usb + " 20480\n" +
		"# HELP node_filesystem_files Filesystem total file nodes.\n" +
		"# TYPE node_filesystem_files gauge\n" +
		"node_filesystem_files" + root + " 65536\n" +
		"node_filesystem_files" + boot + " 0\n" +
		"node_filesystem_files" + data + " 250000\n" +
		"node_filesystem_files" + usb + " 0\n" +
		"# HELP node_filesystem_files_free Filesystem total free file nodes.\n" +
		"# TYPE node_filesystem_files_free gauge\n" +
		"node_filesystem_files_free" + root + " 60000\n" +
		"node_filesystem_files_free" + boot + " 0\n" +
		"node_filesystem_files_free" + data + " 249000\n" +
		"node_filesystem_files_free" + usb + " 0\n" +
		"# HELP node_filesystem_readonly Filesystem read-only status.\n" +
		"# TYPE node_filesystem_readonly gauge\n" +
		"node_filesystem_readonly" + root + " 0\n" +
		"node_filesystem_readonly" + boot + " 1\n" +
		"node_filesystem_readonly" + data + " 0\n" +
		"node_filesystem_readonly" + usb + " 0\n"
	assertText(t, render(t, collectNodeFilesystem), want)
}

// Without /proc/1/mounts, which only root reads on some kernels, the server's
// own mount table answers.
func TestNodeFilesystemFallsBackToSelfMounts(t *testing.T) {
	proc, _ := useNodeRoots(t)
	writeFixture(t, proc, "self/mounts", "/dev/root / ext4 rw 0 0\n")
	setVar(t, &statFS, func(string) (fsStat, error) { return fsStat{blockSize: 1, blocks: 7}, nil })

	assertLines(t, render(t, collectNodeFilesystem),
		[]string{`node_filesystem_size_bytes{device="/dev/root",fstype="ext4",mountpoint="/"} 7`}, nil)
}

const netDevFixture = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:    5000      50    0    0    0     0          0         0     5000      50    0    0    0     0       0          0
  eth0: 1234567    8901    1    2    3     4          5         6   765432    1098    7    8    9    10      11         12
  usb0:     100       2    0    0    0     0          0         0      200       3    0    0    0     0       0          0
`

func TestNodeNetworkReadsNetDevAndOperstate(t *testing.T) {
	proc, sys := useNodeRoots(t)
	writeFixture(t, proc, "net/dev", netDevFixture)
	writeFixture(t, sys, "class/net/eth0/operstate", "up\n")
	writeFixture(t, sys, "class/net/usb0/operstate", "down\n")

	got := render(t, collectNodeNetwork)
	assertLines(t, got, []string{
		"# TYPE node_network_receive_bytes_total counter",
		`node_network_receive_bytes_total{device="eth0"} 1234567`,
		`node_network_receive_bytes_total{device="usb0"} 100`,
		`node_network_receive_packets_total{device="eth0"} 8901`,
		`node_network_receive_errs_total{device="eth0"} 1`,
		`node_network_receive_drop_total{device="eth0"} 2`,
		`node_network_receive_multicast_total{device="eth0"} 6`,
		`node_network_transmit_bytes_total{device="eth0"} 765432`,
		`node_network_transmit_packets_total{device="eth0"} 1098`,
		`node_network_transmit_errs_total{device="eth0"} 7`,
		`node_network_transmit_drop_total{device="eth0"} 8`,
		`node_network_transmit_colls_total{device="eth0"} 10`,
		`node_network_transmit_compressed_total{device="eth0"} 12`,
		`node_network_transmit_bytes_total{device="usb0"} 200`,
		"# TYPE node_network_up gauge",
		`node_network_up{device="eth0"} 1`,
		`node_network_up{device="usb0"} 0`,
	}, []string{`device="lo"`})

	if n := strings.Count(got, "# TYPE "); n != 17 {
		t.Errorf("%d families, want 16 counters and node_network_up", n)
	}
}

const diskstatsFixture = `   1       0 ram0 0 0 0 0 0 0 0 0 0 0 0
   7       0 loop0 10 0 20 1 0 0 0 0 0 1 1
 179       0 mmcblk0 4000 100 800000 12345 2000 300 64000 54321 1 30000 66666 0 0 0 0
 179       1 mmcblk0p1 40 0 8000 100 0 0 0 0 0 90 100
 252       0 zram0 500 0 4000 5 900 0 7200 10 0 20 15
   8       0 sda 7 0 56 3 0 0 0 0 0 3 3
   8       1 sda1 5 0 40 2 0 0 0 0 0 2 2
`

func TestNodeDiskFollowsTheDefaultExcludes(t *testing.T) {
	proc, _ := useNodeRoots(t)
	writeFixture(t, proc, "diskstats", diskstatsFixture)

	got := render(t, collectNodeDisk)
	assertLines(t, got, []string{
		"# TYPE node_disk_reads_completed_total counter",
		`node_disk_reads_completed_total{device="mmcblk0"} 4000`,
		`node_disk_reads_merged_total{device="mmcblk0"} 100`,
		`node_disk_read_bytes_total{device="mmcblk0"} 409600000`,
		`node_disk_read_time_seconds_total{device="mmcblk0"} 12.345`,
		`node_disk_writes_completed_total{device="mmcblk0"} 2000`,
		`node_disk_writes_merged_total{device="mmcblk0"} 300`,
		`node_disk_written_bytes_total{device="mmcblk0"} 32768000`,
		`node_disk_write_time_seconds_total{device="mmcblk0"} 54.321`,
		"# TYPE node_disk_io_now gauge",
		`node_disk_io_now{device="mmcblk0"} 1`,
		`node_disk_io_time_seconds_total{device="mmcblk0"} 30`,
		`node_disk_io_time_weighted_seconds_total{device="mmcblk0"} 66.666`,
		// node_exporter's default keeps mmcblk partitions and whole sd disks.
		`node_disk_read_bytes_total{device="mmcblk0p1"} 4096000`,
		`node_disk_io_time_seconds_total{device="sda"} 0.003`,
	}, []string{`"ram0"`, `"loop0"`, `"zram0"`, `"sda1"`})
}

func TestNodeThermalReadsHwmonAndZones(t *testing.T) {
	_, sys := useNodeRoots(t)
	writeFixture(t, sys, "class/hwmon/hwmon0/name", "cv180x_thermal\n")
	writeFixture(t, sys, "class/hwmon/hwmon0/temp1_input", "45500\n")
	writeFixture(t, sys, "class/hwmon/hwmon1/temp1_input", "-2000\n")
	writeFixture(t, sys, "class/thermal/thermal_zone0/type", "soc_thermal_0\n")
	writeFixture(t, sys, "class/thermal/thermal_zone0/temp", "46000\n")
	writeFixture(t, sys, "class/thermal/thermal_zone10/type", "late\n")
	writeFixture(t, sys, "class/thermal/thermal_zone10/temp", "30125\n")
	writeFixture(t, sys, "class/thermal/thermal_zone2/type", "off\n")

	want := `# HELP node_hwmon_chip_names Annotation metric for human-readable chip names
# TYPE node_hwmon_chip_names gauge
node_hwmon_chip_names{chip="cv180x_thermal",chip_name="cv180x_thermal"} 1
# HELP node_hwmon_temp_celsius Hardware monitor for temperature (input)
# TYPE node_hwmon_temp_celsius gauge
node_hwmon_temp_celsius{chip="cv180x_thermal",sensor="temp1"} 45.5
node_hwmon_temp_celsius{chip="hwmon1",sensor="temp1"} -2
# HELP node_thermal_zone_temp Zone temperature in Celsius
# TYPE node_thermal_zone_temp gauge
node_thermal_zone_temp{type="soc_thermal_0",zone="0"} 46
node_thermal_zone_temp{type="late",zone="10"} 30.125
`
	assertText(t, render(t, collectNodeThermal), want)
}

// node_exporter names a chip after the device its hwmon links to, and a
// dashboard's chip label depends on getting that the same.
func TestNodeHwmonChipFromDeviceLink(t *testing.T) {
	_, sys := useNodeRoots(t)
	device := filepath.Join(sys, "devices", "platform", "cv180x-thermal.0")
	if err := os.MkdirAll(device, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFixture(t, sys, "class/hwmon/hwmon0/name", "soc\n")
	if err := os.Symlink(device, filepath.Join(sys, "class", "hwmon", "hwmon0", "device")); err != nil {
		t.Skipf("this host cannot make a symlink: %s", err)
	}

	if got := hwmonChip(filepath.Join(sys, "class", "hwmon", "hwmon0")); got != "platform_cv180x_thermal_0" {
		t.Fatalf("chip %q, want platform_cv180x_thermal_0", got)
	}
}

// With no /proc and no /sys, every node section but the clock writes nothing.
func TestNodeSectionsOnABareHost(t *testing.T) {
	useNodeRoots(t)

	for _, collect := range []func(*Writer){
		collectNodeMemory, collectNodePressure, collectNodeFilesystem,
		collectNodeNetwork, collectNodeDisk, collectNodeThermal,
	} {
		if got := render(t, collect); got != "" {
			t.Errorf("wrote without a source:\n%s", got)
		}
	}

	got := render(t, collectNodeSystem)
	if !strings.HasPrefix(got, "# HELP node_time_seconds ") || strings.Count(got, "# TYPE ") != 1 {
		t.Errorf("want node_time_seconds alone, got:\n%s", got)
	}
}
