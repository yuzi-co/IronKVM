package vm

import (
	"strings"
	"testing"
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
