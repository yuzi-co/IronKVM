// Package vpn holds what the Tailscale and NetBird add-ons share: the tail of
// a failed command's output, which the page shows as the reason, the runner
// for their boot scripts, the login URL a CLI prints, the daemon's uptime and
// memory, the addons group's memory, the update check's cache, and the
// handlers for start at boot and for refusing while the other VPN runs.
package vpn

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// TailLines is how many lines of a failed command's output reach the page.
const TailLines = 8

// maxTailBytes caps one tail, so a CLI that prints a wall of text cannot fill
// the page.
const maxTailBytes = 2048

// logTailBytes is how much of a daemon's log is read for its last lines.
const logTailBytes = 16 * 1024

// Tail returns the last n non-blank lines of out, trimmed, one per line.
func Tail(out []byte, n int) string {
	var lines []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		if line := strings.TrimSpace(sc.Text()); line != "" {
			lines = append(lines, line)
		}
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	s := strings.Join(lines, "\n")
	if len(s) > maxTailBytes {
		s = s[len(s)-maxTailBytes:]
	}
	return s
}

// CmdError is a command that failed, with the last lines it printed.
type CmdError struct {
	Err  error
	Tail string
}

func (e *CmdError) Error() string {
	if e.Tail == "" {
		return e.Err.Error()
	}
	return e.Err.Error() + ": " + e.Tail
}

func (e *CmdError) Unwrap() error { return e.Err }

// Run runs cmd and returns its combined output. On failure the error is a
// *CmdError that carries the output's last lines.
func Run(cmd *exec.Cmd) ([]byte, error) {
	out, err := cmd.CombinedOutput()
	if err != nil {
		return out, &CmdError{Err: err, Tail: Tail(out, TailLines)}
	}
	return out, nil
}

// Message is what a failing handler returns: what failed, then why. The why
// is the command's last lines when there are any, and the error otherwise.
func Message(what string, err error) string {
	if err == nil {
		return what
	}
	var ce *CmdError
	if errors.As(err, &ce) && ce.Tail != "" {
		return what + ":\n" + ce.Tail
	}
	return what + ": " + err.Error()
}

// Script runs an add-on's boot script with one action.
//
// The scripts exit 0 even when the daemon did not come up; they print
// "Starting <daemon>: FAIL" instead. So a start or restart that prints such a
// line fails here too. logFile, when set, is the daemon's own log: a daemon
// that dies at once says why there and not in the script's output.
func Script(path, action, logFile string) error {
	out, err := Run(exec.Command("sh", path, action))
	if err == nil && (action == "start" || action == "restart") && startFailed(out) {
		err = &CmdError{
			Err:  fmt.Errorf("%s %s failed", filepath.Base(path), action),
			Tail: Tail(out, TailLines),
		}
	}
	if err != nil && logFile != "" {
		var ce *CmdError
		if errors.As(err, &ce) {
			if lines := Tail(readEnd(logFile, logTailBytes), TailLines); lines != "" {
				ce.Tail = strings.TrimSpace(ce.Tail + "\n" + lines)
			}
		}
	}
	return err
}

func startFailed(out []byte) bool {
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Starting ") && strings.HasSuffix(line, "FAIL") {
			return true
		}
	}
	return false
}

// readEnd returns up to n bytes from the end of the file, or nothing.
func readEnd(path string, n int64) []byte {
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()
	if fi, err := f.Stat(); err == nil && fi.Size() > n {
		if _, err := f.Seek(-n, io.SeekEnd); err != nil {
			return nil
		}
	}
	b, _ := io.ReadAll(io.LimitReader(f, n))
	return b
}
