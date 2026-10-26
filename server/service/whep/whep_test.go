package whep

import (
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pion/webrtc/v4"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/config"
	"NanoKVM-Server/service/apikey"
	"NanoKVM-Server/service/audit"
	stream "NanoKVM-Server/service/stream/webrtc"
	"NanoKVM-Server/service/vm"
)

// fixture is a WHEP endpoint over the real session code, with an account
// store, a key file and an audit log of its own.
type fixture struct {
	t       *testing.T
	server  *httptest.Server
	service *Service
	enabled atomic.Bool
	viewKey string
	fullKey string
	audit   *audit.Store
}

func newFixture(t *testing.T, max int) *fixture {
	t.Helper()
	gin.SetMode(gin.TestMode)

	store := authn.NewStore(filepath.Join(t.TempDir(), "pwd"))
	if _, ok, err := store.Authenticate("admin", "admin"); err != nil || !ok {
		t.Fatalf("default login: ok=%v err=%v", ok, err)
	}
	originalStore, originalShares := authn.DefaultStore, authn.DefaultShares
	authn.DefaultStore = store
	authn.DefaultShares = authn.NewShareStore(filepath.Join(t.TempDir(), "share_links.json"))
	originalKeys := apikey.File
	apikey.File = filepath.Join(t.TempDir(), "api_keys.json")
	conf := config.GetInstance()
	originalAuthentication := conf.Authentication
	conf.Authentication = "enable"
	// No STUN: the test must not wait on the internet to gather.
	originalStun := stream.CurrentStun()
	stream.SetStun(config.StunDisabled)
	auditStore := audit.NewStore(t.TempDir())
	restoreAudit := audit.UseStore(auditStore)
	t.Cleanup(func() {
		authn.DefaultStore, authn.DefaultShares = originalStore, originalShares
		apikey.File = originalKeys
		conf.Authentication = originalAuthentication
		stream.SetStun(originalStun)
		restoreAudit()
	})

	viewKey, _, err := apikey.CreateWithScope("obs", "admin", apikey.ScopeView)
	if err != nil {
		t.Fatal(err)
	}
	fullKey, _, err := apikey.Create("script", "admin")
	if err != nil {
		t.Fatal(err)
	}

	f := &fixture{t: t, viewKey: viewKey, fullKey: fullKey, audit: auditStore}
	f.enabled.Store(true)
	f.service = New(Deps{
		Enabled:    f.enabled.Load,
		SetEnabled: func(on bool) error { f.enabled.Store(on); return nil },
		Max:        max,
		Recheck:    50 * time.Millisecond,
	})
	r := gin.New()
	f.service.Register(r)
	f.server = httptest.NewServer(r)
	t.Cleanup(func() {
		f.service.CloseAll("test over")
		f.server.Close()
	})

	return f
}

// player is a WHEP client: a pion peer that only receives.
type player struct {
	pc        *webrtc.PeerConnection
	connected chan struct{}
	video     chan struct{}
}

func newPlayer(t *testing.T) *player {
	t.Helper()

	pc, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = pc.Close() })

	p := &player{pc: pc, connected: make(chan struct{}), video: make(chan struct{})}
	var connectedOnce, videoOnce atomic.Bool
	pc.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		if state == webrtc.ICEConnectionStateConnected && connectedOnce.CompareAndSwap(false, true) {
			close(p.connected)
		}
	})
	pc.OnTrack(func(track *webrtc.TrackRemote, _ *webrtc.RTPReceiver) {
		if track.Kind() == webrtc.RTPCodecTypeVideo && videoOnce.CompareAndSwap(false, true) {
			close(p.video)
		}
	})
	for _, kind := range []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio} {
		if _, err := pc.AddTransceiverFromKind(kind, webrtc.RTPTransceiverInit{Direction: webrtc.RTPTransceiverDirectionRecvonly}); err != nil {
			t.Fatal(err)
		}
	}

	return p
}

