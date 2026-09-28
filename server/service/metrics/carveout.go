package metrics

import "NanoKVM-Server/service/ion"

// readIon is ion.ReadUsage, a variable so tests can hand it a reading. It
// reads the two counters only: ion.Read also parses the summary and works out
// the reserve on every call, which a scrape does not need.
var readIon = ion.ReadUsage

// The heap dump has no peak of its own, and ion.Init resets the watermark for
// another purpose, so the peak is left to Prometheus at scrape resolution.
const helpIon = "ION carveout, from the heap dump. The peak is max_over_time of used."

func collectIon(w *Writer) {
	total, used, ok := readIon()
	if !ok {
		return
	}

	w.Gauge("ironkvm_ion_bytes", helpIon, float64(total), L("kind", "total"))
	w.Gauge("ironkvm_ion_bytes", helpIon, float64(used), L("kind", "used"))
}
