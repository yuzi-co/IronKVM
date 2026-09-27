//go:build linux

package netbird

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/service/extensions/vpn"
)

// A CLI that echoes the key in its error must not put it on the page.
func TestSetupKeyNeverReachesTheMessage(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `echo "login failed: invalid setup key $(cat "$3")" >&2; exit 1`)

	err := NewCli().JoinWithSetupKey(testKey)
	msg := vpn.Message("join failed", err)
	if err == nil || strings.Contains(msg, testKey) || !strings.Contains(msg, "invalid setup key ***") {
		t.Fatalf("message is %q", msg)
	}
	if strings.Contains(err.Error(), testKey) {
		t.Fatalf("the error must not carry the key either: %q", err.Error())
	}
}

func TestLoginSSOReadsTheURLFromStdout(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `[ "$1 $2" = "up --no-browser" ] || exit 9
printf 'Use this URL to log in:\n\nhttps://login.netbird.io/activate?user_code=ABCD-EFGH \n\n'`)

	url, err := NewCli().LoginSSO()
	if err != nil || url != "https://login.netbird.io/activate?user_code=ABCD-EFGH" {
		t.Fatalf("got %q %v", url, err)
	}
}

func TestVersion(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `[ "$1" = version ] && echo 0.78.2`)
	if v, err := NewCli().Version(); err != nil || v != "0.78.2" {
		t.Fatalf("got %q %v", v, err)
	}
}

const testKey = "A1B2C3D4-E5F6-4789-ABCD-0123456789EF"

// The key is masked before the tail is cut to size: a cut through the key
// would otherwise leave a piece of it that no longer matches the whole.
func TestSetupKeyIsMaskedBeforeTheTailIsCut(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `printf 'bad key %s%02040d\n' "$(cat "$3")" 0 >&2; exit 1`)

	err := NewCli().JoinWithSetupKey(testKey)
	msg := vpn.Message("join failed", err)
	if err == nil || strings.Contains(msg, testKey[len(testKey)-8:]) {
		t.Fatalf("a piece of the key reached the message: %q", msg)
	}
}

func TestSetupKeyGoesThroughAPrivateFile(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `[ "$2" = "--setup-key-file" ] || exit 9
echo "$(stat -c %a "$3") $(cat "$3")" > "$(dirname "$3")/../seen"`)
	if err := NewCli().JoinWithSetupKey(testKey); err != nil {
		t.Fatal(err)
	}
	seen, _ := os.ReadFile(filepath.Join(filepath.Dir(KeyDir), "seen"))
	if string(seen) != "600 "+testKey+"\n" {
		t.Fatalf("the CLI saw %q", seen)
	}
	if entries, _ := os.ReadDir(KeyDir); len(entries) != 0 {
		t.Fatalf("the key file must be removed, found %v", entries)
	}
}

func TestNormalizeSetupKey(t *testing.T) {
	for in, want := range map[string]string{
		testKey: testKey,
		" a1b2c3d4-e5f6-4789-abcd-0123456789ef\n": testKey,
	} {
		if got, err := NormalizeSetupKey(in); err != nil || got != want {
			t.Fatalf("%q: got %q %v", in, got, err)
		}
	}
	for _, bad := range []string{"", "KEY-123", testKey + "0", "A1B2C3D4E5F64789ABCD0123456789EF", "G1B2C3D4-E5F6-4789-ABCD-0123456789EF"} {
		_, err := NormalizeSetupKey(bad)
		if err == nil || (bad != "" && strings.Contains(err.Error(), bad)) {
			t.Fatalf("%q: got %v", bad, err)
		}
	}
}

// joinScratch points the daemon's log at a scratch file and makes the join
// time out after timeout.
func joinScratch(t *testing.T, timeout time.Duration) (logFile string) {
	t.Helper()
	scratchImage(t, true)
	logFile = filepath.Join(t.TempDir(), "netbird.log")
	savedLog, savedJoin := LogFile, joinTimeout
	t.Cleanup(func() { LogFile, joinTimeout = savedLog, savedJoin })
	LogFile, joinTimeout = logFile, timeout
	return logFile
}

// A key the management server rejects leaves `netbird up` waiting while the
// daemon retries forever. The join ends, stops the daemon's retries with
// `netbird down`, removes the key file, and says why.
func TestJoinWithARejectedKeyEndsAndSaysWhy(t *testing.T) {
	logFile := joinScratch(t, 300*time.Millisecond)
	calls := filepath.Join(t.TempDir(), "calls")
	stub(t, NetbirdPath, `echo "$1" >> "`+calls+`"
case "$1" in
up)
	echo "2026-09-28T10:00:00Z WARN couldn't add peer: setup key is invalid" >> "`+logFile+`"
	exec sleep 30 ;;
esac`)

	start := time.Now()
	err := NewCli().JoinWithSetupKey(testKey)
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("the join took %s", elapsed)
	}
	if err == nil || err.Error() != "NetBird rejected the setup key: it is invalid, expired, or already used (a one-off key works once)" {
		t.Fatalf("got %v", err)
	}
	if got := callsOf(t, calls); got != "up\ndown\n" {
		t.Fatalf("calls:\n%s", got)
	}
	if entries, _ := os.ReadDir(KeyDir); len(entries) != 0 {
		t.Fatalf("the key file must be removed, found %v", entries)
	}
}

// A timeout with nothing in the log about the key reports the CLI's output.
func TestJoinThatTimesOutWithoutAReason(t *testing.T) {
	joinScratch(t, 300*time.Millisecond)
	stub(t, NetbirdPath, `[ "$1" = up ] || exit 0
echo "connecting to $(cat "$3")"
exec sleep 30`)
	err := NewCli().JoinWithSetupKey(testKey)
	msg := vpn.Message("join failed", err)
	if err == nil || !strings.Contains(msg, "timed out") || !strings.Contains(msg, "connecting to ***") || strings.Contains(msg, testKey) {
		t.Fatalf("message is %q", msg)
	}
	if entries, _ := os.ReadDir(KeyDir); len(entries) != 0 {
		t.Fatalf("the key file must be removed, found %v", entries)
	}
}

// Only what the daemon logged during this join counts: an old rejection in
// the log is not this attempt's reason.
func TestJoinReadsOnlyThisAttemptsLog(t *testing.T) {
	logFile := joinScratch(t, 5*time.Second)
	if err := os.WriteFile(logFile, []byte("old WARN couldn't add peer: setup key is invalid\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	stub(t, NetbirdPath, `[ "$1" = up ] || exit 0
echo "management unreachable" >&2; exit 1`)
	msg := vpn.Message("join failed", NewCli().JoinWithSetupKey(testKey))
	if strings.Contains(msg, "rejected") || !strings.Contains(msg, "management unreachable") {
		t.Fatalf("message is %q", msg)
	}
}

// A rejection the CLI itself reports, with the daemon's log line, is named
// the same way.
func TestJoinFailureWithARejectionInTheLog(t *testing.T) {
	logFile := joinScratch(t, 5*time.Second)
	stub(t, NetbirdPath, `[ "$1" = up ] || exit 0
echo "couldn't add peer: setup key is invalid" >> "`+logFile+`"
echo "login failed" >&2; exit 1`)
	err := NewCli().JoinWithSetupKey(testKey)
	if err == nil || !strings.HasPrefix(err.Error(), "NetBird rejected the setup key") {
		t.Fatalf("got %v", err)
	}
}
