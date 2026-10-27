package slot

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/utils"
)

var (
	errBusy      = errors.New("an install is already running")
	errBadURL    = errors.New("an image URL must be an https:// address")
	errBadFile   = errors.New("an image file must be a .img or .img.zst file on /data")
	errBadSource = errors.New("give an image URL or an image file, not both and not neither")
)

// maxURLLength bounds the URL handed to the tool. A real release URL is a
// fraction of this.
const maxURLLength = 2048

// Job is the install that runs, or ran last. It is kept in JobFile beside the
// tool's progress file, so a restarted server knows which process to watch.
type Job struct {
	Slot string `json:"slot"`
	// Source is what to show for the image: the file's path, or the URL
	// without its credentials and query (see displaySource).
	Source  string    `json:"source"`
	Pid     int       `json:"pid"`
	Started time.Time `json:"started"`

	// Who asked, for the audit line that records the end. Not kept in
	// JobFile: a server restarted during the install records no end.
	actor, via, key, role string
}

// Progress is what GET /api/slot/install answers.
type Progress struct {
	// Running is true while the tool's process is alive.
	Running bool `json:"running"`
	// Phase is the tool's: verify, write, readback, done or failed. It is
	// empty when no install has run since the board booted.
	Phase   string    `json:"phase"`
	Done    int64     `json:"done"`
	Total   int64     `json:"total"`
	Message string    `json:"message,omitempty"`
	Slot    string    `json:"slot,omitempty"`
	Source  string    `json:"source,omitempty"`
	Started time.Time `json:"started,omitzero"`
}

// installer allows one install at a time. While this server started it, the
// process is watched through cmd; after a restart, through the pid in JobFile.
type installer struct {
	mu      sync.Mutex
	running bool
	// onExit is told how an install this server started ended, for the
	// audit log. Nil in most tests.
	onExit func(job Job, ok bool, reason string)
}

// parseProgress reads the one line the tool keeps current:
//
//	phase=<phase> done=<bytes> total=<bytes> [message=<text>]
//
// The message runs to the end of the line.
func parseProgress(line string) Progress {
	var p Progress
	line = strings.TrimRight(line, "\r\n")
	rest := line
	if i := strings.Index(line, " message="); i >= 0 {
		p.Message = line[i+len(" message="):]
		rest = line[:i]
	}
	for _, field := range strings.Fields(rest) {
		k, v, ok := strings.Cut(field, "=")
		if !ok {
			continue
		}
		switch k {
		case "phase":
			p.Phase = v
		case "done":
			p.Done, _ = strconv.ParseInt(v, 10, 64)
		case "total":
			p.Total, _ = strconv.ParseInt(v, 10, 64)
		}
	}
	return p
}

func finished(phase string) bool { return phase == "done" || phase == "failed" }

// validateURL accepts an https URL with a host and nothing a shell or a log
// line would trip on. The tool refuses anything but https too.
func validateURL(raw string) (string, error) {
	if raw == "" || len(raw) > maxURLLength {
		return "", errBadURL
	}
	for _, r := range raw {
		if r <= ' ' || r == 0x7f {
			return "", errBadURL
		}
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.Opaque != "" {
		return "", errBadURL
	}
	return raw, nil
}

// hasImageSuffix is the names the tool installs: a raw image or one
// compressed with zstd.
func hasImageSuffix(name string) bool {
	return strings.HasSuffix(name, ".img") || strings.HasSuffix(name, ".img.zst")
}

// resolveFile returns the image file a request names, with every link
// followed, or errBadFile when that is not a regular image file under
// ImageRoot. The name alone is not enough: a link in /data could point at a
// device or at another partition.
func resolveFile(path string) (string, error) {
	if !filepath.IsAbs(path) {
		return "", errBadFile
	}
	root, err := filepath.EvalSymlinks(ImageRoot)
	if err != nil {
		root = filepath.Clean(ImageRoot)
	}
	resolved, err := filepath.EvalSymlinks(filepath.Clean(path))
	if err != nil {
		return "", errBadFile
	}
	if !utils.IsPathInside(root, resolved) || !hasImageSuffix(resolved) {
		return "", errBadFile
	}
	st, err := os.Stat(resolved)
	if err != nil || !st.Mode().IsRegular() {
		return "", errBadFile
	}
	return resolved, nil
}

// displaySource is a URL as the page and the audit log show it: scheme, host
// and path. A signed download link carries its credential in the query, and a
// URL may carry a password, and neither belongs in a log.
func displaySource(source string) string {
	u, err := url.Parse(source)
	if err != nil || u.Scheme == "" {
		return source
	}
	return u.Scheme + "://" + u.Host + u.Path
}

// alive reports whether pid is a running slot install. The pid alone is not
// proof: after the install ends, another process can be given the number.
func alive(pid int) bool {
	if pid <= 0 || syscall.Kill(pid, 0) != nil {
		return false
	}
	cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		// No /proc to ask. kill answered, so take its word.
		return true
	}
	return bytes.Contains(cmdline, []byte("install")) &&
		bytes.Contains(cmdline, []byte(filepath.Base(Bin)))
}

