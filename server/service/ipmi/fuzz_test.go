package ipmi

import (
	"bytes"
	"encoding/binary"
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/authn"
)

// The fuzz tests feed the packet parsers and the handshake whatever a
// client on the network could send. Each one starts from valid packets
// built with the package's own encoders and the test client, so the fuzzer
// begins past the length and type checks. Besides "no panic", each checks
// what must hold for any input: an answer always parses, the session table
// stays bounded, and no handshake step succeeds without the right HMAC.
//
// Run one with, for example:
//
//	go test -tags novision -run '^$' -fuzz '^FuzzHandlePacket$' -fuzztime 3m ./service/ipmi/

// fuzzAddr is the one client address the fuzz tests send from.
var fuzzAddr = &net.UDPAddr{IP: net.IPv4(192, 0, 2, 1), Port: 50000}

// fuzzUsers are the accounts a fuzz service has, by the index the scripts
// pick them with. The last one does not exist.
var fuzzUsers = []struct{ name, password string }{
	{adminName, adminIPMI},
	{userName, userIPMI},
	{"nobody", "not-a-password"},
}

// fuzzAccounts is an in-memory account list. authn.Store hashes web
// passwords with bcrypt, which would make every fuzz worker wait half a
// minute before its first input.
type fuzzAccounts map[string]authn.User

func (a fuzzAccounts) List() ([]authn.UserInfo, error) { return nil, nil }

func (a fuzzAccounts) Get(username string) (*authn.User, error) {
	user, ok := a[username]
	if !ok {
		return nil, errors.New("no such account")
	}
	return &user, nil
}

func (a fuzzAccounts) Authenticate(username, password string) (*authn.User, bool, error) {
	return nil, false, nil
}

func (a fuzzAccounts) SetIPMIPassword(username, sealed string) (*authn.User, error) {
	return nil, errors.New("read-only")
}

type nopLimiter struct{}

func (nopLimiter) Locked(ip, username string) bool { return false }
func (nopLimiter) Failed(ip, username string)      {}
func (nopLimiter) Succeeded(ip, username string)   {}

// fuzzEnv is a service with no socket, driven through handleLocked. It is
// built once per fuzz test, since accounts and the keyring live in files,
// and reset before every input.
type fuzzEnv struct {
	s   *Service
	now time.Time
}

func newFuzzEnv(f *testing.F) *fuzzEnv {
	f.Helper()
	keyring := NewKeyring(filepath.Join(f.TempDir(), "ipmi.key"))
	accounts := fuzzAccounts{}
	for _, a := range []struct {
		name, password string
		role           authn.Role
	}{
		{adminName, adminIPMI, authn.RoleAdmin},
		{userName, userIPMI, authn.RoleUser},
	} {
		sealed, err := keyring.Seal(a.name, a.password)
		if err != nil {
			f.Fatal(err)
		}
		accounts[a.name] = authn.User{Username: a.name, Role: a.role, Enabled: true, IPMIPassword: sealed}
	}

	// A wrong password is logged at warning level, and the fuzzer sends
	// thousands of them a second.
	level := log.GetLevel()
	log.SetLevel(log.ErrorLevel)
	f.Cleanup(func() { log.SetLevel(level) })

	env := &fuzzEnv{}
	var guid [16]byte
	copy(guid[:], testGUIDValue)
	env.s = New(Deps{
		// A press does nothing: millions of them must not pile up anywhere.
		PressButton:       func(string, time.Duration) error { return nil },
		PowerLED:          func() (bool, error) { return true, nil },
		PowerLEDConnected: func() bool { return true },
		Accounts:          accounts,
		Keyring:           keyring,
		Limiter:           nopLimiter{},
		FirmwareVersion:   func() string { return "2.4.1" },
		GUID:              guid,
		Now:               func() time.Time { return env.now },
	})
	return env
}

func (env *fuzzEnv) reset() {
	env.s.mu.Lock()
	defer env.s.mu.Unlock()
	clear(env.s.sessions)
	env.now = time.Unix(1_800_000_000, 0)
}

// exchange hands one packet to the service and checks the answer.
func (env *fuzzEnv) exchange(t *testing.T, packet []byte) []byte {
	t.Helper()
	env.s.mu.Lock()
	reply := env.s.handleLocked(packet, fuzzAddr)
	env.checkSessionsLocked(t)
	env.s.mu.Unlock()
	checkReply(t, reply)
	return reply
}

