package main

import (
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/config"
	"NanoKVM-Server/logger"
	"NanoKVM-Server/middleware"
	"NanoKVM-Server/router"
	"NanoKVM-Server/service/application"
	"NanoKVM-Server/service/ion"
	"NanoKVM-Server/service/stream/webrtc"
	"NanoKVM-Server/service/vm"
	"NanoKVM-Server/service/vm/jiggler"
	"NanoKVM-Server/utils"

	"github.com/gin-gonic/gin"
	cors "github.com/rs/cors/wrapper/gin"
)

func main() {
	initialize()
	defer dispose()

	run()
}

func initialize() {
	if err := config.EnsurePicoclawInternalToken(); err != nil {
		log.Fatalf("failed to initialize picoclaw internal token: %v", err)
	}

	logger.Init()

	// Record the carveout baseline and reset the peak watermark here, before
	// the first call that can allocate from it. That call is
	// vm.EnableHdmiCapture() below, which reaches libkvm through
	// common.GetKvmVision(); a baseline recorded any later would already
	// include this process's own capture working set and understate what a
	// restart re-pays - the direction the design calls fatal.
	ion.Init(config.GetInstance().Ion.ReserveFloor)

	// restore the memory limit the user configured, which is otherwise only
	// applied to the process that set it and lost on the next boot
	utils.InitGoMemLimit()

	// End any update stand-off. S98supervise leaves the board alone while an
	// update is replacing /kvmapp, and the updater cannot lift that itself: the
	// restart is inside the window, so the process that would clean up is the
	// one being replaced. A server reaching this line is the proof it finished.
	application.ClearUpdateMarker()

	// init screen parameters
	_ = common.GetScreen()

	// init HDMI
	//
	// There used to be a DisableHdmiCapture() and a 10ms sleep in front of
	// this. It was a power cycle of the HDMI receiver, and it did nothing
	// useful in either direction. On alpha and beta boards kvmv_hdmi_control
	// declines the call outright, so the pair was dead code. On the PCIe board
	// it did toggle the receiver, 10ms apart, which is a number copied from
	// libkvm rather than one the receiver was measured against - and
	// EnableHdmiCapture powers the receiver on anyway, so the cycle added a
	// teardown nothing had asked for. `Settings > Reset HDMI` exists for a
	// deliberate cycle and waits a full second between the halves.
	//
	// The idle bookkeeping that DisableHdmiCapture also does is all zero-valued
	// in a process that has only just started, so dropping the call loses it
	// nothing.
	if !utils.IsHdmiDisabled() {
		vm.EnableHdmiCapture()
	}
	vm.SetHdmiViewerCount(0)

	// run mouse jiggler
	jiggler.GetJiggler().Run()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		sig := <-sigChan
		log.Printf("\nReceived signal: %v\n", sig)

		if !disposeWithin(disposeTimeout, dispose) {
			log.Printf("capture teardown did not finish in %s, exiting anyway\n", disposeTimeout)
		}
		os.Exit(0)
	}()
}

func run() {
	conf := config.GetInstance()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	if conf.Authentication == "disable" {
		r.Use(cors.AllowAll())
	}

	router.Init(r)

	httpAddr := utils.ListenAddr(conf.Host, strconv.Itoa(conf.Port.Http))
	loopbackHTTPAddr := utils.ListenAddr("127.0.0.1", strconv.Itoa(conf.Port.Http))
	needsLoopbackHTTP := utils.NeedsDedicatedLoopbackListener(conf.Host)

	if conf.Proto == "https" {
		httpsPortStr := strconv.Itoa(conf.Port.Https)

		go func() {
			server := utils.NewServer(utils.ListenAddr(conf.Host, httpsPortStr), r)
			if err := server.ListenAndServeTLS(conf.Cert.Crt, conf.Cert.Key); err != nil {
				panic("start https server failed")
			}
		}()

		if needsLoopbackHTTP {
			go func() {
				if err := middleware.ListenAndServeLoopbackHTTPRedirect(
					loopbackHTTPAddr,
					httpsPortStr,
					r,
					router.LoopbackHTTPAllowedPaths()...,
				); err != nil {
					panic("start loopback http server failed")
				}
			}()
		}

		if err := middleware.ListenAndServeLoopbackHTTPRedirect(
			httpAddr,
			httpsPortStr,
			r,
			router.LoopbackHTTPAllowedPaths()...,
		); err != nil {
			panic("start http server failed")
		}
	} else {
		if needsLoopbackHTTP {
			go func() {
				if err := utils.NewServer(loopbackHTTPAddr, r).ListenAndServe(); err != nil {
					panic("start loopback http server failed")
				}
			}()
		}

		if err := utils.NewServer(httpAddr, r).ListenAndServe(); err != nil {
			panic("start http server failed")
		}
	}
}

