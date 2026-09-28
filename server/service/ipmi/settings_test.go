package ipmi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

// apiResponse is the web UI's response envelope.
type apiResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

// api calls one of the page's handlers, without the token check the server
// puts in front of them.
func (env *testEnv) api(t *testing.T, method, path, body string) apiResponse {
	t.Helper()
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/api/ipmi/settings", env.service.GetSettings)
	engine.POST("/api/ipmi/settings", env.service.SetSettings)
	engine.POST("/api/ipmi/users/:username/password", env.service.SetPassword)
	engine.DELETE("/api/ipmi/users/:username/password", env.service.ClearPassword)

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	engine.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s: status %d %s", method, path, w.Code, w.Body.String())
	}
	var rsp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("%s %s: %s\n%s", method, path, err, w.Body.String())
	}
	return rsp
}

func TestTheSettingsListTheAccounts(t *testing.T) {
	env := newTestEnv(t)
	if err := env.accounts.Create("a-name-longer-than-16", "web-password-5", "user"); err != nil {
		t.Fatal(err)
	}

	rsp := env.api(t, http.MethodGet, "/api/ipmi/settings", "")
	var got Settings
	if err := json.Unmarshal(rsp.Data, &got); rsp.Code != 0 || err != nil {
		t.Fatalf("code %d, %v", rsp.Code, err)
	}
	if !got.Enabled || got.Port != Port || !got.PowerLED {
		t.Fatalf("settings %+v", got)
	}
	rows := map[string]UserRow{}
	for _, row := range got.Users {
		rows[row.Username] = row
	}
	if !rows[adminName].HasPassword || rows[noIPMIName].HasPassword || !rows[adminName].NameFits {
		t.Fatalf("rows %+v", rows)
	}
	if rows["a-name-longer-than-16"].NameFits {
		t.Fatal("a 21 character name fits")
	}
	if strings.Contains(string(rsp.Data), "v1:") {
		t.Fatal("the settings carry a sealed password")
	}
}

func TestSettingAnIPMIPassword(t *testing.T) {
	env := newTestEnv(t)

	for _, bad := range []struct{ user, password, want string }{
		{noIPMIName, "short", errPasswordLength.Error()},
		{noIPMIName, "twenty-one-characters", errPasswordLength.Error()},
		{noIPMIName, "tab\tin-the-password", errPasswordChars.Error()},
		{noIPMIName, "web-password-3", errSameAsWeb.Error()},
		{"nobody", "a-good-ipmi-pass", errNoSuchAccount.Error()},
	} {
		body, _ := json.Marshal(map[string]string{"password": bad.password})
		rsp := env.api(t, http.MethodPost, "/api/ipmi/users/"+bad.user+"/password", string(body))
		if rsp.Code == 0 || rsp.Msg != bad.want {
			t.Errorf("%s %q: code %d %q, want %q", bad.user, bad.password, rsp.Code, rsp.Msg, bad.want)
		}
	}

	rsp := env.api(t, http.MethodPost, "/api/ipmi/users/"+noIPMIName+"/password", `{"password":"a-good-ipmi-pass"}`)
	if rsp.Code != 0 {
		t.Fatalf("set: %d %s", rsp.Code, rsp.Msg)
	}
	c := dial(t, env.service)
	c.login(3, noIPMIName, "a-good-ipmi-pass", privAdmin|lookupNameOnly)

	// Removing it ends the session and the access.
	rsp = env.api(t, http.MethodDelete, "/api/ipmi/users/"+noIPMIName+"/password", "")
	if rsp.Code != 0 {
		t.Fatalf("clear: %d %s", rsp.Code, rsp.Msg)
	}
	if _, _, err := c.command(netFnApp, cmdGetDeviceID, nil); !errors.Is(err, errNoAnswer) {
		t.Fatalf("the session outlived the password: %v", err)
	}
	user, _ := env.accounts.Get(noIPMIName)
	if user.IPMIPassword != "" {
		t.Fatal("the password is still stored")
	}
}

func TestTheSwitchOpensAndClosesTheSocket(t *testing.T) {
	env := newTestEnv(t)
	env.service.Stop()
	env.enabled = false

	if rsp := env.api(t, http.MethodPost, "/api/ipmi/settings", `{"enabled":true}`); rsp.Code != 0 {
		t.Fatalf("on: %d %s", rsp.Code, rsp.Msg)
	}
	if !env.enabled || env.service.LocalAddr() == nil {
		t.Fatal("on did not open the socket")
	}
	c := dial(t, env.service)
	c.login(3, adminName, adminIPMI, privAdmin|lookupNameOnly)

	if rsp := env.api(t, http.MethodPost, "/api/ipmi/settings", `{"enabled":false}`); rsp.Code != 0 {
		t.Fatalf("off: %d %s", rsp.Code, rsp.Msg)
	}
	if env.enabled || env.service.LocalAddr() != nil {
		t.Fatal("off did not close the socket")
	}
}

func TestSealedPasswordsOpenOnlyForTheirAccount(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ipmi.key")
	k := NewKeyring(path)

	sealed, err := k.Seal("alice", "a-good-ipmi-pass")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(sealed, "a-good-ipmi-pass") {
		t.Fatal("the password is in the clear")
	}
	if got, err := k.Open("alice", sealed); err != nil || string(got) != "a-good-ipmi-pass" {
		t.Fatalf("open: %q, %v", got, err)
	}
	if _, err := k.Open("bob", sealed); err == nil {
		t.Fatal("alice's password opened for bob")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("key file mode %v", info.Mode().Perm())
	}

	// A new keyring reads the same key.
	if got, err := NewKeyring(path).Open("alice", sealed); err != nil || string(got) != "a-good-ipmi-pass" {
		t.Fatalf("reopen: %q, %v", got, err)
	}
	// Without the key file nothing opens, and nothing makes a new key.
	if _, err := NewKeyring(filepath.Join(t.TempDir(), "none")).Open("alice", sealed); !errors.Is(err, errNoKey) {
		t.Fatalf("open without a key: %v", err)
	}
}
