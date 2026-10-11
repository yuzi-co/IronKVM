package metrics

import (
	"NanoKVM-Server/service/dmapool"
	"NanoKVM-Server/service/ion"
)

// readDevicePools reads the mainline kernel's pools. ok is false beside the
// vendor kernel's ION carveout, which collectIon reports, so a slot A scrape
// stays as it was. A variable so tests can hand it a reading.
var readDevicePools = func() (dmapool.Snapshot, bool) {
	if ion.Present() {
		return dmapool.Snapshot{}, false
	}
	return dmapool.Read(), true
}

const (
	helpVideoPool       = "Reserved video pool, from debugfs dma_coherent or its CMA area. A pool the kernel does not measure has its size only."
	helpVideoPoolFailed = "Allocations the reserved video pool refused, since boot or the last reset."
	helpCma             = "CMA area, from debugfs cma (used is what cma_alloc handed out) or sysfs."
	helpCmaFailed       = "Pages a CMA area failed to allocate, from sysfs alloc_pages_fail."
)

func collectDevicePools(w *Writer) {
	snap, ok := readDevicePools()
	if !ok {
		return
	}

	for _, p := range snap.Pools {
		pool := L("pool", p.Name)
		w.Gauge("ironkvm_video_pool_bytes", helpVideoPool, float64(p.Size), pool, L("kind", "size"))
		if !p.Measured {
			continue
		}
		w.Gauge("ironkvm_video_pool_bytes", helpVideoPool, float64(p.Used), pool, L("kind", "used"))
		w.Gauge("ironkvm_video_pool_bytes", helpVideoPool, float64(p.Free), pool, L("kind", "free"))
		if p.LargestFreeKnown {
			w.Gauge("ironkvm_video_pool_bytes", helpVideoPool, float64(p.LargestFree), pool, L("kind", "largest_free"))
		}
		if p.Peak > 0 {
			w.Gauge("ironkvm_video_pool_bytes", helpVideoPool, float64(p.Peak), pool, L("kind", "peak"))
		}
	}
	for _, p := range snap.Pools {
		if p.Measured {
			w.Counter("ironkvm_video_pool_failed_allocations_total", helpVideoPoolFailed, float64(p.Failed), L("pool", p.Name))
		}
	}

	for _, a := range snap.Areas {
		area := L("area", a.Name)
		w.Gauge("ironkvm_cma_bytes", helpCma, float64(a.Total), area, L("kind", "total"))
		w.Gauge("ironkvm_cma_bytes", helpCma, float64(a.Used), area, L("kind", "used"))
		if a.LargestFreeKnown {
			w.Gauge("ironkvm_cma_bytes", helpCma, float64(a.LargestFree), area, L("kind", "largest_free"))
		}
	}
	for _, a := range snap.Areas {
		w.Counter("ironkvm_cma_alloc_fail_pages_total", helpCmaFailed, float64(a.Failed), L("area", a.Name))
	}
}