func (env *fuzzEnv) client(t *testing.T) *testClient {
	return &testClient{t: t, exchange: func(packet []byte) []byte { return env.exchange(t, packet) }}
}

// session returns the service's session by its ID, or nil.
func (env *fuzzEnv) session(id uint32) *session {
	env.s.mu.Lock()
	defer env.s.mu.Unlock()
	return env.s.sessions[id]
}

func (env *fuzzEnv) checkSessionsLocked(t *testing.T) {
	t.Helper()
	if n := len(env.s.sessions); n > maxSessions {
		t.Fatalf("%d sessions, more than %d", n, maxSessions)
	}
	for id, sess := range env.s.sessions {
		if id == 0 || id != sess.id {
			t.Fatalf("session %#x is filed under %#x", sess.id, id)
		}
		if sess.state == stateActive && (len(sess.k1) == 0 || len(sess.k2) < 16 || sess.kuid != nil) {
			t.Fatalf("active session %#x without its keys, or still holding the password", id)
		}
	}
}

// checkReply fails on an answer the service should never send: one that
// does not parse as what it claims to be, or is bigger than any request.
func checkReply(t *testing.T, reply []byte) {
	t.Helper()
	if reply == nil {
		return
	}
	if len(reply) > maxPacket {
		t.Fatalf("a %d byte answer", len(reply))
	}
	authType, err := parseRMCP(reply)
	if err != nil {
		t.Fatalf("answer %x: %v", reply, err)
	}
	switch authType {
	case authTypeNone:
		p, err := parseV15(reply)
		if err != nil {
			t.Fatalf("IPMI 1.5 answer %x: %v", reply, err)
		}
		if _, _, err := parseResponse(p.msg); err != nil {
			t.Fatalf("IPMI 1.5 answer %x: %v", reply, err)
		}
	case authTypeRMCPPlus:
		p, err := parseV20(reply)
		if err != nil {
			t.Fatalf("RMCP+ answer %x: %v", reply, err)
		}
		if p.kind() == payloadIPMI && !p.encrypted() {
			if _, _, err := parseResponse(p.payload); err != nil {
				t.Fatalf("RMCP+ answer %x: %v", reply, err)
			}
		}
	default:
		t.Fatalf("answer %x with auth type %#x", reply, authType)
	}
}

// Seed packets, built the way ipmitool's lanplus interface sends them.

// asfPing is an RMCP ASF Presence Ping, IPMI 2.0 section 13.2.3. Some
// discovery tools send it first. The service does not answer it.
var asfPing = []byte{rmcpVersion, 0, rmcpNoAck, 0x06, 0x00, 0x00, 0x11, 0xbe, 0x80, 0x00, 0x00, 0x00}

func seedV15(netFn, cmd byte, data []byte) []byte {
	msg := buildRequest(netFn, cmd, 0, data)
	packet := []byte{rmcpVersion, 0, rmcpNoAck, rmcpClassIPMI, authTypeNone, 0, 0, 0, 0, 0, 0, 0, 0, byte(len(msg))}
	return append(packet, msg...)
}

func seedPresession(netFn, cmd byte, data []byte) []byte {
	return buildV20(v20Header{payloadType: payloadIPMI}, buildRequest(netFn, cmd, 0, data), nil, nil)
}

func seedOpenRequest(suite *cipherSuite, priv byte) []byte {
	b := []byte{0x42, priv, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, 0xa4a3a2a1)
	b = append(b, 0x00, 0, 0, 8, suite.auth, 0, 0, 0)
	b = append(b, 0x01, 0, 0, 8, suite.integrity, 0, 0, 0)
	b = append(b, 0x02, 0, 0, 8, suite.confidentiality, 0, 0, 0)
	return b
}

func seedRAKP1(systemID uint32, role byte, username string) []byte {
	b := []byte{0x43, 0, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, systemID)
	for i := range 16 {
		b = append(b, byte(i))
	}
	b = append(b, role, 0, 0, byte(len(username)))
	return append(b, username...)
}

// seedKey is a fixed K1 and K2 for the parser seeds.
var seedKey = bytes.Repeat([]byte{0x5a}, 32)

