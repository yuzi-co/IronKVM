package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/service/audit"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	testClientID     = "ironkvm-test"
	testClientSecret = "oidc-client-secret-sentinel"
	testKeyID        = "test-key"
)

// mockIdP is an OpenID provider in the test process: discovery, the key set,
// the authorization step (recorded, not served: the test plays the browser)
// and the token endpoint, which signs an ID token with a test key.
type mockIdP struct {
	t      *testing.T
	server *httptest.Server
	key    *rsa.PrivateKey

	mu sync.Mutex
	// codes are the authorization codes handed out, with what the
	// authorization request carried.
	codes map[string]authRequest
	// claims are the ID token's claims beyond iss, aud, exp, iat and nonce.
	claims map[string]any
	// tamper changes the ID token's standard claims before it is signed.
	tamper func(claims jwt.MapClaims)
	// signer, when set, signs the ID token instead of key.
	signer *rsa.PrivateKey
	// tokenRequests counts calls to the token endpoint.
	tokenRequests int
}

type authRequest struct {
	nonce         string
	challenge     string
	redirectURI   string
	codeChallenge string
}

func newMockIdP(t *testing.T) *mockIdP {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	idp := &mockIdP{t: t, key: key, codes: map[string]authRequest{}, claims: map[string]any{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		base := idp.server.URL
		_ = json.NewEncoder(w).Encode(map[string]any{
			"issuer":                                base,
			"authorization_endpoint":                base + "/authorize",
			"token_endpoint":                        base + "/token",
			"jwks_uri":                              base + "/jwks",
			"response_types_supported":              []string{"code"},
			"subject_types_supported":               []string{"public"},
			"id_token_signing_alg_values_supported": []string{"RS256"},
			"code_challenge_methods_supported":      []string{"S256"},
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		pub := idp.key.PublicKey
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{{
			"kty": "RSA",
			"alg": "RS256",
			"use": "sig",
			"kid": testKeyID,
			"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(pub.E)).Bytes()),
		}}})
	})
	mux.HandleFunc("/token", idp.token)
	idp.server = httptest.NewServer(mux)
	t.Cleanup(idp.server.Close)
	return idp
}

// authorize plays the IdP's login page: it records the request the board
// sent the browser to and returns the code the browser takes back.
func (idp *mockIdP) authorize(location string) (code, state string) {
	idp.t.Helper()
	u, err := url.Parse(location)
	if err != nil || !strings.HasPrefix(location, idp.server.URL+"/authorize") {
		idp.t.Fatalf("login redirected to %q, want the IdP's authorization endpoint", location)
	}
	q := u.Query()
	if q.Get("response_type") != "code" || q.Get("client_id") != testClientID {
		idp.t.Fatalf("authorization request %v", q)
	}
	if q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" {
		idp.t.Fatalf("authorization request without PKCE: %v", q)
	}
	if q.Get("state") == "" || q.Get("nonce") == "" {
		idp.t.Fatalf("authorization request without state or nonce: %v", q)
	}
	if !strings.Contains(" "+q.Get("scope")+" ", " openid ") {
		idp.t.Fatalf("scope %q without openid", q.Get("scope"))
	}
	code = "code-" + q.Get("state")
	idp.mu.Lock()
	idp.codes[code] = authRequest{
		nonce:         q.Get("nonce"),
		redirectURI:   q.Get("redirect_uri"),
		codeChallenge: q.Get("code_challenge"),
	}
	idp.mu.Unlock()
	return code, q.Get("state")
}

