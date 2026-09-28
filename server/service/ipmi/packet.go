package ipmi

import (
	"crypto/hmac"
	"encoding/binary"
	"errors"
)

// RMCP, IPMI 2.0 section 13.1. Every packet starts with the four byte RMCP
// header; the service only speaks the IPMI class.
const (
	rmcpVersion   = 0x06
	rmcpNoAck     = 0xff
	rmcpClassIPMI = 0x07
)

// The session header's first byte. 0x00 is an IPMI 1.5 header without
// authentication, which clients use for Get Channel Authentication
// Capabilities before they know what the BMC speaks. 0x06 is RMCP+.
const (
	authTypeNone     = 0x00
	authTypeRMCPPlus = 0x06
)

// The RMCP+ payload types, IPMI 2.0 table 13-16. The top two bits of the
// byte say whether the payload is encrypted and authenticated.
const (
	payloadIPMI         = 0x00
	payloadOpenRequest  = 0x10
	payloadOpenResponse = 0x11
	payloadRAKP1        = 0x12
	payloadRAKP2        = 0x13
	payloadRAKP3        = 0x14
	payloadRAKP4        = 0x15

	payloadEncrypted     = 0x80
	payloadAuthenticated = 0x40
	payloadTypeMask      = 0x3f
)

// nextHeader is the integrity trailer's Next Header byte, which is always
// the RMCP class.
const nextHeader = 0x07

var errShortPacket = errors.New("short packet")

// v15Packet is an IPMI 1.5 session: auth type none, so there is no
// AuthCode, and the only one the service answers is session 0.
type v15Packet struct {
	seq       uint32
	sessionID uint32
	msg       []byte
}

// v20Header is the RMCP+ session header.
type v20Header struct {
	payloadType byte // with the encrypted and authenticated bits
	sessionID   uint32
	seq         uint32
}

// v20Packet is an RMCP+ packet taken apart. raw is the whole packet, which
// the AuthCode check needs.
type v20Packet struct {
	v20Header
	payload []byte
	raw     []byte
}

func (p *v20Packet) encrypted() bool     { return p.payloadType&payloadEncrypted != 0 }
func (p *v20Packet) authenticated() bool { return p.payloadType&payloadAuthenticated != 0 }
func (p *v20Packet) kind() byte          { return p.payloadType & payloadTypeMask }

// The offsets in an RMCP+ packet.
const (
	rmcpHeaderLen = 4
	v20HeaderLen  = 12 // auth type, payload type, session ID, sequence, length
	v20PayloadAt  = rmcpHeaderLen + v20HeaderLen
	v15HeaderLen  = 10 // auth type, sequence, session ID, length
)

// parseRMCP checks the RMCP header and returns the session header's auth
// type.
func parseRMCP(packet []byte) (byte, error) {
	if len(packet) < rmcpHeaderLen+1 {
		return 0, errShortPacket
	}
	if packet[0] != rmcpVersion || packet[3] != rmcpClassIPMI {
		return 0, errors.New("not an IPMI RMCP packet")
	}
	return packet[rmcpHeaderLen], nil
}

func parseV15(packet []byte) (*v15Packet, error) {
	if len(packet) < rmcpHeaderLen+v15HeaderLen {
		return nil, errShortPacket
	}
	h := packet[rmcpHeaderLen:]
	p := &v15Packet{
		seq:       binary.LittleEndian.Uint32(h[1:5]),
		sessionID: binary.LittleEndian.Uint32(h[5:9]),
	}
	length := int(h[9])
	if len(h) < v15HeaderLen+length {
		return nil, errShortPacket
	}
	p.msg = h[v15HeaderLen : v15HeaderLen+length]
	return p, nil
}

func buildV15(msg []byte) []byte {
	out := make([]byte, 0, rmcpHeaderLen+v15HeaderLen+len(msg)+1)
	out = append(out, rmcpVersion, 0, rmcpNoAck, rmcpClassIPMI)
	out = append(out, authTypeNone, 0, 0, 0, 0, 0, 0, 0, 0, byte(len(msg)))
	out = append(out, msg...)
	// IPMI 1.5 asks for a one byte legacy pad after a message that would
	// otherwise make the UDP payload 56, 84, 112, 128 or 156 bytes long.
	// The service's only 1.5 answers are far shorter, so it never adds one.
	return out
}

