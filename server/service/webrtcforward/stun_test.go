package webrtcforward

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
)

type stunAnswer struct {
	Code int                 `json:"code"`
	Msg  string              `json:"msg"`
	Data proto.StunServerRsp `json:"data"`
}

// stunCall runs one request through a handler and decodes the answer. Saves go
// to saved instead of server.yaml.
func stunCall(t *testing.T, handler gin.HandlerFunc, method, body string, saved *[]string, saveErr error) stunAnswer {
	t.Helper()
	gin.SetMode(gin.TestMode)

	oldSave := saveStun
	saveStun = func(stun string) error {
		if saveErr != nil {
			return saveErr
		}
		*saved = append(*saved, stun)
		return nil
	}
	t.Cleanup(func() { saveStun = oldSave })

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	c.Request = httptest.NewRequest(method, "/api/network/webrtc/stun", strings.NewReader(body))
	if body != "" {
		c.Request.Header.Set("Content-Type", "application/json")
	}
	handler(c)

	if recorder.Code != http.StatusOK {
		t.Fatalf("HTTP %d", recorder.Code)
	}
	var answer stunAnswer
	if err := json.Unmarshal(recorder.Body.Bytes(), &answer); err != nil {
		t.Fatalf("answer %q: %v", recorder.Body.String(), err)
	}
	return answer
}

func keepStun(t *testing.T) {
	t.Helper()
	conf := config.GetInstance()
	old := conf.Stun
	t.Cleanup(func() { conf.Stun = old })
}

func TestStunGetShowsTheSettingAndTheDefault(t *testing.T) {
	keepStun(t)
	config.GetInstance().Stun = config.DefaultStun

	var saved []string
	answer := stunCall(t, NewService().GetStun, http.MethodGet, "", &saved, nil)
	if answer.Code != 0 {
		t.Fatalf("code %d: %s", answer.Code, answer.Msg)
	}
	if answer.Data.Stun != config.DefaultStun || answer.Data.Default != config.DefaultStun {
		t.Fatalf("got %+v", answer.Data)
	}
}

func TestStunSetSavesAndAppliesValidValues(t *testing.T) {
	for _, tc := range []struct {
		name string
		body string
		want string
	}{
		{"own server", `{"stun":" stun.example.com:3478 "}`, "stun.example.com:3478"},
		{"off", `{"stun":"disable"}`, config.StunDisabled},
		{"empty is the default", `{"stun":""}`, config.DefaultStun},
		{"default", `{"stun":"stun.l.google.com:19302"}`, config.DefaultStun},
		{"IPv6", `{"stun":"[2001:db8::1]:3478"}`, "[2001:db8::1]:3478"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			keepStun(t)
			config.GetInstance().Stun = "before.example.com:1"

			var saved []string
			answer := stunCall(t, NewService().SetStun, http.MethodPost, tc.body, &saved, nil)
			if answer.Code != 0 {
				t.Fatalf("code %d: %s", answer.Code, answer.Msg)
			}
			if len(saved) != 1 || saved[0] != tc.want {
				t.Fatalf("saved %v, want [%s]", saved, tc.want)
			}
			if got := config.GetInstance().Stun; got != tc.want {
				t.Fatalf("running setting %q, want %q", got, tc.want)
			}
			if answer.Data.Stun != tc.want {
				t.Fatalf("answer %+v, want stun %q", answer.Data, tc.want)
			}
		})
	}
}

func TestStunSetRefusesBadValues(t *testing.T) {
	for _, body := range []string{
		`{"stun":"stun.example.com"}`,
		`{"stun":"stun.example.com:0"}`,
		`{"stun":"stun.example.com:65536"}`,
		`{"stun":"user:pass@stun.example.com:3478"}`,
		`{"stun":"stun:stun.example.com:3478"}`,
		`{"stun":"bad host:3478"}`,
		`{"stun":"off"}`,
		`not json`,
	} {
		t.Run(body, func(t *testing.T) {
			keepStun(t)
			config.GetInstance().Stun = "before.example.com:1"

			var saved []string
			answer := stunCall(t, NewService().SetStun, http.MethodPost, body, &saved, nil)
			if answer.Code == 0 {
				t.Fatalf("accepted %s", body)
			}
			if len(saved) != 0 {
				t.Fatalf("saved %v for a refused value", saved)
			}
			if got := config.GetInstance().Stun; got != "before.example.com:1" {
				t.Fatalf("running setting changed to %q", got)
			}
		})
	}
}

func TestStunSetKeepsTheRunningSettingWhenTheSaveFails(t *testing.T) {
	keepStun(t)
	config.GetInstance().Stun = "before.example.com:1"

	var saved []string
	answer := stunCall(t, NewService().SetStun, http.MethodPost, `{"stun":"disable"}`, &saved, errors.New("disk full"))
	if answer.Code == 0 {
		t.Fatal("a failed save was reported as success")
	}
	if got := config.GetInstance().Stun; got != "before.example.com:1" {
		t.Fatalf("running setting changed to %q after a failed save", got)
	}
}