func (idp *mockIdP) token(w http.ResponseWriter, r *http.Request) {
	idp.mu.Lock()
	defer idp.mu.Unlock()
	idp.tokenRequests++
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	id, secret, ok := r.BasicAuth()
	if !ok {
		id, secret = r.PostForm.Get("client_id"), r.PostForm.Get("client_secret")
	}
	if id != testClientID || secret != testClientSecret {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_client"})
		return
	}
	req, found := idp.codes[r.PostForm.Get("code")]
	delete(idp.codes, r.PostForm.Get("code"))
	sum := sha256.Sum256([]byte(r.PostForm.Get("code_verifier")))
	if !found || r.PostForm.Get("redirect_uri") != req.redirectURI ||
		base64.RawURLEncoding.EncodeToString(sum[:]) != req.codeChallenge {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
		return
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"iss":   idp.server.URL,
		"sub":   "subject-1",
		"aud":   testClientID,
		"iat":   now.Unix(),
		"exp":   now.Add(5 * time.Minute).Unix(),
		"nonce": req.nonce,
	}
	for k, v := range idp.claims {
		claims[k] = v
	}
	if idp.tamper != nil {
		idp.tamper(claims)
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = testKeyID
	signer := idp.key
	if idp.signer != nil {
		signer = idp.signer
	}
	signed, err := token.SignedString(signer)
	if err != nil {
		idp.t.Errorf("sign: %s", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"access_token": "access",
		"token_type":   "Bearer",
		"expires_in":   300,
		"id_token":     signed,
	})
}

type oidcFixture struct {
	t      *testing.T
	idp    *mockIdP
	store  *authn.Store
	router *gin.Engine
	events []audit.Event
	mu     sync.Mutex
}

func newOIDCFixture(t *testing.T) *oidcFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	f := &oidcFixture{t: t, idp: newMockIdP(t)}
	f.store = authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	t.Cleanup(useTestStore(f.store))
	original := authn.DefaultSessions
	authn.DefaultSessions = authn.NewSessionStore(filepath.Join(t.TempDir(), "login_sessions.json"))
	t.Cleanup(func() { authn.DefaultSessions = original })
	t.Cleanup(useOIDCSettingsFile(filepath.Join(t.TempDir(), "oidc.json")))
	delay := oidcFailureDelay
	oidcFailureDelay = 0
	t.Cleanup(func() { oidcFailureDelay = delay })
	resetOIDCState()
	t.Cleanup(resetOIDCState)

	auditStore := audit.NewStore(t.TempDir())
	auditStore.SetSink(func(e audit.Event) {
		f.mu.Lock()
		f.events = append(f.events, e)
		f.mu.Unlock()
	})
	t.Cleanup(audit.UseStore(auditStore))

	if _, ok, err := f.store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err := f.store.Create("alice", "alice-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}
	f.save(func(s *oidcSettings) {})

	service := NewService()
	r := gin.New()
	r.POST("/login", service.Login)
	r.GET("/api/auth/oidc", service.OIDCInfo)
	r.GET("/api/auth/oidc/login", service.OIDCLogin)
	r.GET("/api/auth/oidc/callback", service.OIDCCallback)
	api := r.Group("/api").Use(middleware.CheckToken())
	api.GET("/auth/account", service.GetAccount)
	adminSession := r.Group("/api").Use(middleware.CheckSession(), middleware.RequireRole(authn.RoleAdmin))
	adminSession.GET("/auth/oidc/settings", service.GetOIDCSettings)
	adminSession.POST("/auth/oidc/settings", service.SetOIDCSettings)
	adminSession.DELETE("/auth/users/:username/oidc", service.UnlinkUserOIDC)
	f.router = r
	return f
}

// save writes the provider's settings, as edit leaves them, over a working
// default.
func (f *oidcFixture) save(edit func(s *oidcSettings)) {
	f.t.Helper()
	s := oidcSettings{
		Enabled:       true,
		ProviderName:  "Test IdP",
		Issuer:        f.idp.server.URL,
		ClientID:      testClientID,
		ClientSecret:  testClientSecret,
		Scopes:        []string{"openid", "profile", "email"},
		UsernameClaim: claimPreferredUsername,
		DefaultRole:   authn.RoleViewer,
		GroupsClaim:   "groups",
	}
	edit(&s)
	if err := saveOIDCSettings(s); err != nil {
		f.t.Fatal(err)
	}
	resetOIDCState()
}

// begin starts a login as a browser would, returning the IdP location and
// the cookie that binds the flow to this browser.
func (f *oidcFixture) begin() (string, *http.Cookie) {
	f.t.Helper()
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "http://kvm.lan/api/auth/oidc/login", nil)
	f.router.ServeHTTP(rec, req)
	if rec.Code != http.StatusFound {
		f.t.Fatalf("login start = %d %s", rec.Code, rec.Body.String())
	}
	var flow *http.Cookie
	for _, c := range rec.Result().Cookies() {
		if c.Name == oidcFlowCookie {
			flow = c
		}
	}
	if flow == nil || !flow.HttpOnly || flow.SameSite != http.SameSiteLaxMode {
		f.t.Fatalf("flow cookie %+v", flow)
	}
	return rec.Header().Get("Location"), flow
}

