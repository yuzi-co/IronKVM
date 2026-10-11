package stream

import (
	"math"
	"slices"
	"sync"
	"time"

	"NanoKVM-Server/proto"
)

// The board's share of each delivery path's latency, for the status bar: from
// when a frame's first line reached the capture (common.VideoTimes
// CaptureStart) to the frame being handed to the network for a viewer. The
// browser adds what it measures of the rest (the way here, its buffering and
// decoding, the display) and shows the sum. GET /api/stream/info reports the
// median of each path over the last boardLatencyWindow.
//
// Only frames whose library gives capture times count (libkvm-v4l2). MJPEG
// reads have no capture times, so there is no MJPEG path.

// BoardPath names a delivery path.
type BoardPath int

const (
	// BoardPathWebRTC ends when the last RTP packet of the frame has been
	// written to the viewer's UDP socket.
	BoardPathWebRTC BoardPath = iota
	// BoardPathDirect ends when the frame's WebSocket message has been
	// written to the viewer's TCP socket.
	BoardPathDirect

	boardPaths
)

// boardLatencyWindow is how far back the median looks. One second is sixty
// frames at 60 fps and enough to ride out a keyframe.
const boardLatencyWindow = time.Second

// boardLatencyRing is how many samples a path keeps, at least a window's
// worth for a few viewers at 60 fps.
const boardLatencyRing = 512

// maxBoardLatency drops a sample that cannot be a latency: a capture time
// from before a wall clock step, for one.
const maxBoardLatency = 10 * time.Second

type latencyRing struct {
	mutex sync.Mutex
	at    [boardLatencyRing]time.Time
	took  [boardLatencyRing]time.Duration
	next  int
	count int
}

var boardLatency [boardPaths]latencyRing

// NoteBoardLatency records that a frame captured at captured, on the wall
// clock, was handed to the network at sent. A zero captured, a frame without
// capture times, is ignored.
func NoteBoardLatency(path BoardPath, captured, sent time.Time) {
	if captured.IsZero() || path < 0 || path >= boardPaths {
		return
	}

	took := sent.Sub(captured)
	if took < 0 || took > maxBoardLatency {
		return
	}

	r := &boardLatency[path]
	r.mutex.Lock()
	r.at[r.next] = sent
	r.took[r.next] = took
	r.next = (r.next + 1) % boardLatencyRing
	r.count = min(r.count+1, boardLatencyRing)
	r.mutex.Unlock()
}

// median is the median of the samples taken in the window before now, and
// false when there are none.
func (r *latencyRing) median(now time.Time) (time.Duration, bool) {
	var recent [boardLatencyRing]time.Duration
	n := 0

	r.mutex.Lock()
	for i := 0; i < r.count; i++ {
		age := now.Sub(r.at[i])
		if age >= 0 && age <= boardLatencyWindow {
			recent[n] = r.took[i]
			n++
		}
	}
	r.mutex.Unlock()

	if n == 0 {
		return 0, false
	}

	values := recent[:n]
	slices.Sort(values)
	if n%2 == 1 {
		return values[n/2], true
	}
	return (values[n/2-1] + values[n/2]) / 2, true
}

// readBoardLatency is what GET /api/stream/info reports, nil when no path
// handed a frame with capture times to a viewer in the last window.
func readBoardLatency(now time.Time) *proto.BoardLatency {
	var out proto.BoardLatency
	some := false

	if d, ok := boardLatency[BoardPathWebRTC].median(now); ok {
		out.WebRTC = roundedMs(d)
		some = true
	}
	if d, ok := boardLatency[BoardPathDirect].median(now); ok {
		out.Direct = roundedMs(d)
		some = true
	}

	if !some {
		return nil
	}
	return &out
}

// roundedMs is d in milliseconds to a tenth.
func roundedMs(d time.Duration) *float64 {
	ms := math.Round(float64(d)/float64(time.Millisecond)*10) / 10
	return &ms
}
