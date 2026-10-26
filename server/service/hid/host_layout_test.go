package hid

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
)

// useHostLayoutFile points HostLayoutFile at a file in a temporary directory
// for the test, with content unless it is empty.
func useHostLayoutFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "keyboard-layout")
	if content != "" {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	old := HostLayoutFile
	HostLayoutFile = path
	t.Cleanup(func() { HostLayoutFile = old })
	return path
}

func TestHostLayoutID(t *testing.T) {
	for _, tt := range []struct {
		name, content, want string
		set                 bool
	}{
		{"no file", "", "us", false},
		{"set", "de\n", "de", true},
		{"alias", "en-gb", "uk", true},
		{"unknown", "xx", "us", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			useHostLayoutFile(t, tt.content)
			if id, set := HostLayoutID(); id != tt.want || set != tt.set {
				t.Fatalf("HostLayoutID() = %q, %v, want %q, %v", id, set, tt.want, tt.set)
			}
		})
	}
}

func TestTypingLayout(t *testing.T) {
	useHostLayoutFile(t, "fr")

	if layout, err := TypingLayout(""); err != nil || layout.ID != "fr" {
		t.Fatalf("no layout named: %v, %v, want the host layout fr", layout, err)
	}
	if layout, err := TypingLayout("DE"); err != nil || layout.ID != "de" {
		t.Fatalf("de: %v, %v", layout, err)
	}
	_, err := TypingLayout("klingon")
	if err == nil || !strings.Contains(err.Error(), `"klingon"`) || !strings.Contains(err.Error(), "us, uk, de") {
		t.Fatalf("unknown layout error = %v, want it to name the layout and the valid IDs", err)
	}
}

func getJSON(t *testing.T, handler gin.HandlerFunc) proto.Response {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/hid/host-layout", nil)
	handler(c)

	var rsp proto.Response
	if err := json.Unmarshal(w.Body.Bytes(), &rsp); err != nil {
		t.Fatalf("response %q: %v", w.Body.String(), err)
	}
	return rsp
}

func hostLayoutOf(t *testing.T, rsp proto.Response) proto.GetHostLayoutRsp {
	t.Helper()
	if rsp.Code != 0 {
		t.Fatalf("code = %d (%s)", rsp.Code, rsp.Msg)
	}
	data, _ := json.Marshal(rsp.Data)
	var got proto.GetHostLayoutRsp
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	return got
}

func TestHostLayoutHandlers(t *testing.T) {
	path := useHostLayoutFile(t, "")
	s := &Service{}

	if got := hostLayoutOf(t, getJSON(t, s.GetHostLayout)); got.Layout != "us" || got.Configured {
		t.Fatalf("unset: %+v, want us and not configured", got)
	}

	if rsp := postJSON(t, s.SetHostLayout, map[string]any{"layout": "pt-BR"}); rsp.Code != 0 {
		t.Fatalf("set pt-BR: code = %d (%s)", rsp.Code, rsp.Msg)
	}
	if data, _ := os.ReadFile(path); string(data) != "pt-br" {
		t.Fatalf("file = %q, want the layout's ID", data)
	}
	if got := hostLayoutOf(t, getJSON(t, s.GetHostLayout)); got.Layout != "pt-br" || !got.Configured {
		t.Fatalf("set: %+v", got)
	}

	if rsp := postJSON(t, s.SetHostLayout, map[string]any{"layout": "xx"}); rsp.Code == 0 {
		t.Fatal("an unknown layout was accepted")
	}
	if data, _ := os.ReadFile(path); string(data) != "pt-br" {
		t.Fatalf("a refused layout changed the file to %q", data)
	}

	if rsp := postJSON(t, s.SetHostLayout, map[string]any{"layout": ""}); rsp.Code != 0 {
		t.Fatalf("reset: code = %d (%s)", rsp.Code, rsp.Msg)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("reset kept the file: %v", err)
	}
}
