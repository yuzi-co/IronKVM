//go:build linux

package vpn

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

const tailscaleLine = "To authenticate, visit:\n\n\thttps://login.tailscale.com/a/abcdef\n\n"

// NetBird's openURL with --no-browser: the URL, a space, and an empty code
// message when the code is already in the URL.
const netbirdLine = "Use this URL to log in:\n\nhttps://login.netbird.io/activate?user_code=ABCD-EFGH \n\n"

func waitForFile(t *testing.T, path string, within time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		if _, err := os.Stat(path); err == nil {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func waitForExit(t *testing.T, pid int, within time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(within)
	for time.Now().Before(deadline) {
		// A killed but unreaped child is still a zombie and signal 0 finds it.
		if err := syscall.Kill(pid, 0); errors.Is(err, syscall.ESRCH) {
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
	return false
}

func TestLoginReturnsTheURL(t *testing.T) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("printf %q >&2", tailscaleLine))
	url, err := LoginURL(cmd, false, 5*time.Second, 0)
	if err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if url != "https://login.tailscale.com/a/abcdef" {
		t.Fatalf("unexpected url %q", url)
	}
}

func TestLoginReadsNetBirdFromStdout(t *testing.T) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("printf %q", netbirdLine))
	url, err := LoginURL(cmd, true, 5*time.Second, 0)
	if err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if url != "https://login.netbird.io/activate?user_code=ABCD-EFGH" {
		t.Fatalf("unexpected url %q", url)
	}
}

func TestLoginLeavesTheCommandRunning(t *testing.T) {
	// The login keeps running until the user finishes in the browser. Closing
	// the pipe as soon as the URL is read hands the child a SIGPIPE on its
	// next line and kills the login it was told to complete.
	marker := filepath.Join(t.TempDir(), "finished")
	script := fmt.Sprintf("printf %q >&2; sleep 0.4; echo still-here >&2; touch %q", tailscaleLine, marker)
	if _, err := LoginURL(exec.Command("sh", "-c", script), false, 5*time.Second, 0); err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if !waitForFile(t, marker, 3*time.Second) {
		t.Fatal("expected the login command to run to completion")
	}
}

func TestLoginGivesUpWhenNoURLAppears(t *testing.T) {
	// Production runs the binary directly, so the test does too: killing a
	// shell wrapper would leave the real command holding the pipe.
	start := time.Now()
	if _, err := LoginURL(exec.Command("sleep", "30"), false, 300*time.Millisecond, 0); err == nil {
		t.Fatal("expected an error when no url is printed")
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("expected LoginURL to give up quickly, took %s", elapsed)
	}
}

func TestLoginReapsACommandThatGaveUp(t *testing.T) {
	cmd := exec.Command("sleep", "30")
	if _, err := LoginURL(cmd, false, 200*time.Millisecond, 0); err == nil {
		t.Fatal("expected an error when no url is printed")
	}
	if !waitForExit(t, cmd.Process.Pid, 5*time.Second) {
		t.Fatal("expected the abandoned command to be reaped")
	}
}

// netbird up has no timeout of its own, so a login nobody finishes is ended
// after its life.
func TestLoginEndsTheCommandAfterItsLife(t *testing.T) {
	cmd := exec.Command("sh", "-c", fmt.Sprintf("printf %q; exec sleep 30", netbirdLine))
	if _, err := LoginURL(cmd, true, 5*time.Second, 300*time.Millisecond); err != nil {
		t.Fatalf("expected the url to be read: %s", err)
	}
	if !waitForExit(t, cmd.Process.Pid, 5*time.Second) {
		t.Fatal("expected the login to be ended after its life")
	}
}

func TestLoginWithoutAURLCarriesTheOutput(t *testing.T) {
	cmd := exec.Command("sh", "-c", "echo 'daemon is not running' >&2")
	_, err := LoginURL(cmd, false, 5*time.Second, 0)
	if err == nil || !strings.Contains(Message("login failed", err), "daemon is not running") {
		t.Fatalf("the CLI's reason is missing: %v", err)
	}
}
