package edid

import (
	"encoding/binary"
	"math"
)

// Spec describes an EDID that Build writes: an E-EDID 1.3 base block and one
// CTA-861 extension, the same layout as the EDID the bridge ships with.
type Spec struct {
	// Manufacturer is a three-letter PNP ID.
	Manufacturer string
	ProductCode  uint16
	// Name is the monitor name a host shows, at most 13 characters.
	Name              string
	WidthCm, HeightCm int
	// Timings go into the base block in order; the first is the preferred
	// mode. At most two, so the name and range limits fit beside them.
	Timings []Timing
	// Established lists modes from the established timings bitmap.
	Established []Mode
	Range       RangeLimits
	// VICs go into the video data block in order. None is marked native:
	// the preferred detailed timing is the native mode.
	VICs []int
	// Audio declares two-channel LPCM at 32, 44.1 and 48 kHz, 16 and 20 bit,
	// front left and right: what the LT6911UXC's own EDID declares.
	Audio bool
}

// Build writes spec as a 256-byte EDID with valid checksums. It panics on a
// spec it cannot encode: specs are constants in this package, checked by
// tests, never input.
func Build(spec Spec) []byte {
	out := make([]byte, 2*BlockSize)
	base, cta := out[:BlockSize], out[BlockSize:]

	copy(base, header)
	if len(spec.Manufacturer) != 3 {
		panic("edid: manufacturer must be three letters")
	}
	var id uint16
	for _, c := range []byte(spec.Manufacturer) {
		if c < 'A' || c > 'Z' {
			panic("edid: manufacturer must be upper-case letters")
		}
		id = id<<5 | uint16(c-'A'+1)
	}
	binary.BigEndian.PutUint16(base[8:], id)
	binary.LittleEndian.PutUint16(base[10:], spec.ProductCode)
	// Serial number 0 and week 0: unknown. Year of manufacture 2026.
	base[17] = 2026 - 1990
	base[18], base[19] = 1, 3
	base[20] = 0x80 // digital input
	base[21], base[22] = byte(spec.WidthCm), byte(spec.HeightCm)
	base[23] = 120  // gamma 2.2
	base[24] = 0x0e // RGB colour, sRGB by default, preferred timing in the first descriptor
	putChromaticity(base[25:35])

	for _, m := range spec.Established {
		found := false
		for i, e := range establishedModes {
			if e == m {
				base[35+i/8] |= 0x80 >> (i % 8)
				found = true
			}
		}
		if !found {
			panic("edid: not an established timing")
		}
	}
	for p := 38; p < 54; p++ {
		base[p] = 0x01 // no standard timings
	}

	if len(spec.Timings) < 1 || len(spec.Timings) > 2 {
		panic("edid: one or two detailed timings")
	}
	p := 54
	for _, t := range spec.Timings {
		putTiming(base[p:p+18], t)
		p += 18
	}
	putRange(base[p:p+18], spec.Range)
	p += 18
	putText(base[p:p+18], 0xfc, spec.Name)
	p += 18
	for ; p < 126; p += 18 {
		base[p+3] = 0x10 // dummy descriptor
	}
	base[126] = 1
	base[127] = checksum(base)

	var blocks []byte
	video := []byte{byte(2<<5 | len(spec.VICs))}
	for _, v := range spec.VICs {
		if v < 1 || v > 127 {
			panic("edid: VIC outside 1 to 127")
		}
		video = append(video, byte(v))
	}
	blocks = append(blocks, video...)
	if spec.Audio {
		blocks = append(blocks, 1<<5|3, 0x09, 0x07, 0x03) // LPCM, 2 ch, 32/44.1/48 kHz, 16/20 bit
		blocks = append(blocks, 4<<5|3, 0x01, 0x00, 0x00) // FL/FR
	}
	var maxClock int
	for _, t := range spec.Timings {
		maxClock = max(maxClock, t.PixelClockKHz)
	}
	tmds := int(math.Ceil(float64(maxClock) / 5000))
	// HDMI 1.x: OUI 00-0C-03, physical address 1.0.0.0, no deep colour, the
	// highest TMDS clock this EDID needs.
	blocks = append(blocks, 3<<5|7, 0x03, 0x0c, 0x00, 0x10, 0x00, 0x00, byte(tmds))
	// Video capability: RGB quantization range selectable, every format
	// underscanned (IT and CE; PT follows them), so a host can send full range
	// and the whole picture.
	blocks = append(blocks, 7<<5|2, 0x00, 0x4a)

	cta[0], cta[1] = 0x02, 3
	cta[2] = byte(4 + len(blocks))
	// Underscanned by default, one native detailed timing: the preferred one.
	cta[3] = 0x80 | 1
	if spec.Audio {
		cta[3] |= 0x40
	}
	copy(cta[4:], blocks)
	cta[127] = checksum(cta)
	return out
}

