package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
)

// passwordRouter serves login and a self password change, the way router/auth.go
// wires them.
func passwordRouter() *gin.Engine {
	service := NewService()
	router := gin.New()
	router.POST("/login", service.Login)
	router.Group("/").Use(middleware.CheckToken()).POST("/password", service.ChangePassword)
	return router
}

// stubSystemPassword keeps the test from shelling out to passwd(1).
func stubSystemPassword(t *testing.T, err error) {
	t.Helper()

	original := systemPasswordUpdater
	t.Cleanup(func() { systemPasswordUpdater = original })
	systemPasswordUpdater = func(string) error { return err }
}

// stubSaveIdentity records whether the identity write-back ran, and lets a test
// make it fail.
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

// The device owner's password lives in /etc/shadow, which belongs to the slot
// that is running. Without the write-back a password change is undone by the
// next reboot and lost outright by a slot switch.
func TestChangePasswordSavesIdentity(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()

	stubSystemPassword(t, nil)
	calls := stubSaveIdentity(t, nil)

	router := passwordRouter()
	cookie := loginCookie(t, router, "admin", "admin")

	if code := changeOwnPassword(t, router, cookie); code != 0 {
		t.Fatalf("password change returned code %d, want 0", code)
	}

	if *calls != 1 {
		t.Fatalf("the identity write-back ran %d times, want 1", *calls)
	}
}

// The password is already written and the old sessions are already revoked when
// the write-back runs, so a failure here cannot be undone by refusing. An
// identity copy that is one boot stale is the better of the two outcomes.
func TestChangePasswordSucceedsWhenIdentitySaveFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()

	stubSystemPassword(t, nil)
	stubSaveIdentity(t, errors.New("no /data"))

	router := passwordRouter()
	cookie := loginCookie(t, router, "admin", "admin")

	if code := changeOwnPassword(t, router, cookie); code != 0 {
		t.Fatalf("a failed identity save must not fail the password change, got code %d", code)
	}

	if _, ok, err := store.Authenticate("admin", "new-password"); err != nil || !ok {
		t.Fatalf("the password change did not take: ok=%v err=%v", ok, err)
	}
}

// Nothing was changed, so there is nothing to persist. Saving here would copy a
// shadow the caller never altered.
func TestChangePasswordSkipsIdentitySaveWhenTheSystemPasswordFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()

	stubSystemPassword(t, errors.New("passwd failed"))
	calls := stubSaveIdentity(t, nil)

	router := passwordRouter()
	cookie := loginCookie(t, router, "admin", "admin")

	if code := changeOwnPassword(t, router, cookie); code == 0 {
		t.Fatal("the password change reported success although the system password failed")
	}

	if *calls != 0 {
		t.Fatalf("the identity write-back ran %d times after a failed change, want 0", *calls)
	}
}

// changeOwnPassword posts a self password change and returns the envelope code.
func changeOwnPassword(t *testing.T, router http.Handler, cookie *http.Cookie) int {
	t.Helper()

	recorder := requestJSONRecorder(router, http.MethodPost, "/password", map[string]any{
		"currentPassword": encryptForRequest("admin"),
		"password":        encryptForRequest("new-password"),
	}, cookie)
	if recorder.Code != http.StatusOK {
		t.Fatalf("transport status = %d", recorder.Code)
	}

	var response proto.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %s", err)
	}

	return response.Code
}

// stubStockPassword stands in for the kernel command line and for the version
// tool. boot is the version id the command line names, "" for a boot that is
// not a version boot. Each run of the tool appends "stock" to order, and each
// identity write-back appends "identity", so a test can see which came first.
func stubStockPassword(t *testing.T, boot string, err error) *[]string {
	t.Helper()

	originalBoot, originalCopy, originalSave := versionBoot, copyRootPasswordToStock, saveIdentity
	t.Cleanup(func() {
		versionBoot, copyRootPasswordToStock, saveIdentity = originalBoot, originalCopy, originalSave
	})

	var order []string
	versionBoot = func() string { return boot }
	copyRootPasswordToStock = func() error {
		order = append(order, "stock")
		return err
	}
	saveIdentity = func() error {
		order = append(order, "identity")
		return nil
	}

	return &order
}

// On a version boot the owner's new root password also goes to stock's root on
// p2, after the identity write-back, so stock's ssh takes the same password.
func TestChangePasswordCopiesRootPasswordToStockOnAVersionBoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()

	stubSystemPassword(t, nil)
	order := stubStockPassword(t, "1.2.1-mainline", nil)

	router := passwordRouter()
	cookie := loginCookie(t, router, "admin", "admin")

	if code := changeOwnPassword(t, router, cookie); code != 0 {
		t.Fatalf("password change returned code %d, want 0", code)
	}

	if got := strings.Join(*order, " "); got != "identity stock" {
		t.Fatalf("ran %q, want the identity write-back and then the stock copy", got)
	}
}

// A slot boot or an upstream image has no stock root beside it.
func TestChangePasswordSkipsTheStockCopyOffAVersionBoot(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()

	stubSystemPassword(t, nil)
	order := stubStockPassword(t, "", nil)

	router := passwordRouter()
	cookie := loginCookie(t, router, "admin", "admin")

	if code := changeOwnPassword(t, router, cookie); code != 0 {
		t.Fatalf("password change returned code %d, want 0", code)
	}

	if got := strings.Join(*order, " "); got != "identity" {
		t.Fatalf("ran %q, want only the identity write-back", got)
	}
}

// The password is already set when the copy runs, so a failure cannot be undone
// by refusing; it is logged, and the next boot copies the hash again.
func TestChangePasswordSucceedsWhenTheStockCopyFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()

	stubSystemPassword(t, nil)
	order := stubStockPassword(t, "1.2.1-mainline", errors.New("store busy"))

	router := passwordRouter()
	cookie := loginCookie(t, router, "admin", "admin")

	if code := changeOwnPassword(t, router, cookie); code != 0 {
		t.Fatalf("a failed stock copy must not fail the password change, got code %d", code)
	}

	if got := strings.Join(*order, " "); got != "identity stock" {
		t.Fatalf("ran %q, want the identity write-back and then the stock copy", got)
	}
	if _, ok, err := store.Authenticate("admin", "new-password"); err != nil || !ok {
		t.Fatalf("the password change did not take: ok=%v err=%v", ok, err)
	}
}

// Only the device owner's change sets root's password, so another user's
// change has nothing to copy to stock.
func TestChangePasswordSkipsTheStockCopyForAnotherUser(t *testing.T) {
	gin.SetMode(gin.TestMode)
	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	restore := useTestStore(store)
	defer restore()
	if _, ok, err := store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err := store.Create("alice", "alice-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}

	stubSystemPassword(t, nil)
	order := stubStockPassword(t, "1.2.1-mainline", nil)

	router := passwordRouter()
	cookie := loginCookie(t, router, "alice", "alice-password")

	recorder := requestJSONRecorder(router, http.MethodPost, "/password", map[string]any{
		"currentPassword": encryptForRequest("alice-password"),
		"password":        encryptForRequest("new-password"),
	}, cookie)
	var response proto.Response
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode response: %s", err)
	}
	if response.Code != 0 {
		t.Fatalf("password change returned code %d, want 0", response.Code)
	}

	if got := strings.Join(*order, " "); got != "identity" {
		t.Fatalf("ran %q, want only the identity write-back", got)
	}
}
