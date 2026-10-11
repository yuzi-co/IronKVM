package failsafe

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// cgoReport is the head of what the Go runtime prints for a fault inside a
// cgo call into libkvm, as the test hook produces it.
const cgoReport = `fatal error: unexpected signal during runtime execution
[signal SIGSEGV: segmentation violation code=0x1 addr=0x0 pc=0x3fa8c1d2e4]

runtime stack:
runtime.throw({0x9a3b1e?, 0x0?})

goroutine 1 gp=0x3f8c000008 m=0 mp=0x1180e40 [syscall]:
runtime.cgocall(0x6c3e10, 0x3f8c0c9e58)
NanoKVM-Server/common._Cfunc_kvmv_failsafe_test_crash()
	_cgo_gotypes.go:310 +0x3c
NanoKVM-Server/common.GetKvmVision.func1()
`

const opusReport = `fatal error: unexpected signal during runtime execution
[signal SIGSEGV: segmentation violation code=0x1 addr=0x8 pc=0x3fa8c1d2e4]
runtime.cgocall(0x6c3e10, 0x3f8c0c9e58)
NanoKVM-Server/service/stream/audio._Cfunc_opus_encode(0x0)
`

// vendorLog is the end of a run on the vendor slot whose libkvm faulted, as
// the board logged it with the test hook on 2026-10-11: Sipeed's library
// catches the signal, says so, tears the pipeline down and exits 1.
const vendorLog = `--- NanoKVM-Server started, pid 2316
[2026-10-11 12:43:18.712] [warning] failsafe test hook: faulting in native code before kvmv_init
mmf_add_vi_channel..
-- [E] Trigger signal, code:SIGSEGV(11)!

maix multi-media driver destroyed.
ISP Vipipe(0) Free pa(0x8bef4000) va(0x0x3fcd55f000)
`

const goPanic =`panic: runtime error: index out of range [3] with length 3

goroutine 47 [running]:
NanoKVM-Server/service/hid.(*Hid).Write(...)
`

func TestClassify(t *testing.T) {
	cases := []struct {
		name   string
		status int
		crash  string
		log    string
		kind   Kind
		signal string
		reason string
	}{
		{"clean", 0, "", "", KindClean, "", "exited cleanly"},
		{"clean even with a stale trace", 0, "", cgoReport, KindClean, "", "exited cleanly"},
		{"a fault in a libkvm call", 2, cgoReport, "", KindVideo, "", "kvmv_failsafe_test_crash"},
		{"the report only in the log", 2, "", cgoReport, KindVideo, "", "SIGSEGV"},
		{"a native fault with no report", 139, "", "", KindVideo, "SIGSEGV", "no Go report"},
		{"an abort with no report", 134, "", "", KindVideo, "SIGABRT", "SIGABRT"},
		{"the vendor library's own handler", 1, "", vendorLog, KindVideo, "SIGSEGV", "vendor video library"},
		{"the audio encoder", 2, opusReport, "", KindNative, "", "audio encoder"},
		{"a Go panic", 2, goPanic, "", KindGo, "", "index out of range"},
		{"the OOM killer", 137, "", "", KindKilled, "SIGKILL", "SIGKILL"},
		{"a stop without a handler", 143, "", "", KindStopped, "SIGTERM", "SIGTERM"},
		{"a plain error exit", 1, "", "some line\n", KindUnknown, "", "status 1"},
		{"no status", -1, "", "", KindUnknown, "", "not known"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kind, sig, reason := Classify(c.status, c.crash, c.log)
			if kind != c.kind || sig != c.signal || !strings.Contains(reason, c.reason) {
				t.Errorf("got %s %q %q, want %s %q containing %q", kind, sig, reason, c.kind, c.signal, c.reason)
			}
		})
	}
}

func record(up int64, status int, crash, log string) string {
	return fmt.Sprintf("uptime %d\ntime 1760000000\nstatus %d\npid 42\n--- crash\n%s--- log\n%s", up, status, crash, log)
}

