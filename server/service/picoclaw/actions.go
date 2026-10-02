package picoclaw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"NanoKVM-Server/service/hid"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	defaultClickHold  = 40 * time.Millisecond
	defaultKeyDelay   = 30 * time.Millisecond
	defaultDragSteps  = 10
	defaultScrollStep = 20 * time.Millisecond
)

func (s *Service) Actions(c *gin.Context) {
	releaseMode, modeErr := s.acquireControlMode()
	if modeErr != nil {
		writePicoclawError(c, modeErr)
		return
	}
	defer releaseMode()

	operationCtx, releaseOperation := s.beginControlOperation(c.Request.Context())
	defer releaseOperation()

	sessionID, sessionErr := s.requireSessionID(c)
	if sessionErr != nil {
		writePicoclawError(c, sessionErr)
		return
	}

	releaseAfter, lockErr := s.lock.AcquireTemporary(sessionID)
	if lockErr != nil {
		writePicoclawError(c, lockErr)
		return
	}
	if releaseAfter {
		s.pointer.forget()
		defer s.lock.Release(sessionID)
	}

	actions, actionErr := normalizeActions(c)
	if actionErr != nil {
		writePicoclawError(c, actionErr)
		return
	}

	result, execErr := s.executeActions(operationCtx, sessionID, actions)
	if execErr != nil {
		writePicoclawError(c, execErr)
		return
	}

	writeSuccess(c, result)
}

func normalizeActions(c *gin.Context) ([]Action, *PicoclawError) {
	body := bytes.NewBuffer(nil)
	if _, err := body.ReadFrom(c.Request.Body); err != nil {
		return nil, newPicoclawError(CodeInvalidAction, "failed to read action payload")
	}

	raw := body.Bytes()
	if len(raw) == 0 {
		return nil, newPicoclawError(CodeInvalidAction, "empty action payload")
	}

	var batch ActionBatch
	if err := json.Unmarshal(raw, &batch); err == nil && len(batch.Actions) > 0 {
		return batch.Actions, nil
	}

	var action Action
	if err := json.Unmarshal(raw, &action); err != nil || action.Action == "" {
		return nil, newPicoclawError(CodeInvalidAction, "invalid action payload")
	}

	return []Action{action}, nil
}

func (s *Service) executeActions(ctx context.Context, sessionID string, actions []Action) (result ActionResult, err *PicoclawError) {
	startedAt := time.Now()
	if len(actions) == 0 {
		return ActionResult{}, newPicoclawError(CodeInvalidAction, "empty actions")
	}

	defer func() {
		if err != nil {
			if releaseErr := s.releaseHeldInput(); releaseErr != nil {
				log.Warnf("PicoClaw action failed with input still held on the host: %v", releaseErr)
			}
		}
	}()

	totalWrites := 0
	for idx, action := range actions {
		if contextErr := controlOperationError(ctx); contextErr != nil {
			contextErr.Index = &idx
			return ActionResult{}, contextErr
		}
		acquired, lockErr := s.lock.acquire(sessionID)
		if lockErr != nil {
			lockErr.Index = &idx
			return ActionResult{}, lockErr
		}
		if acquired {
			// Someone else may have moved the pointer while nobody held the lock.
			s.pointer.forget()
		}

		writes, execErr := s.executeAction(ctx, action)
		if execErr != nil {
			execErr.Index = &idx
			return ActionResult{}, execErr
		}
		totalWrites += writes
	}

	result = ActionResult{
		Action:          actions[0].Action,
		DurationMs:      time.Since(startedAt).Milliseconds(),
		HIDWrites:       totalWrites,
		ExecutedActions: len(actions),
		Pointer:         s.pointerPosition(),
	}
	if len(actions) > 1 {
		result.Action = "batch"
	}

	return result, nil
}

