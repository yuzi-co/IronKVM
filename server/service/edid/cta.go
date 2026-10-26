package edid

import "fmt"

// CTA is a decoded CTA-861 extension block.
type CTA struct {
	Block      int           `json:"block"`
	Revision   int           `json:"revision"`
	Underscan  bool          `json:"underscan"`
	BasicAudio bool          `json:"basicAudio"`
	YCbCr444   bool          `json:"ycbcr444"`
	YCbCr422   bool          `json:"ycbcr422"`
	NativeDTDs int           `json:"nativeDTDs"`
	VICs       []VIC         `json:"vics"`
	Audio      []AudioFormat `json:"audio"`
	Speakers   []string      `json:"speakers"`
	HDMI       *HDMIVSDB     `json:"hdmi"`
	HDMIForum  *HDMIForum    `json:"hdmiForum"`
	DataBlocks []DataBlock   `json:"dataBlocks"`
}

// VIC is one entry of a video data block. Width is zero for a code this
// package has no table entry for.
type VIC struct {
	Code       int     `json:"code"`
	Native     bool    `json:"native"`
	Width      int     `json:"width"`
	Height     int     `json:"height"`
	RefreshHz  float64 `json:"refreshHz"`
	Interlaced bool    `json:"interlaced"`
	Aspect     string  `json:"aspect"`
}

// AudioFormat is one short audio descriptor.
type AudioFormat struct {
	Format         string    `json:"format"`
	Channels       int       `json:"channels"`
	SampleRatesKHz []float64 `json:"sampleRatesKHz"`
	BitDepths      []int     `json:"bitDepths,omitempty"`
	MaxBitrateKbps int       `json:"maxBitrateKbps,omitempty"`
}

// HDMIVSDB is the HDMI 1.x vendor-specific data block (OUI 00-0C-03).
type HDMIVSDB struct {
	PhysicalAddress string `json:"physicalAddress"`
	SupportsAI      bool   `json:"supportsAI"`
	DeepColor30     bool   `json:"deepColor30"`
	DeepColor36     bool   `json:"deepColor36"`
	DeepColor48     bool   `json:"deepColor48"`
	DeepColorY444   bool   `json:"deepColorY444"`
	DualDVI         bool   `json:"dualDVI"`
	MaxTMDSMHz      int    `json:"maxTMDSMHz"`
}

// HDMIForum is the HDMI Forum vendor-specific data block (OUI C4-5D-D8).
type HDMIForum struct {
	Version       int  `json:"version"`
	MaxTMDSMHz    int  `json:"maxTMDSMHz"`
	SCDCPresent   bool `json:"scdcPresent"`
	Scrambling340 bool `json:"scrambling340"`
}

// DataBlock names one data block of the collection, decoded or not.
type DataBlock struct {
	Tag    int    `json:"tag"`
	ExtTag int    `json:"extTag,omitempty"`
	Name   string `json:"name"`
	Length int    `json:"length"`
}

var dataBlockNames = map[int]string{
	1: "Audio",
	2: "Video",
	3: "Vendor-specific",
	4: "Speaker allocation",
	5: "VESA display transfer characteristic",
	7: "Extended",
}

var extendedNames = map[int]string{
	0:   "Video capability",
	1:   "Vendor-specific video",
	2:   "VESA display device",
	3:   "VESA video timing",
	5:   "Colorimetry",
	6:   "HDR static metadata",
	7:   "HDR dynamic metadata",
	13:  "Video format preference",
	14:  "YCbCr 4:2:0 video",
	15:  "YCbCr 4:2:0 capability map",
	17:  "Vendor-specific audio",
	18:  "Room configuration",
	19:  "Speaker location",
	32:  "InfoFrame",
	120: "HDMI Forum EDID extension override",
	121: "HDMI Forum sink capability",
}

var audioFormats = []string{
	"reserved", "LPCM", "AC-3", "MPEG-1", "MP3", "MPEG-2", "AAC LC", "DTS",
	"ATRAC", "DSD", "E-AC-3", "DTS-HD", "MAT", "DST", "WMA Pro", "extended",
}

var sampleRates = []float64{32, 44.1, 48, 88.2, 96, 176.4, 192}

var speakerNames = []string{
	"FL/FR", "LFE", "FC", "RL/RR", "RC", "FLC/FRC", "RLC/RRC", // byte 1
	"FLW/FRW", "TpFL/TpFR", "TpC", "TpFC", // byte 2
}