// callback returns the browser to the board with the code and state.
func (f *oidcFixture) callback(code, state string, flow *http.Cookie) *httptest.ResponseRecorder {
	f.t.Helper()
	q := url.Values{"code": {code}, "state": {state}}
	req := httptest.NewRequest(http.MethodGet, "http://kvm.lan/api/auth/oidc/callback?"+q.Encode(), nil)
	if flow != nil {
		req.AddCookie(flow)
	}
	rec := httptest.NewRecorder()
	f.router.ServeHTTP(rec, req)
	return rec
}

// login runs the whole flow and returns the callback's answer.
func (f *oidcFixture) login() *httptest.ResponseRecorder {
	f.t.Helper()
	location, flow := f.begin()
	code, state := f.idp.authorize(location)
	return f.callback(code, state, flow)
}

func sessionCookieOf(rec *httptest.ResponseRecorder) *http.Cookie {
	for _, c := range rec.Result().Cookies() {
		if c.Name == middleware.CookieName && c.Value != "" {
			return c
		}
	}
	return nil
}

func (f *oidcFixture) account(cookie *http.Cookie) (string, authn.Role) {
	f.t.Helper()
	rec := requestJSONRecorder(f.router, http.MethodGet, "/api/auth/account", nil, cookie)
	if rec.Code != http.StatusOK {
		return "", ""
	}
	var reply struct {
		Data struct {
			Username string     `json:"username"`
			Role     authn.Role `json:"role"`
		} `json:"data"`
	}
	_ = json.Unmarshal(rec.Body.Bytes(), &reply)
	return reply.Data.Username, reply.Data.Role
}

// refused requires that the callback signed nobody in and sent the browser
// back to the login page with the reason.
func (f *oidcFixture) refused(rec *httptest.ResponseRecorder, reason string) {
	f.t.Helper()
	if c := sessionCookieOf(rec); c != nil {
		f.t.Fatalf("refused login set a session cookie")
	}
	want := "/#/auth/login?oidc=" + reason
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != want {
		f.t.Fatalf("callback = %d to %q, want 302 to %q", rec.Code, rec.Header().Get("Location"), want)
	}
	if n := len(authn.DefaultSessions.List("")); n != 0 {
		f.t.Fatalf("%d sessions after a refused login", n)
	}
	e := f.lastEvent()
	if e.Action != "login-oidc" || e.Outcome == audit.OutcomeOK {
		f.t.Fatalf("refusal audited as %+v", e)
	}
}

func (f *oidcFixture) lastEvent() audit.Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	if len(f.events) == 0 {
		f.t.Fatal("nothing audited")
	}
	return f.events[len(f.events)-1]
}

func (f *oidcFixture) sessions(username string) int {
	return len(authn.DefaultSessions.List(username))
}

func TestOIDCLoginCreatesASession(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "alice"

	rec := f.login()

	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/" {
		t.Fatalf("callback = %d to %q: %s", rec.Code, rec.Header().Get("Location"), rec.Body.String())
	}
	cookie := sessionCookieOf(rec)
	if cookie == nil {
		t.Fatal("no session cookie")
	}
	if name, role := f.account(cookie); name != "alice" || role != authn.RoleUser {
		t.Fatalf("signed in as %q/%q", name, role)
	}
	if n := f.sessions("alice"); n != 1 {
		t.Fatalf("%d login sessions for alice, want 1", n)
	}
	e := f.lastEvent()
	if e.Action != "login-oidc" || e.Outcome != audit.OutcomeOK || e.Actor != "alice" || e.Via != "oidc" {
		t.Fatalf("audited %+v", e)
	}
	// The flow cookie is spent.
	for _, c := range rec.Result().Cookies() {
		if c.Name == oidcFlowCookie && c.MaxAge >= 0 {
			t.Fatalf("flow cookie kept: %+v", c)
		}
	}
}

