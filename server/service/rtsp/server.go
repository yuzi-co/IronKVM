// Package rtsp serves the board's picture and sound to RTSP players: OBS's
// Media Source, VLC, ffmpeg and NVRs (ironkvm-dist#142).
//
// It does no media work. The video is the H.264 or H.265 the hardware encoder
// makes for WebRTC and the sound the Opus the audio hub makes for it, both
// taken from the shared tap (service/stream/egress) and cut into RTP as they
// are. Every player shares one stream, so the board packetizes each frame once
// however many watch, and gortsplib gives each player its own bounded write
// queue: a player that cannot keep up loses packets, and nobody else waits.
//
// The server is off until an administrator switches it on. A player
// authenticates with Basic authentication, the password being a token that
// can do no more than watch (middleware.AuthenticateViewToken), as WHEP's
// Bearer token is. The allowed networks list applies before a request is
// read, and at most egress.MaxClients players are served at once.
package rtsp

import (
	"errors"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/bluenviron/gortsplib/v5"
	"github.com/bluenviron/gortsplib/v5/pkg/auth"
	"github.com/bluenviron/gortsplib/v5/pkg/base"
	"github.com/bluenviron/gortsplib/v5/pkg/description"
	"github.com/bluenviron/gortsplib/v5/pkg/format"
	"github.com/bluenviron/gortsplib/v5/pkg/headers"
	"github.com/bluenviron/gortsplib/v5/pkg/liberrors"
	"github.com/pion/rtp"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/common"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/stream/egress"
)

const (
	// Path is the path the settings page shows. The server answers on any
	// path, so a player configured with another one, or none, still plays.
	Path = "/stream"

	// writeQueueSize is how many packets each player may have waiting. A
	// 1080p keyframe is about a hundred packets, and 512 is a third of a
	// second at 10 Mbit/s with the sound; a player further behind than that
	// loses packets rather than holding memory. gortsplib needs a power of
	// two.
	writeQueueSize = 512

	// defaultFirstPicture bounds the wait for the keyframe a new stream is
	// described from. The encoder is asked for one at once, so it normally
	// comes within a frame or two; a board with no HDMI signal sends none.
	defaultFirstPicture = 3 * time.Second

	// defaultLinger is how long the stream stays up with nobody watching. A
	// player DESCRIBEs and then SETUPs on a new connection, and some probe
	// first and come back, and starting the stream again costs a keyframe.
	defaultLinger = 15 * time.Second

	// defaultRecheck is how often a player's credential and the setting are
	// checked again, as WHEP does.
	defaultRecheck = 5 * time.Second

	videoPayloadType = 96
	audioPayloadType = 97
	videoClockRate   = 90000
)

var (
	errNotRunning = errors.New("the RTSP server is not running")
	errNoPicture  = errors.New("no picture from the encoder")
)

// Deps are what the server needs from the rest of NanoKVM-Server. Tests
// replace every one of them.
type Deps struct {
	// Settings returns the current settings with the defaults filled in.
	Settings    func() config.RTSP
	SetSettings func(config.RTSP) error
	// ReservedPorts are the ports the web server listens on.
	ReservedPorts func() []int

	// ClientAllowed reports whether a client at this address may connect.
	// Nil allows every address.
	ClientAllowed func(net.Addr) bool
	Authenticate  egress.Authenticator

	Tap *egress.Tap
	// AudioAvailable reports whether the host's sound can be captured, which
	// decides whether a new stream has an audio track.
	AudioAvailable func() bool

	// FirstPicture, Linger and Recheck are the waits above; zero means the
	// default.
	FirstPicture time.Duration
	Linger       time.Duration
	Recheck      time.Duration
}

// Server is the RTSP server. It listens only while the setting is on.
type Server struct {
	deps Deps

	mu        sync.Mutex
	rtsp      *gortsplib.Server
	running   config.RTSP
	listenErr string
	stopWatch chan struct{}
	feed      *feed
	// idleSince is when the last player left the feed, or when the feed
	// started without one.
	idleSince time.Time
	sessions  map[*gortsplib.ServerSession]*session

	// feedMu makes one stream at a time: a second DESCRIBE waits for the
	// first one's keyframe instead of subscribing again.
	feedMu sync.Mutex
}

// session is a player that has SETUP a track.
type session struct {
	viewer  egress.Viewer
	client  string
	since   time.Time
	playing bool
	session *gortsplib.ServerSession
}

