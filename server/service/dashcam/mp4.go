package dashcam

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h265"

	"NanoKVM-Server/common"
	"NanoKVM-Server/service/stream/egress"
)

// The MP4 writer is this file's own rather than mediacommon's pmp4, although
// the server links mediacommon. pmp4 builds on abema/go-mp4, whose box
// registry alone added 0.9 MB to the riscv64 binary (ironkvm-dist#119); a clip
// needs one video track in one chunk, which is about two hundred lines. The
// tests still read every clip back with pmp4.

// timeScale is the MP4 track's clock, the 90 kHz video clock RTP uses.
const timeScale = 90000

// movieTimeScale is the clock of the movie header: milliseconds.
const movieTimeScale = 1000

// defaultFrame is the duration given to the last sample of a clip when there
// is no sample before it to measure one from.
const defaultFrame = time.Second / 30

// maxSampleDuration bounds one sample. A gap in the stream, as when the
// signal was gone for a while, shows as the picture before it held for the
// gap, up to this long.
const maxSampleDuration = 10 * time.Minute

var errNoParameterSets = errors.New("the clip has no parameter sets to describe the stream")

// paramSets are the parameter sets a segment is described by.
type paramSets struct {
	codec         uint8
	vps, sps, pps []byte
}

func (p paramSets) equal(o paramSets) bool {
	return p.codec == o.codec && bytes.Equal(p.vps, o.vps) && bytes.Equal(p.sps, o.sps) && bytes.Equal(p.pps, o.pps)
}

func (p paramSets) complete() bool {
	if p.sps == nil || p.pps == nil {
		return false
	}
	return p.codec != common.CodecH265 || p.vps != nil
}

// segment is a run of samples that one MP4 sample description covers.
type segment struct {
	params  paramSets
	samples []Sample
}

// split cuts samples into segments at each keyframe whose parameter sets
// differ from the ones before it. The encoder makes new ones when the
// resolution or the codec changes, which a host that reboots does, and an
// avc1 or hvc1 track has one description for the whole file. Each segment
// becomes its own file. Samples before the first keyframe with parameter sets
// are left out: nothing could decode them.
func split(samples []Sample) []segment {
	var out []segment
	for _, s := range samples {
		if s.Key {
			vps, sps, pps := egress.ParameterSets(s.Codec, s.NALUs)
			p := paramSets{codec: s.Codec, vps: vps, sps: sps, pps: pps}
			if p.complete() && (len(out) == 0 || !out[len(out)-1].params.equal(p)) {
				out = append(out, segment{params: p})
			}
		}
		if len(out) == 0 {
			continue
		}
		last := &out[len(out)-1]
		if s.Codec != last.params.codec {
			// A codec change without a keyframe that says so cannot be
			// decoded with this description.
			continue
		}
		last.samples = append(last.samples, s)
	}
	return out
}

// keepNALU reports whether a NALU goes into an MP4 sample. Parameter sets
// live in the sample description, and an access unit delimiter has no place
// in a length-prefixed sample.
func keepNALU(codec uint8, nalu []byte) bool {
	if len(nalu) == 0 {
		return false
	}
	if codec == common.CodecH265 {
		switch h265.NALUType((nalu[0] >> 1) & 0b111111) {
		case h265.NALUType_VPS_NUT, h265.NALUType_SPS_NUT, h265.NALUType_PPS_NUT, h265.NALUType_AUD_NUT:
			return false
		}
		return true
	}
	switch h264.NALUType(nalu[0] & 0x1f) {
	case h264.NALUTypeSPS, h264.NALUTypePPS, h264.NALUTypeAccessUnitDelimiter:
		return false
	}
	return true
}

// payloadSize is the size of a sample in the length-prefixed form.
func payloadSize(codec uint8, nalus [][]byte) uint32 {
	var n uint32
	for _, nalu := range nalus {
		if keepNALU(codec, nalu) {
			n += 4 + uint32(len(nalu))
		}
	}
	return n
}

// ticks is a time on the track's clock, rounded. A sample's duration is the
// difference of two of these rather than its own rounded length, so rounding
// never adds up over a clip: 33.333 ms frames are 3000 ticks each.
func ticks(d time.Duration) int64 {
	return (int64(d)*timeScale + int64(time.Second)/2) / int64(time.Second)
}

