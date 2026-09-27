package metrics

import (
	"math"
	"strings"
	"testing"
)

func TestWriterWritesEachFamilyHeaderOnce(t *testing.T) {
	got := render(t, func(w *Writer) {
		w.Gauge("ironkvm_test", "A test.", 1, L("a", "x"))
		w.Gauge("ironkvm_test", "A test.", 2, L("a", "y"))
		w.Counter("ironkvm_other_total", "Another.", 3)
	})

	want := `# HELP ironkvm_test A test.
# TYPE ironkvm_test gauge
ironkvm_test{a="x"} 1
ironkvm_test{a="y"} 2
# HELP ironkvm_other_total Another.
# TYPE ironkvm_other_total counter
ironkvm_other_total 3
`
	assertText(t, got, want)
}

func TestWriterWritesLabelsInTheOrderGiven(t *testing.T) {
	got := render(t, func(w *Writer) {
		w.Counter("ironkvm_test_total", "A test.", 7, L("resource", "memory"), L("kind", "some"))
	})

	want := `# HELP ironkvm_test_total A test.
# TYPE ironkvm_test_total counter
ironkvm_test_total{resource="memory",kind="some"} 7
`
	assertText(t, got, want)
}

// The format allows three escapes in a label value: backslash, double quote
// and newline. A version string or a Windows path would otherwise end the
// value early and the whole scrape would be rejected.
func TestWriterEscapesLabelValues(t *testing.T) {
	got := render(t, func(w *Writer) {
		w.Gauge("ironkvm_test", "A test.", 1, L("v", "C:\\dir \"quoted\"\nnext"))
	})

	want := "# HELP ironkvm_test A test.\n" +
		"# TYPE ironkvm_test gauge\n" +
		`ironkvm_test{v="C:\\dir \"quoted\"\nnext"} 1` + "\n"
	assertText(t, got, want)
}

// HELP escapes backslash and newline, and leaves a double quote alone.
func TestWriterEscapesHelp(t *testing.T) {
	got := render(t, func(w *Writer) {
		w.Gauge("ironkvm_test", "a\\b \"c\"\nd", 1)
	})

	want := "# HELP ironkvm_test a\\\\b \"c\"\\nd\n" +
		"# TYPE ironkvm_test gauge\n" +
		"ironkvm_test 1\n"
	assertText(t, got, want)
}

func TestWriterFormatsFractionsIntegersAndSpecialValues(t *testing.T) {
	got := render(t, func(w *Writer) {
		w.Gauge("ironkvm_test", "A test.", 0.25, L("v", "fraction"))
		w.Gauge("ironkvm_test", "A test.", 1<<40, L("v", "large"))
		w.Gauge("ironkvm_test", "A test.", math.Inf(1), L("v", "inf"))
		w.Gauge("ironkvm_test", "A test.", math.NaN(), L("v", "nan"))
	})

	want := `# HELP ironkvm_test A test.
# TYPE ironkvm_test gauge
ironkvm_test{v="fraction"} 0.25
ironkvm_test{v="large"} 1099511627776
ironkvm_test{v="inf"} +Inf
ironkvm_test{v="nan"} NaN
`
	assertText(t, got, want)
}

// A family's lines must be contiguous. Prometheus rejects a scrape that
// repeats a family after another one, so the writer refuses it and stops.
func TestWriterRefusesAFamilySplitInTwo(t *testing.T) {
	var out strings.Builder
	w := NewWriter(&out)

	w.Gauge("ironkvm_a", "A.", 1)
	w.Gauge("ironkvm_b", "B.", 2)
	w.Gauge("ironkvm_a", "A.", 3)

	if w.Err() == nil {
		t.Fatal("a family written in two places must be an error")
	}
	if strings.Contains(out.String(), "ironkvm_a 3") {
		t.Fatalf("the refused sample was written:\n%s", out.String())
	}

	w.Gauge("ironkvm_c", "C.", 4)
	if strings.Contains(out.String(), "ironkvm_c") {
		t.Fatalf("the writer kept writing after an error:\n%s", out.String())
	}
}
