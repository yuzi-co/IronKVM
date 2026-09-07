package middleware

import (
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"

	"github.com/gin-gonic/gin"
)

// watchNoteSignedInAdmin replaces the seam with a recorder. It exists because
// the network package's own tests call the observer directly: without a test
// here, deleting the call in authenticateBySession leaves the whole tree green
// and the address trial confirming nothing.
func watchNoteSignedInAdmin(t *testing.T) *[]string {
	t.Helper()

	original := noteSignedInAdmin
	var seen []string
	noteSignedInAdmin = func(c *gin.Context) {
		seen = append(seen, c.Request.URL.Path)
	}

	t.Cleanup(func() {
		noteSignedInAdmin = original
	})

	return &seen
}

func TestOnlyAnAdministratorSessionIsReportedAsReachingTheBoard(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	admin, ok, err := store.Authenticate("admin", "admin")
	if err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	if err = store.Create("alice", "valid-password", authn.RoleUser); err != nil {
		t.Fatal(err)
	}
	alice, err := store.Get("alice")
	if err != nil {
		t.Fatal(err)
	}

	restore := useTestAuthStore(t, store)
	defer restore()

	adminToken, err := GenerateJWT(admin.Username, admin.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}
	userToken, err := GenerateJWT(alice.Username, alice.TokenVersion)
	if err != nil {
		t.Fatal(err)
	}

	router := gin.New()
	router.GET("/api/probe", CheckToken(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name  string
		token string
		want  int
	}{
		{name: "an administrator session", token: adminToken, want: 1},
		{name: "an operator session", token: userToken, want: 0},
		{name: "no credentials at all", token: "", want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			seen := watchNoteSignedInAdmin(t)

			requestWithToken(router, "/api/probe", test.token)

			if len(*seen) != test.want {
				t.Errorf("reported %d times, want %d", len(*seen), test.want)
			}
		})
	}
}

// An api key belongs to a script. A script that can reach the board says
// nothing about whether the operator can, so it must not keep an address
// change by polling.
func TestAnAPIKeyIsNotReportedAsReachingTheBoard(t *testing.T) {
	withAccounts(t)

	secret := issueKey(t, "admin")
	seen := watchNoteSignedInAdmin(t)

	router := gin.New()
	router.GET("/api/probe", CheckToken(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/probe", nil)
	request.Header.Set(apiKeyHeader, secret)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("the key did not authenticate: status = %d", recorder.Code)
	}
	if len(*seen) != 0 {
		t.Errorf("an api key was reported as a client reaching the board: %v", *seen)
	}
}

// With the check off there is no signed-in client to speak of, and a trial that
// confirmed itself on any request at all would confirm on a port scan.
func TestAuthenticationDisabledIsNotReportedAsReachingTheBoard(t *testing.T) {
	withAccounts(t)

	conf := config.GetInstance()
	original := conf.Authentication
	conf.Authentication = "disable"
	t.Cleanup(func() { conf.Authentication = original })

	seen := watchNoteSignedInAdmin(t)

	router := gin.New()
	router.GET("/api/probe", CheckToken(), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	if status := requestWithToken(router, "/api/probe", ""); status != http.StatusNoContent {
		t.Fatalf("the open mode did not admit the request: status = %d", status)
	}
	if len(*seen) != 0 {
		t.Errorf("an unauthenticated request was reported as reaching the board: %v", *seen)
	}
}
