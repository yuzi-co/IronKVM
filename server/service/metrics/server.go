package metrics

import (
	"path/filepath"
	"runtime"
	"time"

	"NanoKVM-Server/service/vm"
)

// The server's own readers, variables so tests can stub them.
var (
	// processStart is set when the package initialises, before main runs.
	// changes() over it counts restarts, including the supervisor's.
	processStart = time.Now()

	versions = vm.Versions

	// ReadMemStats stops the world for a moment. Once per scrape is nothing
	// beside capture.
	goRuntime = func() (heapBytes uint64, goroutines int) {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		return stats.HeapAlloc, runtime.NumGoroutine()
	}
)

const (
	helpBuildInfo  = "Versions of the card image, the kernel and the server. Always 1."
	helpStartTime  = "When this server process started, in Unix seconds."
	helpResident   = "Resident memory of this server process, from /proc/self/status VmRSS."
	helpGoHeap     = "Bytes of allocated Go heap objects."
	helpGoroutines = "Goroutines in this server process."
)

func collectServer(w *Writer) {
	image, kernel, app := versions()
	w.Gauge("ironkvm_build_info", helpBuildInfo, 1, L("image", image), L("kernel", kernel), L("app", app))

	w.Gauge("ironkvm_process_start_time_seconds", helpStartTime, float64(processStart.Unix()))

	if body, ok := readText(filepath.Join(procDir, "self", "status")); ok {
		if rss, ok := parseKBFields(body)["VmRSS"]; ok {
			w.Gauge("ironkvm_process_resident_bytes", helpResident, float64(rss))
		}
	}

	heap, goroutines := goRuntime()
	w.Gauge("ironkvm_go_heap_bytes", helpGoHeap, float64(heap))
	w.Gauge("ironkvm_go_goroutines", helpGoroutines, float64(goroutines))
}
