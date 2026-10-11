// Package dmapool reports the memory the mainline kernel sets aside for
// devices: the reserved pools the video drivers allocate from, and the CMA
// areas.
//
// The vendor kernel has neither. It hands the video pipeline an ION carveout,
// which package ion reads. Which kernel is running is told by what it exposes,
// never by its version string: a caller asks ion.Present first and comes here
// only when the carveout is absent.
//
// On the mainline build the capture, the Coda980 and the WAVE420L share one
// no-map shared-dma-pool, video-pool@86800000. Patch 0939 publishes its use in
// debugfs under dma_coherent/. A kernel without that patch still describes the
// pool in the device tree, so its size is known even when its use is not. A
// pool made reusable is a CMA area instead, and its use comes from the CMA
// files.
//
// libkvm-v4l2 allocates the buffers between the scaler and the encoders from
// a dma-heap, /dev/dma_heap/default_cma_region first, which is the default CMA
// area. CmaFree in /proc/meminfo understates what is left of it, because
// movable pages borrow free CMA pages and still count as taken there. The
// debugfs counters count only what cma_alloc handed out.
package dmapool

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// The roots every source is read from. Variables so tests point them at
// fixtures.
var (
	// CoherentDir holds one file per no-map pool (patch 0939).
	CoherentDir = "/sys/kernel/debug/dma_coherent"
	// CMADebugDir holds one directory per CMA area (CONFIG_CMA_DEBUGFS).
	CMADebugDir = "/sys/kernel/debug/cma"
	// CMASysDir holds one directory per CMA area (CONFIG_CMA_SYSFS).
	CMASysDir = "/sys/kernel/mm/cma"
	// ReservedMemDir is the device tree's reserved-memory node.
	ReservedMemDir = "/sys/firmware/devicetree/base/reserved-memory"
)

// pageSize converts the CMA files' page counts. A variable for the tests.
var pageSize = uint64(os.Getpagesize())

// Pool is one reserved pool a device allocates from. Sizes are in bytes.
type Pool struct {
	Name string
	Size uint64
	// Measured is false when only the size is known: the kernel describes the
	// pool but publishes nothing about its use. Every other field is 0 then.
	Measured    bool
	Used        uint64
	Free        uint64
	LargestFree uint64
	// LargestFreeKnown is false for a CMA-backed pool read from sysfs only.
	LargestFreeKnown bool
	// Peak is the most the pool has held since boot or since the last write to
	// its debugfs file. A CMA-backed pool keeps no peak, and reports 0.
	Peak   uint64
	Failed uint64
}

// Area is one CMA area that is not a device's own pool. Sizes are in bytes.
type Area struct {
	Name  string
	Total uint64
	Used  uint64
	// LargestFree is the longest free run, known only from debugfs.
	LargestFree      uint64
	LargestFreeKnown bool
	// Failed is alloc_pages_fail, in pages, from sysfs; 0 without it.
	Failed uint64
}

// Snapshot is one reading of every pool and area, each sorted by name.
type Snapshot struct {
	Pools []Pool
	Areas []Area
}

// Read takes one reading. It never fails: a source it cannot read is left out,
// and a kernel with none of them gives an empty snapshot.
func Read() Snapshot {
	var snap Snapshot

	nodes := readReservedNodes()
	areas := readAreas()
	seen := make(map[string]bool)

	for _, name := range listNames(CoherentDir, false) {
		body, err := os.ReadFile(filepath.Join(CoherentDir, name))
		if err != nil {
			continue
		}
		pool, ok := ParseCoherent(string(body))
		if !ok {
			continue
		}
		pool.Name = name
		snap.Pools = append(snap.Pools, pool)
		seen[name] = true
	}

	// A CMA area named after a reserved-memory node that is not the default
	// area is that node's pool, made reusable.
	for _, a := range areas {
		node, ok := nodes[a.Name]
		if ok && node.reusable && !node.cmaDefault && !seen[a.Name] {
			pool := Pool{
				Name:     a.Name,
				Size:     a.Total,
				Measured: true,
				Used:     a.Used,
				Failed:   a.Failed,
			}
			if a.Used <= a.Total {
				pool.Free = a.Total - a.Used
			}
			pool.LargestFree, pool.LargestFreeKnown = a.LargestFree, a.LargestFreeKnown
			snap.Pools = append(snap.Pools, pool)
			seen[a.Name] = true
			continue
		}
		snap.Areas = append(snap.Areas, a)
	}

	// A no-map pool the kernel does not measure still has a size.
	for name, node := range nodes {
		if seen[name] || !node.devicePool || node.size == 0 {
			continue
		}
		snap.Pools = append(snap.Pools, Pool{Name: name, Size: node.size})
	}

	sort.Slice(snap.Pools, func(i, j int) bool { return snap.Pools[i].Name < snap.Pools[j].Name })
	return snap
}

