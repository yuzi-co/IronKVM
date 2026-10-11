//go:build linux

package tailscale

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNormalizeLoginServer(t *testing.T) {
	good := map[string]string{
		"":                                    DefaultLoginServer,
		"  ":                                  DefaultLoginServer,
		"https://headscale.example.com":       "https://headscale.example.com",
		" https://headscale.example.com/ ":    "https://headscale.example.com",
		"https://hs.example.com:8443":         "https://hs.example.com:8443",
		"https://10.0.0.5":                    "https://10.0.0.5",
		"https://[fd00::1]:443/":              "https://[fd00::1]:443",
		"https://example.com/headscale/":      "https://example.com/headscale",
		"https://controlplane.tailscale.com/": DefaultLoginServer,
	}
	for in, want := range good {
		got, err := NormalizeLoginServer(in)
		if err != nil || got != want {
			t.Errorf("%q: got %q, %v; want %q", in, got, err, want)
		}
	}

	for _, in := range []string{
		"http://headscale.example.com",
		"headscale.example.com",
		"ftp://x",
		"https://",
		"https:///path",
		"https://user:pass@headscale.example.com",
		"https://headscale.example.com?x=1",
		"https://headscale.example.com#top",
		"https://head scale.example.com",
		"https://headscale.example.com:99999",
		"https://-x.example.com",
		"--auth-key=tskey-abc",
		"https://x.example.com\n--reset",
	} {
		if got, err := NormalizeLoginServer(in); err == nil {
			t.Errorf("%q was accepted as %q", in, got)
		}
	}
}

func TestLoginServerArgs(t *testing.T) {
	for _, def := range []string{"", DefaultLoginServer, "https://login.tailscale.com"} {
		if args := loginServerArgs(def); args != nil {
			t.Errorf("%q adds %v", def, args)
		}
	}
	if args := loginServerArgs("https://hs.example.com"); len(args) != 1 || args[0] != "--login-server=https://hs.example.com" {
		t.Errorf("custom server adds %v", args)
	}
}

// loginServerStub replaces the CLI stub with one that keeps its ControlURL in
// a file, as tailscaled keeps it in its preferences:
//
//	STUB_STATE       the backend state status reports (default NeedsLogin)
//	STUB_SET         1: `set --help` lists --login-server
//	STUB_LOGIN_FAIL  1: a login stores the server and then fails, printing no URL
//	STUB_SET_FAIL    1: `set --login-server` fails without changing anything
func loginServerStub(t *testing.T, controlURL string) (calls, prefsFile string) {
	t.Helper()
	calls = lifecycle(t)
	prefsFile = filepath.Join(filepath.Dir(calls), "controlurl")
	if err := os.WriteFile(prefsFile, []byte(controlURL), 0o644); err != nil {
		t.Fatal(err)
	}
	forgetLoginServer()
	t.Cleanup(forgetLoginServer)

	savedTimeout, savedLife := applyURLTimeout, rollbackLife
	t.Cleanup(func() { applyURLTimeout, rollbackLife = savedTimeout, savedLife })
	applyURLTimeout, rollbackLife = 3*time.Second, 100*time.Millisecond

	stub(t, TailscalePath, `echo "tailscale $*" >> "`+calls+`"
P="`+prefsFile+`"
server=
for a in "$@"; do case "$a" in --login-server=*) server=${a#--login-server=} ;; esac; done
case "$1" in
status) echo "{\"BackendState\": \"${STUB_STATE:-NeedsLogin}\"}" ;;
debug) printf '{\n  "ControlURL": "%s",\n  "WantRunning": false\n}\n' "$(cat "$P")" ;;
set)
	if [ "$2" = --help ]; then
		echo "USAGE"; echo "  --hostname string"
		[ "$STUB_SET" = 1 ] && echo "  --login-server string"
		exit 0
	fi
	[ "$STUB_SET_FAIL" = 1 ] && { echo "set failed: control server unreachable" >&2; exit 1; }
	[ -n "$server" ] && printf %s "$server" > "$P" ;;
login)
	[ -n "$server" ] && printf %s "$server" > "$P"
	if [ "$STUB_LOGIN_FAIL" = 1 ] && [ "$server" != "${STUB_GOOD:-}" ]; then
		echo "backend error: no such host" >&2; exit 1
	fi
	printf '\nTo authenticate, visit:\n\n\t%s/register/nodekey123\n\n' "$server" >&2 ;;
up) ;;
esac`)
	return calls, prefsFile
}

