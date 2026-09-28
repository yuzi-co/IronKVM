package ipmi

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"
)

// testClient is a remote console written from the specification, the way
// ipmitool does it, so the tests can drive the handshake and break it on
// purpose.
type testClient struct {
	t    *testing.T
	conn net.Conn

	suite     *cipherSuite
	consoleID uint32
	systemID  uint32
	in        rakpInput
	k1, k2    []byte
	seq       uint32
	rqSeq     byte

	// ignoreRAKP2 skips the RAKP 2 HMAC check, so a test can send RAKP 3
	// with a wrong password and see what the service does with it.
	ignoreRAKP2 bool
}

func dial(t *testing.T, s *Service) *testClient {
	t.Helper()
	conn, err := net.Dial("udp", s.LocalAddr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })
	return &testClient{t: t, conn: conn}
}

func (c *testClient) send(packet []byte) {
	c.t.Helper()
	if _, err := c.conn.Write(packet); err != nil {
		c.t.Fatal(err)
	}
}

var errNoAnswer = errors.New("no answer")

// receive waits a short while for a packet. Silence is an answer too: the
// service drops what it will not answer.
func (c *testClient) receive() ([]byte, error) {
	buf := make([]byte, 2048)
	_ = c.conn.SetReadDeadline(time.Now().Add(300 * time.Millisecond))
	n, err := c.conn.Read(buf)
	if err != nil {
		var ne net.Error
		if errors.As(err, &ne) && ne.Timeout() {
			return nil, errNoAnswer
		}
		return nil, err
	}
	return buf[:n], nil
}

func (c *testClient) roundTrip(packet []byte) ([]byte, error) {
	c.send(packet)
	return c.receive()
}

func buildRequest(netFn, cmd, rqSeq byte, data []byte) []byte {
	out := []byte{bmcAddr, netFn << 2}
	out = append(out, checksum(out))
	out = append(out, consoleAddr, rqSeq<<2, cmd)
	out = append(out, data...)
	return append(out, checksum(out[3:]))
}

// parseResponse returns a response's completion code and data.
func parseResponse(msg []byte) (byte, []byte, error) {
	if len(msg) < 8 {
		return 0, nil, errShortPacket
	}
	if checksum(msg[:2]) != msg[2] || checksum(msg[3:len(msg)-1]) != msg[len(msg)-1] {
		return 0, nil, errors.New("bad checksum")
	}
	return msg[6], msg[7 : len(msg)-1], nil
}

// v15Command sends a request in IPMI 1.5 framing outside a session.
func (c *testClient) v15Command(netFn, cmd byte, data []byte) (byte, []byte, error) {
	msg := buildRequest(netFn, cmd, 0, data)
	packet := []byte{rmcpVersion, 0, rmcpNoAck, rmcpClassIPMI, authTypeNone, 0, 0, 0, 0, 0, 0, 0, 0, byte(len(msg))}
	reply, err := c.roundTrip(append(packet, msg...))
	if err != nil {
		return 0, nil, err
	}
	p, err := parseV15(reply)
	if err != nil {
		return 0, nil, err
	}
	return parseResponse(p.msg)
}

// presession sends an unauthenticated RMCP+ request with session ID 0.
func (c *testClient) presession(netFn, cmd byte, data []byte) (byte, []byte, error) {
	reply, err := c.roundTrip(buildV20(v20Header{payloadType: payloadIPMI}, buildRequest(netFn, cmd, 0, data), nil, nil))
	if err != nil {
		return 0, nil, err
	}
	p, err := parseV20(reply)
	if err != nil {
		return 0, nil, err
	}
	return parseResponse(p.payload)
}

// openSession sends Open Session with three algorithms and returns the
// status.
func (c *testClient) openSession(auth, integrity, confidentiality, priv byte) (byte, error) {
	var id [4]byte
	_, _ = rand.Read(id[:])
	c.consoleID = binary.LittleEndian.Uint32(id[:]) | 1

	b := []byte{0x42, priv, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, c.consoleID)
	b = append(b, 0x00, 0, 0, 8, auth, 0, 0, 0)
	b = append(b, 0x01, 0, 0, 8, integrity, 0, 0, 0)
	b = append(b, 0x02, 0, 0, 8, confidentiality, 0, 0, 0)
	reply, err := c.roundTrip(buildV20(v20Header{payloadType: payloadOpenRequest}, b, nil, nil))
	if err != nil {
		return 0, err
	}
	p, err := parseV20(reply)
	if err != nil {
		return 0, err
	}
	if p.kind() != payloadOpenResponse || len(p.payload) < 8 {
		return 0, fmt.Errorf("unexpected answer %x", reply)
	}
	if p.payload[0] != 0x42 || binary.LittleEndian.Uint32(p.payload[4:8]) != c.consoleID {
		return 0, errors.New("the answer does not match the request")
	}
	if status := p.payload[1]; status != statusOK {
		return status, nil
	}
	if len(p.payload) < 36 {
		return 0, errors.New("short open session response")
	}
	c.systemID = binary.LittleEndian.Uint32(p.payload[8:12])
	c.suite = findCipherSuite(p.payload[16], p.payload[24], p.payload[32])
	if c.suite == nil {
		return 0, errors.New("the service chose algorithms it does not offer")
	}
	return statusOK, nil
}

