package settingsio

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/dashcam"
	"NanoKVM-Server/service/hid"
	"NanoKVM-Server/service/notify"
	"NanoKVM-Server/service/roommic"
	"NanoKVM-Server/service/sequence"
	"NanoKVM-Server/service/vm/jiggler"
	"NanoKVM-Server/utils"
)

// The files in /etc/kvm the sections read and write. Each one belongs to the
// service named beside it, which reads it at start or on every request.
const (
	webTitleFile        = "/etc/kvm/web-title"       // service/vm/web_title.go
	leaderKeyFile       = "/etc/kvm/leader-key"      // service/hid/leader_key.go
	keyboardLayoutFile  = "/etc/kvm/keyboard-layout" // service/hid/host_layout.go
	shortcutsFile       = "/etc/kvm/shortcuts.json"  // service/hid/shortcut.go
	macrosFile          = "/etc/kvm/macros.json"     // service/sequence/macro.go
	mouseJigglerFile    = "/etc/kvm/mouse-jiggler"   // service/vm/jiggler/jiggler.go
	keyJigglerFile      = "/etc/kvm/key-jiggler"     // service/vm/jiggler/keys.go
	oledSleepFile       = "/etc/kvm/oled_sleep"      // service/vm/oled.go
	oledBrightnessFile  = "/etc/kvm/oled_contrast"   // service/vm/oled.go
	hdmiIdleTimeoutFile = utils.HDMIIdleTimeoutFile  // utils/hdmi.go
	roomMicFile         = "/etc/kvm/room-mic"        // service/roommic/settings.go
	notifyFile          = notify.File                // service/notify/settings.go
	dashcamFile         = dashcam.File               // service/dashcam/settings.go

	// vncPasswordFile is only checked for existence: plain VNC
	// authentication can be imported only where a password is set.
	vncPasswordFile = "/etc/kvm/vnc.passwd" // service/vnc/password.go

	// srtPassphraseFile is only checked for existence: SRT can be imported
	// switched on only where a passphrase is set.
	srtPassphraseFile = "/etc/kvm/srt.passphrase" // service/srt/passphrase.go
)

// The limits that mirror the service that owns the file.
const (
	// service/vm/web_title.go removes the file for the default title.
	defaultWebTitle  = "IronKVM"
	maxWebTitleRunes = 128

	maxShortcuts    = 100
	maxShortcutKeys = 10
	maxShortcutText = 64

	// The mouse jiggler's modes, service/vm/jiggler/jiggler.go.
	mouseModeRelative = "relative"
	mouseModeAbsolute = "absolute"

	// minSleepSeconds, maxSleepSeconds, minBrightness, maxBrightness and
	// defaultBrightness in service/vm/oled.go, which mirror kvm_system.
	oledMinSleepSeconds = 10
	oledMaxSleepSeconds = 65535
	oledMinBrightness   = 0x10
	oledMaxBrightness   = 0xFF
	oledDefaultBright   = 0xCF

	settingsFileMode = 0o644
	// notify.json holds the channel tokens.
	secretsFileMode = 0o600
)

// trimmedText reads a one-line file, "" when it is absent.
func trimmedText(s *snapshot, path string) (string, error) {
	data, ok, err := s.readFile(path)
	if err != nil || !ok {
		return "", err
	}
	return strings.TrimSpace(strings.ReplaceAll(string(data), "\n", "")), nil
}

type webTitleSettings struct {
	Title string `json:"title"`
}

var webTitleSection = spec[webTitleSettings]{
	id: "webTitle",
	get: func(s *snapshot) (webTitleSettings, error) {
		title, err := trimmedText(s, webTitleFile)
		return webTitleSettings{Title: title}, err
	},
	check: func(_ *snapshot, v *webTitleSettings) error {
		v.Title = strings.TrimSpace(v.Title)
		if utf8.RuneCountInString(v.Title) > maxWebTitleRunes {
			return fmt.Errorf("the title is at most %d characters", maxWebTitleRunes)
		}
		for _, r := range v.Title {
			if unicode.IsControl(r) {
				return errors.New("the title holds a control character")
			}
		}
		return nil
	},
	put: func(p *plan, v webTitleSettings) error {
		if v.Title == "" || v.Title == defaultWebTitle {
			p.remove(webTitleFile)
		} else {
			p.write(webTitleFile, []byte(v.Title), settingsFileMode)
		}
		return nil
	},
}

