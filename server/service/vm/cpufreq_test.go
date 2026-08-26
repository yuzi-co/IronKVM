package vm

import (
	"os"
	"path/filepath"
	"testing"
)

// useCpuFreqPaths points the config, the installed script, and the packaged
// source at a temporary tree, and restores them when the test ends. It writes a
// stand-in source so installCpuFreqInitScript has something to copy.
func useCpuFreqPaths(t *testing.T) (config string, script string, source string) {
	t.Helper()

	dir := t.TempDir()
	config = filepath.Join(dir, "cpufreq")
	script = filepath.Join(dir, "S00cpufreq")
	source = filepath.Join(dir, "S00cpufreq.src")

	if err := os.WriteFile(source, []byte("#!/bin/sh\nexit 0\n"), 0o644); err != nil {
		t.Fatalf("failed to write source: %s", err)
	}

	origConfig, origScript, origSource := cpuFreqConfigPath, cpuFreqInitScript, cpuFreqInitSource
	t.Cleanup(func() {
		cpuFreqConfigPath, cpuFreqInitScript, cpuFreqInitSource = origConfig, origScript, origSource
	})
	cpuFreqConfigPath, cpuFreqInitScript, cpuFreqInitSource = config, script, source

	return config, script, source
}

// pathExists reports whether a path is present. It is local to this test so the
// feature stays independent of helpers defined by other files in the package.
func pathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// useRegisters replaces the register reader with a fixed map, restoring the
// real one afterwards. A missing address reports ok=false, standing in for a
// register devmem cannot read.
func useRegisters(t *testing.T, values map[uint32]uint32) {
	t.Helper()

	original := readRegister
	t.Cleanup(func() { readRegister = original })

	readRegister = func(addr uint32) (uint32, bool) {
		v, ok := values[addr]
		return v, ok
	}
}

func TestPllRateMHz(t *testing.T) {
	// The three register values are read from SG2002 hardware: MPLL at stock,
	// TPLL programmed to 1000, and TPLL at its own stock 1400.
	cases := []struct {
		csr  uint32
		want int
	}{
		{0x00448101, 850},
		{0x07508101, 1000},
		{0x07708101, 1400},
	}

	for _, tc := range cases {
		if got := pllRateMHz(tc.csr); got != tc.want {
			t.Errorf("pllRateMHz(0x%08x) = %d, want %d", tc.csr, got, tc.want)
		}
	}
}

func TestPllRateMHzGuardsZeroDivider(t *testing.T) {
	// pre_div_sel and post_div_sel are both 0 here, and div_sel is 2. Without
	// the guard this divides by zero and panics; with it the dividers count as
	// 1 and the rate is 25 * 2 = 50 MHz. The number is not meaningful; the
	// point is that the call returns.
	if got := pllRateMHz(0x00040000); got != 50 {
		t.Errorf("pllRateMHz on a zero-divider register = %d, want 50", got)
	}
}

func TestReadRunningMHzStock(t *testing.T) {
	// The state a factory board boots in: mux on MPLL, MPLL at 850.
	useRegisters(t, map[uint32]uint32{
		muxReg:  0x00010309,
		mpllReg: 0x00448101,
	})

	mhz, measured := readRunningMHz()
	if !measured {
		t.Fatal("readRunningMHz reported the stock state as unknown")
	}
	if mhz != 850 {
		t.Errorf("readRunningMHz = %d, want 850", mhz)
	}
}

func TestReadRunningMHzOverclocked(t *testing.T) {
	// The state S00cpufreq produces: mux on TPLL, TPLL at 1000.
	useRegisters(t, map[uint32]uint32{
		muxReg:  0x00010009,
		tpllReg: 0x07508101,
	})

	mhz, measured := readRunningMHz()
	if !measured {
		t.Fatal("readRunningMHz reported the 1000 MHz state as unknown")
	}
	if mhz != 1000 {
		t.Errorf("readRunningMHz = %d, want 1000", mhz)
	}
}

