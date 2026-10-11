// Package failsafe starts the server without video after a video crash loop.
//
// The boot watchdog covers a slot that does not come up. It does not cover a
// server that crash-loops on a good slot, and the usual cause of that is the
// capture library: libkvm is native code, and a fault in it takes the whole
// process with it. S98supervise starts the server again every time, so the
// board answers for a few seconds at a time and the web UI, the update page
// and the logs are out of reach for as long as the fault lasts.
//
// The record of each exit is written by the shell that started the server.
// A process killed by a signal can record nothing itself, and the exit status
// is only known to its parent, so S95nanokvm and S98supervise both start the
// server through run_recorded (the "exit record" block they share). It writes
// one file per exit into Dir/exits: the status, the time and the last lines of
// the server's log, and the Go runtime's own crash report when there is one.
// The server asks the runtime for that report through debug.SetCrashOutput,
// into Dir/crash.<pid>, because the log holds only what fits in its tail.
//
// At start, Init reads the records and decides. Threshold exits with a video
// signature inside Window put the server in failsafe mode: libkvm is never
// initialised, every capture call answers "no frame", and the page shows a
// banner with the reason, the log excerpt and a button to try video again.
//
// The state lives in tmpfs on purpose. A reboot clears it, so a reboot always
// tries video again, and a clean exit (status 0, which is what a SIGTERM
// produces) removes the records, so a server that was stopped on purpose
// starts with video next time.
package failsafe

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// Threshold is how many video exits inside Window start the server
	// without video. Three, not two: one crash is an accident, and the second
	// can be the same accident, a restart that met the hardware still busy.
	Threshold = 3

	// Window is how far back an exit counts. S98supervise backs off from 5 to
	// 60 seconds between restarts and polls every 10, so three crashes at
	// start-up take about a minute and a half. Ten minutes also catches a
	// crash that needs a viewer to open the stream first.
	Window = 10 * time.Minute

	// keepRecords bounds the files in tmpfs. Only Threshold of them are ever
	// needed to decide; the rest are history for the banner.
	keepRecords = 8

	// excerptLines is how much of a record's log reaches the page.
	excerptLines = 30

	// TestEnv names the test hook. common/kvm_vision.go reads it right before
	// kvmv_init and faults on purpose, inside a C function, so the whole chain
	// (the fault, the record, the decision, the banner) can be driven on a
	// board without a broken library. Nothing on the board sets it: the boot
	// scripts start the server with a fixed environment.
	TestEnv = "NANOKVM_FAILSAFE_TEST"

	// StartMarker opens each run's part of the server log. See parseRecord.
	StartMarker = "--- NanoKVM-Server started"
)

// Variables so the tests can point them at a scratch directory.
var (
	// Dir holds the records. NANOKVM_FAILSAFE_DIR moves it, which a test server
	// on a board needs so that it does not read or clear the state of the
	// production server.
	Dir = envOr("NANOKVM_FAILSAFE_DIR", "/tmp/nanokvm-failsafe")

	// UptimeFile is read for the time of a decision. Records carry the uptime
	// they were written at, which no clock change after boot can move.
	UptimeFile = "/proc/uptime"

	// ProcDir is searched for the boot watchdog; see watchdogWatching.
	ProcDir = "/proc"
)

// Kind is what an exit looked like.
type Kind string

const (
	KindClean   Kind = "clean"   // status 0: stopped on purpose
	KindVideo   Kind = "video"   // died in libkvm, or in native code with no other owner
	KindNative  Kind = "native"  // died in native code that is not video (the Opus encoder)
	KindGo      Kind = "go"      // a Go panic or fatal error outside cgo
	KindKilled  Kind = "killed"  // SIGKILL: the OOM killer or the supervisor's hang cure
	KindStopped Kind = "stopped" // another signal that ends a process on request
	KindUnknown Kind = "unknown"
)

// Exit is one record, as the page shows it.
type Exit struct {
	Uptime  int64  `json:"uptime"` // seconds since boot when the server exited
	Time    int64  `json:"time"`   // unix seconds, for display only
	Status  int    `json:"status"` // the shell's exit status, -1 when not known
	Signal  string `json:"signal"` // SIGSEGV and so on, empty for a plain exit
	Kind    Kind   `json:"kind"`
	Reason  string `json:"reason"`  // one line
	Excerpt string `json:"excerpt"` // the last lines of the crash report and the log
}

