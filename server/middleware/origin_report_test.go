package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
)

type originReply struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Origin string `json:"origin"`
		Host   string `json:"host"`
	} `json:"data"`
}

func originCheckRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/api/origin-check", CheckToken(), OriginCheck)
	return router
}

func postOriginCheck(router *gin.Engine, host, origin string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(http.MethodPost, "/api/origin-check", nil)
	request.Host = host
	if origin != "" {
		request.Header.Set("Origin", origin)
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

// watchOriginLog swaps the clock and the logger for the test and returns the
// warnings seen so far, as "origin host" pairs, through a pointer.
func watchOriginLog(t *testing.T, now *time.Time) *[]string {
	t.Helper()
	var seen []string
	oldNow, oldWarn := originNow, originWarn
	originLogMu.Lock()
	oldLast := originLogLast
	originLogLast = time.Time{}
	originLogMu.Unlock()
	originNow = func() time.Time { return *now }
	originWarn = func(origin, host string) { seen = append(seen, origin+" "+host) }
	t.Cleanup(func() {
		originNow, originWarn = oldNow, oldWarn
		originLogMu.Lock()
		originLogLast = oldLast
		originLogMu.Unlock()
	})
	return &seen
}

func useAuthentication(t *testing.T, mode string) {
	t.Helper()
	conf := config.GetInstance()
	original := conf.Authentication
	conf.Authentication = mode
	t.Cleanup(func() { conf.Authentication = original })
}

func TestForbiddenOriginAnswersEnvelopeNamingBothHosts(t *testing.T) {
	useAuthentication(t, "enable")
	now := time.Unix(1000, 0)
	watchOriginLog(t, &now)

	recorder := postOriginCheck(originCheckRouter(), "ironkvm.lan", "https://kvm.lan.yuzi.co")
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403", recorder.Code)
	}
	var reply originReply
	if err := json.Unmarshal(recorder.Body.Bytes(), &reply); err != nil {
		t.Fatalf("body is not the envelope: %v: %s", err, recorder.Body.String())
	}
	if reply.Code != -1 {
		t.Fatalf("code = %d, want -1", reply.Code)
	}
	if reply.Data.Origin != "kvm.lan.yuzi.co" || reply.Data.Host != "ironkvm.lan" {
		t.Fatalf("data = %+v", reply.Data)
	}
	for _, name := range []string{"kvm.lan.yuzi.co", "ironkvm.lan"} {
		if !contains(reply.Msg, name) {
			t.Fatalf("msg %q does not name %s", reply.Msg, name)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

func TestForbiddenOriginLogsOncePerMinute(t *testing.T) {
	useAuthentication(t, "enable")
	now := time.Unix(1000, 0)
	seen := watchOriginLog(t, &now)
	router := originCheckRouter()

	postOriginCheck(router, "ironkvm.lan", "https://a.example")
	now = now.Add(30 * time.Second)
	postOriginCheck(router, "ironkvm.lan", "https://a.example")
	if len(*seen) != 1 {
		t.Fatalf("warnings inside the minute = %d, want 1", len(*seen))
	}
	if (*seen)[0] != "https://a.example ironkvm.lan" {
		t.Fatalf("warning = %q", (*seen)[0])
	}

	now = now.Add(31 * time.Second)
	postOriginCheck(router, "ironkvm.lan", "https://a.example")
	if len(*seen) != 2 {
		t.Fatalf("warnings after the minute = %d, want 2", len(*seen))
	}
}

func TestMatchingOriginLogsNothing(t *testing.T) {
	useAuthentication(t, "enable")
	now := time.Unix(1000, 0)
	seen := watchOriginLog(t, &now)

	recorder := postOriginCheck(originCheckRouter(), "ironkvm.lan", "https://ironkvm.lan")
	// No session: the auth path decides, and it is not a 403.
	if recorder.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", recorder.Code)
	}
	if len(*seen) != 0 {
		t.Fatalf("warnings = %v, want none", *seen)
	}
}

func TestOriginCheckOkWhenAuthenticationIsOff(t *testing.T) {
	useAuthentication(t, "disable")
	now := time.Unix(1000, 0)
	watchOriginLog(t, &now)

	recorder := postOriginCheck(originCheckRouter(), "ironkvm.lan", "https://ironkvm.lan")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var reply originReply
	if err := json.Unmarshal(recorder.Body.Bytes(), &reply); err != nil || reply.Code != 0 {
		t.Fatalf("reply = %+v err=%v", reply, err)
	}
}
