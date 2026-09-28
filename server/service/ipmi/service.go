// Package ipmi serves IPMI 2.0 over LAN (RMCP+, what ipmitool calls
// lanplus) for the managed host's power: chassis status, power on, off,
// cycle, reset and soft off through the front-panel buttons, and enough of
// the rest for ipmitool to log in and run `mc info`. There is no SEL, SDR,
// FRU or Serial over LAN, and no IPMI 1.5 session.
package ipmi

import (
	"errors"
	"net"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
)

// Port is the RMCP port, where every IPMI client looks.
const Port = 623

// The buttons PressButton accepts. They are the names service/vm uses.
const (
	ButtonPower = "power"
	ButtonReset = "reset"
)

// Deps is everything the service reaches outside this package. The router
// wires the real ones; tests pass fakes.
type Deps struct {
	// PressButton holds ButtonPower or ButtonReset for d.
	PressButton func(kind string, d time.Duration) error
	// PowerLED reports whether the host's power LED is lit.
	PowerLED func() (bool, error)
	// PowerLEDConnected reports whether the owner has said the LED header is
	// wired. When it is not, the LED is never read: the state is unknown.
	PowerLEDConnected func() bool

	Accounts Accounts
	// Keyring opens the sealed IPMI passwords.
	Keyring *Keyring
	// Limiter is the login brute-force limit. Nil means none.
	Limiter Limiter

	// FirmwareVersion is the version Get Device ID reports.
	FirmwareVersion func() string
	// GUID is the managed system's GUID, in the byte order it goes over
	// the wire.
	GUID [16]byte

	// Addr is the UDP address the service listens on while it is on.
	Addr string
	// Enabled reports the ipmi.enabled setting.
	Enabled func() bool
	// SetEnabled saves the setting. Enabled reports the new value once it
	// returns nil.
	SetEnabled func(on bool) error

	// Now is the clock the session timeouts run on. Nil means time.Now.
	Now func() time.Time
}

// Accounts is the part of authn.Store the service uses.
type Accounts interface {
	List() ([]authn.UserInfo, error)
	Get(username string) (*authn.User, error)
	Authenticate(username, password string) (*authn.User, bool, error)
	SetIPMIPassword(username, sealed string) (*authn.User, error)
}

// Limiter is the login brute-force limit, keyed by client address and
// account.
type Limiter interface {
	Locked(ip, username string) bool
	Failed(ip, username string)
	Succeeded(ip, username string)
}

// Service is the IPMI responder.
type Service struct {
	deps Deps

	// mu guards the socket and the sessions. One goroutine reads the
	// socket and handles a packet at a time; the web UI's handlers take
	// the lock to change the setting or end sessions.
	mu       sync.Mutex
	conn     net.PacketConn
	sessions map[uint32]*session

	// powerMu is held from the LED read that plans a Chassis Control press
	// until the host has had time to follow it.
	powerMu sync.Mutex
	// switchMu serializes changes to the enabled setting.
	switchMu sync.Mutex
}

func New(deps Deps) *Service {
	if deps.Now == nil {
		deps.Now = time.Now
	}
	if deps.PowerLEDConnected == nil {
		deps.PowerLEDConnected = func() bool { return true }
	}
	if deps.Enabled == nil {
		deps.Enabled = func() bool { return false }
	}
	if deps.FirmwareVersion == nil {
		deps.FirmwareVersion = func() string { return "" }
	}
	return &Service{deps: deps, sessions: map[uint32]*session{}}
}

// Start opens the socket if the service is on. A socket that cannot be
// opened is logged and the service stays off until it is switched again.
func (s *Service) Start() {
	if !s.deps.Enabled() {
		return
	}
	if err := s.open(); err != nil {
		log.Errorf("ipmi: listen on %s: %s", s.deps.Addr, err)
	}
}

// Stop closes the socket and ends every session.
func (s *Service) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closeLocked()
}