// State is what the page reads.
type State struct {
	Active    bool   `json:"active"`
	Reason    string `json:"reason"`
	Deferred  string `json:"deferred,omitempty"` // why a crash loop did not start failsafe mode
	Threshold int    `json:"threshold"`
	WindowSec int64  `json:"windowSec"`
	Exits     []Exit `json:"exits"` // newest first
}

var (
	mu      sync.Mutex
	current = State{Threshold: Threshold, WindowSec: int64(Window / time.Second), Exits: []Exit{}}
)

func envOr(name, fallback string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return fallback
}

// Current is the state Init decided.
func Current() State {
	mu.Lock()
	defer mu.Unlock()

	st := current
	st.Exits = append([]Exit(nil), current.Exits...)
	return st
}

// Init reads the records, decides, and arranges for this process's own crash
// report. It returns true when the server must start without video. It never
// fails: a directory it cannot use is a server that starts with video, which
// is what it did before this package existed.
func Init() bool {
	fmt.Fprintf(os.Stderr, "%s, pid %d\n", StartMarker, os.Getpid())

	if err := os.MkdirAll(filepath.Join(Dir, "exits"), 0o755); err != nil {
		log.Warnf("failsafe: %s: %s", Dir, err)
		return false
	}

	adoptCrashReports()
	exits := readExits()
	prune()

	up, _ := readUptime()
	st := decide(exits, up, watchdogWatching())

	mu.Lock()
	current = st
	mu.Unlock()

	activeFile := filepath.Join(Dir, "active")
	if st.Active {
		_ = os.WriteFile(activeFile, []byte(st.Reason+"\n"), 0o644)
		log.Warnf("failsafe: starting without video: %s", st.Reason)
	} else {
		_ = os.Remove(activeFile)
		if st.Deferred != "" {
			log.Warnf("failsafe: %s", st.Deferred)
		}
	}

	setCrashOutput()
	return st.Active
}

// Restart starts this server again in place. main.go sets it: it tears down
// what runs and execs the binary, so the process keeps its pid and the shell
// waiting in run_recorded records nothing. Nil until then.
var Restart func()

