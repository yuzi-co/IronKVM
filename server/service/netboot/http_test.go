package netboot

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type linkFixture struct {
	server   *linkServer
	imageDir string
	tftpRoot string
	image    []byte
}

func newLinkFixture(t *testing.T) *linkFixture {
	t.Helper()

	base := t.TempDir()
	f := &linkFixture{
		imageDir: filepath.Join(base, "data"),
		tftpRoot: filepath.Join(base, "tftp"),
		image:    []byte(strings.Repeat("0123456789abcdef", 4096)),
	}
	for _, dir := range []string{f.imageDir, f.tftpRoot, filepath.Join(base, "outside")} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	write := func(path string, content []byte) {
		if err := os.WriteFile(path, content, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write(filepath.Join(f.imageDir, "debian.iso"), f.image)
	write(filepath.Join(base, "outside", "secret.iso"), []byte("secret"))
	write(filepath.Join(f.tftpRoot, netbootXYZEFIx64), []byte("netboot.xyz efi"))
	write(filepath.Join(f.tftpRoot, ipxeEFIx64), []byte("ipxe efi"))
	// A link that looks like an image and leads out of the directory.
	if err := os.Symlink(filepath.Join(base, "outside", "secret.iso"), filepath.Join(f.imageDir, "escape.iso")); err != nil {
		t.Fatal(err)
	}

	f.server = &linkServer{
		subnet:   netip.MustParsePrefix("172.31.255.0/30"),
		base:     "http://172.31.255.1:8069",
		imageDir: f.imageDir,
		tftpRoot: f.tftpRoot,
		images:   func() []string { return listImages(f.imageDir) },
		boots:    &bootLog{},
	}
	return f
}

func (f *linkFixture) do(method, path, client string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, "http://172.31.255.1:8069"+path, nil)
	req.RemoteAddr = client
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rec := httptest.NewRecorder()
	f.server.ServeHTTP(rec, req)
	return rec
}

const hostOnLink = "172.31.255.2:40000"

// sanboot reads an image in ranges. Without them it would have to read the
// whole image for every block.
func TestImagesAreServedWithRangeRequests(t *testing.T) {
	f := newLinkFixture(t)

	rec := f.do(http.MethodGet, "/iso/debian.iso", hostOnLink, map[string]string{"Range": "bytes=16-47"})
	if rec.Code != http.StatusPartialContent {
		t.Fatalf("status %d, want 206", rec.Code)
	}
	if got := rec.Body.String(); got != string(f.image[16:48]) {
		t.Fatalf("body %q, want %q", got, f.image[16:48])
	}
	if got := rec.Header().Get("Content-Range"); got != "bytes 16-47/65536" {
		t.Fatalf("Content-Range %q", got)
	}

	rec = f.do(http.MethodGet, "/iso/debian.iso", hostOnLink, map[string]string{"Range": "bytes=65530-"})
	if rec.Code != http.StatusPartialContent || rec.Body.String() != string(f.image[65530:]) {
		t.Fatalf("an open range: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = f.do(http.MethodHead, "/iso/debian.iso", hostOnLink, nil)
	if rec.Code != http.StatusOK || rec.Header().Get("Content-Length") != "65536" ||
		rec.Header().Get("Accept-Ranges") != "bytes" {
		t.Fatalf("HEAD: status %d, headers %v", rec.Code, rec.Header())
	}

	rec = f.do(http.MethodGet, "/iso/debian.iso", hostOnLink, map[string]string{"Range": "bytes=70000-80000"})
	if rec.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Fatalf("a range past the end: status %d", rec.Code)
	}
}

// The listener is on the link's address, and Linux would still take a packet
// for it from the LAN. So the client's address is checked too.
func TestOnlyTheHostOnTheLinkIsAnswered(t *testing.T) {
	f := newLinkFixture(t)

	for _, client := range []string{"192.168.1.50:5000", "172.31.255.6:5000", "10.0.0.1:1", "[fd00::1]:80", "garbage"} {
		for _, path := range []string{"/menu.ipxe", "/iso/debian.iso", "/boot/netboot.xyz.efi"} {
			if rec := f.do(http.MethodGet, path, client, nil); rec.Code != http.StatusForbidden {
				t.Errorf("%s from %s: status %d, want 403", path, client, rec.Code)
			}
		}
	}
	if len(f.server.boots.list()) != 0 {
		t.Error("a refused request was logged as a boot")
	}
	if rec := f.do(http.MethodGet, "/menu.ipxe", "[::ffff:172.31.255.2]:40000", nil); rec.Code != http.StatusOK {
		t.Errorf("the host as a mapped address: status %d", rec.Code)
	}
}

func TestOnlyListedImagesAndTheNetbootXYZBinariesAreServed(t *testing.T) {
	f := newLinkFixture(t)

	for _, path := range []string{
		"/iso/escape.iso",            // a link out of the image directory
		"/iso/missing.iso",           // not there
		"/iso/../outside/secret.iso", // not in the listing
		"/iso/%2e%2e/outside/secret.iso",
		"/boot/" + ipxeEFIx64, // in the TFTP root, but not chained from the menu
		"/boot/../data/debian.iso",
		"/boot/boot.ipxe",
		"/",
		"/data/debian.iso",
	} {
		rec := f.do(http.MethodGet, path, hostOnLink, nil)
		if rec.Code == http.StatusOK || rec.Code == http.StatusPartialContent {
			t.Errorf("%s was served: %q", path, rec.Body.String())
		}
	}

	rec := f.do(http.MethodGet, "/boot/"+netbootXYZEFIx64, hostOnLink, nil)
	if rec.Code != http.StatusOK || rec.Body.String() != "netboot.xyz efi" {
		t.Errorf("the netboot.xyz binary: status %d, body %q", rec.Code, rec.Body.String())
	}

	if rec := f.do(http.MethodPost, "/menu.ipxe", hostOnLink, nil); rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("POST: status %d", rec.Code)
	}
}

func TestTheMenuListsTheImagesItServes(t *testing.T) {
	f := newLinkFixture(t)

	rec := f.do(http.MethodGet, "/menu.ipxe", hostOnLink, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d", rec.Code)
	}
	body, _ := io.ReadAll(rec.Body)
	menu := string(body)
	if !strings.Contains(menu, "set base http://172.31.255.1:8069\n") {
		t.Errorf("the menu does not point at the board:\n%s", menu)
	}
	if !strings.Contains(menu, "${base}/iso/debian.iso") {
		t.Errorf("the image is not in the menu:\n%s", menu)
	}
	// escape.iso is in the listing by name; it is refused when fetched,
	// which the test above covers.
	if rec.Header().Get("Cache-Control") != "no-store" {
		t.Error("the menu may be cached, and a new image would not show")
	}
}

// The page shows what the host fetched. sanboot's later ranges are not boots.
func TestBootsAreLoggedOnceAnImageIsStarted(t *testing.T) {
	f := newLinkFixture(t)

	f.do(http.MethodGet, "/menu.ipxe", hostOnLink, nil)
	f.do(http.MethodGet, "/iso/debian.iso", hostOnLink, map[string]string{"Range": "bytes=0-2047"})
	for i := 0; i < 20; i++ {
		f.do(http.MethodGet, "/iso/debian.iso", hostOnLink, map[string]string{"Range": "bytes=4096-8191"})
	}

	boots := f.server.boots.list()
	if len(boots) != 2 {
		t.Fatalf("boots %+v, want the menu and the image", boots)
	}
	if boots[0].What != "debian.iso" || boots[1].What != "menu" || boots[0].Client != "172.31.255.2" {
		t.Fatalf("boots %+v", boots)
	}

	for i := 0; i < 30; i++ {
		f.do(http.MethodGet, "/menu.ipxe", hostOnLink, nil)
	}
	if n := len(f.server.boots.list()); n != bootLogSize {
		t.Fatalf("the log holds %d, want %d", n, bootLogSize)
	}
}
