package wireguard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

type apiResult struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
	raw  string
}

func newTestAPI(t *testing.T) (*gin.Engine, *fakeSystem) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	useTempStore(t)
	f := newFake()
	s := &Service{m: &Manager{sys: f, iface: "wg0"}}
	r := gin.New()
	s.Register(r.Group("/api"))
	return r, f
}

func call(t *testing.T, r *gin.Engine, method, path string, body any) apiResult {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		data, _ := json.Marshal(body)
		reader = bytes.NewReader(data)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("%s %s: HTTP %d", method, path, w.Code)
	}
	var res apiResult
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatal(err)
	}
	res.raw = w.Body.String()
	return res
}

func noSecrets(t *testing.T, res apiResult) {
	t.Helper()
	for _, secret := range []string{testPrivate, testPSK} {
		if strings.Contains(res.raw, secret) {
			t.Fatalf("a secret is in the answer: %s", res.raw)
		}
	}
}

func TestImportStoresTheKeyAndNeverReturnsIt(t *testing.T) {
	r, f := newTestAPI(t)

	res := call(t, r, "POST", "/api/network/wireguard/import", map[string]string{"config": sample()})
	if res.Code != 0 {
		t.Fatalf("import: %s", res.Msg)
	}
	noSecrets(t, res)
	if !strings.Contains(res.raw, "PostUp") {
		t.Errorf("the ignored PostUp is not reported: %s", res.raw)
	}
	if !strings.Contains(res.raw, `"hasPresharedKey":true`) {
		t.Errorf("no preshared key flag: %s", res.raw)
	}
	if !strings.Contains(res.raw, mustPublic(t, testPrivate)) {
		t.Errorf("the board's public key is not shown: %s", res.raw)
	}
	// Importing does not bring the tunnel up by itself.
	if len(f.calls) != 0 {
		t.Errorf("calls %v", f.calls)
	}

	stored, err := Load()
	if err != nil || stored.PrivateKey != testPrivate || stored.Peers[0].PresharedKey != testPSK {
		t.Fatalf("stored %+v %v", stored, err)
	}

	res = call(t, r, "GET", "/api/network/wireguard", nil)
	if res.Code != 0 {
		t.Fatal(res.Msg)
	}
	noSecrets(t, res)
}

func TestImportRefusesABadFile(t *testing.T) {
	r, _ := newTestAPI(t)
	res := call(t, r, "POST", "/api/network/wireguard/import", map[string]string{"config": "[Interface]\nPrivateKey = x\n"})
	if res.Code == 0 {
		t.Fatal("accepted")
	}
	if c, _ := Load(); c.PrivateKey != "" {
		t.Error("a refused file was stored")
	}
}

func TestSaveGeneratesAKeyAndKeepsSecrets(t *testing.T) {
	r, _ := newTestAPI(t)

	req := map[string]any{
		"addresses": []string{"10.8.0.2/24"},
		"peers": []map[string]any{{
			"publicKey":    testPeerKey,
			"presharedKey": testPSK,
			"endpoint":     "203.0.113.5:51820",
			"allowedIPs":   []string{"10.8.0.0/24"},
		}},
	}
	res := call(t, r, "POST", "/api/network/wireguard", req)
	if res.Code != 0 {
		t.Fatal(res.Msg)
	}
	noSecrets(t, res)
	first, _ := Load()
	if first.PrivateKey == "" {
		t.Fatal("no key was generated")
	}
	if strings.Contains(res.raw, first.PrivateKey) {
		t.Fatal("the generated key is in the answer")
	}

	// A save without the keys keeps both.
	req["peers"] = []map[string]any{{
		"publicKey":           testPeerKey,
		"endpoint":            "203.0.113.5:51820",
		"allowedIPs":          []string{"10.8.0.0/24"},
		"persistentKeepalive": 25,
	}}
	if res := call(t, r, "POST", "/api/network/wireguard", req); res.Code != 0 {
		t.Fatal(res.Msg)
	}
	second, _ := Load()
	if second.PrivateKey != first.PrivateKey || second.Peers[0].PresharedKey != testPSK || second.Peers[0].PersistentKeepalive != 25 {
		t.Errorf("stored %+v", second)
	}

	// clearPresharedKey drops the preshared key, and a given private key replaces the board's.
	req["privateKey"] = testPrivate
	req["peers"] = []map[string]any{{
		"publicKey":         testPeerKey,
		"clearPresharedKey": true,
		"allowedIPs":        []string{"10.8.0.0/24"},
	}}
	if res := call(t, r, "POST", "/api/network/wireguard", req); res.Code != 0 {
		t.Fatal(res.Msg)
	}
	third, _ := Load()
	if third.PrivateKey != testPrivate || third.Peers[0].PresharedKey != "" {
		t.Errorf("stored %+v", third)
	}
}

