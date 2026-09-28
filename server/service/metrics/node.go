package metrics

import (
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// The node_* families repeat what node_exporter's default collectors report,
// under the same names and labels, so dashboards built for node_exporter read
// this endpoint unchanged. Only a subset is here: what a single-core board
// with one SD card needs. Everything is read from /proc and /sys on the
// scrape; nothing runs between scrapes.

// sysDir is the root the /sys sources are read from. A variable so tests point
// it at fixtures.
var sysDir = "/sys"

// nodeNow is the clock node_time_seconds reads. A variable so tests fix it.
var nodeNow = time.Now

// userHZ is the kernel's USER_HZ, the unit of the /proc/stat CPU times. It is
// 100 on every architecture Linux supports, and node_exporter assumes the same.
const userHZ = 100

// cpuModes are the /proc/stat cpuN columns, in the kernel's order. The guest
// columns that follow are left out, as node_exporter's node_cpu_seconds_total
// leaves them out.
var cpuModes = []string{"user", "nice", "system", "idle", "iowait", "irq", "softirq", "steal"}

func collectNodeSystem(w *Writer) {
	writeUname(w)
	w.Gauge("node_time_seconds", "System time in seconds since epoch (1970).",
		float64(nodeNow().UnixNano())/1e9)
	writeStat(w)
	writeLoadavg(w)
}

// writeUname reports what uname(2) reports, from /proc/sys/kernel. The labels
// are in node_exporter's order, which is alphabetical.
func writeUname(w *Writer) {
	kernel := filepath.Join(procDir, "sys", "kernel")
	sysname, ok := readLine(filepath.Join(kernel, "ostype"))
	if !ok {
		return
	}
	release, _ := readLine(filepath.Join(kernel, "osrelease"))
	version, _ := readLine(filepath.Join(kernel, "version"))
	nodename, _ := readLine(filepath.Join(kernel, "hostname"))
	domainname, _ := readLine(filepath.Join(kernel, "domainname"))

	w.Gauge("node_uname_info", "Labeled system information as provided by the uname system call.", 1,
		L("domainname", domainname), L("machine", unameMachine()), L("nodename", nodename),
		L("release", release), L("sysname", sysname), L("version", version))
}

// unameMachine returns uname's machine field. /proc/sys/kernel/arch has it on
// kernels from 6.1; older ones do not, and the GOARCH this binary was built for
// is the same answer under a Go name.
func unameMachine() string {
	if arch, ok := readLine(filepath.Join(procDir, "sys", "kernel", "arch")); ok && arch != "" {
		return arch
	}

	switch runtime.GOARCH {
	case "amd64":
		return "x86_64"
	case "arm64":
		return "aarch64"
	case "386":
		return "i686"
	default:
		return runtime.GOARCH
	}
}

// writeStat reports the CPU times and the system-wide counters of /proc/stat.
func writeStat(w *Writer) {
	body, ok := readText(filepath.Join(procDir, "stat"))
	if !ok {
		return
	}

	type cpu struct {
		name  string
		ticks []uint64
	}
	var cpus []cpu
	scalars := make(map[string]uint64)

	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		key := fields[0]

		if index, found := strings.CutPrefix(key, "cpu"); found && index != "" {
			c := cpu{name: index}
			for _, raw := range fields[1:min(len(fields), 1+len(cpuModes))] {
				value, err := strconv.ParseUint(raw, 10, 64)
				if err != nil {
					break
				}
				c.ticks = append(c.ticks, value)
			}
			cpus = append(cpus, c)
			continue
		}

		// intr and softirq carry a per-source list after the total.
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err == nil {
			scalars[key] = value
		}
	}

	for _, c := range cpus {
		for i, ticks := range c.ticks {
			w.Counter("node_cpu_seconds_total", "Seconds the CPUs spent in each mode.",
				float64(ticks)/userHZ, L("cpu", c.name), L("mode", cpuModes[i]))
		}
	}

	stats := []struct {
		key, name, help string
		counter         bool
	}{
		{"btime", "node_boot_time_seconds", "Node boot time, in unixtime.", false},
		{"ctxt", "node_context_switches_total", "Total number of context switches.", true},
		{"intr", "node_intr_total", "Total number of interrupts serviced.", true},
		{"processes", "node_forks_total", "Total number of forks.", true},
		{"procs_running", "node_procs_running", "Number of processes in runnable state.", false},
		{"procs_blocked", "node_procs_blocked", "Number of processes blocked waiting for I/O to complete.", false},
	}
	for _, s := range stats {
		value, ok := scalars[s.key]
		if !ok {
			continue
		}
		if s.counter {
			w.Counter(s.name, s.help, float64(value))
		} else {
			w.Gauge(s.name, s.help, float64(value))
		}
	}
}

