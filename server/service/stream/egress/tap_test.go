package egress

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/audio"
)

// fakeVideo is a capture loop driven by the test.
type fakeVideo struct {
	frames chan stream.H264Frame
	demand func() bool
	closed atomic.Bool
}

func (v *fakeVideo) Frames() <-chan stream.H264Frame { return v.frames }
func (v *fakeVideo) Close()                          { v.closed.Store(true) }

type fakeAudio struct {
	frames chan audio.Frame
	closed atomic.Bool
}

func (a *fakeAudio) Frames() <-chan audio.Frame { return a.frames }
func (a *fakeAudio) Close()                     { a.closed.Store(true) }

// FakeSource is a tap source a test drives: Video and Audio are what the tap
// subscribed to, once it has.
type FakeSource struct {
	mu        sync.Mutex
	video     *fakeVideo
	sound     *fakeAudio
	codec     atomic.Uint32
	keyframes atomic.Int32
	readers   map[string]int
}

func newFakeSource() *FakeSource {
	f := &FakeSource{readers: make(map[string]int)}
	f.codec.Store(uint32(common.CodecH264))
	return f
}

func (f *FakeSource) source() Source {
	return Source{
		SubscribeVideo: func(demand func() bool) VideoSubscription {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.video = &fakeVideo{frames: make(chan stream.H264Frame), demand: demand}
			return f.video
		},
		SubscribeAudio: func() AudioSubscription {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.sound = &fakeAudio{frames: make(chan audio.Frame)}
			return f.sound
		},
		Codec:           func() uint8 { return uint8(f.codec.Load()) },
		RequestKeyframe: func() { f.keyframes.Add(1) },
		ReportReaders: func(kind string, count int, _ uint64) {
			f.mu.Lock()
			defer f.mu.Unlock()
			f.readers[kind] = count
		},
	}
}