// Retry removes every record and the marker, so the next start tries video.
func Retry() error {
	entries, err := os.ReadDir(filepath.Join(Dir, "exits"))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	for _, e := range entries {
		if err := os.Remove(filepath.Join(Dir, "exits", e.Name())); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	if err := os.Remove(filepath.Join(Dir, "active")); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// decide is the rule, with every input passed in.
//
// watching is the boot watchdog's verdict still being open. A failsafe server
// answers its API, and on a slot whose health.d hooks ask for that answer it
// would let a slot whose video crash-loops pass its check and be confirmed.
// So while the watchdog judges this boot the crash loop goes on: the hooks
// see a server that does not answer, and a trial goes back to the slot that
// was trusted. The watchdog stands down within a few seconds on a slot whose
// ssh answers and no hook asks more, and after at most its deadline (300s)
// otherwise, and the next start after that is in failsafe mode.
func decide(exits []Exit, uptime int64, watching bool) State {
	st := State{Threshold: Threshold, WindowSec: int64(Window / time.Second), Exits: exits}
	if st.Exits == nil {
		st.Exits = []Exit{}
	}

	video := 0
	var newest *Exit
	for i := range exits {
		e := &exits[i]
		if e.Kind != KindVideo {
			continue
		}
		// A record from the future is a clock this package does not use;
		// uptime only goes forward within a boot. It counts.
		if uptime-e.Uptime > int64(Window/time.Second) {
			continue
		}
		video++
		if newest == nil {
			newest = e
		}
	}

	if video < Threshold {
		return st
	}

	reason := fmt.Sprintf("%d video crashes in the last %d minutes; the last one: %s",
		video, int64(Window/time.Minute), newest.Reason)

	if watching {
		st.Deferred = "video crash loop, but the boot watchdog is still judging this boot: starting with video (" + reason + ")"
		return st
	}

	st.Active = true
	st.Reason = reason
	return st
}

// watchdogWatching reports whether S00awatchdog's watcher still runs. The
// watcher is a subshell of the init script, so its comm is the script's name.
func watchdogWatching() bool {
	entries, err := os.ReadDir(ProcDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		b, err := os.ReadFile(filepath.Join(ProcDir, e.Name(), "comm"))
		if err == nil && strings.TrimSpace(string(b)) == "S00awatchdog" {
			return true
		}
	}
	return false
}

func readUptime() (int64, error) {
	b, err := os.ReadFile(UptimeFile)
	if err != nil {
		return 0, err
	}
	field, _, _ := strings.Cut(strings.TrimSpace(string(b)), " ")
	whole, _, _ := strings.Cut(field, ".")
	return strconv.ParseInt(whole, 10, 64)
}

// setCrashOutput sends the runtime's report of a fatal error to
// Dir/crash.<pid> as well as to stderr. A fault inside a cgo call is one: the
// report names the C function, which is how a libkvm crash is told from any
// other. run_recorded takes the file into the exit record.
func setCrashOutput() {
	name := filepath.Join(Dir, "crash."+strconv.Itoa(os.Getpid()))
	f, err := os.OpenFile(name, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		log.Warnf("failsafe: crash output: %s", err)
		return
	}
	if err := debug.SetCrashOutput(f, debug.CrashOptions{}); err != nil {
		log.Warnf("failsafe: crash output: %s", err)
	}
	// SetCrashOutput keeps its own duplicate of the descriptor.
	_ = f.Close()
}

// adoptCrashReports turns the crash report of a server that ran without
// run_recorded into a record of its own. kvm_system starts the server
// directly after an update, and nothing waits for that one. The exit status
// is not known then, but a report that names libkvm is enough to classify.
//
// The file is renamed before it is read, and run_recorded does the same, so
// one exit is never recorded twice. A report of a process that still runs,
// this one's predecessor still on its way out included, is left alone.
func adoptCrashReports() {
	matches, _ := filepath.Glob(filepath.Join(Dir, "crash.*"))
	up, _ := readUptime()
	for _, path := range matches {
		pid, err := strconv.Atoi(strings.TrimPrefix(filepath.Base(path), "crash."))
		if err != nil || pid == os.Getpid() {
			continue
		}
		if _, err := os.Stat(filepath.Join(ProcDir, strconv.Itoa(pid))); err == nil {
			continue
		}
		taken := path + ".adopted"
		if os.Rename(path, taken) != nil {
			continue
		}
		report, _ := os.ReadFile(taken)
		_ = os.Remove(taken)
		if len(bytes.TrimSpace(report)) == 0 {
			continue
		}
		var b bytes.Buffer
		fmt.Fprintf(&b, "uptime %d\ntime %d\nstatus -1\npid %d\n--- crash\n", up, time.Now().Unix(), pid)
		b.Write(report)
		name := filepath.Join(Dir, "exits", fmt.Sprintf("%d.%d", up, pid))
		_ = os.WriteFile(name, b.Bytes(), 0o644)
	}
}

// readExits reads every record, newest first.
func readExits() []Exit {
	dir := filepath.Join(Dir, "exits")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	exits := []Exit{}
	for _, e := range entries {
		if e.IsDir() || strings.HasSuffix(e.Name(), ".new") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		exits = append(exits, parseRecord(b))
	}
	sort.SliceStable(exits, func(i, j int) bool { return exits[i].Uptime > exits[j].Uptime })
	return exits
}

// prune keeps the newest keepRecords files.
func prune() {
	dir := filepath.Join(Dir, "exits")
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) <= keepRecords {
		return
	}
	type rec struct {
		name string
		up   int64
	}
	var recs []rec
	for _, e := range entries {
		upStr, _, _ := strings.Cut(e.Name(), ".")
		up, _ := strconv.ParseInt(upStr, 10, 64)
		recs = append(recs, rec{e.Name(), up})
	}
	sort.Slice(recs, func(i, j int) bool { return recs[i].up > recs[j].up })
	for _, r := range recs[keepRecords:] {
		_ = os.Remove(filepath.Join(dir, r.name))
	}
}

// parseRecord reads what run_recorded writes:
//
//	uptime 812
//	time 1760180000
//	status 139
//	pid 1234
//	--- crash
//	<the runtime's report, may be empty>
//	--- log
//	<the last lines of /tmp/nanokvm-server.log>
func parseRecord(b []byte) Exit {
	e := Exit{Status: -1}
	var crash, logTail strings.Builder
	section := ""

	sc := bufio.NewScanner(bytes.NewReader(b))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		switch line {
		case "--- crash":
			section = "crash"
			continue
		case "--- log":
			section = "log"
			continue
		}
		switch section {
		case "crash":
			crash.WriteString(line + "\n")
		case "log":
			logTail.WriteString(line + "\n")
		default:
			key, value, _ := strings.Cut(line, " ")
			n, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err != nil {
				continue
			}
			switch key {
			case "uptime":
				e.Uptime = n
			case "time":
				e.Time = n
			case "status":
				e.Status = int(n)
			}
		}
	}

	// The log is appended to across restarts, so its tail can still hold the
	// report of the run before. Only what this run printed counts.
	tail := logTail.String()
	if i := strings.LastIndex(tail, StartMarker); i >= 0 {
		tail = tail[i:]
	}

	e.Kind, e.Signal, e.Reason = Classify(e.Status, crash.String(), tail)
	e.Excerpt = excerpt(crash.String(), tail)
	return e
}

