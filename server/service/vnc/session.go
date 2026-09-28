package vnc

import (
	"bufio"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

// handshakeTimeout bounds everything from the connection to ServerInit.
const handshakeTimeout = 30 * time.Second

// writeTimeout bounds one update. A client that stops reading for this long is
// gone.
const writeTimeout = 10 * time.Second

// Fallback framebuffer size, for a capture that reports none.
const (
	fallbackWidth  = 1920
	fallbackHeight = 1080
)

var (
	errServerStopped = errors.New("the VNC server stopped")
	errDisconnected  = errors.New("the session was ended from the settings page")
	errSessionActive = errors.New("another VNC session is active")
	errNoTight       = errors.New("the client does not support Tight encoding, which the server needs to send JPEG frames")
	errPixelFormat   = errors.New("the client asked for a pixel format that cannot take JPEG frames")
	errNoResize      = errors.New("the host resolution changed and the client does not support DesktopSize")
	errStreamEnded   = errors.New("the video stream ended")
)

// sessionEndReason reports whether an error that ended a session is worth
// showing on the settings page. A client that simply went away is not.
func sessionEndReason(err error) error {
	for _, shown := range []error{errNoTight, errPixelFormat, errNoResize, errStreamEnded} {
		if errors.Is(err, shown) {
			return err
		}
	}
	return nil
}

// session is one client connection, from the handshake to the end.
type session struct {
	srv    *Server
	conn   net.Conn
	reader *bufio.Reader
	ip     string
	remote string
	minor  int

	auth  authResult
	since time.Time

	// done closes when the session ends, whoever ends it.
	done    chan struct{}
	endOnce sync.Once
	endErr  error

	// mu guards what the reader learns and the writer uses.
	mu            sync.Mutex
	pixelFormat   pixelFormat
	tight         bool
	desktopSize   bool
	width, height int
	pending       bool
	incremental   bool

	// notify wakes the writer when a request arrives.
	notify chan struct{}

	framesSent atomic.Uint64
	bytesSent  atomic.Uint64
}

func newSession(srv *Server, conn net.Conn) *session {
	remote := conn.RemoteAddr().String()
	ip := remote
	if host, _, err := net.SplitHostPort(remote); err == nil {
		ip = host
	}
	return &session{
		srv:         srv,
		conn:        conn,
		reader:      bufio.NewReader(conn),
		ip:          ip,
		remote:      remote,
		done:        make(chan struct{}),
		notify:      make(chan struct{}, 1),
		pixelFormat: serverPixelFormat,
	}
}

// end closes the session. The first reason wins.
func (s *session) end(reason error) {
	s.endOnce.Do(func() {
		s.endErr = reason
		close(s.done)
		_ = s.conn.Close()
	})
}

func (s *session) info() SessionInfo {
	s.mu.Lock()
	width, height := s.width, s.height
	s.mu.Unlock()

	return SessionInfo{
		Client:     s.remote,
		User:       s.auth.user,
		Method:     s.auth.method,
		Since:      s.since.UTC().Format(time.RFC3339),
		Width:      width,
		Height:     height,
		FramesSent: s.framesSent.Load(),
		BytesSent:  s.bytesSent.Load(),
	}
}

// serve runs the handshake, then the session, and cleans up after both.
func (s *session) serve() {
	claimed, err := s.handshake()
	if err != nil {
		log.Infof("vnc: %s: %s", s.remote, err)
		_ = s.conn.Close()
		if claimed {
			s.srv.release(s, nil)
		}
		return
	}

	log.Infof("vnc: session from %s opened (%s %s)", s.remote, s.auth.method, s.auth.user)
	err = s.run()
	s.end(err)

	reason := s.endErr
	if reason != errDisconnected && reason != errServerStopped {
		reason = sessionEndReason(reason)
	}
	s.srv.release(s, reason)
	log.Infof("vnc: session from %s closed: %v", s.remote, s.endErr)
}

// handshake runs everything up to ServerInit. It reports whether the session
// was made the open one, which the caller must undo if it then fails.
func (s *session) handshake() (claimed bool, err error) {
	// The slot is held for the handshake only, not for the session.
	defer func() { <-s.srv.handshakes }()

	_ = s.conn.SetDeadline(time.Now().Add(handshakeTimeout))

	if _, err := io.WriteString(s.conn, protocolVersion); err != nil {
		return false, err
	}
	version := make([]byte, 12)
	if _, err := io.ReadFull(s.reader, version); err != nil {
		return false, err
	}
	if s.minor, err = parseClientVersion(version); err != nil {
		return false, err
	}

	types := []byte{securityVeNCrypt}
	settings := s.srv.deps.Settings()
	if settings.VNCAuth {
		if _, ok, _ := s.srv.deps.Password.Get(); ok {
			types = append(types, securityVNCAuth)
		}
	}
	if _, err := s.conn.Write(append([]byte{byte(len(types))}, types...)); err != nil {
		return false, err
	}
	chosen, err := s.reader.ReadByte()
	if err != nil {
		return false, err
	}

	switch {
	case chosen == securityVeNCrypt:
		s.auth, err = s.vencrypt()
	case chosen == securityVNCAuth && len(types) > 1:
		s.auth, err = s.vncAuth()
	default:
		err = fmt.Errorf("the client chose security type %d, which was not offered", chosen)
	}
	if err != nil {
		var silent silentError
		if !errors.As(err, &silent) {
			_ = writeSecurityResult(s.conn, s.minor, err.Error())
		}
		return false, fmt.Errorf("authentication: %w", err)
	}

	// Set before the claim, because from the claim on the settings page
	// reads the session.
	s.since = time.Now()

	// The second client is refused only once it has proved who it is, so a
	// stranger learns nothing about the open session.
	if !s.srv.claim(s) {
		_ = writeSecurityResult(s.conn, s.minor, errSessionActive.Error())
		return false, errSessionActive
	}
	if err := writeSecurityResult(s.conn, s.minor, ""); err != nil {
		return true, err
	}

	// ClientInit carries the shared flag. There is one session at a time
	// whatever it says.
	if _, err := s.reader.ReadByte(); err != nil {
		return true, err
	}

	width, height := s.srv.deps.ScreenSize()
	if width <= 0 || height <= 0 {
		width, height = fallbackWidth, fallbackHeight
	}
	s.mu.Lock()
	s.width, s.height = width, height
	s.mu.Unlock()
	if _, err := s.conn.Write(serverInit(width, height, s.srv.deps.Name())); err != nil {
		return true, err
	}

	_ = s.conn.SetDeadline(time.Time{})
	return true, nil
}

// run serves the session until the client leaves or something ends it. The
// reader runs here, and the writer in its own goroutine.
func (s *session) run() error {
	frames := s.srv.deps.Subscribe()
	defer frames.Close()

	input := s.srv.deps.NewInput()
	defer input.Close()

	writerDone := make(chan error, 1)
	go func() {
		err := s.write(frames)
		s.end(err)
		writerDone <- err
	}()

	err := s.read(input)
	s.end(err)
	if werr := <-writerDone; werr != nil && sessionEndReason(werr) != nil {
		return werr
	}
	return err
}

// read handles the client's messages until the connection fails.
func (s *session) read(input Input) error {
	keys := newKeyboard()
	var mouse pointer
	buf := make([]byte, 20)

	for {
		msgType, err := s.reader.ReadByte()
		if err != nil {
			return err
		}

		switch msgType {
		case msgSetPixelFormat:
			if _, err := io.ReadFull(s.reader, buf[:19]); err != nil {
				return err
			}
			format := parsePixelFormat(buf[3:19])
			s.mu.Lock()
			s.pixelFormat = format
			s.mu.Unlock()

		case msgSetEncodings:
			if _, err := io.ReadFull(s.reader, buf[:3]); err != nil {
				return err
			}
			count := int(binary.BigEndian.Uint16(buf[1:3]))
			if count > maxEncodings {
				return fmt.Errorf("the client listed %d encodings", count)
			}
			tight, desktopSize := false, false
			for range count {
				if _, err := io.ReadFull(s.reader, buf[:4]); err != nil {
					return err
				}
				switch int32(binary.BigEndian.Uint32(buf[:4])) {
				case encodingTight:
					tight = true
				case encodingDesktopSize:
					desktopSize = true
				}
			}
			s.mu.Lock()
			s.tight, s.desktopSize = tight, desktopSize
			s.mu.Unlock()

		case msgFramebufferUpdateRequest:
			if _, err := io.ReadFull(s.reader, buf[:9]); err != nil {
				return err
			}
			s.request(buf[0] != 0)

		case msgKeyEvent:
			if _, err := io.ReadFull(s.reader, buf[:7]); err != nil {
				return err
			}
			if report := keys.event(buf[0] != 0, binary.BigEndian.Uint32(buf[3:7])); report != nil {
				input.Keyboard(report)
			}

		case msgPointerEvent:
			if _, err := io.ReadFull(s.reader, buf[:5]); err != nil {
				return err
			}
			x := int(binary.BigEndian.Uint16(buf[1:3]))
			y := int(binary.BigEndian.Uint16(buf[3:5]))
			s.mu.Lock()
			width, height := s.width, s.height
			s.mu.Unlock()
			for _, report := range mouse.event(buf[0], x, y, width, height) {
				input.Pointer(report)
			}

		case msgClientCutText:
			// The clipboard is not supported. The text is read and dropped.
			if _, err := io.ReadFull(s.reader, buf[:7]); err != nil {
				return err
			}
			length := int64(int32(binary.BigEndian.Uint32(buf[3:7])))
			if length < 0 {
				// The extended clipboard sends a negative length.
				length = -length
			}
			if length > maxCutText {
				return fmt.Errorf("the client sent %d bytes of clipboard text", length)
			}
			if _, err := io.CopyN(io.Discard, s.reader, length); err != nil {
				return err
			}

		default:
			return fmt.Errorf("unknown client message type %d", msgType)
		}
	}
}

// request records a FramebufferUpdateRequest. Requests that arrive before the
// writer answers merge into one, which is incremental only if all of them were.
func (s *session) request(incremental bool) {
	s.mu.Lock()
	if s.pending {
		s.incremental = s.incremental && incremental
	} else {
		s.pending = true
		s.incremental = incremental
	}
	s.mu.Unlock()

	select {
	case s.notify <- struct{}{}:
	default:
	}
}

// writeState is the writer's view of the session when it answers a request.
type writeState struct {
	pending     bool
	incremental bool
	tight       bool
	jpeg        bool
	desktopSize bool
}

func (s *session) snapshot() writeState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return writeState{
		pending:     s.pending,
		incremental: s.incremental,
		tight:       s.tight,
		jpeg:        s.pixelFormat.jpegCapable(),
		desktopSize: s.desktopSize,
	}
}

