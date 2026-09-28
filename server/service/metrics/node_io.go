package metrics

import (
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// netDevColumns are the sixteen /proc/net/dev counters, in the kernel's order,
// under node_exporter's family names.
var netDevColumns = []string{
	"receive_bytes", "receive_packets", "receive_errs", "receive_drop",
	"receive_fifo", "receive_frame", "receive_compressed", "receive_multicast",
	"transmit_bytes", "transmit_packets", "transmit_errs", "transmit_drop",
	"transmit_fifo", "transmit_colls", "transmit_carrier", "transmit_compressed",
}

type netDevice struct {
	name     string
	counters []uint64
}

// collectNodeNetwork writes node_exporter's node_network_*_total families from
// /proc/net/dev, and node_network_up from each interface's operstate. The
// loopback interface is left out.
func collectNodeNetwork(w *Writer) {
	devices := readNetDev()

	for i, column := range netDevColumns {
		name := "node_network_" + column + "_total"
		help := "Network device statistic " + column + "."
		for _, d := range devices {
			if i < len(d.counters) {
				w.Counter(name, help, float64(d.counters[i]), L("device", d.name))
			}
		}
	}

	for _, d := range devices {
		state, ok := readLine(filepath.Join(sysDir, "class", "net", d.name, "operstate"))
		if !ok {
			continue
		}
		up := 0.0
		if state == "up" {
			up = 1
		}
		w.Gauge("node_network_up", "Value is 1 if operstate is 'up', 0 otherwise.", up, L("device", d.name))
	}
}

func readNetDev() []netDevice {
	body, ok := readText(filepath.Join(procDir, "net", "dev"))
	if !ok {
		return nil
	}

	var devices []netDevice
	for _, line := range strings.Split(body, "\n") {
		// The two header lines have no colon before the counters.
		name, rest, found := strings.Cut(line, ":")
		if !found {
			continue
		}
		name = strings.TrimSpace(name)
		if name == "lo" || name == "" {
			continue
		}

		d := netDevice{name: name}
		for _, raw := range strings.Fields(rest) {
			value, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				break
			}
			d.counters = append(d.counters, value)
		}
		devices = append(devices, d)
	}

	return devices
}

// diskDeviceExclude is node_exporter's default for
// --collector.diskstats.device-exclude. It leaves the mmcblk partitions in.
var diskDeviceExclude = regexp.MustCompile(`^(z?ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\d+n\d+p)\d+$`)

// diskSectorBytes is the unit of the diskstats sector counts, 512 whatever the
// device's own sector size.
const diskSectorBytes = 512

const millisPerSecond = 1000

// Units of the diskstats columns. Each is converted to node_exporter's unit.
const (
	diskCount = iota
	diskSectors
	diskMillis
)

// diskColumns are the eleven /proc/diskstats counters after the device name.
var diskColumns = []struct {
	name, help string
	unit       int
	counter    bool
}{
	{"node_disk_reads_completed_total", "The total number of reads completed successfully.", diskCount, true},
	{"node_disk_reads_merged_total", "The total number of reads merged.", diskCount, true},
	{"node_disk_read_bytes_total", "The total number of bytes read successfully.", diskSectors, true},
	{"node_disk_read_time_seconds_total", "The total number of seconds spent by all reads.", diskMillis, true},
	{"node_disk_writes_completed_total", "The total number of writes completed successfully.", diskCount, true},
	{"node_disk_writes_merged_total", "The number of writes merged.", diskCount, true},
	{"node_disk_written_bytes_total", "The total number of bytes written successfully.", diskSectors, true},
	{"node_disk_write_time_seconds_total", "This is the total number of seconds spent by all writes.", diskMillis, true},
	{"node_disk_io_now", "The number of I/Os currently in progress.", diskCount, false},
	{"node_disk_io_time_seconds_total", "Total seconds spent doing I/Os.", diskMillis, true},
	{"node_disk_io_time_weighted_seconds_total", "The weighted # of seconds spent doing I/Os.", diskMillis, true},
}

// collectNodeDisk writes node_exporter's node_disk_* families from
// /proc/diskstats.
func collectNodeDisk(w *Writer) {
	body, ok := readText(filepath.Join(procDir, "diskstats"))
	if !ok {
		return
	}

	type disk struct {
		name     string
		counters []uint64
	}
	var disks []disk
	for _, line := range strings.Split(body, "\n") {
		// major minor name, then the counters.
		fields := strings.Fields(line)
		if len(fields) < 3+len(diskColumns) || diskDeviceExclude.MatchString(fields[2]) {
			continue
		}
		d := disk{name: fields[2]}
		for _, raw := range fields[3 : 3+len(diskColumns)] {
			value, err := strconv.ParseUint(raw, 10, 64)
			if err != nil {
				break
			}
			d.counters = append(d.counters, value)
		}
		disks = append(disks, d)
	}

	for i, column := range diskColumns {
		for _, d := range disks {
			if i >= len(d.counters) {
				continue
			}
			value := float64(d.counters[i])
			switch column.unit {
			case diskSectors:
				value *= diskSectorBytes
			case diskMillis:
				value /= millisPerSecond
			}
			if column.counter {
				w.Counter(column.name, column.help, value, L("device", d.name))
			} else {
				w.Gauge(column.name, column.help, value, L("device", d.name))
			}
		}
	}
}
