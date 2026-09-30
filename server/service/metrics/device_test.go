package metrics

import (
	"context"
	"strings"
	"testing"
	"time"

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

// fakeVpn is a provider whose CLI calls are counted.
type fakeVpn struct {
	installed, running, connected bool
	calls                         int
}

func (f *fakeVpn) provider(name string) vpnProvider {
	return vpnProvider{
		name:      name,
		installed: func() bool { return f.installed },
		running:   func() bool { return f.running },
		connected: func(context.Context) bool { f.calls++; return f.connected },
	}
}

func stubVpn(t *testing.T, providers ...vpnProvider) *time.Time {
	t.Helper()
	clock := time.Unix(1_000_000, 0)
	setVar(t, &vpnProviders, providers)
	setVar(t, &vpnAnswers, map[string]vpnAnswer{})
	setVar(t, &now, func() time.Time { return clock })
	return &clock
}

func TestVpnOnlyInstalledProvidersAreWritten(t *testing.T) {
	ts := &fakeVpn{installed: true, running: true, connected: true}
	nb := &fakeVpn{installed: false}
	stubVpn(t, ts.provider("tailscale"), nb.provider("netbird"))

	want := `# HELP ironkvm_vpn_connected Whether an installed VPN is connected. Asked of its CLI at most once a minute.
# TYPE ironkvm_vpn_connected gauge
ironkvm_vpn_connected{provider="tailscale"} 1
`
	assertText(t, render(t, collectVpn), want)
}

func TestVpnNotRunningIsNotConnectedWithoutAskingTheCli(t *testing.T) {
	ts := &fakeVpn{installed: true, running: false, connected: true}
	nb := &fakeVpn{installed: true, running: true, connected: false}
	stubVpn(t, ts.provider("tailscale"), nb.provider("netbird"))

	want := `# HELP ironkvm_vpn_connected Whether an installed VPN is connected. Asked of its CLI at most once a minute.
# TYPE ironkvm_vpn_connected gauge
ironkvm_vpn_connected{provider="tailscale"} 0
ironkvm_vpn_connected{provider="netbird"} 0
`
	assertText(t, render(t, collectVpn), want)
	if ts.calls != 0 {
		t.Fatalf("the CLI of a stopped daemon was run %d times", ts.calls)
	}
}

func TestVpnAnswerIsKeptForAMinute(t *testing.T) {
	ts := &fakeVpn{installed: true, running: true, connected: true}
	clock := stubVpn(t, ts.provider("tailscale"))

	render(t, collectVpn)
	*clock = clock.Add(59 * time.Second)
	ts.connected = false
	if got := render(t, collectVpn); !contains(got, `ironkvm_vpn_connected{provider="tailscale"} 1`) {
		t.Fatalf("the kept answer was not used:\n%s", got)
	}
	if ts.calls != 1 {
		t.Fatalf("the CLI ran %d times within a minute, want 1", ts.calls)
	}

	*clock = clock.Add(time.Second)
	if got := render(t, collectVpn); !contains(got, `ironkvm_vpn_connected{provider="tailscale"} 0`) {
		t.Fatalf("a stale answer was served:\n%s", got)
	}
	if ts.calls != 2 {
		t.Fatalf("the CLI ran %d times, want 2 after a minute", ts.calls)
	}
}

func TestVpnDaemonThatStartsIsAskedAtOnce(t *testing.T) {
	ts := &fakeVpn{installed: true, running: true, connected: false}
	stubVpn(t, ts.provider("tailscale"))

	render(t, collectVpn)
	ts.running = false
	render(t, collectVpn)
	ts.running, ts.connected = true, true
	if got := render(t, collectVpn); !contains(got, `ironkvm_vpn_connected{provider="tailscale"} 1`) {
		t.Fatalf("a restarted daemon was served the old answer:\n%s", got)
	}
}

func contains(text, line string) bool {
	return strings.Contains(text, line+"\n")
}
