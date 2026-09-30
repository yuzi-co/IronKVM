package metrics

import (
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/vm"
	"NanoKVM-Server/service/vm/jiggler"
)

// What the page shows about the device, as gauges. The readers are variables
// so tests can stub them.
var (
	hdmiSignal = vm.HdmiSignal
	audioState = audio.Shared.State
	jigglers   = func() (mouse, key bool) {
		j := jiggler.GetJiggler()
		key, _ = j.KeyJiggler()
		return j.IsEnabled(), key
	}
)

const (
	helpHdmiSignal = "Whether the HDMI input has a signal, as the page shows it. 0 while capture is off."
	helpAudioState = "What audio capture is doing, one-hot. All 0 while no viewer listens."
	helpJiggler    = "Whether the mouse or key jiggler is on."
)

// audioStates are the states the gauge set names, in the order it writes them.
var audioStates = []audio.State{audio.StatePlaying, audio.StateIdle, audio.StateFailing}

func collectHdmi(w *Writer) {
	w.Gauge("ironkvm_hdmi_signal", helpHdmiSignal, boolValue(hdmiSignal()))
}

func collectAudio(w *Writer) {
	current := audioState()
	for _, state := range audioStates {
		w.Gauge("ironkvm_audio_state", helpAudioState, boolValue(current == state), L("state", state.String()))
	}
}

func collectJiggler(w *Writer) {
	mouse, key := jigglers()
	w.Gauge("ironkvm_jiggler_enabled", helpJiggler, boolValue(mouse), L("kind", "mouse"))
	w.Gauge("ironkvm_jiggler_enabled", helpJiggler, boolValue(key), L("kind", "key"))
}

func boolValue(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