func TestParseRecord(t *testing.T) {
	e := parseRecord([]byte(record(812, 2, cgoReport, StartMarker+", pid 42\nstarting\n")))
	if e.Uptime != 812 || e.Time != 1760000000 || e.Status != 2 || e.Kind != KindVideo {
		t.Fatalf("parsed %+v", e)
	}
	if !strings.Contains(e.Excerpt, "_Cfunc_kvmv_failsafe_test_crash") || !strings.Contains(e.Excerpt, "starting") {
		t.Errorf("excerpt %q", e.Excerpt)
	}
}

// The log is appended to across restarts. A run that was killed must not be
// read as a video crash because the run before it was one.
func TestParseRecordReadsOnlyItsOwnRun(t *testing.T) {
	logTail := StartMarker + ", pid 41\n" + cgoReport + StartMarker + ", pid 42\nserving\n"
	e := parseRecord([]byte(record(900, 137, "", logTail)))
	if e.Kind != KindKilled {
		t.Errorf("kind %s, want killed: the earlier run's report was read", e.Kind)
	}
	if strings.Contains(e.Excerpt, "_Cfunc_") {
		t.Errorf("the excerpt holds the earlier run: %q", e.Excerpt)
	}
}

func TestParseRecordWithoutStatus(t *testing.T) {
	e := parseRecord([]byte("garbage\n--- crash\n--- log\n"))
	if e.Status != -1 || e.Kind != KindUnknown {
		t.Errorf("parsed %+v", e)
	}
}

func videoExit(up int64) Exit {
	return Exit{Uptime: up, Status: 139, Kind: KindVideo, Reason: fmt.Sprintf("crash at %d", up)}
}

func TestDecide(t *testing.T) {
	window := int64(Window.Seconds())
	cases := []struct {
		name     string
		exits    []Exit
		up       int64
		watching bool
		active   bool
		deferred bool
	}{
		{"no exits", nil, 1000, false, false, false},
		{"two crashes", []Exit{videoExit(990), videoExit(980)}, 1000, false, false, false},
		{"three crashes", []Exit{videoExit(990), videoExit(980), videoExit(970)}, 1000, false, true, false},
		{"one of three outside the window", []Exit{videoExit(990), videoExit(980), videoExit(1000 - window - 1)}, 1000, false, false, false},
		{"the oldest just inside", []Exit{videoExit(990), videoExit(980), videoExit(1000 - window)}, 1000, false, true, false},
		{"other exits do not count", []Exit{videoExit(990), {Uptime: 985, Kind: KindKilled}, videoExit(980), {Uptime: 975, Kind: KindGo}}, 1000, false, false, false},
		{"the boot watchdog still judges", []Exit{videoExit(90), videoExit(80), videoExit(70)}, 100, true, false, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			st := decide(c.exits, c.up, c.watching)
			if st.Active != c.active || (st.Deferred != "") != c.deferred {
				t.Fatalf("active %v deferred %q, want %v %v", st.Active, st.Deferred, c.active, c.deferred)
			}
			if c.active && !strings.Contains(st.Reason, "crash at 990") {
				t.Errorf("reason %q does not name the newest crash", st.Reason)
			}
			if st.Exits == nil {
				t.Error("exits is nil; the page reads a list")
			}
		})
	}
}