func seedSessionPacket(suite *cipherSuite, seq uint32, msg []byte) []byte {
	payload, err := encryptPayload(seedKey, msg)
	if err != nil {
		panic(err)
	}
	return buildV20(v20Header{
		payloadType: payloadIPMI | payloadEncrypted | payloadAuthenticated,
		sessionID:   0x12345678,
		seq:         seq,
	}, payload, suite, seedKey)
}

// seedPackets is every kind of packet a client sends, for the parser
// fuzzers.
func seedPackets() [][]byte {
	s3, s17 := cipherSuites[0], cipherSuites[1]
	return [][]byte{
		{},
		{rmcpVersion, 0, rmcpNoAck, rmcpClassIPMI},
		asfPing,
		seedV15(netFnApp, cmdGetChannelAuthCaps, []byte{0x8e, privAdmin}),
		seedV15(netFnApp, cmdGetChannelCipherSuite, []byte{0x0e, 0x00, 0x80}),
		seedPresession(netFnApp, cmdGetChannelAuthCaps, []byte{0x8e, privAdmin}),
		seedPresession(netFnApp, cmdGetChannelCipherSuite, []byte{0x0e, 0x00, 0x80}),
		buildV20(v20Header{payloadType: payloadOpenRequest}, seedOpenRequest(s3, 0), nil, nil),
		buildV20(v20Header{payloadType: payloadOpenRequest}, seedOpenRequest(s17, privAdmin), nil, nil),
		buildV20(v20Header{payloadType: payloadRAKP1}, seedRAKP1(0x12345678, privAdmin|lookupNameOnly, adminName), nil, nil),
		buildV20(v20Header{payloadType: payloadRAKP3}, append([]byte{0x44, 0, 0, 0, 0x78, 0x56, 0x34, 0x12}, seedKey[:20]...), nil, nil),
		seedSessionPacket(s3, 1, buildRequest(netFnApp, cmdGetDeviceID, 1, nil)),
		seedSessionPacket(s17, 2, buildRequest(netFnChassis, cmdChassisControl, 2, []byte{controlPowerUp})),
	}
}

// FuzzParseRMCP is the RMCP header, IPMI 2.0 section 13.1.3, the first
// thing every packet meets. ASF packets, such as a presence ping, have
// class 0x06 and are refused here.
func FuzzParseRMCP(f *testing.F) {
	for _, p := range seedPackets() {
		f.Add(p)
	}
	f.Fuzz(func(t *testing.T, packet []byte) {
		authType, err := parseRMCP(packet)
		if err != nil {
			return
		}
		if len(packet) <= rmcpHeaderLen || packet[0] != rmcpVersion || packet[3] != rmcpClassIPMI {
			t.Fatalf("accepted %x", packet)
		}
		if authType != packet[rmcpHeaderLen] {
			t.Fatalf("auth type %#x from %x", authType, packet)
		}
	})
}

// FuzzParseV15 is the IPMI 1.5 session wrapper and the IPMI message in it,
// IPMI 2.0 sections 13.6 and 13.8, as Get Channel Authentication
// Capabilities arrives before a session.
func FuzzParseV15(f *testing.F) {
	for _, p := range seedPackets() {
		f.Add(p)
	}
	f.Fuzz(func(t *testing.T, packet []byte) {
		p, err := parseV15(packet)
		if err != nil {
			return
		}
		if len(p.msg) != int(packet[rmcpHeaderLen+9]) {
			t.Fatalf("a %d byte message from length byte %d", len(p.msg), packet[rmcpHeaderLen+9])
		}
		if p.seq != binary.LittleEndian.Uint32(packet[5:9]) || p.sessionID != binary.LittleEndian.Uint32(packet[9:13]) {
			t.Fatalf("header %+v from %x", p, packet)
		}

		req, err := parseRequest(p.msg)
		if err != nil {
			return
		}
		if len(req.data) != len(p.msg)-7 || req.netFn > 0x3f || req.rqSeq > 0x3f || req.rqLUN > 3 || req.rsLUN > 3 {
			t.Fatalf("request %+v from %x", req, p.msg)
		}
		reply := buildV15(buildResponse(req, ccOK, req.data))
		checkReply(t, reply)
		back, _ := parseV15(reply)
		cc, data, err := parseResponse(back.msg)
		if err != nil || cc != ccOK || !bytes.Equal(data, req.data) {
			t.Fatalf("response to %+v reads back as %#x %x %v", req, cc, data, err)
		}
	})
}

