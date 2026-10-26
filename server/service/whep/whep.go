// Package whep serves the board's video over WHEP, the WebRTC-HTTP Egress
// Protocol (ironkvm-dist#145): a player POSTs an SDP offer to /api/whep and
// gets the answer, with the session's URL in Location, and DELETEs that URL
// to leave. GStreamer's whepsrc and whepclientsrc, and browser WHEP players,
// speak it. It is modelled on WHIP (RFC 9725), which the WHEP draft follows
// for everything but the direction of the media.
//
// The endpoint is off until an administrator switches it on. It answers only
// to a Bearer token that can do no more than watch (middleware.AuthenticateWHEP),
// and each session is an ordinary WebRTC viewer of the board, counted as one.
package whep

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/audit"
	"NanoKVM-Server/service/stream/webrtc"
)

const (
	// Endpoint is where a player POSTs its offer. A session lives under it.
	Endpoint = "/api/whep"

	// DefaultMax is how many WHEP sessions may run at once. Each one costs the
	// board's single core about as much as a browser viewer (SRTP in
	// software, one UDP send a packet); a few players beside the web UI is
	// what it carries without the browser's picture suffering.
	DefaultMax = 4

	// maxBody bounds an offer or a PATCH. An offer with every candidate is a
	// few KB.
	maxBody = 64 << 10

	// defaultRecheck is how often a live session's credential and the setting
	// are checked again, as the session middleware rechecks a login.
	defaultRecheck = 5 * time.Second

	contentTypeSDP     = "application/sdp"
	contentTypeTrickle = "application/trickle-ice-sdpfrag"
)

// Session is what the service keeps of a WHEP session. webrtc.WHEPSession is
// the real one.
type Session interface {
	AddCandidates(fragment string) error
	Close(reason string)
	Done() <-chan struct{}
	Reason() string
}

// Deps are what the service needs from the rest of the server.
type Deps struct {
	// Enabled reports the setting, and SetEnabled saves and applies it.
	Enabled    func() bool
	SetEnabled func(bool) error

	// Max is how many sessions may run at once; 0 means DefaultMax.
	Max int

	// Authenticate checks a request's credential; middleware.AuthenticateWHEP
	// when nil.
	Authenticate func(*http.Request) (middleware.WHEPViewer, int)

	// NewSession answers an offer; webrtc.NewWHEPSession when nil.
	NewSession func(offer, user string) (Session, string, error)

	// Recheck is how often a live session's credential is checked again;
	// defaultRecheck when zero.
	Recheck time.Duration
}

type entry struct {
	id      string
	session Session
	viewer  middleware.WHEPViewer
	ip      string
	started time.Time
	stop    chan struct{}
}

// Service is the WHEP endpoint and its setting.
type Service struct {
	deps Deps

	mutex    sync.Mutex
	sessions map[string]*entry
	// pending counts offers being answered, so the limit holds while they
	// are.
	pending int
}

func New(deps Deps) *Service {
	if deps.Max <= 0 {
		deps.Max = DefaultMax
	}
	if deps.Authenticate == nil {
		deps.Authenticate = middleware.AuthenticateWHEP
	}
	if deps.NewSession == nil {
		deps.NewSession = func(offer, user string) (Session, string, error) {
			return webrtc.NewWHEPSession(offer, user)
		}
	}
	if deps.Recheck <= 0 {
		deps.Recheck = defaultRecheck
	}

	return &Service{deps: deps, sessions: make(map[string]*entry)}
}

// Register adds the WHEP routes. They take no session middleware: the
// credential is the Bearer token alone, and the cookie is never read.
func (s *Service) Register(r gin.IRouter) {
	r.OPTIONS(Endpoint, s.options)
	r.POST(Endpoint, s.post)
	r.OPTIONS(Endpoint+"/:id", s.options)
	r.PATCH(Endpoint+"/:id", s.patch)
	r.DELETE(Endpoint+"/:id", s.delete)
}

// Count is how many sessions are live.
func (s *Service) Count() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.sessions)
}

// cors lets a browser player on another origin call the endpoint. No
// credentials are allowed: the token travels in a header the page sets, and
// the cookie is never read, so a wildcard origin gives a page nothing it did
// not already hold.
func cors(c *gin.Context) {
	h := c.Writer.Header()
	h.Set("Access-Control-Allow-Origin", "*")
	h.Set("Access-Control-Allow-Methods", "OPTIONS, POST, PATCH, DELETE")
	h.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, If-Match")
	h.Set("Access-Control-Expose-Headers", "Location, Link, Accept-Post, Accept-Patch, Retry-After")
	h.Set("Access-Control-Max-Age", "600")
}

func (s *Service) options(c *gin.Context) {
	cors(c)
	if c.FullPath() == Endpoint {
		c.Header("Accept-Post", contentTypeSDP)
	} else {
		c.Header("Accept-Patch", contentTypeTrickle)
	}
	c.Status(http.StatusNoContent)
}

