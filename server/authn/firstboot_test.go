package authn

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testFirstBootPassword = "Kq7mWx3Rtb9P"

func writeFirstBootFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "first-boot-password")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestFirstBootPasswordReplacesTheDefaultOnAFreshBoard(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	file := writeFirstBootFile(t, testFirstBootPassword+"\n")

	provisioned, err := store.ProvisionFirstBoot(file)
	if err != nil || !provisioned {
		t.Fatalf("provision: %v, %v", provisioned, err)
	}

	if _, ok, _ := store.Authenticate("admin", "admin"); ok {
		t.Fatal("admin/admin still works on a provisioned board")
	}
	user, ok, err := store.Authenticate("admin", testFirstBootPassword)
	if err != nil || !ok {
		t.Fatalf("the generated password does not log in: ok=%v err=%v", ok, err)
	}
	if !user.MustChangePassword || !user.SystemAccount || user.Role != RoleAdmin {
		t.Fatalf("unexpected owner account: %+v", user)
	}
	if !store.FirstBootPending() {
		t.Fatal("the first-boot password is live, so it must be pending")
	}

	data, err := os.ReadFile(store.path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), testFirstBootPassword) {
		t.Fatal("the account file holds the password in the clear")
	}
}

func TestFirstBootPasswordNeverReplacesAnAccount(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if err := os.WriteFile(store.path, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	before, _ := os.ReadFile(store.path)

	provisioned, err := store.ProvisionFirstBoot(writeFirstBootFile(t, testFirstBootPassword))
	if err != nil || provisioned {
		t.Fatalf("provisioned over an existing account file: %v, %v", provisioned, err)
	}
	after, _ := os.ReadFile(store.path)
	if string(before) != string(after) {
		t.Fatal("the account file changed")
	}

	// An existing board whose owner changed the password.
	changed := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, _, err = changed.Authenticate("admin", "admin"); err != nil {
		t.Fatal(err)
	}
	if _, err = changed.SetPassword("admin", "owner-chose-this"); err != nil {
		t.Fatal(err)
	}
	if provisioned, err = changed.ProvisionFirstBoot(writeFirstBootFile(t, testFirstBootPassword)); err != nil || provisioned {
		t.Fatalf("provisioned over a changed password: %v, %v", provisioned, err)
	}
	if _, ok, _ := changed.Authenticate("admin", "owner-chose-this"); !ok {
		t.Fatal("the owner's password stopped working")
	}
}

func TestFirstBootWithoutAPasswordFileKeepsTheDefault(t *testing.T) {
	// An unmodified vendor image updated to this server has no account file
	// and uses admin/admin. Its owner must not be locked out by the update.
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	provisioned, err := store.ProvisionFirstBoot(filepath.Join(t.TempDir(), "absent"))
	if err != nil || provisioned {
		t.Fatalf("provision with no file: %v, %v", provisioned, err)
	}
	if _, ok, _ := store.Authenticate("admin", "admin"); !ok {
		t.Fatal("admin/admin stopped working")
	}
}

func TestFirstBootRejectsAWeakPasswordFile(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	provisioned, err := store.ProvisionFirstBoot(writeFirstBootFile(t, "short"))
	if err == nil || provisioned {
		t.Fatalf("a 5 byte password was accepted: %v, %v", provisioned, err)
	}
	if strings.Contains(err.Error(), "short") {
		t.Fatalf("the error repeats the password: %v", err)
	}
	if _, statErr := os.Stat(store.path); statErr == nil {
		t.Fatal("an account file was written")
	}
}

func TestFirstBootIsNotPendingOnceThePasswordIsChosen(t *testing.T) {
	store := NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, err := store.ProvisionFirstBoot(writeFirstBootFile(t, testFirstBootPassword)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.SetPassword("admin", "owner-chose-this"); err != nil {
		t.Fatal(err)
	}
	if store.FirstBootPending() {
		t.Fatal("still pending after the owner chose a password")
	}
}

func TestClearFirstBootRemovesTheFilesAndToleratesMissingOnes(t *testing.T) {
	file := writeFirstBootFile(t, testFirstBootPassword)
	missing := filepath.Join(t.TempDir(), "absent")
	if err := ClearFirstBoot(file, missing); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(file); err == nil {
		t.Fatal("the file is still there")
	}
}

func TestViewerIsStoredAsAUserThatAnOlderServerCannotLogInAs(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pwd")
	store := NewStore(path)
	if err := store.Create("guest", "guest-password", RoleViewer); err != nil {
		t.Fatal(err)
	}

	// This server reads it back as a viewer.
	reread := NewStore(path)
	user, ok, err := reread.Authenticate("guest", "guest-password")
	if err != nil || !ok || user.Role != RoleViewer {
		t.Fatalf("viewer after reload: %+v ok=%v err=%v", user, ok, err)
	}

	// A server that knows only admin and user rejects the whole file over an
	// unknown role, and slot A may run one against this file. So the file must
	// read as a valid one there, with the viewer unable to log in.
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var old struct {
		Users []struct {
			Username string `json:"username"`
			Password string `json:"password"`
			Role     string `json:"role"`
		} `json:"users"`
	}
	if err = json.Unmarshal(data, &old); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, entry := range old.Users {
		if entry.Username != "guest" {
			continue
		}
		found = true
		if entry.Role != string(RoleUser) {
			t.Fatalf("an older server would see role %q", entry.Role)
		}
		if !strings.HasPrefix(entry.Password, viewerHashPrefix) {
			t.Fatal("the hash an older server reads is a real bcrypt hash")
		}
		if passwordMatches(entry.Password, "guest-password") {
			t.Fatal("an older server could authenticate the viewer")
		}
	}
	if !found {
		t.Fatal("the viewer was not saved")
	}
}
