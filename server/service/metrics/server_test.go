package metrics

import (
	"strings"
	"testing"
	"time"
)

func useServerSources(t *testing.T) (proc string) {
	t.Helper()

	proc = t.TempDir()
	setVar(t, &procDir, proc)
	setVar(t, &processStart, time.Unix(1790000000, 0))
	setVar(t, &goRuntime, func() (uint64, int) { return 1048576, 42 })

	return proc
}

// The image version is a real one: an IronKVM card names its base in
// parentheses, so the label value carries spaces and brackets.
func TestServerWritesEveryFamily(t *testing.T) {
	proc := useServerSources(t)
	writeFixture(t, proc, "self/status", "Name:\tNanoKVM-Server\nVmRSS:\t   40960 kB\n")
	setVar(t, &versions, func() (string, string, string) {
		return "ironkvm-1.2.0 (based on v1.4.3)", "5.10.270-ironkvm0", "2.4.3"
	})

	want := `# HELP ironkvm_build_info Versions of the card image, the kernel and the server. Always 1.
# TYPE ironkvm_build_info gauge
ironkvm_build_info{image="ironkvm-1.2.0 (based on v1.4.3)",kernel="5.10.270-ironkvm0",app="2.4.3"} 1
# HELP ironkvm_process_start_time_seconds When this server process started, in Unix seconds.
# TYPE ironkvm_process_start_time_seconds gauge
ironkvm_process_start_time_seconds 1790000000
# HELP ironkvm_process_resident_bytes Resident memory of this server process, from /proc/self/status VmRSS.
# TYPE ironkvm_process_resident_bytes gauge
ironkvm_process_resident_bytes 41943040
# HELP ironkvm_go_heap_bytes Bytes of allocated Go heap objects.
# TYPE ironkvm_go_heap_bytes gauge
ironkvm_go_heap_bytes 1048576
# HELP ironkvm_go_goroutines Goroutines in this server process.
# TYPE ironkvm_go_goroutines gauge
ironkvm_go_goroutines 42
`
	assertText(t, render(t, collectServer), want)
}

// A version file that holds a quote must not end the label early.
func TestServerEscapesAVersionWithAQuote(t *testing.T) {
	useServerSources(t)
	setVar(t, &versions, func() (string, string, string) { return `v1 "beta"`, "", "" })

	got := render(t, collectServer)
	if !strings.Contains(got, `ironkvm_build_info{image="v1 \"beta\"",kernel="",app=""} 1`) {
		t.Fatalf("the version was not escaped:\n%s", got)
	}
}

func TestServerLeavesOutRSSWhenStatusCannotBeRead(t *testing.T) {
	useServerSources(t)
	setVar(t, &versions, func() (string, string, string) { return "", "", "" })

	got := render(t, collectServer)
	if strings.Contains(got, "ironkvm_process_resident_bytes") {
		t.Fatalf("RSS was written without /proc/self/status:\n%s", got)
	}
	if !strings.Contains(got, "ironkvm_go_goroutines 42\n") {
		t.Fatalf("the rest of the section went with it:\n%s", got)
	}
}

// The version files change only across a restart, so they are read once.
func TestOnceVersionsReadsTheVersionsOnce(t *testing.T) {
	calls := 0
	read := onceVersions(func() (string, string, string) {
		calls++
		return "v1.4.3", "5.10.270-ironkvm0", "2.4.3"
	})

	for range 3 {
		image, kernel, app := read()
		if image != "v1.4.3" || kernel != "5.10.270-ironkvm0" || app != "2.4.3" {
			t.Fatalf("read() = %q, %q, %q", image, kernel, app)
		}
	}
	if calls != 1 {
		t.Fatalf("the versions were read %d times, want 1", calls)
	}
}
