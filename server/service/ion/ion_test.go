package ion

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

// fakeCarveout builds a directory with the same file names as the real debugfs
// entry and points Root at it for the duration of the test.
func fakeCarveout(t *testing.T, total, alloc, peak uint64, summary string) string {
	t.Helper()

	dir := t.TempDir()
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %s", name, err)
		}
	}

	write("total_mem", strconv.FormatUint(total, 10))
	write("alloc_mem", strconv.FormatUint(alloc, 10))
	write("peak", strconv.FormatUint(peak, 10))
	if summary != "" {
		write("summary", summary)
	}

	original := Root
	Root = dir
	t.Cleanup(func() { Root = original })

	return dir
}

func TestInitResetsPeakSoTheRequirementCanBeMeasured(t *testing.T) {
	dir := fakeCarveout(t, 78643200, 19050496, 19050496, summaryIdle)

	Init(25165824)

	body, err := os.ReadFile(filepath.Join(dir, "peak"))
	if err != nil {
		t.Fatalf("read peak: %s", err)
	}
	if got := string(body); got != "0" {
		t.Fatalf("peak = %q after Init, want %q", got, "0")
	}
}

// TestTheReserveFallsAsTheGenerationSpendsTheFloor holds the model the reserve
// uses: what has to stay free is what this generation has not allocated yet.
//
// A generation that has taken 11,341,824 bytes of a 12,582,912 floor has
// 1,241,088 of that cost still ahead of it, so that is what it must find. It
// does not have to find the whole floor again, and it does not have to find
// what it already holds.
func TestTheReserveFallsAsTheGenerationSpendsTheFloor(t *testing.T) {
	dir := fakeCarveout(t, 58720256, 19050496, 19050496, summaryIdle)
	Init(12582912)

	// 30392320 is what opening a stream costs, measured on the device.
	writeCounter(t, dir, "alloc_mem", 30392320)
	writeCounter(t, dir, "peak", 30392320)

	got := Read()

	// spent 11341824 of the 12582912 floor, so 1241088 is still ahead - and
	// that is under the orphan cost, which is the value that has to win.
	if got.Reserve != 6516736 {
		t.Fatalf("Reserve = %d, want the orphan cost 6516736", got.Reserve)
	}
	if !got.Measured {
		t.Fatal("Measured = false, want true once a peak has been read")
	}
}

// TestAFullyExercisedGenerationReservesOnlyTheOrphanCost is the case the
// 2026-08-20 carveout resize turned up on real hardware.
//
// The board had run every delivery path, so its peak was its whole working set
// and nothing more was pending. The old model graded it against that high-water
// mark, found 15,777,792 free against a 23,891,968 "reserve", and told the
// operator to reboot a healthy board. Requiring the carveout to hold the working
// set a second time was right only while a dying process kept its buffers.
func TestAFullyExercisedGenerationReservesOnlyTheOrphanCost(t *testing.T) {
	dir := fakeCarveout(t, 58720256, 19050496, 19050496, summaryIdle)
	Init(12582912)

	writeCounter(t, dir, "alloc_mem", 42942464)
	writeCounter(t, dir, "peak", 42942464)

	got := Read()

	if got.Reserve != 6516736 {
		t.Fatalf("Reserve = %d, want the orphan cost 6516736", got.Reserve)
	}
	if got.Free != 15777792 {
		t.Fatalf("Free = %d, want 15777792", got.Free)
	}
	if got.Verdict != VerdictOK {
		t.Fatalf("Verdict = %q, want %q: this board is healthy", got.Verdict, VerdictOK)
	}
}

// TestTheFloorStandsBeforeTheGenerationHasGrown keeps the half of the old
// asymmetry that is still true. A board that has not opened a stream has the
// whole cost of opening one ahead of it, and reporting a small reserve there
// would read "ok" right up to the allocation that segfaults the server.
func TestTheFloorStandsBeforeTheGenerationHasGrown(t *testing.T) {
	dir := fakeCarveout(t, 58720256, 19050496, 19050496, summaryIdle)
	Init(12582912)

	writeCounter(t, dir, "alloc_mem", 19050496)
	writeCounter(t, dir, "peak", 19050496)

	got := Read()

	if got.Reserve != 12582912 {
		t.Fatalf("Reserve = %d, want the whole floor 12582912", got.Reserve)
	}
}

