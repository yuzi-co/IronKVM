package vm

import (
	"testing"

	hdmistate "NanoKVM-Server/service/vm/hdmi_state"
)

// useHdmiDemand gives one test a fresh demand record and restores the real one.
func useHdmiDemand(t *testing.T) {
	t.Helper()

	hdmiMutex.Lock()
	original := hdmiDemand
	hdmiDemand = hdmistate.New()
	hdmiMutex.Unlock()

	t.Cleanup(func() {
		hdmiMutex.Lock()
		hdmiDemand = original
		hdmiMutex.Unlock()
	})
}

func TestHdmiViewerCountReadsEachSource(t *testing.T) {
	useHdmiDemand(t)

	hdmiMutex.Lock()
	hdmiDemand.UpdateViewer("mjpeg", 2, 1)
	hdmiDemand.UpdateViewer("webrtc", 1, 1)
	hdmiMutex.Unlock()

	for source, want := range map[string]int{"mjpeg": 2, "webrtc": 1, "direct": 0} {
		if got := HdmiViewerCount(source); got != want {
			t.Errorf("HdmiViewerCount(%q) = %d, want %d", source, got, want)
		}
	}
}
