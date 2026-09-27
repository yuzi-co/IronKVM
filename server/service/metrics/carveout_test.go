package metrics

import (
	"testing"

	"NanoKVM-Server/service/ion"
)

func TestIonWritesTotalAndUsed(t *testing.T) {
	setVar(t, &readIon, func() ion.Status {
		return ion.Status{Total: 78643200, Used: 19050496, Verdict: ion.VerdictOK}
	})

	want := `# HELP ironkvm_ion_bytes ION carveout, from the heap dump. The peak is max_over_time of used.
# TYPE ironkvm_ion_bytes gauge
ironkvm_ion_bytes{kind="total"} 78643200
ironkvm_ion_bytes{kind="used"} 19050496
`
	assertText(t, render(t, collectIon), want)
}

func TestIonIsLeftOutWhenTheCarveoutCannotBeRead(t *testing.T) {
	setVar(t, &readIon, func() ion.Status {
		return ion.Status{Verdict: ion.VerdictUnavailable}
	})

	assertText(t, render(t, collectIon), "")
}
