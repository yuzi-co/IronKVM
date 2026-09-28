package ipmi

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/authn"
)

// fakeHost stands in for the buttons and the power LED. No test in this
// package reaches a real GPIO line.
type fakeHost struct {
	mu sync.Mutex

	// led is the power LED; nil means the board has none wired.
	led      *bool
	ledErr   error
	pressErr error
	presses  []string
}

func (f *fakeHost) pressButton(kind string, d time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.pressErr != nil {
		return f.pressErr
	}
	f.presses = append(f.presses, fmt.Sprintf("%s %s", kind, d))
	return nil
}

func (f *fakeHost) powerLED() (bool, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.ledErr != nil {
		return false, f.ledErr
	}
	if f.led == nil {
		return false, nil
	}
	return *f.led, nil
}

func (f *fakeHost) ledConnected() bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.led != nil || f.ledErr != nil
}

func (f *fakeHost) setLED(on *bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.led = on
}

func (f *fakeHost) pressed() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.presses...)
}

// waitForPresses waits for the presses a Chassis Control started, and for
// the lock after them to be free again.
func (f *fakeHost) waitForPresses(t *testing.T, s *Service, want int) []string {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if len(f.pressed()) >= want && s.powerMu.TryLock() {
			s.powerMu.Unlock()
			return f.pressed()
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("waited for %d presses, got %v", want, f.pressed())
	return nil
}

func on() *bool  { v := true; return &v }
func off() *bool { v := false; return &v }

// fakeLimiter records failures and can lock everyone out.
type fakeLimiter struct {
	mu        sync.Mutex
	locked    bool
	failed    []string
	succeeded []string
}

func (l *fakeLimiter) Locked(ip, username string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.locked
}

func (l *fakeLimiter) Failed(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failed = append(l.failed, ip+"/"+username)
}

func (l *fakeLimiter) Succeeded(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.succeeded = append(l.succeeded, ip+"/"+username)
}

// The accounts every test service has. "operator" is a user account, which
// IPMI limits to USER.
const (
	adminName     = "admin"
	adminWeb      = "web-password-1"
	adminIPMI     = "ipmi-admin-pass"
	userName      = "operator"
	userWeb       = "web-password-2"
	userIPMI      = "ipmi-user-passw"
	noIPMIName    = "noipmi"
	disabledName  = "disabled"
	disabledIPMI  = "ipmi-disabled-1"
	testGUIDValue = "0123456789abcdef"
)

type testEnv struct {
	service  *Service
	host     *fakeHost
	limiter  *fakeLimiter
	accounts *authn.Store
	keyring  *Keyring
	enabled  bool
}

func init() {
	// A Chassis Control keeps its lock this long after its presses; the
	// tests have no host to wait for.
	pressSettle = 0
	cyclePause = 0
}

func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	dir := t.TempDir()
	env := &testEnv{
		host:     &fakeHost{led: on()},
		limiter:  &fakeLimiter{},
		accounts: authn.NewStore(filepath.Join(dir, "pwd")),
		keyring:  NewKeyring(filepath.Join(dir, "ipmi.key")),
		enabled:  true,
	}

	for _, a := range []struct {
		name, web string
		role      authn.Role
	}{
		{adminName, adminWeb, authn.RoleAdmin},
		{userName, userWeb, authn.RoleUser},
		{noIPMIName, "web-password-3", authn.RoleAdmin},
		{disabledName, "web-password-4", authn.RoleUser},
	} {
		// The default account is admin/admin; replace its password.
		if a.name == adminName {
			if _, err := env.accounts.SetPassword(adminName, adminWeb); err != nil {
				t.Fatal(err)
			}
			continue
		}
		if err := env.accounts.Create(a.name, a.web, a.role); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := env.accounts.Update(adminName, disabledName, authn.UserPatch{Enabled: off()}); err != nil {
		t.Fatal(err)
	}

	var guid [16]byte
	copy(guid[:], testGUIDValue)
	env.service = New(Deps{
		PressButton:       env.host.pressButton,
		PowerLED:          env.host.powerLED,
		PowerLEDConnected: env.host.ledConnected,
		Accounts:          env.accounts,
		Keyring:           env.keyring,
		Limiter:           env.limiter,
		FirmwareVersion:   func() string { return "2.4.1 (image v1.4.0)" },
		GUID:              guid,
		Addr:              "127.0.0.1:0",
		Enabled:           func() bool { return env.enabled },
		SetEnabled: func(v bool) error {
			env.enabled = v
			return nil
		},
	})

	for name, password := range map[string]string{adminName: adminIPMI, userName: userIPMI, disabledName: disabledIPMI} {
		env.setIPMIPassword(t, name, password)
	}

	env.service.Start()
	t.Cleanup(env.service.Stop)
	if env.service.LocalAddr() == nil {
		t.Fatal("the service did not start")
	}
	return env
}

func (env *testEnv) setIPMIPassword(t *testing.T, name, password string) {
	t.Helper()
	sealed, err := env.keyring.Seal(name, password)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := env.accounts.SetIPMIPassword(name, sealed); err != nil {
		t.Fatal(err)
	}
}
