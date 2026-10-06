package common

import "testing"

// The aspect setting decides whether a source of another shape than the
// chosen resolution keeps its shape (ironkvm-dist#37). It is stored and
// restored like the other screen settings, and handed to the library.

func TestTheAspectIsStoredOnTheCard(t *testing.T) {
	file, ok := ScreenFileMap["aspect"]
	if !ok {
		t.Fatal("aspect has no file in ScreenFileMap, so a choice would not survive a restart")
	}
	if file != "/kvmapp/kvm/aspect" {
		t.Fatalf("aspect file = %q, want /kvmapp/kvm/aspect", file)
	}
}

func TestAnUnconfiguredBoardKeepsTheSourceShape(t *testing.T) {
	withScreenFiles(t, map[string]string{})

	if values := loadScreenValues(); values.Aspect != AspectKeep {
		t.Fatalf("aspect = %d, want keep (%d)", values.Aspect, AspectKeep)
	}
}

func TestAStoredAspectIsRestored(t *testing.T) {
	withScreenFiles(t, map[string]string{"aspect": "1"})

	if values := loadScreenValues(); values.Aspect != AspectStretch {
		t.Fatalf("aspect = %d, want stretch (%d)", values.Aspect, AspectStretch)
	}
}

func TestAnUnknownAspectIsLeftAlone(t *testing.T) {
	for _, aspect := range []int{-1, 2, 255} {
		values := defaultScreenValues
		values.Aspect = AspectStretch
		applyScreenValue(&values, "aspect", aspect)

		if values.Aspect != AspectStretch {
			t.Fatalf("applyScreenValue(aspect=%d) gave %d, want it left alone", aspect, values.Aspect)
		}
	}
}

func TestSettingTheAspectTellsTheLibrary(t *testing.T) {
	// The first GetScreen tells the library the stored setting; that is not
	// what this test counts.
	before := GetScreen().Snapshot().Aspect

	var told []bool
	previous := setLibraryKeepAspect
	setLibraryKeepAspect = func(keep bool) bool {
		told = append(told, keep)
		return true
	}
	t.Cleanup(func() { setLibraryKeepAspect = previous })
	t.Cleanup(func() { SetScreen("aspect", int(before)) })

	SetScreen("aspect", AspectStretch)
	SetScreen("aspect", AspectKeep)

	if len(told) != 2 || told[0] || !told[1] {
		t.Fatalf("library told %v, want [false true]", told)
	}
}
