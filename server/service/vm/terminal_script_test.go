package vm

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// nanokvmInitScript is the boot script that has to agree with this package.
const nanokvmInitScript = "../../../kvmapp/system/init.d/S95nanokvm"

func readNanokvmInitScript(t *testing.T) string {
	t.Helper()

	body, err := os.ReadFile(filepath.FromSlash(nanokvmInitScript))
	if err != nil {
		t.Skipf("cannot read %s: %s", nanokvmInitScript, err)
	}

	return string(body)
}

// The web terminal starts its shell in terminalWorkingDir, and pty.Start does
// the chdir itself. A directory that is not there is not a degraded terminal,
// it is no terminal: the chdir fails, pty.Start returns the error, and the
// handler closes the socket before the operator sees a prompt.
//
// That directory is /root, and this fork's own image does not have to have it.
// tools/abslots/manifest/root.manifest carries `remove /root`, because a
// running board accumulates about 132MB of operator state there and none of it
// belongs in an image. The stock Sipeed rootfs does have /root, which is why
// the upstream change that introduced this chdir could not have seen the
// problem: on a fork-built slot the same code has nowhere to go.
//
// So the script has to create the directory before it starts the server. This
// test is the thing that keeps the three files agreeing with each other: the
// manifest may keep removing /root, and the terminal keeps working, only
// because S95nanokvm makes it again on the way up.
func TestTheBootScriptCreatesTheTerminalWorkingDirectory(t *testing.T) {
	script := readNanokvmInitScript(t)

	made := regexp.MustCompile(`(?m)^\s*mkdir\s+(-\S+\s+)*` + regexp.QuoteMeta(terminalWorkingDir) + `\s*$`)
	if !made.MatchString(script) {
		t.Errorf("S95nanokvm never creates %s, so the web terminal cannot start there",
			terminalWorkingDir)
	}
}

// Creating the directory is worth nothing if it happens after the thing that
// needs it. The one use inside this script is the profile written into that
// directory, and this test requires the mkdir to come first. The server is
// started from start_services, further down the same `start)` arm, so putting
// the mkdir above the profile write puts it above the server as well.
func TestTheBootScriptCreatesThatDirectoryBeforeItIsUsed(t *testing.T) {
	script := readNanokvmInitScript(t)
	lines := strings.Split(script, "\n")

	made := regexp.MustCompile(`^\s*mkdir\s+(-\S+\s+)*` + regexp.QuoteMeta(terminalWorkingDir) + `\s*$`)

	makeAt := -1
	for i, line := range lines {
		if made.MatchString(line) {
			makeAt = i
			break
		}
	}
	if makeAt < 0 {
		t.Skip("the script does not create the directory at all, which the test above reports")
	}

	// Uses are counted whichever side of the mkdir they fall on, so that a use
	// that moved above it is reported as the ordering fault it is rather than
	// as an absence of uses.
	uses := 0
	for i, line := range lines {
		if !strings.Contains(line, terminalWorkingDir+"/") {
			continue
		}

		uses++
		if i <= makeAt {
			t.Errorf("line %d writes into %s before it is created: %q",
				i+1, terminalWorkingDir, strings.TrimSpace(line))
		}
	}

	if uses == 0 {
		t.Errorf("nothing in S95nanokvm uses %s, so this ordering test is measuring nothing",
			terminalWorkingDir)
	}
}
