package middleware

import (
	"context"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
)

// One goroutine rechecks every running session, and it goes away with the
// last of them rather than lingering for the life of the server.
func TestSessionRecheckStopsWithTheLastSession(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	user, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}

	registry := newSessionRegistry(10 * time.Millisecond)
	var unregisters []func()
	for range 3 {
		_, cancel := context.WithCancel(context.Background())
		defer cancel()
		unregisters = append(unregisters, registry.register(session{
			username:     user.Username,
			tokenVersion: user.TokenVersion,
			store:        store,
			cancel:       cancel,
		}))
	}
	for _, unregister := range unregisters {
		unregister()
	}

	deadline := time.Now().Add(2 * time.Second)
	for {
		registry.mutex.Lock()
		watching := registry.watching
		registry.mutex.Unlock()
		if !watching {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("recheck goroutine outlived the last session")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

// A session whose token is still good survives the recheck.
func TestSessionRecheckKeepsValidSessions(t *testing.T) {
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	user, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	registry := newSessionRegistry(time.Hour)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	unregister := registry.register(session{
		username:     user.Username,
		tokenVersion: user.TokenVersion,
		store:        store,
		cancel:       cancel,
	})
	defer unregister()

	registry.recheck()
	if ctx.Err() != nil {
		t.Fatal("valid session was cancelled")
	}

	if _, err = store.Revoke(user.Username); err != nil {
		t.Fatal(err)
	}
	registry.recheck()
	if ctx.Err() == nil {
		t.Fatal("revoked session survived the recheck")
	}
}

// Revoking through the store and the registry, as the user endpoints do,
// ends a running request at once, without waiting for the recheck.
func TestCheckTokenRequestEndsOnRevoke(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	user, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	restore := useTestAuthStore(t, store)
	defer restore()
	token, err := GenerateJWT(user.Username, user.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}

	entered := make(chan struct{})
	ended := make(chan error, 1)
	router := gin.New()
	router.GET("/stream", CheckToken(), func(c *gin.Context) {
		close(entered)
		select {
		case <-c.Request.Context().Done():
			ended <- c.Request.Context().Err()
		case <-time.After(5 * time.Second):
			ended <- nil
		}
		c.Status(http.StatusNoContent)
	})
	go requestWithToken(router, "/stream", token)
	<-entered

	if _, err = store.Revoke(user.Username); err != nil {
		t.Fatal(err)
	}
	RevokeUserSessions(user.Username)
	select {
	case err := <-ended:
		if err == nil {
			t.Fatal("request outlived the revoke")
		}
	case <-time.After(time.Second):
		t.Fatal("request outlived the revoke")
	}
}

// The request's context still ends when the token expires.
func TestCheckTokenRequestEndsAtTokenExpiry(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	user, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	restore := useTestAuthStore(t, store)
	defer restore()
	config.GetInstance().JWT.RefreshTokenDuration = 1
	token, err := GenerateJWT(user.Username, user.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}

	ended := make(chan bool, 1)
	router := gin.New()
	router.GET("/stream", CheckToken(), func(c *gin.Context) {
		select {
		case <-c.Request.Context().Done():
			ended <- true
		case <-time.After(5 * time.Second):
			ended <- false
		}
		c.Status(http.StatusNoContent)
	})
	requestWithToken(router, "/stream", token)
	if !<-ended {
		t.Fatal("request outlived its token")
	}
}