// offer is the player's offer with its candidates gathered, the
// non-trickle way.
func (p *player) offer(t *testing.T) string {
	t.Helper()

	offer, err := p.pc.CreateOffer(nil)
	if err != nil {
		t.Fatal(err)
	}
	gathered := webrtc.GatheringCompletePromise(p.pc)
	if err := p.pc.SetLocalDescription(offer); err != nil {
		t.Fatal(err)
	}
	<-gathered

	return p.pc.LocalDescription().SDP
}

func (f *fixture) do(method, path, token, contentType, body string) *http.Response {
	f.t.Helper()

	req, err := http.NewRequest(method, f.server.URL+path, strings.NewReader(body))
	if err != nil {
		f.t.Fatal(err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		f.t.Fatal(err)
	}
	f.t.Cleanup(func() { _ = rsp.Body.Close() })

	return rsp
}

// start posts a player's offer and applies the answer.
func (f *fixture) start(p *player) string {
	f.t.Helper()

	rsp := f.do(http.MethodPost, "/api/whep", f.viewKey, "application/sdp", p.offer(f.t))
	body, _ := io.ReadAll(rsp.Body)
	if rsp.StatusCode != http.StatusCreated {
		f.t.Fatalf("POST: %d %s", rsp.StatusCode, body)
	}
	if got := rsp.Header.Get("Content-Type"); !strings.HasPrefix(got, "application/sdp") {
		f.t.Fatalf("answer content type %q", got)
	}
	location := rsp.Header.Get("Location")
	if !strings.HasPrefix(location, "/api/whep/") || len(location) <= len("/api/whep/") {
		f.t.Fatalf("Location %q", location)
	}
	if err := p.pc.SetRemoteDescription(webrtc.SessionDescription{Type: webrtc.SDPTypeAnswer, SDP: string(body)}); err != nil {
		f.t.Fatalf("apply the answer: %s", err)
	}

	return location
}

func webrtcViewers() int {
	return vm.HdmiViewerCounts([]string{"webrtc"})[0]
}

func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()

	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for %s", what)
}

func (f *fixture) auditActions() []string {
	f.audit.Flush()
	page, err := f.audit.Query(audit.Query{Limit: 100})
	if err != nil {
		f.t.Fatal(err)
	}
	actions := make([]string, 0, len(page.Events))
	for _, e := range page.Events {
		actions = append(actions, e.Action+":"+e.Outcome)
	}
	return actions
}

func contains(list []string, want string) bool {
	for _, item := range list {
		if item == want {
			return true
		}
	}
	return false
}

// The whole exchange: offer, answer, Location; the player connects and counts
// as a WebRTC viewer; DELETE on the Location ends it and lets the viewer go.
func TestOfferAnswerConnectAndDelete(t *testing.T) {
	f := newFixture(t, 4)
	p := newPlayer(t)

	location := f.start(p)

	select {
	case <-p.connected:
	case <-time.After(15 * time.Second):
		t.Fatal("the player never connected")
	}
	waitFor(t, "the WHEP session to count as a viewer", func() bool { return webrtcViewers() == 1 })
	if got := f.service.Count(); got != 1 {
		t.Fatalf("%d sessions, want 1", got)
	}

	rsp := f.do(http.MethodDelete, location, f.viewKey, "", "")
	if rsp.StatusCode != http.StatusOK {
		t.Fatalf("DELETE: %d", rsp.StatusCode)
	}
	waitFor(t, "the viewer to be let go", func() bool { return webrtcViewers() == 0 })
	if got := f.service.Count(); got != 0 {
		t.Fatalf("%d sessions after DELETE, want 0", got)
	}

	if rsp := f.do(http.MethodDelete, location, f.viewKey, "", ""); rsp.StatusCode != http.StatusNotFound {
		t.Fatalf("second DELETE: %d, want 404", rsp.StatusCode)
	}

	waitFor(t, "the session's start and end in the audit log", func() bool {
		actions := f.auditActions()
		return contains(actions, "whep-start:ok") && contains(actions, "whep-end:ok")
	})
}

