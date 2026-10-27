// Package egress is the tap the RTSP server (ironkvm-dist#142) and the SRT
// listener (ironkvm-dist#143) take the board's picture and sound from.
//
// Neither does any media work of its own. The video is the H.264 or H.265 the
// hardware encoder already makes for WebRTC, taken from the shared capture loop
// as one more subscriber (stream.SubscribeH264), and the sound is the Opus the
// audio hub already encodes for WebRTC (audio.Shared). The tap adds three
// things both outputs need and that would otherwise be written twice:
//
//   - One clock. Video and audio arrive on different goroutines from different
//     devices, and a player can only line them up if both carry times on the
//     same clock. Every unit carries a PTS measured from the tap's start.
//   - Readers that never hold anything up. A reader is fed through a bounded
//     slot; one that falls behind loses frames up to the next keyframe, as a
//     WebRTC viewer does, and the capture loop and the other readers never
//     wait for it.
//   - The demand bookkeeping. The tap subscribes while it has a reader and
//     leaves when the last one goes, and it reports its readers to the HDMI
//     idle logic, so a board with only an RTSP or SRT client keeps capturing.
package egress

import (
	"sync"
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/vm"

	"github.com/bluenviron/mediacommon/v2/pkg/codecs/h264"
	log "github.com/sirupsen/logrus"
)

// Unit is one access unit of video or one Opus packet, on the tap's clock.
//
// NALUs and Opus point into the frame the capture loop handed every
// subscriber. Nobody may write to them.
type Unit struct {
	Audio bool

	// Codec is common.CodecH264 or common.CodecH265, for video.
	Codec    uint8
	KeyFrame bool
	// NALUs is the access unit without start codes.
	NALUs [][]byte

	Opus []byte

	// PTS is when the picture was captured, or when the sound was, measured
	// from the start of the tap. Both tracks use it, which is what keeps them
	// in step.
	PTS time.Duration
	// Wall is the same moment on the wall clock, for RTCP sender reports.
	Wall time.Time
}

// readerDepth is how many units a reader may fall behind by. Audio is 50
// units a second and video up to 60, so 64 is about half a second at 1080p60
// with sound: enough for a TCP send to stall through a retransmission, and
// little enough that a reader that is really behind is cut back to the next
// keyframe instead of being fed stale picture.
const readerDepth = 64

// audioFrame is the length of one Opus packet from the hub.
const audioFrame = time.Duration(audio.SamplesPerFrame) * time.Second / time.Duration(audio.SampleRate)

// audioResync is how far the audio clock may drift from the arrival times
// before it is set again. The host's USB clock and the board's are not the
// same clock, and a capture that stalled has a gap the sequence does not show.
const audioResync = 200 * time.Millisecond

// idleCheck is how often the loop looks for a reader that wants audio while
// none is captured, and for the last reader having gone.
const idleCheck = time.Second

// VideoSubscription is the capture loop seen from the tap.
// *stream.H264Subscription is the one the server runs with.
type VideoSubscription interface {
	Frames() <-chan stream.H264Frame
	Close()
}

// AudioSubscription is the audio hub seen from the tap. *audio.Subscription is
// the one the server runs with.
type AudioSubscription interface {
	Frames() <-chan audio.Frame
	Close()
}

// Source is what a tap is built on. Tests replace every part of it.
type Source struct {
	// SubscribeVideo joins the capture loop. demand says whether the tap
	// wants a frame at all.
	SubscribeVideo func(demand func() bool) VideoSubscription
	// SubscribeAudio joins the audio hub, or returns nil when this board has
	// nothing to capture right now.
	SubscribeAudio func() AudioSubscription
	// Codec is the codec the encoder runs.
	Codec func() uint8
	// RequestKeyframe asks the encoder for a keyframe, for a reader that
	// joined or fell behind.
	RequestKeyframe func()
	// ReportReaders tells the HDMI idle logic how many readers of a kind
	// there are.
	ReportReaders func(kind string, count int, version uint64)
}

// DefaultSource is the capture loop, the audio hub and the HDMI viewer
// counts.
func DefaultSource() Source {
	return Source{
		SubscribeVideo: func(demand func() bool) VideoSubscription {
			return stream.SubscribeH264(demand)
		},
		SubscribeAudio: func() AudioSubscription {
			// A nil *audio.Subscription in the interface would not be nil.
			if sub := audio.Shared.Subscribe(); sub != nil {
				return sub
			}
			return nil
		},
		Codec: func() uint8 {
			return common.GetScreen().Snapshot().Codec
		},
		RequestKeyframe: func() {
			stream.RequestKeyframe(stream.KeyframeReasonWaiting)
		},
		ReportReaders: vm.UpdateHdmiViewerSnapshot,
	}
}

// Tap is the shared feed. Shared is the one the server runs with.
type Tap struct {
	source Source

	mutex   sync.Mutex
	readers map[*Reader]struct{}
	kinds   map[string]int
	version uint64
	running bool
	// wake tells the loop a reader joined, so one that wants audio does not
	// wait for the next idle check to get it.
	wake chan struct{}
}

