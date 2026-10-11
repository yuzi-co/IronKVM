package sequence

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"NanoKVM-Server/proto"
)

func TestCompileStepsRefusesWhatItCannotPlay(t *testing.T) {
	cases := []struct {
		name string
		step proto.SequenceStep
	}{
		{"unknown type", proto.SequenceStep{Type: "type"}},
		{"seven keys", proto.SequenceStep{Type: StepKey, Keys: []int{4, 5, 6, 7, 8, 9, 10}}},
		{"error usage", proto.SequenceStep{Type: StepKey, Keys: []int{0x01}}},
		{"modifier as a key", proto.SequenceStep{Type: StepKey, Keys: []int{0xe0}}},
		{"usage over a byte", proto.SequenceStep{Type: StepKey, Keys: []int{0x100}}},
		{"modifiers out of range", proto.SequenceStep{Type: StepKey, Modifiers: 0x100}},
		{"negative pause", proto.SequenceStep{Type: StepDelay, Ms: -1}},
		{"long pause", proto.SequenceStep{Type: StepDelay, Ms: 60001}},
		{"unknown button", proto.SequenceStep{Type: StepATX, Button: "sleep"}},
		{"long press", proto.SequenceStep{Type: StepATX, Button: ButtonPower, Ms: 10001}},
		{"absolute off screen", proto.SequenceStep{Type: StepMouse, Absolute: true, X: 0x8000}},
		{"absolute horizontal wheel", proto.SequenceStep{Type: StepMouse, Absolute: true, HWheel: 1}},
		{"relative too far", proto.SequenceStep{Type: StepMouse, X: 128}},
		{"wheel too far", proto.SequenceStep{Type: StepMouse, Wheel: -128}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := compileSteps([]proto.SequenceStep{{Type: StepDelay, Ms: 1}, c.step})
			if err == nil {
				t.Fatal("the step was accepted")
			}
			if !strings.HasPrefix(err.Error(), "step 2:") {
				t.Fatalf("the error should name the step: %v", err)
			}
		})
	}

	if _, err := compileSteps(nil); err == nil {
		t.Error("an empty macro was accepted")
	}
	if _, err := compileSteps(make([]proto.SequenceStep, maxSteps+1)); err == nil {
		t.Error("a macro over the step limit was accepted")
	}
}

func TestATXStepDefaultsToAShortPress(t *testing.T) {
	steps, err := compileSteps([]proto.SequenceStep{{Type: StepATX, Button: ButtonReset}})
	if err != nil {
		t.Fatalf("compileSteps: %v", err)
	}
	if steps[0].kind != stepATX || steps[0].button != ButtonReset || steps[0].wait != defaultATXPress {
		t.Fatalf("step = %+v", steps[0])
	}
}

func TestAutoTapProgram(t *testing.T) {
	p, err := autoTapProgram(proto.AutoTapReq{Label: "F2", Keys: []int{0x3b}})
	if err != nil {
		t.Fatalf("autoTapProgram: %v", err)
	}
	if p.kind != KindAutoTap || p.name != "F2" || p.timeout != defaultTapTimeout || len(p.steps) != 0 {
		t.Fatalf("program = %+v", p)
	}
	var total time.Duration
	for _, s := range p.loop {
		total += s.wait
	}
	if total != defaultTapInterval {
		t.Fatalf("one tap takes %s, want %s", total, defaultTapInterval)
	}
	if p.loop[1].wait != tapHold {
		t.Fatalf("hold = %s, want %s", p.loop[1].wait, tapHold)
	}

	// The hold never takes more than half the interval.
	p, _ = autoTapProgram(proto.AutoTapReq{Keys: []int{0x3b}, IntervalMs: 50})
	if p.loop[1].wait != 25*time.Millisecond {
		t.Fatalf("hold at 50 ms = %s, want 25ms", p.loop[1].wait)
	}

	p, err = autoTapProgram(proto.AutoTapReq{Keys: []int{0x4c}, ATX: ButtonPower, TimeoutS: 120})
	if err != nil {
		t.Fatalf("autoTapProgram with ATX: %v", err)
	}
	if len(p.steps) != 1 || p.steps[0].kind != stepATX || p.timeout != 2*time.Minute {
		t.Fatalf("program = %+v", p)
	}
	if p.name != "0x4c" {
		t.Fatalf("name without a label = %q", p.name)
	}

	for _, req := range []proto.AutoTapReq{
		{},
		{Keys: []int{0x3b}, IntervalMs: 49},
		{Keys: []int{0x3b}, IntervalMs: 5001},
		{Keys: []int{0x3b}, TimeoutS: 601},
		{Keys: []int{0x3b}, TimeoutS: -1},
		{Keys: []int{0x3b}, ATX: "sleep"},
	} {
		if _, err := autoTapProgram(req); err == nil {
			t.Errorf("request %+v was accepted", req)
		}
	}
}

