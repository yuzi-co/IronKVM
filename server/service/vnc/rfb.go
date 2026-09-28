package vnc

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strconv"
)

// The protocol version this server speaks. A client that answers 3.7 is served
// as 3.7, and any later 3.x as 3.8. RFB 3.3 lets only the server pick the
// security type, which leaves no room for VeNCrypt, so it is refused.
const protocolVersion = "RFB 003.008\n"

// Security types (RFC 6143 section 7.2.1, and the VeNCrypt extension).
const (
	securityVNCAuth  byte = 2
	securityVeNCrypt byte = 19
)

// VeNCrypt X509Plain: TLS with the server's certificate, then a username and a
// password.
const vencryptX509Plain uint32 = 262

// Client to server message types.
const (
	msgSetPixelFormat           byte = 0
	msgSetEncodings             byte = 2
	msgFramebufferUpdateRequest byte = 3
	msgKeyEvent                 byte = 4
	msgPointerEvent             byte = 5
	msgClientCutText            byte = 6
)

// Server to client message types.
const msgFramebufferUpdate byte = 0

// Encodings. Tight is the only one the server sends pixels in; DesktopSize is
// the pseudo-encoding that tells the client the framebuffer changed size.
const (
	encodingTight       int32 = 7
	encodingDesktopSize int32 = -223
)

// maxEncodings bounds SetEncodings. Real clients list about twenty.
const maxEncodings = 1024

// maxCutText bounds the clipboard text the server reads and throws away.
const maxCutText = 1 << 20

// pixelFormat is the 16-byte PIXEL_FORMAT structure.
type pixelFormat struct {
	BitsPerPixel uint8
	Depth        uint8
	BigEndian    uint8
	TrueColour   uint8
	RedMax       uint16
	GreenMax     uint16
	BlueMax      uint16
	RedShift     uint8
	GreenShift   uint8
	BlueShift    uint8
	_            [3]byte
}

// serverPixelFormat is what ServerInit announces. The pixels travel as JPEG,
// which the client decodes into whatever format it asks for, so this only has
// to be a format every client accepts.
var serverPixelFormat = pixelFormat{
	BitsPerPixel: 32,
	Depth:        24,
	TrueColour:   1,
	RedMax:       255,
	GreenMax:     255,
	BlueMax:      255,
	RedShift:     16,
	GreenShift:   8,
	BlueShift:    0,
}

// jpegCapable reports whether a client in this pixel format can take a Tight
// JPEG rectangle. The Tight specification allows JPEG only for true colour at
// 16 or 32 bits a pixel.
func (p pixelFormat) jpegCapable() bool {
	return p.TrueColour != 0 && (p.BitsPerPixel == 16 || p.BitsPerPixel == 32)
}

// parseClientVersion reads the minor version from a ProtocolVersion message.
// It answers 7 or 8, the two versions the server speaks.
func parseClientVersion(msg []byte) (int, error) {
	if len(msg) != 12 || string(msg[:4]) != "RFB " || msg[7] != '.' || msg[11] != '\n' {
		return 0, fmt.Errorf("invalid protocol version %q", msg)
	}
	major, errMajor := strconv.Atoi(string(msg[4:7]))
	minor, errMinor := strconv.Atoi(string(msg[8:11]))
	if errMajor != nil || errMinor != nil {
		return 0, fmt.Errorf("invalid protocol version %q", msg)
	}
	if major != 3 || minor < 7 {
		return 0, fmt.Errorf("unsupported protocol version %d.%d", major, minor)
	}
	if minor > 8 {
		minor = 8
	}
	return minor, nil
}

// writeSecurityResult sends SecurityResult. RFB 3.8 follows a failure with the
// reason, and 3.7 does not.
func writeSecurityResult(w io.Writer, minor int, reason string) error {
	if reason == "" {
		return binary.Write(w, binary.BigEndian, uint32(0))
	}
	buf := binary.BigEndian.AppendUint32(nil, 1)
	if minor >= 8 {
		buf = binary.BigEndian.AppendUint32(buf, uint32(len(reason)))
		buf = append(buf, reason...)
	}
	_, err := w.Write(buf)
	return err
}

// serverInit builds the ServerInit message.
func serverInit(width, height int, name string) []byte {
	buf := make([]byte, 0, 24+len(name))
	buf = binary.BigEndian.AppendUint16(buf, uint16(width))
	buf = binary.BigEndian.AppendUint16(buf, uint16(height))
	buf = appendPixelFormat(buf, serverPixelFormat)
	buf = binary.BigEndian.AppendUint32(buf, uint32(len(name)))
	return append(buf, name...)
}

func appendPixelFormat(buf []byte, p pixelFormat) []byte {
	buf = append(buf, p.BitsPerPixel, p.Depth, p.BigEndian, p.TrueColour)
	buf = binary.BigEndian.AppendUint16(buf, p.RedMax)
	buf = binary.BigEndian.AppendUint16(buf, p.GreenMax)
	buf = binary.BigEndian.AppendUint16(buf, p.BlueMax)
	return append(buf, p.RedShift, p.GreenShift, p.BlueShift, 0, 0, 0)
}

func parsePixelFormat(b []byte) pixelFormat {
	return pixelFormat{
		BitsPerPixel: b[0],
		Depth:        b[1],
		BigEndian:    b[2],
		TrueColour:   b[3],
		RedMax:       binary.BigEndian.Uint16(b[4:]),
		GreenMax:     binary.BigEndian.Uint16(b[6:]),
		BlueMax:      binary.BigEndian.Uint16(b[8:]),
		RedShift:     b[10],
		GreenShift:   b[11],
		BlueShift:    b[12],
	}
}

// desktopSizeUpdate builds a FramebufferUpdate with the one DesktopSize
// pseudo-rectangle, which tells the client the framebuffer is now this size.
func desktopSizeUpdate(width, height int) []byte {
	buf := []byte{msgFramebufferUpdate, 0, 0, 1}
	return appendRectHeader(buf, 0, 0, width, height, encodingDesktopSize)
}

func appendRectHeader(buf []byte, x, y, width, height int, encoding int32) []byte {
	buf = binary.BigEndian.AppendUint16(buf, uint16(x))
	buf = binary.BigEndian.AppendUint16(buf, uint16(y))
	buf = binary.BigEndian.AppendUint16(buf, uint16(width))
	buf = binary.BigEndian.AppendUint16(buf, uint16(height))
	return binary.BigEndian.AppendUint32(buf, uint32(encoding))
}

// bufferedConn reads through the handshake's buffered reader, so a byte the
// reader took ahead of the TLS handshake is not lost to it.
type bufferedConn struct {
	net.Conn
	r io.Reader
}

func (c *bufferedConn) Read(p []byte) (int, error) {
	return c.r.Read(p)
}
