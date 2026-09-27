package metrics

import "testing"

func TestIonWritesTotalAndUsed(t *testing.T) {
	setVar(t, &readIon, func() (uint64, uint64, bool) { return 78643200, 19050496, true })

	want := `# HELP ironkvm_ion_bytes ION carveout, from the heap dump. The peak is max_over_time of used.
# TYPE ironkvm_ion_bytes gauge
ironkvm_ion_bytes{kind="total"} 78643200
ironkvm_ion_bytes{kind="used"} 19050496
`
	assertText(t, render(t, collectIon), want)
}

func TestIonIsLeftOutWhenTheCarveoutCannotBeRead(t *testing.T) {
	setVar(t, &readIon, func() (uint64, uint64, bool) { return 0, 0, false })

	assertText(t, render(t, collectIon), "")
}
