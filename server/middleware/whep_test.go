package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/apikey"
)

// issueViewKey returns a view-only key belonging to username.
func issueViewKey(t *testing.T, username string) string {
	t.Helper()

	secret, _, err := apikey.CreateWithScope("obs", username, apikey.ScopeView)
	if err != nil {
		t.Fatalf("failed to issue a view key: %s", err)
	}

	return secret
}

// withShares points the share links at a temporary file.
func withShares(t *testing.T) *authn.ShareStore {
	t.Helper()

	shares := authn.NewShareStore(filepath.Join(t.TempDir(), "share_links.json"))
	original := authn.DefaultShares
	authn.DefaultShares = shares
	t.Cleanup(func() { authn.DefaultShares = original })

	return shares
}

func whepRequest(authorization string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/api/whep", nil)
	if authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	return req
}

// A view-only key is a viewer wherever it is used: it reaches the video and
// nothing that drives the host.
func TestViewKeyIsAViewerOnEveryRoute(t *testing.T) {
	withAccounts(t)
	secret := issueViewKey(t, "admin")

	r := gin.New()
	ok := func(c *gin.Context) { c.String(http.StatusOK, "ok") }
	r.POST("/api/hid/paste", CheckToken(), ok)
	r.GET("/api/hid/ws", CheckToken(), ok)
	r.GET("/api/stream/h264", CheckToken(), ok)

	for _, tc := range []struct {
		method, path string
		want         int
	}{
		{http.MethodPost, "/api/hid/paste", http.StatusForbidden},
		{http.MethodGet, "/api/hid/ws", http.StatusForbidden},
		{http.MethodGet, "/api/stream/h264", http.StatusOK},
	} {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		req.Header.Set("Authorization", "Bearer "+secret)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != tc.want {
			t.Errorf("%s %s with a view key: got %d, want %d", tc.method, tc.path, w.Code, tc.want)
		}
	}
}

func TestWHEPRefusesAMissingOrWrongToken(t *testing.T) {
	withAccounts(t)
	withShares(t)
	issueViewKey(t, "admin")

	for _, authorization := range []string{"", "Bearer", "Bearer nkvm_not-a-real-key", "Bearer not-a-share-token", "Basic YWRtaW46YWRtaW4="} {
		if _, status := AuthenticateWHEP(whepRequest(authorization)); status != http.StatusUnauthorized {
			t.Errorf("Authorization %q: got %d, want 401", authorization, status)
		}
	}
}

// A key with the run of the API is not a WHEP credential: the token goes into
// an OBS profile or a shell history, and one that could also type on the host
// must not end up there.
func TestWHEPRefusesAFullKey(t *testing.T) {
	withAccounts(t)
	secret := issueKey(t, "admin")

	if _, status := AuthenticateWHEP(whepRequest("Bearer " + secret)); status != http.StatusForbidden {
		t.Fatalf("a full key: got %d, want 403", status)
	}
}

func TestWHEPTakesAViewKey(t *testing.T) {
	withAccounts(t)
	secret := issueViewKey(t, "alice")

	viewer, status := AuthenticateWHEP(whepRequest("Bearer " + secret))
	if status != 0 {
		t.Fatalf("a view key: got %d, want accepted", status)
	}
	if viewer.Principal.Username != "alice" || viewer.Principal.Role != authn.RoleViewer || viewer.Principal.Via != "apikey" || viewer.Principal.Key != "obs" {
		t.Fatalf("principal %+v, want alice as a viewer through the key obs", viewer.Principal)
	}
	if !viewer.Valid() {
		t.Fatal("the credential should still be good")
	}
}

// The key of a view-only account is view-only whatever its scope.
func TestWHEPTakesAViewerAccountsKey(t *testing.T) {
	store := withAccounts(t)
	if err := store.Create("guest", "guest-password", authn.RoleViewer); err != nil {
		t.Fatal(err)
	}
	secret := issueKey(t, "guest")

	if _, status := AuthenticateWHEP(whepRequest("Bearer " + secret)); status != 0 {
		t.Fatalf("a viewer account's key: got %d, want accepted", status)
	}
}

func TestWHEPCredentialLapsesWhenTheKeyIsRevoked(t *testing.T) {
	withAccounts(t)
	secret, key, err := apikey.CreateWithScope("obs", "admin", apikey.ScopeView)
	if err != nil {
		t.Fatal(err)
	}

	viewer, status := AuthenticateWHEP(whepRequest("Bearer " + secret))
	if status != 0 {
		t.Fatalf("got %d", status)
	}
	if err := apikey.RevokeAny(key.ID); err != nil {
		t.Fatal(err)
	}
	if viewer.Valid() {
		t.Fatal("a revoked key must end the session it opened")
	}
}

func TestWHEPTakesAShareLinkToken(t *testing.T) {
	withAccounts(t)
	shares := withShares(t)
	link, token, err := shares.Create("tv", "admin", time.Hour)
	if err != nil {
		t.Fatal(err)
	}

	viewer, status := AuthenticateWHEP(whepRequest("Bearer " + token))
	if status != 0 {
		t.Fatalf("a share token: got %d, want accepted", status)
	}
	if viewer.Principal.ShareID != link.ID || viewer.Principal.Role != authn.RoleViewer || viewer.Principal.Via != "share" {
		t.Fatalf("principal %+v, want the share link as a viewer", viewer.Principal)
	}

	if err := shares.Revoke(link.ID); err != nil {
		t.Fatal(err)
	}
	if viewer.Valid() {
		t.Fatal("a revoked share link must end the session it opened")
	}
}

// The session cookie is never a WHEP credential. The endpoint answers other
// origins (CORS), so a cookie it took would let any page a signed-in user
// visits pull the video.
func TestWHEPIgnoresTheSessionCookie(t *testing.T) {
	withAccounts(t)
	cookie, err := GenerateJWT("admin", 1)
	if err != nil {
		t.Fatal(err)
	}

	req := whepRequest("")
	req.AddCookie(&http.Cookie{Name: CookieName, Value: cookie})
	req.Header.Set("Origin", "https://evil.example")
	if _, status := AuthenticateWHEP(req); status != http.StatusUnauthorized {
		t.Fatalf("a cookie alone: got %d, want 401", status)
	}
}

// With authentication switched off the whole board is open, and WHEP with it.
func TestWHEPIsOpenWhenAuthenticationIsOff(t *testing.T) {
	withAccounts(t)
	conf := config.GetInstance()
	conf.Authentication = "disable"

	viewer, status := AuthenticateWHEP(whepRequest(""))
	if status != 0 {
		t.Fatalf("got %d, want accepted", status)
	}
	if viewer.Principal.Role != authn.RoleViewer {
		t.Fatalf("role %q, want viewer", viewer.Principal.Role)
	}
}