// An offer with a data channel is refused: WHEP is for watching, and nothing
// on a connection a view-only token opened may reach the keyboard or mouse.
func TestOfferWithADataChannelIsRefused(t *testing.T) {
	f := newFixture(t, 4)
	p := newPlayer(t)
	if _, err := p.pc.CreateDataChannel("hid", nil); err != nil {
		t.Fatal(err)
	}

	rsp := f.do(http.MethodPost, "/api/whep", f.viewKey, "application/sdp", p.offer(t))
	if rsp.StatusCode != http.StatusBadRequest {
		t.Fatalf("POST with a data channel: %d, want 400", rsp.StatusCode)
	}
	if got := f.service.Count(); got != 0 {
		t.Fatalf("%d sessions, want 0", got)
	}
}

func TestPostRefusesWithoutAValidToken(t *testing.T) {
	f := newFixture(t, 4)
	offer := newPlayer(t).offer(t)

	for _, tc := range []struct {
		name, token string
		want        int
	}{
		{"no token", "", http.StatusUnauthorized},
		{"wrong token", "nkvm_wrong", http.StatusUnauthorized},
		{"full key", f.fullKey, http.StatusForbidden},
	} {
		rsp := f.do(http.MethodPost, "/api/whep", tc.token, "application/sdp", offer)
		if rsp.StatusCode != tc.want {
			t.Errorf("%s: %d, want %d", tc.name, rsp.StatusCode, tc.want)
		}
		if tc.want == http.StatusUnauthorized && !strings.HasPrefix(rsp.Header.Get("WWW-Authenticate"), "Bearer") {
			t.Errorf("%s: WWW-Authenticate %q", tc.name, rsp.Header.Get("WWW-Authenticate"))
		}
	}
	if got := f.service.Count(); got != 0 {
		t.Fatalf("%d sessions, want 0", got)
	}
	if !contains(f.auditActions(), "whep-start:denied") {
		t.Fatal("a refused start is not in the audit log")
	}
}

func TestDisabledEndpointIsNotFound(t *testing.T) {
	f := newFixture(t, 4)
	f.enabled.Store(false)

	rsp := f.do(http.MethodPost, "/api/whep", f.viewKey, "application/sdp", newPlayer(t).offer(t))
	if rsp.StatusCode != http.StatusNotFound {
		t.Fatalf("POST while off: %d, want 404", rsp.StatusCode)
	}
}

// Switching WHEP off ends the sessions it had.
func TestDisablingEndsLiveSessions(t *testing.T) {
	f := newFixture(t, 4)
	f.start(newPlayer(t))

	if err := f.service.SetEnabled(false); err != nil {
		t.Fatal(err)
	}
	waitFor(t, "the session to end", func() bool { return f.service.Count() == 0 })
}