// FuzzParseV20 is the RMCP+ session header, IPMI 2.0 section 13.6: the
// payload type with its encrypted and authenticated bits, the session ID,
// the sequence number and the payload length, then the integrity trailer
// and the AES-CBC-128 payload.
func FuzzParseV20(f *testing.F) {
	for _, p := range seedPackets() {
		f.Add(p, byte(0))
		f.Add(p, byte(1))
	}
	f.Fuzz(func(t *testing.T, packet []byte, suiteSel byte) {
		p, err := parseV20(packet)
		if err != nil {
			return
		}
		if len(packet) < v20PayloadAt {
			t.Fatalf("accepted a %d byte packet", len(packet))
		}
		if n := int(binary.LittleEndian.Uint16(packet[14:16])); len(p.payload) != n {
			t.Fatalf("a %d byte payload from length %d", len(p.payload), n)
		}
		if p.kind() == 0x02 {
			t.Fatal("accepted an OEM payload")
		}
		again, err := parseV20(buildV20(p.v20Header, p.payload, nil, nil))
		if err != nil || again.v20Header != p.v20Header || !bytes.Equal(again.payload, p.payload) {
			t.Fatalf("%+v does not survive a round trip: %+v %v", p.v20Header, again, err)
		}

		suite := cipherSuites[int(suiteSel)%len(cipherSuites)]
		if checkTrailer(p, suite, seedKey) {
			// Only a seed carries a good AuthCode. Flipping any bit of it
			// has to break it.
			tampered := append([]byte(nil), packet...)
			tampered[len(tampered)-1] ^= 0x01
			tp, _ := parseV20(tampered)
			if checkTrailer(tp, suite, seedKey) {
				t.Fatalf("a tampered AuthCode passes: %x", tampered)
			}
		}

		plain, err := decryptPayload(seedKey, p.payload)
		if err != nil {
			return
		}
		if len(plain) > len(p.payload)-2*16+15 {
			t.Fatalf("%d bytes of plaintext from a %d byte payload", len(plain), len(p.payload))
		}
		sealed, err := encryptPayload(seedKey, plain)
		if err != nil {
			t.Fatal(err)
		}
		back, err := decryptPayload(seedKey, sealed)
		if err != nil || !bytes.Equal(back, plain) {
			t.Fatalf("plaintext %x does not survive a round trip: %x %v", plain, back, err)
		}
	})
}

// FuzzOpenSession is the RMCP+ Open Session Request, IPMI 2.0 section
// 13.17. prefill opens that many sessions first, so a full table is
// reached too.
func FuzzOpenSession(f *testing.F) {
	for _, suite := range cipherSuites {
		for _, priv := range []byte{0, privUser, privAdmin, 0x0f} {
			f.Add(seedOpenRequest(suite, priv), byte(0), uint32(0), byte(0))
		}
	}
	f.Add(seedOpenRequest(&cipherSuite{}, 0), byte(0), uint32(0), byte(0))
	f.Add(seedOpenRequest(cipherSuites[0], 0), byte(payloadAuthenticated), uint32(0), byte(0))
	f.Add(seedOpenRequest(cipherSuites[0], 0), byte(0), uint32(7), byte(0))
	f.Add(seedOpenRequest(cipherSuites[1], 0), byte(0), uint32(0), byte(maxSessions))
	f.Add(seedOpenRequest(cipherSuites[1], 0)[:31], byte(0), uint32(0), byte(0))

	env := newFuzzEnv(f)
	f.Fuzz(func(t *testing.T, payload []byte, flags byte, sessionID uint32, prefill byte) {
		env.reset()
		for range int(prefill) % (maxSessions + 1) {
			env.exchange(t, buildV20(v20Header{payloadType: payloadOpenRequest}, seedOpenRequest(cipherSuites[0], 0), nil, nil))
		}
		env.s.mu.Lock()
		before := len(env.s.sessions)
		env.s.mu.Unlock()

		reply := env.exchange(t, buildV20(v20Header{payloadType: payloadOpenRequest | flags&0xc0, sessionID: sessionID}, payload, nil, nil))
		env.s.mu.Lock()
		after := len(env.s.sessions)
		env.s.mu.Unlock()
		if reply == nil {
			if after != before {
				t.Fatalf("a dropped request changed the sessions from %d to %d", before, after)
			}
			return
		}

		p, _ := parseV20(reply)
		if p.kind() != payloadOpenResponse || len(p.payload) < 8 || p.payload[0] != payload[0] ||
			!bytes.Equal(p.payload[4:8], payload[4:8]) {
			t.Fatalf("answer %x to %x", reply, payload)
		}
		if p.payload[1] != statusOK {
			if after != before {
				t.Fatalf("a refused request changed the sessions from %d to %d", before, after)
			}
			return
		}

		if len(p.payload) != 36 {
			t.Fatalf("a %d byte Open Session Response", len(p.payload))
		}
		sess := env.session(binary.LittleEndian.Uint32(p.payload[8:12]))
		if sess == nil || sess.state != stateOpened || sess.consoleID != binary.LittleEndian.Uint32(payload[4:8]) {
			t.Fatalf("the session of answer %x is %+v", reply, sess)
		}
		if findCipherSuite(p.payload[16], p.payload[24], p.payload[32]) != sess.suite {
			t.Fatalf("answer %x names another suite than %d", reply, sess.suite.id)
		}
	})
}