func putTiming(d []byte, t Timing) {
	clock := t.PixelClockKHz / 10
	binary.LittleEndian.PutUint16(d[0:], uint16(clock))
	d[2] = byte(t.HActive)
	d[3] = byte(t.HBlank)
	d[4] = byte(t.HActive>>8)<<4 | byte(t.HBlank>>8)
	d[5] = byte(t.VActive)
	d[6] = byte(t.VBlank)
	d[7] = byte(t.VActive>>8)<<4 | byte(t.VBlank>>8)
	d[8] = byte(t.HFrontPorch)
	d[9] = byte(t.HSync)
	d[10] = byte(t.VFrontPorch&15)<<4 | byte(t.VSync&15)
	d[11] = byte(t.HFrontPorch>>8)<<6 | byte(t.HSync>>8)<<4 | byte(t.VFrontPorch>>4)<<2 | byte(t.VSync>>4)
	d[12] = byte(t.WidthMm)
	d[13] = byte(t.HeightMm)
	d[14] = byte(t.WidthMm>>8)<<4 | byte(t.HeightMm>>8)
	flags := byte(0x18) // digital separate sync
	if t.VSyncPositive {
		flags |= 0x04
	}
	if t.HSyncPositive {
		flags |= 0x02
	}
	if t.Interlaced {
		flags |= 0x80
	}
	d[17] = flags
}

func putRange(d []byte, r RangeLimits) {
	d[3] = 0xfd
	d[5], d[6] = byte(r.MinVHz), byte(r.MaxVHz)
	d[7], d[8] = byte(r.MinHKHz), byte(r.MaxHKHz)
	d[9] = byte((r.MaxPixelClockMHz + 9) / 10)
	d[10] = 0x00 // default GTF; the base block does not claim GTF support
	d[11] = 0x0a
	for i := 12; i < 18; i++ {
		d[i] = 0x20
	}
}

func putText(d []byte, tag byte, text string) {
	if len(text) > 13 {
		panic("edid: descriptor text longer than 13 characters")
	}
	d[3] = tag
	n := copy(d[5:], text)
	if n < 13 {
		d[5+n] = 0x0a
		for i := 5 + n + 1; i < 18; i++ {
			d[i] = 0x20
		}
	}
}

// putChromaticity writes the sRGB primaries and D65 white point, each as a
// 10-bit fraction of 1024.
func putChromaticity(d []byte) {
	v := func(c float64) int { return int(math.Round(c * 1024)) }
	rx, ry := v(0.640), v(0.330)
	gx, gy := v(0.300), v(0.600)
	bx, by := v(0.150), v(0.060)
	wx, wy := v(0.3127), v(0.3290)
	d[0] = byte((rx&3)<<6 | (ry&3)<<4 | (gx&3)<<2 | gy&3)
	d[1] = byte((bx&3)<<6 | (by&3)<<4 | (wx&3)<<2 | wy&3)
	for i, c := range []int{rx, ry, gx, gy, bx, by, wx, wy} {
		d[2+i] = byte(c >> 2)
	}
}