func TestOIDCStateIsSingleUse(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "alice"
	location, flow := f.begin()
	code, state := f.idp.authorize(location)
	if rec := f.callback(code, state, flow); sessionCookieOf(rec) == nil {
		t.Fatal("first callback refused")
	}
	f.refused(f.callbackFresh(code, state, flow), "state")
}

// callbackFresh is a callback after the session store was emptied, so
// refused can count sessions.
func (f *oidcFixture) callbackFresh(code, state string, flow *http.Cookie) *httptest.ResponseRecorder {
	authn.DefaultSessions = authn.NewSessionStore(filepath.Join(f.t.TempDir(), "login_sessions.json"))
	return f.callback(code, state, flow)
}

func TestOIDCWrongStateIsRefused(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "alice"
	location, flow := f.begin()
	code, state := f.idp.authorize(location)

	f.refused(f.callback(code, state+"x", flow), "state")

	// A state that is real but was started in another browser: the login
	// CSRF case, where an attacker hands the victim their own callback.
	_, otherFlow := f.begin()
	f.refused(f.callback(code, state, otherFlow), "state")
	f.refused(f.callback(code, state, nil), "state")
	if f.idp.tokenRequests != 0 {
		t.Fatalf("code exchanged %d times for a refused state", f.idp.tokenRequests)
	}
}

func TestOIDCTokenChecks(t *testing.T) {
	other, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name   string
		tamper func(c jwt.MapClaims)
		signer *rsa.PrivateKey
	}{
		{"wrong nonce", func(c jwt.MapClaims) { c["nonce"] = "not-the-nonce" }, nil},
		{"no nonce", func(c jwt.MapClaims) { delete(c, "nonce") }, nil},
		{"wrong audience", func(c jwt.MapClaims) { c["aud"] = "another-client" }, nil},
		{"wrong issuer", func(c jwt.MapClaims) { c["iss"] = "https://evil.example" }, nil},
		{"expired", func(c jwt.MapClaims) {
			c["iat"] = time.Now().Add(-2 * time.Hour).Unix()
			c["exp"] = time.Now().Add(-time.Hour).Unix()
		}, nil},
		{"bad signature", nil, other},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := newOIDCFixture(t)
			f.idp.claims["preferred_username"] = "alice"
			f.idp.tamper = tc.tamper
			f.idp.signer = tc.signer
			f.refused(f.login(), "token")
		})
	}
}

func TestOIDCUnmappedUserIsRefused(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "mallory"
	f.refused(f.login(), "account")
	if _, err := f.store.Get("mallory"); err == nil {
		t.Fatal("an account was created with auto-create off")
	}

	// No claim at all, and a claim that is not a valid local name.
	delete(f.idp.claims, "preferred_username")
	f.refused(f.login(), "account")
	f.idp.claims["preferred_username"] = "../admin"
	f.refused(f.login(), "account")
}

func TestOIDCNeverSignsInAsTheDeviceOwnerOrADisabledUser(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "admin"
	f.refused(f.login(), "account")

	disabled := false
	if _, err := f.store.Update("", "alice", authn.UserPatch{Enabled: &disabled}); err != nil {
		t.Fatal(err)
	}
	f.idp.claims["preferred_username"] = "alice"
	f.refused(f.login(), "account")
}

func TestOIDCEmailClaimNeedsTheDomainAndAVerifiedAddress(t *testing.T) {
	f := newOIDCFixture(t)
	f.save(func(s *oidcSettings) {
		s.UsernameClaim = claimEmail
		s.EmailDomain = "corp.example"
	})
	f.idp.claims["email"] = "alice@evil.example"
	f.refused(f.login(), "account")

	f.idp.claims["email"] = "alice@corp.example"
	f.idp.claims["email_verified"] = false
	f.refused(f.login(), "account")

	f.idp.claims["email_verified"] = true
	f.idp.claims["email"] = "alice@Corp.Example"
	rec := f.login()
	if name, _ := f.account(sessionCookieOf(rec)); name != "alice" {
		t.Fatalf("email login signed in as %q", name)
	}
}