// refuse ends a request with a status and a line of text for whoever reads
// the player's log.
func refuse(c *gin.Context, status int, text string) {
	if status == http.StatusUnauthorized {
		c.Header("WWW-Authenticate", `Bearer realm="whep"`)
	}
	c.Data(status, "text/plain; charset=utf-8", []byte(text+"\n"))
	c.Abort()
}

func (s *Service) enabled() bool {
	return s.deps.Enabled != nil && s.deps.Enabled()
}

// authenticate checks the credential and ends the request when it fails.
func (s *Service) authenticate(c *gin.Context) (middleware.WHEPViewer, bool) {
	viewer, status := s.deps.Authenticate(c.Request)
	switch status {
	case 0:
		return viewer, true
	case http.StatusForbidden:
		refuse(c, status, "this token can do more than watch; use a view-only API key or a share link")
	default:
		refuse(c, http.StatusUnauthorized, "a view-only API key or a share link token is needed as a Bearer token")
	}
	return viewer, false
}

// hasType reports whether the request's body is of the media type want.
func hasType(c *gin.Context, want string) bool {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	return err == nil && strings.EqualFold(mediaType, want)
}

// body reads a request body of at most maxBody bytes.
func body(c *gin.Context) (string, bool) {
	data, err := io.ReadAll(io.LimitReader(c.Request.Body, maxBody+1))
	if err != nil {
		refuse(c, http.StatusBadRequest, "could not read the body")
		return "", false
	}
	if len(data) > maxBody {
		refuse(c, http.StatusRequestEntityTooLarge, "the body is too large")
		return "", false
	}
	return string(data), true
}

func (s *Service) post(c *gin.Context) {
	cors(c)
	if !s.enabled() {
		refuse(c, http.StatusNotFound, "WHEP is switched off on this board")
		return
	}

	viewer, ok := s.authenticate(c)
	if !ok {
		s.record(c, viewer.Principal, "whep-start", audit.OutcomeDenied, "", "refused credential")
		return
	}
	if !hasType(c, contentTypeSDP) {
		refuse(c, http.StatusUnsupportedMediaType, "the offer must be application/sdp")
		return
	}
	offer, ok := body(c)
	if !ok {
		return
	}

	if !s.reserve() {
		c.Header("Retry-After", "10")
		refuse(c, http.StatusServiceUnavailable, fmt.Sprintf("this board serves at most %d WHEP sessions at once", s.deps.Max))
		s.record(c, viewer.Principal, "whep-start", audit.OutcomeDenied, "", "session limit")
		return
	}

	session, answer, err := s.deps.NewSession(offer, viewer.Principal.Username)
	if err != nil {
		s.unreserve()
		if errors.Is(err, webrtc.ErrWHEPOffer) || errors.Is(err, webrtc.ErrWHEPDataChannel) {
			refuse(c, http.StatusBadRequest, err.Error())
			return
		}
		log.Errorf("whep: failed to answer an offer: %s", err)
		refuse(c, http.StatusInternalServerError, "could not set up the session")
		return
	}

	id, err := newID()
	if err != nil {
		s.unreserve()
		session.Close("no session id")
		refuse(c, http.StatusInternalServerError, "could not set up the session")
		return
	}

	e := &entry{id: id, session: session, viewer: viewer, ip: audit.ClientIP(c), started: time.Now(), stop: make(chan struct{})}
	s.mutex.Lock()
	s.pending--
	s.sessions[id] = e
	s.mutex.Unlock()
	go s.watch(e)

	s.record(c, viewer.Principal, "whep-start", audit.OutcomeOK, id, "")

	location := Endpoint + "/" + id
	for _, link := range iceServerLinks() {
		c.Writer.Header().Add("Link", link)
	}
	c.Header("Location", location)
	c.Data(http.StatusCreated, contentTypeSDP, []byte(answer))
}

// reserve takes a place under the limit for an offer being answered.
func (s *Service) reserve() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.sessions)+s.pending >= s.deps.Max {
		return false
	}
	s.pending++
	return true
}

func (s *Service) unreserve() {
	s.mutex.Lock()
	s.pending--
	s.mutex.Unlock()
}

// session finds the session a request names, checks that the request comes
// from whoever opened it, and ends the request when either fails.
func (s *Service) session(c *gin.Context) (*entry, bool) {
	if !s.enabled() {
		refuse(c, http.StatusNotFound, "WHEP is switched off on this board")
		return nil, false
	}
	viewer, ok := s.authenticate(c)
	if !ok {
		return nil, false
	}

	s.mutex.Lock()
	e, found := s.sessions[c.Param("id")]
	s.mutex.Unlock()
	// Someone else's session is not found rather than forbidden, so a
	// request says nothing about which ids exist.
	if !found || e.viewer.Principal.Username != viewer.Principal.Username {
		refuse(c, http.StatusNotFound, "no such session")
		return nil, false
	}
	return e, true
}

