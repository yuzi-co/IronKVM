package mcpservice

import (
	"strings"
	"testing"

	"NanoKVM-Server/service/audit"
)

// cube_media inserts an image from a URL through MountURL, on a board that
// has it (ironkvm-dist#106).
func TestMediaInsertFromAURL(t *testing.T) {
	store := isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierMedia)
	host := &fakeHost{}
	devices := host.devices()
	var mounted []string
	devices.MountURL = func(drive, url string) error {
		mounted = append(mounted, drive+" "+url)
		return nil
	}
	s := newSession(t, devices)

	if out := s.call("cube_media", `{"action":"insert","drive":"cdrom","url":"https://u:secret@isos.lan/x.iso"}`); isError(out) {
		t.Fatalf("insert from a URL failed: %s", out)
	}
	if len(mounted) != 1 || mounted[0] != "cdrom https://u:secret@isos.lan/x.iso" || len(host.inserted) != 0 {
		t.Fatalf("mounted = %v, inserted = %v", mounted, host.inserted)
	}
	for _, e := range events(t, store) {
		if strings.Contains(e.Target+e.Detail, "secret") {
			t.Fatalf("the audit log has the URL's credentials: %+v", e)
		}
		if e.Action == "cube_media" && e.Target != "isos.lan" {
			t.Fatalf("audit target %q, want the URL's host", e.Target)
		}
	}

	for _, args := range []string{
		`{"action":"insert","drive":"cdrom","url":"http://isos.lan/x.iso","file":"/data/x.iso"}`,
		`{"action":"insert","drive":"disk","url":"http://isos.lan/x.img","readOnly":false}`,
	} {
		if out := s.call("cube_media", args); !isError(out) {
			t.Errorf("%s was accepted: %s", args, out)
		}
	}
	if len(mounted) != 1 {
		t.Fatalf("a refused call mounted: %v", mounted)
	}
}

func TestMediaRefusesAURLWithoutMountURL(t *testing.T) {
	isolateAudit(t)
	allowTiers(t, audit.TierObserve, audit.TierMedia)
	host := &fakeHost{}
	s := newSession(t, host.devices())

	if out := s.call("cube_media", `{"action":"insert","drive":"cdrom","url":"http://isos.lan/x.iso"}`); !isError(out) {
		t.Fatalf("a URL was accepted without MountURL: %s", out)
	}
}