func TestSaveRefusesInvalidFields(t *testing.T) {
	r, _ := newTestAPI(t)
	res := call(t, r, "POST", "/api/network/wireguard", map[string]any{
		"addresses": []string{"10.8.0.2/24"},
		"peers": []map[string]any{{
			"publicKey":           testPeerKey,
			"allowedIPs":          []string{"10.8.0.0/24"},
			"persistentKeepalive": 70000,
		}},
	})
	if res.Code == 0 {
		t.Fatal("accepted")
	}
}

func TestUpDownAndBoot(t *testing.T) {
	r, f := newTestAPI(t)
	if res := call(t, r, "POST", "/api/network/wireguard/up", nil); res.Code == 0 {
		t.Fatal("up without a configuration")
	}
	if err := Save(validConfig()); err != nil {
		t.Fatal(err)
	}

	res := call(t, r, "POST", "/api/network/wireguard/up", nil)
	if res.Code != 0 {
		t.Fatal(res.Msg)
	}
	if !f.links["wg0"] {
		t.Fatal("not up")
	}

	// A save while up applies the change.
	f.calls = nil
	res = call(t, r, "POST", "/api/network/wireguard", map[string]any{
		"addresses": []string{"10.8.0.3/24"},
		"peers":     []map[string]any{{"publicKey": testPeerKey, "allowedIPs": []string{"10.8.0.0/24"}}},
	})
	if res.Code != 0 || !f.has("addAddr wg0 10.8.0.3/24") {
		t.Fatalf("%s %v", res.Msg, f.calls)
	}

	if res := call(t, r, "POST", "/api/network/wireguard/boot", map[string]bool{"enabled": true}); res.Code != 0 {
		t.Fatal(res.Msg)
	}
	if c, _ := Load(); !c.StartAtBoot {
		t.Error("start at boot not saved")
	}

	if res := call(t, r, "POST", "/api/network/wireguard/down", nil); res.Code != 0 {
		t.Fatal(res.Msg)
	}
	if f.links["wg0"] {
		t.Error("still up")
	}
}

func TestRegenerateKey(t *testing.T) {
	r, _ := newTestAPI(t)
	if err := Save(validConfig()); err != nil {
		t.Fatal(err)
	}
	res := call(t, r, "POST", "/api/network/wireguard/key", nil)
	if res.Code != 0 {
		t.Fatal(res.Msg)
	}
	c, _ := Load()
	if c.PrivateKey == testPrivate || c.PrivateKey == "" {
		t.Error("the key did not change")
	}
	if strings.Contains(res.raw, c.PrivateKey) {
		t.Error("the new key is in the answer")
	}
}

func TestDeleteRemovesTheConfiguration(t *testing.T) {
	r, f := newTestAPI(t)
	if err := Save(validConfig()); err != nil {
		t.Fatal(err)
	}
	f.links["wg0"] = true
	if res := call(t, r, "DELETE", "/api/network/wireguard", nil); res.Code != 0 {
		t.Fatal(res.Msg)
	}
	if f.links["wg0"] {
		t.Error("still up")
	}
	if c, _ := Load(); c.PrivateKey != "" {
		t.Error("the configuration is still there")
	}
}