func (s *session) clearRequest() {
	s.mu.Lock()
	s.pending = false
	s.mu.Unlock()
}

// write answers update requests with frames from the stream.
//
// It takes a frame only while a request is outstanding and not more often than
// the frame rate limit allows. A frame it does not take stays in the
// subscription's slot, and the capture loop reads nothing while every viewer
// holds one, so the capture rate follows the client.
func (s *session) write(frames FrameSubscription) error {
	minInterval := time.Second / time.Duration(max(1, s.srv.deps.Settings().MaxFPS))

	s.mu.Lock()
	width, height := s.width, s.height
	s.mu.Unlock()

	var lastSent time.Time
	// last is the frame the client has, for a request that asks for the
	// whole picture again. held is a frame taken from the stream and not
	// sent yet, because a resize had to go first.
	var last, held []byte

	for {
		state := s.snapshot()
		if !state.pending {
			select {
			case <-s.notify:
				continue
			case <-s.done:
				return nil
			}
		}
		if !state.tight {
			return errNoTight
		}
		if !state.jpeg {
			return errPixelFormat
		}

		if wait := minInterval - time.Since(lastSent); wait > 0 {
			timer := time.NewTimer(wait)
			select {
			case <-timer.C:
			case <-s.done:
				timer.Stop()
				return nil
			}
		}

		frame := held
		held = nil
		if frame == nil && !state.incremental {
			frame = last
		}
		if frame == nil {
			// A client that asks for the whole picture before it has one
			// must not wait for the host to change something.
			if !state.incremental {
				frames.Refresh()
			}
			select {
			case next, ok := <-frames.Frames():
				if !ok {
					return errStreamEnded
				}
				frame = next
			case <-s.notify:
				// A newer request may ask for the whole picture, which
				// the last frame answers at once.
				continue
			case <-s.done:
				return nil
			}
		}

		frameWidth, frameHeight, err := jpegSize(frame)
		if err != nil {
			log.Debugf("vnc: dropped a frame: %s", err)
			continue
		}

		if frameWidth != width || frameHeight != height {
			if !state.desktopSize {
				return errNoResize
			}
			// The resize is an update of its own. The frame waits for the
			// client's next request, which is for the new size.
			s.clearRequest()
			if err := s.send(desktopSizeUpdate(frameWidth, frameHeight)); err != nil {
				return err
			}
			width, height = frameWidth, frameHeight
			s.mu.Lock()
			s.width, s.height = width, height
			s.mu.Unlock()
			held, last = frame, nil
			continue
		}

		s.clearRequest()
		_ = s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
		n, err := writeJPEGUpdate(s.conn, width, height, frame)
		if err != nil {
			return err
		}
		s.framesSent.Add(1)
		s.bytesSent.Add(uint64(n))
		last = frame
		lastSent = time.Now()
	}
}

func (s *session) send(data []byte) error {
	_ = s.conn.SetWriteDeadline(time.Now().Add(writeTimeout))
	n, err := s.conn.Write(data)
	s.bytesSent.Add(uint64(n))
	return err
}