// FuzzRAKP1 is RAKP Message 1, IPMI 2.0 section 13.20, after a good Open
// Session. Bit 0 of flags writes the real session ID into the payload,
// which the fuzzer cannot guess, bit 1 sends a nonzero session header ID,
// and the top two bits are the payload type's.
func FuzzRAKP1(f *testing.F) {
	for _, u := range fuzzUsers {
		f.Add(seedRAKP1(0, privAdmin|lookupNameOnly, u.name), byte(1), byte(0))
		f.Add(seedRAKP1(0, privUser, u.name), byte(1), byte(1))
	}
	f.Add(seedRAKP1(0, privAdmin, ""), byte(1), byte(0))
	f.Add(seedRAKP1(0, 0x07, adminName), byte(1), byte(0))
	f.Add(seedRAKP1(0, privAdmin, "seventeen-chars-x"), byte(1), byte(0))
	f.Add(seedRAKP1(0, privAdmin, adminName)[:30], byte(1), byte(0))
	f.Add(seedRAKP1(0, privAdmin, adminName), byte(0), byte(0))
	f.Add(seedRAKP1(0, privAdmin, adminName), byte(1|payloadEncrypted), byte(0))

	env := newFuzzEnv(f)
	f.Fuzz(func(t *testing.T, payload []byte, flags, suiteSel byte) {
		env.reset()
		c := env.client(t)
		suite := cipherSuites[int(suiteSel)%len(cipherSuites)]
		if status, err := c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0); err != nil || status != statusOK {
			t.Fatalf("open session: %#x %v", status, err)
		}
		payload = append([]byte(nil), payload...)
		if flags&1 != 0 && len(payload) >= 8 {
			binary.LittleEndian.PutUint32(payload[4:8], c.systemID)
		}
		h := v20Header{payloadType: payloadRAKP1 | flags&0xc0}
		if flags&2 != 0 {
			h.sessionID = c.systemID
		}

		reply := env.exchange(t, buildV20(h, payload, nil, nil))
		sess := env.session(c.systemID)
		if reply == nil {
			if sess == nil || sess.state != stateOpened {
				t.Fatalf("a dropped RAKP 1 changed the session: %+v", sess)
			}
			return
		}

		p, _ := parseV20(reply)
		if p.kind() != payloadRAKP2 || len(p.payload) < 8 || p.payload[0] != payload[0] ||
			binary.LittleEndian.Uint32(p.payload[4:8]) != c.consoleID {
			t.Fatalf("answer %x to %x", reply, payload)
		}
		if p.payload[1] != statusOK {
			if sess != nil {
				t.Fatalf("the session outlived RAKP 2 status %#x", p.payload[1])
			}
			return
		}

		if sess == nil || sess.state != stateRAKP2Sent {
			t.Fatalf("RAKP 2 went out, the session is %+v", sess)
		}
		if len(p.payload) != 40+suite.hash().Size() {
			t.Fatalf("a %d byte RAKP 2", len(p.payload))
		}
		// The HMAC has to be under the password of the account the request
		// named, over the name exactly as it came.
		name := string(payload[28 : 28+int(payload[27])])
		var password string
		for _, u := range fuzzUsers[:2] {
			if u.name == name {
				password = u.password
			}
		}
		if password == "" {
			t.Fatalf("RAKP 2 went out for %q", name)
		}
		if want := suite.rakp2AuthCode([]byte(password), &sess.rakp); !bytes.Equal(p.payload[40:], want) {
			t.Fatalf("RAKP 2 HMAC %x, want %x", p.payload[40:], want)
		}
	})
}

