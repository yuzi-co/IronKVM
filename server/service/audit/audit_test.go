package audit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
)

// testEngine builds an engine with the audit middleware and a store of its
// own. who is the identity every request has, if any.
func testEngine(t *testing.T, who *Identity) (*gin.Engine, *Store) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	oldStd, oldIdentify := std, identify
	std = newTestStore(t)
	identify = func(*gin.Context) (Identity, bool) {
		if who == nil {
			return Identity{}, false
		}
		return *who, true
	}
	t.Cleanup(func() { std, identify = oldStd, oldIdentify })

	r := gin.New()
	r.Use(Middleware())
	return r, std
}

func call(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = "192.168.1.20:5555"
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func events(t *testing.T, s *Store) []Event {
	t.Helper()
	page, err := s.Query(Query{})
	if err != nil {
		t.Fatal(err)
	}
	return page.Events
}

func TestMiddlewareRecordsAChangeWithWhoAndWhere(t *testing.T) {
	r, s := testEngine(t, &Identity{Actor: "alice", Via: "apikey", Key: "ci"})
	r.POST("/api/storage/drives/:id/eject", func(c *gin.Context) {
		SetTarget(c, "cdrom")
		c.JSON(http.StatusOK, gin.H{"code": 0})
	})
	call(r, "POST", "/api/storage/drives/cdrom/eject")

	got := events(t, s)
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
	e := got[0]
	if e.Actor != "alice" || e.Via != "apikey" || e.Key != "ci" || e.IP != "192.168.1.20" ||
		e.Category != CategoryMedia || e.Action != "eject" || e.Target != "cdrom" || e.Outcome != OutcomeOK {
		t.Fatalf("event wrong: %+v", e)
	}
}

func TestMiddlewareReadsFailureFromTheEnvelope(t *testing.T) {
	r, s := testEngine(t, &Identity{Actor: "alice", Via: "session"})
	r.POST("/api/vm/hostname", func(c *gin.Context) {
		var rsp proto.Response
		rsp.ErrRsp(c, -2, "invalid hostname")
	})
	call(r, "POST", "/api/vm/hostname")

	e := events(t, s)[0]
	if e.Outcome != OutcomeFail || e.Detail != "invalid hostname" || e.Category != CategorySettings || e.Action != "post /vm/hostname" {
		t.Fatalf("event wrong: %+v", e)
	}
}

func TestMiddlewareMarksForbiddenAsDenied(t *testing.T) {
	r, s := testEngine(t, &Identity{Actor: "oscar", Via: "session"})
	r.POST("/api/vm/hostname", func(c *gin.Context) { c.JSON(http.StatusForbidden, "forbidden") })
	call(r, "POST", "/api/vm/hostname")

	if e := events(t, s)[0]; e.Outcome != OutcomeDenied {
		t.Fatalf("outcome = %q, want denied", e.Outcome)
	}
}

func TestMiddlewareIgnoresReadsProbesAndSkippedRoutes(t *testing.T) {
	r, s := testEngine(t, nil)
	r.GET("/api/vm/info", func(c *gin.Context) { c.JSON(200, "x") })
	r.POST("/api/vm/hostname", func(c *gin.Context) { c.JSON(401, "unauthorized") })
	call(r, "GET", "/api/vm/info")
	call(r, "POST", "/api/vm/hostname")
	if n := len(events(t, s)); n != 0 {
		t.Fatalf("anonymous probe and read left %d events", n)
	}

	r2, s2 := testEngine(t, &Identity{Actor: "alice", Via: "session"})
	r2.POST("/api/hid/paste/check", func(c *gin.Context) { c.JSON(200, "x") })
	r2.POST("/api/mcp", func(c *gin.Context) { c.JSON(200, "x") })
	call(r2, "POST", "/api/hid/paste/check")
	call(r2, "POST", "/api/mcp")
	if n := len(events(t, s2)); n != 0 {
		t.Fatalf("skipped routes left %d events", n)
	}
}

func TestLoginIsRecordedWithoutASessionAndNamesTheActor(t *testing.T) {
	r, s := testEngine(t, nil)
	r.POST("/api/auth/login", func(c *gin.Context) {
		SetActor(c, "root")
		var rsp proto.Response
		rsp.ErrRsp(c, -2, "invalid username or password")
	})
	call(r, "POST", "/api/auth/login")

	e := events(t, s)[0]
	if e.Actor != "root" || e.Action != "login" || e.Outcome != OutcomeFail || e.Category != CategoryAuth {
		t.Fatalf("event wrong: %+v", e)
	}
}

func TestForwardedForIsNotTrusted(t *testing.T) {
	r, s := testEngine(t, &Identity{Actor: "alice"})
	r.POST("/api/vm/hostname", func(c *gin.Context) { c.JSON(200, "x") })
	req := httptest.NewRequest("POST", "/api/vm/hostname", nil)
	req.RemoteAddr = "192.168.1.20:5555"
	req.Header.Set("X-Forwarded-For", "6.6.6.6")
	r.ServeHTTP(httptest.NewRecorder(), req)

	if ip := events(t, s)[0].IP; ip != "192.168.1.20" {
		t.Fatalf("ip = %q, want the peer address", ip)
	}
}

func TestClassify(t *testing.T) {
	cases := []struct {
		method, route, category, action string
		audited                         bool
	}{
		{"POST", "/api/vm/gpio", CategoryPower, "press", true},
		{"POST", "/api/storage/image/mount", CategoryMedia, "mount", true},
		{"POST", "/api/hid/macro/play", CategoryMacro, "play", true},
		{"POST", "/api/auth/users", CategoryAuth, "post /auth/users", true},
		{"DELETE", "/api/vm/ssh/keys", CategorySettings, "delete /vm/ssh/keys", true},
		{"POST", "/api/mcp/config", CategorySettings, "post /mcp/config", true},
		{"POST", "/api/mcp", "", "", false},
		{"POST", "/api/vm/screen", "", "", false},
		{"POST", "/redfish/v1/Systems/1/Actions", "", "", false},
	}
	for _, c := range cases {
		rule, ok := Classify(c.method, c.route)
		if ok != c.audited || rule.Category != c.category || rule.Action != c.action {
			t.Errorf("%s %s = %+v, %t; want %s %q audited=%t", c.method, c.route, rule, ok, c.category, c.action, c.audited)
		}
	}
}