func dispose() {
	// Stop the audio child before this process goes away. It does not follow us
	// out: arecord sees the closed pipe only when it writes, and while the host
	// plays nothing it blocks in the ALSA read forever. The orphan keeps the
	// capture card open, so the next server cannot record and audio stays dead
	// until somebody kills it by hand.
	//
	// This covers SIGTERM, which is what `S95nanokvm restart` and the in-place
	// updater send. S95nanokvm kills arecord as well, because a SIGKILL or a
	// crash reaches no code in this process.
	before := ion.Read()
	log.Printf("teardown: carveout before: used=%d free=%d generations=%d",
		before.Used, before.Free, before.Generations)

	teardownStep("audio capture", webrtc.StopAudioCapture)
	teardownStep("capture pipeline", func() { common.GetKvmVision().Close() })

	after := ion.Read()
	log.Printf("teardown: carveout after: used=%d free=%d generations=%d",
		after.Used, after.Free, after.Generations)
}

// teardownStep runs one step of dispose and records both ends of it.
//
// The carveout readings on either side say whether the teardown gave anything
// back, and these say which step spent the budget when it does not finish. Only
// the opening line survives a step that never returns, because the process
// exits where it stands when disposeTimeout expires, so the last line in the
// log names the step that was still running.
func teardownStep(name string, run func()) {
	log.Printf("teardown: %s started", name)

	started := time.Now()
	run()

	log.Printf("teardown: %s done in %s", name, time.Since(started).Round(time.Millisecond))
}

// disposeTimeout bounds the teardown that runs when the process is asked to
// stop. It has to stay under the wait S95nanokvm allows after SIGTERM, which is
// ten seconds: a budget at or above that wait ends in SIGKILL, and a SIGKILL
// reaches no code at all.
//
// A teardown that works needs about a second of it. Measured on the device
// 2026-08-20 with the steps instrumented: audio child 0s, reader gate 0s,
// kvmv_deinit 823ms, and the carveout gave back 6,516,736 bytes - one VI channel
// pool and one ISP shared buffer, the exact amount a stop used to leak.
//
// It did not work before that day, and no budget would have rescued it. The
// teardown ran past 5s and past 8s without returning, because kvmv_deinit joins
// a libkvm thread whose loop could not exit: watchdog_sf_feed fell off the end
// of a function declared void*, so GCC treated the break as unreachable and
// deleted the test that reaches it. The exit flag was also a plain field. Both
// are fixed in support/sg2002/additional/kvm/src/kvm_vision.cpp, and
// tools/vidiag/test-libkvm-thread-exit.sh holds the shipped library to it.
//
// The margin over the measurement is deliberate. The teardown walks the whole
// capture pipeline down, and how long that takes depends on what the pipeline
// was doing; eight seconds covers a slow one and still leaves the init script
// two seconds of its own.
const disposeTimeout = 8 * time.Second

// disposeWithin runs teardown and returns whether it finished before timeout.
//
// dispose reaches libkvm through cgo, and kvmv_deinit joins libkvm's threads.
// One of those threads does not always return: auto_try_res spins while the
// HDMI input is unreadable and never tests try_exit_thread, so the join waits
// for a cable. The signal handler then never reaches os.Exit, and the process
// stays alive after a SIGTERM while it still owns the VI pipeline.
//
// S95nanokvm sends that signal and starts the next server, so the survivor is
// what makes the new one fail: its channel enable finds the carveout already
// committed and reports ENOMEM. Not leaving costs the restart, so this leaves.
//
// Leaving early costs a video buffer pool and an ISP shared buffer, and the
// driver does not hand either back on the next start. Only a reboot does. That
// is what every stop cost until 2026-08-20, when the teardown could not finish
// at all: kvmv_deinit joined a libkvm thread whose exit test the compiler had
// deleted. The measurement then was 6516736 bytes a stop, four times running.
//
// The teardown finishes now, in about 823ms, and a stop gives those bytes back.
// So this timeout is a bound on a teardown that works rather than a way out of
// one that cannot, and the cost of hitting it is real: what expires here is
// what leaks.
func disposeWithin(timeout time.Duration, teardown func()) bool {
	done := make(chan struct{})

	go func() {
		defer close(done)
		teardown()
	}()

	select {
	case <-done:
		return true
	case <-time.After(timeout):
		return false
	}
}
