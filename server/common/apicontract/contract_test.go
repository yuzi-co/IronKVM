package apicontract

import (
	"regexp"
	"sort"
	"testing"
)

var featureName = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)

func TestFeaturesAreSortedUniqueAndWellFormed(t *testing.T) {
	// The table is kept sorted in the source so that branches adding lines
	// touch different places; Features sorts anyway, so a slip costs nothing
	// at run time, but it is caught here.
	if !sort.StringsAreSorted(features) {
		t.Fatalf("the registry is not in alphabetical order: %v", features)
	}
	seen := map[string]bool{}
	for _, f := range features {
		if !featureName.MatchString(f) {
			t.Errorf("feature %q is not lower-case words joined by hyphens", f)
		}
		if seen[f] {
			t.Errorf("feature %q is registered twice", f)
		}
		seen[f] = true
	}
}

func TestFeaturesReturnsACopy(t *testing.T) {
	got := Features()
	got[0] = "changed"
	if Features()[0] == "changed" {
		t.Fatal("a caller changed the registry through the returned slice")
	}
}

func TestHas(t *testing.T) {
	if !Has("status-bar") {
		t.Fatal("status-bar should be registered")
	}
	if Has("no-such-feature") {
		t.Fatal("an unregistered name was reported as present")
	}
}

func TestAPIVersionIsPositive(t *testing.T) {
	// 0 is what the UI would read from a broken answer; 1 is the version a
	// server without the endpoint is assumed to speak.
	if APIVersion < 1 {
		t.Fatalf("APIVersion must be at least 1, got %d", APIVersion)
	}
}