func TestOIDCAutoCreateAndGroupRoles(t *testing.T) {
	f := newOIDCFixture(t)
	f.save(func(s *oidcSettings) {
		s.AutoCreate = true
		s.AdminGroup = "kvm-admins"
		s.UserGroup = "kvm-users"
		s.ViewerGroup = "kvm-viewers"
	})

	// In no mapped group: refused, and no account is made.
	f.idp.claims["preferred_username"] = "carol"
	f.idp.claims["groups"] = []string{"staff"}
	f.refused(f.login(), "role")
	if _, err := f.store.Get("carol"); err == nil {
		t.Fatal("carol was created without a role")
	}

	// The highest mapped group wins.
	f.idp.claims["groups"] = []string{"staff", "kvm-viewers", "kvm-users"}
	rec := f.login()
	if name, role := f.account(sessionCookieOf(rec)); name != "carol" || role != authn.RoleUser {
		t.Fatalf("created %q/%q", name, role)
	}
	carol, err := f.store.Get("carol")
	if err != nil || carol.Role != authn.RoleUser {
		t.Fatalf("carol = %+v, %v", carol, err)
	}
	// An account made for OIDC has no password anyone knows.
	if _, ok, _ := f.store.Authenticate("carol", ""); ok {
		t.Fatal("empty password accepted")
	}

	// The IdP's groups set the role of an existing account at each login.
	f.idp.claims["preferred_username"] = "alice"
	f.idp.claims["groups"] = "kvm-viewers"
	rec = f.login()
	if name, role := f.account(sessionCookieOf(rec)); name != "alice" || role != authn.RoleViewer {
		t.Fatalf("alice signed in as %q/%q", name, role)
	}
}

func TestOIDCAutoCreateUsesTheDefaultRoleWithoutGroups(t *testing.T) {
	f := newOIDCFixture(t)
	f.save(func(s *oidcSettings) { s.AutoCreate = true })
	f.idp.claims["preferred_username"] = "dave"
	rec := f.login()
	if name, role := f.account(sessionCookieOf(rec)); name != "dave" || role != authn.RoleViewer {
		t.Fatalf("created %q/%q", name, role)
	}
}

func TestOIDCSkipsLocalTOTP(t *testing.T) {
	f := newOIDCFixture(t)
	if _, err := f.store.SetTOTP("alice", "sealed-secret", nil); err != nil {
		t.Fatal(err)
	}
	f.idp.claims["preferred_username"] = "alice"
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("OIDC login asked for the local second factor")
	}
}

func TestOIDCDisabledOrAuthOffRefusesTheFlow(t *testing.T) {
	f := newOIDCFixture(t)
	f.save(func(s *oidcSettings) { s.Enabled = false })
	rec := httptest.NewRecorder()
	f.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/auth/oidc/login", nil))
	if rec.Code != http.StatusFound || rec.Header().Get("Location") != "/#/auth/login?oidc=disabled" {
		t.Fatalf("disabled start = %d to %q", rec.Code, rec.Header().Get("Location"))
	}
}

func TestOIDCCallbackIsRateLimited(t *testing.T) {
	f := newOIDCFixture(t)
	conf := config.GetInstance()
	duration, failures := conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures
	conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures = 60, 3
	t.Cleanup(func() {
		conf.Security.LoginLockoutDuration, conf.Security.LoginMaxFailures = duration, failures
		ClearLoginAttempt("192.0.2.1")
	})
	ClearLoginAttempt("192.0.2.1")

	for i := 0; i < 3; i++ {
		f.refused(f.callback("code", "forged", nil), "state")
	}
	// Locked now: even a good flow is turned away, and the IdP is not asked.
	f.idp.claims["preferred_username"] = "alice"
	location, flow := f.begin()
	code, state := f.idp.authorize(location)
	f.refused(f.callback(code, state, flow), "locked")
	if f.idp.tokenRequests != 0 {
		t.Fatalf("locked callback exchanged the code")
	}
}

