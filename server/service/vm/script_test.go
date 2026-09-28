package vm

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestScriptCommandRunsShellScriptsThroughSh(t *testing.T) {
	// Upload sets the execute bit, but a .sh written as a plain list of
	// commands carries no shebang, and executing such a path directly fails
	// with ENOEXEC. The interpreter has to stay in the command.
	cmd := scriptCommand("backup.sh", "/etc/kvm/scripts/backup.sh")

	if len(cmd.Args) != 2 || cmd.Args[0] != "sh" || cmd.Args[1] != "/etc/kvm/scripts/backup.sh" {
		t.Fatalf("args are %q, want [sh /etc/kvm/scripts/backup.sh]", cmd.Args)
	}
}

func TestScriptCommandRunsPythonScriptsThroughPython(t *testing.T) {
	cmd := scriptCommand("report.PY", "/etc/kvm/scripts/report.PY")

	if len(cmd.Args) != 2 || cmd.Args[0] != "python" {
		t.Fatalf("args are %q, want python first", cmd.Args)
	}
}

func TestScriptCommandNeverPassesTheNameToAShell(t *testing.T) {
	// SecureJoin rejects this name before it reaches here. This asserts the
	// other half: even a name that got through stays one argument, so no part
	// of it is parsed as shell text.
	cmd := scriptCommand("a.sh; reboot", "/etc/kvm/scripts/a.sh; reboot")

	for _, arg := range cmd.Args {
		if strings.Contains(arg, "-c") {
			t.Fatalf("args are %q, want no shell -c", cmd.Args)
		}
	}
	if cmd.Args[len(cmd.Args)-1] != "/etc/kvm/scripts/a.sh; reboot" {
		t.Fatalf("the path must stay one argument, got %q", cmd.Args)
	}
}

func scriptRequest(t *testing.T, method, body string, handler gin.HandlerFunc) string {
	t.Helper()

	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, "/api/vm/script", strings.NewReader(body))
	if body != "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	handler(c)
	return recorder.Body.String()
}

// Upload creates the script directory, so a board that never had a script
// has none, and that is an empty list.
func TestNoScriptDirectoryIsAnEmptyList(t *testing.T) {
	original := ScriptDirectory
	t.Cleanup(func() { ScriptDirectory = original })
	ScriptDirectory = filepath.Join(t.TempDir(), "scripts")

	body := scriptRequest(t, http.MethodGet, "", (&Service{}).GetScripts)
	if !strings.Contains(body, `"code":0`) || !strings.Contains(body, `"files":[]`) {
		t.Fatalf("GetScripts answered %s", body)
	}
}

// A script that fails says why in its output, so the output comes back with
// the error.
func TestAFailedScriptReturnsItsOutput(t *testing.T) {
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("no sh")
	}
	original := ScriptDirectory
	t.Cleanup(func() { ScriptDirectory = original })
	ScriptDirectory = t.TempDir()

	script := "echo disk full\nexit 3\n"
	if err := os.WriteFile(filepath.Join(ScriptDirectory, "fail.sh"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	body := scriptRequest(t, http.MethodPost, `{"name":"fail.sh","type":"foreground"}`, (&Service{}).RunScript)
	if strings.Contains(body, `"code":0`) {
		t.Fatalf("a failing script reported success: %s", body)
	}
	if !strings.Contains(body, "disk full") {
		t.Fatalf("the output was lost: %s", body)
	}
}

// A foreground script that outlives the limit is killed, and so is what it
// started: the sleep below is a child of the shell, in the shell's group.
func TestRunForegroundKillsTheProcessGroupAtTheLimit(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "child.pid")

	cmd := exec.Command("sh", "-c", "echo started; sleep 30 & echo $! > "+pidFile+"; wait")

	start := time.Now()
	output, err := runForeground(cmd, 300*time.Millisecond)
	elapsed := time.Since(start)

	if !errors.Is(err, errScriptTimedOut) {
		t.Fatalf("err = %v, want errScriptTimedOut", err)
	}
	if elapsed > 10*time.Second {
		t.Fatalf("returned after %s, want soon after the limit", elapsed)
	}
	if !strings.Contains(string(output), "started") {
		t.Fatalf("output %q lost what the script printed before the kill", output)
	}

	raw, readErr := os.ReadFile(pidFile)
	if readErr != nil {
		t.Fatalf("read child pid: %v", readErr)
	}
	pid, convErr := strconv.Atoi(strings.TrimSpace(string(raw)))
	if convErr != nil {
		t.Fatalf("child pid %q: %v", raw, convErr)
	}

	// A killed child may linger as a zombie until something reaps it; that
	// counts as gone.
	deadline := time.Now().Add(5 * time.Second)
	for processRunning(pid) {
		if time.Now().After(deadline) {
			_ = syscall.Kill(pid, syscall.SIGKILL)
			t.Fatalf("child %d outlived the kill", pid)
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestRunForegroundReturnsOutputAndExitError(t *testing.T) {
	output, err := runForeground(exec.Command("sh", "-c", "echo out; echo err >&2; exit 3"), time.Minute)

	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) || exitErr.ExitCode() != 3 {
		t.Fatalf("err = %v, want exit status 3", err)
	}
	if !strings.Contains(string(output), "out") || !strings.Contains(string(output), "err") {
		t.Fatalf("output %q, want both streams", output)
	}
}

// processRunning reports whether pid names a process that is not a zombie.
func processRunning(pid int) bool {
	stat, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/stat")
	if err != nil {
		return false
	}
	// The state follows the command name, which is in parentheses.
	rest := string(stat)
	if i := strings.LastIndexByte(rest, ')'); i >= 0 {
		rest = rest[i+1:]
	}
	fields := strings.Fields(rest)
	return len(fields) > 0 && fields[0] != "Z"
}