func New(deps Deps) *Server {
	if deps.FirstPicture <= 0 {
		deps.FirstPicture = defaultFirstPicture
	}
	if deps.Linger <= 0 {
		deps.Linger = defaultLinger
	}
	if deps.Recheck <= 0 {
		deps.Recheck = defaultRecheck
	}
	if deps.AudioAvailable == nil {
		deps.AudioAvailable = func() bool { return false }
	}
	return &Server{deps: deps, sessions: make(map[*gortsplib.ServerSession]*session)}
}

// Apply brings the listener in line with the settings: it starts the server
// when it is on, stops it when it is off, and starts it again on a change.
// Stopping ends every player.
func (s *Server) Apply() error {
	settings := s.deps.Settings()

	s.feedMu.Lock()
	defer s.feedMu.Unlock()

	s.mu.Lock()
	if s.rtsp != nil && settings.Enabled && settings == s.running {
		s.mu.Unlock()
		return nil
	}
	old, oldFeed, oldWatch := s.rtsp, s.feed, s.stopWatch
	s.rtsp, s.feed, s.stopWatch = nil, nil, nil
	s.running = config.RTSP{}
	s.listenErr = ""
	s.mu.Unlock()

	// Outside the lock: closing a stream and the server closes sessions,
	// and each one reports back through OnSessionClose, which takes it.
	if oldWatch != nil {
		close(oldWatch)
	}
	if oldFeed != nil {
		oldFeed.close()
	}
	if old != nil {
		old.Close()
		log.Infof("rtsp: stopped")
	}

	if !settings.Enabled {
		return nil
	}

	server := &gortsplib.Server{
		Handler:        &handler{s: s},
		RTSPAddress:    ":" + strconv.Itoa(settings.Port),
		WriteQueueSize: writeQueueSize,
		// The password is a token the board keeps only a hash of, so it
		// cannot answer a digest challenge: Basic is the only method.
		AuthMethods: []auth.VerifyMethod{auth.VerifyMethodBasic},
		Listen:      s.listen,
	}
	if settings.UDP {
		server.UDPRTPAddress = ":" + strconv.Itoa(settings.UDPPort)
		server.UDPRTCPAddress = ":" + strconv.Itoa(settings.UDPPort+1)
	}
	if err := server.Start(); err != nil {
		s.mu.Lock()
		s.listenErr = err.Error()
		s.mu.Unlock()
		log.Errorf("rtsp: listen on port %d: %s", settings.Port, err)
		return err
	}

	stop := make(chan struct{})
	s.mu.Lock()
	s.rtsp, s.running, s.stopWatch = server, settings, stop
	s.mu.Unlock()
	go s.watch(stop)

	log.Infof("rtsp: listening on port %d (UDP %t)", settings.Port, settings.UDP)
	return nil
}

// Stop ends every player and closes the listener.
func (s *Server) Stop() {
	s.feedMu.Lock()
	defer s.feedMu.Unlock()

	s.mu.Lock()
	old, oldFeed, oldWatch := s.rtsp, s.feed, s.stopWatch
	s.rtsp, s.feed, s.stopWatch = nil, nil, nil
	s.mu.Unlock()

	if oldWatch != nil {
		close(oldWatch)
	}
	if oldFeed != nil {
		oldFeed.close()
	}
	if old != nil {
		old.Close()
	}
}

// Addr is the listener's address, or nil when the server is not listening.
func (s *Server) Addr() net.Addr {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.rtsp == nil || s.rtsp.NetListener() == nil {
		return nil
	}
	return s.rtsp.NetListener().Addr()
}

// listen opens the TCP listener behind the allowed networks list.
func (s *Server) listen(network, address string) (net.Listener, error) {
	l, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return &allowListener{
		Listener: l,
		allowed:  s.deps.ClientAllowed,
		open:     make(chan struct{}, maxConns),
	}, nil
}

// maxConns bounds the RTSP connections open at once. The player limit counts
// sessions, which begin at SETUP; this one keeps connections that never get
// that far, each a goroutine and its buffers, from piling up on the board. A
// player opens one or two.
const maxConns = 8

// allowListener closes a connection from outside the allowed networks, or one
// past maxConns, before gortsplib reads a byte of it.
type allowListener struct {
	net.Listener
	allowed func(net.Addr) bool
	open    chan struct{}
}

