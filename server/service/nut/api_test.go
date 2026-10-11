package nut

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
)

type envelope struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func call(t *testing.T, r *gin.Engine, method, path, body string) (*httptest.ResponseRecorder, envelope) {
	t.Helper()

	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	var env envelope
	if strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
			t.Fatalf("%s %s: %s", method, path, rec.Body.String())
		}
	}
	return rec, env
}

// apiHarness is a monitor whose settings and password live in memory.
func apiHarness(t *testing.T) (*harness, *gin.Engine) {
	t.Helper()
	h := newHarness(t)
	h.settings = config.NUT{}
	h.password = ""
	h.mon.deps.SetSettings = func(s config.NUT) error { h.settings = s; return nil }
	h.mon.deps.SetPassword = func(p string) error { h.password = p; return nil }

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/settings", h.mon.GetSettings)
	r.POST("/settings", h.mon.SetSettings)
	r.GET("/state", h.mon.GetState)
	return h, r
}

const goodSettings = `{"enabled":true,"host":"nas.lan","port":3493,"ups":"ups","username":"kvm","password":"hunter2",` +
	`"onBatteryMinutes":10,"forceOffMinutes":5,"powerOnRestore":true,"powerOnMinCharge":20,"powerOnDelaySeconds":90}`

func TestTheSettingsAreSavedAndTheSecretStaysOut(t *testing.T) {
	h, r := apiHarness(t)

	rec, env := call(t, r, http.MethodPost, "/settings", goodSettings)
	if env.Code != 0 {
		t.Fatalf("%+v", env)
	}
	if strings.Contains(rec.Body.String(), "hunter2") {
		t.Fatal("the response carries the password")
	}
	if h.password != "hunter2" {
		t.Fatalf("password %q", h.password)
	}
	want := config.NUT{
		Enabled: true, Host: "nas.lan", Port: 3493, UPS: "ups", Username: "kvm",
		OnBatteryMinutes: 10, ForceOffMinutes: 5, PowerOnRestore: true, PowerOnMinCharge: 20, PowerOnDelaySeconds: 90,
	}
	if h.settings != want {
		t.Fatalf("saved %+v", h.settings)
	}

	rec, env = call(t, r, http.MethodGet, "/settings", "")
	if strings.Contains(rec.Body.String(), "hunter2") {
		t.Fatal("the settings carry the password")
	}
	var got Settings
	if err := json.Unmarshal(env.Data, &got); err != nil {
		t.Fatal(err)
	}
	if !got.PasswordSet || got.Host != "nas.lan" || got.PowerOnDelaySeconds != 90 {
		t.Fatalf("%+v", got)
	}
}

func TestAnOmittedPasswordIsKeptAndAClearedOneRemoved(t *testing.T) {
	h, r := apiHarness(t)
	call(t, r, http.MethodPost, "/settings", goodSettings)

	keep := strings.Replace(goodSettings, `"password":"hunter2",`, "", 1)
	if _, env := call(t, r, http.MethodPost, "/settings", keep); env.Code != 0 {
		t.Fatalf("%+v", env)
	}
	if h.password != "hunter2" {
		t.Fatalf("password %q", h.password)
	}

	clear := strings.Replace(keep, `"enabled":true,`, `"enabled":true,"clearPassword":true,`, 1)
	if _, env := call(t, r, http.MethodPost, "/settings", clear); env.Code != 0 {
		t.Fatalf("%+v", env)
	}
	if h.password != "" {
		t.Fatalf("password %q", h.password)
	}
}

