package mjpeg

import (
	"sync"
	"testing"
	"time"

	"NanoKVM-Server/common"

	"github.com/gin-gonic/gin"
)

func withCaptureFPS(t *testing.T) func() []int {
	t.Helper()

	var mutex sync.Mutex
	var told []int

	original := setCaptureFPS
	t.Cleanup(func() { setCaptureFPS = original })
	setCaptureFPS = func(fps int) {
		mutex.Lock()
		defer mutex.Unlock()
		told = append(told, fps)
	}

	return func() []int {
		mutex.Lock()
		defer mutex.Unlock()

		return append([]int(nil), told...)
	}
}

// withScreenFPS puts one FPS setting in place and puts the previous one back.
// The screen singleton is process wide.
func withScreenFPS(t *testing.T, fps int) {
	t.Helper()

	before := common.GetScreen().Snapshot().FPS
	t.Cleanup(func() { common.SetScreen("fps", before) })
	common.SetScreen("fps", fps)
}

func waitFor(t *testing.T, what string, condition func() bool) {
	t.Helper()

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}

	t.Fatalf("timed out waiting for %s", what)
}

// addBareClient puts a client in the map without starting its writer, so the
// loop keeps running without anything touching a response.
func addBareClient(s *Streamer) *gin.Context {
	c := &gin.Context{}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.clients[c] = newClient(c)
	s.updateClientSnapshotLocked()

	return c
}

func removeBareClient(s *Streamer, c *gin.Context) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	delete(s.clients, c)
	s.updateClientSnapshotLocked()
}

// MJPEG has to say this as well as H.264. The capture channel is shared, it
// hands out every frame the source produces unless told a rate, and a frame
// this loop never reads is still written to memory in full.
func TestTheCaptureChannelIsToldBeforeTheLoopStarts(t *testing.T) {
	withScreenFPS(t, 30)
	told := withCaptureFPS(t)

	s := NewStreamer()

	// No clients, so the loop reports the rate and then returns on its first
	// tick rather than running on.
	s.run()

	values := told()
	if len(values) == 0 {
		t.Fatal("the capture channel was told nothing")
	}

	if values[0] != 30 {
		t.Fatalf("the capture channel was told %d, want the configured 30", values[0])
	}
}

func TestAChangedFrameRateReachesTheCaptureChannel(t *testing.T) {
	withScreenFPS(t, 30)
	told := withCaptureFPS(t)

	s := NewStreamer()
	c := addBareClient(s)

	done := make(chan struct{})
	go func() {
		defer close(done)
		s.run()
	}()

	waitFor(t, "the first frame rate", func() bool { return len(told()) > 0 })

	common.SetScreen("fps", 60)

	waitFor(t, "the new frame rate", func() bool {
		values := told()

		return len(values) > 1 && values[len(values)-1] == 60
	})

	removeBareClient(s, c)
	<-done
}