func (l *allowListener) Accept() (net.Conn, error) {
	for {
		conn, err := l.Listener.Accept()
		if err != nil {
			return nil, err
		}
		if l.allowed != nil && !l.allowed(conn.RemoteAddr()) {
			log.Debugf("rtsp: refused %s, not an allowed network", conn.RemoteAddr())
			_ = conn.Close()
			continue
		}
		select {
		case l.open <- struct{}{}:
			return &countedConn{Conn: conn, open: l.open}, nil
		default:
			log.Debugf("rtsp: refused %s, %d connections open", conn.RemoteAddr(), maxConns)
			_ = conn.Close()
		}
	}
}

// countedConn gives its place under maxConns back when it closes.
type countedConn struct {
	net.Conn
	open chan struct{}
	once sync.Once
}

func (c *countedConn) Close() error {
	err := c.Conn.Close()
	c.once.Do(func() { <-c.open })
	return err
}

// credential is the token in a request's Basic authentication: the password,
// or the user name for a player that has only one field to fill in.
func credential(req *base.Request) string {
	var a headers.Authorization
	if err := a.Unmarshal(req.Header["Authorization"]); err != nil || a.Method != headers.AuthMethodBasic {
		return ""
	}
	if a.BasicPass != "" {
		return a.BasicPass
	}
	return a.Username
}

func (s *Server) authenticate(req *base.Request) (egress.Viewer, bool) {
	if s.deps.Authenticate == nil {
		return egress.Viewer{}, false
	}
	return s.deps.Authenticate(credential(req))
}

// unauthorized answers a request without a good credential. gortsplib adds
// the Basic challenge when the request had no credential, and closes the
// connection when it had a wrong one.
func unauthorized() (*base.Response, error) {
	return &base.Response{StatusCode: base.StatusUnauthorized}, liberrors.ErrServerAuth{}
}

// handler is the server's side of gortsplib's callbacks.
type handler struct {
	s *Server
}

func (h *handler) OnDescribe(ctx *gortsplib.ServerHandlerOnDescribeCtx) (*base.Response, *gortsplib.ServerStream, error) {
	if _, ok := h.s.authenticate(ctx.Request); !ok {
		res, err := unauthorized()
		return res, nil, err
	}

	f, err := h.s.ensureFeed()
	if err != nil {
		log.Debugf("rtsp: DESCRIBE from %s: %s", ctx.Conn.NetConn().RemoteAddr(), err)
		return &base.Response{StatusCode: base.StatusServiceUnavailable}, nil, nil
	}
	return &base.Response{StatusCode: base.StatusOK}, f.stream, nil
}

func (h *handler) OnSetup(ctx *gortsplib.ServerHandlerOnSetupCtx) (*base.Response, *gortsplib.ServerStream, error) {
	viewer, ok := h.s.authenticate(ctx.Request)
	if !ok {
		res, err := unauthorized()
		return res, nil, err
	}

	f, err := h.s.ensureFeed()
	if err != nil {
		return &base.Response{StatusCode: base.StatusServiceUnavailable}, nil, nil
	}

	client := ctx.Conn.NetConn().RemoteAddr().String()
	if !h.s.admit(ctx.Session, viewer, client) {
		log.Debugf("rtsp: refused %s, %d players already", client, egress.MaxClients)
		return &base.Response{StatusCode: base.StatusServiceUnavailable}, nil, nil
	}
	return &base.Response{StatusCode: base.StatusOK}, f.stream, nil
}

func (h *handler) OnPlay(ctx *gortsplib.ServerHandlerOnPlayCtx) (*base.Response, error) {
	if _, ok := h.s.authenticate(ctx.Request); !ok {
		return unauthorized()
	}
	h.s.played(ctx.Session)
	return &base.Response{StatusCode: base.StatusOK}, nil
}

func (h *handler) OnSessionClose(ctx *gortsplib.ServerHandlerOnSessionCloseCtx) {
	h.s.forget(ctx.Session, ctx.Error)
}

// admit records a session's first SETUP, and refuses it when the server
// already serves as many players as it may. A further SETUP of the same
// session, for its second track, is always let through.
func (s *Server) admit(ss *gortsplib.ServerSession, viewer egress.Viewer, client string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, known := s.sessions[ss]; known {
		return true
	}
	if len(s.sessions) >= egress.MaxClients {
		return false
	}
	s.sessions[ss] = &session{viewer: viewer, client: client, since: time.Now(), session: ss}
	return true
}

func (s *Server) played(ss *gortsplib.ServerSession) {
	s.mu.Lock()
	sess, ok := s.sessions[ss]
	first := ok && !sess.playing
	if first {
		sess.playing = true
	}
	s.mu.Unlock()

	if first {
		record(sess, "rtsp-start", "")
	}
}