// rakp runs RAKP 1 to 4 and returns the first failing status, from RAKP 2
// or RAKP 4. The RAKP 2 HMAC is checked against password.
func (c *testClient) rakp(username, password string, role byte) (byte, error) {
	in := &c.in
	in.consoleID = c.consoleID
	in.systemID = c.systemID
	_, _ = rand.Read(in.rm[:])
	in.role = role
	in.username = []byte(username)

	b := []byte{0x43, 0, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, c.systemID)
	b = append(b, in.rm[:]...)
	b = append(b, role, 0, 0, byte(len(username)))
	b = append(b, username...)
	reply, err := c.roundTrip(buildV20(v20Header{payloadType: payloadRAKP1}, b, nil, nil))
	if err != nil {
		return 0, err
	}
	p, err := parseV20(reply)
	if err != nil {
		return 0, err
	}
	if p.kind() != payloadRAKP2 || len(p.payload) < 8 || p.payload[0] != 0x43 {
		return 0, fmt.Errorf("unexpected RAKP 2 %x", reply)
	}
	if status := p.payload[1]; status != statusOK {
		return status, nil
	}
	if len(p.payload) < 40+c.suite.hash().Size() {
		return 0, errors.New("short RAKP 2")
	}
	copy(in.rc[:], p.payload[8:24])
	copy(in.guid[:], p.payload[24:40])
	kuid := []byte(password)
	if got, want := p.payload[40:], c.suite.rakp2AuthCode(kuid, in); !c.ignoreRAKP2 && string(got) != string(want) {
		return 0, errors.New("the RAKP 2 HMAC does not match the password")
	}

	b = []byte{0x44, statusOK, 0, 0}
	b = binary.LittleEndian.AppendUint32(b, c.systemID)
	b = append(b, c.suite.rakp3AuthCode(kuid, in)...)
	reply, err = c.roundTrip(buildV20(v20Header{payloadType: payloadRAKP3}, b, nil, nil))
	if err != nil {
		return 0, err
	}
	p, err = parseV20(reply)
	if err != nil {
		return 0, err
	}
	if p.kind() != payloadRAKP4 || len(p.payload) < 8 || p.payload[0] != 0x44 {
		return 0, fmt.Errorf("unexpected RAKP 4 %x", reply)
	}
	if status := p.payload[1]; status != statusOK {
		return status, nil
	}
	sik := c.suite.sik(kuid, in)
	if got, want := p.payload[8:], c.suite.rakp4ICV(sik, in); string(got) != string(want) {
		return 0, errors.New("the RAKP 4 ICV does not match")
	}
	c.k1 = c.suite.k1(sik)
	c.k2 = c.suite.k2(sik)
	return statusOK, nil
}

// login opens a session with a cipher suite and runs RAKP, failing the
// test on any error.
func (c *testClient) login(suiteID byte, username, password string, role byte) {
	c.t.Helper()
	suite := suiteByID(c.t, suiteID)
	if status, err := c.openSession(suite.auth, suite.integrity, suite.confidentiality, 0); err != nil || status != statusOK {
		c.t.Fatalf("open session: status %#x, %v", status, err)
	}
	if status, err := c.rakp(username, password, role); err != nil || status != statusOK {
		c.t.Fatalf("RAKP: status %#x, %v", status, err)
	}
}

// sessionPacket builds a request in the session with a given sequence
// number.
func (c *testClient) sessionPacket(seq uint32, rqSeq, netFn, cmd byte, data []byte) []byte {
	c.t.Helper()
	payload, err := encryptPayload(c.k2, buildRequest(netFn, cmd, rqSeq, data))
	if err != nil {
		c.t.Fatal(err)
	}
	return buildV20(v20Header{
		payloadType: payloadIPMI | payloadEncrypted | payloadAuthenticated,
		sessionID:   c.systemID,
		seq:         seq,
	}, payload, c.suite, c.k1)
}

// readSessionAnswer checks and decrypts an answer in the session.
func (c *testClient) readSessionAnswer(reply []byte) (byte, []byte, error) {
	p, err := parseV20(reply)
	if err != nil {
		return 0, nil, err
	}
	if p.sessionID != c.consoleID || !p.authenticated() || !p.encrypted() {
		return 0, nil, fmt.Errorf("unexpected answer header %+v", p.v20Header)
	}
	if !checkTrailer(p, c.suite, c.k1) {
		return 0, nil, errors.New("the answer's AuthCode is wrong")
	}
	plain, err := decryptPayload(c.k2, p.payload)
	if err != nil {
		return 0, nil, err
	}
	return parseResponse(plain)
}

// command sends one request in the session.
func (c *testClient) command(netFn, cmd byte, data []byte) (byte, []byte, error) {
	c.seq++
	c.rqSeq = (c.rqSeq + 1) & 0x3f
	reply, err := c.roundTrip(c.sessionPacket(c.seq, c.rqSeq, netFn, cmd, data))
	if err != nil {
		return 0, nil, err
	}
	return c.readSessionAnswer(reply)
}

// mustCommand sends a request and wants a completion code.
func (c *testClient) mustCommand(netFn, cmd byte, data []byte, wantCC byte) []byte {
	c.t.Helper()
	cc, out, err := c.command(netFn, cmd, data)
	if err != nil {
		c.t.Fatalf("netFn %#x cmd %#x: %v", netFn, cmd, err)
	}
	if cc != wantCC {
		c.t.Fatalf("netFn %#x cmd %#x: completion code %#x, want %#x", netFn, cmd, cc, wantCC)
	}
	return out
}
