package netboot

import (
	"net/http"
	"net/netip"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/utils"
)

// linkServer answers the host on the USB link: the menu, the netboot.xyz
// binaries the menu chains, and the images in the image directory.
//
// It listens on the board's address on the link and nowhere else. That alone
// is not enough, because Linux accepts a packet for any of its addresses on
// any interface, so a machine on the LAN could send to the link's address
// through the board's LAN port. Every request from outside the link's subnet
// is refused as well. The answer to such a request would go out on the link,
// where the sender is not, so it could not even finish a TCP handshake with a
// forged address.
type linkServer struct {
	subnet   netip.Prefix
	base     string
	imageDir string
	tftpRoot string
	images   func() []string
	boots    *bootLog
}

func (s *linkServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	client, ok := clientAddr(r)
	if !ok || !s.subnet.Contains(client) {
		http.Error(w, "network boot answers the host on the USB link only", http.StatusForbidden)
		return
	}

	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	switch p := r.URL.Path; {
	case p == "/menu.ipxe":
		s.boots.add(client, "menu")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Cache-Control", "no-store")
		_, _ = w.Write([]byte(menu(s.base, s.images())))

	case strings.HasPrefix(p, "/boot/"):
		name := strings.TrimPrefix(p, "/boot/")
		if !slices.Contains(chainedFiles(), name) {
			http.NotFound(w, r)
			return
		}
		s.boots.add(client, name)
		s.serveFile(w, r, filepath.Join(s.tftpRoot, name), s.tftpRoot)

	case strings.HasPrefix(p, "/iso/"):
		rel := strings.TrimPrefix(p, "/iso/")
		if !slices.Contains(s.images(), rel) {
			http.NotFound(w, r)
			return
		}
		// sanboot reads the image in many ranges. The first read is the
		// boot; the rest would push everything else out of the log.
		if rg := r.Header.Get("Range"); rg == "" || strings.HasPrefix(rg, "bytes=0-") {
			s.boots.add(client, rel)
		}
		s.serveFile(w, r, filepath.Join(s.imageDir, filepath.FromSlash(rel)), s.imageDir)

	default:
		http.NotFound(w, r)
	}
}

// serveFile serves a regular file inside root with range requests, which
// iPXE's sanboot needs. A link that leads out of root is refused, as the
// virtual drives refuse one.
func (s *linkServer) serveFile(w http.ResponseWriter, r *http.Request, name, root string) {
	realRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	resolved, err := filepath.EvalSymlinks(name)
	if err != nil || !utils.IsPathInside(realRoot, resolved) {
		http.NotFound(w, r)
		return
	}

	f, err := os.Open(resolved)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer func() { _ = f.Close() }()

	fi, err := f.Stat()
	if err != nil || !fi.Mode().IsRegular() {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeContent(w, r, path.Base(name), fi.ModTime(), f)
}

// chainedFiles are the files under /boot/: the netboot.xyz binaries, and
// nothing else from the TFTP root.
func chainedFiles() []string {
	return []string{netbootXYZBIOS, netbootXYZEFIx64, netbootXYZEFIarm64}
}

func clientAddr(r *http.Request) (netip.Addr, bool) {
	ap, err := netip.ParseAddrPort(r.RemoteAddr)
	if err != nil {
		return netip.Addr{}, false
	}
	return ap.Addr().Unmap(), true
}

// Boot is one thing the host fetched from the board: the menu, a netboot.xyz
// binary, or the start of an image.
type Boot struct {
	Time   time.Time `json:"time"`
	Client string    `json:"client"`
	What   string    `json:"what"`
}

// bootLog keeps the last few boots for the page, newest first.
type bootLog struct {
	mu    sync.Mutex
	boots []Boot
}

const bootLogSize = 10

func (l *bootLog) add(client netip.Addr, what string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.boots = append([]Boot{{Time: time.Now(), Client: client.String(), What: what}}, l.boots...)
	if len(l.boots) > bootLogSize {
		l.boots = l.boots[:bootLogSize]
	}
	log.Debugf("netboot: %s fetched %s", client, what)
}

func (l *bootLog) list() []Boot {
	l.mu.Lock()
	defer l.mu.Unlock()

	return append([]Boot(nil), l.boots...)
}