func parseV20(packet []byte) (*v20Packet, error) {
	if len(packet) < v20PayloadAt {
		return nil, errShortPacket
	}
	h := packet[rmcpHeaderLen:]
	p := &v20Packet{raw: packet}
	p.payloadType = h[1]
	if p.kind() == 0x02 {
		// An OEM explicit payload carries six more header bytes. The
		// service has no OEM payloads.
		return nil, errors.New("OEM payload")
	}
	p.sessionID = binary.LittleEndian.Uint32(h[2:6])
	p.seq = binary.LittleEndian.Uint32(h[6:10])
	length := int(binary.LittleEndian.Uint16(h[10:12]))
	if len(packet) < v20PayloadAt+length {
		return nil, errShortPacket
	}
	p.payload = packet[v20PayloadAt : v20PayloadAt+length]
	return p, nil
}

// checkTrailer verifies an authenticated packet's integrity trailer: the
// pad, the pad length, the Next Header byte and the AuthCode, which covers
// everything from the session header to the Next Header byte.
func checkTrailer(p *v20Packet, suite *cipherSuite, k1 []byte) bool {
	end := v20PayloadAt + len(p.payload)
	trailer := p.raw[end:]
	if len(trailer) < 2+suite.authCodeLen {
		return false
	}
	codeAt := len(p.raw) - suite.authCodeLen
	if p.raw[codeAt-1] != nextHeader {
		return false
	}
	padLen := int(p.raw[codeAt-2])
	if end+padLen+2 != codeAt {
		return false
	}
	for _, b := range p.raw[end : end+padLen] {
		if b != 0xff {
			return false
		}
	}
	want := suite.authCode(k1, p.raw[rmcpHeaderLen:codeAt])
	return hmac.Equal(want, p.raw[codeAt:])
}

// buildV20 frames an RMCP+ payload. With a suite and K1 it adds the
// integrity trailer: 0xff pad to a multiple of four bytes from the session
// header on, the pad length, Next Header, and the AuthCode.
func buildV20(h v20Header, payload []byte, suite *cipherSuite, k1 []byte) []byte {
	out := make([]byte, 0, v20PayloadAt+len(payload)+8+32)
	out = append(out, rmcpVersion, 0, rmcpNoAck, rmcpClassIPMI)
	out = append(out, authTypeRMCPPlus, h.payloadType)
	out = binary.LittleEndian.AppendUint32(out, h.sessionID)
	out = binary.LittleEndian.AppendUint32(out, h.seq)
	out = binary.LittleEndian.AppendUint16(out, uint16(len(payload)))
	out = append(out, payload...)

	if suite == nil {
		return out
	}
	padLen := (4 - (len(out)-rmcpHeaderLen+2)%4) % 4
	for i := 0; i < padLen; i++ {
		out = append(out, 0xff)
	}
	out = append(out, byte(padLen), nextHeader)
	return append(out, suite.authCode(k1, out[rmcpHeaderLen:])...)
}

// The IPMI message inside a payload, IPMI 2.0 section 13.8. The service is
// the BMC at slave address 0x20; the remote console software is 0x81.
const (
	bmcAddr     = 0x20
	consoleAddr = 0x81
)

type request struct {
	netFn byte
	rqSeq byte // the request sequence number, 6 bits
	rqLUN byte
	rsLUN byte
	cmd   byte
	data  []byte
}

func checksum(b []byte) byte {
	var sum byte
	for _, v := range b {
		sum += v
	}
	return -sum
}

func parseRequest(msg []byte) (*request, error) {
	if len(msg) < 7 {
		return nil, errShortPacket
	}
	if checksum(msg[:2]) != msg[2] || checksum(msg[3:len(msg)-1]) != msg[len(msg)-1] {
		return nil, errors.New("bad IPMI message checksum")
	}
	return &request{
		netFn: msg[1] >> 2,
		rsLUN: msg[1] & 0x03,
		rqSeq: msg[4] >> 2,
		rqLUN: msg[4] & 0x03,
		cmd:   msg[5],
		data:  msg[6 : len(msg)-1],
	}, nil
}

// buildResponse is the answer to req with a completion code and data.
func buildResponse(req *request, cc byte, data []byte) []byte {
	out := make([]byte, 0, 8+len(data))
	out = append(out, consoleAddr, (req.netFn|1)<<2|req.rqLUN)
	out = append(out, checksum(out[:2]))
	out = append(out, bmcAddr, req.rqSeq<<2|req.rsLUN, req.cmd, cc)
	out = append(out, data...)
	return append(out, checksum(out[3:]))
}