// Native code in this server: libkvm through common/kvm_vision.go, and the
// Opus encoder through service/stream/audio. A report names the cgo stub of
// the C function it died in, _Cfunc_<name>, and the Go frames around it.
var (
	videoMarkers = []string{
		"_Cfunc_kvmv_", "_Cfunc_free_kvmv_data", "_Cfunc_set_h264_", "_Cfunc_set_capture_fps",
		"_Cfunc_set_frame_detact", "common.(*KvmVision)", "common.GetKvmVision",
	}
	otherNativeMarkers = []string{"_Cfunc_opus", "service/stream/audio"}
)

// signals is the subset of Linux signal numbers an exit status can name.
var signals = map[int]string{
	1: "SIGHUP", 2: "SIGINT", 3: "SIGQUIT", 4: "SIGILL", 5: "SIGTRAP", 6: "SIGABRT",
	7: "SIGBUS", 8: "SIGFPE", 9: "SIGKILL", 11: "SIGSEGV", 13: "SIGPIPE", 15: "SIGTERM",
}

// nativeFaults are the signals native code dies of.
var nativeFaults = map[string]bool{"SIGILL": true, "SIGABRT": true, "SIGBUS": true, "SIGFPE": true, "SIGSEGV": true, "SIGTRAP": true}

// Classify says what an exit was. status is the shell's: 128+n for a process
// killed by signal n, -1 when not known.
//
// A fault the Go runtime catches ends in its report and exit status 2. A
// fault it does not catch, one on a thread libkvm created itself, ends in the
// signal with no report at all. libkvm is the only native code that runs
// threads of its own, so a native fault with no report is counted as video.
func Classify(status int, crash, logTail string) (Kind, string, string) {
	sig := ""
	if status > 128 && status < 160 {
		sig = signals[status-128]
		if sig == "" {
			sig = "signal " + strconv.Itoa(status-128)
		}
	}

	report := crash
	if strings.TrimSpace(report) == "" {
		report = logTail
	}
	text := crash + "\n" + logTail

	if status == 0 {
		return KindClean, "", "exited cleanly"
	}
	if fn := firstMarker(text, videoMarkers); fn != "" {
		return KindVideo, sig, "died in the video library (" + describeFault(report, sig) + ", in " + cfunc(text, fn) + ")"
	}
	if name, ok := vendorHandlerSignal(text); ok {
		return KindVideo, name, "killed by " + name + " in native code; the vendor video library caught it and exited"
	}
	if fn := firstMarker(text, otherNativeMarkers); fn != "" && strings.Contains(text, "[signal ") {
		return KindNative, sig, "died in the audio encoder (" + describeFault(report, sig) + ")"
	}

	switch {
	case nativeFaults[sig]:
		return KindVideo, sig, "killed by " + sig + " outside Go code, with no Go report (libkvm runs the only native threads)"
	case sig == "SIGKILL":
		return KindKilled, sig, "killed by SIGKILL (out of memory, or a supervisor hang cure)"
	case sig != "":
		return KindStopped, sig, "ended by " + sig
	case strings.Contains(report, "runtime.cgocall") && strings.Contains(report, "[signal "):
		// A fault inside a cgo call whose C function the report does not
		// name. The only other native code names itself above.
		return KindVideo, sig, "died in native code (" + describeFault(report, sig) + ")"
	case strings.Contains(report, "panic: ") || strings.Contains(report, "fatal error: "):
		return KindGo, sig, "Go " + firstLine(report, "panic: ", "fatal error: ")
	case status > 0:
		return KindUnknown, sig, "exited with status " + strconv.Itoa(status)
	}
	return KindUnknown, sig, "exit status not known"
}

