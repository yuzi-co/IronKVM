package vnc

import (
	"bufio"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/binary"
	"errors"
	"image/jpeg"
	"io"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
)

// The tests drive the server with an RFB client written here, over a real TCP
// connection, with a fake frame source, fake accounts and a fake HID.

type fakeFrames struct {
	ch        chan []byte
	refreshes chan struct{}
	closeOnce sync.Once
	closed    chan struct{}
}

func (f *fakeFrames) Frames() <-chan []byte { return f.ch }
func (f *fakeFrames) Refresh() {
	select {
	case f.refreshes <- struct{}{}:
	default:
	}
}
func (f *fakeFrames) Close() { f.closeOnce.Do(func() { close(f.closed) }) }

type fakeInput struct {
	mu       sync.Mutex
	keyboard [][]byte
	pointer  [][]byte
	closed   bool
}

func (f *fakeInput) Keyboard(r []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.keyboard = append(f.keyboard, r)
}

func (f *fakeInput) Pointer(r []byte) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pointer = append(f.pointer, r)
}

func (f *fakeInput) Close() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.closed = true
}

func (f *fakeInput) snapshot() (keyboard, pointer [][]byte, closed bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([][]byte(nil), f.keyboard...), append([][]byte(nil), f.pointer...), f.closed
}

type fakeAccounts map[string]string

func (a fakeAccounts) Authenticate(username, password string) (*authn.User, bool, error) {
	if want, ok := a[username]; ok && want == password {
		return &authn.User{Username: username, Role: authn.RoleAdmin, Enabled: true}, true, nil
	}
	return nil, false, nil
}

type fakeLimiter struct {
	mu       sync.Mutex
	failures []string
}

func (l *fakeLimiter) Locked(string, string) bool { return false }
func (l *fakeLimiter) Failed(ip, username string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.failures = append(l.failures, ip+"/"+username)
}
func (l *fakeLimiter) Succeeded(string, string) {}
func (l *fakeLimiter) count() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.failures)
}

type memoryPassword struct {
	mu       sync.Mutex
	password string
}

func (m *memoryPassword) Get() (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.password, m.password != "", nil
}

func (m *memoryPassword) Set(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.password = p
	return nil
}

type harness struct {
	srv      *Server
	addr     string
	settings config.VNC
	limiter  *fakeLimiter
	password *memoryPassword

	mu     sync.Mutex
	frames []*fakeFrames
	inputs []*fakeInput
}

func newHarness(t *testing.T, settings config.VNC) *harness {
	t.Helper()
	settings.Enabled = true
	if settings.MaxFPS == 0 {
		settings.MaxFPS = 1000
	}

	h := &harness{settings: settings, limiter: &fakeLimiter{}, password: &memoryPassword{}}
	cert := testCertificate(t)

	h.srv = New(Deps{
		Settings:    h.getSettings,
		SetSettings: func(v config.VNC) error { h.updateSettings(func(s *config.VNC) { *s = v }); return nil },
		Subscribe: func() FrameSubscription {
			f := &fakeFrames{ch: make(chan []byte, 1), refreshes: make(chan struct{}, 1), closed: make(chan struct{})}
			h.mu.Lock()
			h.frames = append(h.frames, f)
			h.mu.Unlock()
			return f
		},
		ScreenSize: func() (int, int) { return 64, 48 },
		NewInput: func() Input {
			in := &fakeInput{}
			h.mu.Lock()
			h.inputs = append(h.inputs, in)
			h.mu.Unlock()
			return in
		},
		Accounts:    fakeAccounts{"admin": "right-password"},
		Limiter:     h.limiter,
		Password:    h.password,
		Certificate: func() (*tls.Certificate, error) { return cert, nil },
		Name:        func() string { return "test-kvm" },
		Listen: func(network, _ string) (net.Listener, error) {
			return net.Listen(network, "127.0.0.1:0")
		},
	})
	if err := h.srv.Apply(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(h.srv.Stop)
	h.addr = h.srv.Addr().String()
	return h
}

func (h *harness) getSettings() config.VNC {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.settings
}

func (h *harness) updateSettings(change func(*config.VNC)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	change(&h.settings)
}

// session returns the frame source and the input of the nth session.
func (h *harness) session(t *testing.T, n int) (*fakeFrames, *fakeInput) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		h.mu.Lock()
		if len(h.frames) > n && len(h.inputs) > n {
			f, in := h.frames[n], h.inputs[n]
			h.mu.Unlock()
			return f, in
		}
		h.mu.Unlock()
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("session %d did not start", n)
	return nil, nil
}

