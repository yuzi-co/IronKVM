package metrics

import (
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/service/vm"
)

// The server's own readers, variables so tests can stub them.
var (
	// processStart is set when the package initialises, before main runs.
	// changes() over it counts restarts, including the supervisor's.
	processStart = time.Now()

	// The version files change only across a restart, so they are read on
	// the first scrape and not again.
	versions = onceVersions(vm.Versions)

	// ReadMemStats stops the world for a moment. Once per scrape is nothing
	// beside capture.
	goRuntime = func() (heapBytes uint64, goroutines int) {
		var stats runtime.MemStats
		runtime.ReadMemStats(&stats)
		return stats.HeapAlloc, runtime.NumGoroutine()
	}

	// cgoCalls counts the calls from Go into C since the process started. Each
	// one that blocks long enough costs a P handoff on a board with one P, and
	// with libkvm-v4l2's capture worker (KVM_NATIVE_CAPTURE) a frame costs none.
	cgoCalls = runtime.NumCgoCall

	// nativeCapture says whether video frames come from that worker.
	nativeCapture = common.NativeCaptureActive
)

const (
	helpBuildInfo  = "Versions of the card image, the kernel and the server. Always 1."
	helpStartTime  = "When this server process started, in Unix seconds."
	helpResident   = "Resident memory of this server process, from /proc/self/status VmRSS."
	helpGoHeap     = "Bytes of allocated Go heap objects."
	helpGoroutines = "Goroutines in this server process."
	helpCgoCalls   = "Calls from Go into C since the process started. With one P, each blocking one hands the P to another thread."
	helpNative     = "1 when video frames come from libkvm's capture worker (KVM_NATIVE_CAPTURE), 0 when the server calls the library for each frame."
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
	w.Counter("ironkvm_go_cgo_calls_total", helpCgoCalls, float64(cgoCalls()))

	native := 0.0
	if nativeCapture() {
		native = 1
	}
	w.Gauge("ironkvm_capture_native", helpNative, native)
}

// onceVersions wraps read so it runs on the first call only, and every later
// call returns what it returned then.
func onceVersions(read func() (string, string, string)) func() (string, string, string) {
	type triple struct{ image, kernel, app string }

	cached := sync.OnceValue(func() triple {
		image, kernel, app := read()
		return triple{image, kernel, app}
	})

	return func() (string, string, string) {
		v := cached()
		return v.image, v.kernel, v.app
	}
}
