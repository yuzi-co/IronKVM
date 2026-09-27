package redfish

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/proto"
)

// fakeHost stands in for the buttons, the power LED and the drives.
type fakeHost struct {
	mu sync.Mutex

	// led is the power LED; nil means the board has none wired.
	led      *bool
	pressErr error
	presses  []string

	// pressDelay is how long a press takes. offAfterPower turns the LED off
	// once a power press is done, as a host that shuts down would.
	pressDelay    time.Duration
	offAfterPower bool

	drives    []proto.DriveInfo
	insertErr error
	ejectErr  error
	inserts   []string
	ejects    []string
}

func (f *fakeHost) pressButton(kind string, d time.Duration) error {
	f.mu.Lock()
	delay := f.pressDelay
	f.mu.Unlock()
	time.Sleep(delay)

	f.mu.Lock()
	defer f.mu.Unlock()
	if f.pressErr != nil {
		return f.pressErr
	}
	f.presses = append(f.presses, fmt.Sprintf("%s %s", kind, d))
	if f.offAfterPower && kind == ButtonPower {
		off := false
		f.led = &off
	}
	return nil
}

func (f *fakeHost) powerLED() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.led == nil {
		return false, fmt.Errorf("no led")
	}
	return *f.led, nil
}

func (f *fakeHost) setLED(on *bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.led = on
}

func (f *fakeHost) listDrives() ([]proto.DriveInfo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]proto.DriveInfo{}, f.drives...), nil
}

func (f *fakeHost) insertDrive(id string, file string, ro bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.insertErr != nil {
		return f.insertErr
	}
	f.inserts = append(f.inserts, fmt.Sprintf("%s %s ro=%t", id, file, ro))
	for i := range f.drives {
		if f.drives[i].ID == id {
			f.drives[i].File = file
			f.drives[i].Ro = ro
		}
	}
	return nil
}

func (f *fakeHost) ejectDrive(id string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ejectErr != nil {
		return f.ejectErr
	}
	f.ejects = append(f.ejects, id)
	for i := range f.drives {
		if f.drives[i].ID == id {
			f.drives[i].File = ""
		}
	}
	return nil
}

// fakeLimiter records the brute-force calls and locks the addresses in locked.
type fakeLimiter struct {
	mu        sync.Mutex
	locked    map[string]bool
	failed    []string
	succeeded []string
}

func (f *fakeLimiter) Locked(ip string) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.locked[ip]
}

func (f *fakeLimiter) Failed(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.failed = append(f.failed, ip)
}

func (f *fakeLimiter) Succeeded(ip string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.succeeded = append(f.succeeded, ip)
}

// countingAccounts counts the password checks that reach the store.
type countingAccounts struct {
	*authn.Store
	mu     sync.Mutex
	checks int
}

func (a *countingAccounts) Authenticate(username, password string) (*authn.User, bool, error) {
	a.mu.Lock()
	a.checks++
	a.mu.Unlock()
	return a.Store.Authenticate(username, password)
}

func (a *countingAccounts) passwordChecks() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.checks
}

const testUUID = "8e1b3a52-6c1f-4c55-9a8e-0f3c2d1b4a77"

// clientIP is where every test request comes from.
const clientIP = "192.0.2.10"

type harness struct {
	t        *testing.T
	service  *Service
	engine   *gin.Engine
	host     *fakeHost
	limiter  *fakeLimiter
	accounts *authn.Store
	counted  *countingAccounts
	// keys maps an API key secret to its account.
	keys         map[string]string
	clock        time.Time
	authDisabled bool
	imageDir     string
	nics         []NIC
}

