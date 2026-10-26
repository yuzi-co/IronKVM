package router

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"

	"github.com/gin-gonic/gin"
)

type edidFixture struct {
	engine *gin.Engine
	token  string
}

func newEdidFixture(t *testing.T) edidFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, _, err := store.Authenticate("admin", "admin"); err != nil {
		t.Fatal(err)
	}
	originalStore := authn.DefaultStore
	authn.DefaultStore = store
	conf := config.GetInstance()
	originalAuthentication, originalSecret, originalDuration :=
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration
	conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration = "enable", "test-secret", 3600
	t.Cleanup(func() {
		authn.DefaultStore = originalStore
		conf.Authentication, conf.JWT.SecretKey, conf.JWT.RefreshTokenDuration =
			originalAuthentication, originalSecret, originalDuration
	})

	user, err := store.Get("admin")
	if err != nil {
		t.Fatal(err)
	}
	token, err := middleware.GenerateJWT(user.Username, user.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	edidRouter(r)
	return edidFixture{engine: r, token: token}
}

func (f edidFixture) do(t *testing.T, method, target string, body []byte, signedIn bool) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(method, target, bytes.NewReader(body))
	if signedIn {
		request.AddCookie(&http.Cookie{Name: middleware.CookieName, Value: f.token})
	}
	recorder := httptest.NewRecorder()
	f.engine.ServeHTTP(recorder, request)
	return recorder
}

// envelope decodes {code, msg, data} with data left raw.
func envelope(t *testing.T, recorder *httptest.ResponseRecorder) (int, string, json.RawMessage) {
	t.Helper()
	var rsp struct {
		Code int             `json:"code"`
		Msg  string          `json:"msg"`
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("not an envelope (%d): %s", recorder.Code, recorder.Body.String())
	}
	return rsp.Code, rsp.Msg, rsp.Data
}

func TestEdidRoutesNeedASession(t *testing.T) {
	f := newEdidFixture(t)
	for _, route := range [][2]string{
		{http.MethodGet, "/api/edid/current"},
		{http.MethodGet, "/api/edid/presets"},
		{http.MethodGet, "/api/edid/presets/sipeed-nanokvm"},
		{http.MethodGet, "/api/edid/presets/sipeed-nanokvm/bin"},
		{http.MethodPost, "/api/edid/decode"},
	} {
		if recorder := f.do(t, route[0], route[1], nil, false); recorder.Code != http.StatusUnauthorized {
			t.Errorf("%s %s without a session: %d", route[0], route[1], recorder.Code)
		}
	}
}

func TestEdidCurrentSaysNotReadable(t *testing.T) {
	f := newEdidFixture(t)
	code, _, data := envelope(t, f.do(t, http.MethodGet, "/api/edid/current", nil, true))
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	var cur struct {
		Readable      bool   `json:"readable"`
		Reason        string `json:"reason"`
		BuiltinPreset string `json:"builtinPreset"`
	}
	if err := json.Unmarshal(data, &cur); err != nil {
		t.Fatal(err)
	}
	if cur.Readable || cur.Reason == "" || cur.BuiltinPreset != "lt6911uxc-builtin" {
		t.Fatalf("current = %+v", cur)
	}
}