func TestOIDCInfoIsPublicAndCarriesNoSecret(t *testing.T) {
	f := newOIDCFixture(t)
	rec := requestJSONRecorder(f.router, http.MethodGet, "/api/auth/oidc", nil, nil)
	body := rec.Body.String()
	if !strings.Contains(body, `"enabled":true`) || !strings.Contains(body, "Test IdP") {
		t.Fatalf("info = %s", body)
	}
	for _, leak := range []string{testClientSecret, testClientID, f.idp.server.URL} {
		if strings.Contains(body, leak) {
			t.Fatalf("public info holds %q: %s", leak, body)
		}
	}
}

func TestOIDCSettingsNeverReturnTheSecret(t *testing.T) {
	f := newOIDCFixture(t)
	admin := loginCookie(t, f.router, "admin", "admin")

	rec := requestJSONRecorder(f.router, http.MethodGet, "/api/auth/oidc/settings", nil, admin)
	body := rec.Body.String()
	if strings.Contains(body, testClientSecret) {
		t.Fatalf("settings leak the secret: %s", body)
	}
	if !strings.Contains(body, `"clientSecretSet":true`) || !strings.Contains(body, `"redirectUrl":"http://example.com/api/auth/oidc/callback"`) {
		t.Fatalf("settings = %s", body)
	}

	// Saving without a secret keeps the one on file, and the answer to the
	// save does not carry it either.
	update := map[string]any{
		"enabled": true, "providerName": "Corp", "issuer": f.idp.server.URL, "clientId": testClientID,
		"scopes": []string{"profile"}, "usernameClaim": "preferred_username", "defaultRole": "viewer",
	}
	rec = requestJSONRecorder(f.router, http.MethodPost, "/api/auth/oidc/settings", update, admin)
	if strings.Contains(rec.Body.String(), testClientSecret) || !strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("save = %s", rec.Body.String())
	}
	s, err := loadOIDCSettings()
	if err != nil || s.ClientSecret != testClientSecret || s.ProviderName != "Corp" {
		t.Fatalf("after save %+v, %v", s, err)
	}
	if s.Scopes[0] != "openid" {
		t.Fatalf("scopes %v without openid first", s.Scopes)
	}
	info, err := os.Stat(oidcSettingsFile)
	if err != nil || info.Mode().Perm() != 0o600 {
		t.Fatalf("settings file mode %v, %v", info.Mode(), err)
	}

	// A new secret replaces it; clearing removes it.
	update["clientSecret"] = "new-secret"
	requestJSONRecorder(f.router, http.MethodPost, "/api/auth/oidc/settings", update, admin)
	if s, _ = loadOIDCSettings(); s.ClientSecret != "new-secret" {
		t.Fatalf("secret %q after replace", s.ClientSecret)
	}
	delete(update, "clientSecret")
	update["clearClientSecret"] = true
	requestJSONRecorder(f.router, http.MethodPost, "/api/auth/oidc/settings", update, admin)
	if s, _ = loadOIDCSettings(); s.ClientSecret != "" {
		t.Fatalf("secret %q after clear", s.ClientSecret)
	}

	// A bad issuer, a bad claim or a bad role is refused.
	for field, value := range map[string]any{"issuer": "ftp://x", "usernameClaim": "sub", "defaultRole": "root"} {
		bad := map[string]any{}
		for k, v := range update {
			bad[k] = v
		}
		bad[field] = value
		rec = requestJSONRecorder(f.router, http.MethodPost, "/api/auth/oidc/settings", bad, admin)
		if strings.Contains(rec.Body.String(), `"code":0`) {
			t.Fatalf("%s=%v accepted", field, value)
		}
	}
}