// scratch points the package at a temporary directory, an uptime file and an
// empty /proc.
func scratch(t *testing.T, uptime string) string {
	t.Helper()
	base := t.TempDir()
	savedDir, savedUp, savedProc := Dir, UptimeFile, ProcDir
	t.Cleanup(func() { Dir, UptimeFile, ProcDir = savedDir, savedUp, savedProc })

	Dir = filepath.Join(base, "failsafe")
	UptimeFile = filepath.Join(base, "uptime")
	ProcDir = filepath.Join(base, "proc")
	if err := os.MkdirAll(filepath.Join(Dir, "exits"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(ProcDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(UptimeFile, []byte(uptime+" 1234.56\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return base
}

func writeRecord(t *testing.T, name, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(Dir, "exits", name), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestInitStartsWithoutVideoAfterALoop(t *testing.T) {
	scratch(t, "1000.42")
	for i, up := range []int64{950, 970, 990} {
		writeRecord(t, fmt.Sprintf("%d.%d", up, 100+i), record(up, 2, cgoReport, ""))
	}

	if !Init() {
		t.Fatal("three video crashes in a minute did not start failsafe mode")
	}
	if b, err := os.ReadFile(filepath.Join(Dir, "active")); err != nil || !strings.Contains(string(b), "3 video crashes") {
		t.Errorf("marker %q, %v", b, err)
	}
	st := Current()
	if !st.Active || len(st.Exits) != 3 || st.Exits[0].Uptime != 990 {
		t.Errorf("state %+v", st)
	}
	if _, err := os.Stat(filepath.Join(Dir, fmt.Sprintf("crash.%d", os.Getpid()))); err != nil {
		t.Errorf("no crash output file for this process: %v", err)
	}

	if err := Retry(); err != nil {
		t.Fatal(err)
	}
	if Init() {
		t.Error("still in failsafe mode after a retry")
	}
	if _, err := os.Stat(filepath.Join(Dir, "active")); !os.IsNotExist(err) {
		t.Errorf("the marker survived the retry: %v", err)
	}
}

func TestInitWaitsForTheBootWatchdog(t *testing.T) {
	scratch(t, "100.00")
	for i, up := range []int64{70, 80, 90} {
		writeRecord(t, fmt.Sprintf("%d.%d", up, 100+i), record(up, 139, "", ""))
	}
	watcher := filepath.Join(ProcDir, "123")
	if err := os.MkdirAll(watcher, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(watcher, "comm"), []byte("S00awatchdog\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if Init() {
		t.Fatal("failsafe mode started while the boot watchdog judged the boot")
	}
	if Current().Deferred == "" {
		t.Error("the deferral is not reported")
	}

	if err := os.RemoveAll(watcher); err != nil {
		t.Fatal(err)
	}
	if !Init() {
		t.Error("failsafe mode did not start once the watchdog stood down")
	}
}

// A server that kvm_system started directly has no run_recorded behind it.
// Its crash report is still found, once.
func TestInitAdoptsAReportNobodyTook(t *testing.T) {
	scratch(t, "500")
	orphan := filepath.Join(Dir, "crash.99999")
	if err := os.WriteFile(orphan, []byte(cgoReport), 0o644); err != nil {
		t.Fatal(err)
	}
	empty := filepath.Join(Dir, "crash.99998")
	if err := os.WriteFile(empty, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	alive := filepath.Join(Dir, "crash.4242")
	if err := os.WriteFile(alive, []byte(cgoReport), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(ProcDir, "4242"), 0o755); err != nil {
		t.Fatal(err)
	}

	Init()

	st := Current()
	if len(st.Exits) != 1 || st.Exits[0].Kind != KindVideo || st.Exits[0].Status != -1 {
		t.Fatalf("exits %+v", st.Exits)
	}
	for _, gone := range []string{orphan, empty} {
		if _, err := os.Stat(gone); !os.IsNotExist(err) {
			t.Errorf("%s is still there", gone)
		}
	}
	if _, err := os.Stat(alive); err != nil {
		t.Errorf("the report of a live process was taken: %v", err)
	}

	Init()
	if n := len(Current().Exits); n != 1 {
		t.Errorf("%d exits after a second start, want the one", n)
	}
}

func TestInitKeepsTheNewestRecords(t *testing.T) {
	scratch(t, "5000")
	for i := 0; i < keepRecords+4; i++ {
		writeRecord(t, fmt.Sprintf("%d.%d", 100+i, i), record(int64(100+i), 137, "", ""))
	}
	Init()
	entries, err := os.ReadDir(filepath.Join(Dir, "exits"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != keepRecords {
		t.Fatalf("%d records kept, want %d", len(entries), keepRecords)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "100.") {
			t.Errorf("the oldest record %s was kept", e.Name())
		}
	}
}

func TestInitWithoutADirectoryStartsWithVideo(t *testing.T) {
	base := scratch(t, "10")
	blocker := filepath.Join(base, "file")
	if err := os.WriteFile(blocker, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	Dir = filepath.Join(blocker, "failsafe")
	if Init() {
		t.Error("an unusable directory started failsafe mode")
	}
}
