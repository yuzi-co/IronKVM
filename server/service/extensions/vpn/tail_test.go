//go:build linux

package vpn

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTailKeepsTheLastNonBlankLines(t *testing.T) {
	got := Tail([]byte("one\n\n two \nthree\r\nfour\n"), 2)
	if got != "three\nfour" {
		t.Fatalf("got %q", got)
	}
}

func TestTailCapsItsSize(t *testing.T) {
	got := Tail([]byte(strings.Repeat("x", 5000)), TailLines)
	if len(got) != maxTailBytes {
		t.Fatalf("got %d bytes, want %d", len(got), maxTailBytes)
	}
}

func TestRunCarriesTheTail(t *testing.T) {
	_, err := Run(exec.Command("sh", "-c", "echo one; echo two; echo why >&2; exit 3"))
	var ce *CmdError
	if !errors.As(err, &ce) {
		t.Fatalf("want a *CmdError, got %v", err)
	}
	if ce.Tail != "one\ntwo\nwhy" {
		t.Fatalf("tail is %q", ce.Tail)
	}
	if got := Message("start failed", err); got != "start failed:\none\ntwo\nwhy" {
		t.Fatalf("message is %q", got)
	}
}

func TestMessageWithoutOutputUsesTheError(t *testing.T) {
	if got := Message("stop failed", errors.New("boom")); got != "stop failed: boom" {
		t.Fatalf("message is %q", got)
	}
}

func script(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "S98test")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

// The boot scripts exit 0 when the daemon did not come up and print FAIL.
func TestScriptTakesAFailLineAsAFailedStart(t *testing.T) {
	s := script(t, `echo "GOMEMLIMIT set to 56MiB"; echo "Starting netbird: FAIL"`)
	err := Script(s, "start", "")
	if err == nil {
		t.Fatal("a start that printed FAIL must fail")
	}
	if msg := Message("start failed", err); !strings.Contains(msg, "Starting netbird: FAIL") {
		t.Fatalf("message is %q", msg)
	}
}

func TestScriptAddsTheDaemonsLog(t *testing.T) {
	logFile := filepath.Join(t.TempDir(), "netbird.log")
	if err := os.WriteFile(logFile, []byte(
		"2026-09-27T10:00:00Z INFO starting\n"+
			"2026-09-27T10:00:01Z FATA failed to create interface wt0: operation not permitted\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	s := script(t, `echo "Starting netbird: FAIL"`)
	msg := Message("start failed", Script(s, "start", logFile))
	if !strings.Contains(msg, "failed to create interface wt0") {
		t.Fatalf("the daemon's own reason is missing: %q", msg)
	}
}

func TestScriptStartThatSucceeds(t *testing.T) {
	if err := Script(script(t, `echo "Starting netbird: OK"`), "start", ""); err != nil {
		t.Fatal(err)
	}
}

// Stop with nothing running prints FAIL and is still a successful stop.
func TestScriptStopWithNothingRunningIsNotAnError(t *testing.T) {
	if err := Script(script(t, `echo "Stopping netbird: FAIL"`), "stop", ""); err != nil {
		t.Fatal(err)
	}
}

func TestScriptExitStatusIsAFailure(t *testing.T) {
	err := Script(script(t, `echo "/usr/bin/netbird not found"; exit 1`), "start", "")
	if msg := Message("start failed", err); !strings.Contains(msg, "/usr/bin/netbird not found") {
		t.Fatalf("message is %q", msg)
	}
}
