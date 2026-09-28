package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The screenshot shows the host screen, so it needs the same login as the
// streams beside it.
func TestTheScreenshotNeedsALogin(t *testing.T) {
	r := newTestEngine(t, t.TempDir())
	streamRouter(r)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/stream/screenshot", nil))
	if w.Code != http.StatusUnauthorized {
		t.Errorf("GET /api/stream/screenshot without a login answered %d %s", w.Code, w.Body.String())
	}
}
