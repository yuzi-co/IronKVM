package hostdevices

import (
	"errors"
	"strings"
	"testing"
)

func TestMountErrorLosesTheURLCredentials(t *testing.T) {
	raw := "https://alice:secret@images.example/win11.iso?token=hunter2"
	cause := errors.New(`get "` + raw + `": refused; user alice:secret, query token=hunter2`)
	devices := Devices{MountURL: func(string, string) error { return cause }}

	_, err := devices.ChangeMedia(Change{Action: "insert", Drive: "cdrom", URL: raw}, "list")
	if err == nil || !errors.Is(err, cause) {
		t.Fatalf("err = %v, want one that unwraps to the mount error", err)
	}
	for _, secret := range []string{"secret", "hunter2", "alice@"} {
		if strings.Contains(err.Error(), secret) {
			t.Fatalf("err = %q carries %q", err, secret)
		}
	}
	if !strings.Contains(err.Error(), "https://images.example/win11.iso") {
		t.Fatalf("err = %q, want the URL without its credentials", err)
	}
}

func TestDescribeChangeKeepsOnlyTheHostOfAURL(t *testing.T) {
	target, detail := DescribeChange(Change{Action: "insert", Drive: "cdrom", URL: "https://alice:secret@images.example/x.iso?t=1"})
	if target != "images.example" || detail != "insert cdrom ro url" {
		t.Fatalf("target, detail = %q, %q", target, detail)
	}
}