func TestEdidPresetsListAndDecode(t *testing.T) {
	f := newEdidFixture(t)

	code, _, data := envelope(t, f.do(t, http.MethodGet, "/api/edid/presets", nil, true))
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	var list struct {
		Presets []struct {
			ID         string   `json:"id"`
			Origin     string   `json:"origin"`
			Source     string   `json:"source"`
			Licence    string   `json:"licence"`
			CapturedOn []string `json:"capturedOn"`
		} `json:"presets"`
	}
	if err := json.Unmarshal(data, &list); err != nil {
		t.Fatal(err)
	}
	ids := map[string]bool{}
	for _, p := range list.Presets {
		ids[p.ID] = true
	}
	for _, id := range []string{"sipeed-nanokvm", "lt6911uxc-builtin", "1080p60"} {
		if !ids[id] {
			t.Errorf("preset %s not listed: %v", id, ids)
		}
	}

	code, _, data = envelope(t, f.do(t, http.MethodGet, "/api/edid/presets/sipeed-nanokvm", nil, true))
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	var one struct {
		Hex     string `json:"hex"`
		Decoded struct {
			Valid     bool `json:"valid"`
			Preferred struct {
				HActive int `json:"hActive"`
				VActive int `json:"vActive"`
			} `json:"preferred"`
		} `json:"decoded"`
	}
	if err := json.Unmarshal(data, &one); err != nil {
		t.Fatal(err)
	}
	stock, err := os.ReadFile("../../tools/nanokvm_update_edid/E21_NanoKVM.bin")
	if err != nil {
		t.Fatal(err)
	}
	if one.Hex != hex.EncodeToString(stock) || !one.Decoded.Valid ||
		one.Decoded.Preferred.HActive != 1920 || one.Decoded.Preferred.VActive != 1080 {
		t.Fatalf("preset = %+v", one)
	}

	if code, _, _ := envelope(t, f.do(t, http.MethodGet, "/api/edid/presets/nope", nil, true)); code == 0 {
		t.Error("an unknown preset answered code 0")
	}
}

func TestEdidPresetDownload(t *testing.T) {
	f := newEdidFixture(t)
	recorder := f.do(t, http.MethodGet, "/api/edid/presets/sipeed-nanokvm/bin", nil, true)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status %d", recorder.Code)
	}
	stock, err := os.ReadFile("../../tools/nanokvm_update_edid/E21_NanoKVM.bin")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(recorder.Body.Bytes(), stock) {
		t.Fatal("download differs from the preset")
	}
	if ct := recorder.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("content type %q", ct)
	}
	if cd := recorder.Header().Get("Content-Disposition"); !strings.Contains(cd, `filename="edid-sipeed-nanokvm.bin"`) {
		t.Errorf("content disposition %q", cd)
	}

	if recorder := f.do(t, http.MethodGet, "/api/edid/presets/nope/bin", nil, true); recorder.Code != http.StatusNotFound {
		t.Errorf("unknown preset: %d", recorder.Code)
	}
}

func TestEdidDecodeUpload(t *testing.T) {
	f := newEdidFixture(t)
	stock, err := os.ReadFile("../../tools/nanokvm_update_edid/E21_NanoKVM.bin")
	if err != nil {
		t.Fatal(err)
	}

	for name, body := range map[string][]byte{
		"binary": stock,
		"hex":    []byte(hex.EncodeToString(stock)),
	} {
		t.Run(name, func(t *testing.T) {
			code, msg, data := envelope(t, f.do(t, http.MethodPost, "/api/edid/decode", body, true))
			if code != 0 {
				t.Fatalf("code %d %s", code, msg)
			}
			var rsp struct {
				Hex     string `json:"hex"`
				Decoded struct {
					Valid bool   `json:"valid"`
					Name  string `json:"name"`
				} `json:"decoded"`
			}
			if err := json.Unmarshal(data, &rsp); err != nil {
				t.Fatal(err)
			}
			if rsp.Hex != hex.EncodeToString(stock) || !rsp.Decoded.Valid || rsp.Decoded.Name != "NanoKVM" {
				t.Fatalf("decoded = %+v", rsp)
			}
		})
	}

	// A checksum error still decodes, marked invalid.
	broken := append([]byte(nil), stock...)
	broken[0x7f]++
	code, _, data := envelope(t, f.do(t, http.MethodPost, "/api/edid/decode", broken, true))
	if code != 0 {
		t.Fatalf("checksum error: code %d", code)
	}
	if !strings.Contains(string(data), `"valid":false`) {
		t.Errorf("checksum error decoded as valid: %s", data)
	}

	for name, body := range map[string][]byte{
		"empty":     nil,
		"text":      []byte("not an edid"),
		"too large": make([]byte, 64<<10),
	} {
		t.Run(name, func(t *testing.T) {
			if code, msg, _ := envelope(t, f.do(t, http.MethodPost, "/api/edid/decode", body, true)); code == 0 {
				t.Fatalf("code 0 (%s)", msg)
			}
		})
	}
}
