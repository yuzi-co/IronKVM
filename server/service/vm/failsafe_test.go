package vm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/failsafe"

	"github.com/gin-gonic/gin"
)

// failsafeScratch points the failsafe package at a temporary directory with
// the given number of video crashes in it, and runs its Init.
func failsafeScratch(t *testing.T, crashes int) {
	t.Helper()
	base := t.TempDir()
	savedDir, savedUp, savedProc, savedRestart := failsafe.Dir, failsafe.UptimeFile, failsafe.ProcDir, failsafe.Restart
	savedDelay := restartDelay
	t.Cleanup(func() {
		failsafe.Dir, failsafe.UptimeFile, failsafe.ProcDir, failsafe.Restart = savedDir, savedUp, savedProc, savedRestart
		restartDelay = savedDelay
	})
	failsafe.Dir = filepath.Join(base, "failsafe")
	failsafe.UptimeFile = filepath.Join(base, "uptime")
	failsafe.ProcDir = filepath.Join(base, "proc")
	restartDelay = 0

	if err := os.MkdirAll(filepath.Join(failsafe.Dir, "exits"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(failsafe.UptimeFile, []byte("1000.00 0.00\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < crashes; i++ {
		body := fmt.Sprintf("uptime %d\ntime 0\nstatus 139\npid %d\n--- crash\n--- log\n", 990-i, i)
		name := filepath.Join(failsafe.Dir, "exits", fmt.Sprintf("%d.%d", 990-i, i))
		if err := os.WriteFile(name, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	failsafe.Init()
}

func callFailsafe(t *testing.T, h gin.HandlerFunc, method string) proto.Response {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", nil)
	h(c)
	var rsp proto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %v", w.Body.String(), err)
	}
	return rsp
}

func TestGetFailsafeReportsTheState(t *testing.T) {
	failsafeScratch(t, 3)

	rsp := callFailsafe(t, NewService().GetFailsafe, http.MethodGet)
	data, _ := json.Marshal(rsp.Data)
	var st failsafe.State
	if err := json.Unmarshal(data, &st); err != nil {
		t.Fatal(err)
	}
	if rsp.Code != 0 || !st.Active || len(st.Exits) != 3 || st.Threshold != failsafe.Threshold {
		t.Errorf("code %d, state %+v", rsp.Code, st)
	}
}

func TestRetryFailsafeRefusesWithVideoOn(t *testing.T) {
	failsafeScratch(t, 1)
	failsafe.Restart = func() { t.Error("restarted a server that has video") }

	if rsp := callFailsafe(t, NewService().RetryFailsafe, http.MethodPost); rsp.Code == 0 {
		t.Error("a retry with video on succeeded")
	}
}

func TestRetryFailsafeClearsAndRestarts(t *testing.T) {
	failsafeScratch(t, 3)
	restarted := make(chan struct{}, 1)
	failsafe.Restart = func() { restarted <- struct{}{} }

	if rsp := callFailsafe(t, NewService().RetryFailsafe, http.MethodPost); rsp.Code != 0 {
		t.Fatalf("retry failed: %s", rsp.Msg)
	}
	select {
	case <-restarted:
	case <-time.After(2 * time.Second):
		t.Fatal("the server did not restart")
	}

	entries, _ := os.ReadDir(filepath.Join(failsafe.Dir, "exits"))
	if len(entries) != 0 {
		t.Errorf("%d records left after the retry", len(entries))
	}
	if failsafe.Init() {
		t.Error("the next start is still in failsafe mode")
	}
}