// FuzzRAKP3 is RAKP Message 3, IPMI 2.0 section 13.22, after a good RAKP 1
// for admin. Bit 0 of flags writes the real session ID into the payload,
// bit 1 the right HMAC after it, and the top two bits are the payload
// type's. Nothing else may make the session active.
func FuzzRAKP3(f *testing.F) {
	good := []byte{0x44, statusOK, 0, 0, 0, 0, 0, 0}
	for _, sel := range []byte{0, 1} {
		f.Add(good, byte(3), sel)
		f.Add(good, byte(1), sel)
		f.Add(append(append([]byte(nil), good...), seedKey[:20]...), byte(1), sel)
	}
	f.Add([]byte{0x44, statusInvalidIntegrityChk, 0, 0, 0, 0, 0, 0}, byte(3), byte(0))
	f.Add(good[:7], byte(1), byte(0))
	f.Add(good, byte(3|payloadAuthenticated), byte(0))
	f.Add(good, byte(2), byte(1))

	env := newFuzzEnv(f)
	f.Fuzz(func(t *testing.T, payload []byte, flags, suiteSel byte) {
		env.reset()
		c := env.client(t)
		suite := cipherSuites[int(suiteSel)%len(cipherSuites)]
		if status, err := c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0); err != nil || status != statusOK {
			t.Fatalf("open session: %#x %v", status, err)
		}
		if status, err := c.rakp1(adminName, adminIPMI, privAdmin|lookupNameOnly); err != nil || status != statusOK {
			t.Fatalf("RAKP 1: %#x %v", status, err)
		}
		want := suite.rakp3AuthCode([]byte(adminIPMI), &c.in)

		payload = append([]byte(nil), payload...)
		if flags&1 != 0 && len(payload) >= 8 {
			binary.LittleEndian.PutUint32(payload[4:8], c.systemID)
		}
		if flags&2 != 0 && len(payload) >= 8 {
			payload = append(payload[:8], want...)
		}

		reply := env.exchange(t, buildV20(v20Header{payloadType: payloadRAKP3 | flags&0xc0}, payload, nil, nil))
		sess := env.session(c.systemID)
		if reply == nil {
			if sess != nil && sess.state != stateRAKP2Sent {
				t.Fatalf("a dropped RAKP 3 changed the session: %+v", sess)
			}
			return
		}

		p, _ := parseV20(reply)
		if p.kind() != payloadRAKP4 || len(p.payload) < 8 || p.payload[0] != payload[0] ||
			binary.LittleEndian.Uint32(p.payload[4:8]) != c.consoleID {
			t.Fatalf("answer %x to %x", reply, payload)
		}
		if p.payload[1] != statusOK {
			if sess != nil {
				t.Fatalf("the session outlived RAKP 4 status %#x", p.payload[1])
			}
			return
		}

		if payload[1] != statusOK || !bytes.Equal(payload[8:], want) {
			t.Fatalf("RAKP 3 %x logged in; the HMAC is %x", payload, want)
		}
		if sess == nil || sess.state != stateActive {
			t.Fatalf("RAKP 4 went out, the session is %+v", sess)
		}
		sik := suite.sik([]byte(adminIPMI), &c.in)
		if icv := suite.rakp4ICV(sik, &c.in); !bytes.Equal(p.payload[8:], icv) {
			t.Fatalf("RAKP 4 ICV %x, want %x", p.payload[8:], icv)
		}
	})
}