func setLoginServer(t *testing.T, body string) (code int, msg string, data map[string]any) {
	t.Helper()
	rsp := call(t, NewService().SetLoginServer, body)
	b, _ := json.Marshal(rsp.Data)
	_ = json.Unmarshal(b, &data)
	return rsp.Code, rsp.Msg, data
}

func readFile(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestSetLoginServerStartsALoginAgainstHeadscale(t *testing.T) {
	calls, prefs := loginServerStub(t, DefaultLoginServer)

	code, msg, data := setLoginServer(t, `{"loginServer":"https://hs.example.com/"}`)
	if code != 0 {
		t.Fatalf("code %d: %s", code, msg)
	}
	if data["loginServer"] != "https://hs.example.com" || data["url"] != "https://hs.example.com/register/nodekey123" {
		t.Errorf("data %v", data)
	}
	if got := readFile(t, prefs); got != "https://hs.example.com" {
		t.Errorf("ControlURL %q", got)
	}
	if !strings.Contains(callsOf(t, calls), "tailscale login --accept-dns=false --timeout=10m --login-server=https://hs.example.com") {
		t.Errorf("calls:\n%s", callsOf(t, calls))
	}
}

func TestSetLoginServerRollsBackAServerThatDoesNotAnswer(t *testing.T) {
	calls, prefs := loginServerStub(t, "https://old.example.com")
	t.Setenv("STUB_LOGIN_FAIL", "1")
	t.Setenv("STUB_GOOD", "https://old.example.com")

	code, msg, _ := setLoginServer(t, `{"loginServer":"https://new.example.com"}`)
	if code != -2 || !strings.Contains(msg, "no such host") || !strings.Contains(msg, "https://old.example.com is in use again") {
		t.Errorf("code %d: %s", code, msg)
	}
	if got := readFile(t, prefs); got != "https://old.example.com" {
		t.Errorf("ControlURL %q after the rollback", got)
	}
	if !strings.Contains(callsOf(t, calls), "--login-server=https://old.example.com") {
		t.Errorf("no rollback login:\n%s", callsOf(t, calls))
	}
}

func TestSetLoginServerReportsAFailedRollback(t *testing.T) {
	_, prefs := loginServerStub(t, "https://old.example.com")
	t.Setenv("STUB_LOGIN_FAIL", "1") // every login fails, the old server's too

	code, msg, _ := setLoginServer(t, `{"loginServer":"https://new.example.com"}`)
	if code != -2 || !strings.Contains(msg, "putting https://old.example.com back also failed") {
		t.Errorf("code %d: %s", code, msg)
	}
	_ = prefs
}

func TestSetLoginServerRefusesWhileLoggedIn(t *testing.T) {
	for _, state := range []string{"Running", "Stopped"} {
		t.Run(state, func(t *testing.T) {
			calls, prefs := loginServerStub(t, DefaultLoginServer)
			t.Setenv("STUB_STATE", state)

			code, msg, _ := setLoginServer(t, `{"loginServer":"https://hs.example.com"}`)
			if code != -3 || !strings.Contains(msg, "log out") {
				t.Errorf("code %d: %s", code, msg)
			}
			if got := readFile(t, prefs); got != DefaultLoginServer {
				t.Errorf("ControlURL changed to %q", got)
			}
			if strings.Contains(callsOf(t, calls), "tailscale login") {
				t.Errorf("a login ran:\n%s", callsOf(t, calls))
			}
		})
	}
}

func TestSetLoginServerUsesSetWhereTheCLIHasIt(t *testing.T) {
	calls, prefs := loginServerStub(t, DefaultLoginServer)
	t.Setenv("STUB_SET", "1")
	t.Setenv("STUB_STATE", "Running")

	code, msg, data := setLoginServer(t, `{"loginServer":"https://hs.example.com"}`)
	if code != 0 || data["url"] != "" {
		t.Fatalf("code %d: %s, %v", code, msg, data)
	}
	if got := readFile(t, prefs); got != "https://hs.example.com" {
		t.Errorf("ControlURL %q", got)
	}
	if !strings.Contains(callsOf(t, calls), "tailscale set --login-server=https://hs.example.com") {
		t.Errorf("calls:\n%s", callsOf(t, calls))
	}
}

func TestSetLoginServerRollsBackAFailedSet(t *testing.T) {
	_, prefs := loginServerStub(t, "https://old.example.com")
	t.Setenv("STUB_SET", "1")
	t.Setenv("STUB_SET_FAIL", "1")

	code, msg, _ := setLoginServer(t, `{"loginServer":"https://new.example.com"}`)
	if code == 0 || !strings.Contains(msg, "control server unreachable") {
		t.Errorf("code %d: %s", code, msg)
	}
	if got := readFile(t, prefs); got != "https://old.example.com" {
		t.Errorf("ControlURL %q", got)
	}
}

func TestSetLoginServerRejectsInvalidInput(t *testing.T) {
	calls, _ := loginServerStub(t, DefaultLoginServer)

	for _, body := range []string{`{"loginServer":"http://hs.example.com"}`, `{"loginServer":"--auth-key=tskey-x"}`, `not json`} {
		if code, _, _ := setLoginServer(t, body); code != -1 {
			t.Errorf("%s: code %d", body, code)
		}
	}
	if c := callsOf(t, calls); c != "" {
		t.Errorf("the CLI ran for invalid input:\n%s", c)
	}
}

func TestSetLoginServerToTheSameServerChangesNothing(t *testing.T) {
	calls, _ := loginServerStub(t, "")
	t.Setenv("STUB_STATE", "Running")

	if code, msg, _ := setLoginServer(t, `{"loginServer":""}`); code != 0 {
		t.Fatalf("code %d: %s", code, msg)
	}
	if strings.Contains(callsOf(t, calls), "login") || strings.Contains(callsOf(t, calls), "set --login") {
		t.Errorf("calls:\n%s", callsOf(t, calls))
	}
}

// A node on Headscale has to keep naming it, or tailscale refuses up.
func TestUpAndLoginKeepACustomServer(t *testing.T) {
	calls, _ := loginServerStub(t, "https://hs.example.com")

	if err := NewCli().Up(); err != nil {
		t.Fatal(err)
	}
	if _, err := NewCli().Login(); err != nil {
		t.Fatal(err)
	}
	c := callsOf(t, calls)
	if !strings.Contains(c, "tailscale up --accept-dns=false --login-server=https://hs.example.com") ||
		!strings.Contains(c, "tailscale login --accept-dns=false --timeout=10m --login-server=https://hs.example.com") {
		t.Errorf("calls:\n%s", c)
	}
}

func TestUpWithTailscalesServerIsUnchanged(t *testing.T) {
	calls, _ := loginServerStub(t, DefaultLoginServer)

	if err := NewCli().Up(); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(callsOf(t, calls), "tailscale up --accept-dns=false\n") {
		t.Errorf("calls:\n%s", callsOf(t, calls))
	}
}

func TestStatusShowsACustomServer(t *testing.T) {
	_, _ = loginServerStub(t, "https://hs.example.com")
	t.Setenv("STUB_STATE", "Running")

	rsp := call(t, NewService().GetStatus, "")
	b, _ := json.Marshal(rsp.Data)
	var st struct {
		LoginServer string `json:"loginServer"`
	}
	_ = json.Unmarshal(b, &st)
	if st.LoginServer != "https://hs.example.com" {
		t.Errorf("status %s", b)
	}
}
