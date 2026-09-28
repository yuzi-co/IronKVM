package vm

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