type leaderKeySettings struct {
	Key string `json:"key"`
}

// leaderKeyPattern is a KeyboardEvent.code, such as ControlRight.
var leaderKeyPattern = regexp.MustCompile(`^[A-Za-z0-9]{0,32}$`)

var leaderKeySection = spec[leaderKeySettings]{
	id: "leaderKey",
	get: func(s *snapshot) (leaderKeySettings, error) {
		key, err := trimmedText(s, leaderKeyFile)
		return leaderKeySettings{Key: key}, err
	},
	check: func(_ *snapshot, v *leaderKeySettings) error {
		if !leaderKeyPattern.MatchString(v.Key) {
			return errors.New("the leader key is a key code of letters and digits")
		}
		return nil
	},
	put: func(p *plan, v leaderKeySettings) error {
		if v.Key == "" {
			p.remove(leaderKeyFile)
		} else {
			p.write(leaderKeyFile, []byte(v.Key), settingsFileMode)
		}
		return nil
	},
}

type keyboardLayoutSettings struct {
	Layout string `json:"layout"`
}

// keyboardLayoutSection is the keyboard layout active on the host. An empty
// layout is the default, US.
var keyboardLayoutSection = spec[keyboardLayoutSettings]{
	id: "keyboardLayout",
	get: func(s *snapshot) (keyboardLayoutSettings, error) {
		layout, err := trimmedText(s, keyboardLayoutFile)
		return keyboardLayoutSettings{Layout: layout}, err
	},
	check: func(_ *snapshot, v *keyboardLayoutSettings) error {
		v.Layout = strings.TrimSpace(v.Layout)
		if v.Layout == "" {
			return nil
		}
		layout, err := hid.GetLayout(v.Layout)
		if err != nil {
			return fmt.Errorf("unknown keyboard layout %q", v.Layout)
		}
		v.Layout = layout.ID
		return nil
	},
	put: func(p *plan, v keyboardLayoutSettings) error {
		if v.Layout == "" {
			p.remove(keyboardLayoutFile)
		} else {
			p.write(keyboardLayoutFile, []byte(v.Layout), settingsFileMode)
		}
		return nil
	},
}

type shortcutsSettings struct {
	Shortcuts []proto.Shortcut `json:"shortcuts"`
}

var shortcutsSection = spec[shortcutsSettings]{
	id: "shortcuts",
	get: func(s *snapshot) (shortcutsSettings, error) {
		v := shortcutsSettings{Shortcuts: []proto.Shortcut{}}
		data, ok, err := s.readFile(shortcutsFile)
		if err != nil || !ok || len(data) == 0 {
			return v, err
		}
		if err := json.Unmarshal(data, &v); err != nil {
			return shortcutsSettings{Shortcuts: []proto.Shortcut{}}, errors.New("shortcuts.json is not valid JSON")
		}
		if v.Shortcuts == nil {
			v.Shortcuts = []proto.Shortcut{}
		}
		return v, nil
	},
	check: func(_ *snapshot, v *shortcutsSettings) error {
		if v.Shortcuts == nil {
			v.Shortcuts = []proto.Shortcut{}
		}
		if len(v.Shortcuts) > maxShortcuts {
			return fmt.Errorf("at most %d shortcuts", maxShortcuts)
		}
		for _, sc := range v.Shortcuts {
			if sc.ID == "" || len(sc.ID) > maxShortcutText {
				return errors.New("a shortcut has no valid ID")
			}
			if len(sc.Keys) == 0 || len(sc.Keys) > maxShortcutKeys {
				return fmt.Errorf("a shortcut has 1 to %d keys", maxShortcutKeys)
			}
			for _, key := range sc.Keys {
				if key.Code == "" || len(key.Code) > maxShortcutText || len(key.Label) > maxShortcutText {
					return errors.New("a shortcut key has no valid code or label")
				}
			}
		}
		return nil
	},
	put: func(p *plan, v shortcutsSettings) error {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		p.write(shortcutsFile, data, settingsFileMode)
		return nil
	},
}

