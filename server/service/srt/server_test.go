package srt

import (
	"bufio"
	"bytes"
	"errors"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/mpegts/codecs"
	gosrt "github.com/datarhei/gosrt"

	"NanoKVM-Server/config"
	"NanoKVM-Server/service/stream/egress"
	"NanoKVM-Server/service/stream/egress/egresstest"
)

const (
	goodToken  = "nkvm_view-only"
	passphrase = "correct horse battery"
)

var opusPacket = []byte{0xfc, 0xff, 0xfe}

type memoryPassphrase struct {
	mu    sync.Mutex
	value string
}

func (m *memoryPassphrase) Get() (string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.value, m.value != "", nil
}

func (m *memoryPassphrase) Set(p string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.value = p
	return nil
}

func freeUDPPort(t *testing.T) int {
	t.Helper()
	c, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	return c.LocalAddr().(*net.UDPAddr).Port
}

type harness struct {
	server  *Server
	port    int
	allowed atomic.Bool
	valid   atomic.Bool
}

func start(t *testing.T, withAudio bool) *harness {
	t.Helper()
	tap, feed := egresstest.New()
	h := &harness{port: freeUDPPort(t)}
	h.allowed.Store(true)
	h.valid.Store(true)

	settings := config.SRT{Enabled: true, Port: h.port, LatencyMs: config.DefaultSRTLatencyMs}
	h.server = New(Deps{
		Settings:      func() config.SRT { return settings },
		Passphrase:    &memoryPassphrase{value: passphrase},
		ClientAllowed: func(net.Addr) bool { return h.allowed.Load() },
		Authenticate: func(token string) (egress.Viewer, bool) {
			if token != goodToken {
				return egress.Viewer{}, false
			}
			return egress.Viewer{Username: "watcher", Role: "viewer", Via: "apikey", Valid: h.valid.Load}, true
		},
		Tap:            tap,
		AudioAvailable: func() bool { return withAudio },
		Recheck:        50 * time.Millisecond,
	})
	if err := h.server.Apply(); err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	feed.Pump(stop, 30, 10)
	if withAudio {
		feed.PumpAudio(stop, opusPacket)
	}
	t.Cleanup(func() {
		close(stop)
		h.server.Stop()
	})
	return h
}

func (h *harness) dial(streamID, pass string) (gosrt.Conn, error) {
	cfg := gosrt.DefaultConfig()
	cfg.StreamId = streamID
	cfg.Passphrase = pass
	cfg.ConnectionTimeout = 2 * time.Second
	return gosrt.Dial("srt", "127.0.0.1:"+strconv.Itoa(h.port), cfg)
}