// TestTheReserveNeverFallsBelowOneOrphan states the part that does not depend
// on what the generation has done. A process that dies without its teardown
// leaves 6,516,736 bytes that only a reboot returns, so a board with less than
// that free is one crash away from a carveout it cannot recover.
func TestTheReserveNeverFallsBelowOneOrphan(t *testing.T) {
	dir := fakeCarveout(t, 58720256, 19050496, 19050496, summaryIdle)
	Init(1024)

	writeCounter(t, dir, "alloc_mem", 42942464)
	writeCounter(t, dir, "peak", 42942464)

	got := Read()

	if got.Reserve != 6516736 {
		t.Fatalf("Reserve = %d, want the orphan cost 6516736", got.Reserve)
	}
}

// TestOrphansEatTheHeadroomUntilItIsCritical walks the failure the grade exists
// to catch, on the resized carveout. Two crashed generations leave 13,033,472
// bytes stranded, and what is left is under one more orphan.
func TestOrphansEatTheHeadroomUntilItIsCritical(t *testing.T) {
	dir := fakeCarveout(t, 58720256, 19050496, 19050496, summaryIdle)
	Init(12582912)

	writeCounter(t, dir, "alloc_mem", 55975936)
	writeCounter(t, dir, "peak", 55975936)

	got := Read()

	if got.Free != 2744320 {
		t.Fatalf("Free = %d, want 2744320", got.Free)
	}
	if got.Verdict != VerdictCritical {
		t.Fatalf("Verdict = %q, want %q", got.Verdict, VerdictCritical)
	}
}

func TestReadFallsBackToTheFloorBeforeAnythingHasBeenCaptured(t *testing.T) {
	fakeCarveout(t, 78643200, 19050496, 19050496, summaryIdle)
	Init(25165824)

	got := Read()

	if got.Reserve != 25165824 {
		t.Fatalf("Reserve = %d, want the floor 25165824", got.Reserve)
	}
	if got.Measured {
		t.Fatal("Measured = true, want false while the floor is in use")
	}
}

func TestReadReportsTheDerivedFields(t *testing.T) {
	fakeCarveout(t, 78643200, 19050496, 19050496, summaryIdle)
	Init(25165824)

	got := Read()

	if got.Total != 78643200 || got.Used != 19050496 {
		t.Fatalf("Total/Used = %d/%d, want 78643200/19050496", got.Total, got.Used)
	}
	if got.Free != 59592704 {
		t.Fatalf("Free = %d, want 59592704", got.Free)
	}
	if got.UsageRate != 24 {
		t.Fatalf("UsageRate = %d, want 24", got.UsageRate)
	}
	if got.Generations != 1 {
		t.Fatalf("Generations = %d, want 1", got.Generations)
	}
	if got.Verdict != VerdictOK {
		t.Fatalf("Verdict = %q, want %q", got.Verdict, VerdictOK)
	}
}

// TestReadReportsCritical is the one mutation this feature must never survive:
// a hard-coded VerdictOK in Read passes every other test in this package, but
// verdict == "critical" is the sole input to the desktop stream gate, so an
// always-healthy Read would ship silently. free (7,864,320) is below the
// floor (25,165,824) here, so the assertion depends on Verdict actually being
// called, not on the argument order Verdict itself already pins.
func TestReadReportsCritical(t *testing.T) {
	fakeCarveout(t, 78643200, 70778880, 70778880, summaryIdle)
	Init(25165824)

	got := Read()

	if got.Verdict != VerdictCritical {
		t.Fatalf("Verdict = %q, want %q", got.Verdict, VerdictCritical)
	}
}

// TestReadReportsWarn covers the middle band: free (38,643,200) sits between
// the floor and twice the floor, so Read must reach the warn branch of
// Verdict and not just the ok/critical extremes the other fixtures exercise.
func TestReadReportsWarn(t *testing.T) {
	fakeCarveout(t, 78643200, 40000000, 40000000, summaryIdle)
	Init(25165824)

	got := Read()

	if got.Verdict != VerdictWarn {
		t.Fatalf("Verdict = %q, want %q", got.Verdict, VerdictWarn)
	}
}