func readJob() (Job, bool) {
	data, err := os.ReadFile(JobFile)
	if err != nil {
		return Job{}, false
	}
	var job Job
	if json.Unmarshal(data, &job) != nil {
		return Job{}, false
	}
	return job, true
}

func writeJob(job Job) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	tmp := JobFile + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, JobFile)
}

// lastLogLine is the tool's reason for an install that ended without saying
// so in the progress file, which happens only when it was killed.
func lastLogLine() string {
	data, err := os.ReadFile(LogFile)
	if err != nil {
		return ""
	}
	if reason := lastReason(string(data)); reason != "" {
		return reason
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	return strings.TrimSpace(lines[len(lines)-1])
}

// busy reports whether an install is running: one this server started, or
// one a restarted server finds in JobFile.
func (in *installer) busy() bool {
	if in.running {
		return true
	}
	job, ok := readJob()
	return ok && alive(job.Pid)
}

// progress is the state of the install that runs, or ran last.
func (in *installer) progress() Progress {
	in.mu.Lock()
	running := in.running
	in.mu.Unlock()

	job, haveJob := readJob()
	if !running && haveJob {
		running = alive(job.Pid)
	}

	var p Progress
	if data, err := os.ReadFile(ProgressFile); err == nil {
		p = parseProgress(string(data))
	}
	p.Running = running
	if haveJob {
		p.Slot, p.Source, p.Started = job.Slot, job.Source, job.Started
	}
	// The tool writes "failed" itself on every refusal. A process that is
	// gone and never wrote an end was killed, and the record of the slot
	// says "interrupted" by the next status.
	if haveJob && !running && !finished(p.Phase) {
		p.Phase = "failed"
		p.Message = lastLogLine()
		if p.Message == "" {
			p.Message = "the install stopped before it finished"
		}
	}
	return p
}

// start runs `slot install` for slot from source, in the background.
//
// The tool runs in a session of its own, with its output in LogFile and not
// in a pipe to this server. A server restarted during an install, by an
// application update or by a crash, would otherwise take the install with it:
// a signal to the server's process group reaches its children, and a write to
// a pipe nobody reads any more kills the writer. A slot write is never worth
// stopping halfway.
//
// --unsigned is never passed. The tool holds an install that no person at a
// shell asked for to the keyring, and an install from the API is one.
func (in *installer) start(slot, source string, force bool, who audit.Identity) (Job, error) {
	in.mu.Lock()
	defer in.mu.Unlock()
	if in.busy() {
		return Job{}, errBusy
	}

	_ = os.Remove(ProgressFile)
	logFile, err := os.OpenFile(LogFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o600)
	if err != nil {
		return Job{}, fmt.Errorf("cannot keep the install's log: %w", err)
	}

	args := []string{"install", "--progress", ProgressFile}
	if force {
		args = append(args, "--force")
	}
	args = append(args, "--", slot, source)
	cmd := exec.Command(Bin, args...)
	cmd.Stdin = nil
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return Job{}, err
	}
	// The child has its own copy of the descriptor.
	_ = logFile.Close()

	job := Job{
		Slot: slot, Source: displaySource(source), Pid: cmd.Process.Pid, Started: time.Now().UTC(),
		actor: who.Actor, via: who.Via, key: who.Key, role: who.Role,
	}
	if err := writeJob(job); err != nil {
		log.Warnf("slot install: cannot record the job in %s: %v", JobFile, err)
	}
	in.running = true
	log.Infof("slot install of %s from %s started, pid %d", slot, job.Source, job.Pid)

	go in.wait(cmd, job)
	return job, nil
}

// wait reaps the install and says how it ended.
func (in *installer) wait(cmd *exec.Cmd, job Job) {
	err := cmd.Wait()

	in.mu.Lock()
	in.running = false
	onExit := in.onExit
	in.mu.Unlock()

	p := in.progress()
	ok := err == nil && p.Phase == "done"
	reason := p.Message
	if !ok && reason == "" && err != nil {
		reason = err.Error()
	}
	if ok {
		log.Infof("slot install of %s from %s finished", job.Slot, job.Source)
	} else {
		log.Warnf("slot install of %s from %s failed: %s", job.Slot, job.Source, reason)
	}
	if onExit != nil {
		onExit(job, ok, reason)
	}
}
