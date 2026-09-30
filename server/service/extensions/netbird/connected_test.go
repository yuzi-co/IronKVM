//go:build linux

package netbird

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// fakeCli points NetbirdPath at a script that prints out.
func fakeCli(t *testing.T, out string) {
	t.Helper()
	dir := t.TempDir()
	body := filepath.Join(dir, "out")
	if err := os.WriteFile(body, []byte(out), 0o644); err != nil {
		t.Fatal(err)
	}
	saved := NetbirdPath
	t.Cleanup(func() { NetbirdPath = saved })
	NetbirdPath = filepath.Join(dir, "netbird")
	stub(t, NetbirdPath, "cat "+body)
}

func TestConnectedWhenTheDaemonIsConnected(t *testing.T) {
	connected, err := os.ReadFile("testdata/status-connected.json")
	if err != nil {
		t.Fatal(err)
	}
	fakeCli(t, string(connected))
	if !Connected(context.Background()) {
		t.Fatal("a connected daemon must read as connected")
	}
}

func TestNotConnectedWhileConnectingOrOtherwise(t *testing.T) {
	for _, out := range []string{`{"daemonStatus":"Connecting"}`, `{"daemonStatus":"NeedsLogin"}`, `{"daemonStatus":"Idle"}`, "daemon is not running"} {
		fakeCli(t, out)
		if Connected(context.Background()) {
			t.Fatalf("%q read as connected", out)
		}
	}
}