// LocalAddr is the address the socket is bound to, or nil while the
// service is off. Tests listen on port 0 and ask.
func (s *Service) LocalAddr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn == nil {
		return nil
	}
	return s.conn.LocalAddr()
}

func (s *Service) open() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.conn != nil {
		return nil
	}
	conn, err := net.ListenPacket("udp", s.deps.Addr)
	if err != nil {
		return err
	}
	s.conn = conn
	go s.serve(conn)
	log.Infof("ipmi: listening on %s", conn.LocalAddr())
	return nil
}

func (s *Service) closeLocked() {
	if s.conn == nil {
		return
	}
	_ = s.conn.Close()
	s.conn = nil
	clear(s.sessions)
}

// maxPacket is more than any request the service answers can take. An
// RMCP+ packet with a full IPMI message is well under 300 bytes.
const maxPacket = 1024

func (s *Service) serve(conn net.PacketConn) {
	buf := make([]byte, maxPacket)
	for {
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Debugf("ipmi: read: %s", err)
			continue
		}

		packet := append([]byte(nil), buf[:n]...)
		s.mu.Lock()
		// A packet read just before the socket was closed, or a new one
		// opened, belongs to nothing now.
		if s.conn != conn {
			s.mu.Unlock()
			return
		}
		reply := s.handleLocked(packet, addr)
		s.mu.Unlock()

		if reply != nil {
			if _, err := conn.WriteTo(reply, addr); err != nil {
				log.Debugf("ipmi: write to %s: %s", addr, err)
			}
		}
	}
}

// handleLocked answers one packet, or returns nil to drop it. A packet
// that is malformed, or fails its integrity check, gets no answer: IPMI
// has nowhere to report either, and silence tells a prober least.
func (s *Service) handleLocked(packet []byte, addr net.Addr) []byte {
	authType, err := parseRMCP(packet)
	if err != nil {
		return nil
	}

	switch authType {
	case authTypeNone:
		p, err := parseV15(packet)
		if err != nil || p.sessionID != 0 {
			return nil
		}
		return s.handleV15(p)
	case authTypeRMCPPlus:
		p, err := parseV20(packet)
		if err != nil {
			return nil
		}
		return s.handleV20(p, addr)
	}
	// MD2, MD5 and straight password are IPMI 1.5 sessions, which are off.
	return nil
}

// handleV15 answers an IPMI 1.5 message outside a session. Clients send
// Get Channel Authentication Capabilities this way before they know what
// the BMC speaks, and that is the reason to answer at all.
func (s *Service) handleV15(p *v15Packet) []byte {
	req, err := parseRequest(p.msg)
	if err != nil {
		return nil
	}
	cc, data := s.dispatch(nil, req)
	return buildV15(buildResponse(req, cc, data))
}

func (s *Service) handleV20(p *v20Packet, addr net.Addr) []byte {
	now := s.deps.Now()
	s.expireLocked(now)

	switch p.kind() {
	case payloadOpenRequest:
		return s.openSession(p, addr, now)
	case payloadRAKP1:
		return s.rakp1(p, addr, now)
	case payloadRAKP3:
		return s.rakp3(p, addr, now)
	case payloadIPMI:
		if p.sessionID == 0 {
			return s.presession(p)
		}
		return s.inSession(p, addr, now)
	}
	return nil
}

// presession answers an unauthenticated RMCP+ IPMI message with session
// ID 0. Only the discovery commands answer anything useful there.
func (s *Service) presession(p *v20Packet) []byte {
	if p.encrypted() || p.authenticated() {
		return nil
	}
	req, err := parseRequest(p.payload)
	if err != nil {
		return nil
	}
	cc, data := s.dispatch(nil, req)
	return buildV20(v20Header{payloadType: payloadIPMI}, buildResponse(req, cc, data), nil, nil)
}

// EndUserSessions ends the sessions of one account, after its IPMI
// password changed or went away.
func (s *Service) EndUserSessions(username string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if sess.username == username {
			delete(s.sessions, id)
		}
	}
}