func writeLoadavg(w *Writer) {
	body, ok := readText(filepath.Join(procDir, "loadavg"))
	if !ok {
		return
	}

	fields := strings.Fields(body)
	for i, name := range []string{"node_load1", "node_load5", "node_load15"} {
		if i >= len(fields) {
			return
		}
		value, err := strconv.ParseFloat(fields[i], 64)
		if err != nil {
			return
		}
		minutes := strings.TrimPrefix(name, "node_load")
		w.Gauge(name, minutes+"m load average.", value)
	}
}

// meminfoParens matches the "(anon)" of "Active(anon)", which node_exporter
// writes as "_anon".
var meminfoParens = regexp.MustCompile(`\((.*)\)`)

// collectNodeMemory writes every /proc/meminfo field as its own family, named
// the way node_exporter names it: node_memory_MemAvailable_bytes, and without
// the _bytes for a field that is a count, such as HugePages_Total.
func collectNodeMemory(w *Writer) {
	body, ok := readText(filepath.Join(procDir, "meminfo"))
	if !ok {
		return
	}

	written := make(map[string]bool)
	for _, line := range strings.Split(body, "\n") {
		key, rest, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		value, err := strconv.ParseFloat(fields[0], 64)
		if err != nil {
			continue
		}

		name := "node_memory_" + meminfoParens.ReplaceAllString(strings.TrimSpace(key), "_${1}")
		help := "Memory information field " + strings.TrimPrefix(name, "node_memory_")
		if len(fields) > 1 && fields[1] == "kB" {
			value *= bytesPerKilobyte
			name += "_bytes"
			help += "_bytes"
		}
		help += "."

		// A repeated field would split its family, and the writer would drop
		// the whole section for it.
		if written[name] {
			continue
		}
		written[name] = true
		w.Gauge(name, help, value)
	}
}

// collectNodePressure writes node_exporter's PSI families: "waiting" is the
// some line and "stalled" the full line. A kernel without PSI has none of the
// files, and the section writes nothing.
func collectNodePressure(w *Writer) {
	resources := []struct{ name, waiting, stalled string }{
		{"cpu", "Total time in seconds that processes have waited for CPU time",
			"Total time in seconds no process could make progress due to CPU congestion"},
		{"io", "Total time in seconds that processes have waited due to IO congestion",
			"Total time in seconds no process could make progress due to IO congestion"},
		{"memory", "Total time in seconds that processes have waited for memory",
			"Total time in seconds no process could make progress due to memory congestion"},
	}

	for _, r := range resources {
		body, ok := readText(filepath.Join(procDir, "pressure", r.name))
		if !ok {
			continue
		}
		if micros, ok := parsePressureTotal(body, "some"); ok {
			w.Counter("node_pressure_"+r.name+"_waiting_seconds_total", r.waiting,
				float64(micros)/microsPerSecond)
		}
		// Before 5.13 the cpu file has no full line, and the family is left out.
		if micros, ok := parsePressureTotal(body, "full"); ok {
			w.Counter("node_pressure_"+r.name+"_stalled_seconds_total", r.stalled,
				float64(micros)/microsPerSecond)
		}
	}
}

// readLine returns the first line of a file without its newline.
func readLine(path string) (string, bool) {
	body, ok := readText(path)
	if !ok {
		return "", false
	}
	line, _, _ := strings.Cut(body, "\n")

	return strings.TrimSpace(line), true
}