func (s *Server) forget(ss *gortsplib.ServerSession, reason error) {
	s.mu.Lock()
	sess, ok := s.sessions[ss]
	delete(s.sessions, ss)
	if len(s.sessions) == 0 {
		s.idleSince = time.Now()
	}
	s.mu.Unlock()

	if ok && sess.playing {
		detail := "after " + time.Since(sess.since).Round(time.Second).String()
		if reason != nil {
			detail = reason.Error() + " " + detail
		}
		record(sess, "rtsp-end", detail)
	}
}

// record writes a player's start or end to the audit log.
func record(sess *session, action, detail string) {
	host, _, err := net.SplitHostPort(sess.client)
	if err != nil {
		host = sess.client
	}
	audit.Record(audit.Event{
		Actor:    sess.viewer.Username,
		Via:      sess.viewer.Via,
		Key:      sess.viewer.Key,
		Role:     sess.viewer.Role,
		IP:       host,
		Category: audit.CategoryAuth,
		Action:   action,
		Target:   "rtsp",
		Outcome:  audit.OutcomeOK,
		Detail:   detail,
	})
}

// watch ends a player whose credential no longer holds, until stop closes.
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
		sessions := make([]*session, 0, len(s.sessions))
		for _, sess := range s.sessions {
			sessions = append(sessions, sess)
		}
		s.mu.Unlock()

		// Valid reads the key store, so it runs outside the lock.
		for _, sess := range sessions {
			if sess.viewer.Valid != nil && !sess.viewer.Valid() {
				log.Infof("rtsp: ended %s, the credential was revoked or expired", sess.client)
				sess.session.Close()
			}
		}
	}
}

// feed is the stream every player shares, and the tap reader behind it.
type feed struct {
	stream *gortsplib.ServerStream
	reader *egress.Reader
	codec  uint8
	video  *description.Media
	audio  *description.Media
	once   sync.Once
}

// close ends the stream, and with it every player on it.
func (f *feed) close() {
	f.once.Do(func() {
		f.reader.Close()
		f.stream.Close()
	})
}

// ensureFeed returns the running stream, starting one if there is none. A new
// stream is described from its first keyframe, so the parameter sets a
// player needs are in the SDP from the start.
func (s *Server) ensureFeed() (*feed, error) {
	s.feedMu.Lock()
	defer s.feedMu.Unlock()

	s.mu.Lock()
	server, current := s.rtsp, s.feed
	s.mu.Unlock()
	if server == nil {
		return nil, errNotRunning
	}
	if current != nil {
		return current, nil
	}

	withAudio := s.deps.AudioAvailable()
	reader := s.deps.Tap.Subscribe("rtsp", withAudio)

	timer := time.NewTimer(s.deps.FirstPicture)
	defer timer.Stop()

	var first egress.Unit
	select {
	case unit, ok := <-reader.Units():
		if !ok {
			reader.Close()
			return nil, errNoPicture
		}
		first = unit
	case <-timer.C:
		reader.Close()
		return nil, errNoPicture
	}

	f := &feed{reader: reader, codec: first.Codec}
	f.video = &description.Media{
		Type:    description.MediaTypeVideo,
		Formats: []format.Format{videoFormat(first)},
	}
	medias := []*description.Media{f.video}
	if withAudio {
		f.audio = &description.Media{
			Type: description.MediaTypeAudio,
			Formats: []format.Format{&format.Opus{
				PayloadTyp:   audioPayloadType,
				ChannelCount: audio.HostFormat.Channels,
			}},
		}
		medias = append(medias, f.audio)
	}

	f.stream = &gortsplib.ServerStream{Server: server, Desc: &description.Session{Medias: medias}}
	if err := f.stream.Initialize(); err != nil {
		reader.Close()
		return nil, err
	}

	s.mu.Lock()
	s.feed, s.idleSince = f, time.Now()
	s.mu.Unlock()

	go s.run(f, first)
	return f, nil
}

// videoFormat describes the codec of a keyframe, with its parameter sets.
func videoFormat(first egress.Unit) format.Format {
	// The parameter sets point into the capture loop's frame, which the
	// description outlives.
	vps, sps, pps := egress.ParameterSets(first.Codec, first.NALUs)
	vps, sps, pps = cloneOrNil(vps), cloneOrNil(sps), cloneOrNil(pps)
	if first.Codec == common.CodecH265 {
		return &format.H265{PayloadTyp: videoPayloadType, VPS: vps, SPS: sps, PPS: pps}
	}
	return &format.H264{PayloadTyp: videoPayloadType, SPS: sps, PPS: pps, PacketizationMode: 1}
}