func testCertificate(t *testing.T) *tls.Certificate {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "test-kvm"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key}
}

// client is a minimal RFB client.
type client struct {
	t      *testing.T
	conn   net.Conn
	r      *bufio.Reader
	width  int
	height int
	name   string
}

func dial(t *testing.T, addr string) *client {
	t.Helper()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	t.Cleanup(func() { _ = conn.Close() })
	return &client{t: t, conn: conn, r: bufio.NewReader(conn)}
}

func (c *client) read(n int) []byte {
	c.t.Helper()
	buf := make([]byte, n)
	if _, err := io.ReadFull(c.r, buf); err != nil {
		c.t.Fatalf("read %d bytes: %s", n, err)
	}
	return buf
}

func (c *client) write(b []byte) {
	c.t.Helper()
	if _, err := c.conn.Write(b); err != nil {
		c.t.Fatalf("write: %s", err)
	}
}

func (c *client) u32() uint32 { return binary.BigEndian.Uint32(c.read(4)) }

// securityTypes runs the version exchange and returns the offered types.
func (c *client) securityTypes(version string) []byte {
	c.t.Helper()
	if got := string(c.read(12)); got != "RFB 003.008\n" {
		c.t.Fatalf("server version %q", got)
	}
	c.write([]byte(version))
	n := int(c.read(1)[0])
	if n == 0 {
		c.t.Fatalf("no security types: %s", c.read(int(c.u32())))
	}
	return c.read(n)
}

// vencrypt runs VeNCrypt X509Plain and returns the SecurityResult and its
// reason.
func (c *client) vencrypt(username, password string) (uint32, string) {
	c.t.Helper()
	c.write([]byte{securityVeNCrypt})
	if v := c.read(2); v[0] != 0 || v[1] != 2 {
		c.t.Fatalf("VeNCrypt version %v", v)
	}
	c.write([]byte{0, 2})
	if ack := c.read(1)[0]; ack != 0 {
		c.t.Fatalf("VeNCrypt version refused: %d", ack)
	}
	subtypes := c.read(int(c.read(1)[0]) * 4)
	if !bytes.Equal(subtypes, []byte{0, 0, 1, 6}) {
		c.t.Fatalf("VeNCrypt subtypes %v, want X509Plain only", subtypes)
	}
	c.write(binary.BigEndian.AppendUint32(nil, vencryptX509Plain))
	if ok := c.read(1)[0]; ok != 1 {
		c.t.Fatalf("X509Plain refused: %d", ok)
	}

	tlsConn := tls.Client(c.conn, &tls.Config{InsecureSkipVerify: true})
	if err := tlsConn.Handshake(); err != nil {
		c.t.Fatalf("TLS handshake: %s", err)
	}
	c.conn = tlsConn
	c.r = bufio.NewReader(tlsConn)

	msg := binary.BigEndian.AppendUint32(nil, uint32(len(username)))
	msg = binary.BigEndian.AppendUint32(msg, uint32(len(password)))
	msg = append(msg, username...)
	c.write(append(msg, password...))
	return c.securityResult()
}

func (c *client) securityResult() (uint32, string) {
	c.t.Helper()
	result := c.u32()
	if result == 0 {
		return 0, ""
	}
	return result, string(c.read(int(c.u32())))
}

// init sends ClientInit and reads ServerInit.
func (c *client) init() {
	c.t.Helper()
	c.write([]byte{1})
	head := c.read(24)
	c.width = int(binary.BigEndian.Uint16(head[0:]))
	c.height = int(binary.BigEndian.Uint16(head[2:]))
	c.name = string(c.read(int(binary.BigEndian.Uint32(head[20:]))))
}

func (c *client) setEncodings(encodings ...int32) {
	msg := []byte{msgSetEncodings, 0}
	msg = binary.BigEndian.AppendUint16(msg, uint16(len(encodings)))
	for _, e := range encodings {
		msg = binary.BigEndian.AppendUint32(msg, uint32(e))
	}
	c.write(msg)
}

func (c *client) requestUpdate(incremental bool) {
	msg := []byte{msgFramebufferUpdateRequest, 0}
	if incremental {
		msg[1] = 1
	}
	msg = binary.BigEndian.AppendUint16(msg, 0)
	msg = binary.BigEndian.AppendUint16(msg, 0)
	msg = binary.BigEndian.AppendUint16(msg, uint16(c.width))
	msg = binary.BigEndian.AppendUint16(msg, uint16(c.height))
	c.write(msg)
}

