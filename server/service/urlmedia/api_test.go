package urlmedia

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/service/audit"
)

func postJSON(r *gin.Engine, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// The audit target is the drive and the host; neither the answer nor the
// target carries the URL's credentials or query.
func TestMountURLAuditsTheHostOnly(t *testing.T) {
	newFakeBoard(t)
	gin.SetMode(gin.TestMode)
	s := New()

	store := audit.NewStore(filepath.Join(t.TempDir(), "audit"))
	t.Cleanup(audit.UseStore(store))
	audit.SetIdentity(func(*gin.Context) (audit.Identity, bool) {
		return audit.Identity{Actor: "admin", Via: "session"}, true
	})
	t.Cleanup(func() { audit.SetIdentity(nil) })

	r := gin.New()
	r.Use(audit.Middleware())
	r.POST("/api/storage/url/mount", s.MountURL)
	r.POST("/api/storage/url/eject", s.EjectURL)
	r.GET("/api/storage/url", s.GetStatus)

	w := postJSON(r, "/api/storage/url/mount", `{"drive":"cdrom","url":"https://bob:hunter2@isos.lan:8443/x.iso?token=abc"}`)
	var rsp struct {
		Code int         `json:"code"`
		Data MountStatus `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil || rsp.Code != 0 {
		t.Fatalf("mount answered %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), "hunter2") || strings.Contains(w.Body.String(), "token=abc") {
		t.Fatalf("the answer leaks the URL's secrets: %s", w.Body.String())
	}
	page, err := store.Query(audit.Query{})
	if err != nil || len(page.Events) != 1 {
		t.Fatalf("audit: %v, %+v", err, page)
	}
	if e := page.Events[0]; e.Action != "mount-url" || e.Target != "cdrom: isos.lan:8443" {
		t.Fatalf("audit event %+v, want mount-url of the drive and host", e)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/storage/url", nil)
	sw := httptest.NewRecorder()
	r.ServeHTTP(sw, req)
	if strings.Contains(sw.Body.String(), "hunter2") || !strings.Contains(sw.Body.String(), `"available":true`) {
		t.Fatalf("status: %s", sw.Body.String())
	}

	if w := postJSON(r, "/api/storage/url/eject", `{"drive":"cdrom"}`); !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("eject answered %s", w.Body.String())
	}
	if w := postJSON(r, "/api/storage/url/eject", `{"drive":"cdrom"}`); strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("a second eject succeeded: %s", w.Body.String())
	}
}

func TestMountURLSaysWhyItIsUnavailable(t *testing.T) {
	b := newFakeBoard(t)
	b.available = errNoNBD
	gin.SetMode(gin.TestMode)
	s := New()
	r := gin.New()
	r.POST("/mount", s.MountURL)

	w := postJSON(r, "/mount", `{"drive":"cdrom","url":"http://example.com/x.iso"}`)
	if !strings.Contains(w.Body.String(), `"code":-3`) || !strings.Contains(w.Body.String(), "slot B") {
		t.Fatalf("mount on a kernel without NBD answered %s", w.Body.String())
	}
}