// auEncoder is rtph264.Encoder or rtph265.Encoder.
type auEncoder interface {
	Encode(au [][]byte) ([]*rtp.Packet, error)
}

// rtpTime is a PTS on an RTP clock.
func rtpTime(pts time.Duration, clockRate int64) uint32 {
	return uint32(pts.Microseconds() * clockRate / 1_000_000)
}

// run writes the tap's units to the stream until the feed ends: no player
// for a while, a codec change, which the description cannot follow, or the
// server stopping.
func (s *Server) run(f *feed, first egress.Unit) {
	defer s.endFeed(f)

	var video auEncoder
	var err error
	switch forma := f.video.Formats[0].(type) {
	case *format.H264:
		video, err = forma.CreateEncoder()
	case *format.H265:
		video, err = forma.CreateEncoder()
	}
	if err != nil {
		log.Errorf("rtsp: video packetizer: %s", err)
		return
	}

	var opus interface {
		Encode(frame []byte) (*rtp.Packet, error)
	}
	if f.audio != nil {
		enc, err := f.audio.Formats[0].(*format.Opus).CreateEncoder()
		if err != nil {
			log.Errorf("rtsp: audio packetizer: %s", err)
			return
		}
		opus = enc
	}

	writeVideo := func(unit egress.Unit) {
		if unit.KeyFrame {
			s.refreshParams(f, unit)
		}
		packets, err := video.Encode(unit.NALUs)
		if err != nil {
			log.Debugf("rtsp: packetize: %s", err)
			return
		}
		ts := rtpTime(unit.PTS, videoClockRate)
		for _, packet := range packets {
			packet.Timestamp = ts
			// The error is a closed stream, and the reader closing ends
			// this loop.
			_ = f.stream.WritePacketRTPWithNTP(f.video, packet, unit.Wall)
		}
	}

	writeVideo(first)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case unit, ok := <-f.reader.Units():
			if !ok {
				return
			}
			if unit.Audio {
				if opus == nil {
					continue
				}
				packet, err := opus.Encode(unit.Opus)
				if err != nil {
					continue
				}
				packet.Timestamp = rtpTime(unit.PTS, int64(audio.SampleRate))
				_ = f.stream.WritePacketRTPWithNTP(f.audio, packet, unit.Wall)
				continue
			}
			if unit.Codec != f.codec {
				// The players were told another codec in the SDP. They are
				// let go, and the next DESCRIBE describes the new one.
				log.Infof("rtsp: the codec changed; ending the stream")
				return
			}
			writeVideo(unit)

		case <-ticker.C:
			if s.idleFor(f) > s.deps.Linger {
				log.Debugf("rtsp: nobody watching; ending the stream")
				return
			}
		}
	}
}

// refreshParams puts a keyframe's parameter sets in the description when they
// changed, as they do with the resolution, so the next DESCRIBE has them.
func (s *Server) refreshParams(f *feed, unit egress.Unit) {
	vps, sps, pps := egress.ParameterSets(unit.Codec, unit.NALUs)
	if sps == nil || pps == nil {
		return
	}
	switch forma := f.video.Formats[0].(type) {
	case *format.H264:
		oldSPS, oldPPS := forma.SafeParams()
		if string(oldSPS) == string(sps) && string(oldPPS) == string(pps) {
			return
		}
		forma.SafeSetParams(clone(sps), clone(pps))
	case *format.H265:
		oldVPS, oldSPS, oldPPS := forma.SafeParams()
		if string(oldVPS) == string(vps) && string(oldSPS) == string(sps) && string(oldPPS) == string(pps) {
			return
		}
		forma.SafeSetParams(clone(vps), clone(sps), clone(pps))
	default:
		return
	}
	f.stream.ReloadDesc()
}

func clone(b []byte) []byte {
	return append([]byte(nil), b...)
}

// cloneOrNil is clone, keeping a missing parameter set missing.
func cloneOrNil(b []byte) []byte {
	if b == nil {
		return nil
	}
	return clone(b)
}

// idleFor is how long the feed has had no player, or 0 while it has one.
func (s *Server) idleFor(f *feed) time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.feed != f || len(s.sessions) > 0 {
		return 0
	}
	return time.Since(s.idleSince)
}

// endFeed forgets the feed and closes it.
func (s *Server) endFeed(f *feed) {
	s.mu.Lock()
	if s.feed == f {
		s.feed = nil
	}
	s.mu.Unlock()

	f.close()
}
