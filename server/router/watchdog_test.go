package router

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// The Watchdog page's routes need a web UI login.
func TestTheWatchdogRoutesNeedAWebLogin(t *testing.T) {
	original := watchdogLogDir
	watchdogLogDir = filepath.Join(t.TempDir(), "watchdog")
	t.Cleanup(func() { watchdogLogDir = original })

	r := newTestEngine(t, t.TempDir())
	watchdogRouter(r)

	for _, path := range []string{
		"/api/watchdog/settings",
		"/api/watchdog/state",
		"/api/watchdog/log",
		"/api/watchdog/log/1/screenshot",
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("GET %s without a login answered %d %s", path, w.Code, w.Body.String())
		}
	}
}
