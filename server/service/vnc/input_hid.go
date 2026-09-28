package vnc

import (
	"context"
	"errors"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"
	"NanoKVM-Server/service/picoclaw"
	"NanoKVM-Server/service/vm/jiggler"
)

// manualPreemptTimeout matches the web UI websocket's: how long one report may
// wait for an AI operation to give the HID back.
const manualPreemptTimeout = 2 * time.Second

// hidQueueLength matches the web UI websocket's queues.
const hidQueueLength = 200

// hidInput sends a session's reports through the same HID writers and the same
// manual-input reservation as the web UI websocket (service/ws), so VNC input
// is arbitrated against MCP and PicoClaw exactly as a browser's is.
type hidInput struct {
	manual   *inputcontrol.ManualSession
	keyboard chan hid.QueuedReport
	mouse    chan hid.QueuedReport
	workers  sync.WaitGroup
	once     sync.Once
}

// NewHIDInput opens the HID devices and starts the writers for one session.
func NewHIDInput() Input {
	h := hid.GetHid()
	h.Open()

	input := &hidInput{
		manual:   inputcontrol.NewManualSession(controlmode.GetManager(), inputcontrol.GetCoordinator()),
		keyboard: make(chan hid.QueuedReport, hidQueueLength),
		mouse:    make(chan hid.QueuedReport, hidQueueLength),
	}

	input.workers.Add(2)
	go func() {
		defer input.workers.Done()
		h.KeyboardReports(input.keyboard)
	}()
	go func() {
		defer input.workers.Done()
		h.MouseReports(input.mouse)
	}()

	return input
}

func (i *hidInput) Keyboard(report []byte) {
	held := false
	for _, b := range report {
		if b != 0 {
			held = true
			break
		}
	}
	i.queue(i.keyboard, inputcontrol.ManualKeyboard, report, held, true)
}

func (i *hidInput) Pointer(report []byte) {
	// Movement alone is not input that suspends the jiggler, as in the
	// websocket. A button or a wheel step is.
	startCooldown := report[0] != 0 || report[5] != 0
	i.queue(i.mouse, inputcontrol.ManualAbsoluteMouse, report, report[0] != 0, startCooldown)
}

// queue reserves the HID for manual input and hands the report to its writer.
// It is the websocket's queueManualReport, with the same PicoClaw check.
func (i *hidInput) queue(queue chan hid.QueuedReport, kind inputcontrol.ManualReportKind, report []byte, held bool, startCooldown bool) {
	ctx, cancel := context.WithTimeout(context.Background(), manualPreemptTimeout)
	defer cancel()

	reservation, err := i.manual.ReserveWithCooldown(ctx, kind, held, startCooldown, func(mode controlmode.Mode) bool {
		return mode != controlmode.ModePicoclaw || !picoclaw.GetSessionLock().BlocksManualInput()
	})
	if err != nil {
		if errors.Is(err, inputcontrol.ErrManualInputBlocked) {
			log.Debug("vnc: HID input dropped while PicoClaw session holds control")
		} else {
			log.Errorf("vnc: HID input failed to acquire control: %s", err)
		}
		return
	}

	queued := hid.QueuedReport{
		Data:               report,
		Execute:            i.manual.Execute,
		Complete:           reservation.Complete,
		ResetKeyboard:      func() { i.manual.Reset(inputcontrol.ManualKeyboard) },
		ResetRelativeMouse: func() { i.manual.Reset(inputcontrol.ManualRelativeMouse) },
		ResetAbsoluteMouse: func() { i.manual.Reset(inputcontrol.ManualAbsoluteMouse) },
	}
	if !sendQueue(queue, queued) {
		queued.Complete(false)
		return
	}
	jiggler.GetJiggler().Update()
}

// Close stops the writers, which release every key and button still held, and
// ends the manual reservation.
func (i *hidInput) Close() {
	i.once.Do(func() {
		close(i.keyboard)
		close(i.mouse)
	})
	i.workers.Wait()
	i.manual.Close()
}

// sendQueue hands a report to a writer without blocking, as the websocket
// does: when the queue is full the stalest report goes, because each report
// carries absolute state and the newest one describes the truth on its own.
// The evicted report is completed, so its reservation is released.
func sendQueue(queue chan hid.QueuedReport, report hid.QueuedReport) bool {
	for range cap(queue) + 1 {
		select {
		case queue <- report:
			return true
		default:
		}

		select {
		case stale := <-queue:
			if stale.Complete != nil {
				stale.Complete(false)
			}
		default:
		}
	}
	return false
}