type rect struct {
	x, y, w, h int
	encoding   int32
	jpeg       []byte
}

// update reads one FramebufferUpdate.
func (c *client) update() []rect {
	c.t.Helper()
	head := c.read(4)
	if head[0] != msgFramebufferUpdate {
		c.t.Fatalf("message type %d, want FramebufferUpdate", head[0])
	}
	rects := make([]rect, binary.BigEndian.Uint16(head[2:]))
	for i := range rects {
		h := c.read(12)
		r := rect{
			x:        int(binary.BigEndian.Uint16(h[0:])),
			y:        int(binary.BigEndian.Uint16(h[2:])),
			w:        int(binary.BigEndian.Uint16(h[4:])),
			h:        int(binary.BigEndian.Uint16(h[6:])),
			encoding: int32(binary.BigEndian.Uint32(h[8:])),
		}
		switch r.encoding {
		case encodingTight:
			if control := c.read(1)[0]; control != tightJPEG {
				c.t.Fatalf("Tight compression control %#x, want JPEG", control)
			}
			length := 0
			for shift := 0; shift < 21; shift += 7 {
				b := c.read(1)[0]
				if shift == 14 {
					length |= int(b) << 14
					break
				}
				length |= int(b&0x7f) << shift
				if b&0x80 == 0 {
					break
				}
			}
			r.jpeg = c.read(length)
		case encodingDesktopSize:
			c.width, c.height = r.w, r.h
		default:
			c.t.Fatalf("unexpected encoding %d", r.encoding)
		}
		rects[i] = r
	}
	return rects
}

// connect runs a whole VeNCrypt handshake as admin.
func (h *harness) connect(t *testing.T) *client {
	t.Helper()
	c := dial(t, h.addr)
	c.securityTypes("RFB 003.008\n")
	if result, reason := c.vencrypt("admin", "right-password"); result != 0 {
		t.Fatalf("login failed: %s", reason)
	}
	c.init()
	return c
}

func expectClosed(t *testing.T, c *client) {
	t.Helper()
	_ = c.conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	_, err := c.r.ReadByte()
	var netErr net.Error
	if err == nil || (errors.As(err, &netErr) && netErr.Timeout()) {
		t.Fatalf("the connection stayed open: %v", err)
	}
}

