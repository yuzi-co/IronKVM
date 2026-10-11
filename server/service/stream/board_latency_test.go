package stream

import (
	"encoding/json"
	"testing"
	"time"

	"NanoKVM-Server/common"
)

func resetBoardLatency() {
	for i := range boardLatency {
		boardLatency[i].mutex.Lock()
		boardLatency[i].next = 0
		boardLatency[i].count = 0
		boardLatency[i].mutex.Unlock()
	}
}

func TestBoardLatencyIsTheMedianOfTheLastSecond(t *testing.T) {
	resetBoardLatency()
	t.Cleanup(resetBoardLatency)

	now := time.Unix(1760000000, 0)
	note := func(sentAgo, took time.Duration) {
		sent := now.Add(-sentAgo)
		NoteBoardLatency(BoardPathWebRTC, sent.Add(-took), sent)
	}

	// Too old to count, however large.
	note(1500*time.Millisecond, 900*time.Millisecond)
	note(10*time.Millisecond, 30*time.Millisecond)
	note(20*time.Millisecond, 20*time.Millisecond)
	note(30*time.Millisecond, 50*time.Millisecond)
	note(40*time.Millisecond, 22*time.Millisecond)
	note(50*time.Millisecond, 21*time.Millisecond)

	got, ok := boardLatency[BoardPathWebRTC].median(now)
	if !ok || got != 22*time.Millisecond {
		t.Fatalf("median %v %v, want 22ms", got, ok)
	}

	// An even count takes the middle two.
	note(60*time.Millisecond, 24*time.Millisecond)
	got, _ = boardLatency[BoardPathWebRTC].median(now)
	if got != 23*time.Millisecond {
		t.Fatalf("median of six %v, want 23ms", got)
	}
}

func TestBoardLatencyIgnoresFramesWithoutCaptureTimes(t *testing.T) {
	resetBoardLatency()
	t.Cleanup(resetBoardLatency)

	now := time.Now()
	NoteBoardLatency(BoardPathDirect, time.Time{}, now)
	// A capture time after the send: the wall clock stepped back.
	NoteBoardLatency(BoardPathDirect, now.Add(time.Second), now)
	// A capture time from before a step forward.
	NoteBoardLatency(BoardPathDirect, now.Add(-time.Hour), now)

	if got := readBoardLatency(now); got != nil {
		t.Fatalf("board latency %+v, want none", got)
	}
}

// The field is left out while nothing was sent, so the answer is the same as
// an older server's; with samples, each path that has some is reported.
func TestStreamInfoCarriesBoardLatency(t *testing.T) {
	resetBoardLatency()
	t.Cleanup(resetBoardLatency)

	info := readStreamInfo()
	raw, _ := json.Marshal(info)
	var fields map[string]any
	_ = json.Unmarshal(raw, &fields)
	if _, ok := fields["boardLatency"]; ok {
		t.Fatalf("boardLatency present with no samples: %s", raw)
	}

	now := time.Now()
	NoteBoardLatency(BoardPathWebRTC, now.Add(-21340*time.Microsecond), now)

	raw, _ = json.Marshal(readStreamInfo())
	var got struct {
		BoardLatency map[string]float64 `json:"boardLatency"`
	}
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.BoardLatency["webrtc"] != 21.3 {
		t.Fatalf("webrtc %v, want 21.3: %s", got.BoardLatency["webrtc"], raw)
	}
	if _, ok := got.BoardLatency["direct"]; ok {
		t.Fatalf("direct reported without samples: %s", raw)
	}
}

func TestNoteSentFeedsTheWebRTCPath(t *testing.T) {
	resetBoardLatency()
	t.Cleanup(resetBoardLatency)

	now := time.Now()
	times := common.VideoTimes{CaptureEnd: now.Add(-30 * time.Millisecond), FrameTime: 16 * time.Millisecond}
	NoteSent(times, now)

	got, ok := boardLatency[BoardPathWebRTC].median(now)
	if !ok || got != 46*time.Millisecond {
		t.Fatalf("webrtc %v %v, want 46ms (frame time plus 30ms)", got, ok)
	}
}
