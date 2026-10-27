// Package srt serves the board's picture and sound over SRT, in MPEG-TS, for
// OBS's Media Source and other SRT callers (ironkvm-dist#143).
//
// Like the RTSP server it does no media work: the H.264 or H.265 and the Opus
// come from the shared tap (service/stream/egress) as WebRTC gets them, and
// are only put in MPEG-TS. Each caller has its own muxer, so it starts on a
// keyframe with fresh tables, and its own tap reader, so a caller on a bad
// link loses frames to the next keyframe while the others play on.
//
// The listener is off until an administrator switches it on, and it serves
// only an encrypted caller with the board's passphrase. The caller's stream ID
// carries a token that can do no more than watch, as WHEP's Bearer token
// does, which names it in the audit log and ends it when the token is
// revoked. The allowed networks list applies before the handshake's key is
// checked, and at most egress.MaxClients callers are served at once.
package srt

import (
	"bytes"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/substructs"
	gosrt "github.com/datarhei/gosrt"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/common"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/stream/egress"
)

const (
	// payloadSize is seven MPEG-TS packets, the payload every SRT
	// implementation sends live MPEG-TS in.
	payloadSize = 7 * 188

	// tsBase lifts every timestamp clear of zero. The muxer puts the PCR
	// 100 ms before the first DTS, and audio may begin a little before the
	// first picture.
	tsBase = 2 * 90000

	// defaultRecheck is how often a caller's credential is checked again.
	defaultRecheck = 5 * time.Second
)

// Deps are what the server needs from the rest of NanoKVM-Server. Tests
// replace every one of them.
type Deps struct {
	// Settings returns the current settings with the defaults filled in.
	Settings    func() config.SRT
	SetSettings func(config.SRT) error
	Passphrase  PassphraseStore

	// ClientAllowed reports whether a caller at this address may connect.
	// Nil allows every address.
	ClientAllowed func(net.Addr) bool
	Authenticate  egress.Authenticator

	Tap *egress.Tap
	// AudioAvailable reports whether the host's sound can be captured, which
	// decides whether a caller's stream has an audio track.
	AudioAvailable func() bool

	// Recheck is how often a caller's credential is checked again; zero
	// means the default.
	Recheck time.Duration
}

// Server is the SRT listener. It listens only while the setting is on and a
// passphrase is set.
type Server struct {
	deps Deps

	mu         sync.Mutex
	listener   gosrt.Listener
	running    config.SRT
	passphrase string
	listenErr  string
	stopWatch  chan struct{}
	clients    map[*client]struct{}
}

// client is a caller being served.
type client struct {
	conn    gosrt.Conn
	viewer  egress.Viewer
	address string
	since   time.Time
}

func New(deps Deps) *Server {
	if deps.Recheck <= 0 {
		deps.Recheck = defaultRecheck
	}
	if deps.AudioAvailable == nil {
		deps.AudioAvailable = func() bool { return false }
	}
	return &Server{deps: deps, clients: make(map[*client]struct{})}
}

var errNoPassphrase = errors.New("set a passphrase first: SRT here is always encrypted")

// Apply brings the listener in line with the settings and the passphrase: it
// opens it when SRT is on, closes it when SRT is off, and opens it again on a
// change. Closing it ends every caller.
func (s *Server) Apply() error {
	settings := s.deps.Settings()
	passphrase := ""
	if s.deps.Passphrase != nil {
		p, ok, err := s.deps.Passphrase.Get()
		if err != nil {
			log.Errorf("srt: read the passphrase: %s", err)
		}
		if ok {
			passphrase = p
		}
	}

	s.mu.Lock()
	if s.listener != nil && settings.Enabled && settings == s.running && passphrase == s.passphrase {
		s.mu.Unlock()
		return nil
	}
	old, oldWatch := s.listener, s.stopWatch
	s.listener, s.stopWatch = nil, nil
	s.running, s.passphrase, s.listenErr = config.SRT{}, "", ""
	s.mu.Unlock()

	if oldWatch != nil {
		close(oldWatch)
	}
	if old != nil {
		// Closing the listener closes every connection on it, and each
		// caller's goroutine sees its write fail and leaves.
		old.Close()
		log.Infof("srt: stopped")
	}

	if !settings.Enabled {
		return nil
	}
	if passphrase == "" {
		s.mu.Lock()
		s.listenErr = errNoPassphrase.Error()
		s.mu.Unlock()
		return errNoPassphrase
	}

	cfg := gosrt.DefaultConfig()
	latency := time.Duration(settings.LatencyMs) * time.Millisecond
	cfg.ReceiverLatency = latency
	cfg.PeerLatency = latency
	cfg.PayloadSize = payloadSize
	cfg.EnforcedEncryption = true

	listener, err := gosrt.Listen("srt", ":"+strconv.Itoa(settings.Port), cfg)
	if err != nil {
		s.mu.Lock()
		s.listenErr = err.Error()
		s.mu.Unlock()
		log.Errorf("srt: listen on port %d: %s", settings.Port, err)
		return err
	}

	stop := make(chan struct{})
	s.mu.Lock()
	s.listener, s.running, s.passphrase, s.stopWatch = listener, settings, passphrase, stop
	s.mu.Unlock()

	go s.accept(listener, passphrase)
	go s.watch(stop)

	log.Infof("srt: listening on UDP port %d, latency %d ms", settings.Port, settings.LatencyMs)
	return nil
}

