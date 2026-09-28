package metrics

import (
	"strings"
	"testing"

	"NanoKVM-Server/service/stream"
)

func TestStreamsWritesEveryFamily(t *testing.T) {
	setVar(t, &viewerCounts, func(sources []string) []int {
		counts := make([]int, len(sources))
		for i, source := range sources {
			counts[i] = map[string]int{"mjpeg": 2, "webrtc": 1}[source]
		}
		return counts
	})
	setVar(t, &streamFPS, func() int32 { return 24 })
	setVar(t, &sentBytes, []byteSource{
		{path: "mjpeg", read: func() uint64 { return 1000 }},
		{path: "direct", read: func() uint64 { return 500 }},
	})
	setVar(t, &suppressedFrames, func() uint64 { return 7 })

	want := `# HELP ironkvm_stream_viewers Clients watching the HDMI stream, by delivery path.
# TYPE ironkvm_stream_viewers gauge
ironkvm_stream_viewers{path="mjpeg"} 2
ironkvm_stream_viewers{path="direct"} 0
ironkvm_stream_viewers{path="webrtc"} 1
# HELP ironkvm_stream_fps Frames per second the capture loops delivered, averaged over three seconds.
# TYPE ironkvm_stream_fps gauge
ironkvm_stream_fps 24
# HELP ironkvm_stream_sent_bytes_total Video bytes written to viewers, by delivery path. WebRTC is not counted.
# TYPE ironkvm_stream_sent_bytes_total counter
ironkvm_stream_sent_bytes_total{path="mjpeg"} 1000
ironkvm_stream_sent_bytes_total{path="direct"} 500
# HELP ironkvm_stream_suppressed_frames_total MJPEG frames kept off the wire because they matched the previous frame.
# TYPE ironkvm_stream_suppressed_frames_total counter
ironkvm_stream_suppressed_frames_total 7
`
	assertText(t, render(t, collectStreams), want)
}

// A scrape before any stream must not start the frame counter: its ticker
// writes now_fps to the card every three seconds from then on.
func TestAScrapeLeavesTheFrameCounterUnstarted(t *testing.T) {
	if _, started := stream.CurrentFPS(); started {
		t.Skip("another test in this binary started the counter")
	}

	got := render(t, collectStreams)

	if _, started := stream.CurrentFPS(); started {
		t.Fatal("the scrape started the frame counter")
	}
	if !strings.Contains(got, "ironkvm_stream_fps 0\n") {
		t.Fatalf("the rate of a counter never started is not 0:\n%s", got)
	}
}