func TestReadRunningMHzUnknownParent(t *testing.T) {
	// A mux that selects a parent this feature never sets must report unknown,
	// not a guessed number.
	useRegisters(t, map[uint32]uint32{
		muxReg:  0x00010109, // sel 1: neither MPLL nor TPLL
		mpllReg: 0x00448101,
		tpllReg: 0x07708101,
	})

	if _, measured := readRunningMHz(); measured {
		t.Error("readRunningMHz claimed to know an unrecognized mux state")
	}
}

func TestReadRunningMHzUnreadable(t *testing.T) {
	// devmem absent: every read fails, and the result is honest ignorance.
	useRegisters(t, map[uint32]uint32{})

	if _, measured := readRunningMHz(); measured {
		t.Error("readRunningMHz claimed a reading with no registers")
	}
}

func TestParseDevmemValue(t *testing.T) {
	if v, ok := parseDevmemValue("0x00448101\n"); !ok || v != 0x00448101 {
		t.Errorf("parseDevmemValue = 0x%08x, %v; want 0x00448101, true", v, ok)
	}
	if _, ok := parseDevmemValue("not a number"); ok {
		t.Error("parseDevmemValue accepted garbage")
	}
}

func TestReadTargetFreqDefaultsToStock(t *testing.T) {
	useCpuFreqPaths(t)

	// No config file at all.
	if got := readTargetFreq(); got != stockFreqMHz {
		t.Errorf("readTargetFreq with no file = %d, want %d", got, stockFreqMHz)
	}

	// A config naming a frequency the feature does not offer falls back to
	// stock, matching what the boot script does with the same value.
	if err := os.WriteFile(cpuFreqConfigPath, []byte("1200\n"), 0o644); err != nil {
		t.Fatalf("failed to write config: %s", err)
	}
	if got := readTargetFreq(); got != stockFreqMHz {
		t.Errorf("readTargetFreq with an unoffered value = %d, want %d", got, stockFreqMHz)
	}
}

func TestApplyTargetFreqOverclockWritesConfigAndInstallsScript(t *testing.T) {
	config, script, _ := useCpuFreqPaths(t)

	if err := applyTargetFreq(1000); err != nil {
		t.Fatalf("applyTargetFreq(1000) = %s", err)
	}

	if got := readTargetFreq(); got != 1000 {
		t.Errorf("target after applyTargetFreq(1000) = %d, want 1000", got)
	}
	if !pathExists(script) {
		t.Error("applyTargetFreq(1000) did not install the boot script")
	}
	if !pathExists(config) {
		t.Error("applyTargetFreq(1000) did not write the config")
	}
}

func TestApplyTargetFreqStockIsPristine(t *testing.T) {
	config, script, _ := useCpuFreqPaths(t)

	// Start from the overclocked state.
	if err := applyTargetFreq(1000); err != nil {
		t.Fatalf("applyTargetFreq(1000) = %s", err)
	}

	// Selecting stock must remove both the config and the boot script, so the
	// next boot follows the factory path with nothing extra to run.
	if err := applyTargetFreq(stockFreqMHz); err != nil {
		t.Fatalf("applyTargetFreq(stock) = %s", err)
	}

	if pathExists(config) {
		t.Error("applyTargetFreq(stock) left the config in place")
	}
	if pathExists(script) {
		t.Error("applyTargetFreq(stock) left the boot script in place")
	}
	if got := readTargetFreq(); got != stockFreqMHz {
		t.Errorf("target after applyTargetFreq(stock) = %d, want %d", got, stockFreqMHz)
	}
}

func TestApplyTargetFreqRejectsUnsupported(t *testing.T) {
	useCpuFreqPaths(t)

	if err := applyTargetFreq(1200); err == nil {
		t.Error("applyTargetFreq accepted an unsupported frequency")
	}
	if pathExists(cpuFreqConfigPath) {
		t.Error("a rejected frequency still wrote a config")
	}
}

func TestIsAllowedFreq(t *testing.T) {
	for _, mhz := range []int{850, 1000} {
		if !isAllowedFreq(mhz) {
			t.Errorf("isAllowedFreq(%d) = false, want true", mhz)
		}
	}
	for _, mhz := range []int{0, 900, 1100, 1200} {
		if isAllowedFreq(mhz) {
			t.Errorf("isAllowedFreq(%d) = true, want false", mhz)
		}
	}
}