func (s *Service) patch(c *gin.Context) {
	cors(c)
	e, ok := s.session(c)
	if !ok {
		return
	}
	if !hasType(c, contentTypeTrickle) {
		refuse(c, http.StatusUnsupportedMediaType, "a PATCH must be application/trickle-ice-sdpfrag")
		return
	}
	fragment, ok := body(c)
	if !ok {
		return
	}

	switch err := e.session.AddCandidates(fragment); {
	case err == nil:
		c.Status(http.StatusNoContent)
	case errors.Is(err, webrtc.ErrWHEPRestart):
		// Trickle ICE is supported and ICE restarts are not, which RFC 9725
		// (section 4.3.1) answers with 422. The player starts a new session.
		refuse(c, http.StatusUnprocessableEntity, err.Error())
	default:
		refuse(c, http.StatusBadRequest, err.Error())
	}
}

func (s *Service) delete(c *gin.Context) {
	cors(c)
	e, ok := s.session(c)
	if !ok {
		return
	}

	s.end(e, "ended by the player")
	c.Status(http.StatusOK)
}

// end closes a session and waits for watch to let it go.
func (s *Service) end(e *entry, reason string) {
	e.session.Close(reason)
	<-e.stop
}

// watch waits for a session to end, by the player, the connection or a
// credential that no longer holds, and takes it off the list.
func (s *Service) watch(e *entry) {
	ticker := time.NewTicker(s.deps.Recheck)
	defer ticker.Stop()

	for {
		select {
		case <-e.session.Done():
			s.forget(e)
			return
		case <-ticker.C:
			switch {
			case !s.enabled():
				e.session.Close("WHEP was switched off")
			case e.viewer.Valid != nil && !e.viewer.Valid():
				e.session.Close("the credential was revoked or expired")
			}
		}
	}
}

// forget takes an ended session off the list and records its end.
func (s *Service) forget(e *entry) {
	s.mutex.Lock()
	delete(s.sessions, e.id)
	s.mutex.Unlock()

	p := e.viewer.Principal
	audit.Record(audit.Event{
		Actor:    p.Username,
		Via:      p.Via,
		Key:      p.Key,
		Role:     string(p.Role),
		IP:       e.ip,
		Category: audit.CategoryAuth,
		Action:   "whep-end",
		Target:   e.id,
		Outcome:  audit.OutcomeOK,
		Detail:   e.session.Reason() + " after " + time.Since(e.started).Round(time.Second).String(),
	})
	close(e.stop)
}

// CloseAll ends every session.
func (s *Service) CloseAll(reason string) {
	s.mutex.Lock()
	entries := make([]*entry, 0, len(s.sessions))
	for _, e := range s.sessions {
		entries = append(entries, e)
	}
	s.mutex.Unlock()

	for _, e := range entries {
		s.end(e, reason)
	}
}

// SetEnabled saves and applies the setting. Switching it off ends the
// sessions there are.
func (s *Service) SetEnabled(enabled bool) error {
	if s.deps.SetEnabled == nil {
		return errors.New("the setting cannot be changed")
	}
	if err := s.deps.SetEnabled(enabled); err != nil {
		return err
	}
	if !enabled {
		s.CloseAll("WHEP was switched off")
	}
	return nil
}

// record writes a WHEP event to the audit log. The request's middleware does
// not: the routes are self-audited (service/audit/rules.go).
func (s *Service) record(c *gin.Context, p middleware.Principal, action, outcome, target, detail string) {
	audit.Request(c, audit.Event{
		Actor:    p.Username,
		Via:      p.Via,
		Key:      p.Key,
		Role:     string(p.Role),
		Category: audit.CategoryAuth,
		Action:   action,
		Target:   target,
		Outcome:  outcome,
		Detail:   detail,
	})
}

// GetSettings is the web UI's view of the setting.
func (s *Service) GetSettings(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.settings())
}

// SetSettings switches WHEP on or off.
func (s *Service) SetSettings(c *gin.Context) {
	var req proto.SetWHEPSettingsReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil || req.Enabled == nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}
	if err := s.SetEnabled(*req.Enabled); err != nil {
		log.Errorf("whep: failed to save the setting: %s", err)
		rsp.ErrRsp(c, -2, "failed to save the setting")
		return
	}

	log.Infof("whep: endpoint switched %s", map[bool]string{true: "on", false: "off"}[*req.Enabled])
	rsp.OkRspWithData(c, s.settings())
}

func (s *Service) settings() *proto.WHEPSettingsRsp {
	return &proto.WHEPSettingsRsp{
		Enabled:     s.enabled(),
		Path:        Endpoint,
		Sessions:    s.Count(),
		MaxSessions: s.deps.Max,
	}
}

// iceServerLinks are the Link header values for the board's STUN and TURN
// servers (RFC 9725, section 4.6). A browser viewer is given the same list.
func iceServerLinks() []string {
	var links []string
	for _, server := range webrtc.ICEServers() {
		for _, url := range server.URLs {
			link := "<" + url + `>; rel="ice-server"`
			if server.Username != "" {
				link += "; username=" + strconv.Quote(server.Username)
			}
			if credential, ok := server.Credential.(string); ok && credential != "" {
				link += "; credential=" + strconv.Quote(credential) + `; credential-type="password"`
			}
			links = append(links, link)
		}
	}
	return links
}

func newID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