// TestATornReadClampsFreeAndUsageRate covers the used <= total guard added in
// response to a review minor. total_mem and alloc_mem are two separate,
// non-atomic reads of files the kernel can update between them, so a torn
// read can briefly report used > total; without the guard, Free underflows
// and UsageRate divides out to something other than the clamped 100.
func TestATornReadClampsFreeAndUsageRate(t *testing.T) {
	fakeCarveout(t, 78643200, 78643201, 78643201, summaryIdle)
	Init(25165824)

	got := Read()

	if got.Free != 0 {
		t.Fatalf("Free = %d, want 0 when a torn read reports used > total", got.Free)
	}
	if got.UsageRate != 100 {
		t.Fatalf("UsageRate = %d, want 100 when a torn read reports used > total", got.UsageRate)
	}
}

func TestReadCountsTheOrphanedGeneration(t *testing.T) {
	fakeCarveout(t, 78643200, 49459200, 49459200, summaryTwoGenerations)
	Init(25165824)

	if got := Read().Generations; got != 2 {
		t.Fatalf("Generations = %d, want 2", got)
	}
}

func TestMissingCountersAreUnavailableAndNotAnError(t *testing.T) {
	original := Root
	Root = filepath.Join(t.TempDir(), "not-here")
	t.Cleanup(func() { Root = original })

	Init(25165824)
	got := Read()

	if got.Verdict != VerdictUnavailable {
		t.Fatalf("Verdict = %q, want %q", got.Verdict, VerdictUnavailable)
	}
	if got.Total != 0 || got.Used != 0 {
		t.Fatalf("Total/Used = %d/%d, want 0/0", got.Total, got.Used)
	}
}

func TestAMissingSummaryStillReportsTheCounters(t *testing.T) {
	fakeCarveout(t, 78643200, 19050496, 19050496, "")
	Init(25165824)

	got := Read()

	if got.Generations != 0 {
		t.Fatalf("Generations = %d, want 0 when the summary cannot be read", got.Generations)
	}
	if got.Verdict == VerdictUnavailable {
		t.Fatal("Verdict is unavailable, want the counters to still be graded")
	}
}

// TestAReadOnlyPeakDoesNotBreakTheEndpoint isolates Init's write-failure
// handling specifically. peak stays a normal, readable file holding a value
// whose growth would exceed the floor if Read got to see it - the only thing
// that fails is the write inside Init, staged through the writeFile seam
// rather than a filesystem permission, because root ignores the write bit and
// a directory-shaped peak fails the later read too, which would let this test
// pass even if Init did not check the write's error at all.
func TestAReadOnlyPeakDoesNotBreakTheEndpoint(t *testing.T) {
	// alloc_mem/peak both start at a value whose eventual growth, if measured,
	// would exceed the floor - so if the reset failure were ignored and Read
	// still measured from it, this test would catch that too.
	fakeCarveout(t, 78643200, 19050496, 49459200, summaryIdle)

	original := writeFile
	writeFile = func(string, []byte, os.FileMode) error { return errors.New("read-only peak") }
	t.Cleanup(func() { writeFile = original })

	Init(25165824)
	got := Read()

	if got.Verdict == VerdictUnavailable {
		t.Fatal("Verdict is unavailable, want a graded reading from the floor")
	}
	if got.Measured {
		t.Fatal("Measured = true, want false when peak could not be reset")
	}
}

// TestAnUnreadablePeakFallsBackToTheFloor covers the filesystem case: peak
// exists as a directory instead of a file, so both the write in Init and the
// later read in Read fail against it. Root ignores the write bit, so a
// permission-based construction cannot stage a write failure on its own; this
// is a different, real failure path from a directory-shaped peak, not a
// substitute for TestAReadOnlyPeakDoesNotBreakTheEndpoint above.
func TestAnUnreadablePeakFallsBackToTheFloor(t *testing.T) {
	dir := fakeCarveout(t, 78643200, 19050496, 19050496, summaryIdle)
	if err := os.Remove(filepath.Join(dir, "peak")); err != nil {
		t.Fatalf("remove peak: %s", err)
	}
	if err := os.Mkdir(filepath.Join(dir, "peak"), 0o755); err != nil {
		t.Fatalf("mkdir peak: %s", err)
	}

	Init(25165824)
	got := Read()

	if got.Verdict == VerdictUnavailable {
		t.Fatal("Verdict is unavailable, want a graded reading from the floor")
	}
	if got.Measured {
		t.Fatal("Measured = true, want false when peak could not be reset")
	}
}

func writeCounter(t *testing.T, dir, name string, v uint64) {
	t.Helper()
	body := []byte(strconv.FormatUint(v, 10))
	if err := os.WriteFile(filepath.Join(dir, name), body, 0o644); err != nil {
		t.Fatalf("write %s: %s", name, err)
	}
}
