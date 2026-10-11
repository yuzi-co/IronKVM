package stream

import (
	"os"
	"slices"
	"strconv"
	"sync"
	"time"

	"NanoKVM-Server/common"

	log "github.com/sirupsen/logrus"
)

// The board's share of the video latency, as PiKVM's ustreamer-dump shows its
// own (GRAB, ENC, LAT), for measuring (ironkvm-dist trial 78). With
// KVM_LATENCY_LOG set, the server logs every latencyLogEvery, per stage, the
// median, the 95th percentile and the largest of:
//
//   - grab: the source's frame time, the first line to the last (the capture
//     timestamp is the frame-done interrupt)
//   - wait: capture done to the picture going to the encoder (the read's wait,
//     the scaler)
//   - enc: the encoder's work
//   - copy: the access unit out of the encoder's buffer and the library's scan
//   - hand: the library returning to the frame reaching the delivery paths
//   - lat: the frame's first line to that hand-over, the sum of the above
//   - sent: the first line to the last packet of the frame leaving a WebRTC
//     viewer's socket (the writer's queue, SRTP and the send), any viewer
//
// Only frames whose library gives capture times count (libkvm-v4l2). Off, the
// cost is one check per frame.
const latencyLogEnv = "KVM_LATENCY_LOG"

const latencyLogEvery = 10 * time.Second

var latencyLogOn = os.Getenv(latencyLogEnv) != ""

type latencySamples struct {
	grab, wait, enc, copy, hand, lat, sent []time.Duration
}

type latencyLog struct {
	mutex   sync.Mutex
	since   time.Time
	samples latencySamples
}

var sharedLatencyLog = &latencyLog{}

// newLatencyLog answers the shared log when KVM_LATENCY_LOG is set, else nil.
// A nil log takes nothing.
func newLatencyLog() *latencyLog {
	if !latencyLogOn {
		return nil
	}

	return sharedLatencyLog
}

// add records one captured frame, handed to the delivery paths at handed.
func (l *latencyLog) add(frame H264Frame, handed time.Time) {
	if l == nil || !frame.HasTimes {
		return
	}

	t := frame.Times

	l.mutex.Lock()
	defer l.mutex.Unlock()

	s := &l.samples
	s.grab = append(s.grab, t.FrameTime)
	s.wait = append(s.wait, t.EncodeStart.Sub(t.CaptureEnd))
	s.enc = append(s.enc, t.EncodeEnd.Sub(t.EncodeStart))
	s.copy = append(s.copy, t.Returned.Sub(t.EncodeEnd))
	s.hand = append(s.hand, handed.Sub(t.Returned))
	s.lat = append(s.lat, handed.Sub(t.CaptureStart()))

	l.flushLocked(handed)
}

// NoteSent records that the last packet of a frame left a WebRTC viewer's
// socket at sent. The WebRTC writer calls it. The status bar's board latency
// takes it whether or not the log is on.
func NoteSent(times common.VideoTimes, sent time.Time) {
	if times.CaptureEnd.IsZero() {
		return
	}

	NoteBoardLatency(BoardPathWebRTC, times.CaptureStart(), sent)

	if !latencyLogOn {
		return
	}

	l := sharedLatencyLog

	l.mutex.Lock()
	l.samples.sent = append(l.samples.sent, sent.Sub(times.CaptureStart()))
	l.mutex.Unlock()
}

func (l *latencyLog) flushLocked(now time.Time) {
	if l.since.IsZero() {
		l.since = now

		return
	}

	if now.Sub(l.since) < latencyLogEvery {
		return
	}

	s := l.samples
	log.Infof("latency (ms, p50/p95/max) over %d frames: grab %s, wait %s, enc %s, copy %s, hand %s, lat %s; sent %s (%d)",
		len(s.lat), spread(s.grab), spread(s.wait), spread(s.enc), spread(s.copy), spread(s.hand),
		spread(s.lat), spread(s.sent), len(s.sent))

	l.samples = latencySamples{}
	l.since = now
}

// spread is "p50/p95/max" in milliseconds with one decimal, or "-".
func spread(values []time.Duration) string {
	if len(values) == 0 {
		return "-"
	}

	sorted := slices.Clone(values)
	slices.Sort(sorted)

	at := func(q float64) time.Duration {
		return sorted[min(len(sorted)-1, int(q*float64(len(sorted))))]
	}

	return ms(at(0.5)) + "/" + ms(at(0.95)) + "/" + ms(sorted[len(sorted)-1])
}

func ms(d time.Duration) string {
	return strconv.FormatFloat(float64(d)/float64(time.Millisecond), 'f', 1, 64)
}
