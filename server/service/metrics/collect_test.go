package metrics

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func sectionWriting(name string, value float64) section {
	return section{name: name, collect: func(w *Writer) {
		w.Gauge("ironkvm_"+name, "Section "+name+".", value)
	}}
}

func collectString(t *testing.T) string {
	t.Helper()

	var out strings.Builder
	if err := Collect(&out); err != nil {
		t.Fatalf("Collect: %s", err)
	}

	return out.String()
}

func TestCollectWritesEverySectionInOrder(t *testing.T) {
	setVar(t, &sections, []section{sectionWriting("a", 1), sectionWriting("b", 2)})

	want := `# HELP ironkvm_a Section a.
# TYPE ironkvm_a gauge
ironkvm_a 1
# HELP ironkvm_b Section b.
# TYPE ironkvm_b gauge
ironkvm_b 2
`
	assertText(t, collectString(t), want)
}

// A bug in one collector costs that section, not the scrape. What it wrote
// before it panicked is thrown away with it, so no half family is served.
func TestCollectLeavesOutASectionThatPanics(t *testing.T) {
	setVar(t, &sections, []section{
		sectionWriting("a", 1),
		{name: "broken", collect: func(w *Writer) {
			w.Gauge("ironkvm_broken", "Broken.", 1)
			panic("collector bug")
		}},
		sectionWriting("c", 3),
	})

	got := collectString(t)
	if strings.Contains(got, "ironkvm_broken") {
		t.Fatalf("the panicking section's output was served:\n%s", got)
	}
	if !strings.Contains(got, "ironkvm_a 1\n") || !strings.Contains(got, "ironkvm_c 3\n") {
		t.Fatalf("the other sections were lost:\n%s", got)
	}
}

// Two sections that write the same family would produce a scrape Prometheus
// rejects whole. The second one is dropped instead.
func TestCollectLeavesOutASectionThatRepeatsAFamily(t *testing.T) {
	setVar(t, &sections, []section{
		sectionWriting("a", 1),
		{name: "repeat", collect: func(w *Writer) {
			w.Gauge("ironkvm_a", "Section a.", 9)
		}},
		sectionWriting("c", 3),
	})

	got := collectString(t)
	if strings.Contains(got, "ironkvm_a 9") {
		t.Fatalf("the repeated family was served:\n%s", got)
	}
	if !strings.Contains(got, "ironkvm_c 3\n") {
		t.Fatalf("the section after it was lost:\n%s", got)
	}
}

// useBareHost runs the real sections against a host with no /proc or /sys
// fixtures, no ION and no UDC. The frame counter is the real one, never
// started.
func useBareHost(t *testing.T) {
	t.Helper()

	root := t.TempDir()
	setVar(t, &procDir, root+"/proc")
	setVar(t, &zramDir, root+"/zram0")
	setVar(t, &cgroupDir, root+"/cgroup")
	setVar(t, &readIon, func() (uint64, uint64, bool) { return 0, 0, false })
	setVar(t, &usbLinkState, func() (string, error) { return "", errors.New("no controller") })
}

// Every missing source drops its own families, and the scrape still serves the
// ones that cannot be missing.
func TestCollectServesTheRestOnABareHost(t *testing.T) {
	useBareHost(t)

	got := collectString(t)
	for _, present := range []string{
		`ironkvm_stream_viewers{path="mjpeg"} `,
		`ironkvm_hid_endpoint_state{endpoint="keyboard",state="unknown"} 1`,
		`ironkvm_usb_recoveries_total{action="rebind"} `,
		`ironkvm_build_info{`,
		`ironkvm_go_goroutines `,
	} {
		if !strings.Contains(got, present) {
			t.Errorf("missing %q", present)
		}
	}
	for _, absent := range []string{
		"ironkvm_memory_bytes", "ironkvm_zram_bytes", "ironkvm_pressure_stall_seconds_total",
		"ironkvm_cgroup_memory_bytes", "ironkvm_ion_bytes", "ironkvm_usb_udc_state",
		"ironkvm_process_resident_bytes",
	} {
		if strings.Contains(got, absent) {
			t.Errorf("%s was written without its source", absent)
		}
	}
}

// Two Prometheus replicas scrape at once. Run with -race.
func TestCollectIsSafeToRunConcurrently(t *testing.T) {
	useBareHost(t)

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			var out strings.Builder
			if err := Collect(&out); err != nil {
				t.Errorf("Collect: %s", err)
			}
		}()
	}
	wg.Wait()
}