// A session whose key is revoked ends at the next check.
func TestRevokedKeyEndsTheSession(t *testing.T) {
	f := newFixture(t, 4)
	f.start(newPlayer(t))

	keys, err := apikey.List("admin")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range keys {
		if key.Scope == apikey.ScopeView {
			if err := apikey.RevokeAny(key.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	waitFor(t, "the session to end", func() bool { return f.service.Count() == 0 })
}

func TestSessionLimit(t *testing.T) {
	f := newFixture(t, 1)
	f.start(newPlayer(t))

	rsp := f.do(http.MethodPost, "/api/whep", f.viewKey, "application/sdp", newPlayer(t).offer(t))
	if rsp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("POST over the limit: %d, want 503", rsp.StatusCode)
	}
	if rsp.Header.Get("Retry-After") == "" {
		t.Fatal("no Retry-After over the limit")
	}
}

func TestPostWantsSDP(t *testing.T) {
	f := newFixture(t, 4)

	rsp := f.do(http.MethodPost, "/api/whep", f.viewKey, "application/json", "{}")
	if rsp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("JSON body: %d, want 415", rsp.StatusCode)
	}
	rsp = f.do(http.MethodPost, "/api/whep", f.viewKey, "application/sdp", "not sdp")
	if rsp.StatusCode != http.StatusBadRequest {
		t.Fatalf("garbage offer: %d, want 400", rsp.StatusCode)
	}
}

// Trickle ICE is supported; an ICE restart is not, which RFC 9725 answers
// with 422.
func TestPatchTrickleAndRestart(t *testing.T) {
	f := newFixture(t, 4)
	p := newPlayer(t)
	location := f.start(p)

	ufrag := ""
	for _, line := range strings.Split(p.pc.LocalDescription().SDP, "\r\n") {
		if strings.HasPrefix(line, "a=ice-ufrag:") {
			ufrag = strings.TrimPrefix(line, "a=ice-ufrag:")
			break
		}
	}

	trickle := "a=ice-ufrag:" + ufrag + "\r\nm=video 9 UDP/TLS/RTP/SAVPF 0\r\na=mid:0\r\n" +
		"a=candidate:1 1 udp 2130706431 192.0.2.7 50000 typ host\r\na=end-of-candidates\r\n"
	if rsp := f.do(http.MethodPatch, location, f.viewKey, "application/trickle-ice-sdpfrag", trickle); rsp.StatusCode != http.StatusNoContent {
		t.Fatalf("trickle PATCH: %d, want 204", rsp.StatusCode)
	}

	restart := "a=ice-ufrag:brandnew\r\na=ice-pwd:brandnewpasswordbrandnewpw\r\nm=video 9 UDP/TLS/RTP/SAVPF 0\r\na=mid:0\r\n"
	if rsp := f.do(http.MethodPatch, location, f.viewKey, "application/trickle-ice-sdpfrag", restart); rsp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("restart PATCH: %d, want 422", rsp.StatusCode)
	}

	if rsp := f.do(http.MethodPatch, location, f.viewKey, "application/sdp", trickle); rsp.StatusCode != http.StatusUnsupportedMediaType {
		t.Fatalf("PATCH with the wrong type: %d, want 415", rsp.StatusCode)
	}
	if rsp := f.do(http.MethodPatch, location, "", "application/trickle-ice-sdpfrag", trickle); rsp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("PATCH without a token: %d, want 401", rsp.StatusCode)
	}
}

// Browser players on another origin need CORS. The preflight is answered
// without credentials, and the answer to it says what the endpoint takes.
func TestOptionsAndCORS(t *testing.T) {
	f := newFixture(t, 4)

	req, _ := http.NewRequest(http.MethodOptions, f.server.URL+"/api/whep", nil)
	req.Header.Set("Origin", "https://player.example")
	req.Header.Set("Access-Control-Request-Method", "POST")
	req.Header.Set("Access-Control-Request-Headers", "authorization, content-type")
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	_ = rsp.Body.Close()

	if rsp.StatusCode != http.StatusNoContent && rsp.StatusCode != http.StatusOK {
		t.Fatalf("OPTIONS: %d", rsp.StatusCode)
	}
	if got := rsp.Header.Get("Accept-Post"); got != "application/sdp" {
		t.Errorf("Accept-Post %q", got)
	}
	if got := rsp.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("Access-Control-Allow-Origin %q, want *", got)
	}
	if got := rsp.Header.Get("Access-Control-Allow-Credentials"); got != "" {
		t.Errorf("Access-Control-Allow-Credentials %q: cookies are never taken", got)
	}
	if got := strings.ToLower(rsp.Header.Get("Access-Control-Allow-Headers")); !strings.Contains(got, "authorization") {
		t.Errorf("Access-Control-Allow-Headers %q", got)
	}

	location := f.start(newPlayer(t))
	post := f.do(http.MethodDelete, location, f.viewKey, "", "")
	if got := post.Header.Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("DELETE Access-Control-Allow-Origin %q", got)
	}
}
