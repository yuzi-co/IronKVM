package mjpeg

import (
	"NanoKVM-Server/common"
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/vm"
	"bytes"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// setCaptureFPS is a variable so the loop can be driven without the capture
// hardware, the same way setFrameDetect is in frame-detect.go.
var setCaptureFPS = func(fps int) {
	common.GetKvmVision().SetCaptureFPS(uint8(fps))
}

// readMjpeg is a variable for the same reason: the stub build returns -1 for
// every read, so a test that wants to see a frame has to supply one.
var readMjpeg = func(width uint16, height uint16, quality uint16) ([]byte, int) {
	return common.GetKvmVision().ReadMjpeg(width, height, quality)
}

// refreshInterval bounds how long a viewer may hold a frame that duplicate
// suppression decided not to resend.
//
// Suppression is safe on its own: a viewer that misses a frame gets the next
// one, because the slot always holds the newest. This exists for the case
// suppression cannot see, which is a frame that left this process and never
// arrived. Without it the screen would stay wrong until the host changed
// something, and on an idle console that can be a long time.
//
// One frame every five seconds is 70KB/s against the 10.5MB/s a stream costs
// when nothing is suppressed, so the insurance is close to free.
//
// A variable so a test does not have to wait five seconds for it.
var refreshInterval = 5 * time.Second

type Streamer struct {
	mutex          sync.Mutex
	clients        map[*gin.Context]*client
	clientSnapshot atomic.Pointer[[]*client]
	running        int32
	frameMutex     sync.RWMutex
	latestFrame    LatestFrame
	cacheRefs      int32
	viewerVersion  uint64

	// forceNext makes the next frame go out whatever duplicate suppression
	// thinks of it. AddClient sets it, because a viewer that arrives while the
	// screen is still would otherwise wait for the host to change something
	// before it saw anything at all.
	forceNext atomic.Bool

	// suppressed counts the frames duplicate suppression kept off the wire.
	suppressed atomic.Uint64

	// lastFrame and lastSent belong to the capture loop and are touched by
	// nothing else, so they need no lock.
	lastFrame []byte
	lastSent  time.Time
}

func NewStreamer() *Streamer {
	s := &Streamer{
		clients: make(map[*gin.Context]*client),
	}
	s.updateClientSnapshotLocked()

	return s
}

func (s *Streamer) AddClient(c *gin.Context) *client {
	client := newClient(c)
	go client.write()

	s.mutex.Lock()
	s.clients[c] = client
	count := s.updateClientSnapshotLocked()
	s.viewerVersion++
	version := s.viewerVersion
	s.mutex.Unlock()
	vm.UpdateHdmiViewerSnapshot("mjpeg", count, version)

	// Before the loop starts, so the first frame of a new stream is sent even
	// if it matches the one the previous stream ended on.
	s.forceNext.Store(true)

	if atomic.CompareAndSwapInt32(&s.running, 0, 1) {
		go s.run()
		log.Debug("mjpeg stream started")
	}

	return client
}

func (s *Streamer) RemoveClient(c *gin.Context) {
	s.mutex.Lock()
	client, exists := s.clients[c]
	delete(s.clients, c)
	count := s.updateClientSnapshotLocked()
	s.viewerVersion++
	version := s.viewerVersion
	s.mutex.Unlock()
	vm.UpdateHdmiViewerSnapshot("mjpeg", count, version)

	if exists {
		client.stop()
	}

	log.Debugf("mjpeg connection removed, remaining clients: %d", count)
}

func (s *Streamer) updateClientSnapshotLocked() int {
	clients := make([]*client, 0, len(s.clients))
	for _, c := range s.clients {
		clients = append(clients, c)
	}
	s.clientSnapshot.Store(&clients)

	return len(clients)
}

func (s *Streamer) getClients() []*client {
	clients := s.clientSnapshot.Load()
	if clients == nil {
		return nil
	}

	return *clients
}

func (s *Streamer) run() {
	defer atomic.StoreInt32(&s.running, 0)

	screen := common.GetScreen()
	common.CheckScreen()
	values := screen.Snapshot()
	fps := values.FPS

	// The capture channel hands out every frame the source produces unless it
	// is told otherwise, and a frame this loop never reads is still written to
	// memory in full. The H.264 loop says the same thing for the same reason.
	setCaptureFPS(fps)

	// The comparison holds a whole JPEG, so let go of it when the stream ends
	// rather than keeping it for as long as the process runs.
	defer func() { s.lastFrame = nil }()

	ticker := time.NewTicker(time.Second / time.Duration(fps))
	defer ticker.Stop()

	for range ticker.C {
		clients := s.getClients()
		if len(clients) == 0 {
			log.Debug("mjpeg stream stopped due to no clients")
			return
		}

		values = screen.Snapshot()

		// Ahead of the read, and ahead of the early return under it. This used
		// to sit at the end of the loop, past a continue that a failed read and
		// an unchanged frame both take, so while the screen was still a rate
		// change reached neither the ticker nor the capture channel. The H.264
		// loop has always read the setting here.
		if values.FPS != fps && values.FPS != 0 {
			fps = values.FPS
			setCaptureFPS(fps)
			ticker.Reset(time.Second / time.Duration(fps))
		}

		// Read nothing while every viewer still holds the frame it was given
		// last. The slot keeps one frame per client and Replace throws away
		// whatever it finds, so a read taken now is encoded, copied twice and
		// then discarded.
		//
		// This is not a rare case. Measured 2026-09-06 at 1080p over HTTPS,
		// the loop read 30 frames a second while about 10 reached the browser,
		// so two reads in three were waste: a cgo call, a 350KB copy inside
		// libkvm and another in C.GoBytes, about 14MB/s of copying and 7MB/s
		// of heap for frames nobody could take.
		//
		// It also makes the read rate follow the viewer instead of the ticker.
		// A viewer on a slow link now slows capture rather than filling the
		// carveout with frames that are dropped a moment later.
		if allPending(clients) {
			continue
		}

		data, result := readMjpeg(values.Width, values.Height, values.Quality)
		stream.UpdateCaptureStatus(stream.CaptureModeMJPEG, result)
		if result < 0 || result == 5 || len(data) == 0 {
			continue
		}

		// Ahead of the suppression check on purpose. A screenshot asks what is
		// on the screen now, and a frame that is identical to the last one is
		// still the answer; only its timestamp changes.
		if s.frameCacheEnabled() {
			s.setLatestFrame(data, values.Width, values.Height)
		}

		if !s.shouldSend(data) {
			continue
		}

		// Handing the frame over never blocks: a client that is behind gets
		// the newest frame and the older one is dropped.
		for _, client := range clients {
			client.enqueue(data)
		}

		stream.GetFrameRateCounter().Update()
	}
}

// allPending reports whether every viewer still holds the frame it was given
// last. The caller has already established that there is at least one.
func allPending(clients []*client) bool {
	for _, c := range clients {
		if !c.slot.Pending() {
			return false
		}
	}

	return true
}

// shouldSend decides whether this frame is worth putting on the wire.
//
// The JPEG encoder runs at fixed quantisation (VENC channel 0 reports
// RcMode: FIXQP, Qfactor 80), so an unchanged screen produces byte-identical
// output and a comparison of the encoded frame is an exact test for "nothing
// happened". It costs a length check, which settles almost every real change,
// and a memcmp of memory this process already owns. No raw frame is mapped and
// no cache is invalidated, which is what the detector inside libkvm has to do.
//
// This is the case the libkvm detector cannot serve. That one samples one frame
// in sixty and is all-or-nothing, so a screen whose only motion is a clock
// reads as "changed" and all thirty frames a second go out, though twenty-nine
// of every thirty are identical to the frame before them.
//
// If the encoder turns out not to be deterministic, nothing ever matches, the
// comparison costs a memcmp and the stream behaves exactly as it did before.
// The failure mode is a lost optimisation, not a broken picture.
func (s *Streamer) shouldSend(data []byte) bool {
	forced := s.forceNext.Swap(false)
	duplicate := bytes.Equal(s.lastFrame, data)

	// Kept whatever is decided below, so the next frame is compared against
	// what the screen last showed rather than what was last sent.
	s.lastFrame = data

	if forced || !duplicate || time.Since(s.lastSent) >= refreshInterval {
		s.lastSent = time.Now()

		return true
	}

	s.suppressed.Add(1)

	return false
}

// Suppressed is the number of frames duplicate suppression kept off the wire.
func (s *Streamer) Suppressed() uint64 {
	return s.suppressed.Load()
}

// setLatestFrame caches the frame for the screenshot API. The capture loop
// hands over a freshly allocated slice per frame and nobody mutates it, so the
// cache shares it rather than copying a whole JPEG on every tick. getLatestFrame
// still copies, so callers cannot reach back into it.
func (s *Streamer) setLatestFrame(data []byte, width uint16, height uint16) {
	s.frameMutex.Lock()
	defer s.frameMutex.Unlock()

	s.latestFrame = LatestFrame{
		Data:       data,
		Width:      width,
		Height:     height,
		CapturedAt: time.Now(),
	}
}

func (s *Streamer) clearLatestFrame() {
	s.frameMutex.Lock()
	defer s.frameMutex.Unlock()

	s.latestFrame = LatestFrame{}
}

func (s *Streamer) enableLatestFrameCache() {
	atomic.AddInt32(&s.cacheRefs, 1)
}

func (s *Streamer) disableLatestFrameCache() {
	for {
		current := atomic.LoadInt32(&s.cacheRefs)
		if current <= 0 {
			return
		}

		if atomic.CompareAndSwapInt32(&s.cacheRefs, current, current-1) {
			if current == 1 {
				s.clearLatestFrame()
			}
			return
		}
	}
}

func (s *Streamer) frameCacheEnabled() bool {
	return atomic.LoadInt32(&s.cacheRefs) > 0
}

func (s *Streamer) getLatestFrame() (LatestFrame, bool) {
	if !s.frameCacheEnabled() {
		return LatestFrame{}, false
	}

	s.frameMutex.RLock()
	defer s.frameMutex.RUnlock()

	if len(s.latestFrame.Data) == 0 {
		return LatestFrame{}, false
	}

	return LatestFrame{
		Data:       append([]byte(nil), s.latestFrame.Data...),
		Width:      s.latestFrame.Width,
		Height:     s.latestFrame.Height,
		CapturedAt: s.latestFrame.CapturedAt,
	}, true
}
