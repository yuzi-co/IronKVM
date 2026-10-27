package dashcam

import (
	"bytes"
	"encoding/binary"
	"testing"
	"time"

	"github.com/bluenviron/mediacommon/v2/pkg/formats/mp4/codecs"
	"github.com/bluenviron/mediacommon/v2/pkg/formats/pmp4"

	"NanoKVM-Server/common"
	"NanoKVM-Server/service/stream/egress/egresstest"
)

var (
	idr   = []byte{0x65, 0x88, 0x84, 0x00, 0x33, 0xff}
	slice = []byte{0x41, 0x9a, 0x02, 0x03, 0x04}
	aud   = []byte{0x09, 0xf0}
)

// H.265 parameter sets, from mediacommon's own tests.
var (
	h265VPS = []byte{
		0x40, 0x01, 0x0c, 0x01, 0xff, 0xff, 0x02, 0x20,
		0x00, 0x00, 0x03, 0x00, 0xb0, 0x00, 0x00, 0x03,
		0x00, 0x00, 0x03, 0x00, 0x7b, 0x18, 0xb0, 0x24,
	}
	h265SPS = []byte{
		0x42, 0x01, 0x01, 0x02, 0x20, 0x00, 0x00, 0x03,
		0x00, 0xb0, 0x00, 0x00, 0x03, 0x00, 0x00, 0x03,
		0x00, 0x7b, 0xa0, 0x07, 0x82, 0x00, 0x88, 0x7d,
		0xb6, 0x71, 0x8b, 0x92, 0x44, 0x80, 0x53, 0x88,
		0x88, 0x92, 0xcf, 0x24, 0xa6, 0x92, 0x72, 0xc9,
		0x12, 0x49, 0x22, 0xdc, 0x91, 0xaa, 0x48, 0xfc,
		0xa2, 0x23, 0xff, 0x00, 0x01, 0x00, 0x01, 0x6a,
		0x02, 0x02, 0x02, 0x01,
	}
	h265PPS   = []byte{0x44, 0x01, 0xc0, 0x25, 0x2f, 0x05, 0x32, 0x40}
	h265IDR   = []byte{0x26, 0x01, 0xaf, 0x06, 0xb8}
	h265Trail = []byte{0x02, 0x01, 0xd0, 0x2f}
)

// h264Clip is n access units at fps with a keyframe every gop, as the encoder
// makes them: an AUD, and SPS and PPS before each IDR.
func h264Clip(n, fps, gop int) []Sample {
	frame := time.Second / time.Duration(fps)
	out := make([]Sample, n)
	for i := range out {
		s := Sample{Codec: common.CodecH264, PTS: time.Second + time.Duration(i)*frame}
		if i%gop == 0 {
			s.Key = true
			s.NALUs = [][]byte{aud, egresstest.SPS, egresstest.PPS, idr}
		} else {
			s.NALUs = [][]byte{aud, slice}
		}
		out[i] = s
	}
	return out
}

// topBoxes lists the top-level boxes of an MP4 file.
func topBoxes(t *testing.T, data []byte) []string {
	t.Helper()
	var names []string
	for off := 0; off < len(data); {
		if len(data)-off < 8 {
			t.Fatalf("a box header is cut at offset %d", off)
		}
		size := int(binary.BigEndian.Uint32(data[off:]))
		if size < 8 || off+size > len(data) {
			t.Fatalf("box %q at %d has size %d in a file of %d", data[off+4:off+8], off, size, len(data))
		}
		names = append(names, string(data[off+4:off+8]))
		off += size
	}
	return names
}

func TestWriteMP4H264(t *testing.T) {
	samples := h264Clip(90, 30, 30)
	segs := split(samples)
	if len(segs) != 1 || len(segs[0].samples) != 90 {
		t.Fatalf("split gave %d segments", len(segs))
	}

	var buf bytes.Buffer
	if err := writeMP4(&buf, segs[0]); err != nil {
		t.Fatal(err)
	}

	if got := topBoxes(t, buf.Bytes()); len(got) != 3 || got[0] != "ftyp" || got[1] != "moov" || got[2] != "mdat" {
		t.Fatalf("the top-level boxes are %v, want ftyp, moov, mdat", got)
	}

	var p pmp4.Presentation
	if err := p.Unmarshal(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("the file does not parse back: %s", err)
	}
	if len(p.Tracks) != 1 {
		t.Fatalf("%d tracks", len(p.Tracks))
	}
	track := p.Tracks[0]
	codec, ok := track.Codec.(*codecs.H264)
	if !ok || !bytes.Equal(codec.SPS, egresstest.SPS) || !bytes.Equal(codec.PPS, egresstest.PPS) {
		t.Fatalf("the track's codec is %#v", track.Codec)
	}
	if track.TimeScale != timeScale || len(track.Samples) != 90 {
		t.Fatalf("time scale %d and %d samples, want %d and 90", track.TimeScale, len(track.Samples), timeScale)
	}

	var total uint32
	for i, s := range track.Samples {
		if s.Duration != timeScale/30 {
			t.Fatalf("sample %d lasts %d ticks, want %d", i, s.Duration, timeScale/30)
		}
		total += s.Duration
		if s.IsNonSyncSample != (i%30 != 0) {
			t.Fatalf("sample %d: non-sync %t", i, s.IsNonSyncSample)
		}
		data, err := s.GetPayload()
		if err != nil {
			t.Fatal(err)
		}
		// Length-prefixed, without the AUD and the parameter sets.
		want := slice
		if i%30 == 0 {
			want = idr
		}
		if len(data) != 4+len(want) || int(binary.BigEndian.Uint32(data)) != len(want) || !bytes.Equal(data[4:], want) {
			t.Fatalf("sample %d holds % x", i, data)
		}
	}
	if total != 3*timeScale {
		t.Fatalf("the track lasts %d ticks, want %d", total, 3*timeScale)
	}
}

