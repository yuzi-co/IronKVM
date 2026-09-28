package auth

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/apikey"
)

type apiKeysEnvelope struct {
	Code int `json:"code"`
	Data struct {
		ID   string `json:"id"`
		Keys []struct {
			ID       string `json:"id"`
			Name     string `json:"name"`
			Username string `json:"username"`
		} `json:"keys"`
	} `json:"data"`
}

func apiKeyRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	t.Cleanup(restore)
	if err := store.Create("alice", "alice-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}

	original := apikey.File
	apikey.File = filepath.Join(t.TempDir(), "api_keys.json")
	t.Cleanup(func() { apikey.File = original })

	service := NewService()
	router := gin.New()
	router.POST("/login", service.Login)
	keys := router.Group("/").Use(middleware.CheckSession())
	keys.GET("/api-keys", service.GetAPIKeys)
	keys.POST("/api-keys", service.CreateAPIKey)
	keys.DELETE("/api-keys/:id", service.DeleteAPIKey)
	return router
}

func apiKeyCall(t *testing.T, router http.Handler, method, path string, body any, cookie *http.Cookie) apiKeysEnvelope {
	t.Helper()
	recorder := requestJSONRecorder(router, method, path, body, cookie)
	if recorder.Code != http.StatusOK {
		t.Fatalf("%s %s status = %d", method, path, recorder.Code)
	}
	var envelope apiKeysEnvelope
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("%s %s: %s", method, path, err)
	}
	return envelope
}

func TestAUserSeesAndRevokesOnlyTheirOwnKeys(t *testing.T) {
	router := apiKeyRouter(t)
	adminCookie := loginCookie(t, router, "admin", "admin")
	aliceCookie := loginCookie(t, router, "alice", "alice-password")

	adminKey := apiKeyCall(t, router, http.MethodPost, "/api-keys", map[string]string{"name": "prometheus"}, adminCookie).Data.ID
	apiKeyCall(t, router, http.MethodPost, "/api-keys", map[string]string{"name": "script"}, aliceCookie)

	listed := apiKeyCall(t, router, http.MethodGet, "/api-keys", nil, aliceCookie).Data.Keys
	if len(listed) != 1 || listed[0].Name != "script" {
		t.Fatalf("alice listed %+v, want only her own key", listed)
	}

	if rsp := apiKeyCall(t, router, http.MethodDelete, "/api-keys/"+adminKey, nil, aliceCookie); rsp.Code == 0 {
		t.Fatal("alice revoked a key that belongs to admin")
	}

	listed = apiKeyCall(t, router, http.MethodGet, "/api-keys", nil, adminCookie).Data.Keys
	if len(listed) != 2 {
		t.Fatalf("admin key must survive alice's revoke; admin listed %+v", listed)
	}
}

func TestAnAdministratorSeesEveryKeyWithItsOwnerAndCanRevokeIt(t *testing.T) {
	router := apiKeyRouter(t)
	adminCookie := loginCookie(t, router, "admin", "admin")
	aliceCookie := loginCookie(t, router, "alice", "alice-password")

	apiKeyCall(t, router, http.MethodPost, "/api-keys", map[string]string{"name": "prometheus"}, adminCookie)
	aliceKey := apiKeyCall(t, router, http.MethodPost, "/api-keys", map[string]string{"name": "script"}, aliceCookie).Data.ID

	listed := apiKeyCall(t, router, http.MethodGet, "/api-keys", nil, adminCookie).Data.Keys
	owners := map[string]string{}
	for _, key := range listed {
		owners[key.Name] = key.Username
	}
	if len(listed) != 2 || owners["prometheus"] != "admin" || owners["script"] != "alice" {
		t.Fatalf("admin listed %+v, want every key with its owner", listed)
	}

	if rsp := apiKeyCall(t, router, http.MethodDelete, "/api-keys/"+aliceKey, nil, adminCookie); rsp.Code != 0 {
		t.Fatalf("admin revoke of alice's key returned code %d", rsp.Code)
	}

	if listed := apiKeyCall(t, router, http.MethodGet, "/api-keys", nil, aliceCookie).Data.Keys; len(listed) != 0 {
		t.Fatalf("alice still holds %+v after admin revoked her key", listed)
	}
}