func TestOIDCRedirectURLOverride(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "alice"
	const override = "https://kvm.example.org/api/auth/oidc/callback"
	f.save(func(s *oidcSettings) { s.RedirectURL = override })

	// The IdP is sent the override, and the code exchange carries it too.
	location, flow := f.begin()
	u, err := url.Parse(location)
	if err != nil || u.Query().Get("redirect_uri") != override {
		t.Fatalf("authorization request %q, want redirect_uri %q", location, override)
	}
	code, state := f.idp.authorize(location)
	if rec := f.callback(code, state, flow); sessionCookieOf(rec) == nil {
		t.Fatalf("callback refused: %d %s", rec.Code, rec.Body.String())
	}

	admin := loginCookie(t, f.router, "admin", "admin")
	body := requestJSONRecorder(f.router, http.MethodGet, "/api/auth/oidc/settings", nil, admin).Body.String()
	if !strings.Contains(body, `"redirectUrl":"`+override+`"`) || !strings.Contains(body, `"redirectUrlOverride":"`+override+`"`) {
		t.Fatalf("settings = %s", body)
	}

	update := map[string]any{
		"enabled": true, "issuer": f.idp.server.URL, "clientId": testClientID,
		"usernameClaim": "preferred_username", "defaultRole": "viewer",
		"redirectUrlOverride": "kvm.example.org/callback",
	}
	rec := requestJSONRecorder(f.router, http.MethodPost, "/api/auth/oidc/settings", update, admin)
	if strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("a redirect URL without a scheme was accepted: %s", rec.Body.String())
	}

	// Saving without one goes back to the address the request came in on.
	update["redirectUrlOverride"] = ""
	rec = requestJSONRecorder(f.router, http.MethodPost, "/api/auth/oidc/settings", update, admin)
	if !strings.Contains(rec.Body.String(), `"redirectUrl":"http://example.com/api/auth/oidc/callback"`) {
		t.Fatalf("save without override = %s", rec.Body.String())
	}
}

// loginFresh is a login after the session store was emptied, so refused can
// count sessions after earlier logins succeeded.
func (f *oidcFixture) loginFresh() *httptest.ResponseRecorder {
	authn.DefaultSessions = authn.NewSessionStore(filepath.Join(f.t.TempDir(), "login_sessions.json"))
	return f.login()
}

// eventsOf returns the audited events with an action.
func (f *oidcFixture) eventsOf(action string) []audit.Event {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []audit.Event
	for _, e := range f.events {
		if e.Action == action {
			out = append(out, e)
		}
	}
	return out
}

// link returns the issuer and sub an account is linked to.
func (f *oidcFixture) link(username string) (string, string) {
	f.t.Helper()
	user, err := f.store.Get(username)
	if err != nil {
		f.t.Fatal(err)
	}
	return user.OIDCIssuer, user.OIDCSubject
}

func TestOIDCFirstLoginLinksTheAccountAndLaterOnesMustMatch(t *testing.T) {
	f := newOIDCFixture(t)
	f.save(func(s *oidcSettings) { s.UserGroup = "kvm-users"; s.ViewerGroup = "kvm-viewers" })
	f.idp.claims["preferred_username"] = "alice"
	f.idp.claims["groups"] = "kvm-users"

	// An account from before the link existed signs in, and is linked.
	if issuer, sub := f.link("alice"); issuer != "" || sub != "" {
		t.Fatalf("alice linked before any login: %q %q", issuer, sub)
	}
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("first login refused")
	}
	if issuer, sub := f.link("alice"); issuer != f.idp.server.URL || sub != "subject-1" {
		t.Fatalf("alice linked to %q %q", issuer, sub)
	}
	if links := f.eventsOf("oidc-link"); len(links) != 1 || links[0].Target != "alice" || !strings.Contains(links[0].Detail, f.idp.server.URL) {
		t.Fatalf("link audited as %+v", links)
	}

	// The same IdP account signs in again, and nothing is linked anew.
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("second login refused")
	}
	if n := len(f.eventsOf("oidc-link")); n != 1 {
		t.Fatalf("%d link events after a second login", n)
	}

	// Another IdP account that took the name alice is refused, and its
	// groups do not touch alice's role.
	f.idp.claims["sub"] = "subject-2"
	f.idp.claims["groups"] = "kvm-viewers"
	f.refused(f.loginFresh(), "binding")
	if e := f.lastEvent(); e.Outcome != audit.OutcomeDenied || e.Actor != "alice" || !strings.Contains(e.Detail, "subject-2") {
		t.Fatalf("mismatch audited as %+v", e)
	}
	alice, _ := f.store.Get("alice")
	if alice.Role != authn.RoleUser || alice.OIDCSubject != "subject-1" {
		t.Fatalf("after a refused login alice is %s, linked to %q", alice.Role, alice.OIDCSubject)
	}
}