// NewTap builds a tap on source.
func NewTap(source Source) *Tap {
	return &Tap{
		source:  source,
		readers: make(map[*Reader]struct{}),
		kinds:   make(map[string]int),
		wake:    make(chan struct{}, 1),
	}
}

var sharedTap = sync.OnceValue(func() *Tap { return NewTap(DefaultSource()) })

// Shared is the tap on the board's capture loop and audio hub.
func Shared() *Tap {
	return sharedTap()
}

// Reader is one output's view of the tap: an RTSP stream, which every RTSP
// player shares, or one SRT caller.
type Reader struct {
	tap   *Tap
	kind  string
	audio bool
	slot  *stream.FrameSlot[Unit]
	once  sync.Once

	// started and waiting are read and written by the tap's loop alone.
	// started turns true on the reader's first keyframe, and waiting while it
	// is behind and holds out for the next one.
	started bool
	waiting bool
}

// Subscribe adds a reader of a kind ("rtsp", "srt"), which names it to the
// HDMI idle logic. withAudio asks for the host's sound beside the picture.
//
// The first unit a reader gets is a keyframe, which the encoder is asked for
// now, and audio starts with it, so an output can begin its stream on a
// picture a player can decode.
func (t *Tap) Subscribe(kind string, withAudio bool) *Reader {
	r := &Reader{
		tap:   t,
		kind:  kind,
		audio: withAudio,
		slot:  stream.NewFrameSlotDepth[Unit](readerDepth),
	}

	t.mutex.Lock()
	t.readers[r] = struct{}{}
	t.kinds[kind]++
	t.version++
	count, version := t.kinds[kind], t.version
	start := !t.running
	t.running = true
	t.mutex.Unlock()

	t.source.ReportReaders(kind, count, version)
	if start {
		go t.run()
	} else {
		select {
		case t.wake <- struct{}{}:
		default:
		}
	}
	t.source.RequestKeyframe()

	return r
}

// Units delivers the reader's units in order. It closes with the reader.
func (r *Reader) Units() <-chan Unit {
	return r.slot.Channel()
}

// Dropped counts the units this reader was not ready to take.
func (r *Reader) Dropped() uint64 {
	return r.slot.Dropped()
}

// Close leaves the tap. It is safe to call more than once.
func (r *Reader) Close() {
	r.once.Do(func() {
		t := r.tap
		t.mutex.Lock()
		delete(t.readers, r)
		t.kinds[r.kind]--
		t.version++
		count, version := t.kinds[r.kind], t.version
		t.mutex.Unlock()

		t.source.ReportReaders(r.kind, count, version)
		r.slot.Close()
	})
}

// Readers is how many readers the tap has.
func (t *Tap) Readers() int {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return len(t.readers)
}

func (t *Tap) snapshot() (readers []*Reader, wantAudio bool) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	readers = make([]*Reader, 0, len(t.readers))
	for r := range t.readers {
		readers = append(readers, r)
		wantAudio = wantAudio || r.audio
	}
	return readers, wantAudio
}

// stopIfIdle marks the loop stopped when the last reader has gone, under the
// mutex Subscribe takes, so a reader arriving now either finds the loop
// running or starts the next one.
func (t *Tap) stopIfIdle() bool {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if len(t.readers) > 0 {
		return false
	}
	t.running = false
	return true
}

// clock turns capture and arrival times into PTS on one timeline.
type clock struct {
	epoch time.Time

	lastVideo time.Duration
	hasVideo  bool

	// The audio clock counts packets from an anchor, so it runs at exactly
	// 20 ms a packet whatever the jitter of their arrival.
	audioAnchor time.Duration
	audioSeq    uint64
	hasAudio    bool
	lastAudio   time.Duration
}

// video is the PTS of a picture captured at captured, or of one that arrived
// at now when the library cannot say when it was captured.
func (c *clock) video(captured time.Time, now time.Time) (time.Duration, time.Time) {
	wall := now
	if !captured.IsZero() {
		wall = captured
	}
	pts := wall.Sub(c.epoch)
	// A muxer refuses time going back, which a step of the wall clock or a
	// capture time from before the tap started would make.
	if c.hasVideo && pts <= c.lastVideo {
		pts = c.lastVideo + time.Millisecond
	}
	c.lastVideo, c.hasVideo = pts, true
	return pts, c.epoch.Add(pts)
}

// audio is the PTS of the packet seq that arrived at now. The packet holds
// the 20 ms of sound before it was handed over, so it began a frame earlier.
func (c *clock) audio(seq uint64, now time.Time) (time.Duration, time.Time) {
	arrived := now.Sub(c.epoch) - audioFrame

	started := c.hasAudio
	pts := c.audioAnchor + time.Duration(seq-c.audioSeq)*audioFrame
	if !started || seq < c.audioSeq || pts-arrived > audioResync || arrived-pts > audioResync {
		c.audioAnchor, c.audioSeq, c.hasAudio = arrived, seq, true
		pts = arrived
	}
	if started && pts <= c.lastAudio {
		pts = c.lastAudio + time.Millisecond
	}
	c.lastAudio = pts
	return pts, c.epoch.Add(pts)
}

