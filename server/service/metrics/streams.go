package metrics

import (
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/direct"
	"NanoKVM-Server/service/stream/mjpeg"
	"NanoKVM-Server/service/vm"
)

// streamPaths are the sources that report viewers to the HDMI demand count,
// under the names they report with.
var streamPaths = []string{"mjpeg", "direct", "webrtc"}

// byteSource is one delivery path's sent-bytes counter. WebRTC has none: pion
// sends its bytes, and reading them would need a stats collector per peer.
type byteSource struct {
	path string
	read func() uint64
}

// The stream readers, variables so tests can stub them.
var (
	viewerCount = vm.HdmiViewerCount

	// GetFrameRateCounter starts the counter's three-second ticker on its first
	// call, which a scrape before any stream would be. That is the same ticker
	// the first stream starts, and it only writes now_fps when the rate changes.
	streamFPS = func() int32 { return stream.GetFrameRateCounter().GetFPS() }

	sentBytes = []byteSource{
		{path: "mjpeg", read: mjpeg.SentBytes},
		{path: "direct", read: direct.SentBytes},
	}

	suppressedFrames = mjpeg.Suppressed
)

const (
	helpViewers    = "Clients watching the HDMI stream, by delivery path."
	helpFPS        = "Frames per second the capture loops delivered, averaged over three seconds."
	helpSentBytes  = "Video bytes written to viewers, by delivery path. WebRTC is not counted."
	helpSuppressed = "MJPEG frames kept off the wire because they matched the previous frame."
)

func collectStreams(w *Writer) {
	for _, path := range streamPaths {
		w.Gauge("ironkvm_stream_viewers", helpViewers, float64(viewerCount(path)), L("path", path))
	}

	w.Gauge("ironkvm_stream_fps", helpFPS, float64(streamFPS()))

	for _, source := range sentBytes {
		w.Counter("ironkvm_stream_sent_bytes_total", helpSentBytes, float64(source.read()), L("path", source.path))
	}

	w.Counter("ironkvm_stream_suppressed_frames_total", helpSuppressed, float64(suppressedFrames()))
}
