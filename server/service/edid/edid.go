// Package edid decodes EDIDs (VESA E-EDID 1.3/1.4 base block and CTA-861
// extensions) and holds the presets the web UI shows (ironkvm-dist#95).
//
// It is read-only towards the hardware: nothing here talks to the HDMI bridge.
// The LT6911 keeps its EDID in its own SPI flash, reachable only over I2C with
// register writes, and writing it is parked until an operator decides on it
// (ironkvm-dist#95, trial 79).
package edid

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

// BlockSize is the size of every EDID block.
const BlockSize = 128

// MaxSize is the largest EDID: a base block and 255 extensions.
const MaxSize = 256 * BlockSize

var (
	ErrTruncated = errors.New("shorter than one 128-byte EDID block")
	ErrNotEDID   = errors.New("not an EDID: the 8-byte header is missing")
	ErrTooLarge  = fmt.Errorf("larger than %d bytes, the most an EDID can hold", MaxSize)
)

var header = []byte{0x00, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0x00}

// Info is a decoded EDID. Problems lists what makes it unfit to hand to a
// host (a bad checksum, a missing extension, a malformed block); Valid is true
// only when there are none. A problem does not stop the decoding, so a broken
// file can still be inspected.
type Info struct {
	Size      int        `json:"size"`
	Blocks    int        `json:"blocks"`
	Valid     bool       `json:"valid"`
	Problems  []string   `json:"problems"`
	Checksums []Checksum `json:"checksums"`

	Version      string `json:"version"`
	Manufacturer string `json:"manufacturer"`
	ProductCode  uint16 `json:"productCode"`
	SerialNumber uint32 `json:"serialNumber"`
	SerialText   string `json:"serialText,omitempty"`
	Name         string `json:"name,omitempty"`
	Text         string `json:"text,omitempty"`
	Week         int    `json:"week"`
	Year         int    `json:"year"`
	ModelYear    bool   `json:"modelYear"`

	Digital   bool    `json:"digital"`
	BitDepth  int     `json:"bitDepth,omitempty"`
	Interface string  `json:"interface,omitempty"`
	WidthCm   int     `json:"widthCm"`
	HeightCm  int     `json:"heightCm"`
	Gamma     float64 `json:"gamma,omitempty"`

	Preferred          *Timing      `json:"preferred"`
	DetailedTimings    []Timing     `json:"detailedTimings"`
	EstablishedTimings []Mode       `json:"establishedTimings"`
	StandardTimings    []Mode       `json:"standardTimings"`
	RangeLimits        *RangeLimits `json:"rangeLimits"`

	Extensions []Extension `json:"extensions"`
	CTA        *CTA        `json:"cta"`
}

// Checksum is one block's checksum byte against the one its other 127 bytes
// call for.
type Checksum struct {
	Block    int  `json:"block"`
	Stored   byte `json:"stored"`
	Computed byte `json:"computed"`
	Valid    bool `json:"valid"`
}

// Timing is a detailed timing descriptor. For an interlaced timing the
// vertical numbers are per field and RefreshHz is the field rate.
type Timing struct {
	Block         int     `json:"block"`
	PixelClockKHz int     `json:"pixelClockKHz"`
	HActive       int     `json:"hActive"`
	HBlank        int     `json:"hBlank"`
	HFrontPorch   int     `json:"hFrontPorch"`
	HSync         int     `json:"hSync"`
	VActive       int     `json:"vActive"`
	VBlank        int     `json:"vBlank"`
	VFrontPorch   int     `json:"vFrontPorch"`
	VSync         int     `json:"vSync"`
	WidthMm       int     `json:"widthMm"`
	HeightMm      int     `json:"heightMm"`
	Interlaced    bool    `json:"interlaced"`
	HSyncPositive bool    `json:"hSyncPositive"`
	VSyncPositive bool    `json:"vSyncPositive"`
	RefreshHz     float64 `json:"refreshHz"`
	HFreqKHz      float64 `json:"hFreqKHz"`
}

// Mode is a resolution and rate from a list (established, standard, VIC).
type Mode struct {
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	RefreshHz  float64 `json:"refreshHz"`
	Interlaced bool    `json:"interlaced,omitempty"`
}

