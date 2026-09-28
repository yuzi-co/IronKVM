package screenshot

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	mcpservice "NanoKVM-Server/service/mcp"
)

type fakeSnapshotter struct {
	snap    mcpservice.Snapshot
	err     error
	request mcpservice.SnapshotRequest
}

func (f *fakeSnapshotter) Capture(_ context.Context, req mcpservice.SnapshotRequest) (mcpservice.Snapshot, error) {
	f.request = req
	return f.snap, f.err
}

func serve(t *testing.T, snapshotter mcpservice.Snapshotter) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/screenshot", Handler(snapshotter))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/screenshot", nil))
	return w
}

func envelope(t *testing.T, w *httptest.ResponseRecorder) (int, string) {
	t.Helper()

	var body struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("the answer is not an envelope: %q", w.Body.String())
	}
	return body.Code, body.Msg
}

func TestAFrameIsServedAsAJPEG(t *testing.T) {
	jpeg := []byte{0xff, 0xd8, 0xff, 0xd9}
	snapshotter := &fakeSnapshotter{snap: mcpservice.Snapshot{OK: true, Width: 1920, Height: 1080, JPEG: jpeg}}

	w := serve(t, snapshotter)

	if w.Code != http.StatusOK {
		t.Fatalf("status %d", w.Code)
	}
	if got := w.Header().Get("Content-Type"); got != "image/jpeg" {
		t.Errorf("Content-Type %q", got)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control %q", got)
	}
	if w.Body.String() != string(jpeg) {
		t.Errorf("body %v", w.Body.Bytes())
	}
	if snapshotter.request.Quality != quality {
		t.Errorf("asked for quality %d", snapshotter.request.Quality)
	}
	if snapshotter.request.X != 0 || snapshotter.request.Y != 0 || snapshotter.request.W != 0 || snapshotter.request.H != 0 {
		t.Errorf("asked for a crop: %+v", snapshotter.request)
	}
}

func TestNoFrameAnswersWithTheSnapshotterMessage(t *testing.T) {
	w := serve(t, &fakeSnapshotter{snap: mcpservice.Snapshot{Message: "no HDMI signal or frame unavailable"}})

	code, msg := envelope(t, w)
	if code == 0 || msg != "no HDMI signal or frame unavailable" {
		t.Errorf("code %d, msg %q", code, msg)
	}
}

func TestNoFrameWithoutAMessageStillSaysWhy(t *testing.T) {
	w := serve(t, &fakeSnapshotter{})

	code, msg := envelope(t, w)
	if code == 0 || msg == "" {
		t.Errorf("code %d, msg %q", code, msg)
	}
}

func TestACaptureErrorIsReported(t *testing.T) {
	w := serve(t, &fakeSnapshotter{err: errors.New("screenshot capture is unavailable")})

	code, msg := envelope(t, w)
	if code == 0 || msg != "screenshot capture is unavailable" {
		t.Errorf("code %d, msg %q", code, msg)
	}
}