type macrosSettings struct {
	Macros []proto.Macro `json:"macros"`
}

var macrosSection = spec[macrosSettings]{
	id: "macros",
	get: func(s *snapshot) (macrosSettings, error) {
		v := macrosSettings{Macros: []proto.Macro{}}
		data, ok, err := s.readFile(macrosFile)
		if err != nil || !ok || len(data) == 0 {
			return v, err
		}
		if err := json.Unmarshal(data, &v); err != nil {
			return macrosSettings{Macros: []proto.Macro{}}, errors.New("macros.json is not valid JSON")
		}
		if v.Macros == nil {
			v.Macros = []proto.Macro{}
		}
		return v, nil
	},
	check: func(_ *snapshot, v *macrosSettings) error {
		if v.Macros == nil {
			v.Macros = []proto.Macro{}
		}
		return sequence.ValidateMacros(v.Macros)
	},
	put: func(p *plan, v macrosSettings) error {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		p.write(macrosFile, data, settingsFileMode)
		return nil
	},
}

type mouseJigglerSettings struct {
	Enabled bool   `json:"enabled"`
	Mode    string `json:"mode"`
}

var mouseJigglerSection = spec[mouseJigglerSettings]{
	id: "mouseJiggler",
	// The file exists while the jiggler is on and holds the mode.
	get: func(s *snapshot) (mouseJigglerSettings, error) {
		data, ok, err := s.readFile(mouseJigglerFile)
		if err != nil || !ok {
			return mouseJigglerSettings{Mode: mouseModeRelative}, err
		}
		mode := strings.TrimSpace(string(data))
		if mode == "" {
			mode = mouseModeRelative
		}
		return mouseJigglerSettings{Enabled: true, Mode: mode}, nil
	},
	check: func(_ *snapshot, v *mouseJigglerSettings) error {
		if v.Mode == "" {
			v.Mode = mouseModeRelative
		}
		if v.Mode != mouseModeRelative && v.Mode != mouseModeAbsolute {
			return errors.New("the mode is relative or absolute")
		}
		return nil
	},
	put: func(p *plan, v mouseJigglerSettings) error {
		if v.Enabled {
			p.write(mouseJigglerFile, []byte(v.Mode), settingsFileMode)
		} else {
			p.remove(mouseJigglerFile)
		}
		return nil
	},
}

type keyJigglerSettings struct {
	Enabled bool   `json:"enabled"`
	Key     string `json:"key"`
}

var keyJigglerSection = spec[keyJigglerSettings]{
	id: "keyJiggler",
	// "on" or "off" on the first line, the key on the second.
	get: func(s *snapshot) (keyJigglerSettings, error) {
		v := keyJigglerSettings{Key: jiggler.DefaultKey}
		data, ok, err := s.readFile(keyJigglerFile)
		if err != nil || !ok {
			return v, err
		}
		lines := strings.Split(strings.ReplaceAll(string(data), "\r", ""), "\n")
		v.Enabled = strings.TrimSpace(lines[0]) == "on"
		if len(lines) > 1 && jiggler.ValidKey(strings.TrimSpace(lines[1])) {
			v.Key = strings.TrimSpace(lines[1])
		}
		return v, nil
	},
	check: func(_ *snapshot, v *keyJigglerSettings) error {
		if !jiggler.ValidKey(v.Key) {
			return fmt.Errorf("unknown jiggler key %q", v.Key)
		}
		return nil
	},
	put: func(p *plan, v keyJigglerSettings) error {
		state := "off"
		if v.Enabled {
			state = "on"
		}
		p.write(keyJigglerFile, []byte(state+"\n"+v.Key+"\n"), settingsFileMode)
		return nil
	},
}

type oledSettings struct {
	SleepSeconds int `json:"sleepSeconds"`
	Brightness   int `json:"brightness"`
}

