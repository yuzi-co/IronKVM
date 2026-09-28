package ipmi

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"encoding/binary"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
)

// The privilege levels, IPMI 2.0 table 6-2.
const (
	privCallback = 0x01
	privUser     = 0x02
	privOperator = 0x03
	privAdmin    = 0x04
)

// The RMCP+ status codes of Open Session and RAKP, IPMI 2.0 table 13-15.
const (
	statusOK                  = 0x00
	statusNoResources         = 0x01
	statusInvalidSessionID    = 0x02
	statusInvalidRole         = 0x09
	statusUnauthorizedRole    = 0x0a
	statusInvalidNameLength   = 0x0c
	statusUnauthorizedName    = 0x0d
	statusInvalidIntegrityChk = 0x0f
	statusNoCipherSuiteMatch  = 0x11
	statusIllegalParameter    = 0x12
)

const (
	// lookupNameOnly is RAKP 1's bit for a name-only user lookup.
	lookupNameOnly = 0x10
	// maxUsernameLen is the IPMI limit on a user name.
	maxUsernameLen = 16
	// maxSessions bounds the sessions, those still in the handshake
	// included.
	maxSessions = 16
	// handshakeTimeout drops a session that has not finished RAKP, and
	// sessionTimeout one that has been quiet, as the specification's
	// default does.
	handshakeTimeout = 10 * time.Second
	sessionTimeout   = 60 * time.Second
	// replayWindow is how far behind the highest inbound sequence number
	// a late one may be.
	replayWindow = 32
	// repeatedRequestLifetime is how long the last answer is kept for a
	// request sent again.
	repeatedRequestLifetime = 5 * time.Second

	openSessionRequestMinBytes = 32
)

type sessionState int

const (
	stateOpened sessionState = iota
	stateRAKP2Sent
	stateActive
)

// session is one RMCP+ session, from Open Session on. id is the managed
// system's session ID, which the console sends; consoleID is the one the
// service sends back.
type session struct {
	id        uint32
	consoleID uint32
	addr      string
	suite     *cipherSuite
	state     sessionState
	lastSeen  time.Time

	// The handshake.
	rakp     rakpInput
	username string
	kuid     []byte
	// sealed is the account's sealed password the session was opened with.
	// A session whose account no longer has it ends.
	sealed string

	// The session proper.
	k1    []byte
	k2    []byte
	limit byte // the highest privilege the session may take
	priv  byte // the privilege it has now

	// highestSeq and seen are the inbound sliding window: bit i of seen is
	// set when highestSeq-i arrived.
	highestSeq uint32
	seen       uint32
	outSeq     uint32

	// last is the last answer, kept for a request sent again.
	last *answered
	// closing is set by Close Session: the answer goes out, then the
	// session is gone.
	closing bool
}

// answered is one request and the IPMI response it got.
type answered struct {
	netFn, cmd, rqSeq byte
	data              []byte
	response          []byte
	at                time.Time
}

func (sess *session) timeout() time.Duration {
	if sess.state == stateActive {
		return sessionTimeout
	}
	return handshakeTimeout
}

func (s *Service) expireLocked(now time.Time) {
	for id, sess := range s.sessions {
		if now.Sub(sess.lastSeen) >= sess.timeout() {
			delete(s.sessions, id)
		}
	}
}

func (s *Service) newSessionIDLocked() (uint32, bool) {
	var b [4]byte
	for range 8 {
		if _, err := rand.Read(b[:]); err != nil {
			return 0, false
		}
		id := binary.LittleEndian.Uint32(b[:])
		if _, taken := s.sessions[id]; id != 0 && !taken {
			return id, true
		}
	}
	return 0, false
}

