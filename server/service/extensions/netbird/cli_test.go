//go:build linux

package netbird

import (
	"strings"
	"testing"

	"NanoKVM-Server/service/extensions/vpn"
)

// A CLI that echoes the key in its error must not put it on the page.
func TestSetupKeyNeverReachesTheMessage(t *testing.T) {
	scratchImage(t, true)
	stub(t, NetbirdPath, `echo "login failed: invalid setup key $3" >&2; exit 1`)

	err := NewCli().JoinWithSetupKey("SECRET-KEY-123")
	msg := vpn.Message("join failed", err)
	if err == nil || strings.Contains(msg, "SECRET-KEY-123") || !strings.Contains(msg, "invalid setup key ***") {
		t.Fatalf("message is %q", msg)
	}
	if strings.Contains(err.Error(), "SECRET-KEY-123") {
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
