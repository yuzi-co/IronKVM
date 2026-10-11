package metrics

import (
	"testing"

	"NanoKVM-Server/service/dmapool"
)

// boardSnapshot is slot B's reading on 2026-10-11 (Linux 7.2.9), as
// dmapool.Read returns it with 4 KiB pages.
var boardSnapshot = dmapool.Snapshot{
	Pools: []dmapool.Pool{{
		Name:             "video-pool@86800000",
		Size:             29360128,
		Measured:         true,
		Used:             17858560,
		Free:             11501568,
		LargestFree:      8388608,
		LargestFreeKnown: true,
		Peak:             24973312,
	}},
	Areas: []dmapool.Area{{
		Name:             "reserved",
		Total:            25165824,
		Used:             7602176,
		LargestFree:      13643776,
		LargestFreeKnown: true,
	}},
}

func TestDevicePoolsWriteThePoolAndTheCmaArea(t *testing.T) {
	setVar(t, &readDevicePools, func() (dmapool.Snapshot, bool) { return boardSnapshot, true })

	want := `# HELP ironkvm_video_pool_bytes Reserved video pool, from debugfs dma_coherent or its CMA area. A pool the kernel does not measure has its size only.
# TYPE ironkvm_video_pool_bytes gauge
ironkvm_video_pool_bytes{pool="video-pool@86800000",kind="size"} 29360128
ironkvm_video_pool_bytes{pool="video-pool@86800000",kind="used"} 17858560
ironkvm_video_pool_bytes{pool="video-pool@86800000",kind="free"} 11501568
ironkvm_video_pool_bytes{pool="video-pool@86800000",kind="largest_free"} 8388608
ironkvm_video_pool_bytes{pool="video-pool@86800000",kind="peak"} 24973312
# HELP ironkvm_video_pool_failed_allocations_total Allocations the reserved video pool refused, since boot or the last reset.
# TYPE ironkvm_video_pool_failed_allocations_total counter
ironkvm_video_pool_failed_allocations_total{pool="video-pool@86800000"} 0
# HELP ironkvm_cma_bytes CMA area, from debugfs cma (used is what cma_alloc handed out) or sysfs.
# TYPE ironkvm_cma_bytes gauge
ironkvm_cma_bytes{area="reserved",kind="total"} 25165824
ironkvm_cma_bytes{area="reserved",kind="used"} 7602176
ironkvm_cma_bytes{area="reserved",kind="largest_free"} 13643776
# HELP ironkvm_cma_alloc_fail_pages_total Pages a CMA area failed to allocate, from sysfs alloc_pages_fail.
# TYPE ironkvm_cma_alloc_fail_pages_total counter
ironkvm_cma_alloc_fail_pages_total{area="reserved"} 0
`
	assertText(t, render(t, collectDevicePools), want)
}

func TestAnUnmeasuredPoolWritesItsSizeOnly(t *testing.T) {
	setVar(t, &readDevicePools, func() (dmapool.Snapshot, bool) {
		return dmapool.Snapshot{Pools: []dmapool.Pool{{Name: "video-pool@86800000", Size: 29360128}}}, true
	})

	want := `# HELP ironkvm_video_pool_bytes Reserved video pool, from debugfs dma_coherent or its CMA area. A pool the kernel does not measure has its size only.
# TYPE ironkvm_video_pool_bytes gauge
ironkvm_video_pool_bytes{pool="video-pool@86800000",kind="size"} 29360128
`
	assertText(t, render(t, collectDevicePools), want)
}

func TestDevicePoolsAreLeftOutBesideTheCarveout(t *testing.T) {
	setVar(t, &readDevicePools, func() (dmapool.Snapshot, bool) { return boardSnapshot, false })

	assertText(t, render(t, collectDevicePools), "")
}