func TestOIDCAdministratorUnlinksAnAccount(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "alice"
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("first login refused")
	}

	// The user was recreated at the IdP, with a new sub.
	f.idp.claims["sub"] = "subject-2"
	f.refused(f.loginFresh(), "binding")

	// Only an administrator unlinks.
	alice := loginCookie(t, f.router, "alice", "alice-password")
	if code := requestJSON(f.router, http.MethodDelete, "/api/auth/users/alice/oidc", nil, alice); code != http.StatusForbidden {
		t.Fatalf("a user unlinked an account: %d", code)
	}
	admin := loginCookie(t, f.router, "admin", "admin")
	rec := requestJSONRecorder(f.router, http.MethodDelete, "/api/auth/users/alice/oidc", nil, admin)
	if !strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("unlink = %s", rec.Body.String())
	}
	if issuer, sub := f.link("alice"); issuer != "" || sub != "" {
		t.Fatalf("alice still linked to %q %q", issuer, sub)
	}
	// The unlink ends alice's sessions, and not the administrator's.
	if name, _ := f.account(alice); name != "" {
		t.Fatalf("alice's session survived the unlink as %q", name)
	}
	if name, _ := f.account(admin); name != "admin" {
		t.Fatalf("the administrator's session ended: %q", name)
	}
	rec = requestJSONRecorder(f.router, http.MethodDelete, "/api/auth/users/nobody/oidc", nil, admin)
	if strings.Contains(rec.Body.String(), `"code":0`) {
		t.Fatalf("unlinked an account that does not exist: %s", rec.Body.String())
	}

	// The next login links the new IdP account.
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("login after unlink refused")
	}
	if _, sub := f.link("alice"); sub != "subject-2" {
		t.Fatalf("alice linked to %q after unlink", sub)
	}
}

func TestOIDCAutoCreatedAccountIsLinkedAtCreation(t *testing.T) {
	f := newOIDCFixture(t)
	f.save(func(s *oidcSettings) { s.AutoCreate = true })
	f.idp.claims["preferred_username"] = "dave"
	f.idp.claims["sub"] = "dave-sub"
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("auto-create login refused")
	}
	if issuer, sub := f.link("dave"); issuer != f.idp.server.URL || sub != "dave-sub" {
		t.Fatalf("dave created linked to %q %q", issuer, sub)
	}
	if links := f.eventsOf("oidc-link"); len(links) != 1 || !strings.Contains(links[0].Detail, "created") {
		t.Fatalf("creation audited as %+v", links)
	}

	f.idp.claims["sub"] = "someone-else"
	f.refused(f.loginFresh(), "binding")
}

// A sub means nothing outside its issuer: after the administrator points the
// settings at another provider, each account links again at its next login
// instead of every one being locked out.
func TestOIDCIssuerChangeLinksAccountsAgain(t *testing.T) {
	f := newOIDCFixture(t)
	const oldIssuer = "https://old-idp.example"
	if _, _, err := f.store.BindOIDC("alice", oldIssuer, "old-sub"); err != nil {
		t.Fatal(err)
	}
	f.idp.claims["preferred_username"] = "alice"
	if sessionCookieOf(f.login()) == nil {
		t.Fatal("login under the new issuer refused")
	}
	if issuer, sub := f.link("alice"); issuer != f.idp.server.URL || sub != "subject-1" {
		t.Fatalf("alice linked to %q %q", issuer, sub)
	}
	links := f.eventsOf("oidc-link")
	if len(links) != 1 || !strings.Contains(links[0].Detail, "in place of "+oldIssuer) {
		t.Fatalf("relink audited as %+v", links)
	}
	// And under the new issuer the link holds.
	f.idp.claims["sub"] = "subject-2"
	f.refused(f.loginFresh(), "binding")
}

func TestOIDCTokenWithoutSubIsRefused(t *testing.T) {
	f := newOIDCFixture(t)
	f.idp.claims["preferred_username"] = "alice"
	f.idp.tamper = func(c jwt.MapClaims) { delete(c, "sub") }
	f.refused(f.login(), "token")
	if _, sub := f.link("alice"); sub != "" {
		t.Fatalf("alice linked to %q without a sub", sub)
	}
}