func TestWriteMP4H265(t *testing.T) {
	var samples []Sample
	for i := range 20 {
		s := Sample{Codec: common.CodecH265, PTS: time.Duration(i) * 40 * time.Millisecond}
		if i%10 == 0 {
			s.Key = true
			s.NALUs = [][]byte{h265VPS, h265SPS, h265PPS, h265IDR}
		} else {
			s.NALUs = [][]byte{h265Trail}
		}
		samples = append(samples, s)
	}
	segs := split(samples)
	if len(segs) != 1 {
		t.Fatalf("split gave %d segments", len(segs))
	}

	var buf bytes.Buffer
	if err := writeMP4(&buf, segs[0]); err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("hvc1")) || !bytes.Contains(buf.Bytes(), []byte("hvcC")) {
		t.Fatal("an H.265 clip has no hvc1 sample entry")
	}

	var p pmp4.Presentation
	if err := p.Unmarshal(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatalf("the file does not parse back: %s", err)
	}
	codec, ok := p.Tracks[0].Codec.(*codecs.H265)
	if !ok || !bytes.Equal(codec.VPS, h265VPS) || !bytes.Equal(codec.SPS, h265SPS) || !bytes.Equal(codec.PPS, h265PPS) {
		t.Fatalf("the track's codec is %#v", p.Tracks[0].Codec)
	}
	if n := len(p.Tracks[0].Samples); n != 20 {
		t.Fatalf("%d samples, want 20", n)
	}
	for i, s := range p.Tracks[0].Samples {
		if s.Duration != timeScale*40/1000 || s.IsNonSyncSample != (i%10 != 0) {
			t.Fatalf("sample %d: %d ticks, non-sync %t", i, s.Duration, s.IsNonSyncSample)
		}
	}
}

// A change of parameter sets, as a host's change of resolution makes, starts
// a new file; samples before the first keyframe with parameter sets are left
// out.
func TestSplitAtAChangeOfParameterSets(t *testing.T) {
	otherSPS := append([]byte(nil), egresstest.SPS...)
	otherSPS[len(otherSPS)-1] ^= 0x01

	samples := []Sample{
		{Codec: common.CodecH264, Key: true, NALUs: [][]byte{idr}, PTS: 0},
		{Codec: common.CodecH264, NALUs: [][]byte{slice}, PTS: 1},
		{Codec: common.CodecH264, Key: true, NALUs: [][]byte{egresstest.SPS, egresstest.PPS, idr}, PTS: 2},
		{Codec: common.CodecH264, NALUs: [][]byte{slice}, PTS: 3},
		// A keyframe without parameter sets continues the segment.
		{Codec: common.CodecH264, Key: true, NALUs: [][]byte{idr}, PTS: 4},
		// The same parameter sets again continue it as well.
		{Codec: common.CodecH264, Key: true, NALUs: [][]byte{egresstest.SPS, egresstest.PPS, idr}, PTS: 5},
		{Codec: common.CodecH264, Key: true, NALUs: [][]byte{otherSPS, egresstest.PPS, idr}, PTS: 6},
		{Codec: common.CodecH264, NALUs: [][]byte{slice}, PTS: 7},
		{Codec: common.CodecH265, Key: true, NALUs: [][]byte{h265VPS, h265SPS, h265PPS, h265IDR}, PTS: 8},
	}
	segs := split(samples)
	if len(segs) != 3 {
		t.Fatalf("split gave %d segments, want 3", len(segs))
	}
	for i, want := range []int{4, 2, 1} {
		if len(segs[i].samples) != want {
			t.Fatalf("segment %d has %d samples, want %d", i, len(segs[i].samples), want)
		}
		if !segs[i].samples[0].Key {
			t.Fatalf("segment %d starts on a delta frame", i)
		}
	}
	if segs[2].params.codec != common.CodecH265 {
		t.Fatal("the last segment is not H.265")
	}
}

func TestWriteMP4RefusesASegmentWithoutParameterSets(t *testing.T) {
	var buf bytes.Buffer
	if err := writeMP4(&buf, segment{samples: h264Clip(3, 30, 30)}); err != errNoParameterSets {
		t.Fatalf("got %v, want errNoParameterSets", err)
	}
}

// A gap in the stream holds the picture before it, and a sample never lasts
// longer than maxSampleDuration.
func TestWriteMP4Gap(t *testing.T) {
	samples := h264Clip(3, 30, 30)
	samples[2].PTS = samples[1].PTS + time.Hour
	segs := split(samples)

	var buf bytes.Buffer
	if err := writeMP4(&buf, segs[0]); err != nil {
		t.Fatal(err)
	}
	var p pmp4.Presentation
	if err := p.Unmarshal(bytes.NewReader(buf.Bytes())); err != nil {
		t.Fatal(err)
	}
	got := p.Tracks[0].Samples
	if got[1].Duration != uint32(maxSampleDuration/time.Second)*timeScale {
		t.Fatalf("the sample before the gap lasts %d ticks", got[1].Duration)
	}
	if got[2].Duration != timeScale/30 {
		t.Fatalf("the last sample lasts %d ticks, want the frame time before it", got[2].Duration)
	}
}
