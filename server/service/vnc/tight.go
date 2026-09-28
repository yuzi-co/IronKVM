package vnc

import (
	"errors"
	"io"
	"net"
)

// tightJPEG is the Tight compression-control byte for a JPEG rectangle: the
// upper nibble 1001 selects JPEG, and the lower nibble resets no zlib stream,
// because none is ever used.
const tightJPEG byte = 0x90

// maxTightLength is the largest length the three-byte compact form can carry.
const maxTightLength = 1<<22 - 1

var errFrameTooLarge = errors.New("the JPEG frame is too large for a Tight rectangle")

// appendCompactLength appends a Tight compact length: seven bits a byte, low
// bits first, with the high bit set on every byte but the last. The third byte
// carries eight bits.
func appendCompactLength(buf []byte, n int) []byte {
	b := byte(n & 0x7f)
	if n <= 0x7f {
		return append(buf, b)
	}
	buf = append(buf, b|0x80)
	b = byte(n >> 7 & 0x7f)
	if n <= 0x3fff {
		return append(buf, b)
	}
	return append(buf, b|0x80, byte(n>>14))
}

// jpegUpdateHeader builds everything a FramebufferUpdate with one Tight JPEG
// rectangle over the whole framebuffer carries before the JPEG bytes.
func jpegUpdateHeader(width, height, length int) ([]byte, error) {
	if length > maxTightLength {
		return nil, errFrameTooLarge
	}
	buf := make([]byte, 0, 4+12+1+3)
	buf = append(buf, msgFramebufferUpdate, 0, 0, 1)
	buf = appendRectHeader(buf, 0, 0, width, height, encodingTight)
	buf = append(buf, tightJPEG)
	return appendCompactLength(buf, length), nil
}

// writeJPEGUpdate sends one frame. The JPEG goes to the socket as the encoder
// made it: the server never copies, decodes or re-encodes it.
func writeJPEGUpdate(w io.Writer, width, height int, frame []byte) (int64, error) {
	header, err := jpegUpdateHeader(width, height, len(frame))
	if err != nil {
		return 0, err
	}
	buffers := net.Buffers{header, frame}
	return buffers.WriteTo(w)
}

var errNotJPEG = errors.New("the frame is not a JPEG image")

// jpegSize reads the width and height from the frame header of a JPEG. It
// walks the marker segments up to the first start-of-frame and touches nothing
// after it, so it costs a few dozen byte reads a frame.
func jpegSize(data []byte) (width, height int, err error) {
	if len(data) < 4 || data[0] != 0xff || data[1] != 0xd8 {
		return 0, 0, errNotJPEG
	}
	i := 2
	for i+4 <= len(data) {
		if data[i] != 0xff {
			return 0, 0, errNotJPEG
		}
		marker := data[i+1]
		// A marker may be preceded by any number of fill bytes.
		if marker == 0xff {
			i++
			continue
		}
		// Markers without a length.
		if marker == 0x01 || (marker >= 0xd0 && marker <= 0xd7) {
			i += 2
			continue
		}
		length := int(data[i+2])<<8 | int(data[i+3])
		if length < 2 {
			return 0, 0, errNotJPEG
		}
		// SOF0 to SOF15, except DHT (C4), JPG (C8) and DAC (CC).
		if marker >= 0xc0 && marker <= 0xcf && marker != 0xc4 && marker != 0xc8 && marker != 0xcc {
			if i+9 > len(data) {
				return 0, 0, errNotJPEG
			}
			height = int(data[i+5])<<8 | int(data[i+6])
			width = int(data[i+7])<<8 | int(data[i+8])
			if width == 0 || height == 0 {
				return 0, 0, errNotJPEG
			}
			return width, height, nil
		}
		if marker == 0xda || marker == 0xd9 {
			break
		}
		i += 2 + length
	}
	return 0, 0, errNotJPEG
}