// timeNow is a variable so a test can run the clock.
var timeNow = time.Now

func (t *Tap) run() {
	video := t.source.SubscribeVideo(func() bool { return t.Readers() > 0 })
	defer video.Close()

	var sound AudioSubscription
	var soundFrames <-chan audio.Frame
	defer func() {
		if sound != nil {
			sound.Close()
		}
	}()

	idle := time.NewTicker(idleCheck)
	defer idle.Stop()

	// The epoch is a second before the loop starts, so a packet stamped a
	// frame before it arrived, as sound is, still has a positive PTS.
	c := clock{epoch: timeNow().Add(-time.Second)}

	// syncAudio holds a subscription to the hub while a reader wants sound,
	// and lets go of it when none does. The hub runs arecord for as long as
	// anyone holds one.
	syncAudio := func(want bool) {
		switch {
		case want && sound == nil:
			if sound = t.source.SubscribeAudio(); sound != nil {
				soundFrames = sound.Frames()
			}
		case !want && sound != nil:
			sound.Close()
			sound, soundFrames = nil, nil
		}
	}
	_, wantAudio := t.snapshot()
	syncAudio(wantAudio)

	for {
		select {
		case <-idle.C:
			readers, wantAudio := t.snapshot()
			if len(readers) == 0 && t.stopIfIdle() {
				return
			}
			syncAudio(wantAudio)

		case <-t.wake:
			_, wantAudio := t.snapshot()
			syncAudio(wantAudio)

		case frame, ok := <-video.Frames():
			if !ok {
				// The capture loop does not end a subscription on its own;
				// this is a test's source, or the server going down.
				t.closeAll()
				return
			}
			if frame.Result < 0 || len(frame.Data) == 0 {
				continue
			}
			t.deliverVideo(&c, frame)

		case packet, ok := <-soundFrames:
			if !ok {
				// The capture ended by itself. The next idle check starts
				// another if a reader still wants sound.
				sound, soundFrames = nil, nil
				continue
			}
			if packet.State != audio.StateUnknown || len(packet.Data) == 0 {
				continue
			}
			t.deliverAudio(&c, packet)
		}
	}
}

func (t *Tap) deliverVideo(c *clock, frame stream.H264Frame) {
	var au h264.AnnexB
	// The H.265 start code is the same as the H.264 one, so one splitter
	// serves both.
	if err := au.Unmarshal(frame.Data); err != nil || len(au) == 0 {
		log.Debugf("egress: dropped an access unit that does not split: %v", err)
		return
	}

	var captured time.Time
	if frame.HasTimes {
		captured = frame.Times.CaptureStart()
	}
	pts, wall := c.video(captured, timeNow())

	unit := Unit{
		Codec:    t.source.Codec(),
		KeyFrame: frame.KeyFrame,
		NALUs:    au,
		PTS:      pts,
		Wall:     wall,
	}

	readers, _ := t.snapshot()
	for _, r := range readers {
		r.offerVideo(unit)
	}
}

func (t *Tap) deliverAudio(c *clock, packet audio.Frame) {
	pts, wall := c.audio(packet.Seq, timeNow())
	unit := Unit{Audio: true, Opus: packet.Data, PTS: pts, Wall: wall}

	readers, _ := t.snapshot()
	for _, r := range readers {
		r.offerAudio(unit)
	}
}

// offerVideo hands a picture to the reader, and holds the reader's video back
// to the next keyframe once it has missed one, the way the capture loop holds
// back a delivery path (stream.H264Subscription.deliver): a frame after a gap
// is predicted from a picture the player never got.
func (r *Reader) offerVideo(unit Unit) {
	if unit.KeyFrame {
		r.started, r.waiting = true, false
		// A keyframe needs nothing before it, so stale units make room
		// for it rather than it being lost.
		r.slot.PutIndependent(unit)
		return
	}
	if !r.started || r.waiting {
		return
	}
	if !r.slot.TryPut(unit) {
		r.waiting = true
		r.tap.source.RequestKeyframe()
	}
}

// offerAudio hands a packet to the reader once its picture has started. A
// lost Opus packet is a 20 ms gap a decoder conceals, so a full slot only
// loses this one.
func (r *Reader) offerAudio(unit Unit) {
	if !r.audio || !r.started {
		return
	}
	r.slot.TryPut(unit)
}

// closeAll ends every reader, for a capture loop that went away.
func (t *Tap) closeAll() {
	t.mutex.Lock()
	readers := make([]*Reader, 0, len(t.readers))
	for r := range t.readers {
		readers = append(readers, r)
	}
	t.running = false
	t.mutex.Unlock()

	for _, r := range readers {
		r.Close()
	}
}
