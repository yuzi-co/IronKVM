package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/utils"

	"github.com/gin-gonic/gin"
	"github.com/mervick/aes-everywhere/go/aes256"
	"golang.org/x/crypto/bcrypt"
)

// encryptPassword mirrors what the web UI sends: the password encrypted with
// the shared key, then URL-escaped.
func encryptPassword(password string) string {
	return url.QueryEscape(aes256.Encrypt(password, utils.SecretKey))
}

// useTempAccountFile points the account store at a throwaway file. When
// currentPassword is empty no account file is created, which is how a
// factory-fresh device looks.
func useTempAccountFile(t *testing.T, currentPassword string) string {
	t.Helper()

	original := AccountFile
	t.Cleanup(func() { AccountFile = original })

	AccountFile = filepath.Join(t.TempDir(), "pwd")

	if currentPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(currentPassword), bcrypt.DefaultCost)
		if err != nil {
			t.Fatalf("failed to hash password: %s", err)
		}
		if err := SetAccount("admin", string(hashed)); err != nil {
			t.Fatalf("failed to write account: %s", err)
		}
	}

	return AccountFile
}

// stubRootPassword prevents the test from shelling out to passwd(1).
func stubRootPassword(t *testing.T) *string {
	t.Helper()

	original := setRootPassword
	t.Cleanup(func() { setRootPassword = original })

	var applied string
	setRootPassword = func(password string) error {
		applied = password
		return nil
	}

	return &applied
}

// stubSaveIdentity records whether the identity write-back ran, and lets a
// test make it fail.
func stubSaveIdentity(t *testing.T, err error) *int {
	t.Helper()

	original := saveIdentity
	t.Cleanup(func() { saveIdentity = original })

	var calls int
	saveIdentity = func() error {
		calls++
		return err
	}

	return &calls
}

func changePassword(t *testing.T, body string) *httptest.ResponseRecorder {
	t.Helper()

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/api/auth/password", NewService().ChangePassword)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/password", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func TestChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	accountFile := useTempAccountFile(t, "correct-horse")
	stubRootPassword(t)

	before, err := os.ReadFile(accountFile)
	if err != nil {
		t.Fatalf("failed to read account: %s", err)
	}

	w := changePassword(t, `{"username":"admin","oldPassword":"`+encryptPassword("wrong-guess")+
		`","password":"`+encryptPassword("new-password")+`"}`)

	if !strings.Contains(w.Body.String(), `"code"`) || strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("expected an error response, got %s", w.Body.String())
	}

	after, err := os.ReadFile(accountFile)
	if err != nil {
		t.Fatalf("failed to re-read account: %s", err)
	}
	if string(before) != string(after) {
		t.Fatal("account must not be modified when the current password is wrong")
	}
}

func TestChangePasswordRejectsMissingCurrentPassword(t *testing.T) {
	useTempAccountFile(t, "correct-horse")
	stubRootPassword(t)

	w := changePassword(t, `{"username":"admin","password":"`+encryptPassword("new-password")+`"}`)

	if strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("expected an error response, got %s", w.Body.String())
	}
}

func TestChangePasswordAcceptsCorrectCurrentPassword(t *testing.T) {
	useTempAccountFile(t, "correct-horse")
	applied := stubRootPassword(t)

	w := changePassword(t, `{"username":"admin","oldPassword":"`+encryptPassword("correct-horse")+
		`","password":"`+encryptPassword("new-password")+`"}`)

	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("expected success, got %s", w.Body.String())
	}

	if !CompareAccount("admin", encryptPassword("new-password")) {
		t.Fatal("the new password should authenticate after the change")
	}

	if *applied != "new-password" {
		t.Fatalf("root password should be updated too, got %q", *applied)
	}
}

func TestChangePasswordAllowsFirstTimeSetupWithoutCurrentPassword(t *testing.T) {
	// A factory-fresh device has no account file and uses the documented
	// admin/admin default, so demanding the old password adds nothing and
	// would block the initial setup flow.
	useTempAccountFile(t, "")
	stubRootPassword(t)

	w := changePassword(t, `{"username":"admin","password":"`+encryptPassword("new-password")+`"}`)

	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("expected success on first-time setup, got %s", w.Body.String())
	}
}

// A changed password has to outlive the slot it was set on.
//
// /etc/shadow lives in the slot's own root filesystem, and S02identity restores
// it from /data at every boot. Without a write-back, changing the password and
// rebooting silently restores the old one, and switching slots loses it
// outright. Asking the operator to remember a command afterwards is not a
// mechanism: they will not, and nothing tells them it was needed.
func TestChangePasswordSavesIdentity(t *testing.T) {
	useTempAccountFile(t, "correct-horse")
	stubRootPassword(t)
	calls := stubSaveIdentity(t, nil)

	w := changePassword(t, `{"username":"admin","oldPassword":"`+encryptPassword("correct-horse")+
		`","password":"`+encryptPassword("new-password")+`"}`)

	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("expected success, got %s", w.Body.String())
	}
	if *calls != 1 {
		t.Fatalf("expected the identity write-back to run once, ran %d times", *calls)
	}
}

// The password is already written when the write-back runs, so a failure here
// cannot be undone by refusing. Reporting failure would also delete the account
// file on the way out, which drops the web UI back to admin/admin: a worse
// outcome than an identity copy that is one boot stale.
func TestChangePasswordSucceedsWhenIdentitySaveFails(t *testing.T) {
	useTempAccountFile(t, "correct-horse")
	applied := stubRootPassword(t)
	stubSaveIdentity(t, errors.New("no /data"))

	w := changePassword(t, `{"username":"admin","oldPassword":"`+encryptPassword("correct-horse")+
		`","password":"`+encryptPassword("new-password")+`"}`)

	if !strings.Contains(w.Body.String(), `"code":0`) {
		t.Fatalf("a failed identity save must not fail the password change, got %s", w.Body.String())
	}
	if *applied != "new-password" {
		t.Fatalf("the root password should still have been set, got %q", *applied)
	}
}

// Nothing was changed, so there is nothing to persist. Saving here would copy
// a shadow the caller never altered.
func TestChangePasswordSkipsIdentitySaveWhenRootPasswordFails(t *testing.T) {
	useTempAccountFile(t, "correct-horse")

	original := setRootPassword
	t.Cleanup(func() { setRootPassword = original })
	setRootPassword = func(string) error { return errors.New("passwd failed") }

	calls := stubSaveIdentity(t, nil)

	changePassword(t, `{"username":"admin","oldPassword":"`+encryptPassword("correct-horse")+
		`","password":"`+encryptPassword("new-password")+`"}`)

	if *calls != 0 {
		t.Fatalf("the identity write-back must not run when the password change failed, ran %d times", *calls)
	}
}