var maxSampleTicks = ticks(maxSampleDuration)

// durations are the samples' durations on the track's clock.
func durations(samples []Sample) []uint32 {
	out := make([]uint32, len(samples))
	if len(samples) == 0 {
		return out
	}
	base := samples[0].PTS
	last := ticks(defaultFrame)
	for i, s := range samples {
		d := last
		if i+1 < len(samples) {
			d = ticks(samples[i+1].PTS-base) - ticks(s.PTS-base)
			if d > 0 && d < timeScale {
				// The last sample takes the frame time the stream had.
				last = d
			}
		}
		out[i] = uint32(min(max(d, 1), maxSampleTicks))
	}
	return out
}

// boxes builds ISO BMFF boxes in memory. A box is opened with its type, its
// body appended, and closed, which writes its size.
type boxes struct{ buf []byte }

func (b *boxes) open(typ string) int {
	pos := len(b.buf)
	b.buf = append(b.buf, 0, 0, 0, 0)
	b.buf = append(b.buf, typ...)
	return pos
}

// openFull opens a full box: one with a version and flags.
func (b *boxes) openFull(typ string, version uint8, flags uint32) int {
	pos := b.open(typ)
	b.u32(uint32(version)<<24 | flags)
	return pos
}

func (b *boxes) close(pos int) {
	binary.BigEndian.PutUint32(b.buf[pos:], uint32(len(b.buf)-pos))
}

func (b *boxes) u8(v uint8)       { b.buf = append(b.buf, v) }
func (b *boxes) u16(v uint16)     { b.buf = binary.BigEndian.AppendUint16(b.buf, v) }
func (b *boxes) u32(v uint32)     { b.buf = binary.BigEndian.AppendUint32(b.buf, v) }
func (b *boxes) bytes(v []byte)   { b.buf = append(b.buf, v...) }
func (b *boxes) zeros(n int)      { b.buf = append(b.buf, make([]byte, n)...) }
func (b *boxes) matrix()          { b.unity(); b.zeros(12); b.unity(); b.zeros(12); b.u32(0x40000000) }
func (b *boxes) unity()           { b.u32(0x00010000) }
func (b *boxes) nalu(nalu []byte) { b.u16(uint16(len(nalu))); b.bytes(nalu) }

// description is what the sample entry says about the stream.
type description struct {
	entry         string
	width, height int
	config        func(b *boxes)
}

// describe reads the picture size and the decoder configuration from the
// parameter sets.
func describe(p paramSets) (description, error) {
	if p.codec == common.CodecH265 {
		return describeH265(p)
	}

	var sps h264.SPS
	if err := sps.Unmarshal(p.sps); err != nil {
		return description{}, fmt.Errorf("the H.264 SPS does not parse: %w", err)
	}
	if len(p.sps) < 4 {
		return description{}, errors.New("the H.264 SPS is too short")
	}
	return description{
		entry:  "avc1",
		width:  sps.Width(),
		height: sps.Height(),
		// AVCDecoderConfigurationRecord, ISO/IEC 14496-15 5.3.3.1.
		config: func(b *boxes) {
			pos := b.open("avcC")
			b.u8(1)
			b.bytes(p.sps[1:4]) // profile, compatibility, level
			b.u8(0xfc | 3)      // four-byte lengths
			b.u8(0xe0 | 1)      // one SPS
			b.nalu(p.sps)
			b.u8(1) // one PPS
			b.nalu(p.pps)
			switch sps.ProfileIdc {
			case 100, 110, 122, 144:
				b.u8(0xfc | uint8(sps.ChromaFormatIdc))
				b.u8(0xf8 | uint8(sps.BitDepthLumaMinus8))
				b.u8(0xf8 | uint8(sps.BitDepthChromaMinus8))
				b.u8(0) // no SPS extensions
			}
			b.close(pos)
		},
	}, nil
}

