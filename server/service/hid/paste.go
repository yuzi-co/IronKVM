package hid

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/inputcontrol"
)

const (
	defaultPasteDelay = 30 * time.Millisecond
	minPasteDelay     = 10 * time.Millisecond
	maxPasteDelay     = 500 * time.Millisecond

	// maxPasteRunes bounds a paste. At the default delay it takes about ten
	// minutes, which the background job, its progress and its cancel make
	// bearable; beyond it a file transfer is the better tool.
	maxPasteRunes = 20000

	// pasteCancelWait bounds how long a cancel waits for the job to stop. The
	// job checks between key presses, so it stops within one delay.
	pasteCancelWait = 5 * time.Second
)

const (
	pasteStatusIdle     = "idle"
	pasteStatusTyping   = "typing"
	pasteStatusDone     = "done"
	pasteStatusCanceled = "canceled"
	pasteStatusFailed   = "failed"

	pasteErrorControlBusy = "control_busy"
	pasteErrorHID         = "hid_error"
)

var (
	errPasteInProgress  = errors.New("a paste is already in progress")
	errNoPaste          = errors.New("no paste in progress")
	errPasteCanceled    = errors.New("paste canceled")
	errPasteControlBusy = errors.New("HID control is busy")
	errPasteCancelWait  = errors.New("cancel paste timed out")
)

var keyUpReport = []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}

// pasteTyper presses keys for a paste job.
type pasteTyper interface {
	// typeStep presses and releases the keys of one character, waiting delay
	// between them. It holds the keyboard only while it does, so a control
	// mode switch waits at most one character, not the whole paste.
	typeStep(ctx context.Context, strokes []Char, delay time.Duration) error
	// close releases the keys and the keyboard at the end of the job.
	close()
}

// pasteManager runs one paste at a time in the background and keeps its
// progress for the status and cancel requests.
type pasteManager struct {
	mu     sync.Mutex
	cancel context.CancelCauseFunc
	done   chan struct{}
	status proto.PasteStatusRsp
}

func newPasteManager() *pasteManager {
	return &pasteManager{status: proto.PasteStatusRsp{Status: pasteStatusIdle}}
}

// start begins typing plan with typer, and returns once the job is running.
func (m *pasteManager) start(plan pastePlan, delay time.Duration, typer pasteTyper) (proto.PasteStatusRsp, error) {
	m.mu.Lock()
	if m.done != nil {
		status := m.status
		m.mu.Unlock()
		return status, errPasteInProgress
	}

	ctx, cancel := context.WithCancelCause(context.Background())
	done := make(chan struct{})
	m.cancel = cancel
	m.done = done
	m.status = proto.PasteStatusRsp{Status: pasteStatusTyping, Total: len(plan.steps)}
	status := m.status
	m.mu.Unlock()

	go m.run(ctx, done, plan, delay, typer)
	return status, nil
}

func (m *pasteManager) run(ctx context.Context, done chan struct{}, plan pastePlan, delay time.Duration, typer pasteTyper) {
	var err error
	for i, step := range plan.steps {
		if i > 0 {
			if err = sleepPasteContext(ctx, delay); err != nil {
				break
			}
		}
		if err = context.Cause(ctx); err != nil {
			break
		}
		if err = typer.typeStep(ctx, step.strokes, delay); err != nil {
			break
		}
		m.setTyped(done, i+1)
	}
	typer.close()
	m.finish(done, err)
}

func (m *pasteManager) setTyped(done chan struct{}, typed int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.done == done {
		m.status.Typed = typed
	}
}

func (m *pasteManager) finish(done chan struct{}, err error) {
	m.mu.Lock()
	if m.done != done {
		m.mu.Unlock()
		return
	}

	switch {
	case err == nil:
		m.status.Status = pasteStatusDone
		log.Debugf("hid paste success, %d characters typed", m.status.Typed)
	case errors.Is(err, errPasteCanceled):
		m.status.Status = pasteStatusCanceled
		log.Debugf("hid paste canceled after %d of %d characters", m.status.Typed, m.status.Total)
	case errors.Is(err, errPasteControlBusy):
		m.status.Status = pasteStatusFailed
		m.status.Error = pasteErrorControlBusy
		log.Errorf("hid paste failed: %v", err)
	default:
		m.status.Status = pasteStatusFailed
		m.status.Error = pasteErrorHID
		log.Errorf("hid paste failed: %v", err)
	}
	m.cancel(nil)
	m.cancel = nil
	m.done = nil
	m.mu.Unlock()

	close(done)
}

// stop cancels the running paste and waits for it to release the keyboard.
func (m *pasteManager) stop(wait time.Duration) (proto.PasteStatusRsp, error) {
	m.mu.Lock()
	cancel := m.cancel
	done := m.done
	m.mu.Unlock()

	if cancel == nil || done == nil {
		return m.getStatus(), errNoPaste
	}

	cancel(errPasteCanceled)
	select {
	case <-done:
		return m.getStatus(), nil
	case <-time.After(wait):
		return m.getStatus(), errPasteCancelWait
	}
}

func (m *pasteManager) getStatus() proto.PasteStatusRsp {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.status
}

