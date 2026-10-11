package dmapool

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The fixtures were read from the board on slot B (Linux 7.2.9, trial 80) on
// 2026-10-11, with an H.265 viewer open.

const coherentFixture = `size_kib 28672
used_kib 17440
free_kib 11232
largest_free_kib 8192
peak_used_kib 24388
failed 0
map_chunk_kib 1024
map #+######+###+###+..#........
`

// reserved-memory/video-pool@86800000/reg with #address-cells and
// #size-cells both 1: 0x86800000, 28 MiB.
var videoPoolReg = []byte{0x86, 0x80, 0x00, 0x00, 0x01, 0xc0, 0x00, 0x00}

// ramoops@867e0000/reg: 128 KiB, no-map but not a shared-dma-pool.
var ramoopsReg = []byte{0x86, 0x7e, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00}

type roots struct{ coherent, cmaDebug, cmaSys, reserved string }

func useRoots(t *testing.T) roots {
	t.Helper()
	base := t.TempDir()
	r := roots{
		coherent: filepath.Join(base, "dma_coherent"),
		cmaDebug: filepath.Join(base, "cma"),
		cmaSys:   filepath.Join(base, "mm-cma"),
		reserved: filepath.Join(base, "reserved-memory"),
	}
	saved := []string{CoherentDir, CMADebugDir, CMASysDir, ReservedMemDir}
	savedPage := pageSize
	t.Cleanup(func() {
		CoherentDir, CMADebugDir, CMASysDir, ReservedMemDir = saved[0], saved[1], saved[2], saved[3]
		pageSize = savedPage
	})
	CoherentDir, CMADebugDir, CMASysDir, ReservedMemDir = r.coherent, r.cmaDebug, r.cmaSys, r.reserved
	pageSize = 4096
	return r
}