func describeH265(p paramSets) (description, error) {
	var sps h265.SPS
	if err := sps.Unmarshal(p.sps); err != nil {
		return description{}, fmt.Errorf("the H.265 SPS does not parse: %w", err)
	}
	// The general profile, tier and level are 12 bytes at a fixed place in
	// the SPS, after its NAL header and its first byte, and the record takes
	// them as they are.
	rbsp := h264.EmulationPreventionRemove(p.sps)
	if len(rbsp) < 15 {
		return description{}, errors.New("the H.265 SPS is too short")
	}
	ptl := rbsp[3:15]

	return description{
		entry:  "hvc1",
		width:  sps.Width(),
		height: sps.Height(),
		// HEVCDecoderConfigurationRecord, ISO/IEC 14496-15 8.3.3.1.
		config: func(b *boxes) {
			pos := b.open("hvcC")
			b.u8(1)
			b.bytes(ptl)
			b.u16(0xf000) // min_spatial_segmentation_idc 0
			b.u8(0xfc)    // parallelismType unknown
			b.u8(0xfc | uint8(sps.ChromaFormatIdc))
			b.u8(0xf8 | uint8(sps.BitDepthLumaMinus8))
			b.u8(0xf8 | uint8(sps.BitDepthChromaMinus8))
			b.u16(0) // avgFrameRate unknown
			nested := uint8(0)
			if sps.TemporalIDNestingFlag {
				nested = 1
			}
			b.u8((sps.MaxSubLayersMinus1+1)<<3 | nested<<2 | 3)
			b.u8(3) // VPS, SPS and PPS, one of each
			for _, set := range []struct {
				typ  h265.NALUType
				nalu []byte
			}{{h265.NALUType_VPS_NUT, p.vps}, {h265.NALUType_SPS_NUT, p.sps}, {h265.NALUType_PPS_NUT, p.pps}} {
				// Complete: the stream carries no others.
				b.u8(0x80 | uint8(set.typ))
				b.u16(1)
				b.nalu(set.nalu)
			}
			b.close(pos)
		},
	}, nil
}

