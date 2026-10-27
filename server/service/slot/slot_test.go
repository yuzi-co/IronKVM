package slot

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/service/audit"
)

// What `slot status` prints on a board running slot a, with b installed and a
// trial of it armed, in the tool's own layout (generic/slot, cmd_status).
const statusArmed = `running    a
trusted    a
trial      b
prev       none
recovery   no
bootloader installed
kernel     slot
slot-a     unknown
slot-b     installed version=2026.10.27 sha256=abc installed=2026-10-27T10:00:00Z trial=armed
slot-recovery installed

  device lichee-rv-nano
  vermagic 5.10.4-tag- mod_unload riscv
  slotkernel 6.12.0
  flavour none
  libkvm libkvm-vendor.so 0123
`

// stub stands a shell script in for the slot tool. It records every argument
// list it is given in dir/args, prints dir/status for status, and installs as
// dir/mode says: "fail" refuses as the tool does, "hang" stops in the write
// until dir/release exists, and anything else finishes.
type stub struct {
	dir string
}

func newStub(t *testing.T) *stub {
	t.Helper()
	dir := t.TempDir()
	s := &stub{dir: dir}
	script := fmt.Sprintf(`#!/bin/sh
D=%q
echo "$*" >> "$D/args"
case "$1" in
status)
	# An older tool ignores --json and prints the text form, which is what
	# this does until a case writes status.json.
	if [ "$2" = --json ] && [ -e "$D/status.json" ]; then
		cat "$D/status.warn" >&2 2>/dev/null
		cat "$D/status.json"
	else
		cat "$D/status"
	fi
	;;
install)
	shift
	progress=
	while [ $# -gt 0 ]; do
		case "$1" in
		--progress) progress=$2; shift 2 ;;
		--force) shift ;;
		--) shift; break ;;
		*) break ;;
		esac
	done
	echo "phase=verify done=0 total=0" > "$progress"
	case "$(cat "$D/mode" 2>/dev/null)" in
	fail)
		echo "slot: no key in the keyring verifies $2.manifest" >&2
		echo "phase=failed done=0 total=0 message=no key in the keyring verifies $2.manifest" > "$progress"
		exit 1
		;;
	hang)
		echo "phase=write done=1048576 total=4194304" > "$progress"
		echo "slot: writing $2"
		while [ ! -e "$D/release" ]; do sleep 0.05; done
		;;
	esac
	echo "phase=done done=4194304 total=4194304" > "$progress"
	echo "slot: $1 installed and verified (4194304 bytes)"
	;;
try|confirm|revert)
	if [ -e "$D/refuse" ]; then
		echo "slot: $(cat "$D/refuse")" >&2
		exit 1
	fi
	echo "slot: done"
	;;
esac
`, dir)
	bin := filepath.Join(dir, "slot")
	if err := os.WriteFile(bin, []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	s.write(t, "status", statusArmed)

	data := filepath.Join(dir, "data")
	if err := os.Mkdir(data, 0o755); err != nil {
		t.Fatal(err)
	}

	saved := []string{Bin, ImageRoot, ProgressFile, LogFile, JobFile}
	Bin = bin
	ImageRoot = data
	ProgressFile = filepath.Join(dir, "progress")
	LogFile = filepath.Join(dir, "log")
	JobFile = filepath.Join(dir, "job")
	t.Cleanup(func() {
		Bin, ImageRoot, ProgressFile, LogFile, JobFile = saved[0], saved[1], saved[2], saved[3], saved[4]
		// A hung stub is let go, so no test leaves a process behind.
		s.write(t, "release", "")
	})
	return s
}

func (s *stub) write(t *testing.T, name, content string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(s.dir, name), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (s *stub) args(t *testing.T) []string {
	t.Helper()
	data, _ := os.ReadFile(filepath.Join(s.dir, "args"))
	return strings.Split(strings.TrimSpace(string(data)), "\n")
}

func (s *stub) image(t *testing.T, name string, beside ...string) string {
	t.Helper()
	path := filepath.Join(ImageRoot, name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("image"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, suffix := range beside {
		content := "sig"
		if suffix == ".manifest" {
			content = "ironkvm-slot-manifest 1\ndevice    lichee-rv-nano\nversion   2026.10.27\n"
		}
		if err := os.WriteFile(path+suffix, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return path
}

func TestParseStatusReadsTheSlotsAndTheRunningImage(t *testing.T) {
	st := parseStatus(statusArmed)
	if st.Running != "a" || st.Trusted != "a" || st.Trial != "b" || st.Prev != "none" ||
		st.Recovery != "no" || st.Bootloader != "installed" || st.Kernel != "slot" {
		t.Fatalf("status lines: %+v", st)
	}
	if len(st.Slots) != 3 {
		t.Fatalf("slots: %+v", st.Slots)
	}
	b, _ := st.Slot("b")
	want := Info{Name: "b", State: "installed", Version: "2026.10.27", Sha256: "abc",
		Installed: "2026-10-27T10:00:00Z", Trial: "armed"}
	if b != want {
		t.Fatalf("slot b: %+v", b)
	}
	if a, _ := st.Slot("a"); a.State != "unknown" {
		t.Fatalf("slot a: %+v", a)
	}
	// The recovery line is the slot's record, not the backend's "recovery no".
	if r, _ := st.Slot("recovery"); r.State != "installed" {
		t.Fatalf("slot recovery: %+v", r)
	}
	if st.Image["device"] != "lichee-rv-nano" || st.Image["slotkernel"] != "6.12.0" {
		t.Fatalf("image: %v", st.Image)
	}
	if _, ok := st.Image["vermagic"]; ok {
		t.Fatal("the module line is not shown")
	}
	if st.Installing() != "" {
		t.Fatal("nothing is being installed")
	}
}

func TestParseStatusOfABackendWithoutItsOwnLines(t *testing.T) {
	st := parseStatus("running    b\ntrusted    b\nslot-a     installing\nslot-b     installed\n")
	if st.Running != "b" || st.Trial != "" || st.Prev != "" || st.Installing() != "a" {
		t.Fatalf("%+v", st)
	}
}

func TestParseProgress(t *testing.T) {
	p := parseProgress("phase=write done=1048576 total=4194304\n")
	if p.Phase != "write" || p.Done != 1048576 || p.Total != 4194304 || p.Message != "" {
		t.Fatalf("%+v", p)
	}
	p = parseProgress("phase=failed done=0 total=0 message=image is 9 bytes, /dev/x holds 4 done=7\n")
	if p.Phase != "failed" || p.Done != 0 || p.Message != "image is 9 bytes, /dev/x holds 4 done=7" {
		t.Fatalf("the message runs to the end of the line: %+v", p)
	}
}

func TestValidateURL(t *testing.T) {
	for _, ok := range []string{
		"https://github.com/yuzi-co/ironkvm-dist/releases/download/v1/slot-b.img.zst",
		"https://10.0.30.5:8443/slot-b.img",
	} {
		if _, err := validateURL(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{
		"", "http://example.com/slot-b.img.zst", "ftp://example.com/x", "https:///x",
		"https://example.com/a b", "https://example.com/x\n--unsigned", "/data/slot-b.img",
		"https:example.com", "https://" + strings.Repeat("a", maxURLLength),
	} {
		if _, err := validateURL(bad); err == nil {
			t.Errorf("%q was accepted", bad)
		}
	}
}

func TestDisplaySourceDropsCredentials(t *testing.T) {
	got := displaySource("https://user:secret@example.com/r/slot-b.img.zst?token=abc#x")
	if got != "https://example.com/r/slot-b.img.zst" {
		t.Fatal(got)
	}
	if got := displaySource("/data/slot-b.img.zst"); got != "/data/slot-b.img.zst" {
		t.Fatal(got)
	}
}

func TestResolveFileKeepsToImagesOnData(t *testing.T) {
	s := newStub(t)
	good := s.image(t, "slot-b.img.zst")
	if got, err := resolveFile(good); err != nil || got != mustEval(t, good) {
		t.Fatalf("%s: %s %v", good, got, err)
	}
	if _, err := resolveFile(filepath.Join(ImageRoot, "sub", "..", "slot-b.img.zst")); err != nil {
		t.Fatalf("a path with .. inside /data: %v", err)
	}

	outside := filepath.Join(s.dir, "outside.img")
	if err := os.WriteFile(outside, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(ImageRoot, "link.img")
	if err := os.Symlink(outside, link); err != nil {
		t.Fatal(err)
	}
	notImage := s.image(t, "notes.txt")
	for _, bad := range []string{
		outside, link, notImage, "slot-b.img.zst", filepath.Join(ImageRoot, "missing.img"),
		filepath.Join(ImageRoot, "..", "outside.img"), ImageRoot,
	} {
		if _, err := resolveFile(bad); err == nil {
			t.Errorf("%s was accepted", bad)
		}
	}
}

func mustEval(t *testing.T, path string) string {
	t.Helper()
	got, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestListImagesFindsSlotImagesAndNotDiskImages(t *testing.T) {
	s := newStub(t)
	s.image(t, "slot-b.img.zst", ".manifest", ".manifest.minisig")
	s.image(t, "builds/slot-a.img", ".minisig")
	s.image(t, "builds/deep/slot-x.img.zst")
	s.image(t, "usb.img")
	s.image(t, "win11.iso")
	s.image(t, "plain.img.zst")

	images := listImages()
	got := map[string]Image{}
	for _, img := range images {
		rel, _ := filepath.Rel(mustEval(t, ImageRoot), img.Path)
		got[rel] = img
	}
	if len(got) != 3 {
		t.Fatalf("found %v", got)
	}
	b := got["slot-b.img.zst"]
	if !b.Signed || b.Version != "2026.10.27" || b.Device != "lichee-rv-nano" {
		t.Fatalf("slot-b: %+v", b)
	}
	if !got[filepath.Join("builds", "slot-a.img")].Signed {
		t.Fatal("slot-a has a signature beside it")
	}
	if p := got["plain.img.zst"]; p.Signed || p.Version != "" {
		t.Fatalf("plain: %+v", p)
	}
}

// api builds the routes without the gate, which router tests cover.
func api(t *testing.T) (*gin.Engine, *Service, chan bool) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	svc := NewService()
	ended := make(chan bool, 4)
	svc.install.onExit = func(_ Job, ok bool, _ string) { ended <- ok }
	r := gin.New()
	r.GET("/api/slot", svc.GetStatus)
	r.GET("/api/slot/images", svc.GetImages)
	r.GET("/api/slot/install", svc.GetProgress)
	r.POST("/api/slot/install", svc.Install)
	r.POST("/api/slot/try", svc.Try)
	r.POST("/api/slot/confirm", svc.Confirm)
	r.POST("/api/slot/revert", svc.Revert)
	return r, svc, ended
}

type reply struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func call(t *testing.T, r *gin.Engine, method, path, body string) reply {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s: HTTP %d", method, path, w.Code)
	}
	var rsp reply
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("%s %s: %s", method, path, w.Body.String())
	}
	return rsp
}

func waitEnd(t *testing.T, ended chan bool) bool {
	t.Helper()
	select {
	case ok := <-ended:
		return ok
	case <-time.After(10 * time.Second):
		t.Fatal("the install never ended")
		return false
	}
}

func progressOf(t *testing.T, r *gin.Engine) Progress {
	t.Helper()
	rsp := call(t, r, http.MethodGet, "/api/slot/install", "")
	var p Progress
	if err := json.Unmarshal(rsp.Data, &p); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestInstallFromAURLRunsTheToolWithoutUnsigned(t *testing.T) {
	s := newStub(t)
	r, _, ended := api(t)

	if p := progressOf(t, r); p.Phase != "" || p.Running {
		t.Fatalf("before any install: %+v", p)
	}

	url := "https://example.com/r/slot-b.img.zst?token=secret"
	rsp := call(t, r, http.MethodPost, "/api/slot/install", `{"slot":"b","url":"`+url+`"}`)
	if rsp.Code != 0 {
		t.Fatalf("install: %+v", rsp)
	}
	if !waitEnd(t, ended) {
		t.Fatal("the install failed")
	}

	args := s.args(t)
	last := args[len(args)-1]
	if last != "install --progress "+ProgressFile+" -- b "+url {
		t.Fatalf("the tool was run as %q", last)
	}
	for _, a := range args {
		if strings.Contains(a, "--unsigned") {
			t.Fatalf("--unsigned was passed: %q", a)
		}
	}

	p := progressOf(t, r)
	if p.Running || p.Phase != "done" || p.Done != 4194304 || p.Total != 4194304 || p.Slot != "b" {
		t.Fatalf("after: %+v", p)
	}
	if p.Source != "https://example.com/r/slot-b.img.zst" {
		t.Fatalf("the source shows its query: %s", p.Source)
	}
}

func TestInstallFromAFileOnDataWithForce(t *testing.T) {
	s := newStub(t)
	r, _, ended := api(t)
	img := s.image(t, "slot-b.img.zst", ".manifest")

	rsp := call(t, r, http.MethodPost, "/api/slot/install", `{"slot":"b","file":"`+img+`","force":true}`)
	if rsp.Code != 0 {
		t.Fatalf("install: %+v", rsp)
	}
	waitEnd(t, ended)
	args := s.args(t)
	if last := args[len(args)-1]; last != "install --progress "+ProgressFile+" --force -- b "+mustEval(t, img) {
		t.Fatalf("the tool was run as %q", last)
	}
}

func TestInstallRefusals(t *testing.T) {
	s := newStub(t)
	r, _, _ := api(t)
	img := s.image(t, "slot-b.img.zst")

	cases := map[string]string{
		"the running slot":    `{"slot":"a","file":"` + img + `"}`,
		"an http URL":         `{"slot":"b","url":"http://example.com/slot-b.img.zst"}`,
		"both sources":        `{"slot":"b","file":"` + img + `","url":"https://example.com/x.img"}`,
		"no source":           `{"slot":"b"}`,
		"no such slot name":   `{"slot":"c","file":"` + img + `"}`,
		"a file off /data":    `{"slot":"b","file":"/etc/passwd"}`,
		"an option as a path": `{"slot":"b","file":"--unsigned"}`,
	}
	for name, body := range cases {
		if rsp := call(t, r, http.MethodPost, "/api/slot/install", body); rsp.Code == 0 {
			t.Errorf("%s was accepted", name)
		}
	}
	for _, a := range s.args(t) {
		if strings.HasPrefix(a, "install") {
			t.Fatalf("a refused install ran the tool: %q", a)
		}
	}

	// A slot the board does not have.
	s.write(t, "status", "running    a\ntrusted    a\nslot-a     installed\nslot-b     installed\n")
	if rsp := call(t, r, http.MethodPost, "/api/slot/install", `{"slot":"recovery","file":"`+img+`"}`); rsp.Code == 0 {
		t.Error("an install onto a missing recovery slot was accepted")
	}

	// An install the records show, started from a shell.
	s.write(t, "status", "running    a\ntrusted    a\nslot-b     installing\n")
	if rsp := call(t, r, http.MethodPost, "/api/slot/install", `{"slot":"b","file":"`+img+`"}`); rsp.Code != codeBusy {
		t.Errorf("an install over one in progress: %+v", rsp)
	}
}

func TestOneInstallAtATimeAndAcrossARestart(t *testing.T) {
	s := newStub(t)
	s.write(t, "mode", "hang")
	r, _, ended := api(t)
	img := s.image(t, "slot-b.img.zst")
	body := `{"slot":"b","file":"` + img + `"}`

	if rsp := call(t, r, http.MethodPost, "/api/slot/install", body); rsp.Code != 0 {
		t.Fatalf("install: %+v", rsp)
	}
	waitFor(t, func() bool { return progressOf(t, r).Phase == "write" })

	if rsp := call(t, r, http.MethodPost, "/api/slot/install", body); rsp.Code != codeBusy {
		t.Fatalf("a second install: %+v", rsp)
	}
	p := progressOf(t, r)
	if !p.Running || p.Done != 1048576 || p.Total != 4194304 {
		t.Fatalf("while writing: %+v", p)
	}

	// A server restarted during the install has none of this one's memory,
	// and still sees the install through the job file.
	restarted, _, _ := api(t)
	if rsp := call(t, restarted, http.MethodPost, "/api/slot/install", body); rsp.Code != codeBusy {
		t.Fatalf("an install after a restart: %+v", rsp)
	}
	if p := progressOf(t, restarted); !p.Running || p.Phase != "write" || p.Slot != "b" {
		t.Fatalf("after a restart: %+v", p)
	}

	s.write(t, "release", "")
	if !waitEnd(t, ended) {
		t.Fatal("the install failed")
	}
	if p := progressOf(t, restarted); p.Running || p.Phase != "done" {
		t.Fatalf("after the end, seen by the restarted server: %+v", p)
	}
}

func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for !cond() {
		if time.Now().After(deadline) {
			t.Fatal("timed out")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestAFailedInstallSaysWhy(t *testing.T) {
	s := newStub(t)
	s.write(t, "mode", "fail")
	r, _, ended := api(t)
	img := s.image(t, "slot-b.img.zst")

	if rsp := call(t, r, http.MethodPost, "/api/slot/install", `{"slot":"b","file":"`+img+`"}`); rsp.Code != 0 {
		t.Fatalf("install: %+v", rsp)
	}
	if waitEnd(t, ended) {
		t.Fatal("a refused install was reported as done")
	}
	p := progressOf(t, r)
	if p.Running || p.Phase != "failed" || !strings.HasPrefix(p.Message, "no key in the keyring verifies") {
		t.Fatalf("%+v", p)
	}
}

func TestAKilledInstallIsAFailure(t *testing.T) {
	s := newStub(t)
	s.write(t, "mode", "hang")
	r, _, ended := api(t)
	img := s.image(t, "slot-b.img.zst")

	if rsp := call(t, r, http.MethodPost, "/api/slot/install", `{"slot":"b","file":"`+img+`"}`); rsp.Code != 0 {
		t.Fatalf("install: %+v", rsp)
	}
	waitFor(t, func() bool { return progressOf(t, r).Phase == "write" })
	job, ok := readJob()
	if !ok {
		t.Fatal("no job file")
	}
	proc, err := os.FindProcess(job.Pid)
	if err != nil {
		t.Fatal(err)
	}
	_ = proc.Kill()
	if waitEnd(t, ended) {
		t.Fatal("a killed install was reported as done")
	}
	p := progressOf(t, r)
	if p.Running || p.Phase != "failed" || p.Message == "" {
		t.Fatalf("%+v", p)
	}
}

func TestTheInstallEndIsAudited(t *testing.T) {
	s := newStub(t)
	restore := audit.UseStore(audit.NewStore(t.TempDir()))
	t.Cleanup(restore)
	svc := NewService()
	img := s.image(t, "slot-b.img.zst")
	done := make(chan struct{})
	svc.install.onExit = func(job Job, ok bool, reason string) {
		recordResult(job, ok, reason)
		close(done)
	}
	if _, err := svc.install.start("b", img, false, audit.Identity{Actor: "admin", Via: "session", Role: "admin"}); err != nil {
		t.Fatal(err)
	}
	<-done
	audit.Default().Flush()
	page, err := audit.Default().Query(audit.Query{Category: audit.CategorySystem})
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range page.Events {
		if e.Action == "slot-install-result" && e.Actor == "admin" && e.Outcome == audit.OutcomeOK &&
			strings.Contains(e.Target, "slot b from ") {
			return
		}
	}
	t.Fatalf("no result line in %+v", page.Events)
}

func TestTryConfirmRevert(t *testing.T) {
	s := newStub(t)
	r, _, _ := api(t)

	if rsp := call(t, r, http.MethodPost, "/api/slot/try", `{"slot":"b"}`); rsp.Code != 0 {
		t.Fatalf("try: %+v", rsp)
	}
	if rsp := call(t, r, http.MethodPost, "/api/slot/try", `{"slot":"recovery"}`); rsp.Code == 0 {
		t.Fatal("a trial of recovery was accepted")
	}
	if rsp := call(t, r, http.MethodPost, "/api/slot/confirm", ``); rsp.Code != 0 {
		t.Fatalf("confirm: %+v", rsp)
	}
	s.write(t, "refuse", "no previous slot recorded")
	rsp := call(t, r, http.MethodPost, "/api/slot/revert", ``)
	if rsp.Code == 0 || rsp.Msg != "no previous slot recorded" {
		t.Fatalf("revert: %+v", rsp)
	}

	want := []string{"try b", "status --json", "confirm", "status --json", "revert"}
	if got := s.args(t); strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("the tool was run as %q", got)
	}
}

func TestStatusWithoutTheTool(t *testing.T) {
	newStub(t)
	Bin = filepath.Join(t.TempDir(), "none")
	r, _, _ := api(t)
	if Available() {
		t.Fatal("a missing tool is available")
	}
	if rsp := call(t, r, http.MethodGet, "/api/slot", ""); rsp.Code != codeUnavailable {
		t.Fatalf("%+v", rsp)
	}
	if _, err := readStatus(context.Background()); err == nil {
		t.Fatal("no error")
	}
}

// What `slot status --json` prints for the board statusArmed describes, in
// the tool's own layout (generic/slot, cmd_status_json). The manifest repeats
// a key and carries a value with a quote, a backslash and a tab in it, which
// the column parser could never have read back.
const statusArmedJSON = `{
  "format": 1,
  "running": "a",
  "trusted": "a",
  "trial": "b",
  "prev": "none",
  "recovery": "no",
  "bootloader": "installed",
  "kernel": "slot",
  "slots": [
    {"name": "a", "state": "unknown", "version": "dist-0123456789ab"},
    {"name": "b", "state": "installed", "version": "2026.10.27", "sha256": "abc", "size": "4194304", "installed": "2026-10-27T10:00:00Z", "trial": "armed", "source": "ironkvm-2026.10.27.img.zst"},
    {"name": "recovery", "state": "installed"}
  ],
  "manifest": [
    {"key": "device", "value": "lichee-rv-nano"},
    {"key": "vermagic", "value": "5.10.4-tag- mod_unload riscv"},
    {"key": "slotkernel", "value": "6.12.0"},
    {"key": "flavour", "value": "say \"hi\" \\ to\tall"},
    {"key": "flavour", "value": "second"},
    {"key": "libkvm", "value": "libkvm-vendor.so 0123"},
    {"key": "version", "value": "dist-0123456789ab"}
  ]
}
`

func TestParseStatusJSON(t *testing.T) {
	st, err := parseStatusJSON(statusArmedJSON)
	if err != nil {
		t.Fatal(err)
	}
	if st.Running != "a" || st.Trusted != "a" || st.Trial != "b" || st.Prev != "none" ||
		st.Recovery != "no" || st.Bootloader != "installed" || st.Kernel != "slot" {
		t.Fatalf("status fields: %+v", st)
	}
	if len(st.Slots) != 3 {
		t.Fatalf("slots: %+v", st.Slots)
	}
	b, _ := st.Slot("b")
	want := Info{Name: "b", State: "installed", Version: "2026.10.27", Sha256: "abc",
		Installed: "2026-10-27T10:00:00Z", Trial: "armed", Source: "ironkvm-2026.10.27.img.zst"}
	if b != want {
		t.Fatalf("slot b: %+v", b)
	}
	// The running slot's version, which the tool takes from its manifest when
	// no install recorded one.
	if a, _ := st.Slot("a"); a.State != "unknown" || a.Version != "dist-0123456789ab" {
		t.Fatalf("slot a: %+v", a)
	}
	if r, _ := st.Slot("recovery"); r.State != "installed" {
		t.Fatalf("slot recovery: %+v", r)
	}
	if st.Image["device"] != "lichee-rv-nano" || st.Image["slotkernel"] != "6.12.0" ||
		st.Image["version"] != "dist-0123456789ab" {
		t.Fatalf("image: %v", st.Image)
	}
	// The first of a repeated key, as the column parser keeps it, and the
	// escapes read back.
	if st.Image["flavour"] != "say \"hi\" \\ to\tall" {
		t.Fatalf("flavour: %q", st.Image["flavour"])
	}
	if _, ok := st.Image["vermagic"]; ok {
		t.Fatal("the module line is not shown")
	}
}

// Both forms describe the same board the same way, apart from what only the
// JSON carries.
func TestParseStatusJSONAgreesWithTheText(t *testing.T) {
	fromJSON, err := parseStatusJSON(statusArmedJSON)
	if err != nil {
		t.Fatal(err)
	}
	fromText := parseStatus(statusArmed)
	for i := range fromJSON.Slots {
		fromJSON.Slots[i].Source = ""
		if fromJSON.Slots[i].Name == "a" {
			fromJSON.Slots[i].Version = ""
		}
	}
	a, _ := json.Marshal(fromJSON.Slots)
	b, _ := json.Marshal(fromText.Slots)
	if string(a) != string(b) {
		t.Fatalf("slots differ:\n json %s\n text %s", a, b)
	}
	fromJSON.Slots, fromText.Slots, fromJSON.Image, fromText.Image = nil, nil, nil, nil
	a, _ = json.Marshal(fromJSON)
	b, _ = json.Marshal(fromText)
	if string(a) != string(b) {
		t.Fatalf("fields differ:\n json %s\n text %s", a, b)
	}
}

func TestParseStatusJSONRefusesWhatItCannotRead(t *testing.T) {
	if _, err := parseStatusJSON(strings.Replace(statusArmedJSON, `"format": 1`, `"format": 2`, 1)); err == nil ||
		!strings.Contains(err.Error(), "format 2") {
		t.Fatalf("a later format: %v", err)
	}
	if _, err := parseStatusJSON(`{"running": "a"}`); err == nil {
		t.Fatal("an object with no format was read")
	}
	if _, err := parseStatusJSON(`{"format": 1, "running": `); err == nil {
		t.Fatal("a cut-off object was read")
	}
	st, err := parseStatusJSON(`{"format": 1, "running": "a", "trusted": "a", "slots": []}`)
	if err != nil || st.Slots == nil || len(st.Slots) != 0 || st.Image != nil {
		t.Fatalf("a board with no records and no manifest: %+v, %v", st, err)
	}
}

func TestReadStatusAsksForJSON(t *testing.T) {
	s := newStub(t)
	s.write(t, "status.json", statusArmedJSON)
	// A warning on stderr is not part of the answer.
	s.write(t, "status.warn", "slot: cannot keep a record in /data/ironkvm/slots\n")
	st, err := readStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := st.Slot("b"); b.Source != "ironkvm-2026.10.27.img.zst" {
		t.Fatalf("the JSON was not read: %+v", st)
	}
	if got := s.args(t); strings.Join(got, "|") != "status --json" {
		t.Fatalf("the tool was run as %q", got)
	}
}

// An image from before --json has a tool that ignores the option and prints
// the text form. That is read with the column parser.
func TestReadStatusFromAnOlderTool(t *testing.T) {
	s := newStub(t)
	st, err := readStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := st.Slot("b"); b.Version != "2026.10.27" || b.Trial != "armed" || b.Source != "" {
		t.Fatalf("the text was not read: %+v", st)
	}
	if st.Image["device"] != "lichee-rv-nano" {
		t.Fatalf("image: %v", st.Image)
	}
	if got := s.args(t); strings.Join(got, "|") != "status --json" {
		t.Fatalf("the tool was run as %q", got)
	}
}

func TestReadStatusOfALaterFormat(t *testing.T) {
	s := newStub(t)
	s.write(t, "status.json", `{"format": 2, "board": {}}`)
	if _, err := readStatus(context.Background()); err == nil {
		t.Fatal("a format this server does not know was read")
	}
}
