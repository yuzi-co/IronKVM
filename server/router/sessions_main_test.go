package router

import (
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/authn"
)

// TestMain keeps the tests that sign in away from the board's session file.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "router-sessions-")
	if err != nil {
		panic(err)
	}
	authn.DefaultSessions = authn.NewSessionStore(filepath.Join(dir, "login_sessions.json"))
	code := m.Run()
	_ = os.RemoveAll(dir)
	os.Exit(code)
}