// The operations of a FuzzHandlePacket script. Each is one byte, taken
// modulo opCount, followed by its operands.
const (
	opRaw        = iota // length, packet: sent as it is
	opLogin             // suite, user, role: a whole handshake
	opSealed            // seq (4), length, message: encrypted and signed in the session
	opSigned            // payload type, seq (4), length, payload: signed, not encrypted
	opWait              // seconds: the clock moves on
	opCommand           // netFn, cmd, length, data: a well-formed request in the session
	opPresession        // netFn, cmd, length, data: a well-formed request outside one
	opCount
)

// script builds FuzzHandlePacket seeds.
type script []byte

func (s script) raw(packet []byte) script {
	return append(append(s, opRaw, byte(len(packet))), packet...)
}
func (s script) login(suite, user, role byte) script {
	return append(s, opLogin, suite, user, role)
}
func (s script) sealed(seq uint32, msg []byte) script {
	s = binary.LittleEndian.AppendUint32(append(s, opSealed), seq)
	return append(append(s, byte(len(msg))), msg...)
}
func (s script) signed(payloadType byte, seq uint32, payload []byte) script {
	s = binary.LittleEndian.AppendUint32(append(s, opSigned, payloadType), seq)
	return append(append(s, byte(len(payload))), payload...)
}
func (s script) wait(seconds byte) script { return append(s, opWait, seconds) }
func (s script) command(netFn, cmd byte, data []byte) script {
	return append(append(s, opCommand, netFn, cmd, byte(len(data))), data...)
}
func (s script) presession(netFn, cmd byte, data []byte) script {
	return append(append(s, opPresession, netFn, cmd, byte(len(data))), data...)
}

// take splits n bytes off b, or as many as there are.
func take(b []byte, n int) ([]byte, []byte) {
	n = min(n, len(b))
	return b[:n], b[n:]
}

