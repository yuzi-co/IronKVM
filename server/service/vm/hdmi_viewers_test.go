package vm

import (
	"reflect"
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

func TestHdmiViewerCountsReadsEachSourceInOrder(t *testing.T) {
	useHdmiDemand(t)

	hdmiMutex.Lock()
	hdmiDemand.UpdateViewer("mjpeg", 2, 1)
	hdmiDemand.UpdateViewer("webrtc", 1, 1)
	hdmiMutex.Unlock()

	got := HdmiViewerCounts([]string{"mjpeg", "direct", "webrtc"})
	if want := []int{2, 0, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("HdmiViewerCounts = %v, want %v", got, want)
	}
}
