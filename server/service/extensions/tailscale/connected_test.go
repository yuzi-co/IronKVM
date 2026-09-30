//go:build linux

package tailscale

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// fakeCli points TailscalePath at a script that prints out.
func fakeCli(t *testing.T, out string) {
	t.Helper()
	dir := t.TempDir()
	body := filepath.Join(dir, "out")
	if err := os.WriteFile(body, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	cli := filepath.Join(dir, "tailscale")
	if err := os.WriteFile(cli, []byte("#!/bin/sh\ncat "+body+"\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	saved := TailscalePath
	t.Cleanup(func() { TailscalePath = saved })
	TailscalePath = cli
}

func TestConnectedWhenTheBackendRuns(t *testing.T) {
	running, err := os.ReadFile("testdata/status-running.json")
	if err != nil {
		t.Fatal(err)
	}
	fakeCli(t, string(running))
	if !Connected(context.Background()) {
		t.Fatal("a running backend must read as connected")
	}
}

func TestNotConnectedOtherwise(t *testing.T) {
	for _, out := range []string{`{"BackendState":"Stopped"}`, `{"BackendState":"NeedsLogin"}`, "failed to connect"} {
		fakeCli(t, out)
		if Connected(context.Background()) {
			t.Fatalf("%q read as connected", out)
		}
	}

	TailscalePath = filepath.Join(t.TempDir(), "absent")
	if Connected(context.Background()) {
		t.Fatal("a missing CLI read as connected")
	}
}