// manualTyper types through a manual input session, which is how the browser's
// own key presses reach the target. It takes the keyboard for one character at
// a time.
type manualTyper struct {
	session *inputcontrol.ManualSession
	write   func(report []byte) error
}

func (s *Service) newPasteTyper() *manualTyper {
	session := s.newManualSession()
	return &manualTyper{
		session: session,
		write: func(report []byte) error {
			return session.Execute(func() error {
				return s.hid.WriteKeyboardReport(report)
			})
		},
	}
}

func (t *manualTyper) typeStep(ctx context.Context, strokes []Char, delay time.Duration) error {
	reservation, err := t.session.Reserve(ctx, inputcontrol.ManualKeyboard, false, nil)
	if err != nil {
		if cause := context.Cause(ctx); cause != nil {
			return cause
		}
		return fmt.Errorf("%w: %w", errPasteControlBusy, err)
	}

	// The keys of one character go out together, even when a cancel arrives
	// between them: a dead key left pending would put its accent on whatever
	// the user types next.
	for i, stroke := range strokes {
		if i > 0 {
			time.Sleep(delay)
		}
		keyDown := []byte{byte(stroke.Modifiers), 0x00, byte(stroke.Code), 0x00, 0x00, 0x00, 0x00, 0x00}
		if err = t.write(keyDown); err != nil {
			break
		}
		if err = t.write(keyUpReport); err != nil {
			break
		}
	}
	if err != nil {
		_ = t.write(keyUpReport)
	}
	reservation.Complete(err == nil)
	return err
}

func (t *manualTyper) close() {
	t.session.Close()
}

// pasteLayout returns the layout a request names. Layout wins over the older
// Langue.
func pasteLayout(layout string, langue string) (*Layout, error) {
	if layout == "" {
		layout = langue
	}
	return GetLayout(layout)
}

// pasteDelay returns the delay a request asks for, or the default for 0.
func pasteDelay(ms int) (time.Duration, error) {
	if ms == 0 {
		return defaultPasteDelay, nil
	}
	delay := time.Duration(ms) * time.Millisecond
	if delay < minPasteDelay || delay > maxPasteDelay {
		return 0, fmt.Errorf("delay must be between %d and %d ms", minPasteDelay.Milliseconds(), maxPasteDelay.Milliseconds())
	}
	return delay, nil
}

func checkRsp(plan pastePlan, delay time.Duration) *proto.PasteCheckRsp {
	untypeable := plan.untypeable
	if untypeable == nil {
		untypeable = []proto.PasteUntypeable{}
	}
	return &proto.PasteCheckRsp{
		Characters:      len(plan.steps),
		Keystrokes:      plan.keystrokes,
		DurationMs:      (time.Duration(plan.keystrokes) * delay).Milliseconds(),
		Untypeable:      untypeable,
		UntypeableCount: plan.untypeableCount,
	}
}

// Paste starts typing a text on the target and answers at once. The status
// request reports the progress, and the cancel request stops it.
func (s *Service) Paste(c *gin.Context) {
	var req proto.PasteReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	layout, err := pasteLayout(req.Layout, req.Langue)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	delay, err := pasteDelay(req.Delay)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	if len([]rune(req.Content)) > maxPasteRunes {
		rsp.ErrRsp(c, -2, fmt.Sprintf("content too long, at most %d characters", maxPasteRunes))
		return
	}

	plan := planPaste(req.Content, layout)
	if plan.untypeableCount > 0 && !req.SkipUntypeable {
		rsp.Data = checkRsp(plan, delay)
		rsp.ErrRsp(c, -4, "the layout cannot type some characters")
		return
	}
	if len(plan.steps) == 0 {
		rsp.ErrRsp(c, -4, "nothing to type")
		return
	}

	typer := s.newPasteTyper()
	status, err := s.paste.start(plan, delay, typer)
	if err != nil {
		typer.close()
		rsp.Data = status
		rsp.ErrRsp(c, -3, err.Error())
		return
	}

	rsp.OkRspWithData(c, status)
	log.Debugf("hid paste started, %d characters on layout %s", len(plan.steps), layout.ID)
}

// CheckPaste reports what typing a text would take on a layout: how long, and
// which characters the layout cannot type.
func (s *Service) CheckPaste(c *gin.Context) {
	var req proto.PasteCheckReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	layout, err := GetLayout(req.Layout)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	delay, err := pasteDelay(req.Delay)
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	if len([]rune(req.Content)) > maxPasteRunes {
		rsp.ErrRsp(c, -2, fmt.Sprintf("content too long, at most %d characters", maxPasteRunes))
		return
	}

	rsp.OkRspWithData(c, checkRsp(planPaste(req.Content, layout), delay))
}

// GetPasteStatus reports the paste typing now, or the last one.
func (s *Service) GetPasteStatus(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, s.paste.getStatus())
}

// CancelPaste stops the paste typing now.
func (s *Service) CancelPaste(c *gin.Context) {
	var rsp proto.Response

	status, err := s.paste.stop(pasteCancelWait)
	if err != nil {
		rsp.Data = status
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	rsp.OkRspWithData(c, status)
}

func sleepPasteContext(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-timer.C:
		return nil
	}
}