// readOLEDInt reads a number the way kvm_system does, with strtol base 0, so
// "0xFF" is a value.
func readOLEDInt(s *snapshot, path string, fallback int) (int, error) {
	text, err := trimmedText(s, path)
	if err != nil || text == "" {
		return fallback, err
	}
	value, err := strconv.ParseInt(text, 0, 0)
	if err != nil {
		return fallback, fmt.Errorf("%s does not hold a number", path)
	}
	return int(value), nil
}

var oledSection = spec[oledSettings]{
	id: "oled",
	get: func(s *snapshot) (oledSettings, error) {
		sleep, err := readOLEDInt(s, oledSleepFile, 0)
		if err != nil {
			return oledSettings{Brightness: oledDefaultBright}, err
		}
		brightness, err := readOLEDInt(s, oledBrightnessFile, oledDefaultBright)
		brightness = min(max(brightness, oledMinBrightness), oledMaxBrightness)
		return oledSettings{SleepSeconds: sleep, Brightness: brightness}, err
	},
	// The OLED page clamps rather than refuses; an import is a file somebody
	// may have edited, so it refuses what the page would have changed.
	check: func(_ *snapshot, v *oledSettings) error {
		if v.SleepSeconds != 0 && (v.SleepSeconds < oledMinSleepSeconds || v.SleepSeconds > oledMaxSleepSeconds) {
			return fmt.Errorf("the sleep delay is 0 (never) or %d to %d seconds", oledMinSleepSeconds, oledMaxSleepSeconds)
		}
		if v.Brightness < oledMinBrightness || v.Brightness > oledMaxBrightness {
			return fmt.Errorf("the brightness is %d to %d", oledMinBrightness, oledMaxBrightness)
		}
		return nil
	},
	put: func(p *plan, v oledSettings) error {
		p.write(oledSleepFile, []byte(strconv.Itoa(v.SleepSeconds)), settingsFileMode)
		p.write(oledBrightnessFile, []byte(strconv.Itoa(v.Brightness)), settingsFileMode)
		return nil
	},
}

type hdmiSettings struct {
	IdleTimeoutMinutes int `json:"idleTimeoutMinutes"`
}

var hdmiSection = spec[hdmiSettings]{
	id: "hdmi",
	get: func(s *snapshot) (hdmiSettings, error) {
		text, err := trimmedText(s, hdmiIdleTimeoutFile)
		if err != nil || text == "" {
			return hdmiSettings{IdleTimeoutMinutes: utils.DefaultHDMIIdleTimeout}, err
		}
		minutes, err := strconv.Atoi(text)
		if err != nil || minutes < 0 || minutes > utils.MaxHDMIIdleTimeoutMinutes {
			// utils.GetHDMIIdleTimeout reads such a file as the default.
			return hdmiSettings{IdleTimeoutMinutes: utils.DefaultHDMIIdleTimeout}, nil
		}
		return hdmiSettings{IdleTimeoutMinutes: minutes}, nil
	},
	check: func(_ *snapshot, v *hdmiSettings) error {
		if v.IdleTimeoutMinutes < 0 || v.IdleTimeoutMinutes > utils.MaxHDMIIdleTimeoutMinutes {
			return fmt.Errorf("the idle timeout is 0 to %d minutes", utils.MaxHDMIIdleTimeoutMinutes)
		}
		return nil
	},
	put: func(p *plan, v hdmiSettings) error {
		p.write(hdmiIdleTimeoutFile, []byte(strconv.Itoa(v.IdleTimeoutMinutes)), settingsFileMode)
		return nil
	},
}

type roomMicSettings struct {
	Allowed bool `json:"allowed"`
	Gain    int  `json:"gain"`
}

var roomMicSection = spec[roomMicSettings]{
	id: "roomMic",
	get: func(s *snapshot) (roomMicSettings, error) {
		def := roommic.DefaultSettings()
		v := roomMicSettings{Allowed: def.Allowed, Gain: def.Gain}
		data, ok, err := s.readFile(roomMicFile)
		if err != nil || !ok {
			return v, err
		}
		if err := json.Unmarshal(data, &v); err != nil {
			return roomMicSettings{Allowed: def.Allowed, Gain: def.Gain}, errors.New("room-mic is not valid JSON")
		}
		return v, nil
	},
	check: func(_ *snapshot, v *roomMicSettings) error {
		return roommic.Settings{Allowed: v.Allowed, Gain: v.Gain}.Validate()
	},
	put: func(p *plan, v roomMicSettings) error {
		data, err := json.Marshal(roommic.Settings{Allowed: v.Allowed, Gain: v.Gain})
		if err != nil {
			return err
		}
		p.write(roomMicFile, data, settingsFileMode)
		return nil
	},
}