// Stop closes the listener and ends every caller.
func (s *Server) Stop() {
	s.mu.Lock()
	old, oldWatch := s.listener, s.stopWatch
	s.listener, s.stopWatch = nil, nil
	s.mu.Unlock()

	if oldWatch != nil {
		close(oldWatch)
	}
	if old != nil {
		old.Close()
	}
}

// Addr is the listener's address, or nil when it is not listening.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.listener == nil {
		return nil
	}
	return s.listener.Addr()
}

// accept answers handshakes one at a time until the listener closes. One at a
// time is the point: checking a passphrase is a PBKDF2 derivation, which the
// single core must not be asked for by a flood of callers at once.
func (s *Server) accept(listener gosrt.Listener, passphrase string) {
	for {
		req, err := listener.Accept2()
		if err != nil {
			if errors.Is(err, gosrt.ErrListenerClosed) {
				return
			}
			log.Debugf("srt: accept: %s", err)
			continue
		}
		s.handshake(req, passphrase)
	}
}

// handshake decides on one caller, cheapest check first.
func (s *Server) handshake(req gosrt.ConnRequest, passphrase string) {
	addr := req.RemoteAddr()

	if s.deps.ClientAllowed != nil && !s.deps.ClientAllowed(addr) {
		log.Debugf("srt: refused %s, not an allowed network", addr)
		req.Reject(gosrt.REJX_FORBIDDEN)
		return
	}
	// A version 4 handshake has no stream ID and no encryption, and it is
	// always a caller that wants to publish.
	if req.Version() != 5 {
		req.Reject(gosrt.REJ_VERSION)
		return
	}
	if s.Clients() >= egress.MaxClients {
		log.Debugf("srt: refused %s, %d callers already", addr, egress.MaxClients)
		req.Reject(gosrt.REJX_OVERLOAD)
		return
	}
	if !req.IsEncrypted() {
		log.Debugf("srt: refused %s, which did not encrypt", addr)
		req.Reject(gosrt.REJ_UNSECURE)
		return
	}
	if err := req.SetPassphrase(passphrase); err != nil {
		log.Debugf("srt: refused %s, wrong passphrase", addr)
		req.Reject(gosrt.REJ_BADSECRET)
		return
	}

	var who egress.Viewer
	ok := false
	if s.deps.Authenticate != nil {
		who, ok = s.deps.Authenticate(streamToken(req.StreamId()))
	}
	if !ok {
		log.Debugf("srt: refused %s, the stream ID holds no view-only token", addr)
		req.Reject(gosrt.REJX_UNAUTHORIZED)
		return
	}

	conn, err := req.Accept()
	if err != nil {
		log.Debugf("srt: accept %s: %s", addr, err)
		return
	}

	c := &client{conn: conn, viewer: who, address: addr.String(), since: time.Now()}
	s.mu.Lock()
	s.clients[c] = struct{}{}
	s.mu.Unlock()

	go s.serve(c)
}

// streamToken is the token in a caller's stream ID: the whole ID, or in the
// access control syntax of the SRT specification ("#!::s=...,m=request") the
// session ID s, or the user name u.
func streamToken(streamID string) string {
	rest, ok := strings.CutPrefix(streamID, "#!::")
	if !ok {
		return strings.TrimSpace(streamID)
	}

	var user string
	for _, pair := range strings.Split(rest, ",") {
		key, value, _ := strings.Cut(pair, "=")
		switch strings.TrimSpace(key) {
		case "s":
			return strings.TrimSpace(value)
		case "u":
			user = strings.TrimSpace(value)
		}
	}
	return user
}

// Clients is how many callers are served.
func (s *Server) Clients() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return len(s.clients)
}

