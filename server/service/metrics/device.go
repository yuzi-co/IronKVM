package metrics

import (
	"context"
	"sync"
	"time"

	"NanoKVM-Server/service/extensions/netbird"
	"NanoKVM-Server/service/extensions/tailscale"
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
	vpnProviders = []vpnProvider{
		{name: "tailscale", installed: tailscale.Installed, running: tailscale.Running, connected: tailscale.Connected},
		{name: "netbird", installed: netbird.Installed, running: netbird.Running, connected: netbird.Connected},
	}
	now = time.Now
)

const (
	helpHdmiSignal = "Whether the HDMI input has a signal, as the page shows it. 0 while capture is off."
	helpAudioState = "What audio capture is doing, one-hot. All 0 while no viewer listens."
	helpJiggler    = "Whether the mouse or key jiggler is on."
	helpVpn        = "Whether an installed VPN is connected. Asked of its CLI at most once a minute."
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

// vpnProvider is one VPN add-on. installed and running are a stat and a pid
// file, cheap enough for every scrape. connected runs the CLI, so its answer
// is kept for vpnCacheTTL.
type vpnProvider struct {
	name      string
	installed func() bool
	running   func() bool
	connected func(context.Context) bool
}

const (
	vpnCacheTTL = time.Minute
	// vpnCheckTimeout bounds one CLI call, well inside a scrape's timeout. A
	// CLI that does not answer in time counts as not connected.
	vpnCheckTimeout = 5 * time.Second
)

type vpnAnswer struct {
	at        time.Time
	connected bool
}

var (
	// vpnMutex also keeps two scrapes from running the same CLI at once.
	vpnMutex   sync.Mutex
	vpnAnswers = map[string]vpnAnswer{}
)

// collectVpn writes a series for each installed provider only. A daemon that
// is not running is not connected, which needs no CLI; it also forgets the
// kept answer, so a daemon that starts is asked at once.
func collectVpn(w *Writer) {
	vpnMutex.Lock()
	defer vpnMutex.Unlock()

	for _, p := range vpnProviders {
		if !p.installed() {
			delete(vpnAnswers, p.name)
			continue
		}

		connected := false
		if p.running() {
			answer, ok := vpnAnswers[p.name]
			if !ok || now().Sub(answer.at) >= vpnCacheTTL {
				ctx, cancel := context.WithTimeout(context.Background(), vpnCheckTimeout)
				answer = vpnAnswer{at: now(), connected: p.connected(ctx)}
				cancel()
				vpnAnswers[p.name] = answer
			}
			connected = answer.connected
		} else {
			delete(vpnAnswers, p.name)
		}

		w.Gauge("ironkvm_vpn_connected", helpVpn, boolValue(connected), L("provider", p.name))
	}
}

func boolValue(b bool) float64 {
	if b {
		return 1
	}
	return 0
}
