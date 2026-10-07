package procs

import "testing"

// fake records what Two asked of the runtime.
func fake(t *testing.T, apply bool) *[]int {
	t.Helper()

	var calls []int
	current := 1

	oldSet, oldApplies := setMaxProcs, applies
	setMaxProcs = func(n int) int {
		calls = append(calls, n)
		was := current
		current = n

		return was
	}
	applies = func() bool { return apply }

	t.Cleanup(func() {
		setMaxProcs, applies = oldSet, oldApplies
		holders, previous = 0, 0
	})

	return &calls
}

func equal(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func TestTwoRaisesAndPutsBack(t *testing.T) {
	calls := fake(t, true)

	release := Two()
	release()

	if !equal(*calls, []int{2, 1}) {
		t.Fatalf("GOMAXPROCS calls %v, want [2 1]", *calls)
	}
}

func TestOverlappingHoldersShareOneChange(t *testing.T) {
	calls := fake(t, true)

	first := Two()
	second := Two()
	first()

	if !equal(*calls, []int{2}) {
		t.Fatalf("after the first release: %v, want [2]", *calls)
	}

	second()

	if !equal(*calls, []int{2, 1}) {
		t.Fatalf("after both: %v, want [2 1]", *calls)
	}
}

func TestReleasingTwiceCountsOnce(t *testing.T) {
	calls := fake(t, true)

	first := Two()
	second := Two()
	first()
	first()

	if !equal(*calls, []int{2}) {
		t.Fatalf("a second release of the same holder put the value back: %v", *calls)
	}

	second()

	if !equal(*calls, []int{2, 1}) {
		t.Fatalf("after both: %v, want [2 1]", *calls)
	}
}

func TestNothingChangesWhereItDoesNotApply(t *testing.T) {
	calls := fake(t, false)

	Two()()

	if len(*calls) != 0 {
		t.Fatalf("GOMAXPROCS changed on a board it does not apply to: %v", *calls)
	}
}
