package metrics

import (
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

const milliPerUnit = 1000

// collectNodeThermal writes node_exporter's hwmon temperatures and thermal
// zones. A board without either writes nothing.
func collectNodeThermal(w *Writer) {
	writeHwmon(w)
	writeThermalZones(w)
}

type hwmonTemp struct {
	chip, sensor string
	celsius      float64
}

func writeHwmon(w *Writer) {
	dirs, _ := filepath.Glob(filepath.Join(sysDir, "class", "hwmon", "hwmon*"))
	sort.Strings(dirs)

	type chip struct{ chip, name string }
	var chips []chip
	var temps []hwmonTemp
	for _, dir := range dirs {
		id := hwmonChip(dir)
		if name, ok := readLine(filepath.Join(dir, "name")); ok && name != "" {
			chips = append(chips, chip{id, name})
		}

		inputs, _ := filepath.Glob(filepath.Join(dir, "temp*_input"))
		sort.Strings(inputs)
		for _, input := range inputs {
			raw, ok := readLine(input)
			if !ok {
				continue
			}
			millis, err := strconv.ParseFloat(raw, 64)
			if err != nil {
				continue
			}
			sensor := strings.TrimSuffix(filepath.Base(input), "_input")
			temps = append(temps, hwmonTemp{chip: id, sensor: sensor, celsius: millis / milliPerUnit})
		}
	}

	for _, c := range chips {
		w.Gauge("node_hwmon_chip_names", "Annotation metric for human-readable chip names", 1,
			L("chip", c.chip), L("chip_name", c.name))
	}
	for _, t := range temps {
		w.Gauge("node_hwmon_temp_celsius", "Hardware monitor for temperature (input)", t.celsius,
			L("chip", t.chip), L("sensor", t.sensor))
	}
}

// hwmonChip names a hwmon directory the way node_exporter does, so the chip
// label is the same on both: the bus and device the hwmon link points at, then
// the name file, then the directory itself.
func hwmonChip(dir string) string {
	if device, err := filepath.EvalSymlinks(filepath.Join(dir, "device")); err == nil {
		parent, name := filepath.Split(device)
		bus := filepath.Base(strings.TrimRight(parent, "/"))
		cleanName, cleanBus := cleanMetricName(name), cleanMetricName(bus)
		if cleanBus != "" && cleanName != "" {
			return cleanBus + "_" + cleanName
		}
		if cleanName != "" {
			return cleanName
		}
	}

	if name, ok := readLine(filepath.Join(dir, "name")); ok {
		if clean := cleanMetricName(name); clean != "" {
			return clean
		}
	}

	if real, err := filepath.EvalSymlinks(dir); err == nil {
		dir = real
	}

	return cleanMetricName(filepath.Base(dir))
}

var metricNameUnsafe = regexp.MustCompile(`[^a-z0-9:_]`)

// cleanMetricName is node_exporter's hwmon cleanMetricName.
func cleanMetricName(name string) string {
	return strings.Trim(metricNameUnsafe.ReplaceAllString(strings.ToLower(name), "_"), "_")
}

// writeThermalZones writes node_thermal_zone_temp for each zone that answers.
// A zone whose sensor is off fails the read, and node_exporter leaves it out
// too.
func writeThermalZones(w *Writer) {
	dirs, _ := filepath.Glob(filepath.Join(sysDir, "class", "thermal", "thermal_zone*"))
	sort.Slice(dirs, func(i, j int) bool { return zoneNumber(dirs[i]) < zoneNumber(dirs[j]) })

	for _, dir := range dirs {
		zone := strings.TrimPrefix(filepath.Base(dir), "thermal_zone")
		kind, ok := readLine(filepath.Join(dir, "type"))
		if !ok {
			continue
		}
		raw, ok := readLine(filepath.Join(dir, "temp"))
		if !ok {
			continue
		}
		millis, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			continue
		}
		w.Gauge("node_thermal_zone_temp", "Zone temperature in Celsius", millis/milliPerUnit,
			L("type", kind), L("zone", zone))
	}
}

// zoneNumber orders thermal_zone10 after thermal_zone9.
func zoneNumber(dir string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(dir), "thermal_zone"))
	if err != nil {
		return int(^uint(0) >> 1)
	}

	return n
}
