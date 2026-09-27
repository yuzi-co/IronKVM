//go:build linux

package netbird

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

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