// newHarness builds a service on fakes, with two accounts in a real
// authn.Store: admin/admin (admin) and alice/valid-password (user).
func newHarness(t *testing.T) *harness {
	t.Helper()
	gin.SetMode(gin.TestMode)

	settle := resetSettle
	resetSettle = time.Millisecond
	t.Cleanup(func() { resetSettle = settle })

	accounts := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, ok, err := accounts.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err := accounts.Create("alice", "valid-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}

	on := true
	h := &harness{
		t:        t,
		host:     &fakeHost{led: &on},
		limiter:  &fakeLimiter{locked: map[string]bool{}},
		accounts: accounts,
		counted:  &countingAccounts{Store: accounts},
		keys:     map[string]string{},
		clock:    time.Date(2026, 9, 27, 12, 0, 0, 0, time.UTC),
		imageDir: t.TempDir(),
		nics:     []NIC{{ID: "eth0", MAC: "48:da:35:6e:00:01", IPv4: "10.0.0.222"}},
	}
	h.host.drives = []proto.DriveInfo{
		{ID: "disk", Type: "disk"},
		{ID: "cdrom", Type: "cdrom", Ro: true},
	}

	h.service = New(Deps{
		PressButton: h.host.pressButton,
		PowerLED:    h.host.powerLED,
		ListDrives:  h.host.listDrives,
		InsertDrive: h.host.insertDrive,
		EjectDrive:  h.host.ejectDrive,
		ImageDir:    h.imageDir,
		Accounts:    h.counted,
		APIKeyUser: func(secret string) (string, bool) {
			username, ok := h.keys[secret]
			return username, ok
		},
		Limiter:         h.limiter,
		AuthDisabled:    func() bool { return h.authDisabled },
		FirmwareVersion: func() string { return "2.3.0 (image v1.4.3)" },
		NICs:            func() []NIC { return h.nics },
		UUID:            testUUID,
		Now:             func() time.Time { return h.clock },
	})

	h.engine = gin.New()
	h.service.Register(h.engine)

	return h
}

func basic(username, password string) string {
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
}

// do sends one request. headers are name, value pairs. A body is sent as
// application/json unless the headers set a Content-Type.
func (h *harness) do(method, path, body string, headers ...string) *httptest.ResponseRecorder {
	h.t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	request.RemoteAddr = clientIP + ":40000"
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	for i := 0; i+1 < len(headers); i += 2 {
		request.Header.Set(headers[i], headers[i+1])
	}

	w := httptest.NewRecorder()
	h.engine.ServeHTTP(w, request)
	return w
}

// token opens a session for an account directly in the store, without the
// password check, and returns its X-Auth-Token header pair.
func (h *harness) token(username string) []string {
	h.t.Helper()

	user, err := h.accounts.Get(username)
	if err != nil {
		h.t.Fatalf("get %s: %s", username, err)
	}
	_, token, err := h.service.sessions.create(user.Username, user.TokenVersion)
	if err != nil {
		h.t.Fatalf("create a session: %s", err)
	}
	return []string{"X-Auth-Token", token}
}

func (h *harness) admin() []string {
	return h.token("admin")
}

func (h *harness) user() []string {
	return h.token("alice")
}

func decode(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("response is not a JSON object: %s\n%s", err, w.Body.String())
	}
	return body
}

// messageID returns the MessageId of a Redfish error body.
func messageID(t *testing.T, w *httptest.ResponseRecorder) string {
	t.Helper()

	body := decode(t, w)
	e, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("no error object in %s", w.Body.String())
	}
	info, ok := e["@Message.ExtendedInfo"].([]any)
	if !ok || len(info) != 1 {
		t.Fatalf("no ExtendedInfo in %s", w.Body.String())
	}
	id, _ := info[0].(map[string]any)["MessageId"].(string)
	return id
}

func expectError(t *testing.T, w *httptest.ResponseRecorder, status int, id string) {
	t.Helper()

	if w.Code != status {
		t.Fatalf("status %d, want %d: %s", w.Code, status, w.Body.String())
	}
	if got := messageID(t, w); got != registry+"."+id {
		t.Fatalf("MessageId %q, want %q", got, registry+"."+id)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