// notificationSettings are the notification rules. The channels (webhook,
// ntfy, Telegram) are left out: their addresses and tokens are secrets, and
// an import keeps the channels the board already has.
type notificationSettings struct {
	Enabled            bool                   `json:"enabled"`
	Snapshot           bool                   `json:"snapshot"`
	Events             map[string]notify.Rule `json:"events"`
	LoginThreshold     int                    `json:"loginThreshold"`
	LoginWindowSeconds int                    `json:"loginWindowSeconds"`
}

// readNotify reads notify.json as the board keeps it, channels included. It
// never leaves this package except through notificationSettings.
func readNotify(s *snapshot) (notify.Settings, error) {
	data, ok, err := s.readFile(notifyFile)
	if err != nil || !ok {
		return notify.Settings{}, err
	}
	var full notify.Settings
	if err := json.Unmarshal(data, &full); err != nil {
		return notify.Settings{}, errors.New("notify.json is not valid JSON")
	}
	return full, nil
}

var notificationsSection = spec[notificationSettings]{
	id: "notifications",
	get: func(s *snapshot) (notificationSettings, error) {
		full, err := readNotify(s)
		full = full.WithDefaults()
		return notificationSettings{
			Enabled:            full.Enabled,
			Snapshot:           full.Snapshot,
			Events:             full.Events,
			LoginThreshold:     full.LoginThreshold,
			LoginWindowSeconds: full.LoginWindowSeconds,
		}, err
	},
	// The limits of the notifications page, on the rules alone.
	check: func(_ *snapshot, v *notificationSettings) error {
		rules := notify.Settings{
			Enabled:            v.Enabled,
			Snapshot:           v.Snapshot,
			Events:             v.Events,
			LoginThreshold:     v.LoginThreshold,
			LoginWindowSeconds: v.LoginWindowSeconds,
		}
		// Validate checks the channels too, and the default ntfy server is
		// the only channel value it needs. WithDefaults is not called first,
		// because it drops the event types it does not know, and those must
		// be refused, not lost.
		rules.Ntfy.Server = notify.DefaultNtfyServer
		if err := rules.Validate(); err != nil {
			return err
		}
		v.Events = rules.WithDefaults().Events
		return nil
	},
	put: func(p *plan, v notificationSettings) error {
		full, err := readNotify(p.snap)
		if err != nil {
			// A damaged file loses its channels either way; the rules are
			// still worth writing.
			full = notify.Settings{}
		}
		full.Enabled = v.Enabled
		full.Snapshot = v.Snapshot
		full.Events = v.Events
		full.LoginThreshold = v.LoginThreshold
		full.LoginWindowSeconds = v.LoginWindowSeconds
		data, err := json.Marshal(full.WithDefaults())
		if err != nil {
			return err
		}
		p.write(notifyFile, data, secretsFileMode)
		return nil
	},
}

// dashcamSection is the dashcam's settings. The clips stay on the board they
// were saved on.
var dashcamSection = spec[dashcam.Settings]{
	id: "dashcam",
	get: func(s *snapshot) (dashcam.Settings, error) {
		v := dashcam.Defaults()
		data, ok, err := s.readFile(dashcamFile)
		if err != nil || !ok {
			return v, err
		}
		if err := json.Unmarshal(data, &v); err != nil {
			return dashcam.Defaults(), errors.New("dashcam.json is not valid JSON")
		}
		return v, nil
	},
	check: func(_ *snapshot, v *dashcam.Settings) error {
		return v.Validate()
	},
	put: func(p *plan, v dashcam.Settings) error {
		data, err := json.Marshal(v)
		if err != nil {
			return err
		}
		p.write(dashcamFile, data, settingsFileMode)
		return nil
	},
}
