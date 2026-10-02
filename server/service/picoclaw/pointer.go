package picoclaw

import (
	"fmt"
	"math"
	"strconv"
	"sync"

	"NanoKVM-Server/common"
)

// pointerTracker remembers where PicoClaw last put the mouse pointer through
// the absolute mouse, so that a move can be relative to it. While PicoClaw
// holds the session lock nobody else moves the pointer; the position is
// forgotten whenever the lock is taken afresh or the control mode changes.
type pointerTracker struct {
	mu    sync.Mutex
	known bool
	x, y  float64
}

func (p *pointerTracker) set(x, y float64) {
	p.mu.Lock()
	p.known, p.x, p.y = true, x, y
	p.mu.Unlock()
}

func (p *pointerTracker) get() (float64, float64, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.x, p.y, p.known
}

func (p *pointerTracker) forget() {
	p.mu.Lock()
	p.known = false
	p.mu.Unlock()
}

// screenPixelSize is the size of the remote screen in the pixels the model is
// told about: the full-size screenshot, as the screenshot caption reports it.
func screenPixelSize() (int, int) {
	values := common.GetScreen().Snapshot()
	width, height := int(values.Width), int(values.Height)
	if width <= 0 || height <= 0 {
		return 1920, 1080
	}
	return width, height
}

func (s *Service) pointerPosition() *PointerPosition {
	x, y, known := s.pointer.get()
	if !known {
		return nil
	}
	width, height := screenPixelSize()
	return &PointerPosition{
		X:  math.Round(x*10000) / 10000,
		Y:  math.Round(y*10000) / 10000,
		PX: int(math.Round(x * float64(width-1))),
		PY: int(math.Round(y * float64(height-1))),
	}
}

const (
	coordinateHelp = `"x" and "y" are fractions of the screen from 0 to 1: {"x":0,"y":0} is the top-left corner, {"x":1,"y":1} the bottom-right, {"x":0.5,"y":0.5} the centre. ` +
		`For a point seen in a screenshot, divide its pixel position by the image width and height`
	moveHelp = `move takes either "x" and "y" to go to a point (fractions of the screen from 0 to 1, e.g. {"action":"move","x":0.5,"y":0.5} for the centre) ` +
		`or "dx" and/or "dy" to move from the current pointer position by that many screen pixels (e.g. {"action":"move","dx":10} for 10 pixels right; negative dx is left, negative dy is up)`
)

// moveTarget resolves a move action to an absolute point.
func (s *Service) moveTarget(action Action) (float64, float64, *PicoclawError) {
	absolute := action.X != nil || action.Y != nil
	relative := action.DX != nil || action.DY != nil

	switch {
	case absolute && relative:
		return 0, 0, newPicoclawError(CodeInvalidAction, `move got both "x"/"y" and "dx"/"dy"; use only one pair. `+moveHelp)
	case !absolute && !relative:
		return 0, 0, newPicoclawError(CodeInvalidAction, `move needs coordinates. `+moveHelp)
	case absolute:
		return normalizedPointFor("move", action.X, action.Y)
	}

	currentX, currentY, known := s.pointer.get()
	if !known {
		return 0, 0, newPicoclawError(CodeInvalidAction,
			`move with "dx"/"dy" needs to know where the pointer is, and it is not known yet. `+
				`First move or click with "x" and "y" (fractions of the screen from 0 to 1), then move relative to that`)
	}
	width, height := screenPixelSize()
	x, y := currentX, currentY
	if action.DX != nil {
		if !isFinite(*action.DX) {
			return 0, 0, newPicoclawError(CodeInvalidAction, `"dx" must be a number of screen pixels`)
		}
		x = clampUnit(x + *action.DX/float64(width-1))
	}
	if action.DY != nil {
		if !isFinite(*action.DY) {
			return 0, 0, newPicoclawError(CodeInvalidAction, `"dy" must be a number of screen pixels`)
		}
		y = clampUnit(y + *action.DY/float64(height-1))
	}
	return x, y, nil
}

func normalizedPointFor(name string, x *float64, y *float64) (float64, float64, *PicoclawError) {
	switch {
	case x == nil && y == nil:
		return 0, 0, newPicoclawError(CodeInvalidAction, name+` needs "x" and "y". `+coordinateHelp)
	case x == nil:
		return 0, 0, newPicoclawError(CodeInvalidAction, name+` needs "x" as well as "y". `+coordinateHelp)
	case y == nil:
		return 0, 0, newPicoclawError(CodeInvalidAction, name+` needs "y" as well as "x". `+coordinateHelp)
	}
	if err := unitCoordinate(name, "x", *x); err != nil {
		return 0, 0, err
	}
	if err := unitCoordinate(name, "y", *y); err != nil {
		return 0, 0, err
	}
	return *x, *y, nil
}

func unitCoordinate(name string, field string, value float64) *PicoclawError {
	if isFinite(value) && value >= 0 && value <= 1 {
		return nil
	}
	message := fmt.Sprintf(`%s: "%s" is %s, but it must be a fraction from 0 to 1, not pixels. %s`,
		name, field, strconv.FormatFloat(value, 'f', -1, 64), coordinateHelp)
	if name == "move" {
		message += `. To move by a number of pixels from where the pointer is, use "dx"/"dy" instead`
	}
	return newPicoclawError(CodeInvalidAction, message)
}

func isFinite(value float64) bool {
	return !math.IsNaN(value) && !math.IsInf(value, 0)
}

func clampUnit(value float64) float64 {
	return math.Min(1, math.Max(0, value))
}
