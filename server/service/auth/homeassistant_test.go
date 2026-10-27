package auth

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
)

// The Home Assistant nanokvm integration logs in the way Sipeed's API expects
// (ironkvm-dist#124): code -2 is a wrong password, and a successful login
// carries a token, in the nano-kvm-token cookie or in data.token.

func TestLoginWrongPasswordIsCodeMinusTwo(t *testing.T) {
	gin.SetMode(gin.TestMode)
	restore := useTestStore(authn.NewStore(filepath.Join(t.TempDir(), "pwd")))
	defer restore()
	// httptest requests come from 192.0.2.1.
	ClearLoginAttempt("192.0.2.1")
	defer ClearLoginAttempt("192.0.2.1")

	router := gin.New()
	router.POST("/login", NewService().Login)

	recorder := requestJSONRecorder(router, http.MethodPost, "/login",
		map[string]any{"username": "admin", "password": encryptForRequest("wrong")}, nil)
	var rsp proto.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &rsp); err != nil {
		t.Fatal(err)
	}
	if recorder.Code != http.StatusOK || rsp.Code != -2 {
		t.Fatalf("wrong password: status %d, code %d; the client reads code -2 as a wrong password", recorder.Code, rsp.Code)
	}
}

func TestLoginWithAuthenticationOffAnswersAToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	conf := config.GetInstance()
	original := conf.Authentication
	conf.Authentication = "disable"
	t.Cleanup(func() { conf.Authentication = original })

	router := gin.New()
	router.POST("/login", NewService().Login)

	recorder := requestJSONRecorder(router, http.MethodPost, "/login",
		map[string]any{"username": "admin", "password": "x"}, nil)
	var rsp struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &rsp); err != nil {
		t.Fatal(err)
	}
	if rsp.Code != 0 || rsp.Data.Token != proto.LoginDisabledToken {
		t.Fatalf("login with authentication off: %s", recorder.Body.String())
	}
}
