package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/apikey"
	"NanoKVM-Server/service/roommic"

	"github.com/gin-gonic/gin"
)

// roomMicAccounts gives the test an administrator and an ordinary operator,
// each with an api key, and points the settings file somewhere temporary.
func roomMicAccounts(t *testing.T) (admin string, operator string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, ok, err := store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err := store.Create("alice", "valid-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}

	conf := config.GetInstance()
	originalStore := authn.DefaultStore
	originalAuthentication := conf.Authentication
	originalSecret := conf.JWT.SecretKey
	originalKeys := apikey.File
	originalSettings := roommic.SettingsFile

	authn.DefaultStore = store
	conf.Authentication = "enable"
	conf.JWT.SecretKey = "test-secret"
	apikey.File = filepath.Join(t.TempDir(), "api_keys.json")
	roommic.SettingsFile = filepath.Join(t.TempDir(), "room-mic")

	t.Cleanup(func() {
		authn.DefaultStore = originalStore
		conf.Authentication = originalAuthentication
		conf.JWT.SecretKey = originalSecret
		apikey.File = originalKeys
		roommic.SettingsFile = originalSettings
	})

	adminKey, _, err := apikey.Create("test", "admin")
	if err != nil {
		t.Fatal(err)
	}
	operatorKey, _, err := apikey.Create("test", "alice")
	if err != nil {
		t.Fatal(err)
	}

	return adminKey, operatorKey
}

func roomMicRequest(r *gin.Engine, method string, key string, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "/api/room-mic", strings.NewReader(body))
	req.Host = "nanokvm.local"
	req.Header.Set("Content-Type", "application/json")
	if key != "" {
		req.Header.Set("X-API-Key", key)
	}

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// Only an administrator may allow the microphone or change its gain. Every
// viewer may read whether it is live, because the indicator shows it to all.
func TestOnlyAnAdministratorChangesTheRoomMicrophone(t *testing.T) {
	adminKey, operatorKey := roomMicAccounts(t)

	r := gin.New()
	roomMicRouter(r)

	if w := roomMicRequest(r, http.MethodPost, "", `{"gain":10}`); w.Code != http.StatusUnauthorized {
		t.Fatalf("without a login: %d %s", w.Code, w.Body.String())
	}

	if w := roomMicRequest(r, http.MethodPost, operatorKey, `{"allowed":true,"gain":10}`); w.Code != http.StatusForbidden {
		t.Fatalf("an operator changed the setting: %d %s", w.Code, w.Body.String())
	}
	if _, err := os.Stat(roommic.SettingsFile); !os.IsNotExist(err) {
		t.Fatal("the operator's request wrote the settings file")
	}

	w := roomMicRequest(r, http.MethodPost, adminKey, `{"allowed":false,"gain":10}`)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("the administrator's change failed: %d %s", w.Code, w.Body.String())
	}
	if got := roommic.Shared.Settings().Gain; got != 10 {
		t.Fatalf("gain is %d after the administrator set 10", got)
	}

	w = roomMicRequest(r, http.MethodGet, operatorKey, "")
	if w.Code != http.StatusOK {
		t.Fatalf("an operator could not read the status: %d", w.Code)
	}
	var rsp struct {
		Code int            `json:"code"`
		Data roommic.Status `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatal(err)
	}
	if rsp.Code != 0 || rsp.Data.Gain != 10 || rsp.Data.Live {
		t.Fatalf("status %+v", rsp)
	}
}

// Without the onboard card the setting cannot be turned on: that is slot A,
// where recording from the codec hangs the board.
func TestTheRoomMicrophoneCannotBeAllowedWithoutTheCard(t *testing.T) {
	adminKey, _ := roomMicAccounts(t)
	if roommic.Available() {
		t.Skip("this machine has the onboard card")
	}

	r := gin.New()
	roomMicRouter(r)

	w := roomMicRequest(r, http.MethodPost, adminKey, `{"allowed":true}`)
	if w.Code != http.StatusOK || strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("allowing without the card was accepted: %d %s", w.Code, w.Body.String())
	}
	if roommic.Shared.Settings().Allowed {
		t.Fatal("the setting was saved as allowed")
	}
}