// openSession answers an RMCP+ Open Session Request, IPMI 2.0 section
// 13.17. The three algorithms have to be one of the two suites exactly.
func (s *Service) openSession(p *v20Packet, addr net.Addr, now time.Time) []byte {
	if p.sessionID != 0 || p.encrypted() || p.authenticated() {
		return nil
	}
	b := p.payload
	if len(b) < openSessionRequestMinBytes {
		return nil
	}
	tag := b[0]
	requested := b[1] & 0x0f
	consoleID := binary.LittleEndian.Uint32(b[4:8])

	fail := func(status byte) []byte {
		out := []byte{tag, status, 0, 0}
		out = binary.LittleEndian.AppendUint32(out, consoleID)
		return buildV20(v20Header{payloadType: payloadOpenResponse}, out, nil, nil)
	}

	if consoleID == 0 {
		return fail(statusInvalidSessionID)
	}
	if requested > privAdmin {
		return fail(statusInvalidRole)
	}
	auth, okA := algorithm(b[8:16], 0x00)
	integrity, okI := algorithm(b[16:24], 0x01)
	confidentiality, okC := algorithm(b[24:32], 0x02)
	if !okA || !okI || !okC {
		return fail(statusIllegalParameter)
	}
	suite := findCipherSuite(auth, integrity, confidentiality)
	if suite == nil {
		log.Debugf("ipmi: %s asked for algorithms %d/%d/%d, which no offered cipher suite has", addr, auth, integrity, confidentiality)
		return fail(statusNoCipherSuiteMatch)
	}

	if len(s.sessions) >= maxSessions {
		return fail(statusNoResources)
	}
	id, ok := s.newSessionIDLocked()
	if !ok {
		return fail(statusNoResources)
	}
	s.sessions[id] = &session{
		id:        id,
		consoleID: consoleID,
		addr:      addr.String(),
		suite:     suite,
		state:     stateOpened,
		lastSeen:  now,
	}

	if requested == 0 {
		requested = privAdmin
	}
	out := []byte{tag, statusOK, requested, 0}
	out = binary.LittleEndian.AppendUint32(out, consoleID)
	out = binary.LittleEndian.AppendUint32(out, id)
	out = append(out, 0x00, 0, 0, 8, suite.auth, 0, 0, 0)
	out = append(out, 0x01, 0, 0, 8, suite.integrity, 0, 0, 0)
	out = append(out, 0x02, 0, 0, 8, suite.confidentiality, 0, 0, 0)
	return buildV20(v20Header{payloadType: payloadOpenResponse}, out, nil, nil)
}

// algorithm reads one of the Open Session Request's three algorithm
// payloads. A zero payload length would mean "any", which the service does
// not offer: the client has to name the suite.
func algorithm(b []byte, payloadType byte) (byte, bool) {
	if b[0] != payloadType || b[3] != 8 {
		return 0, false
	}
	return b[4] & 0x3f, true
}

// userLimit is the highest privilege an account may take: ADMINISTRATOR
// for an admin, USER for anyone else, who can read the power state but not
// press a button, as in Redfish.
func userLimit(user *authn.User) byte {
	if user.Role == authn.RoleAdmin {
		return privAdmin
	}
	return privUser
}

