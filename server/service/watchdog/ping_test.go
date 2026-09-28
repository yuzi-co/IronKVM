package watchdog

import (
	"context"
	"errors"
	"os"
	"testing"
)

// The loopback address always answers. The test needs a raw socket, which a
// container running as root has and an ordinary user does not.
func TestPingGetsAnAnswerFromLoopback(t *testing.T) {
	ok, err := ping(context.Background(), "127.0.0.1")
	if errors.Is(err, os.ErrPermission) {
		t.Skipf("no raw socket here: %s", err)
	}
	if err != nil || !ok {
		t.Fatalf("ping 127.0.0.1: %t, %v", ok, err)
	}
}

func TestPingRefusesAName(t *testing.T) {
	if ok, err := ping(context.Background(), "host.lan"); ok || err == nil {
		t.Fatalf("ping host.lan: %t, %v", ok, err)
	}
}