// RangeLimits is the display range limits descriptor.
type RangeLimits struct {
	MinVHz           int `json:"minVHz"`
	MaxVHz           int `json:"maxVHz"`
	MinHKHz          int `json:"minHKHz"`
	MaxHKHz          int `json:"maxHKHz"`
	MaxPixelClockMHz int `json:"maxPixelClockMHz"`
}

// Extension names one extension block by its tag.
type Extension struct {
	Block int    `json:"block"`
	Tag   byte   `json:"tag"`
	Name  string `json:"name"`
}

var extensionNames = map[byte]string{
	0x02: "CTA-861",
	0x10: "Video Timing Block",
	0x40: "Display Information",
	0x50: "Localized String",
	0x60: "Digital Packet Video Link",
	0x70: "DisplayID",
	0xf0: "Block Map",
	0xff: "Manufacturer",
}

// Parse decodes data. It fails only when data cannot be an EDID at all; any
// other fault is a Problem on the result.
func Parse(data []byte) (*Info, error) {
	if len(data) > MaxSize {
		return nil, ErrTooLarge
	}
	if len(data) < BlockSize {
		return nil, ErrTruncated
	}
	if string(data[:8]) != string(header) {
		return nil, ErrNotEDID
	}

	info := &Info{
		Size:               len(data),
		Problems:           []string{},
		Checksums:          []Checksum{},
		DetailedTimings:    []Timing{},
		EstablishedTimings: []Mode{},
		StandardTimings:    []Mode{},
		Extensions:         []Extension{},
	}

	blocks := len(data) / BlockSize
	if rest := len(data) % BlockSize; rest != 0 {
		info.problem("%d trailing bytes after block %d are not a whole block and were ignored", rest, blocks-1)
	}
	declared := int(data[126])
	if blocks-1 < declared {
		info.problem("the base block declares %d extension block(s), the data holds %d", declared, blocks-1)
	} else if blocks-1 > declared {
		info.problem("the data holds %d extension block(s) beyond the %d the base block declares; they were ignored", blocks-1-declared, declared)
		blocks = 1 + declared
	}
	info.Blocks = blocks

	for b := 0; b < blocks; b++ {
		block := data[b*BlockSize : (b+1)*BlockSize]
		c := Checksum{Block: b, Stored: block[127], Computed: checksum(block)}
		c.Valid = c.Stored == c.Computed
		if !c.Valid {
			info.problem("block %d checksum is %#02x, the data needs %#02x", b, c.Stored, c.Computed)
		}
		info.Checksums = append(info.Checksums, c)
	}

	info.parseBase(data[:BlockSize])
	for b := 1; b < blocks; b++ {
		block := data[b*BlockSize : (b+1)*BlockSize]
		ext := Extension{Block: b, Tag: block[0], Name: extensionNames[block[0]]}
		if ext.Name == "" {
			ext.Name = fmt.Sprintf("unknown (%#02x)", block[0])
		}
		info.Extensions = append(info.Extensions, ext)
		if block[0] == 0x02 && info.CTA == nil {
			info.parseCTA(b, block)
		}
	}

	info.Valid = len(info.Problems) == 0
	return info, nil
}

func (info *Info) problem(format string, args ...any) {
	info.Problems = append(info.Problems, fmt.Sprintf(format, args...))
}

// checksum is the byte that makes the block's 128 bytes sum to zero.
func checksum(block []byte) byte {
	var sum byte
	for _, b := range block[:127] {
		sum += b
	}
	return -sum
}

var interfaces = map[byte]string{1: "DVI", 2: "HDMI-a", 3: "HDMI-b", 4: "MDDI", 5: "DisplayPort"}

