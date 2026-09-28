package mjpeg

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestEnqueueQueuesFrameForTheWriter(t *testing.T) {
	c := newClient(nil)

	c.enqueue([]byte("frame"))

	frame, ok := c.slot.Take()
	if !ok || string(frame) != "frame" {
		t.Fatalf("expected the frame to be queued, got %q (ok=%v)", frame, ok)
	}
}

func TestEnqueueKeepsTheNewestFrameForASlowClient(t *testing.T) {
	// Every MJPEG frame stands alone, so a client that fell behind should get
	// the current picture rather than work through a backlog.
	c := newClient(nil)

	c.enqueue([]byte("old"))
	c.enqueue([]byte("new"))

	frame, _ := c.slot.Take()
	if string(frame) != "new" {
		t.Fatalf("expected the newest frame, got %q", frame)
	}

	if c.slot.Dropped() != 1 {
		t.Fatalf("expected 1 dropped frame, got %d", c.slot.Dropped())
	}
}

func TestFailedIsClosedWhenTheWriterGivesUp(t *testing.T) {
	c := newClient(nil)

	c.fail()

	select {
	case <-c.failed:
	default:
		t.Fatal("the handler must be woken when the writer gives up")
	}
}

func TestFailIsIdempotent(t *testing.T) {
	c := newClient(nil)

	c.fail()
	c.fail()
}

// Every byte a viewer received is counted: the part header, the JPEG and the
// trailing CRLF.
func TestWriteFrameCountsTheBytesItSent(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/api/stream/mjpeg", nil)
	c := newClient(ctx)

	before := SentBytes()
	if err := c.writeFrame([]byte("jpeg")); err != nil {
		t.Fatalf("writeFrame: %s", err)
	}

	got := SentBytes() - before
	if got != uint64(recorder.Body.Len()) {
		t.Fatalf("counted %d bytes, the response holds %d", got, recorder.Body.Len())
	}
	if got != 62 {
		t.Fatalf("counted %d bytes, want 62 (56 of header, 4 of JPEG, 2 of CRLF)", got)
	}
}

func TestSuppressedReadsThePackageStreamer(t *testing.T) {
	t.Cleanup(func() { streamer.lastFrame = nil })

	before := Suppressed()
	frame := []byte("the same frame")
	streamer.shouldSend(frame)
	streamer.shouldSend(frame)

	if got := Suppressed() - before; got != 1 {
		t.Fatalf("suppressed rose by %d, want 1", got)
	}
}