// ParseCoherent reads one dma_coherent file, whose sizes are in KiB:
//
//	size_kib 28672
//	used_kib 17440
//	free_kib 11232
//	largest_free_kib 8192
//	peak_used_kib 24388
//	failed 0
//
// ok is false when the size is missing or 0, which no real pool has.
func ParseCoherent(s string) (Pool, bool) {
	values := make(map[string]uint64)
	for _, line := range strings.Split(s, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 2 {
			continue
		}
		v, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			continue
		}
		values[fields[0]] = v
	}

	size := values["size_kib"] * 1024
	if size == 0 {
		return Pool{}, false
	}
	pool := Pool{
		Size:     size,
		Measured: true,
		Used:     values["used_kib"] * 1024,
		Peak:     values["peak_used_kib"] * 1024,
		Failed:   values["failed"],
	}
	if largest, ok := values["largest_free_kib"]; ok {
		pool.LargestFree, pool.LargestFreeKnown = largest*1024, true
	}
	if free, ok := values["free_kib"]; ok {
		pool.Free = free * 1024
	} else if pool.Used <= size {
		pool.Free = size - pool.Used
	}
	return pool, true
}

// readAreas reads every CMA area, from debugfs where it can and from sysfs
// where it cannot, sorted by name.
func readAreas() []Area {
	names := make(map[string]bool)
	for _, n := range listNames(CMADebugDir, true) {
		names[n] = true
	}
	for _, n := range listNames(CMASysDir, true) {
		names[n] = true
	}

	var out []Area
	for name := range names {
		if a, ok := readArea(name); ok {
			out = append(out, a)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func readArea(name string) (Area, bool) {
	area := Area{Name: name}
	debug := filepath.Join(CMADebugDir, name)
	sys := filepath.Join(CMASysDir, name)

	if count, ok := readUint(filepath.Join(debug, "count")); ok && count > 0 {
		used, okUsed := readUint(filepath.Join(debug, "used"))
		if !okUsed {
			return Area{}, false
		}
		area.Total = count * pageSize
		area.Used = used * pageSize
		if chunk, ok := readUint(filepath.Join(debug, "maxchunk")); ok {
			area.LargestFree = chunk * pageSize
			area.LargestFreeKnown = true
		}
	} else {
		total, okTotal := readUint(filepath.Join(sys, "total_pages"))
		avail, okAvail := readUint(filepath.Join(sys, "available_pages"))
		if !okTotal || !okAvail || total == 0 || avail > total {
			return Area{}, false
		}
		area.Total = total * pageSize
		area.Used = (total - avail) * pageSize
	}

	if failed, ok := readUint(filepath.Join(sys, "alloc_pages_fail")); ok {
		area.Failed = failed
	}
	return area, true
}

// reservedNode is what the reserved-memory node says about one child.
type reservedNode struct {
	size uint64
	// devicePool is a no-map shared-dma-pool: a device's own coherent pool.
	devicePool bool
	reusable   bool
	cmaDefault bool
}

func readReservedNodes() map[string]reservedNode {
	out := make(map[string]reservedNode)

	sizeCells := uint32(1)
	if b, err := os.ReadFile(filepath.Join(ReservedMemDir, "#size-cells")); err == nil && len(b) == 4 {
		sizeCells = binary.BigEndian.Uint32(b)
	}
	addrCells := uint32(1)
	if b, err := os.ReadFile(filepath.Join(ReservedMemDir, "#address-cells")); err == nil && len(b) == 4 {
		addrCells = binary.BigEndian.Uint32(b)
	}

	for _, name := range listNames(ReservedMemDir, true) {
		dir := filepath.Join(ReservedMemDir, name)
		node := reservedNode{
			reusable:   exists(filepath.Join(dir, "reusable")),
			cmaDefault: exists(filepath.Join(dir, "linux,cma-default")),
		}
		if b, err := os.ReadFile(filepath.Join(dir, "reg")); err == nil {
			node.size = ParseRegSize(b, addrCells, sizeCells)
		}
		if b, err := os.ReadFile(filepath.Join(dir, "compatible")); err == nil {
			node.devicePool = hasString(b, "shared-dma-pool") &&
				exists(filepath.Join(dir, "no-map")) && !node.reusable
		}
		out[name] = node
	}
	return out
}

// ParseRegSize sums the sizes of every (address, size) pair in a reg property.
// It returns 0 for a property whose length does not fit the cell counts, or
// for cell counts other than 1 or 2.
func ParseRegSize(reg []byte, addrCells, sizeCells uint32) uint64 {
	if sizeCells < 1 || sizeCells > 2 || addrCells > 2 {
		return 0
	}
	entry := int(addrCells+sizeCells) * 4
	if entry == 0 || len(reg) == 0 || len(reg)%entry != 0 {
		return 0
	}

	var total uint64
	for off := 0; off < len(reg); off += entry {
		cells := reg[off+int(addrCells)*4 : off+entry]
		var size uint64
		for i := 0; i < len(cells); i += 4 {
			size = size<<32 | uint64(binary.BigEndian.Uint32(cells[i:i+4]))
		}
		total += size
	}
	return total
}

// hasString reports whether a NUL-separated string list property holds s.
func hasString(prop []byte, s string) bool {
	for _, item := range strings.Split(string(prop), "\x00") {
		if item == s {
			return true
		}
	}
	return false
}

// listNames lists a directory's entries, only the subdirectories when dirs is
// set, and nothing when the directory cannot be read.
func listNames(dir string, dirs bool) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var out []string
	for _, e := range entries {
		if e.IsDir() != dirs {
			continue
		}
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out
}

func readUint(path string) (uint64, bool) {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	v, err := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	return v, err == nil
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