func (info *Info) parseBase(b []byte) {
	info.Version = fmt.Sprintf("%d.%d", b[18], b[19])
	info.Manufacturer = manufacturer(binary.BigEndian.Uint16(b[8:10]))
	info.ProductCode = binary.LittleEndian.Uint16(b[10:12])
	info.SerialNumber = binary.LittleEndian.Uint32(b[12:16])
	info.Year = int(b[17]) + 1990
	if b[16] == 0xff {
		info.ModelYear = true
	} else {
		info.Week = int(b[16])
	}

	input := b[20]
	info.Digital = input&0x80 != 0
	if info.Digital && b[18] == 1 && b[19] >= 4 {
		if depth := (input >> 4) & 7; depth >= 1 && depth <= 6 {
			info.BitDepth = 4 + 2*int(depth)
		}
		info.Interface = interfaces[input&0x0f]
	}
	info.WidthCm, info.HeightCm = int(b[21]), int(b[22])
	if b[23] != 0xff {
		info.Gamma = float64(int(b[23])+100) / 100
	}

	for i, m := range establishedModes {
		if b[35+i/8]&(0x80>>(i%8)) != 0 {
			info.EstablishedTimings = append(info.EstablishedTimings, m)
		}
	}
	for p := 38; p < 54; p += 2 {
		if m, ok := standardTiming(b[p], b[p+1], b[19]); ok {
			info.StandardTimings = append(info.StandardTimings, m)
		}
	}

	for p := 54; p < 126; p += 18 {
		d := b[p : p+18]
		if d[0] != 0 || d[1] != 0 {
			t := detailedTiming(d, 0)
			info.DetailedTimings = append(info.DetailedTimings, t)
			if info.Preferred == nil && p == 54 {
				info.Preferred = &t
			}
			continue
		}
		info.parseDescriptor(d)
	}
}

func (info *Info) parseDescriptor(d []byte) {
	switch d[3] {
	case 0xff:
		info.SerialText = descriptorText(d)
	case 0xfe:
		info.Text = descriptorText(d)
	case 0xfc:
		info.Name = descriptorText(d)
	case 0xfd:
		r := &RangeLimits{
			MinVHz: int(d[5]), MaxVHz: int(d[6]),
			MinHKHz: int(d[7]), MaxHKHz: int(d[8]),
			MaxPixelClockMHz: int(d[9]) * 10,
		}
		// EDID 1.4 offsets: each limit may add 255.
		if d[4]&0x03 == 0x03 {
			r.MinVHz += 255
		}
		if d[4]&0x02 != 0 {
			r.MaxVHz += 255
		}
		if d[4]&0x0c == 0x0c {
			r.MinHKHz += 255
		}
		if d[4]&0x08 != 0 {
			r.MaxHKHz += 255
		}
		info.RangeLimits = r
	case 0xfa:
		for p := 5; p+1 < 17; p += 2 {
			if m, ok := standardTiming(d[p], d[p+1], 4); ok {
				info.StandardTimings = append(info.StandardTimings, m)
			}
		}
	}
}

// descriptorText is the text of a display descriptor: up to 13 characters,
// ended by a line feed and padded with spaces.
func descriptorText(d []byte) string {
	text := d[5:18]
	if i := strings.IndexByte(string(text), '\n'); i >= 0 {
		text = text[:i]
	}
	return strings.TrimRight(strings.Map(func(r rune) rune {
		if r < 0x20 || r > 0x7e {
			return -1
		}
		return r
	}, string(text)), " ")
}

func manufacturer(id uint16) string {
	letter := func(v uint16) byte {
		v &= 0x1f
		if v < 1 || v > 26 {
			return '?'
		}
		return byte('A' + v - 1)
	}
	return string([]byte{letter(id >> 10), letter(id >> 5), letter(id)})
}

// establishedModes are the bits of bytes 35 to 37, most significant first.
// Only the first seventeen are defined; the rest are the manufacturer's.
var establishedModes = []Mode{
	{Width: 720, Height: 400, RefreshHz: 70},
	{Width: 720, Height: 400, RefreshHz: 88},
	{Width: 640, Height: 480, RefreshHz: 60},
	{Width: 640, Height: 480, RefreshHz: 67},
	{Width: 640, Height: 480, RefreshHz: 72},
	{Width: 640, Height: 480, RefreshHz: 75},
	{Width: 800, Height: 600, RefreshHz: 56},
	{Width: 800, Height: 600, RefreshHz: 60},
	{Width: 800, Height: 600, RefreshHz: 72},
	{Width: 800, Height: 600, RefreshHz: 75},
	{Width: 832, Height: 624, RefreshHz: 75},
	{Width: 1024, Height: 768, RefreshHz: 87, Interlaced: true},
	{Width: 1024, Height: 768, RefreshHz: 60},
	{Width: 1024, Height: 768, RefreshHz: 70},
	{Width: 1024, Height: 768, RefreshHz: 75},
	{Width: 1280, Height: 1024, RefreshHz: 75},
	{Width: 1152, Height: 870, RefreshHz: 75},
}

