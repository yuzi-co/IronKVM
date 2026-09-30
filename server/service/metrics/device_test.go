package metrics

import (
	"testing"

	"NanoKVM-Server/service/stream/audio"
)

func TestHdmiSignal(t *testing.T) {
	for signal, value := range map[bool]string{true: "1", false: "0"} {
		setVar(t, &hdmiSignal, func() bool { return signal })

		want := `# HELP ironkvm_hdmi_signal Whether the HDMI input has a signal, as the page shows it. 0 while capture is off.
# TYPE ironkvm_hdmi_signal gauge
ironkvm_hdmi_signal ` + value + "\n"
		assertText(t, render(t, collectHdmi), want)
	}
}

func TestAudioStateIsOneHot(t *testing.T) {
	setVar(t, &audioState, func() audio.State { return audio.StateIdle })

	want := `# HELP ironkvm_audio_state What audio capture is doing, one-hot. All 0 while no viewer listens.
# TYPE ironkvm_audio_state gauge
ironkvm_audio_state{state="playing"} 0
ironkvm_audio_state{state="idle"} 1
ironkvm_audio_state{state="failing"} 0
`
	assertText(t, render(t, collectAudio), want)
}

func TestAudioStateWithNoCaptureIsAllZero(t *testing.T) {
	setVar(t, &audioState, func() audio.State { return audio.StateUnknown })

	want := `# HELP ironkvm_audio_state What audio capture is doing, one-hot. All 0 while no viewer listens.
# TYPE ironkvm_audio_state gauge
ironkvm_audio_state{state="playing"} 0
ironkvm_audio_state{state="idle"} 0
ironkvm_audio_state{state="failing"} 0
`
	assertText(t, render(t, collectAudio), want)
}

func TestJigglerEnabledByKind(t *testing.T) {
	setVar(t, &jigglers, func() (bool, bool) { return false, true })

	want := `# HELP ironkvm_jiggler_enabled Whether the mouse or key jiggler is on.
# TYPE ironkvm_jiggler_enabled gauge
ironkvm_jiggler_enabled{kind="mouse"} 0
ironkvm_jiggler_enabled{kind="key"} 1
`
	assertText(t, render(t, collectJiggler), want)
}