func TestBadSettingsAreRefused(t *testing.T) {
	h, r := apiHarness(t)

	for name, bad := range map[string]string{
		"no enabled":      strings.Replace(goodSettings, `"enabled":true,`, "", 1),
		"bad host":        strings.Replace(goodSettings, "nas.lan", "nas lan;x", 1),
		"bad ups":         strings.Replace(goodSettings, `"ups":"ups"`, `"ups":"a b"`, 1),
		"bad port":        strings.Replace(goodSettings, `"port":3493`, `"port":70000`, 1),
		"charge over 100": strings.Replace(goodSettings, `"powerOnMinCharge":20`, `"powerOnMinCharge":101`, 1),
		"delay too short": strings.Replace(goodSettings, `"powerOnDelaySeconds":90`, `"powerOnDelaySeconds":1`, 1),
		"password space":  strings.Replace(goodSettings, "hunter2", "hunt er2", 1),
		"enabled no host": strings.Replace(goodSettings, `"host":"nas.lan"`, `"host":""`, 1),
		"enabled no ups":  strings.Replace(goodSettings, `"ups":"ups"`, `"ups":""`, 1),
	} {
		if _, env := call(t, r, http.MethodPost, "/settings", bad); env.Code == 0 {
			t.Errorf("accepted %s", name)
		}
	}
	if h.settings != (config.NUT{}) || h.password != "" {
		t.Fatalf("a refused request changed the settings: %+v", h.settings)
	}
}

// A short press on a host that is off switches it on, so the client does not
// run without the power LED.
func TestItCannotBeEnabledWithoutThePowerLED(t *testing.T) {
	h, r := apiHarness(t)
	h.ledConnected = false

	if _, env := call(t, r, http.MethodPost, "/settings", goodSettings); env.Code == 0 {
		t.Fatal("enabled without the power LED")
	}
	off := strings.Replace(goodSettings, `"enabled":true`, `"enabled":false`, 1)
	if _, env := call(t, r, http.MethodPost, "/settings", off); env.Code != 0 {
		t.Fatalf("refused to save the settings while off: %+v", env)
	}
}

func TestTheStateIsServed(t *testing.T) {
	h, r := apiHarness(t)
	call(t, r, http.MethodPost, "/settings", goodSettings)
	h.settings.Host, h.settings.Port = h.srv.hostPort()
	h.srv.set("OB", 64, 840)
	h.tick(step)

	_, env := call(t, r, http.MethodGet, "/state", "")
	var st State
	if err := json.Unmarshal(env.Data, &st); err != nil {
		t.Fatal(err)
	}
	if st.UPS != UPSOnBattery || st.Charge == nil || *st.Charge != 64 || *st.RuntimeSeconds != 840 || st.Host != "on" {
		t.Fatalf("%+v", st)
	}
}

func TestThePasswordFile(t *testing.T) {
	f := FilePassword{Path: filepath.Join(t.TempDir(), "nut.passwd")}

	if p, err := f.Get(); err != nil || p != "" {
		t.Fatalf("%q %v", p, err)
	}
	if err := f.Set("hunter2"); err != nil {
		t.Fatal(err)
	}
	if p, err := f.Get(); err != nil || p != "hunter2" {
		t.Fatalf("%q %v", p, err)
	}
	// Windows does not carry Unix modes.
	if info, err := os.Stat(f.Path); err != nil || (runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0) {
		t.Fatalf("%v %v", info, err)
	}
	if err := f.Set("two words"); err == nil {
		t.Fatal("accepted a password with a space")
	}
	if err := f.Set(""); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(f.Path); !os.IsNotExist(err) {
		t.Fatal("the file is still there")
	}
	if err := f.Set(""); err != nil {
		t.Fatal("removing a missing file failed")
	}
}

func TestValidHost(t *testing.T) {
	for _, ok := range []string{"nas", "nas.lan", "10.0.0.5", "fe80::1", "my-nas.example.com"} {
		if !ValidHost(ok) {
			t.Errorf("refused %q", ok)
		}
	}
	for _, bad := range []string{"", "a b", "-x", "x-", "a;b", "http://x", "[::1]"} {
		if ValidHost(bad) {
			t.Errorf("accepted %q", bad)
		}
	}
}
