package sequence

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"NanoKVM-Server/service/controlmode"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/inputcontrol"
	"NanoKVM-Server/service/vm/jiggler"
)

// reportWait bounds the wait for one report to reach the endpoint. A host that
// stops polling stalls the write, and the run then fails instead of waiting
// with a key held.
const reportWait = 2 * time.Second

var errReportTimeout = errors.New("the HID report was not written in time")

// output is where a run's steps go.
type output interface {
	// report writes one keyboard or pointer report and returns once the
	// endpoint has taken it.
	report(ctx context.Context, kind stepKind, report []byte) error
	// press holds a front-panel button for hold, then releases it.
	press(button string, hold time.Duration) error
	// close releases whatever the run still holds.
	close()
}

// hidOutput writes a run's reports the way a browser's websocket client does:
// through a manual input session, into a keyboard and a pointer queue that the
// HID workers drain. Closing the queues is what releases the keys and buttons
// still held, the same release a browser gets when its socket closes, so a run
// that ends for any reason leaves nothing down on the host.
type hidOutput struct {
	session  *inputcontrol.ManualSession
	keyboard chan hid.QueuedReport
	mouse    chan hid.QueuedReport
	workers  sync.WaitGroup
	once     sync.Once
	pressATX func(string, time.Duration) error
}

func newHIDOutput(press func(string, time.Duration) error) *hidOutput {
	h := hid.GetHid()
	o := &hidOutput{
		session:  inputcontrol.NewManualSession(controlmode.GetManager(), inputcontrol.GetCoordinator()),
		keyboard: make(chan hid.QueuedReport, 1),
		mouse:    make(chan hid.QueuedReport, 1),
		pressATX: press,
	}

	o.workers.Add(2)
	go func() {
		defer o.workers.Done()
		h.KeyboardReports(o.keyboard)
	}()
	go func() {
		defer o.workers.Done()
		h.MouseReports(o.mouse)
	}()
	return o
}

func (o *hidOutput) report(ctx context.Context, kind stepKind, report []byte) error {
	queue := o.mouse
	manual := inputcontrol.ManualRelativeMouse
	held := report[0] != 0
	switch kind {
	case stepKeyboard:
		queue = o.keyboard
		manual = inputcontrol.ManualKeyboard
		held = keyboardHeld(report)
	case stepAbsolute:
		manual = inputcontrol.ManualAbsoluteMouse
	}

	reservation, err := o.session.Reserve(ctx, manual, held, nil)
	if err != nil {
		if cause := context.Cause(ctx); cause != nil {
			return cause
		}
		return fmt.Errorf("%w: %w", errBusy, err)
	}

	written := make(chan bool, 1)
	queued := hid.QueuedReport{
		Data:    report,
		Execute: o.session.Execute,
		Complete: func(ok bool) {
			reservation.Complete(ok)
			written <- ok
		},
		ResetKeyboard:      func() { o.session.Reset(inputcontrol.ManualKeyboard) },
		ResetRelativeMouse: func() { o.session.Reset(inputcontrol.ManualRelativeMouse) },
		ResetAbsoluteMouse: func() { o.session.Reset(inputcontrol.ManualAbsoluteMouse) },
	}

	// The queue holds one report and this run waits for each, so the send
	// does not block for long; the context still bounds it.
	select {
	case queue <- queued:
	case <-ctx.Done():
		reservation.Complete(false)
		return context.Cause(ctx)
	}
	// A run's keys count as input, so the jigglers wait as they do for a
	// browser's.
	jiggler.GetJiggler().Update()

	timer := time.NewTimer(reportWait)
	defer timer.Stop()
	select {
	case ok := <-written:
		if !ok {
			return errHID
		}
		return nil
	case <-timer.C:
		return fmt.Errorf("%w: %w", errHID, errReportTimeout)
	case <-ctx.Done():
		// The report may still go out. Close waits for the worker, which
		// releases it if it held anything.
		return context.Cause(ctx)
	}
}

func (o *hidOutput) press(button string, hold time.Duration) error {
	if o.pressATX == nil {
		return fmt.Errorf("%w: no ATX control", errATX)
	}
	if err := o.pressATX(button, hold); err != nil {
		return fmt.Errorf("%w: %w", errATX, err)
	}
	return nil
}

func (o *hidOutput) close() {
	o.once.Do(func() {
		close(o.keyboard)
		close(o.mouse)
		o.workers.Wait()
		o.session.Close()
	})
}

func keyboardHeld(report []byte) bool {
	for i, b := range report {
		if i != 1 && b != 0 {
			return true
		}
	}
	return false
}