func TestEndToEndJPEGOverVeNCrypt(t *testing.T) {
	h := newHarness(t, config.VNC{})
	c := h.connect(t)

	if c.width != 64 || c.height != 48 || c.name != "test-kvm" {
		t.Fatalf("ServerInit = %dx%d %q", c.width, c.height, c.name)
	}

	frames, input := h.session(t, 0)
	c.setEncodings(encodingTight, encodingDesktopSize, -32)
	c.requestUpdate(false)

	// A full refresh before any frame asks the stream for one at once.
	select {
	case <-frames.refreshes:
	case <-time.After(2 * time.Second):
		t.Fatal("the stream was not asked to refresh")
	}

	frame := testJPEG(t, 64, 48)
	frames.ch <- frame
	rects := c.update()
	if len(rects) != 1 || rects[0].encoding != encodingTight || rects[0].w != 64 || rects[0].h != 48 {
		t.Fatalf("update = %+v", rects)
	}
	if !bytes.Equal(rects[0].jpeg, frame) {
		t.Fatal("the JPEG rectangle does not carry the frame unchanged")
	}
	if _, err := jpeg.Decode(bytes.NewReader(rects[0].jpeg)); err != nil {
		t.Fatalf("the JPEG rectangle does not decode: %s", err)
	}

	// A resize goes out as a DesktopSize update of its own, and the frame
	// follows on the next request.
	c.requestUpdate(true)
	small := testJPEG(t, 32, 24)
	frames.ch <- small
	rects = c.update()
	if len(rects) != 1 || rects[0].encoding != encodingDesktopSize || rects[0].w != 32 || rects[0].h != 24 {
		t.Fatalf("resize update = %+v", rects)
	}
	c.requestUpdate(false)
	rects = c.update()
	if len(rects) != 1 || rects[0].encoding != encodingTight || !bytes.Equal(rects[0].jpeg, small) {
		t.Fatalf("update after the resize = %+v", rects)
	}

	// A full refresh is answered with the last frame, without the stream.
	c.requestUpdate(false)
	rects = c.update()
	if len(rects) != 1 || !bytes.Equal(rects[0].jpeg, small) {
		t.Fatal("a full refresh did not resend the last frame")
	}

	// Input reaches the HID, scaled to the new size.
	key := []byte{msgKeyEvent, 1, 0, 0}
	c.write(binary.BigEndian.AppendUint32(key, 'A'))
	c.write([]byte{msgPointerEvent, 1, 0, 31, 0, 23})
	waitUntil(t, "the input", func() bool {
		keyboard, pointer, _ := input.snapshot()
		return len(keyboard) == 1 && len(pointer) == 1
	})
	keyboard, pointer, _ := input.snapshot()
	if !bytes.Equal(keyboard[0], []byte{modLeftShift, 0, 0x04, 0, 0, 0, 0, 0}) {
		t.Fatalf("keyboard report = %v", keyboard[0])
	}
	if !bytes.Equal(pointer[0], []byte{hidButtonLeft, 0x00, 0x80, 0x00, 0x80, 0}) {
		t.Fatalf("pointer report = %v", pointer[0])
	}

	state := h.srv.State()
	if state.Session == nil || state.Session.User != "admin" || state.Session.Method != methodVeNCrypt ||
		state.Session.FramesSent != 3 || state.Session.Width != 32 {
		t.Fatalf("state = %+v", state.Session)
	}
	if !strings.HasPrefix(state.Session.Client, "127.0.0.1:") {
		t.Fatalf("client address = %q", state.Session.Client)
	}

	// The settings page ends the session, which releases the input and the
	// stream.
	if !h.srv.Disconnect() {
		t.Fatal("Disconnect found no session")
	}
	expectClosed(t, c)
	waitUntil(t, "the cleanup", func() bool {
		_, _, closed := input.snapshot()
		select {
		case <-frames.closed:
			return closed
		default:
			return false
		}
	})
	if h.srv.State().Session != nil {
		t.Fatal("the state still shows a session")
	}
}

// The server takes a frame only while the client has asked for one, and not
// faster than the frame rate limit.
func TestUpdatesFollowRequestsAndTheFrameRateLimit(t *testing.T) {
	h := newHarness(t, config.VNC{MaxFPS: 10})
	c := h.connect(t)
	frames, _ := h.session(t, 0)
	c.setEncodings(encodingTight)

	frame := testJPEG(t, 64, 48)
	frames.ch <- frame
	time.Sleep(100 * time.Millisecond)
	if len(frames.ch) != 1 {
		t.Fatal("a frame was taken with no request outstanding")
	}

	c.requestUpdate(true)
	c.update()
	first := time.Now()

	c.requestUpdate(true)
	frames.ch <- frame
	c.update()
	if gap := time.Since(first); gap < 80*time.Millisecond {
		t.Fatalf("two updates %s apart at a 10 fps limit", gap)
	}
}

func TestAWrongPasswordFails(t *testing.T) {
	h := newHarness(t, config.VNC{})
	c := dial(t, h.addr)
	c.securityTypes("RFB 003.008\n")
	result, reason := c.vencrypt("admin", "wrong")
	if result != 1 || reason != errAuthFailed.Error() {
		t.Fatalf("SecurityResult = %d %q", result, reason)
	}
	if h.limiter.count() != 1 {
		t.Fatalf("the limiter recorded %d failures", h.limiter.count())
	}
	expectClosed(t, c)
	if h.srv.State().Session != nil {
		t.Fatal("a failed login opened a session")
	}
}

func TestOnlyOneSessionAtATime(t *testing.T) {
	h := newHarness(t, config.VNC{})
	first := h.connect(t)

	second := dial(t, h.addr)
	second.securityTypes("RFB 003.008\n")
	result, reason := second.vencrypt("admin", "right-password")
	if result != 1 || reason != errSessionActive.Error() {
		t.Fatalf("second session: SecurityResult = %d %q", result, reason)
	}
	expectClosed(t, second)

	// The first session is untouched.
	frames, _ := h.session(t, 0)
	first.setEncodings(encodingTight)
	first.requestUpdate(false)
	frames.ch <- testJPEG(t, 64, 48)
	first.update()

	// Once it ends, the next client gets in.
	_ = first.conn.Close()
	waitUntil(t, "the session to end", func() bool { return h.srv.State().Session == nil })
	h.connect(t)
}