func (s *Service) executeAction(ctx context.Context, action Action) (int, *PicoclawError) {
	if err := controlOperationError(ctx); err != nil {
		return 0, err
	}

	out := &hidOutput{s: s}
	switch strings.ToLower(strings.TrimSpace(action.Action)) {
	case "click":
		x, y, err := s.clickTarget(action)
		if err != nil {
			return 0, err
		}
		button, err := mouseButton(action.Button)
		if err != nil {
			return 0, err
		}

		out.mouse(x, y, 0x00, 0)
		out.mouse(x, y, button, 0)
		if out.err != nil {
			return out.writes, out.err
		}
		if waitErr := waitForControlOperation(ctx, defaultClickHold); waitErr != nil {
			return out.writes, waitErr
		}
		out.mouse(x, y, 0x00, 0)
		return out.writes, out.err

	case "move":
		x, y, err := s.moveTarget(action)
		if err != nil {
			return 0, err
		}
		out.mouse(x, y, 0x00, 0)
		return out.writes, out.err

	case "wait":
		if action.DurationMs < 0 {
			return 0, newPicoclawError(CodeInvalidAction, "wait duration must be >= 0")
		}
		if action.DurationMs > maxWaitDurationMS {
			return 0, newPicoclawError(CodeInvalidAction, "wait duration must be <= 30000 milliseconds")
		}
		if waitErr := waitForControlOperation(ctx, time.Duration(action.DurationMs)*time.Millisecond); waitErr != nil {
			return 0, waitErr
		}
		return 0, nil

	case "drag":
		fromX, fromY, err := normalizedNestedPoint("drag \"from\"", action.From)
		if err != nil {
			return 0, err
		}
		toX, toY, err := normalizedNestedPoint("drag \"to\"", action.To)
		if err != nil {
			return 0, err
		}
		button, err := mouseButton(action.Button)
		if err != nil {
			return 0, err
		}

		out.mouse(fromX, fromY, 0x00, 0)
		out.mouse(fromX, fromY, button, 0)
		for step := 1; step <= defaultDragSteps && out.err == nil; step++ {
			if contextErr := controlOperationError(ctx); contextErr != nil {
				return out.writes, contextErr
			}
			ratio := float64(step) / float64(defaultDragSteps)
			x := fromX + (toX-fromX)*ratio
			y := fromY + (toY-fromY)*ratio
			out.mouse(x, y, button, 0)
		}
		out.mouse(toX, toY, 0x00, 0)
		return out.writes, out.err

	case "scroll":
		// Scroll where the pointer is, or at the centre when that is unknown.
		x, y, known := s.pointer.get()
		if !known {
			x, y = 0.5, 0.5
		}
		if action.X != nil || action.Y != nil {
			var err *PicoclawError
			x, y, err = normalizedPointFor("scroll", action.X, action.Y)
			if err != nil {
				return 0, err
			}
		}

		amount := action.Amount
		if amount == 0 {
			amount = 1
		}
		if amount < 0 {
			return 0, newPicoclawError(CodeInvalidAction, "scroll amount must be > 0")
		}

		wheel := 1
		switch strings.ToLower(strings.TrimSpace(action.Direction)) {
		case "", "up":
			wheel = 1
		case "down":
			wheel = -1
		default:
			return 0, newPicoclawError(CodeInvalidAction, "invalid scroll direction")
		}

		for range amount {
			if contextErr := controlOperationError(ctx); contextErr != nil {
				return out.writes, contextErr
			}
			out.mouse(x, y, 0x00, wheel)
			out.mouse(x, y, 0x00, 0)
			if out.err != nil {
				return out.writes, out.err
			}
			if waitErr := waitForControlOperation(ctx, defaultScrollStep); waitErr != nil {
				return out.writes, waitErr
			}
		}
		return out.writes, nil

	case "type":
		if action.Text == "" {
			return 0, newPicoclawError(CodeInvalidAction, typeNeedsTextMessage)
		}
		charMap := hid.GetCharMap("")
		for _, char := range action.Text {
			if contextErr := controlOperationError(ctx); contextErr != nil {
				return out.writes, contextErr
			}
			key, ok := charMap[char]
			if !ok {
				return out.writes, newPicoclawError(CodeInvalidAction, "unsupported character in type action")
			}

			out.key([]byte{byte(key.Modifiers), 0x00, byte(key.Code), 0x00, 0x00, 0x00, 0x00, 0x00})
			out.key([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
			if out.err != nil {
				return out.writes, out.err
			}
			if waitErr := waitForControlOperation(ctx, defaultKeyDelay); waitErr != nil {
				return out.writes, waitErr
			}
		}
		return out.writes, nil

	case "hotkey":
		report, err := buildHotkeyReport([]string(action.Keys))
		if err != nil {
			return 0, err
		}
		out.key(report)
		if out.err != nil {
			return out.writes, out.err
		}
		if waitErr := waitForControlOperation(ctx, defaultClickHold); waitErr != nil {
			return out.writes, waitErr
		}
		out.key([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
		return out.writes, out.err
	}

	return 0, newPicoclawError(CodeInvalidAction, fmt.Sprintf(`unknown action %q; "action" must be one of click, move, type, hotkey, scroll, drag, wait`, action.Action))
}

const typeNeedsTextMessage = "type requires text"

func toAbsoluteHidCoord(normalized float64) uint16 {
	if normalized < 0 {
		normalized = 0
	}
	if normalized > 1 {
		normalized = 1
	}
	return uint16(math.Floor(0x7FFF*normalized)) + 1
}

// hidOutput sends the reports of one action and stops at the first one that
// does not reach the host: a later report would act on a host that missed
// the earlier ones, for example a button release without its press.
type hidOutput struct {
	s      *Service
	writes int
	err    *PicoclawError
}

func (o *hidOutput) mouse(x float64, y float64, buttons byte, wheel int) {
	if o.err != nil {
		return
	}
	if o.err = o.s.sendMouseMoveWithButton(x, y, buttons, wheel); o.err == nil {
		o.writes++
	}
}

func (o *hidOutput) key(report []byte) {
	if o.err != nil {
		return
	}
	if o.err = o.s.sendKeyboardReport(report); o.err == nil {
		o.writes++
	}
}

func absoluteMouseReport(x float64, y float64, buttons byte, wheel int) []byte {
	absoluteX := toAbsoluteHidCoord(x)
	absoluteY := toAbsoluteHidCoord(y)

	return []byte{
		buttons,
		byte(absoluteX & 0xff),
		byte(absoluteX >> 8),
		byte(absoluteY & 0xff),
		byte(absoluteY >> 8),
		byte(int8(wheel)),
	}
}

// sendMouseMoveWithButton writes one absolute pointer report. The pointer
// position is remembered only once the report has reached the host, so a
// failed write leaves the next relative move starting from where the
// pointer really is.
func (s *Service) sendMouseMoveWithButton(x float64, y float64, buttons byte, wheel int) *PicoclawError {
	if err := s.hid.WriteAbsoluteMouseReport(absoluteMouseReport(x, y, buttons, wheel)); err != nil {
		return hidWriteError("mouse", err)
	}
	s.pointer.set(clampUnit(x), clampUnit(y))
	s.held.setButtons(buttons != 0)
	return nil
}

func (s *Service) sendKeyboardReport(report []byte) *PicoclawError {
	if err := s.hid.WriteKeyboardReport(report); err != nil {
		return hidWriteError("keyboard", err)
	}
	s.held.setKeys(!bytes.Equal(report, make([]byte, len(report))))
	return nil
}

func hidWriteError(device string, err error) *PicoclawError {
	return newPicoclawError(CodeHIDWriteFailed, fmt.Sprintf(
		"the %s report did not reach the remote host (%v), so this action did not take effect", device, err))
}

func normalizedNestedPoint(name string, point *Point) (float64, float64, *PicoclawError) {
	if point == nil {
		return 0, 0, newPicoclawError(CodeInvalidAction, name+` needs an object with "x" and "y", e.g. {"x":0.2,"y":0.3}. `+coordinateHelp)
	}
	return normalizedPointFor(name, point.X, point.Y)
}

// clickTarget resolves a click to a point: the given "x" and "y", or where
// the pointer is when the click has no coordinates.
func (s *Service) clickTarget(action Action) (float64, float64, *PicoclawError) {
	if action.DX != nil || action.DY != nil {
		return 0, 0, newPicoclawError(CodeInvalidAction, `click does not take "dx"/"dy". Move with "dx"/"dy" first, then click without coordinates to click where the pointer is, or click with "x" and "y". `+coordinateHelp)
	}
	if action.X == nil && action.Y == nil {
		if x, y, known := s.pointer.get(); known {
			return x, y, nil
		}
	}
	return normalizedPointFor("click", action.X, action.Y)
}

func mouseButton(button string) (byte, *PicoclawError) {
	switch strings.ToLower(strings.TrimSpace(button)) {
	case "", "left":
		return 1 << 0, nil
	case "right":
		return 1 << 1, nil
	case "middle":
		return 1 << 2, nil
	case "back":
		return 1 << 3, nil
	case "forward":
		return 1 << 4, nil
	default:
		return 0, newPicoclawError(CodeInvalidAction, "invalid mouse button")
	}
}
