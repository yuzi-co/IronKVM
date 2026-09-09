package common

import "testing"

// The GOP used to be the one capture setting the server did not keep. The API
// handed it to libkvm and stored nothing, so the browser held the only copy: a
// second viewer saw the default, and a restart lost the operator's choice
// without saying so. It is stored and restored now, like every other setting
// the menu offers.

func TestTheGOPIsStoredOnTheCard(t *testing.T) {
	file, ok := ScreenFileMap["gop"]
	if !ok {
		t.Fatal("gop has no file in ScreenFileMap, so a choice would not survive a restart")
	}
	if file != "/kvmapp/kvm/gop" {
		t.Fatalf("gop file = %q, want /kvmapp/kvm/gop", file)
	}
}

func TestAStoredGOPIsRestored(t *testing.T) {
	withScreenFiles(t, map[string]string{"gop": "12"})

	if values := loadScreenValues(); values.GOP != 12 {
		t.Fatalf("gop = %d, want 12", values.GOP)
	}
}

// libkvm clamps the GOP to 1..100 itself, so a value outside that range is not
// the value the encoder will use. Storing it anyway would make the menu report
// a GOP the board is not running.
func TestAnOutOfRangeGOPFallsBackToTheDefault(t *testing.T) {
	for _, gop := range []int{0, -1, 101, 1000} {
		values := defaultScreenValues
		values.GOP = 12
		applyScreenValue(&values, "gop", gop)

		if values.GOP != defaultScreenValues.GOP {
			t.Fatalf("applyScreenValue(gop=%d) gave %d, want the default %d",
				gop, values.GOP, defaultScreenValues.GOP)
		}
	}
}

func TestAGOPInRangeIsKept(t *testing.T) {
	for _, gop := range []int{1, 30, 100} {
		values := defaultScreenValues
		applyScreenValue(&values, "gop", gop)

		if int(values.GOP) != gop {
			t.Fatalf("applyScreenValue(gop=%d) gave %d", gop, values.GOP)
		}
	}
}

// The settings files are plain text on the card and a person can edit them, so
// the consistency check has to hold the GOP to the same rule the API does.
func TestAnImpossibleStoredGOPIsRepaired(t *testing.T) {
	for _, stored := range []uint8{0, 101, 255} {
		if got := validateGOP(int(stored)); got != defaultScreenValues.GOP {
			t.Fatalf("validateGOP(%d) = %d, want the default %d",
				stored, got, defaultScreenValues.GOP)
		}
	}
	for _, stored := range []uint8{1, 30, 100} {
		if got := validateGOP(int(stored)); got != stored {
			t.Fatalf("validateGOP(%d) = %d, want it left alone", stored, got)
		}
	}
}