func (info *Info) parseCTA(index int, b []byte) {
	cta := &CTA{
		Block:      index,
		Revision:   int(b[1]),
		VICs:       []VIC{},
		Audio:      []AudioFormat{},
		Speakers:   []string{},
		DataBlocks: []DataBlock{},
	}
	info.CTA = cta

	dtdOffset := int(b[2])
	if cta.Revision >= 2 {
		cta.Underscan = b[3]&0x80 != 0
		cta.BasicAudio = b[3]&0x40 != 0
		cta.YCbCr444 = b[3]&0x20 != 0
		cta.YCbCr422 = b[3]&0x10 != 0
		cta.NativeDTDs = int(b[3] & 0x0f)
	}
	if dtdOffset == 0 {
		return
	}
	if dtdOffset < 4 || dtdOffset > 127 {
		info.problem("CTA block %d: detailed timing offset %d is outside the block", index, dtdOffset)
		return
	}

	for p := 4; p < dtdOffset; {
		tag, length := int(b[p]>>5), int(b[p]&0x1f)
		if p+1+length > dtdOffset {
			info.problem("CTA block %d: data block at byte %d (length %d) runs past the detailed timing offset %d", index, p, length, dtdOffset)
			break
		}
		payload := b[p+1 : p+1+length]
		cta.dataBlock(tag, payload)
		p += 1 + length
	}

	for p := dtdOffset; p+18 <= 127; p += 18 {
		d := b[p : p+18]
		if d[0] == 0 && d[1] == 0 {
			break
		}
		info.DetailedTimings = append(info.DetailedTimings, detailedTiming(d, index))
	}
}

func (cta *CTA) dataBlock(tag int, payload []byte) {
	block := DataBlock{Tag: tag, Name: dataBlockNames[tag], Length: len(payload)}
	switch tag {
	case 1:
		for p := 0; p+3 <= len(payload); p += 3 {
			cta.Audio = append(cta.Audio, shortAudio(payload[p:p+3]))
		}
	case 2:
		for _, v := range payload {
			cta.VICs = append(cta.VICs, vic(v))
		}
	case 3:
		cta.vendor(payload, &block)
	case 4:
		if len(payload) >= 1 {
			for bit := 0; bit < 7; bit++ {
				if payload[0]&(1<<bit) != 0 {
					cta.Speakers = append(cta.Speakers, speakerNames[bit])
				}
			}
		}
		if len(payload) >= 2 {
			for bit := 0; bit < 4; bit++ {
				if payload[1]&(1<<bit) != 0 {
					cta.Speakers = append(cta.Speakers, speakerNames[7+bit])
				}
			}
		}
	case 7:
		if len(payload) >= 1 {
			block.ExtTag = int(payload[0])
			block.Length--
			if name := extendedNames[block.ExtTag]; name != "" {
				block.Name = name
			} else {
				block.Name = fmt.Sprintf("Extended tag %d", block.ExtTag)
			}
		}
	}
	if block.Name == "" {
		block.Name = fmt.Sprintf("Tag %d", tag)
	}
	cta.DataBlocks = append(cta.DataBlocks, block)
}

func (cta *CTA) vendor(p []byte, block *DataBlock) {
	if len(p) < 3 {
		return
	}
	oui := uint32(p[0]) | uint32(p[1])<<8 | uint32(p[2])<<16
	switch oui {
	case 0x000c03:
		block.Name = "HDMI 1.x"
		if len(p) < 5 {
			return
		}
		h := &HDMIVSDB{PhysicalAddress: fmt.Sprintf("%d.%d.%d.%d", p[3]>>4, p[3]&15, p[4]>>4, p[4]&15)}
		if len(p) >= 6 {
			h.SupportsAI = p[5]&0x80 != 0
			h.DeepColor48 = p[5]&0x40 != 0
			h.DeepColor36 = p[5]&0x20 != 0
			h.DeepColor30 = p[5]&0x10 != 0
			h.DeepColorY444 = p[5]&0x08 != 0
			h.DualDVI = p[5]&0x01 != 0
		}
		if len(p) >= 7 {
			h.MaxTMDSMHz = int(p[6]) * 5
		}
		cta.HDMI = h
	case 0xc45dd8:
		block.Name = "HDMI Forum"
		if len(p) < 6 {
			return
		}
		cta.HDMIForum = &HDMIForum{
			Version:       int(p[3]),
			MaxTMDSMHz:    int(p[4]) * 5,
			SCDCPresent:   p[5]&0x80 != 0,
			Scrambling340: p[5]&0x08 != 0,
		}
	default:
		block.Name = fmt.Sprintf("Vendor-specific (OUI %02X-%02X-%02X)", p[2], p[1], p[0])
	}
}

func shortAudio(p []byte) AudioFormat {
	code := int(p[0]>>3) & 0x0f
	a := AudioFormat{Format: audioFormats[code], Channels: int(p[0]&7) + 1, SampleRatesKHz: []float64{}}
	for bit, rate := range sampleRates {
		if p[1]&(1<<bit) != 0 {
			a.SampleRatesKHz = append(a.SampleRatesKHz, rate)
		}
	}
	switch {
	case code == 1:
		for bit, depth := range []int{16, 20, 24} {
			if p[2]&(1<<bit) != 0 {
				a.BitDepths = append(a.BitDepths, depth)
			}
		}
	case code >= 2 && code <= 8:
		a.MaxBitrateKbps = int(p[2]) * 8
	}
	return a
}

// vic decodes one byte of a video data block. From CTA-861-F on, bit 7 of
// codes 129 to 192 marks a native mode of VIC 1 to 64; codes 193 and above
// are VICs as they stand.
func vic(b byte) VIC {
	v := VIC{Code: int(b)}
	if b >= 129 && b <= 192 {
		v.Code, v.Native = int(b&0x7f), true
	}
	if v.Code < len(vicTable) {
		m := vicTable[v.Code]
		v.Width, v.Height, v.RefreshHz, v.Interlaced, v.Aspect = m.w, m.h, m.hz, m.i, m.aspect
	}
	return v
}