// serve sends one caller its stream until it goes, its write fails, or the
// codec changes, which a muxer's tables cannot follow.
func (s *Server) serve(c *client) {
	withAudio := s.deps.AudioAvailable()
	reader := s.deps.Tap.Subscribe("srt", withAudio)

	reason := "the caller left"
	defer func() {
		reader.Close()
		_ = c.conn.Close()

		s.mu.Lock()
		delete(s.clients, c)
		s.mu.Unlock()

		record(c, "srt-end", reason+" after "+time.Since(c.since).Round(time.Second).String())
	}()

	record(c, "srt-start", "")

	// The caller sends nothing but control packets, which gosrt handles
	// itself, so a read returns only when the connection has gone. That is
	// how a caller that leaves while there is no picture to send is noticed.
	gone := make(chan struct{})
	go func() {
		defer close(gone)
		buf := make([]byte, 1500)
		for {
			if _, err := c.conn.Read(buf); err != nil {
				return
			}
		}
	}()

	var out bytes.Buffer
	var muxer *muxer

	for {
		select {
		case <-gone:
			return
		case unit, ok := <-reader.Units():
			if !ok {
				reason = "the capture ended"
				return
			}
			if muxer == nil {
				// The tap's first unit for a reader is a keyframe.
				m, err := newMuxer(&out, unit, withAudio)
				if err != nil {
					reason = err.Error()
					return
				}
				muxer = m
			}
			if !unit.Audio && unit.Codec != muxer.codec {
				reason = "the codec changed"
				return
			}
			if err := muxer.write(unit); err != nil {
				log.Debugf("srt: mux: %s", err)
				out.Reset()
				continue
			}
			if out.Len() == 0 {
				continue
			}
			// gosrt cuts the buffer into payloadSize packets, whole
			// MPEG-TS packets each, and never blocks: a send queue that is
			// full fails the write, and the caller is let go.
			if _, err := c.conn.Write(out.Bytes()); err != nil {
				reason = "the link could not keep up"
				return
			}
			out.Reset()
		}
	}
}

// muxer is one caller's MPEG-TS writer.
type muxer struct {
	writer *mpegts.Writer
	video  *mpegts.Track
	audio  *mpegts.Track
	codec  uint8
	// start is the first unit's PTS. The stream's timestamps count from it,
	// so they are small whatever the tap's age.
	start time.Duration
}

func newMuxer(out *bytes.Buffer, first egress.Unit, withAudio bool) (*muxer, error) {
	m := &muxer{codec: first.Codec, start: first.PTS}

	var videoCodec codecs.Codec = &codecs.H264{}
	if first.Codec == common.CodecH265 {
		videoCodec = &codecs.H265{}
	}
	m.video = &mpegts.Track{Codec: videoCodec}
	tracks := []*mpegts.Track{m.video}
	if withAudio {
		m.audio = &mpegts.Track{Codec: &codecs.Opus{
			Desc: &substructs.OpusAudioDescriptor{ChannelConfigCode: uint8(audio.HostFormat.Channels)},
		}}
		tracks = append(tracks, m.audio)
	}

	m.writer = &mpegts.Writer{W: out, Tracks: tracks}
	if err := m.writer.Initialize(); err != nil {
		return nil, err
	}
	return m, nil
}

// ts is a PTS in the 90 kHz units of MPEG-TS.
func (m *muxer) ts(pts time.Duration) int64 {
	return tsBase + (pts-m.start).Microseconds()*9/100
}

func (m *muxer) write(unit egress.Unit) error {
	if unit.Audio {
		if m.audio == nil {
			return nil
		}
		return m.writer.WriteOpus(m.audio, m.ts(unit.PTS), [][]byte{unit.Opus})
	}

	pts := m.ts(unit.PTS)
	// The encoder makes no B-frames, so every picture is decoded when it is
	// shown.
	if m.codec == common.CodecH265 {
		return m.writer.WriteH265(m.video, pts, pts, unit.NALUs)
	}
	return m.writer.WriteH264(m.video, pts, pts, unit.NALUs)
}

// record writes a caller's start or end to the audit log.
func record(c *client, action, detail string) {
	host, _, err := net.SplitHostPort(c.address)
	if err != nil {
		host = c.address
	}
	audit.Record(audit.Event{
		Actor:    c.viewer.Username,
		Via:      c.viewer.Via,
		Key:      c.viewer.Key,
		Role:     c.viewer.Role,
		IP:       host,
		Category: audit.CategoryAuth,
		Action:   action,
		Target:   "srt",
		Outcome:  audit.OutcomeOK,
		Detail:   detail,
	})
}

// watch ends a caller whose credential no longer holds, until stop closes.
func (s *Server) watch(stop chan struct{}) {
	ticker := time.NewTicker(s.deps.Recheck)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
		}

		s.mu.Lock()
		clients := make([]*client, 0, len(s.clients))
		for c := range s.clients {
			clients = append(clients, c)
		}
		s.mu.Unlock()

		// Valid reads the key store, so it runs outside the lock.
		for _, c := range clients {
			if c.viewer.Valid != nil && !c.viewer.Valid() {
				log.Infof("srt: ended %s, the credential was revoked or expired", c.address)
				_ = c.conn.Close()
			}
		}
	}
}