// writeMP4 writes one segment as a regular MP4 file: ftyp, moov, then mdat
// with every sample in one chunk. Nothing is encoded again, and the NALUs are
// written from where the capture loop left them, so the file costs its index
// in memory and nothing more.
func writeMP4(w io.Writer, seg segment) error {
	if !seg.params.complete() || len(seg.samples) == 0 {
		return errNoParameterSets
	}
	desc, err := describe(seg.params)
	if err != nil {
		return err
	}

	codec := seg.params.codec
	durs := durations(seg.samples)
	var total uint64
	var mdatSize uint64 = 8
	allSync := true
	for i, s := range seg.samples {
		total += uint64(durs[i])
		mdatSize += uint64(payloadSize(codec, s.NALUs))
		allSync = allSync && s.Key
	}
	if mdatSize > 1<<32-1 {
		return errors.New("the clip is too large for one chunk")
	}

	var b boxes
	ftyp := b.open("ftyp")
	b.bytes([]byte("isom"))
	b.u32(0x200)
	b.bytes([]byte("isomiso2" + desc.entry + "mp41"))
	b.close(ftyp)

	moov := b.open("moov")

	mvhd := b.openFull("mvhd", 0, 0)
	b.u32(0) // creation time
	b.u32(0) // modification time
	b.u32(movieTimeScale)
	b.u32(uint32(total * movieTimeScale / timeScale))
	b.unity()      // rate 1.0
	b.u16(0x0100)  // volume 1.0
	b.zeros(2 + 8) // reserved
	b.matrix()
	b.zeros(24) // pre_defined
	b.u32(2)    // next track ID
	b.close(mvhd)

	trak := b.open("trak")
	tkhd := b.openFull("tkhd", 0, 3) // enabled, in the movie
	b.u32(0)
	b.u32(0)
	b.u32(1) // track ID
	b.u32(0)
	b.u32(uint32(total * movieTimeScale / timeScale))
	b.zeros(8)
	b.u16(0) // layer
	b.u16(0) // alternate group
	b.u16(0) // volume: video has none
	b.u16(0)
	b.matrix()
	b.u32(uint32(desc.width) << 16)
	b.u32(uint32(desc.height) << 16)
	b.close(tkhd)

	// One edit: the whole track, from its first sample, at normal speed.
	edts := b.open("edts")
	elst := b.openFull("elst", 0, 0)
	b.u32(1)
	b.u32(uint32(total * movieTimeScale / timeScale))
	b.u32(0) // media time
	b.unity()
	b.close(elst)
	b.close(edts)

	mdia := b.open("mdia")
	mdhd := b.openFull("mdhd", 0, 0)
	b.u32(0)
	b.u32(0)
	b.u32(timeScale)
	b.u32(uint32(total))
	b.u16(0x55c4) // language "und"
	b.u16(0)
	b.close(mdhd)

	hdlr := b.openFull("hdlr", 0, 0)
	b.u32(0)
	b.bytes([]byte("vide"))
	b.zeros(12)
	b.bytes([]byte("VideoHandler\x00"))
	b.close(hdlr)

	minf := b.open("minf")
	vmhd := b.openFull("vmhd", 0, 1)
	b.zeros(8) // graphics mode, colour
	b.close(vmhd)
	dinf := b.open("dinf")
	dref := b.openFull("dref", 0, 0)
	b.u32(1)
	url := b.openFull("url ", 0, 1) // the media is in this file
	b.close(url)
	b.close(dref)
	b.close(dinf)

	stbl := b.open("stbl")
	stsd := b.openFull("stsd", 0, 0)
	b.u32(1)
	entry := b.open(desc.entry)
	b.zeros(6) // reserved
	b.u16(1)   // data reference index
	b.zeros(16)
	b.u16(uint16(desc.width))
	b.u16(uint16(desc.height))
	b.u32(0x00480000) // 72 dpi
	b.u32(0x00480000)
	b.u32(0)
	b.u16(1)    // one frame per sample
	b.zeros(32) // compressor name
	b.u16(0x0018)
	b.u16(0xffff)
	desc.config(&b)
	b.close(entry)
	b.close(stsd)

	// Decoding times: the durations, run-length coded.
	stts := b.openFull("stts", 0, 0)
	countAt := len(b.buf)
	b.u32(0)
	runs := uint32(0)
	for i := 0; i < len(durs); {
		j := i
		for j < len(durs) && durs[j] == durs[i] {
			j++
		}
		b.u32(uint32(j - i))
		b.u32(durs[i])
		runs++
		i = j
	}
	binary.BigEndian.PutUint32(b.buf[countAt:], runs)
	b.close(stts)

	if !allSync {
		stss := b.openFull("stss", 0, 0)
		countAt := len(b.buf)
		b.u32(0)
		keys := uint32(0)
		for i, s := range seg.samples {
			if s.Key {
				b.u32(uint32(i + 1))
				keys++
			}
		}
		binary.BigEndian.PutUint32(b.buf[countAt:], keys)
		b.close(stss)
	}

	stsc := b.openFull("stsc", 0, 0)
	b.u32(1)
	b.u32(1) // from the first chunk
	b.u32(uint32(len(seg.samples)))
	b.u32(1) // sample description
	b.close(stsc)

	stsz := b.openFull("stsz", 0, 0)
	b.u32(0) // sizes differ
	b.u32(uint32(len(seg.samples)))
	for _, s := range seg.samples {
		b.u32(payloadSize(codec, s.NALUs))
	}
	b.close(stsz)

	stco := b.openFull("stco", 0, 0)
	b.u32(1)
	offsetAt := len(b.buf)
	b.u32(0)
	b.close(stco)

	b.close(stbl)
	b.close(minf)
	b.close(mdia)
	b.close(trak)
	b.close(moov)

	// The chunk starts after the mdat header.
	binary.BigEndian.PutUint32(b.buf[offsetAt:], uint32(len(b.buf)+8))
	b.u32(uint32(mdatSize))
	b.bytes([]byte("mdat"))
	if _, err := w.Write(b.buf); err != nil {
		return err
	}

	var prefix [4]byte
	for _, s := range seg.samples {
		for _, nalu := range s.NALUs {
			if !keepNALU(codec, nalu) {
				continue
			}
			binary.BigEndian.PutUint32(prefix[:], uint32(len(nalu)))
			if _, err := w.Write(prefix[:]); err != nil {
				return err
			}
			if _, err := w.Write(nalu); err != nil {
				return err
			}
		}
	}
	return nil
}

// segmentDuration is how long a segment plays.
func segmentDuration(seg segment) time.Duration {
	n := len(seg.samples)
	if n == 0 {
		return 0
	}
	return seg.samples[n-1].PTS - seg.samples[0].PTS
}