type vicMode struct {
	w, h   int
	hz     float64
	i      bool
	aspect string
}

// vicTable is CTA-861-G table 3, VIC 1 to 107. A rate of 60, 30 or 24 also
// stands for its 1000/1001 twin. Pixel-repeated codes list their active
// width as sent.
var vicTable = func() []vicMode {
	t := make([]vicMode, 108)
	set := func(code, w, h int, hz float64, i bool, aspect string) {
		t[code] = vicMode{w, h, hz, i, aspect}
	}
	set(1, 640, 480, 60, false, "4:3")
	set(2, 720, 480, 60, false, "4:3")
	set(3, 720, 480, 60, false, "16:9")
	set(4, 1280, 720, 60, false, "16:9")
	set(5, 1920, 1080, 60, true, "16:9")
	set(6, 1440, 480, 60, true, "4:3")
	set(7, 1440, 480, 60, true, "16:9")
	set(8, 1440, 240, 60, false, "4:3")
	set(9, 1440, 240, 60, false, "16:9")
	set(10, 2880, 480, 60, true, "4:3")
	set(11, 2880, 480, 60, true, "16:9")
	set(12, 2880, 240, 60, false, "4:3")
	set(13, 2880, 240, 60, false, "16:9")
	set(14, 1440, 480, 60, false, "4:3")
	set(15, 1440, 480, 60, false, "16:9")
	set(16, 1920, 1080, 60, false, "16:9")
	set(17, 720, 576, 50, false, "4:3")
	set(18, 720, 576, 50, false, "16:9")
	set(19, 1280, 720, 50, false, "16:9")
	set(20, 1920, 1080, 50, true, "16:9")
	set(21, 1440, 576, 50, true, "4:3")
	set(22, 1440, 576, 50, true, "16:9")
	set(23, 1440, 288, 50, false, "4:3")
	set(24, 1440, 288, 50, false, "16:9")
	set(25, 2880, 576, 50, true, "4:3")
	set(26, 2880, 576, 50, true, "16:9")
	set(27, 2880, 288, 50, false, "4:3")
	set(28, 2880, 288, 50, false, "16:9")
	set(29, 1440, 576, 50, false, "4:3")
	set(30, 1440, 576, 50, false, "16:9")
	set(31, 1920, 1080, 50, false, "16:9")
	set(32, 1920, 1080, 24, false, "16:9")
	set(33, 1920, 1080, 25, false, "16:9")
	set(34, 1920, 1080, 30, false, "16:9")
	set(35, 2880, 480, 60, false, "4:3")
	set(36, 2880, 480, 60, false, "16:9")
	set(37, 2880, 576, 50, false, "4:3")
	set(38, 2880, 576, 50, false, "16:9")
	set(39, 1920, 1080, 50, true, "16:9")
	set(40, 1920, 1080, 100, true, "16:9")
	set(41, 1280, 720, 100, false, "16:9")
	set(42, 720, 576, 100, false, "4:3")
	set(43, 720, 576, 100, false, "16:9")
	set(44, 1440, 576, 100, true, "4:3")
	set(45, 1440, 576, 100, true, "16:9")
	set(46, 1920, 1080, 120, true, "16:9")
	set(47, 1280, 720, 120, false, "16:9")
	set(48, 720, 480, 120, false, "4:3")
	set(49, 720, 480, 120, false, "16:9")
	set(50, 1440, 480, 120, true, "4:3")
	set(51, 1440, 480, 120, true, "16:9")
	set(52, 720, 576, 200, false, "4:3")
	set(53, 720, 576, 200, false, "16:9")
	set(54, 1440, 576, 200, true, "4:3")
	set(55, 1440, 576, 200, true, "16:9")
	set(56, 720, 480, 240, false, "4:3")
	set(57, 720, 480, 240, false, "16:9")
	set(58, 1440, 480, 240, true, "4:3")
	set(59, 1440, 480, 240, true, "16:9")
	set(60, 1280, 720, 24, false, "16:9")
	set(61, 1280, 720, 25, false, "16:9")
	set(62, 1280, 720, 30, false, "16:9")
	set(63, 1920, 1080, 120, false, "16:9")
	set(64, 1920, 1080, 100, false, "16:9")
	rates := []float64{24, 25, 30, 50, 60, 100, 120}
	for k, hz := range rates {
		set(65+k, 1280, 720, hz, false, "64:27")
		set(72+k, 1920, 1080, hz, false, "64:27")
		set(79+k, 1680, 720, hz, false, "64:27")
		set(86+k, 2560, 1080, hz, false, "64:27")
	}
	for k, hz := range []float64{24, 25, 30, 50, 60} {
		set(93+k, 3840, 2160, hz, false, "16:9")
		set(98+k, 4096, 2160, hz, false, "256:135")
		set(103+k, 3840, 2160, hz, false, "64:27")
	}
	return t
}()