// FuzzHandlePacket is handleLocked, the function the UDP listener calls
// with every packet. A script of operations runs against one service, so
// sessions carry from one packet to the next. The harness does the
// cryptography a fuzzer cannot: it logs in, and it encrypts and signs
// what the script wants sent in the session.
func FuzzHandlePacket(f *testing.F) {
	s3, s17 := cipherSuites[0], cipherSuites[1]
	for _, s := range []script{
		script{}.raw(asfPing),
		script{}.raw(seedV15(netFnApp, cmdGetChannelAuthCaps, []byte{0x8e, privAdmin})).
			raw(seedPresession(netFnApp, cmdGetChannelCipherSuite, []byte{0x0e, 0x00, 0x80})),
		script{}.presession(netFnApp, cmdGetChannelAuthCaps, []byte{0x8e, privAdmin}).
			presession(netFnApp, cmdGetChannelCipherSuite, []byte{0x0e, 0x00, 0x81}).
			presession(netFnApp, cmdGetSystemGUID, nil).
			presession(netFnChassis, cmdChassisControl, []byte{controlPowerDown}),
		script{}.raw(buildV20(v20Header{payloadType: payloadOpenRequest}, seedOpenRequest(s3, 0), nil, nil)).
			raw(buildV20(v20Header{payloadType: payloadRAKP1}, seedRAKP1(0, privAdmin, adminName), nil, nil)),
		script{}.login(0, 0, privAdmin|lookupNameOnly).
			command(netFnApp, cmdGetDeviceID, nil).
			command(netFnApp, cmdSetSessionPrivilege, []byte{privAdmin}).
			command(netFnChassis, cmdGetChassisStatus, nil).
			command(netFnChassis, cmdChassisControl, []byte{controlPowerCycle}).
			command(netFnChassis, cmdChassisIdentify, nil).
			command(netFnApp, cmdCloseSession, []byte{0, 0, 0, 0}),
		script{}.login(1, 1, privUser|lookupNameOnly).
			command(netFnApp, cmdSetSessionPrivilege, []byte{privAdmin}).
			command(netFnChassis, cmdChassisControl, []byte{controlHardReset}),
		script{}.login(0, 0, privAdmin|lookupNameOnly).
			sealed(1, buildRequest(netFnApp, cmdGetDeviceID, 1, nil)).
			sealed(1, buildRequest(netFnApp, cmdGetDeviceID, 1, nil)).
			sealed(40, buildRequest(netFnApp, cmdGetDeviceID, 2, nil)).
			sealed(5, buildRequest(netFnApp, cmdGetDeviceID, 3, nil)),
		script{}.login(1, 0, privAdmin|lookupNameOnly).
			signed(payloadIPMI|payloadAuthenticated, 1, buildRequest(netFnApp, cmdGetDeviceID, 1, nil)).
			signed(payloadIPMI|payloadAuthenticated|payloadEncrypted, 2, bytes.Repeat([]byte{0xaa}, 32)).
			signed(payloadRAKP3|payloadAuthenticated, 3, nil),
		script{}.login(0, 0, privAdmin|lookupNameOnly).wait(61).command(netFnApp, cmdGetDeviceID, nil),
		script{}.login(0, 2, privAdmin).login(0, 0, 0x07),
		script{}.raw(seedSessionPacket(s17, 1, buildRequest(netFnApp, cmdGetDeviceID, 1, nil))),
	} {
		f.Add([]byte(s))
	}

	env := newFuzzEnv(f)
	f.Fuzz(func(t *testing.T, in []byte) {
		env.reset()
		var c *testClient
		var rqSeq byte
		loggedIn := func() bool { return c != nil && c.k1 != nil }
		// expectSessionAnswer checks an answer inside the session, which
		// has to pass its own AuthCode and decrypt.
		expectSessionAnswer := func(reply []byte) {
			if reply == nil {
				return
			}
			if _, _, err := c.readSessionAnswer(reply); err != nil {
				t.Fatalf("answer %x in the session: %v", reply, err)
			}
		}

		for len(in) > 0 {
			op := in[0] % opCount
			in = in[1:]
			switch op {
			case opRaw:
				if len(in) == 0 {
					return
				}
				var packet []byte
				packet, in = take(in[1:], int(in[0]))
				env.exchange(t, packet)

			case opLogin:
				var args []byte
				if args, in = take(in, 3); len(args) < 3 {
					return
				}
				suite := cipherSuites[int(args[0])%len(cipherSuites)]
				u := fuzzUsers[int(args[1])%len(fuzzUsers)]
				c = env.client(t)
				rqSeq = 0
				if status, err := c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0); err == nil && status == statusOK {
					_, _ = c.rakp(u.name, u.password, args[2])
				}

			case opSealed:
				var args, msg []byte
				if args, in = take(in, 5); len(args) < 5 {
					return
				}
				msg, in = take(in, int(args[4]))
				if !loggedIn() {
					continue
				}
				payload, err := encryptPayload(c.k2, msg)
				if err != nil {
					t.Fatal(err)
				}
				expectSessionAnswer(env.exchange(t, buildV20(v20Header{
					payloadType: payloadIPMI | payloadEncrypted | payloadAuthenticated,
					sessionID:   c.systemID,
					seq:         binary.LittleEndian.Uint32(args[:4]),
				}, payload, c.suite, c.k1)))

			case opSigned:
				var args, payload []byte
				if args, in = take(in, 6); len(args) < 6 {
					return
				}
				payload, in = take(in, int(args[5]))
				if !loggedIn() {
					continue
				}
				expectSessionAnswer(env.exchange(t, buildV20(v20Header{
					payloadType: args[0] | payloadAuthenticated,
					sessionID:   c.systemID,
					seq:         binary.LittleEndian.Uint32(args[1:5]),
				}, payload, c.suite, c.k1)))

			case opWait:
				if len(in) == 0 {
					return
				}
				env.now = env.now.Add(time.Duration(in[0]) * time.Second)
				in = in[1:]

			case opCommand, opPresession:
				var args, data []byte
				if args, in = take(in, 3); len(args) < 3 {
					return
				}
				data, in = take(in, int(args[2]))
				netFn, cmd := args[0]&0x3f, args[1]
				if op == opPresession {
					env.exchange(t, seedPresession(netFn, cmd, data))
					env.exchange(t, seedV15(netFn, cmd, data))
					continue
				}
				if !loggedIn() {
					continue
				}
				if netFn == netFnApp && cmd == cmdCloseSession && len(data) >= 4 {
					data = append([]byte(nil), data...)
					binary.LittleEndian.PutUint32(data[:4], c.systemID)
				}
				c.seq++
				rqSeq = (rqSeq + 1) & 0x3f
				expectSessionAnswer(env.exchange(t, c.sessionPacket(c.seq, rqSeq, netFn, cmd, data)))
			}
		}
	})
}