func useMacroFile(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "macros.json")
	old := macroFile
	macroFile = path
	t.Cleanup(func() { macroFile = old })
	return path
}

func TestMacroStore(t *testing.T) {
	path := useMacroFile(t)

	if macros, err := listMacros(); err != nil || len(macros) != 0 {
		t.Fatalf("a missing file should read as no macros: %v, %v", macros, err)
	}

	steps := []proto.SequenceStep{{Type: StepKey, Keys: []int{0x28}}, {Type: StepKey}}
	saved, err := putMacro(proto.Macro{Name: "  Enter  ", Steps: steps})
	if err != nil {
		t.Fatalf("putMacro: %v", err)
	}
	if saved.ID == "" || saved.Name != "Enter" {
		t.Fatalf("saved = %+v", saved)
	}

	saved.Name = "Enter twice"
	saved.Steps = append(steps, steps...)
	if _, err := putMacro(saved); err != nil {
		t.Fatalf("replace: %v", err)
	}
	got, err := getMacro(saved.ID)
	if err != nil || got.Name != "Enter twice" || len(got.Steps) != 4 {
		t.Fatalf("getMacro = %+v, %v", got, err)
	}

	if _, err := putMacro(proto.Macro{ID: "missing", Name: "x", Steps: steps}); !errors.Is(err, errMacroNotFound) {
		t.Fatalf("replacing a missing macro = %v", err)
	}
	if _, err := putMacro(proto.Macro{Name: "", Steps: steps}); err == nil {
		t.Fatal("a macro without a name was saved")
	}
	if _, err := putMacro(proto.Macro{Name: "bad", Steps: []proto.SequenceStep{{Type: "x"}}}); err == nil {
		t.Fatal("a macro that cannot play was saved")
	}

	if err := deleteMacro(saved.ID); err != nil {
		t.Fatalf("deleteMacro: %v", err)
	}
	if err := deleteMacro(saved.ID); !errors.Is(err, errMacroNotFound) {
		t.Fatalf("second delete = %v", err)
	}

	// Nothing but the file itself is left in the directory: the temporary
	// file of each write was renamed or removed.
	entries, _ := os.ReadDir(filepath.Dir(path))
	if len(entries) != 1 {
		t.Fatalf("directory holds %d entries, want only the macro file", len(entries))
	}
}

func TestMacroStoreLimit(t *testing.T) {
	useMacroFile(t)
	steps := []proto.SequenceStep{{Type: StepDelay, Ms: 1}}
	for i := 0; i < maxMacros; i++ {
		if _, err := putMacro(proto.Macro{Name: "m", Steps: steps}); err != nil {
			t.Fatalf("macro %d: %v", i, err)
		}
	}
	if _, err := putMacro(proto.Macro{Name: "one more", Steps: steps}); !errors.Is(err, errTooManyMacros) {
		t.Fatalf("macro over the limit = %v", err)
	}
}

func TestTruncateCountsRunes(t *testing.T) {
	if got := truncate("ääää", 2); got != "ää" {
		t.Fatalf("truncate = %q", got)
	}
}
