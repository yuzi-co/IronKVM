// Package vnc is an RFB 3.8 server that lets a standard VNC client view and
// control the host.
//
// It does no pixel work. The frames are the JPEG frames the hardware encoder
// already makes for the MJPEG stream, sent as Tight JPEG rectangles, and the
// input goes through the same HID writers and the same manual-input
// arbitration as the web UI's websocket. docs/superpowers/specs/
// 2026-09-28-vnc-design.md holds the design.
package vnc

import (
	"crypto/tls"
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
)

// maxHandshakes bounds the handshakes that run at once. Each TLS handshake
// costs the single core a public-key operation, and a password check a bcrypt,
// so a flood of connections must not queue up behind them.
const maxHandshakes = 4

// FrameSubscription is a viewer of the MJPEG stream. mjpeg.Subscription is
// the one the server runs with.
type FrameSubscription interface {
	Frames() <-chan []byte
	Refresh()
	Close()
}

// Accounts checks a username and a password. authn.Store is the one the server
// runs with.
type Accounts interface {
	Authenticate(username, password string) (*authn.User, bool, error)
}

// Limiter is the web login's brute-force limit. redfish.LoginLimiter is the
// one the server runs with.
type Limiter interface {
	Locked(ip, username string) bool
	Failed(ip, username string)
	Succeeded(ip, username string)
}

// PasswordStore keeps the password for plain VNC authentication.
type PasswordStore interface {
	// Get returns the password, and false when none is set.
	Get() (string, bool, error)
	Set(password string) error
}

// Input takes the keyboard and pointer reports of one session. It must not
// block: a host that stops accepting reports must not stall the session.
type Input interface {
	Keyboard(report []byte)
	Pointer(report []byte)
	// Close releases every key and button still held.
	Close()
}

// Deps is what the server needs from the rest of NanoKVM-Server. Tests replace
// every one of them.
type Deps struct {
	// Settings returns the current settings with the defaults filled in.
	Settings    func() config.VNC
	SetSettings func(config.VNC) error
	// ReservedPorts are the ports the web server listens on, which the VNC
	// server must not take.
	ReservedPorts func() []int

	Subscribe  func() FrameSubscription
	ScreenSize func() (width, height int)
	NewInput   func() Input

	Accounts     Accounts
	Limiter      Limiter
	FailureDelay time.Duration
	Password     PasswordStore
	// Certificate returns the certificate for the VeNCrypt TLS handshake.
	// It is called for every connection, so a new certificate applies to the
	// next client without a restart.
	Certificate func() (*tls.Certificate, error)

	// Name is the desktop name the client shows.
	Name func() string

	// Listen opens the listener. net.Listen when nil.
	Listen func(network, address string) (net.Listener, error)
}

// Server is the VNC server. It listens only while the setting is on, and it
// serves one session at a time.
type Server struct {
	deps Deps

	mu        sync.Mutex
	listener  net.Listener
	port      int
	listenErr string
	active    *session
	lastError string

	handshakes chan struct{}
	sleep      func(time.Duration)
}

func New(deps Deps) *Server {
	if deps.Listen == nil {
		deps.Listen = net.Listen
	}
	if deps.Name == nil {
		deps.Name = func() string { return "IronKVM" }
	}
	return &Server{
		deps:       deps,
		handshakes: make(chan struct{}, maxHandshakes),
		sleep:      time.Sleep,
	}
}

// Apply brings the listener in line with the settings: it opens it when the
// server is on, closes it when the server is off, and moves it when the port
// changed. Closing or moving it ends the open session.
func (s *Server) Apply() error {
	settings := s.deps.Settings()

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil && settings.Enabled && s.port == settings.Port {
		return nil
	}

	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
		s.port = 0
		if s.active != nil {
			s.active.end(errServerStopped)
		}
	}
	s.listenErr = ""

	if !settings.Enabled {
		return nil
	}

	listener, err := s.deps.Listen("tcp", ":"+strconv.Itoa(settings.Port))
	if err != nil {
		s.listenErr = err.Error()
		log.Errorf("vnc: listen on port %d: %s", settings.Port, err)
		return err
	}
	s.listener = listener
	s.port = settings.Port
	log.Infof("vnc: listening on port %d", settings.Port)

	go s.accept(listener)
	return nil
}

// Stop closes the listener and ends the open session.
func (s *Server) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener != nil {
		_ = s.listener.Close()
		s.listener = nil
		s.port = 0
	}
	if s.active != nil {
		s.active.end(errServerStopped)
	}
}

// Addr is the listener's address, or nil when the server is not listening.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

func (s *Server) accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				log.Errorf("vnc: accept: %s", err)
			}
			return
		}

		select {
		case s.handshakes <- struct{}{}:
		default:
			log.Debugf("vnc: refused %s, too many handshakes", conn.RemoteAddr())
			_ = conn.Close()
			continue
		}

		go newSession(s, conn).serve()
	}
}

// claim makes a session the open one. It fails while another is open.
func (s *Server) claim(sess *session) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active != nil || s.listener == nil {
		return false
	}
	s.active = sess
	return true
}

// release forgets a session that ended, and keeps the reason it ended for the
// settings page when the reason is worth showing.
func (s *Server) release(sess *session, reason error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active == sess {
		s.active = nil
	}
	if reason != nil {
		s.lastError = reason.Error()
	}
}

// Disconnect ends the open session. It reports whether there was one.
func (s *Server) Disconnect() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.active == nil {
		return false
	}
	s.active.end(errDisconnected)
	return true
}

// State is what the settings page shows about the server.
type State struct {
	Listening bool         `json:"listening"`
	Port      int          `json:"port,omitempty"`
	Error     string       `json:"error,omitempty"`
	LastError string       `json:"lastError,omitempty"`
	Session   *SessionInfo `json:"session,omitempty"`
}

// SessionInfo describes the open session.
type SessionInfo struct {
	Client     string `json:"client"`
	User       string `json:"user,omitempty"`
	Method     string `json:"method"`
	Since      string `json:"since"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	FramesSent uint64 `json:"framesSent"`
	BytesSent  uint64 `json:"bytesSent"`
}

func (s *Server) State() State {
	s.mu.Lock()
	defer s.mu.Unlock()

	state := State{
		Listening: s.listener != nil,
		Port:      s.port,
		Error:     s.listenErr,
		LastError: s.lastError,
	}
	if s.active != nil {
		info := s.active.info()
		state.Session = &info
	}
	return state
}