func hostOf(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// rakp1 answers RAKP Message 1 with RAKP Message 2, IPMI 2.0 sections 13.20
// and 13.21. This is where the weakness of IPMI 2.0 authentication lives:
// message 2 carries an HMAC keyed by the user's password to anyone who
// names the user, before that anyone has proved anything, and it can be
// cracked offline. Nothing in the protocol avoids it.
func (s *Service) rakp1(p *v20Packet, addr net.Addr, now time.Time) []byte {
	if p.sessionID != 0 || p.encrypted() || p.authenticated() {
		return nil
	}
	b := p.payload
	if len(b) < 28 {
		return nil
	}
	tag := b[0]
	sess, ok := s.sessions[binary.LittleEndian.Uint32(b[4:8])]
	if !ok || sess.state != stateOpened || sess.addr != addr.String() {
		return nil
	}

	fail := func(status byte) []byte {
		delete(s.sessions, sess.id)
		out := []byte{tag, status, 0, 0}
		out = binary.LittleEndian.AppendUint32(out, sess.consoleID)
		return buildV20(v20Header{payloadType: payloadRAKP2}, out, nil, nil)
	}

	in := &sess.rakp
	in.consoleID = sess.consoleID
	in.systemID = sess.id
	copy(in.rm[:], b[8:24])
	in.role = b[24]
	nameLen := int(b[27])
	if nameLen > maxUsernameLen || len(b) < 28+nameLen {
		return fail(statusInvalidNameLength)
	}
	in.username = append([]byte(nil), b[28:28+nameLen]...)
	in.guid = s.deps.GUID
	username := string(in.username)

	role := in.role &^ lookupNameOnly
	if role < privCallback || role > privAdmin {
		return fail(statusInvalidRole)
	}
	// Name-only lookup is the only one there is: every account has one
	// privilege limit, whatever the role asked for.
	if nameLen == 0 {
		return fail(statusUnauthorizedName)
	}

	ip := hostOf(sess.addr)
	if s.deps.Limiter != nil && s.deps.Limiter.Locked(ip, username) {
		log.Warnf("ipmi: %s is locked out of %q after failed logins", ip, username)
		return fail(statusUnauthorizedName)
	}
	user, password, sealed := s.ipmiUser(username)
	if user == nil {
		return fail(statusUnauthorizedName)
	}
	limit := userLimit(user)
	if role > limit {
		return fail(statusUnauthorizedRole)
	}

	if _, err := rand.Read(in.rc[:]); err != nil {
		return fail(statusNoResources)
	}
	sess.username = username
	sess.kuid = password
	sess.sealed = sealed
	sess.limit = role
	sess.state = stateRAKP2Sent
	sess.lastSeen = now

	out := []byte{tag, statusOK, 0, 0}
	out = binary.LittleEndian.AppendUint32(out, sess.consoleID)
	out = append(out, in.rc[:]...)
	out = append(out, in.guid[:]...)
	out = append(out, sess.suite.rakp2AuthCode(sess.kuid, in)...)
	return buildV20(v20Header{payloadType: payloadRAKP2}, out, nil, nil)
}

// ipmiUser returns an enabled account that can log in over IPMI, with its
// password and the sealed value it came from, or nil.
func (s *Service) ipmiUser(username string) (*authn.User, []byte, string) {
	if len(username) > maxUsernameLen {
		return nil, nil, ""
	}
	user, err := s.deps.Accounts.Get(username)
	if err != nil || !user.Enabled || user.IPMIPassword == "" || s.deps.Keyring == nil {
		return nil, nil, ""
	}
	password, err := s.deps.Keyring.Open(username, user.IPMIPassword)
	if err != nil {
		log.Errorf("ipmi: open the IPMI password of %q: %s", username, err)
		return nil, nil, ""
	}
	return user, password, user.IPMIPassword
}

// rakp3 checks RAKP Message 3 and answers with RAKP Message 4, IPMI 2.0
// sections 13.22 and 13.23. A good message 3 makes the session active.
func (s *Service) rakp3(p *v20Packet, addr net.Addr, now time.Time) []byte {
	if p.sessionID != 0 || p.encrypted() || p.authenticated() {
		return nil
	}
	b := p.payload
	if len(b) < 8 {
		return nil
	}
	tag := b[0]
	sess, ok := s.sessions[binary.LittleEndian.Uint32(b[4:8])]
	if !ok || sess.state != stateRAKP2Sent || sess.addr != addr.String() {
		return nil
	}
	in := &sess.rakp
	ip := hostOf(sess.addr)

	fail := func(status byte) []byte {
		delete(s.sessions, sess.id)
		out := []byte{tag, status, 0, 0}
		out = binary.LittleEndian.AppendUint32(out, sess.consoleID)
		return buildV20(v20Header{payloadType: payloadRAKP4}, out, nil, nil)
	}

	// The console reports its own failure, such as an RAKP 2 HMAC that did
	// not match its password, with a status here. There is nothing to
	// answer.
	if b[1] != statusOK {
		delete(s.sessions, sess.id)
		return nil
	}

	want := sess.suite.rakp3AuthCode(sess.kuid, in)
	if !hmac.Equal(want, b[8:]) {
		log.Warnf("ipmi: wrong password for %q from %s", sess.username, ip)
		if s.deps.Limiter != nil {
			s.deps.Limiter.Failed(ip, sess.username)
		}
		return fail(statusInvalidIntegrityChk)
	}
	if s.deps.Limiter != nil {
		s.deps.Limiter.Succeeded(ip, sess.username)
	}

	sik := sess.suite.sik(sess.kuid, in)
	sess.k1 = sess.suite.k1(sik)
	sess.k2 = sess.suite.k2(sik)
	clear(sess.kuid)
	sess.kuid = nil
	// A session starts at USER, or lower when that is all it asked for,
	// and Set Session Privilege Level takes it up to its limit.
	sess.priv = min(privUser, sess.limit)
	sess.state = stateActive
	sess.lastSeen = now

	// Debug, not info: ipmitool opens a session for every command, and a
	// tool that polls the power state would fill the log.
	log.Debugf("ipmi: %q logged in from %s with cipher suite %d", sess.username, ip, sess.suite.id)

	out := []byte{tag, statusOK, 0, 0}
	out = binary.LittleEndian.AppendUint32(out, sess.consoleID)
	out = append(out, sess.suite.rakp4ICV(sik, in)...)
	return buildV20(v20Header{payloadType: payloadRAKP4}, out, nil, nil)
}

// acceptSeq is the inbound sliding window. Zero is never valid, a number
// more than the window behind the highest one is refused, and so is one
// seen before.
func (sess *session) acceptSeq(seq uint32) bool {
	if seq == 0 {
		return false
	}
	if sess.highestSeq == 0 || seq > sess.highestSeq {
		shift := seq - sess.highestSeq
		if sess.highestSeq == 0 || shift >= replayWindow {
			sess.seen = 1
		} else {
			sess.seen = sess.seen<<shift | 1
		}
		sess.highestSeq = seq
		return true
	}
	behind := sess.highestSeq - seq
	if behind >= replayWindow || sess.seen&(1<<behind) != 0 {
		return false
	}
	sess.seen |= 1 << behind
	return true
}

// inSession handles an IPMI message in an active session. It has to be
// authenticated and encrypted, pass its AuthCode, and carry a sequence
// number the window accepts, or it is dropped.
func (s *Service) inSession(p *v20Packet, addr net.Addr, now time.Time) []byte {
	sess, ok := s.sessions[p.sessionID]
	if !ok || sess.state != stateActive || sess.addr != addr.String() {
		return nil
	}
	if !p.authenticated() || !p.encrypted() {
		return nil
	}
	if !checkTrailer(p, sess.suite, sess.k1) || !sess.acceptSeq(p.seq) {
		return nil
	}
	plain, err := decryptPayload(sess.k2, p.payload)
	if err != nil {
		return nil
	}
	req, err := parseRequest(plain)
	if err != nil {
		return nil
	}
	sess.lastSeen = now

	// The account may have been disabled, demoted, deleted or given a new
	// IPMI password since the session opened.
	user, err := s.deps.Accounts.Get(sess.username)
	if err != nil || !user.Enabled || user.IPMIPassword != sess.sealed {
		delete(s.sessions, sess.id)
		return nil
	}
	sess.limit = min(sess.limit, userLimit(user))
	sess.priv = min(sess.priv, sess.limit)

	var response []byte
	if last := sess.last; last != nil && now.Sub(last.at) < repeatedRequestLifetime &&
		last.netFn == req.netFn && last.cmd == req.cmd && last.rqSeq == req.rqSeq && bytes.Equal(last.data, req.data) {
		// ipmitool sends a request again, with the same request sequence
		// number, when the answer did not reach it. Doing it twice would
		// press a button twice.
		response = last.response
	} else {
		cc, data := s.dispatch(sess, req)
		response = buildResponse(req, cc, data)
		sess.last = &answered{
			netFn: req.netFn, cmd: req.cmd, rqSeq: req.rqSeq,
			data: append([]byte(nil), req.data...), response: response, at: now,
		}
	}

	payload, err := encryptPayload(sess.k2, response)
	if err != nil {
		return nil
	}
	sess.outSeq++
	if sess.outSeq == 0 {
		sess.outSeq = 1
	}
	out := buildV20(v20Header{
		payloadType: payloadIPMI | payloadEncrypted | payloadAuthenticated,
		sessionID:   sess.consoleID,
		seq:         sess.outSeq,
	}, payload, sess.suite, sess.k1)

	if sess.closing {
		delete(s.sessions, sess.id)
	}
	return out
}
