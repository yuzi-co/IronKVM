//go:build linux

package common

import (
	"bytes"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// fakePort is the shared block in Go memory, and fakeWorker the library's
// thread behind it: it reads the request eventfd and answers through the done
// eventfd as kvmv_worker.c does, with answer deciding what each request gets.
type fakePort struct {
	mu     sync.Mutex
	req    nativeRequest
	reqSeq atomic.Uint32
	resSeq atomic.Uint32
	answer nativeAnswer
}

func (p *fakePort) put(req nativeRequest, seq uint32) {
	p.mu.Lock()
	p.req = req
	p.mu.Unlock()
	p.reqSeq.Store(seq)
}

func (p *fakePort) answered() uint32 { return p.resSeq.Load() }

func (p *fakePort) take() nativeAnswer {
	p.mu.Lock()
	defer p.mu.Unlock()

	a := p.answer
	a.data = bytes.Clone(a.data)

	return a
}

type fakeWorker struct {
	port      *fakePort
	requestFD int
	doneFD    int
	stop      atomic.Bool
	reads     atomic.Int32
	exited    chan struct{}

	// answer is called for each new request, on the worker's goroutine.
	answer func(seq uint32, req nativeRequest) nativeAnswer
}

func eventfd(t *testing.T, flags int) int {
	t.Helper()

	fd, err := unix.Eventfd(0, unix.EFD_CLOEXEC|flags)
	if err != nil {
		t.Fatalf("eventfd: %v", err)
	}

	return fd
}

func signalFD(fd int) error {
	one := uint64(1)
	_, err := unix.Write(fd, (*[8]byte)(unsafe.Pointer(&one))[:])

	return err
}

func startFakeWorker(t *testing.T, answer func(seq uint32, req nativeRequest) nativeAnswer) *fakeWorker {
	t.Helper()

	w := &fakeWorker{
		port:      &fakePort{},
		requestFD: eventfd(t, 0),
		doneFD:    eventfd(t, unix.EFD_NONBLOCK),
		exited:    make(chan struct{}),
		answer:    answer,
	}

	go func() {
		defer close(w.exited)

		var last uint32
		var buf [8]byte
		for {
			if _, err := unix.Read(w.requestFD, buf[:]); err != nil && !errors.Is(err, unix.EINTR) {
				return
			}
			if w.stop.Load() {
				return
			}
			seq := w.port.reqSeq.Load()
			if seq == last {
				continue
			}
			last = seq

			w.port.mu.Lock()
			req := w.port.req
			w.port.mu.Unlock()

			a := w.answer(seq, req)
			a.seq = seq
			w.reads.Add(1)

			w.port.mu.Lock()
			w.port.answer = a
			w.port.mu.Unlock()
			w.port.resSeq.Store(seq)
			_ = signalFD(w.doneFD)
		}
	}()

	t.Cleanup(func() {
		w.stop.Store(true)
		_ = signalFD(w.requestFD)
		<-w.exited
		_ = unix.Close(w.requestFD)
		_ = unix.Close(w.doneFD)
	})

	return w
}

func newTestCapture(t *testing.T, w *fakeWorker) *nativeCapture {
	t.Helper()

	n, err := newNativeCapture(w.port, w.requestFD, w.doneFD)
	if err != nil {
		t.Fatalf("newNativeCapture: %v", err)
	}
	t.Cleanup(n.close)
	n.timeout = 2 * time.Second

	return n
}

func frameAnswer(seq uint32, _ nativeRequest) nativeAnswer {
	return nativeAnswer{result: 4, data: []byte{byte(seq), 0xaa}}
}

func TestNativeCaptureWanted(t *testing.T) {
	for value, want := range map[string]bool{
		"": false, "off": false, "0": false, "no": false, "tick": false,
		"on": true, "ON": true, " on\n": true, "1": true, "true": true, "yes": true,
	} {
		if got := nativeCaptureWanted(value); got != want {
			t.Errorf("nativeCaptureWanted(%q) = %v, want %v", value, got, want)
		}
	}
}

// A read hands the worker the arguments and answers its frame and result.
func TestNativeReadAnswersTheWorkersFrame(t *testing.T) {
	var asked nativeRequest
	w := startFakeWorker(t, func(seq uint32, req nativeRequest) nativeAnswer {
		asked = req

		return nativeAnswer{result: 3, data: []byte{1, 2, 3}}
	})
	n := newTestCapture(t, w)

	req := nativeRequest{width: 1920, height: 1080, bitrate: 5000, codec: 2, gop: 0, fps: 0}
	data, result, err := n.read(req)
	if err != nil || result != 3 || !bytes.Equal(data, []byte{1, 2, 3}) {
		t.Fatalf("read = %v, %d, %v; want [1 2 3], 3, nil", data, result, err)
	}
	if asked != req {
		t.Fatalf("the worker was asked %+v, want %+v", asked, req)
	}

	// Every read is a request of its own.
	for i := 2; i <= 4; i++ {
		w.answer = frameAnswer
		data, result, err = n.read(req)
		if err != nil || result != 4 || data[0] != byte(i) {
			t.Fatalf("read %d = %v, %d, %v", i, data, result, err)
		}
	}
	if got := w.reads.Load(); got != 4 {
		t.Fatalf("the worker read %d times, want 4", got)
	}
}

// A failed read passes the library's result through with no data.
func TestNativeReadPassesAFailureThrough(t *testing.T) {
	w := startFakeWorker(t, func(uint32, nativeRequest) nativeAnswer {
		return nativeAnswer{result: noImage}
	})
	n := newTestCapture(t, w)

	data, result, err := n.read(nativeRequest{})
	if err != nil || result != noImage || data != nil {
		t.Fatalf("read = %v, %d, %v; want nil, -1, nil", data, result, err)
	}
	if _, ok := n.lastTimes(); ok {
		t.Fatal("a failed read reported capture times")
	}
}

// A worker that does not answer in time costs the read -5, not the loop. The
// next read waits for the late answer first, throws it away, and asks again,
// because the worker serves one request at a time.
func TestNativeReadTimesOutAndDropsTheLateAnswer(t *testing.T) {
	release := make(chan struct{})
	w := startFakeWorker(t, func(seq uint32, req nativeRequest) nativeAnswer {
		if seq == 1 {
			<-release
		}

		return frameAnswer(seq, req)
	})
	n := newTestCapture(t, w)
	n.timeout = 50 * time.Millisecond

	if data, result, err := n.read(nativeRequest{}); err != nil || result != retrievingImage || data != nil {
		t.Fatalf("stuck read = %v, %d, %v; want nil, -5, nil", data, result, err)
	}
	if !n.pending {
		t.Fatal("the unanswered request was not remembered")
	}

	// Still stuck: no second request goes out behind the first.
	if _, result, _ := n.read(nativeRequest{}); result != retrievingImage {
		t.Fatalf("read while stuck = %d, want -5", result)
	}
	if got := w.port.reqSeq.Load(); got != 1 {
		t.Fatalf("request %d went out while the worker was still on the first", got)
	}

	close(release)
	n.timeout = 2 * time.Second
	data, result, err := n.read(nativeRequest{})
	if err != nil || result != 4 || data[0] != 2 {
		t.Fatalf("read after the late answer = %v, %d, %v; want request 2's frame", data, result, err)
	}
	if n.pending {
		t.Fatal("still pending after the late answer")
	}
}

// The done eventfd may carry a signal for an answer the read is not waiting
// for; wait goes by the sequence number, not by the signal.
func TestNativeReadIgnoresAStaleSignal(t *testing.T) {
	w := startFakeWorker(t, frameAnswer)
	n := newTestCapture(t, w)

	if err := signalFD(w.doneFD); err != nil {
		t.Fatal(err)
	}
	data, result, err := n.read(nativeRequest{})
	if err != nil || result != 4 || data[0] != 1 {
		t.Fatalf("read = %v, %d, %v; want request 1's frame", data, result, err)
	}
}

// A request the worker cannot be told about is an error the caller falls back
// on, not a frame that never comes.
func TestNativeReadFailsWhenTheWorkerCannotBeSignalled(t *testing.T) {
	w := startFakeWorker(t, frameAnswer)
	n := newTestCapture(t, w)
	n.requestFD = -1

	if _, result, err := n.read(nativeRequest{}); !errors.Is(err, errNativeKick) || result != noImage {
		t.Fatalf("read = %d, %v; want -1 and errNativeKick", result, err)
	}
}

// The times come with the answer and go on the wall clock as
// kvmv_video_times_wall puts them.
func TestNativeAnswerTimes(t *testing.T) {
	a := nativeAnswer{
		times:  [5]uint64{1_000_000, 16_667, 1_002_000, 1_015_000, 1_016_000},
		ntimes: 5,
		mono:   2_000_000_000,                 // 2 s after boot
		real:   1_790_000_000_000_000_000 + 7, // and on the wall clock
	}
	got, ok := a.videoTimes()
	if !ok {
		t.Fatal("five times were not taken")
	}
	wall := func(us int64) time.Time {
		return time.Unix(0, a.real-(a.mono-us*1000))
	}
	if !got.CaptureEnd.Equal(wall(1_000_000)) || !got.EncodeStart.Equal(wall(1_002_000)) ||
		!got.EncodeEnd.Equal(wall(1_015_000)) || !got.Returned.Equal(wall(1_016_000)) {
		t.Fatalf("times = %+v", got)
	}
	if got.FrameTime != 16667*time.Microsecond {
		t.Fatalf("frame time = %s", got.FrameTime)
	}
	if got.EncodeEnd.Sub(got.CaptureEnd) != 15*time.Millisecond {
		t.Fatalf("capture to encode end = %s, want 15ms", got.EncodeEnd.Sub(got.CaptureEnd))
	}

	a.ntimes = 4
	if _, ok := a.videoTimes(); ok {
		t.Fatal("four times were taken for five")
	}
}

func TestNativeReadKeepsTheTimesOfTheLastFrame(t *testing.T) {
	w := startFakeWorker(t, func(seq uint32, req nativeRequest) nativeAnswer {
		a := frameAnswer(seq, req)
		a.times = [5]uint64{uint64(seq) * 1000, 0, 0, 0, 0}
		a.ntimes = 5
		a.mono, a.real = 1e9, 2e9

		return a
	})
	n := newTestCapture(t, w)

	for range 2 {
		if _, _, err := n.read(nativeRequest{}); err != nil {
			t.Fatal(err)
		}
	}
	got, ok := n.lastTimes()
	if !ok || !got.CaptureEnd.Equal(time.Unix(0, 2e9-(1e9-2000*1000))) {
		t.Fatalf("lastTimes = %+v, %v; want request 2's", got, ok)
	}
}

// The netpoller needs a nonblocking descriptor. One that is not refuses the
// worker rather than block the thread on each frame.
func TestNewNativeCaptureRefusesABlockingDescriptor(t *testing.T) {
	blocking := eventfd(t, 0)
	defer unix.Close(blocking)

	if n, err := newNativeCapture(&fakePort{}, blocking, blocking); err == nil {
		n.close()
		t.Fatal("a blocking done descriptor was taken")
	}
	if _, err := newNativeCapture(&fakePort{}, -1, -1); err == nil {
		t.Fatal("missing descriptors were taken")
	}
}

func TestChooseNativeCapture(t *testing.T) {
	w := startFakeWorker(t, frameAnswer)
	opened := 0
	lib := func(available bool, openErr error) nativeCaptureLibrary {
		return nativeCaptureLibrary{
			available: func() bool { return available },
			open: func() (nativePort, int, int, error) {
				opened++
				if openErr != nil {
					return nil, -1, -1, openErr
				}

				return w.port, w.requestFD, w.doneFD, nil
			},
		}
	}

	// Off by default: the library is not even asked.
	for _, env := range []string{"", "off", "0"} {
		if n := chooseNativeCapture(env, lib(true, nil)); n != nil {
			t.Fatalf("%q chose the worker", env)
		}
	}
	if opened != 0 {
		t.Fatalf("the worker was started %d times with the variable off", opened)
	}

	// Sipeed's library, or an older libkvm-v4l2: per-frame calls.
	if n := chooseNativeCapture("on", lib(false, nil)); n != nil {
		t.Fatal("chose a worker the library does not have")
	}
	if opened != 0 {
		t.Fatal("tried to start a worker the library does not have")
	}

	// A worker that does not start: per-frame calls.
	if n := chooseNativeCapture("on", lib(true, errors.New("NULL"))); n != nil {
		t.Fatal("chose a worker that did not start")
	}

	n := chooseNativeCapture("on", lib(true, nil))
	if n == nil {
		t.Fatal("did not choose a worker that started")
	}
	defer n.close()
	data, result, err := n.read(nativeRequest{})
	if err != nil || result != 4 || len(data) != 2 {
		t.Fatalf("read through the chosen worker = %v, %d, %v", data, result, err)
	}
}
