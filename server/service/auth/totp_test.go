package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/authn/totp"
	"NanoKVM-Server/middleware"
)

type totpFixture struct {
	t       *testing.T
	store   *authn.Store
	router  *gin.Engine
	clock   time.Time
	keyPath string
}

func newTOTPFixture(t *testing.T) *totpFixture {
	t.Helper()
	gin.SetMode(gin.TestMode)
	dir := t.TempDir()
	store := authn.NewStore(filepath.Join(dir, "pwd"))
	t.Cleanup(useTestStore(store))
	if _, ok, err := store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err := store.Create("alice", "alice-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}

	f := &totpFixture{t: t, store: store, clock: time.Unix(1_800_000_000, 0), keyPath: filepath.Join(dir, "totp.key")}

	originalGuard, originalKeyring, originalDelay := guard, totpKeyring, codeFailureDelay
	guard = newTOTPGuard()
	guard.now = func() time.Time { return f.clock }
	totpKeyring = authn.NewKeyring(f.keyPath)
	codeFailureDelay = 0
	t.Cleanup(func() { guard, totpKeyring, codeFailureDelay = originalGuard, originalKeyring, originalDelay })

	service := NewService()
	router := gin.New()
	router.POST("/login", service.Login)
	router.POST("/login/totp", service.LoginTOTP)
	authenticated := router.Group("/").Use(middleware.CheckToken())
	authenticated.GET("/account", service.GetAccount)
	session := router.Group("/").Use(middleware.CheckSession())
	session.GET("/totp", service.GetTOTP)
	session.POST("/totp/setup", service.SetupTOTP)
	session.POST("/totp/enable", service.EnableTOTP)
	session.POST("/totp/disable", service.DisableTOTP)
	session.POST("/totp/recovery-codes", service.RegenerateRecoveryCodes)
	adminSession := router.Group("/").Use(middleware.CheckSession(), middleware.RequireRole(authn.RoleAdmin))
	adminSession.DELETE("/users/:username/totp", service.ResetUserTOTP)
	f.router = router
	return f
}

type reply struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (f *totpFixture) call(method, path string, body any, cookie *http.Cookie) (reply, *http.Cookie) {
	f.t.Helper()
	recorder := requestJSONRecorder(f.router, method, path, body, cookie)
	if recorder.Code != http.StatusOK {
		f.t.Fatalf("%s %s: status %d", method, path, recorder.Code)
	}
	var r reply
	if err := json.Unmarshal(recorder.Body.Bytes(), &r); err != nil {
		f.t.Fatalf("%s %s: %v", method, path, err)
	}
	for _, c := range recorder.Result().Cookies() {
		if c.Name == middleware.CookieName && c.Value != "" {
			return r, c
		}
	}
	return r, nil
}

// password logs in with a password and returns the reply and any cookie.
func (f *totpFixture) password(username, password string) (reply, *http.Cookie) {
	return f.call(http.MethodPost, "/login", map[string]any{
		"username": username,
		"password": encryptForRequest(password),
	}, nil)
}

func (f *totpFixture) challenge(username, password string) string {
	f.t.Helper()
	r, cookie := f.password(username, password)
	if r.Code != codeTOTPRequired || cookie != nil {
		f.t.Fatalf("password step: code %d, cookie %v", r.Code, cookie != nil)
	}
	var data struct {
		TOTPRequired bool   `json:"totpRequired"`
		Challenge    string `json:"challenge"`
	}
	if err := json.Unmarshal(r.Data, &data); err != nil || !data.TOTPRequired || data.Challenge == "" {
		f.t.Fatalf("challenge reply %s: %v", r.Data, err)
	}
	return data.Challenge
}

func (f *totpFixture) code(challenge, code string) (reply, *http.Cookie) {
	return f.call(http.MethodPost, "/login/totp", map[string]any{"challenge": challenge, "code": code}, nil)
}

// enrol turns TOTP on for alice and returns her secret and recovery codes.
func (f *totpFixture) enrol() ([]byte, []string) {
	f.t.Helper()
	_, cookie := f.password("alice", "alice-password")
	if cookie == nil {
		f.t.Fatal("no session before enrolment")
	}

	r, _ := f.call(http.MethodPost, "/totp/setup", map[string]any{"password": encryptForRequest("wrong-password")}, cookie)
	if r.Code != -2 {
		f.t.Fatalf("setup with a wrong password: code %d", r.Code)
	}
	r, _ = f.call(http.MethodPost, "/totp/setup", map[string]any{"password": encryptForRequest("alice-password")}, cookie)
	if r.Code != 0 {
		f.t.Fatalf("setup: code %d %s", r.Code, r.Msg)
	}
	var setup struct {
		Secret string `json:"secret"`
		URI    string `json:"uri"`
	}
	if err := json.Unmarshal(r.Data, &setup); err != nil {
		f.t.Fatal(err)
	}
	if !strings.HasPrefix(setup.URI, "otpauth://totp/IronKVM:alice") || !strings.Contains(setup.URI, setup.Secret) {
		f.t.Fatalf("unexpected uri %s", setup.URI)
	}
	secret, err := totp.DecodeSecret(setup.Secret)
	if err != nil {
		f.t.Fatal(err)
	}

	// Not in force until a code is typed back.
	if r, _ := f.password("alice", "alice-password"); r.Code != 0 {
		f.t.Fatalf("login before confirmation: code %d", r.Code)
	}

	r, _ = f.call(http.MethodPost, "/totp/enable", map[string]any{"code": "000000"}, cookie)
	if r.Code != -2 && totp.Code(secret, f.clock) != "000000" {
		f.t.Fatalf("enable with a wrong code: code %d", r.Code)
	}
	r, _ = f.call(http.MethodPost, "/totp/enable", map[string]any{"code": totp.Code(secret, f.clock)}, cookie)
	if r.Code != 0 {
		f.t.Fatalf("enable: code %d %s", r.Code, r.Msg)
	}
	var codes struct {
		RecoveryCodes []string `json:"recoveryCodes"`
	}
	if err := json.Unmarshal(r.Data, &codes); err != nil || len(codes.RecoveryCodes) != totp.RecoveryCodeCount {
		f.t.Fatalf("recovery codes %s: %v", r.Data, err)
	}

	r, _ = f.call(http.MethodGet, "/totp", nil, cookie)
	if !strings.Contains(string(r.Data), `"enabled":true`) {
		f.t.Fatalf("state after enable: %s", r.Data)
	}
	// The enrolling code is spent, so move on a step before logging in.
	f.clock = f.clock.Add(totp.Period)
	return secret, codes.RecoveryCodes
}

func TestTOTPEnrolmentAndLogin(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()

	challenge := f.challenge("alice", "alice-password")
	if r, cookie := f.code(challenge, "123456"); r.Code != -2 || cookie != nil {
		if totp.Code(secret, f.clock) != "123456" {
			t.Fatalf("wrong code: code %d, cookie %v", r.Code, cookie != nil)
		}
	}
	r, cookie := f.code(challenge, totp.Code(secret, f.clock))
	if r.Code != 0 || cookie == nil {
		t.Fatalf("right code: code %d %s", r.Code, r.Msg)
	}
	r, _ = f.call(http.MethodGet, "/account", nil, cookie)
	if !strings.Contains(string(r.Data), `"totpEnabled":true`) {
		t.Fatalf("account: %s", r.Data)
	}
	// The challenge is spent.
	if r, _ := f.code(challenge, totp.Code(secret, f.clock)); r.Code != codeTOTPChallengeGo {
		t.Fatalf("challenge reused: code %d", r.Code)
	}
}

func TestTOTPSecretIsSealedOnDisk(t *testing.T) {
	f := newTOTPFixture(t)
	secret, codes := f.enrol()
	data, err := os.ReadFile(filepath.Join(filepath.Dir(f.keyPath), "pwd"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), totp.EncodeSecret(secret)) {
		t.Fatal("the account file holds the secret in the clear")
	}
	for _, code := range codes {
		if strings.Contains(string(data), code) {
			t.Fatal("the account file holds a recovery code in the clear")
		}
	}
	if _, err := os.Stat(f.keyPath); err != nil {
		t.Fatalf("no sealing key: %v", err)
	}
}

func TestTOTPClockSkew(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()

	// A phone one step behind and one step ahead both work.
	for _, offset := range []time.Duration{-totp.Period, totp.Period} {
		f.clock = f.clock.Add(3 * totp.Period)
		challenge := f.challenge("alice", "alice-password")
		if r, _ := f.code(challenge, totp.Code(secret, f.clock.Add(offset))); r.Code != 0 {
			t.Fatalf("offset %s refused: code %d", offset, r.Code)
		}
	}
	// Two steps off does not.
	f.clock = f.clock.Add(3 * totp.Period)
	challenge := f.challenge("alice", "alice-password")
	if r, _ := f.code(challenge, totp.Code(secret, f.clock.Add(-2*totp.Period))); r.Code != -2 {
		t.Fatalf("two steps behind: code %d", r.Code)
	}
}

func TestTOTPReplayRefused(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()

	code := totp.Code(secret, f.clock)
	if r, _ := f.code(f.challenge("alice", "alice-password"), code); r.Code != 0 {
		t.Fatalf("first use: code %d", r.Code)
	}
	// Same code, new login, same step and the next one: both refused.
	if r, _ := f.code(f.challenge("alice", "alice-password"), code); r.Code != -2 {
		t.Fatalf("replay in the same step: code %d", r.Code)
	}
	f.clock = f.clock.Add(totp.Period)
	if r, _ := f.code(f.challenge("alice", "alice-password"), code); r.Code != -2 {
		t.Fatalf("replay one step later: code %d", r.Code)
	}
	if r, _ := f.code(f.challenge("alice", "alice-password"), totp.Code(secret, f.clock)); r.Code != 0 {
		t.Fatalf("the next code: code %d", r.Code)
	}
}

func TestTOTPRateLimit(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()
	wrong := "000000"
	if totp.Code(secret, f.clock) == wrong {
		wrong = "111111"
	}

	// A challenge takes challengeAttempts guesses and is then gone.
	challenge := f.challenge("alice", "alice-password")
	for i := 0; i < challengeAttempts-1; i++ {
		if r, _ := f.code(challenge, wrong); r.Code != -2 {
			t.Fatalf("guess %d: code %d", i, r.Code)
		}
	}
	// The fifth wrong code in a row locks the account's code step.
	if r, _ := f.code(challenge, wrong); r.Code != -5 {
		t.Fatalf("fifth guess: code %d", r.Code)
	}
	if r, _ := f.code(challenge, totp.Code(secret, f.clock)); r.Code != codeTOTPChallengeGo {
		t.Fatalf("challenge after its last guess: code %d", r.Code)
	}

	// While locked, even the right code on a fresh challenge is refused.
	fresh := f.challenge("alice", "alice-password")
	if r, cookie := f.code(fresh, totp.Code(secret, f.clock)); r.Code != -5 || cookie != nil {
		t.Fatalf("right code while locked: code %d", r.Code)
	}

	// After the lockout it works again.
	f.clock = f.clock.Add(userLockout + time.Second)
	fresh = f.challenge("alice", "alice-password")
	if r, cookie := f.code(fresh, totp.Code(secret, f.clock)); r.Code != 0 || cookie == nil {
		t.Fatalf("right code after the lockout: code %d", r.Code)
	}
}

func TestTOTPChallengeExpires(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()
	challenge := f.challenge("alice", "alice-password")
	f.clock = f.clock.Add(challengeTTL)
	if r, _ := f.code(challenge, totp.Code(secret, f.clock)); r.Code != codeTOTPChallengeGo {
		t.Fatalf("expired challenge: code %d", r.Code)
	}
}

// A password change between the steps ends the half finished login.
func TestTOTPChallengeFollowsTokenVersion(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()
	challenge := f.challenge("alice", "alice-password")
	if _, err := f.store.Revoke("alice"); err != nil {
		t.Fatal(err)
	}
	if r, _ := f.code(challenge, totp.Code(secret, f.clock)); r.Code != codeTOTPChallengeGo {
		t.Fatalf("challenge after revocation: code %d", r.Code)
	}
}

func TestTOTPRecoveryCodeWorksOnce(t *testing.T) {
	f := newTOTPFixture(t)
	_, codes := f.enrol()
	typed := strings.ToLower(codes[3])
	if r, cookie := f.code(f.challenge("alice", "alice-password"), typed); r.Code != 0 || cookie == nil {
		t.Fatalf("recovery code: code %d", r.Code)
	}
	if r, _ := f.code(f.challenge("alice", "alice-password"), codes[3]); r.Code != -2 {
		t.Fatalf("recovery code reused: code %d", r.Code)
	}
	user, err := f.store.Get("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(user.RecoveryCodes) != totp.RecoveryCodeCount-1 {
		t.Fatalf("%d codes left", len(user.RecoveryCodes))
	}
}

func TestTOTPDisableNeedsPassword(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()
	_, cookie := f.code(f.challenge("alice", "alice-password"), totp.Code(secret, f.clock))
	if cookie == nil {
		t.Fatal("no session")
	}
	if r, _ := f.call(http.MethodPost, "/totp/disable", map[string]any{"password": encryptForRequest("nope-nope")}, cookie); r.Code != -2 {
		t.Fatalf("disable with a wrong password: code %d", r.Code)
	}
	if r, _ := f.call(http.MethodPost, "/totp/disable", map[string]any{"password": encryptForRequest("alice-password")}, cookie); r.Code != 0 {
		t.Fatalf("disable: code %d", r.Code)
	}
	if r, cookie := f.password("alice", "alice-password"); r.Code != 0 || cookie == nil {
		t.Fatalf("login after disable: code %d", r.Code)
	}
}

func TestTOTPRegenerateRecoveryCodes(t *testing.T) {
	f := newTOTPFixture(t)
	secret, old := f.enrol()
	_, cookie := f.code(f.challenge("alice", "alice-password"), totp.Code(secret, f.clock))
	r, _ := f.call(http.MethodPost, "/totp/recovery-codes", map[string]any{"password": encryptForRequest("alice-password")}, cookie)
	if r.Code != 0 {
		t.Fatalf("regenerate: code %d", r.Code)
	}
	var codes struct {
		RecoveryCodes []string `json:"recoveryCodes"`
	}
	if err := json.Unmarshal(r.Data, &codes); err != nil || len(codes.RecoveryCodes) != totp.RecoveryCodeCount {
		t.Fatalf("codes %s: %v", r.Data, err)
	}
	if r, _ := f.code(f.challenge("alice", "alice-password"), old[0]); r.Code != -2 {
		t.Fatalf("an old code still works: code %d", r.Code)
	}
	if r, _ := f.code(f.challenge("alice", "alice-password"), codes.RecoveryCodes[0]); r.Code != 0 {
		t.Fatalf("a new code: code %d", r.Code)
	}
}

func TestTOTPAdminReset(t *testing.T) {
	f := newTOTPFixture(t)
	secret, _ := f.enrol()
	_, aliceCookie := f.code(f.challenge("alice", "alice-password"), totp.Code(secret, f.clock))

	// Not by a user.
	if code := requestJSON(f.router, http.MethodDelete, "/users/alice/totp", nil, aliceCookie); code != http.StatusForbidden {
		t.Fatalf("reset by a user: status %d", code)
	}

	_, adminCookie := f.password("admin", "admin")
	if r, _ := f.call(http.MethodDelete, "/users/alice/totp", nil, adminCookie); r.Code != 0 {
		t.Fatalf("reset: code %d", r.Code)
	}
	if r, cookie := f.password("alice", "alice-password"); r.Code != 0 || cookie == nil {
		t.Fatalf("login after reset: code %d", r.Code)
	}
	if r, _ := f.call(http.MethodDelete, "/users/admin/totp", nil, adminCookie); r.Code != -2 {
		t.Fatalf("an administrator reset their own: code %d", r.Code)
	}
	if r, _ := f.call(http.MethodDelete, "/users/nobody/totp", nil, adminCookie); r.Code == 0 {
		t.Fatal("reset of a missing user succeeded")
	}
}

// An API key never reaches the routes that change the second factor.
func TestTOTPRoutesRefuseWithoutSession(t *testing.T) {
	f := newTOTPFixture(t)
	for _, path := range []string{"/totp/setup", "/totp/enable", "/totp/disable", "/totp/recovery-codes"} {
		if code := requestJSON(f.router, http.MethodPost, path, map[string]any{}, nil); code != http.StatusUnauthorized {
			t.Fatalf("%s without a session: status %d", path, code)
		}
	}
}

// The password step must not answer with a session for a TOTP account, and
// wrong passwords still answer as before.
func TestTOTPPasswordStep(t *testing.T) {
	f := newTOTPFixture(t)
	f.enrol()
	if r, cookie := f.password("alice", "wrong-password"); r.Code != -2 || cookie != nil {
		t.Fatalf("wrong password: code %d", r.Code)
	}
	// Other accounts are untouched.
	if r, cookie := f.password("admin", "admin"); r.Code != 0 || cookie == nil {
		t.Fatalf("admin login: code %d", r.Code)
	}
}