func (f *FakeSource) waitVideo(t *testing.T) *fakeVideo {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		v := f.video
		f.mu.Unlock()
		if v != nil {
			return v
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the tap never subscribed to the capture loop")
	return nil
}

func (f *FakeSource) waitAudio(t *testing.T) *fakeAudio {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		f.mu.Lock()
		a := f.sound
		f.mu.Unlock()
		if a != nil {
			return a
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("the tap never subscribed to the audio hub")
	return nil
}

// h264Key is an IDR access unit with its parameter sets.
var h264Key = []byte{
	0, 0, 0, 1, 0x67, 0x42, 0xc0, 0x1f, 0xda, // SPS
	0, 0, 0, 1, 0x68, 0xce, 0x3c, 0x80, // PPS
	0, 0, 0, 1, 0x65, 0x88, 0x84, 0x00, 0x33, // IDR slice
}

var h264Delta = []byte{0, 0, 0, 1, 0x41, 0x9a, 0x02, 0x03}

func receive(t *testing.T, r *Reader) Unit {
	t.Helper()
	select {
	case u, ok := <-r.Units():
		if !ok {
			t.Fatal("the reader closed")
		}
		return u
	case <-time.After(2 * time.Second):
		t.Fatal("no unit arrived")
	}
	return Unit{}
}

func TestAReaderStartsOnAKeyframeWithItsParameterSets(t *testing.T) {
	fake := newFakeSource()
	tap := NewTap(fake.source())

	r := tap.Subscribe("rtsp", false)
	defer r.Close()
	if fake.keyframes.Load() == 0 {
		t.Fatal("a new reader did not ask for a keyframe")
	}
	video := fake.waitVideo(t)

	// A delta frame before the keyframe is no use to a new reader.
	video.frames <- stream.H264Frame{Data: h264Delta, Result: 1}
	video.frames <- stream.H264Frame{Data: h264Key, Result: 3, KeyFrame: true}
	video.frames <- stream.H264Frame{Data: h264Delta, Result: 1}

	first := receive(t, r)
	if first.Audio || !first.KeyFrame || first.Codec != common.CodecH264 {
		t.Fatalf("first unit = %+v, want an H.264 keyframe", first)
	}
	if len(first.NALUs) != 3 {
		t.Fatalf("the keyframe split into %d NAL units, want 3", len(first.NALUs))
	}
	_, sps, pps := ParameterSets(first.Codec, first.NALUs)
	if sps == nil || pps == nil {
		t.Fatal("the keyframe's SPS and PPS were not found")
	}

	second := receive(t, r)
	if second.KeyFrame || second.PTS <= first.PTS {
		t.Fatalf("second unit = %+v after PTS %s, want a later delta frame", second, first.PTS)
	}
}

// The HDMI idle logic counts the tap's readers, so a board with only an RTSP
// or SRT client keeps capturing, and the capture loop is left when the last
// reader goes.
func TestReadersAreCountedAndTheLoopEndsWithTheLast(t *testing.T) {
	fake := newFakeSource()
	tap := NewTap(fake.source())

	a := tap.Subscribe("srt", false)
	b := tap.Subscribe("srt", false)
	video := fake.waitVideo(t)
	if !video.demand() {
		t.Fatal("the tap does not want frames while it has readers")
	}

	fake.mu.Lock()
	got := fake.readers["srt"]
	fake.mu.Unlock()
	if got != 2 {
		t.Fatalf("reported %d srt readers, want 2", got)
	}

	a.Close()
	b.Close()
	if video.demand() {
		t.Fatal("the tap still wants frames with no reader")
	}

	deadline := time.Now().Add(3 * idleCheck)
	for !video.closed.Load() {
		if time.Now().After(deadline) {
			t.Fatal("the tap kept its capture subscription with no reader")
		}
		time.Sleep(5 * time.Millisecond)
	}
	fake.mu.Lock()
	got = fake.readers["srt"]
	fake.mu.Unlock()
	if got != 0 {
		t.Fatalf("reported %d srt readers after both left, want 0", got)
	}
}

// A reader that does not read must never hold up the loop or another
// reader: it loses frames to the next keyframe instead.
func TestASlowReaderDoesNotStallTheOthers(t *testing.T) {
	fake := newFakeSource()
	tap := NewTap(fake.source())

	slow := tap.Subscribe("srt", false)
	defer slow.Close()
	fast := tap.Subscribe("rtsp", false)
	defer fast.Close()
	video := fake.waitVideo(t)

	var received atomic.Int64
	go func() {
		for range fast.Units() {
			received.Add(1)
		}
	}()

	// send hands the loop a frame, which it must take at once however full
	// the slow reader is.
	send := func(frame stream.H264Frame) {
		t.Helper()
		select {
		case video.frames <- frame:
		case <-time.After(3 * time.Second):
			t.Fatal("the capture loop was held up by a reader that does not read")
		}
	}
	// caughtUp waits for the reader that keeps up to have every frame sent,
	// so it is never the one that falls behind, however loaded the machine.
	caughtUp := func(want int64) {
		t.Helper()
		deadline := time.Now().Add(3 * time.Second)
		for received.Load() < want {
			if time.Now().After(deadline) {
				t.Fatalf("the reader that keeps up got %d of %d units", received.Load(), want)
			}
			time.Sleep(time.Millisecond)
		}
	}

	send(stream.H264Frame{Data: h264Key, Result: 3, KeyFrame: true})
	sent := int64(1)
	for range 2 * readerDepth {
		send(stream.H264Frame{Data: h264Delta, Result: 1})
		sent++
		if sent%16 == 0 {
			caughtUp(sent)
		}
	}
	caughtUp(sent)

	if slow.Dropped() == 0 {
		t.Fatal("the slow reader dropped nothing")
	}
	if fake.keyframes.Load() < 3 {
		t.Fatal("the reader that fell behind did not ask for a keyframe")
	}
	// The slow reader holds out for a keyframe and gets it.
	for range readerDepth {
		<-slow.Units()
	}
	video.frames <- stream.H264Frame{Data: h264Delta, Result: 1}
	video.frames <- stream.H264Frame{Data: h264Key, Result: 3, KeyFrame: true}
	if u := receive(t, slow); !u.KeyFrame {
		t.Fatalf("after falling behind the slow reader got %+v, want the next keyframe", u)
	}
}

// Audio starts with the reader's first picture, and both tracks are stamped
// on one clock: the sound's PTS is near the picture's when both arrive
// together.
func TestAudioFollowsThePictureOnTheSameClock(t *testing.T) {
	fake := newFakeSource()
	tap := NewTap(fake.source())

	r := tap.Subscribe("rtsp", true)
	defer r.Close()
	video := fake.waitVideo(t)
	sound := fake.waitAudio(t)

	// Sound before the first picture is dropped.
	sound.frames <- audio.Frame{Seq: 0, Data: []byte{1}}
	video.frames <- stream.H264Frame{Data: h264Key, Result: 3, KeyFrame: true}
	sound.frames <- audio.Frame{Seq: 1, Data: []byte{2}}
	sound.frames <- audio.Frame{Seq: 2, Data: []byte{3}}

	key := receive(t, r)
	if !key.KeyFrame {
		t.Fatalf("first unit = %+v, want the keyframe", key)
	}
	a1 := receive(t, r)
	a2 := receive(t, r)
	if !a1.Audio || !a2.Audio || a1.Opus[0] != 2 {
		t.Fatalf("got %+v and %+v, want the packets after the keyframe", a1, a2)
	}
	if a2.PTS-a1.PTS != audioFrame {
		t.Fatalf("consecutive packets are %s apart, want %s", a2.PTS-a1.PTS, audioFrame)
	}
	if d := a1.PTS - key.PTS; d < -time.Second || d > time.Second {
		t.Fatalf("audio is %s from the picture that arrived with it", d)
	}
}

func TestTheClockFollowsTheCaptureTimes(t *testing.T) {
	epoch := time.Unix(1000, 0)
	c := clock{epoch: epoch}

	pts, wall := c.video(epoch.Add(40*time.Millisecond), epoch.Add(90*time.Millisecond))
	if pts != 40*time.Millisecond || !wall.Equal(epoch.Add(40*time.Millisecond)) {
		t.Fatalf("a picture captured 40 ms in got PTS %s", pts)
	}
	// A capture time that goes back still moves the PTS on.
	if pts, _ := c.video(epoch.Add(10*time.Millisecond), epoch.Add(100*time.Millisecond)); pts <= 40*time.Millisecond {
		t.Fatalf("a capture time from the past gave PTS %s", pts)
	}

	// Audio counts 20 ms a packet from its first one, whatever the jitter.
	first, _ := c.audio(5, epoch.Add(120*time.Millisecond))
	if first != 100*time.Millisecond {
		t.Fatalf("a packet that arrived 120 ms in got PTS %s, want 100 ms", first)
	}
	next, _ := c.audio(6, epoch.Add(150*time.Millisecond))
	if next != first+audioFrame {
		t.Fatalf("the next packet got PTS %s, want %s", next, first+audioFrame)
	}
	// Sound stamped before the epoch still counts 20 ms a packet.
	early := clock{epoch: epoch}
	p0, _ := early.audio(0, epoch.Add(5*time.Millisecond))
	p1, _ := early.audio(1, epoch.Add(26*time.Millisecond))
	if p1-p0 != audioFrame {
		t.Fatalf("packets either side of the epoch are %s apart, want %s", p1-p0, audioFrame)
	}

	// A stall longer than audioResync sets the clock again.
	later, _ := c.audio(7, epoch.Add(2*time.Second))
	if later != 2*time.Second-audioFrame {
		t.Fatalf("a packet after a stall got PTS %s", later)
	}
}