// An SRT caller gets MPEG-TS with the injected H.264 and the Opus beside it.
func TestACallerReceivesTheInjectedFrames(t *testing.T) {
	h := start(t, true)

	conn, err := h.dial(goodToken, passphrase)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	type result struct {
		key, sound bool
		err        error
	}
	done := make(chan result, 1)
	go func() {
		var res result
		r := &mpegts.Reader{R: bufio.NewReader(conn)}
		if err := r.Initialize(); err != nil {
			done <- result{err: err}
			return
		}
		finished := errors.New("finished")
		for _, track := range r.Tracks() {
			switch track.Codec.(type) {
			case *codecs.H264:
				r.OnDataH264(track, func(_, _ int64, au [][]byte) error {
					for _, nalu := range au {
						if len(nalu) > 0 && nalu[0]&0x1f == 5 {
							res.key = true
						}
					}
					if res.key && res.sound {
						return finished
					}
					return nil
				})
			case *codecs.Opus:
				r.OnDataOpus(track, func(_ int64, packets [][]byte) error {
					for _, p := range packets {
						if bytes.Equal(p, opusPacket) {
							res.sound = true
						}
					}
					if res.key && res.sound {
						return finished
					}
					return nil
				})
			}
		}
		for {
			if err := r.Read(); err != nil {
				if !errors.Is(err, finished) {
					res.err = err
				}
				done <- res
				return
			}
		}
	}()

	select {
	case res := <-done:
		if res.err != nil || !res.key || !res.sound {
			t.Fatalf("the caller got an IDR picture %t and Opus %t, error %v", res.key, res.sound, res.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the caller never got both an IDR picture and an Opus packet")
	}

	state := h.server.State()
	if len(state.Clients) != 1 || state.Clients[0].User != "watcher" || state.Clients[0].Sent == 0 {
		t.Fatalf("state lists %+v, want one caller that was sent packets", state.Clients)
	}
}

func TestACallerWithoutThePassphraseOrATokenIsRefused(t *testing.T) {
	h := start(t, false)

	cases := []struct{ name, streamID, pass string }{
		{"no passphrase", goodToken, ""},
		{"a wrong passphrase", goodToken, "not the passphrase"},
		{"no token", "", passphrase},
		{"a wrong token", "nkvm_wrong", passphrase},
	}
	for _, tc := range cases {
		if conn, err := h.dial(tc.streamID, tc.pass); err == nil {
			conn.Close()
			t.Fatalf("a caller with %s was let in", tc.name)
		}
	}
}

func TestACallerOutsideTheAllowedNetworksIsRefused(t *testing.T) {
	h := start(t, false)
	h.allowed.Store(false)

	if conn, err := h.dial(goodToken, passphrase); err == nil {
		conn.Close()
		t.Fatal("a caller from outside the allowed networks was let in")
	}
}

func TestAtMostTwoCallersAreServed(t *testing.T) {
	h := start(t, false)

	for i := range egress.MaxClients {
		conn, err := h.dial(goodToken, passphrase)
		if err != nil {
			t.Fatalf("caller %d was refused: %v", i, err)
		}
		defer conn.Close()
	}
	if conn, err := h.dial(goodToken, passphrase); err == nil {
		conn.Close()
		t.Fatalf("a caller was served past the limit of %d", egress.MaxClients)
	}
}

func TestARevokedCallerIsEnded(t *testing.T) {
	h := start(t, false)

	conn, err := h.dial(goodToken, passphrase)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	deadline := time.Now().Add(5 * time.Second)
	for h.server.Clients() == 0 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	h.valid.Store(false)
	for h.server.Clients() > 0 {
		if time.Now().After(deadline) {
			t.Fatal("the revoked caller was not ended")
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func TestTheListenerNeedsAPassphrase(t *testing.T) {
	tap, _ := egresstest.New()
	s := New(Deps{
		Settings:   func() config.SRT { return config.SRT{Enabled: true, Port: freeUDPPort(t), LatencyMs: 120} },
		Passphrase: &memoryPassphrase{},
		Tap:        tap,
	})
	if err := s.Apply(); err == nil {
		s.Stop()
		t.Fatal("the listener opened without a passphrase")
	}
	if s.Addr() != nil || s.State().Error == "" {
		t.Fatal("the state does not say why SRT is not listening")
	}
}

func TestTheStreamIDTakesTheTokenInEitherForm(t *testing.T) {
	for id, want := range map[string]string{
		"nkvm_abc":                    "nkvm_abc",
		" nkvm_abc ":                  "nkvm_abc",
		"#!::s=nkvm_abc,m=request":    "nkvm_abc",
		"#!::m=request,u=nkvm_abc":    "nkvm_abc",
		"#!::u=someone,s=nkvm_abc":    "nkvm_abc",
		"#!::m=request,r=live/stream": "",
	} {
		if got := streamToken(id); got != want {
			t.Errorf("streamToken(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestPassphraseBounds(t *testing.T) {
	for p, ok := range map[string]bool{
		"short":                 false,
		"tenletters":            true,
		strings.Repeat("a", 79): true,
		strings.Repeat("a", 80): false,
		"with a tab\tinside":    false,
	} {
		if err := ValidatePassphrase(p); (err == nil) != ok {
			t.Errorf("ValidatePassphrase(%q) = %v", p, err)
		}
	}
}