func TestPlainVNCAuthIsOfferedOnlyWhenEnabled(t *testing.T) {
	h := newHarness(t, config.VNC{})
	_ = h.password.Set("secret12")
	c := dial(t, h.addr)
	if types := c.securityTypes("RFB 003.008\n"); !bytes.Equal(types, []byte{securityVeNCrypt}) {
		t.Fatalf("security types %v with plain authentication off", types)
	}

	h.updateSettings(func(s *config.VNC) { s.VNCAuth = true })
	c = dial(t, h.addr)
	if types := c.securityTypes("RFB 003.008\n"); !bytes.Equal(types, []byte{securityVeNCrypt, securityVNCAuth}) {
		t.Fatalf("security types %v with plain authentication on", types)
	}
	c.write([]byte{securityVNCAuth})
	challenge := c.read(16)
	c.write(clientVNCAuthResponse(t, "wrong123", challenge))
	if result, reason := c.securityResult(); result != 1 || reason != errAuthFailed.Error() {
		t.Fatalf("wrong VNC password: %d %q", result, reason)
	}

	c = dial(t, h.addr)
	c.securityTypes("RFB 003.007\n")
	c.write([]byte{securityVNCAuth})
	challenge = c.read(16)
	c.write(clientVNCAuthResponse(t, "secret12", challenge))
	if result := c.u32(); result != 0 {
		t.Fatalf("right VNC password: SecurityResult %d", result)
	}
	c.init()
	if state := h.srv.State(); state.Session == nil || state.Session.Method != methodVNC {
		t.Fatalf("state = %+v", state.Session)
	}
}

// A 3.7 client gets no reason after a failure, and a type that was not
// offered is refused.
func TestATypeNotOfferedIsRefused(t *testing.T) {
	h := newHarness(t, config.VNC{})
	c := dial(t, h.addr)
	c.securityTypes("RFB 003.007\n")
	c.write([]byte{securityVNCAuth})
	if result := c.u32(); result != 1 {
		t.Fatalf("SecurityResult %d", result)
	}
	expectClosed(t, c)
}

func TestVeNCryptRefusesOtherSubtypes(t *testing.T) {
	h := newHarness(t, config.VNC{})
	c := dial(t, h.addr)
	c.securityTypes("RFB 003.008\n")
	c.write([]byte{securityVeNCrypt})
	c.read(2)
	c.write([]byte{0, 2})
	c.read(1)
	c.read(int(c.read(1)[0]) * 4)
	// TLSPlain, anonymous TLS, which the server does not offer.
	c.write(binary.BigEndian.AppendUint32(nil, 259))
	if ack := c.read(1)[0]; ack != 0 {
		t.Fatalf("an unoffered subtype was accepted: %d", ack)
	}
	expectClosed(t, c)
}

func TestAClientWithoutTightIsDisconnectedWithAReason(t *testing.T) {
	h := newHarness(t, config.VNC{})
	c := h.connect(t)
	// Raw and Hextile only.
	c.setEncodings(0, 5)
	c.requestUpdate(false)
	expectClosed(t, c)

	waitUntil(t, "the reason", func() bool {
		return strings.Contains(h.srv.State().LastError, "Tight")
	})
}

func TestAResizeWithoutDesktopSizeEndsTheSession(t *testing.T) {
	h := newHarness(t, config.VNC{})
	c := h.connect(t)
	frames, _ := h.session(t, 0)
	c.setEncodings(encodingTight)
	c.requestUpdate(false)
	frames.ch <- testJPEG(t, 32, 24)
	expectClosed(t, c)

	waitUntil(t, "the reason", func() bool {
		return strings.Contains(h.srv.State().LastError, "DesktopSize")
	})
}

func TestTurningTheServerOffEndsTheSession(t *testing.T) {
	h := newHarness(t, config.VNC{})
	c := h.connect(t)
	h.updateSettings(func(s *config.VNC) { s.Enabled = false })
	if err := h.srv.Apply(); err != nil {
		t.Fatal(err)
	}
	expectClosed(t, c)
	if state := h.srv.State(); state.Listening {
		t.Fatal("the server still listens")
	}
	if _, err := net.DialTimeout("tcp", h.addr, time.Second); err == nil {
		t.Fatal("the port still accepts connections")
	}
}

func waitUntil(t *testing.T, what string, condition func() bool) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}
