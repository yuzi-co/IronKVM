//go:build linux

package netboot

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"NanoKVM-Server/service/upstream"
)

// fetchFake writes "<version> <name>" for every file, or fails on failOn.
func fetchFake(version, failOn string) func(ctx context.Context, name, dst string, from, span int) error {
	return func(_ context.Context, name, dst string, _, _ int) error {
		if name == failOn {
			return errors.New("checksum mismatch")
		}
		return os.WriteFile(dst, []byte(version+" "+name), 0o644)
	}
}

func bootFileContents(t *testing.T) map[string]string {
	t.Helper()
	out := map[string]string{}
	for _, name := range netbootXYZNames() {
		data, err := os.ReadFile(filepath.Join(tftpRoot(), name))
		if err != nil {
			t.Fatal(err)
		}
		out[name] = string(data)
	}
	return out
}

func TestUpdateSwapsTheBootFilesAndRecordsTheRelease(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)
	if bootFilesVersion() != netbootXYZVersion {
		t.Fatal("a fresh install is not the pinned release")
	}

	rel := upstream.Release{Version: "9.9.9"}
	sums := upstream.Checksums{}
	for _, name := range netbootXYZNames() {
		sums[name] = name + "-sum"
	}
	if err := f.svc.updateBootFiles(context.Background(), rel, sums, fetchFake("9.9.9", "")); err != nil {
		t.Fatal(err)
	}
	for name, content := range bootFileContents(t) {
		if content != "9.9.9 "+name {
			t.Fatalf("%s holds %q", name, content)
		}
	}
	rec, ok := readBootRecord()
	if !ok || rec.Version != "9.9.9" || rec.Files[netbootXYZEFIx64] != netbootXYZEFIx64+"-sum" {
		t.Fatalf("record %+v", rec)
	}
	if bootFilesVersion() != "9.9.9" {
		t.Fatal("the version does not follow the record")
	}
	// iPXE and boot.ipxe are not the update's.
	if data, _ := os.ReadFile(filepath.Join(tftpRoot(), ipxeBIOS)); string(data) != ipxeBIOS {
		t.Fatal("the update touched iPXE")
	}
	if _, err := os.Stat(filepath.Join(addonDir(), ".update")); err == nil {
		t.Fatal("the staging directory was left")
	}
}

func TestAFailedUpdateKeepsTheBootFiles(t *testing.T) {
	f := newServiceFixture(t)
	fakeInstall(t)
	before := bootFileContents(t)

	err := f.svc.updateBootFiles(context.Background(), upstream.Release{Version: "9.9.9"}, upstream.Checksums{},
		fetchFake("9.9.9", netbootXYZEFIarm64))
	if err == nil {
		t.Fatal("a failed download was not reported")
	}
	after := bootFileContents(t)
	for name := range before {
		if before[name] != after[name] {
			t.Fatalf("%s changed", name)
		}
	}
	if _, ok := readBootRecord(); ok {
		t.Fatal("a failed update left a record")
	}
}

func TestUpdateNeedsTheAddon(t *testing.T) {
	f := newServiceFixture(t)
	err := f.svc.updateBootFiles(context.Background(), upstream.Release{Version: "9.9.9"}, nil, fetchFake("9.9.9", ""))
	if !errors.Is(err, errNotInstalled) {
		t.Fatalf("update without the add-on: %v", err)
	}
	if st := f.svc.Updater(nil).Status(); st.Installed {
		t.Fatalf("status %+v", st)
	}
}
