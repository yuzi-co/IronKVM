package vm

import (
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/dmapool"
	"NanoKVM-Server/service/extensions/addon"
	"NanoKVM-Server/service/extensions/vpn"
	"NanoKVM-Server/service/ion"
)

// useMemory builds on the health fixture, which already moves /proc and the
// VPN daemons, and moves the rest of what GetMemory reads.
func useMemory(t *testing.T) *healthFixture {
	t.Helper()
	f := useHealth(t)

	savedZram, savedSelf, savedGroup := memZramDir, memSelfPID, vpn.CgroupDir
	savedIon := ion.Root
	savedPools := []string{dmapool.CoherentDir, dmapool.CMADebugDir, dmapool.CMASysDir, dmapool.ReservedMemDir}
	t.Cleanup(func() {
		memZramDir, memSelfPID, vpn.CgroupDir = savedZram, savedSelf, savedGroup
		ion.Root = savedIon
		dmapool.CoherentDir, dmapool.CMADebugDir = savedPools[0], savedPools[1]
		dmapool.CMASysDir, dmapool.ReservedMemDir = savedPools[2], savedPools[3]
		kvmSystemPID.pid = 0
	})
	memZramDir = filepath.Join(f.base, "zram0")
	ion.Root = filepath.Join(f.base, "ion")
	dmapool.CoherentDir = filepath.Join(f.base, "dma_coherent")
	dmapool.CMADebugDir = filepath.Join(f.base, "cma")
	dmapool.CMASysDir = filepath.Join(f.base, "mm-cma")
	dmapool.ReservedMemDir = filepath.Join(f.base, "reserved-memory")
	vpn.CgroupDir = filepath.Join(f.base, "addons")
	memSelfPID = func() int { return 100 }
	kvmSystemPID.pid = 0

	return f
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func writeProc(t *testing.T, pid int, comm string, rssKB int) {
	t.Helper()
	dir := filepath.Join(addon.ProcDir, strconv.Itoa(pid))
	writeFile(t, filepath.Join(dir, "comm"), comm+"\n")
	writeFile(t, filepath.Join(dir, "status"),
		"Name:\t"+comm+"\nVmRSS:\t    "+strconv.Itoa(rssKB)+" kB\nThreads:\t4\n")
}

func TestReadMemoryWithNothingToReadIsEmpty(t *testing.T) {
	useMemory(t)

	got := readMemory()
	if got.Total != 0 || len(got.Swaps) != 0 || len(got.Processes) != 0 || got.Addons != nil || got.ZramMemUsed != 0 ||
		got.VideoPools != nil || got.Cma != nil {
		t.Errorf("readMemory() = %+v, want an empty report", got)
	}
}

func TestReadMemoryReportsEverySource(t *testing.T) {
	f := useMemory(t)

	writeFile(t, filepath.Join(addon.ProcDir, "meminfo"),
		"MemTotal:         250000 kB\nMemFree:           20000 kB\nMemAvailable:      60000 kB\nBuffers: 1 kB\n")
	writeFile(t, filepath.Join(addon.ProcDir, "swaps"),
		"Filename\t\t\t\tType\t\tSize\t\tUsed\t\tPriority\n"+
			"/dev/zram0                              partition\t65532\t\t1024\t\t100\n"+
			"/swapfile                               file\t\t131068\t\t0\t\t-2\n")
	writeFile(t, filepath.Join(memZramDir, "mm_stat"), "4194304 1048576 1310720 0 1310720 0 0 0 0\n")
	writeFile(t, filepath.Join(vpn.CgroupDir, "memory.current"), "41943040\n")
	writeFile(t, filepath.Join(vpn.CgroupDir, "memory.high"), "62914560\n")
	writeFile(t, filepath.Join(vpn.CgroupDir, "memory.max"), "max\n")

	writeProc(t, 100, "NanoKVM-Server", 30000)
	writeProc(t, 200, "sh", 500)
	writeProc(t, 300, "kvm_system", 12000)
	f.running(t, f.vpnA, 400)
	writeProc(t, 400, "a-daemon", 25000)

	got := readMemory()
	if got.Total != 250000*1024 || got.Available != 60000*1024 || got.Free != 20000*1024 {
		t.Errorf("RAM = %d/%d/%d", got.Total, got.Available, got.Free)
	}

	wantSwaps := []proto.MemorySwap{
		{Name: "/dev/zram0", Kind: "zram", Size: 65532 * 1024, Used: 1024 * 1024},
		{Name: "/swapfile", Kind: "file", Size: 131068 * 1024, Used: 0},
	}
	if !reflect.DeepEqual(got.Swaps, wantSwaps) {
		t.Errorf("swaps = %+v, want %+v", got.Swaps, wantSwaps)
	}
	if got.ZramMemUsed != 1310720 {
		t.Errorf("zramMemUsed = %d, want 1310720", got.ZramMemUsed)
	}

	wantProcs := []proto.MemoryProcess{
		{Name: "NanoKVM-Server", RSS: 30000 * 1024},
		{Name: "kvm_system", RSS: 12000 * 1024},
		{Name: "a-daemon", RSS: 25000 * 1024},
	}
	if !reflect.DeepEqual(got.Processes, wantProcs) {
		t.Errorf("processes = %+v, want %+v", got.Processes, wantProcs)
	}

	want := &proto.MemoryGroup{Current: 41943040, High: 62914560, Max: 0}
	if !reflect.DeepEqual(got.Addons, want) {
		t.Errorf("addons = %+v, want %+v", got.Addons, want)
	}
}

// kvm_system is looked for again once the pid it had names something else,
// as after a restart.
func TestFindKvmSystemFollowsARestart(t *testing.T) {
	useMemory(t)

	writeProc(t, 300, "kvm_system", 1)
	if pid, ok := findKvmSystem(addon.ProcDir); !ok || pid != 300 {
		t.Fatalf("findKvmSystem = %d, %v, want 300", pid, ok)
	}

	writeProc(t, 300, "sh", 1)
	writeProc(t, 310, "kvm_system", 1)
	if pid, ok := findKvmSystem(addon.ProcDir); !ok || pid != 310 {
		t.Fatalf("after a restart findKvmSystem = %d, %v, want 310", pid, ok)
	}

	if err := os.RemoveAll(filepath.Join(addon.ProcDir, "310")); err != nil {
		t.Fatal(err)
	}
	if pid, ok := findKvmSystem(addon.ProcDir); ok {
		t.Errorf("with no kvm_system findKvmSystem = %d, want none", pid)
	}
}

// A swap without zram reads no mm_stat, which may hold a stale device's.
func TestReadMemoryReadsZramOnlyWhileItSwaps(t *testing.T) {
	useMemory(t)
	writeFile(t, filepath.Join(addon.ProcDir, "swaps"),
		"Filename Type Size Used Priority\n/swapfile file 1024 0 -2\n")
	writeFile(t, filepath.Join(memZramDir, "mm_stat"), "1 2 3 0 3\n")

	if got := readMemory(); got.ZramMemUsed != 0 {
		t.Errorf("zramMemUsed = %d, want 0 while zram is not a swap", got.ZramMemUsed)
	}
}

// writeMainlinePools writes the video pool and the CMA area as slot B's
// kernel publishes them (Linux 7.2.9, read on 2026-10-11).
func writeMainlinePools(t *testing.T) {
	t.Helper()
	writeFile(t, filepath.Join(dmapool.CoherentDir, "video-pool@86800000"),
		"size_kib 28672\nused_kib 17440\nfree_kib 11232\nlargest_free_kib 8192\n"+
			"peak_used_kib 24388\nfailed 0\nmap_chunk_kib 1024\nmap #+######+###+###+..#........\n")
	writeFile(t, filepath.Join(dmapool.CMADebugDir, "reserved", "count"), "6144\n")
	writeFile(t, filepath.Join(dmapool.CMADebugDir, "reserved", "used"), "1856\n")
	writeFile(t, filepath.Join(dmapool.CMADebugDir, "reserved", "maxchunk"), "3331\n")
	writeFile(t, filepath.Join(dmapool.CMASysDir, "reserved", "alloc_pages_fail"), "0\n")
}

func TestReadMemoryReportsTheMainlinePools(t *testing.T) {
	useMemory(t)
	writeMainlinePools(t)

	got := readMemory()
	page := int64(os.Getpagesize())
	poolLargest, cmaLargest := int64(8192*1024), 3331*page
	wantPools := []proto.MemoryPool{{
		Name:        "video-pool@86800000",
		Size:        28672 * 1024,
		Measured:    true,
		Used:        17440 * 1024,
		Free:        11232 * 1024,
		LargestFree: &poolLargest,
		Peak:        24388 * 1024,
	}}
	if !reflect.DeepEqual(got.VideoPools, wantPools) {
		t.Errorf("videoPools = %+v, want %+v", got.VideoPools, wantPools)
	}
	wantCma := []proto.MemoryCma{{
		Name:        "reserved",
		Total:       6144 * page,
		Used:        1856 * page,
		LargestFree: &cmaLargest,
	}}
	if !reflect.DeepEqual(got.Cma, wantCma) {
		t.Errorf("cma = %+v, want %+v", got.Cma, wantCma)
	}
}

// Slot A keeps its report as it was: with the carveout dump in place the
// pools are not read, even where the vendor kernel has a CMA area.
func TestReadMemoryLeavesThePoolsOutBesideTheCarveout(t *testing.T) {
	useMemory(t)
	writeMainlinePools(t)
	writeFile(t, filepath.Join(ion.Root, "total_mem"), "78643200\n")

	if got := readMemory(); got.VideoPools != nil || got.Cma != nil {
		t.Errorf("with ION videoPools = %+v, cma = %+v, want neither", got.VideoPools, got.Cma)
	}
}