// vendorHandlerMarker is what Sipeed's library prints from the handler it
// installs for SIGSEGV, SIGBUS, SIGABRT and the other faults, from a
// constructor, at load (maix_camera_mmf.hpp, maix_vision_register_signal):
//
//	-- [E] Trigger signal, code:SIGSEGV(11)!
//
// The handler deinitialises the pipeline and calls exit(1). The Go runtime
// forwards a fault in non-Go code to a handler that was there before its own,
// so on the vendor slot a libkvm crash ends in status 1 and this line, with no
// Go report. Measured on the board with the test hook, 2026-10-11.
// libkvm-v4l2 installs no handler, and there the runtime's report says it.
const vendorHandlerMarker = "Trigger signal, code:"

func vendorHandlerSignal(text string) (string, bool) {
	i := strings.LastIndex(text, vendorHandlerMarker)
	if i < 0 {
		return "", false
	}
	name, _, _ := strings.Cut(text[i+len(vendorHandlerMarker):], "(")
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, " \n!") {
		name = "a fault"
	}
	return name, true
}

func firstMarker(text string, markers []string) string {
	for _, m := range markers {
		if strings.Contains(text, m) {
			return m
		}
	}
	return ""
}

// cfunc names the C function from the report, kvmv_read_img for
// "main._Cfunc_kvmv_read_img(...)", or the marker itself.
func cfunc(text, marker string) string {
	i := strings.Index(text, "_Cfunc_")
	if i < 0 {
		return marker
	}
	rest := text[i+len("_Cfunc_"):]
	end := strings.IndexAny(rest, "( \n\t")
	if end < 0 {
		end = len(rest)
	}
	return rest[:end]
}

func describeFault(report, sig string) string {
	if i := strings.Index(report, "[signal "); i >= 0 {
		rest := report[i+len("[signal "):]
		name, _, _ := strings.Cut(rest, ":")
		if name, _, _ = strings.Cut(name, " "); name != "" {
			return name
		}
	}
	if sig != "" {
		return sig
	}
	if l := firstLine(report, "fatal error: ", "panic: "); l != "" {
		return l
	}
	return "a fault"
}

// firstLine is the first line holding one of the prefixes, from the prefix on.
func firstLine(text string, prefixes ...string) string {
	for _, line := range strings.Split(text, "\n") {
		for _, p := range prefixes {
			if i := strings.Index(line, p); i >= 0 {
				return strings.TrimSpace(line[i:])
			}
		}
	}
	return ""
}

// excerpt is what the banner shows: the head of the crash report, which names
// the fault and the function, and then the last lines of the log.
func excerpt(crash, logTail string) string {
	var parts []string
	if c := strings.TrimSpace(crash); c != "" {
		lines := strings.Split(c, "\n")
		if len(lines) > 20 {
			lines = lines[:20]
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}
	if l := strings.TrimSpace(logTail); l != "" {
		lines := strings.Split(l, "\n")
		if len(lines) > excerptLines {
			lines = lines[len(lines)-excerptLines:]
		}
		parts = append(parts, strings.Join(lines, "\n"))
	}
	return strings.Join(parts, "\n...\n")
}
