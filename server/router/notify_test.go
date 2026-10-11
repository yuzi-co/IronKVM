package router

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
)

// The Notifications page's routes need a web UI login.
func TestTheNotifyRoutesNeedAWebLogin(t *testing.T) {
	original := notifyFile
	notifyFile = filepath.Join(t.TempDir(), "notify.json")
	t.Cleanup(func() { notifyFile = original })

	r := newTestEngine(t, t.TempDir())
	notifyRouter(r)

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/notify/settings"},
		{http.MethodPost, "/api/notify/settings"},
		{http.MethodGet, "/api/notify/state"},
		{http.MethodPost, "/api/notify/test"},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(route.method, route.path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a login answered %d %s", route.method, route.path, w.Code, w.Body.String())
		}
	}
}

// The Host log page's routes need a web UI login, the live tail included.
func TestTheHostlogRoutesNeedAWebLogin(t *testing.T) {
	original := hostlogDiskDir
	hostlogDiskDir = filepath.Join(t.TempDir(), "hostlog")
	t.Cleanup(func() { hostlogDiskDir = original })

	r := newTestEngine(t, t.TempDir())
	hostlogRouter(r)

	for _, route := range []struct{ method, path string }{
		{http.MethodGet, "/api/hostlog/settings"},
		{http.MethodPost, "/api/hostlog/settings"},
		{http.MethodGet, "/api/hostlog/state"},
		{http.MethodGet, "/api/hostlog/lines"},
		{http.MethodDelete, "/api/hostlog/lines"},
		{http.MethodGet, "/api/hostlog/download"},
		{http.MethodGet, "/api/hostlog/stream"},
	} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(route.method, route.path, nil))
		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a login answered %d %s", route.method, route.path, w.Code, w.Body.String())
		}
	}
}
