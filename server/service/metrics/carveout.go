package metrics

import "NanoKVM-Server/service/ion"

// readIon is ion.Read, a variable so tests can hand it a reading.
var readIon = ion.Read

// The heap dump has no peak of its own, and ion.Init resets the watermark for
// another purpose, so the peak is left to Prometheus at scrape resolution.
const helpIon = "ION carveout, from the heap dump. The peak is max_over_time of used."

func collectIon(w *Writer) {
	status := readIon()
	if status.Verdict == ion.VerdictUnavailable {
		return
	}

	w.Gauge("ironkvm_ion_bytes", helpIon, float64(status.Total), L("kind", "total"))
	w.Gauge("ironkvm_ion_bytes", helpIon, float64(status.Used), L("kind", "used"))
}