// standardTiming decodes a two-byte standard timing. Aspect code 0 is 16:10
// from EDID 1.3 on and 1:1 before.
func standardTiming(b0, b1, revision byte) (Mode, bool) {
	if (b0 == 0x01 && b1 == 0x01) || (b0 == 0x00 && b1 == 0x00) || b0 == 0x00 {
		return Mode{}, false
	}
	w := (int(b0) + 31) * 8
	var h int
	switch b1 >> 6 {
	case 0:
		if revision < 3 {
			h = w
		} else {
			h = w * 10 / 16
		}
	case 1:
		h = w * 3 / 4
	case 2:
		h = w * 4 / 5
	case 3:
		h = w * 9 / 16
	}
	return Mode{Width: w, Height: h, RefreshHz: float64(int(b1&0x3f) + 60)}, true
}

func detailedTiming(d []byte, block int) Timing {
	t := Timing{
		Block:         block,
		PixelClockKHz: int(binary.LittleEndian.Uint16(d[0:2])) * 10,
		HActive:       int(d[2]) | int(d[4]&0xf0)<<4,
		HBlank:        int(d[3]) | int(d[4]&0x0f)<<8,
		VActive:       int(d[5]) | int(d[7]&0xf0)<<4,
		VBlank:        int(d[6]) | int(d[7]&0x0f)<<8,
		HFrontPorch:   int(d[8]) | int(d[11]&0xc0)<<2,
		HSync:         int(d[9]) | int(d[11]&0x30)<<4,
		VFrontPorch:   int(d[10]>>4) | int(d[11]&0x0c)<<2,
		VSync:         int(d[10]&0x0f) | int(d[11]&0x03)<<4,
		WidthMm:       int(d[12]) | int(d[14]&0xf0)<<4,
		HeightMm:      int(d[13]) | int(d[14]&0x0f)<<8,
		Interlaced:    d[17]&0x80 != 0,
	}
	// Digital separate sync carries both polarities; digital composite only
	// the horizontal one (bit 1); analog sync carries none.
	switch (d[17] >> 3) & 3 {
	case 3:
		t.VSyncPositive = d[17]&0x04 != 0
		t.HSyncPositive = d[17]&0x02 != 0
	case 2:
		t.HSyncPositive = d[17]&0x02 != 0
	}
	htotal, vtotal := t.HActive+t.HBlank, t.VActive+t.VBlank
	if htotal > 0 && vtotal > 0 {
		hz := float64(t.PixelClockKHz) * 1000 / float64(htotal*vtotal)
		t.RefreshHz = round3(hz)
		t.HFreqKHz = round3(float64(t.PixelClockKHz) / float64(htotal))
	}
	return t
}

func round3(v float64) float64 { return math.Round(v*1000) / 1000 }

// AllModes lists every mode an EDID offers, from every list, as frame sizes
// (an interlaced detailed timing counts both fields).
func AllModes(info *Info) []Mode {
	var out []Mode
	out = append(out, info.EstablishedTimings...)
	out = append(out, info.StandardTimings...)
	for _, t := range info.DetailedTimings {
		m := Mode{Width: t.HActive, Height: t.VActive, RefreshHz: t.RefreshHz, Interlaced: t.Interlaced}
		if t.Interlaced {
			m.Height *= 2
		}
		out = append(out, m)
	}
	if info.CTA != nil {
		for _, v := range info.CTA.VICs {
			if v.Width == 0 {
				continue
			}
			out = append(out, Mode{Width: v.Width, Height: v.Height, RefreshHz: v.RefreshHz, Interlaced: v.Interlaced})
		}
	}
	return out
}