func write(t *testing.T, path string, body []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeBoardTree writes the reserved-memory node and the CMA area as slot B
// has them.
func writeBoardTree(t *testing.T, r roots) {
	t.Helper()
	write(t, filepath.Join(r.reserved, "#address-cells"), []byte{0, 0, 0, 1})
	write(t, filepath.Join(r.reserved, "#size-cells"), []byte{0, 0, 0, 1})
	write(t, filepath.Join(r.reserved, "mmode_resv0@80000000", "reg"),
		[]byte{0x80, 0, 0, 0, 0, 0x04, 0, 0})
	write(t, filepath.Join(r.reserved, "ramoops@867e0000", "compatible"), []byte("ramoops\x00"))
	write(t, filepath.Join(r.reserved, "ramoops@867e0000", "no-map"), nil)
	write(t, filepath.Join(r.reserved, "ramoops@867e0000", "reg"), ramoopsReg)
	write(t, filepath.Join(r.reserved, "video-pool@86800000", "compatible"), []byte("shared-dma-pool\x00"))
	write(t, filepath.Join(r.reserved, "video-pool@86800000", "no-map"), nil)
	write(t, filepath.Join(r.reserved, "video-pool@86800000", "reg"), videoPoolReg)

	write(t, filepath.Join(r.cmaDebug, "reserved", "count"), []byte("6144\n"))
	write(t, filepath.Join(r.cmaDebug, "reserved", "used"), []byte("1856\n"))
	write(t, filepath.Join(r.cmaDebug, "reserved", "maxchunk"), []byte("3331\n"))
	write(t, filepath.Join(r.cmaDebug, "reserved", "order_per_bit"), []byte("0\n"))
	write(t, filepath.Join(r.cmaSys, "reserved", "total_pages"), []byte("6144\n"))
	write(t, filepath.Join(r.cmaSys, "reserved", "available_pages"), []byte("4288\n"))
	write(t, filepath.Join(r.cmaSys, "reserved", "alloc_pages_fail"), []byte("0\n"))
}

func TestParseCoherentReadsTheBoardsFile(t *testing.T) {
	got, ok := ParseCoherent(coherentFixture)
	want := Pool{
		Size:             28672 * 1024,
		Measured:         true,
		Used:             17440 * 1024,
		Free:             11232 * 1024,
		LargestFree:      8192 * 1024,
		LargestFreeKnown: true,
		Peak:             24388 * 1024,
	}
	if !ok || got != want {
		t.Fatalf("ParseCoherent = %+v, %v; want %+v, true", got, ok, want)
	}
}

func TestParseCoherentRejectsTextWithoutASize(t *testing.T) {
	if _, ok := ParseCoherent("used_kib 12\n"); ok {
		t.Fatal("a file without size_kib parsed")
	}
	if _, ok := ParseCoherent(""); ok {
		t.Fatal("an empty file parsed")
	}
}

func TestParseRegSize(t *testing.T) {
	cases := []struct {
		name        string
		reg         []byte
		addr, sizeC uint32
		want        uint64
	}{
		{"one cell each", videoPoolReg, 1, 1, 28 << 20},
		{"two cells each", []byte{0, 0, 0, 0, 0x86, 0x80, 0, 0, 0, 0, 0, 0, 0x01, 0xc0, 0, 0}, 2, 2, 28 << 20},
		{"two entries", append(append([]byte{}, videoPoolReg...), ramoopsReg...), 1, 1, 28<<20 + 128<<10},
		{"short property", videoPoolReg[:6], 1, 1, 0},
		{"no size cells", videoPoolReg, 1, 0, 0},
	}
	for _, c := range cases {
		if got := ParseRegSize(c.reg, c.addr, c.sizeC); got != c.want {
			t.Errorf("%s: ParseRegSize = %d, want %d", c.name, got, c.want)
		}
	}
}

func TestReadOnTheBoardsKernel(t *testing.T) {
	r := useRoots(t)
	writeBoardTree(t, r)
	write(t, filepath.Join(r.coherent, "video-pool@86800000"), []byte(coherentFixture))

	got := Read()
	want := Snapshot{
		Pools: []Pool{{
			Name:             "video-pool@86800000",
			Size:             28672 * 1024,
			Measured:         true,
			Used:             17440 * 1024,
			Free:             11232 * 1024,
			LargestFree:      8192 * 1024,
			LargestFreeKnown: true,
			Peak:             24388 * 1024,
		}},
		Areas: []Area{{
			Name:             "reserved",
			Total:            24 << 20,
			Used:             1856 * 4096,
			LargestFree:      3331 * 4096,
			LargestFreeKnown: true,
		}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Read() =\n%+v\nwant\n%+v", got, want)
	}
}

// A kernel without patch 0939 still describes the pool in the device tree, so
// the pool keeps its size and says its use is not known.
func TestAPoolWithoutDebugfsHasItsSizeOnly(t *testing.T) {
	r := useRoots(t)
	writeBoardTree(t, r)

	got := Read().Pools
	want := []Pool{{Name: "video-pool@86800000", Size: 28 << 20}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Pools = %+v, want %+v", got, want)
	}
}

// Without CMA_DEBUGFS the totals come from sysfs and the longest free run is
// unknown.
func TestACmaAreaWithoutDebugfsFallsBackToSysfs(t *testing.T) {
	r := useRoots(t)
	writeBoardTree(t, r)
	if err := os.RemoveAll(r.cmaDebug); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(r.cmaSys, "reserved", "alloc_pages_fail"), []byte("3\n"))

	got := Read().Areas
	want := []Area{{Name: "reserved", Total: 24 << 20, Used: 1856 * 4096, Failed: 3}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Areas = %+v, want %+v", got, want)
	}
}

// Trial 77's r28 kernel makes the pool reusable. It is then a CMA area named
// after the node, and reported as the pool rather than as general CMA.
func TestAReusablePoolIsReportedAsAPool(t *testing.T) {
	r := useRoots(t)
	writeBoardTree(t, r)
	node := filepath.Join(r.reserved, "video-pool@86800000")
	if err := os.Remove(filepath.Join(node, "no-map")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(node, "reusable"), nil)
	write(t, filepath.Join(r.cmaDebug, "video-pool@86800000", "count"), []byte("7168\n"))
	write(t, filepath.Join(r.cmaDebug, "video-pool@86800000", "used"), []byte("4360\n"))
	write(t, filepath.Join(r.cmaDebug, "video-pool@86800000", "maxchunk"), []byte("2048\n"))

	got := Read()
	wantPools := []Pool{{
		Name:             "video-pool@86800000",
		Size:             28 << 20,
		Measured:         true,
		Used:             4360 * 4096,
		Free:             (7168 - 4360) * 4096,
		LargestFree:      2048 * 4096,
		LargestFreeKnown: true,
	}}
	if !reflect.DeepEqual(got.Pools, wantPools) {
		t.Fatalf("Pools = %+v, want %+v", got.Pools, wantPools)
	}
	if len(got.Areas) != 1 || got.Areas[0].Name != "reserved" {
		t.Fatalf("Areas = %+v, want the default area only", got.Areas)
	}
}

// The vendor kernel's tree and a container have none of these sources.
func TestReadWithNothingToReadIsEmpty(t *testing.T) {
	useRoots(t)

	if got := Read(); len(got.Pools) != 0 || len(got.Areas) != 0 {
		t.Fatalf("Read() = %+v, want nothing", got)
	}
}
