package metrics

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"NanoKVM-Server/service/vm"
)

// The roots every memory source is read from. Variables so tests point them
// at fixtures.
var (
	procDir      = "/proc"
	zramDir      = "/sys/block/zram0"
	cgroupDir    = "/sys/fs/cgroup"
	cgroupGroups = []string{"kvm", "addons"}
)

// meminfoKinds are the /proc/meminfo fields reported, in output order.
var meminfoKinds = []string{"MemTotal", "MemAvailable", "MemFree", "SwapTotal", "SwapFree", "CmaTotal", "CmaFree"}

// cgroupEvents are the memory.events keys reported, in output order.
var cgroupEvents = []string{"high", "max", "oom", "oom_kill"}

const (
	helpMemory       = "Memory from /proc/meminfo, by field."
	helpSwap         = "Pages swapped in and out, from /proc/vmstat."
	helpZram         = "zram0 sizes, from mm_stat."
	helpPressure     = "Time tasks stalled waiting for a resource, from /proc/pressure."
	helpCgroupMemory = "Memory charged to a cgroup, from memory.current."
	helpCgroupEvents = "Memory events in a cgroup, from memory.events."
	microsPerSecond  = 1e6
	bytesPerKilobyte = 1024
)

func collectMemory(w *Writer) {
	writeMeminfo(w)
	writeSwap(w)
	writeZram(w)
	writePressure(w)
	writeCgroups(w)
}

func writeMeminfo(w *Writer) {
	body, ok := readText(filepath.Join(procDir, "meminfo"))
	if !ok {
		return
	}

	values := parseKBFields(body)
	for _, kind := range meminfoKinds {
		value, ok := values[kind]
		if !ok {
			continue
		}
		w.Gauge("ironkvm_memory_bytes", helpMemory, float64(value), L("kind", kind))
	}
}

func writeSwap(w *Writer) {
	body, ok := readText(filepath.Join(procDir, "vmstat"))
	if !ok {
		return
	}

	// A kernel without swap support has neither counter, which is not 0 pages.
	in, out, found := vm.ParseVmstatSwap(body)
	if !found {
		return
	}
	w.Counter("ironkvm_swap_pages_total", helpSwap, float64(in), L("direction", "in"))
	w.Counter("ironkvm_swap_pages_total", helpSwap, float64(out), L("direction", "out"))
}

func writeZram(w *Writer) {
	body, ok := readText(filepath.Join(zramDir, "mm_stat"))
	if !ok {
		return
	}

	// Parsing stops at the first bad field. Only the fields before it are
	// written: a field it never reached reads 0, and 0 is a real size.
	stat := vm.ParseZramMmStat(body)
	kinds := []struct {
		kind  string
		value int64
	}{
		{"orig_data_size", stat.Original},
		{"compr_data_size", stat.Compressed},
		{"mem_used_total", stat.MemUsed},
	}
	for _, k := range kinds[:stat.Fields] {
		w.Gauge("ironkvm_zram_bytes", helpZram, float64(k.value), L("kind", k.kind))
	}
}

// writePressure reports the PSI totals. A kernel booted without psi=1 fails the
// read, and the family is left out.
func writePressure(w *Writer) {
	body, ok := readText(filepath.Join(procDir, "pressure", "memory"))
	if !ok {
		return
	}

	for _, kind := range []string{"some", "full"} {
		micros, ok := parsePressureTotal(body, kind)
		if !ok {
			continue
		}
		w.Counter("ironkvm_pressure_stall_seconds_total", helpPressure,
			float64(micros)/microsPerSecond, L("resource", "memory"), L("kind", kind))
	}
}

// writeCgroups reports each group that exists. The two families are written
// one after the other, so every group's bytes come before any group's events.
func writeCgroups(w *Writer) {
	type group struct {
		name    string
		current uint64
		events  map[string]uint64
	}

	var groups []group
	for _, name := range cgroupGroups {
		body, ok := readText(filepath.Join(cgroupDir, name, "memory.current"))
		if !ok {
			continue
		}
		current, err := strconv.ParseUint(strings.TrimSpace(body), 10, 64)
		if err != nil {
			continue
		}

		g := group{name: name, current: current}
		if body, ok := readText(filepath.Join(cgroupDir, name, "memory.events")); ok {
			g.events = parseKeyValues(body)
		}
		groups = append(groups, g)
	}

	for _, g := range groups {
		w.Gauge("ironkvm_cgroup_memory_bytes", helpCgroupMemory, float64(g.current), L("group", g.name))
	}
	for _, g := range groups {
		for _, event := range cgroupEvents {
			value, ok := g.events[event]
			if !ok {
				continue
			}
			w.Counter("ironkvm_cgroup_memory_events_total", helpCgroupEvents, float64(value),
				L("group", g.name), L("event", event))
		}
	}
}

// readText returns a file's contents, or false when it cannot be read. A
// missing source is a normal board, so the error itself is not interesting.
func readText(path string) (string, bool) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}

	return string(body), true
}

// parseKBFields reads "Name:  value [kB]" lines, the shape of /proc/meminfo and
// /proc/self/status. A value in kB is returned in bytes. A line whose value is
// not a number is skipped.
func parseKBFields(body string) map[string]uint64 {
	values := make(map[string]uint64)
	for _, line := range strings.Split(body, "\n") {
		name, rest, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		fields := strings.Fields(rest)
		if len(fields) == 0 {
			continue
		}
		value, err := strconv.ParseUint(fields[0], 10, 64)
		if err != nil {
			continue
		}
		if len(fields) > 1 && fields[1] == "kB" {
			value *= bytesPerKilobyte
		}
		values[strings.TrimSpace(name)] = value
	}

	return values
}

// parseKeyValues reads "key value" lines, the shape of memory.events.
func parseKeyValues(body string) map[string]uint64 {
	values := make(map[string]uint64)
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		values[fields[0]] = value
	}

	return values
}

// parsePressureTotal finds total= on the line for kind ("some" or "full") of a
// /proc/pressure file. The value is in microseconds.
func parsePressureTotal(body, kind string) (uint64, bool) {
	for _, line := range strings.Split(body, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 || fields[0] != kind {
			continue
		}
		for _, field := range fields[1:] {
			raw, found := strings.CutPrefix(field, "total=")
			if !found {
				continue
			}
			value, err := strconv.ParseUint(raw, 10, 64)
			return value, err == nil
		}
	}

	return 0, false
}
