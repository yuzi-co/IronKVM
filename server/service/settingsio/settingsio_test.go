package settingsio

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	"NanoKVM-Server/config"
)

const sentinel = "SENTINEL-SECRET"

// board is a fake board: a root directory with /etc/kvm in it.
type board struct {
	t    *testing.T
	root string
	dev  *Device
}

func newBoard(t *testing.T) *board {
	t.Helper()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "etc", "kvm"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &board{t: t, root: root, dev: &Device{Root: root}}
}

func (b *board) write(path, content string) {
	b.t.Helper()
	full := filepath.Join(b.root, path)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		b.t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
		b.t.Fatal(err)
	}
}

func (b *board) read(path string) string {
	b.t.Helper()
	data, err := os.ReadFile(filepath.Join(b.root, path))
	if err != nil {
		b.t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func (b *board) exists(path string) bool {
	_, err := os.Stat(filepath.Join(b.root, path))
	return err == nil
}

func (b *board) writeConfig(conf *config.Config) {
	b.t.Helper()
	data, err := yaml.Marshal(conf)
	if err != nil {
		b.t.Fatal(err)
	}
	b.write(config.ConfigurationFile, string(data))
}

func (b *board) config() *config.Config {
	b.t.Helper()
	var conf config.Config
	if err := yaml.Unmarshal([]byte(b.read(config.ConfigurationFile)), &conf); err != nil {
		b.t.Fatal(err)
	}
	return &conf
}

func (b *board) export() []byte {
	b.t.Helper()
	doc, skipped, err := b.dev.Export(Source{Application: "2.4.3", Hostname: "kvm"}, time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC))
	if err != nil {
		b.t.Fatal(err)
	}
	if len(skipped) > 0 {
		b.t.Fatalf("sections skipped: %v", skipped)
	}
	data, err := json.Marshal(doc)
	if err != nil {
		b.t.Fatal(err)
	}
	return data
}

func baseConfig() *config.Config {
	enabled := true
	return &config.Config{
		Proto:          "http",
		Port:           config.Port{Http: 80, Https: 443},
		Logger:         config.Logger{Level: "info", File: "stdout"},
		Authentication: "enable",
		JWT:            config.JWT{RefreshTokenDuration: 2678400, RevokeTokensOnLogout: true},
		Stun:           "stun.l.google.com:19302",
		Security:       config.Security{LoginMaxFailures: 5},
		Redfish:        config.Redfish{Enabled: &enabled},
	}
}

// seedSecrets puts a sentinel in every secret this repository keeps on the
// board: in server.yaml and in the files beside it.
func seedSecrets(b *board) {
	conf := baseConfig()
	conf.JWT.SecretKey = sentinel + "-jwt"
	conf.Turn = config.Turn{TurnAddr: "turn:" + sentinel + "-turnaddr", TurnUser: sentinel + "-turnuser", TurnCred: sentinel + "-turncred"}
	conf.Proxy = "http://user:" + sentinel + "-proxy@proxy.lan:3128"
	conf.Cert = config.Cert{Crt: "/etc/kvm/" + sentinel + "-crt", Key: "/etc/kvm/" + sentinel + "-key"}
	conf.Security.AllowedNetworks = []string{"10.99.0.0/16"}
	conf.Host = sentinel + "-host"
	b.writeConfig(conf)

	for _, path := range []string{
		"/etc/kvm/pwd", "/etc/kvm/.jwt_secret", "/etc/kvm/totp.key", "/etc/kvm/api_keys.json",
		"/etc/kvm/share_links.json", "/etc/kvm/login_sessions.json", "/etc/kvm/mcp.json", "/etc/kvm/.picoclaw_internal_token",
		"/etc/kvm/first-boot-password", "/etc/kvm/server.key", "/etc/kvm/uploaded.key",
		"/etc/kvm/server.crt", "/etc/kvm/uploaded.crt", "/etc/kvm/wifi.ssid", "/etc/kvm/wifi.pass",
		"/etc/kvm/vnc.passwd", "/etc/kvm/nut.passwd", "/etc/kvm/ipmi.key",
		"/etc/kvm/application-update.json", "/etc/kvm/syslog-remote", "/etc/kvm/ssh_port",
		"/etc/kvm/redfish-uuid", "/etc/kvm/usb-identity", "/etc/kvm/agent-tiers.json",
		"/etc/kvm/oidc.json",
		"/boot/ironkvm-first-password.txt", "/boot/eth.nodhcp",
	} {
		b.write(path, sentinel+"-"+filepath.Base(path))
	}

	b.write("/etc/kvm/notify.json", fmt.Sprintf(`{
  "enabled": true,
  "webhook": {"enabled": true, "url": "https://hooks.example/%[1]s-webhook"},
  "ntfy": {"enabled": true, "server": "https://%[1]s-ntfyserver.example", "topic": "%[1]s-topic", "token": "%[1]s-ntfytoken"},
  "telegram": {"enabled": true, "botToken": "%[1]s-bot", "chatId": "%[1]s-chat"}
}`, sentinel))
}

func TestExportCarriesNoSecret(t *testing.T) {
	b := newBoard(t)
	seedSecrets(b)

	out := string(b.export())

	if strings.Contains(out, sentinel) {
		i := strings.Index(out, sentinel)
		t.Fatalf("the export holds a secret near %q", out[max(0, i-80):min(len(out), i+40)])
	}
	if strings.Contains(out, "10.99.0.0") {
		t.Fatal("the export holds the allowed networks")
	}
}

// fillStrings sets every string in v, however deep, to a value that names its
// path, and every list of strings to one such value.
func fillStrings(v reflect.Value, path string) {
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		fillStrings(v.Elem(), path)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			f := v.Type().Field(i)
			if !f.IsExported() {
				continue
			}
			fillStrings(v.Field(i), path+"."+f.Name)
		}
	case reflect.String:
		v.SetString("FIELD" + path + "END")
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.String {
			v.Set(reflect.ValueOf([]string{"FIELD" + path + "END"}))
		}
	}
}

// TestExportIsAnAllowlist puts a distinct value in every string field of
// server.yaml, the ones that exist today and any added later, and requires
// that only the fields the allowlist names reach the file. A new field,
// secret or not, fails this test until somebody decides about it.
func TestExportIsAnAllowlist(t *testing.T) {
	b := newBoard(t)
	conf := baseConfig()
	fillStrings(reflect.ValueOf(conf).Elem(), "")
	b.writeConfig(conf)

	doc, _, err := b.dev.Export(Source{}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(doc.Sections)

	var found []string
	rest := string(data)
	for {
		i := strings.Index(rest, "FIELD")
		if i < 0 {
			break
		}
		j := strings.Index(rest[i:], "END")
		found = append(found, rest[i+len("FIELD"):i+j])
		rest = rest[i+j:]
	}
	sort.Strings(found)

	want := []string{
		".HostLog.Bind",
		".HostLog.Patterns",
		".HostLog.Sources",
		".Logger.Level",
		".NUT.Host",
		".NUT.UPS",
		".NUT.Username",
		".Stun",
		".Watchdog.Action",
		".Watchdog.PingHost",
		".WebRTC.PublicAddress",
	}
	if !reflect.DeepEqual(found, want) {
		t.Fatalf("the export carries these server.yaml strings:\n%v\nwant\n%v", found, want)
	}
}

func TestExportSectionsAreTheAllowlist(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())

	var doc Document
	if err := json.Unmarshal(b.export(), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Format != Format || doc.Version != Version {
		t.Fatalf("format %q version %d", doc.Format, doc.Version)
	}
	var names []string
	for name := range doc.Sections {
		names = append(names, name)
	}
	sort.Strings(names)
	want := append([]string(nil), SectionNames()...)
	sort.Strings(want)
	if !reflect.DeepEqual(names, want) {
		t.Fatalf("sections %v, want %v", names, want)
	}
}

func TestNotificationsExportKeepsRulesOnly(t *testing.T) {
	b := newBoard(t)
	seedSecrets(b)

	var doc Document
	if err := json.Unmarshal(b.export(), &doc); err != nil {
		t.Fatal(err)
	}
	var rules map[string]json.RawMessage
	if err := json.Unmarshal(doc.Sections["notifications"], &rules); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"webhook", "ntfy", "telegram"} {
		if _, ok := rules[key]; ok {
			t.Errorf("the notifications section holds %s", key)
		}
	}
	if string(rules["enabled"]) != "true" {
		t.Errorf("enabled = %s", rules["enabled"])
	}
}

func parse(t *testing.T, data []byte) *Document {
	t.Helper()
	doc, err := ParseDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

// source is a board with a setting in every section that differs from a
// fresh board's.
func sourceBoard(t *testing.T) *board {
	b := newBoard(t)
	conf := baseConfig()
	conf.Logger.Level = "debug"
	conf.HardwareSettings.PowerLED = true
	conf.Watchdog = config.Watchdog{Enabled: true, TimeoutMinutes: 10, Action: "power", PingHost: "192.168.1.10"}
	conf.VNC = config.VNC{Enabled: true, Port: 5901}
	conf.HostLog.Syslog = config.HostLogListener{Enabled: true}
	b.writeConfig(conf)
	b.write(webTitleFile, "Rack 4")
	b.write(leaderKeyFile, "ControlRight")
	b.write(keyboardLayoutFile, "de")
	b.write(mouseJigglerFile, "absolute")
	b.write(keyJigglerFile, "on\nshift\n")
	b.write(oledSleepFile, "120")
	b.write(oledBrightnessFile, "0xFF")
	b.write(hdmiIdleTimeoutFile, "30")
	b.write(roomMicFile, `{"allowed":true,"gain":12}`)
	b.write(shortcutsFile, `{"shortcuts":[{"id":"a","keys":[{"code":"ControlLeft","label":"Ctrl"},{"code":"Delete","label":"Del"}]}]}`)
	b.write(macrosFile, `{"macros":[{"id":"m1","name":"Login","steps":[{"type":"key","keys":[4]},{"type":"key"}]}]}`)
	return b
}

func TestImportRoundTrip(t *testing.T) {
	src := sourceBoard(t)
	data := src.export()

	dst := newBoard(t)
	dstConf := baseConfig()
	dstConf.JWT.SecretKey = "keep-me"
	dstConf.Port = config.Port{Http: 8080, Https: 8443}
	dstConf.Security.AllowedNetworks = []string{"10.1.0.0/16"}
	dstConf.Turn = config.Turn{TurnAddr: "turn:x", TurnUser: "u", TurnCred: "c"}
	dst.writeConfig(dstConf)
	dst.write(notifyFile, `{"webhook":{"enabled":true,"url":"https://hooks.example/keep"}}`)

	applied, err := dst.dev.Import(parse(t, data), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(applied) != len(sections) {
		t.Fatalf("applied %v", applied)
	}

	conf := dst.config()
	if conf.Logger.Level != "debug" || !conf.HardwareSettings.PowerLED || conf.Watchdog.Action != "power" || conf.VNC.Port != 5901 {
		t.Fatalf("server.yaml not imported: %+v", conf)
	}
	// What the export never carries stays as the board had it.
	if conf.JWT.SecretKey != "keep-me" || conf.Port.Http != 8080 || conf.Security.AllowedNetworks[0] != "10.1.0.0/16" || conf.Turn.TurnCred != "c" {
		t.Fatalf("the import changed a setting outside the allowlist: %+v", conf)
	}

	checks := map[string]string{
		webTitleFile:        "Rack 4",
		leaderKeyFile:       "ControlRight",
		keyboardLayoutFile:  "de",
		mouseJigglerFile:    "absolute",
		keyJigglerFile:      "on\nshift\n",
		oledSleepFile:       "120",
		oledBrightnessFile:  "255",
		hdmiIdleTimeoutFile: "30",
	}
	for path, want := range checks {
		if got := dst.read(path); got != want {
			t.Errorf("%s = %q, want %q", path, got, want)
		}
	}
	if !strings.Contains(dst.read(macrosFile), `"Login"`) || !strings.Contains(dst.read(shortcutsFile), "ControlLeft") {
		t.Error("macros or shortcuts not imported")
	}
	// The notification channels the board had are kept.
	if !strings.Contains(dst.read(notifyFile), "https://hooks.example/keep") {
		t.Error("the import dropped the board's own notification channel")
	}

	// A second export of the imported board matches the first, apart from
	// what names the board.
	var a, c Document
	_ = json.Unmarshal(data, &a)
	_ = json.Unmarshal(dst.export(), &c)
	for name, raw := range a.Sections {
		if string(raw) != string(c.Sections[name]) {
			t.Errorf("section %s: exported %s, after the import %s", name, raw, c.Sections[name])
		}
	}
}

func TestPreviewListsChangesAndWritesNothing(t *testing.T) {
	src := sourceBoard(t)
	dst := newBoard(t)
	dst.writeConfig(baseConfig())
	before := dst.read(config.ConfigurationFile)

	preview := dst.dev.Preview(parse(t, src.export()))

	if dst.read(config.ConfigurationFile) != before || dst.exists(webTitleFile) {
		t.Fatal("the preview wrote a file")
	}
	byName := map[string]SectionPreview{}
	for _, s := range preview.Sections {
		byName[s.Section] = s
	}
	if s := byName["webTitle"]; s.Status != StatusChange || len(s.Changes) != 1 || s.Changes[0].Field != "title" || s.Changes[0].To != "Rack 4" {
		t.Fatalf("webTitle preview: %+v", s)
	}
	if s := byName["logging"]; s.Status != StatusChange || s.Changes[0].From != "info" || s.Changes[0].To != "debug" {
		t.Fatalf("logging preview: %+v", s)
	}
	if s := byName["redfish"]; s.Status != StatusSame {
		t.Fatalf("redfish preview: %+v", s)
	}
	if s := byName["hostLog"]; s.Status != StatusChange || s.Changes[0].Field != "syslog.enabled" {
		t.Fatalf("hostLog preview: %+v", s)
	}
}

func TestDocumentVersions(t *testing.T) {
	for _, tc := range []struct {
		name string
		doc  string
		ok   bool
		is   error
	}{
		{"current", `{"format":"ironkvm-settings","version":1,"sections":{}}`, true, nil},
		{"newer", `{"format":"ironkvm-settings","version":2,"sections":{}}`, false, ErrNewerVersion},
		{"no version", `{"format":"ironkvm-settings","sections":{}}`, false, nil},
		{"other format", `{"format":"something","version":1,"sections":{}}`, false, nil},
		{"no sections", `{"format":"ironkvm-settings","version":1}`, false, nil},
		{"not json", `format: x`, false, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseDocument([]byte(tc.doc))
			if (err == nil) != tc.ok {
				t.Fatalf("err = %v", err)
			}
			if tc.is != nil && !errors.Is(err, tc.is) {
				t.Fatalf("err = %v, want %v", err, tc.is)
			}
		})
	}
}

func TestUnknownSectionsAndFields(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{
		"future":{"x":1},
		"logging":{"level":"warn","colour":"blue"}
	}}`))

	preview := b.dev.Preview(doc)
	got := map[string]SectionPreview{}
	for _, s := range preview.Sections {
		got[s.Section] = s
	}
	if got["future"].Status != StatusUnknown {
		t.Fatalf("future: %+v", got["future"])
	}
	if got["logging"].Status != StatusChange || len(got["logging"].Warnings) != 1 || !strings.Contains(got["logging"].Warnings[0], "colour") {
		t.Fatalf("logging: %+v", got["logging"])
	}

	applied, err := b.dev.Import(doc, nil)
	if err != nil || !reflect.DeepEqual(applied, []string{"logging"}) {
		t.Fatalf("applied %v, err %v", applied, err)
	}
	if b.config().Logger.Level != "warn" {
		t.Fatal("the level was not imported")
	}
	// Choosing a section the server does not know refuses the import.
	if _, err := b.dev.Import(doc, []string{"future"}); err == nil {
		t.Fatal("an unknown section was accepted")
	}
}

func TestMissingFieldsKeepTheBoardValue(t *testing.T) {
	b := newBoard(t)
	conf := baseConfig()
	conf.Watchdog = config.Watchdog{TimeoutMinutes: 7, PingHost: "10.0.0.5"}
	b.writeConfig(conf)

	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"watchdog":{"maxPerHour":4}}}`))
	if _, err := b.dev.Import(doc, nil); err != nil {
		t.Fatal(err)
	}
	w := b.config().Watchdog
	if w.MaxPerHour != 4 || w.TimeoutMinutes != 7 || w.PingHost != "10.0.0.5" {
		t.Fatalf("watchdog %+v", w)
	}
}

func TestInvalidSectionStopsTheWholeImport(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	before := b.read(config.ConfigurationFile)

	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{
		"webTitle":{"title":"New"},
		"logging":{"level":"debug"},
		"vnc":{"port":80}
	}}`))

	preview := b.dev.Preview(doc)
	for _, s := range preview.Sections {
		if s.Section == "vnc" && (s.Status != StatusInvalid || !strings.Contains(s.Error, "web server")) {
			t.Fatalf("vnc preview %+v", s)
		}
	}

	_, err := b.dev.Import(doc, nil)
	var importErr *ImportError
	if !errors.As(err, &importErr) || importErr.Sections["vnc"] == "" {
		t.Fatalf("err = %v", err)
	}
	if b.read(config.ConfigurationFile) != before || b.exists(webTitleFile) {
		t.Fatal("a refused import wrote a file")
	}

	// Leaving the bad section out imports the rest.
	applied, err := b.dev.Import(doc, []string{"webTitle", "logging"})
	if err != nil || len(applied) != 2 {
		t.Fatalf("applied %v, err %v", applied, err)
	}
	if b.read(webTitleFile) != "New" {
		t.Fatal("the title was not imported")
	}
}

func TestValidation(t *testing.T) {
	for _, tc := range []struct {
		section string
		value   string
		wantErr string
	}{
		{"logging", `{"level":"loud"}`, "log level"},
		{"sessions", `{"refreshTokenDuration":0}`, "session length"},
		{"loginSecurity", `{"loginLockoutDuration":60,"loginMaxFailures":0}`, "at least one failure"},
		{"watchdog", `{"action":"explode"}`, "reset or power"},
		{"watchdog", `{"pingHost":"not an ip"}`, "IP address"},
		{"nut", `{"enabled":true}`, "UPS name"},
		{"nut", `{"host":"bad host!"}`, "server address"},
		{"vnc", `{"maxFps":200}`, "frame rate"},
		{"webrtc", `{"udpPort":80}`, "1024"},
		{"webrtc", `{"stun":"user:pass@stun.example:3478"}`, "STUN"},
		{"webrtc", `{"stun":"stun.example:70000"}`, "STUN"},
		{"webrtc", `{"stun":"stun.example"}`, "STUN"},
		{"picoclaw", `{"screenshotWidth":5}`, "width"},
		{"hostLog", `{"bind":"eth0"}`, "bind"},
		{"webTitle", `{"title":"a\u0007b"}`, "control"},
		{"leaderKey", `{"key":"rm -rf"}`, "leader key"},
		{"keyboardLayout", `{"layout":"klingon"}`, "keyboard layout"},
		{"shortcuts", `{"shortcuts":[{"id":"x","keys":[]}]}`, "keys"},
		{"macros", `{"macros":[{"id":"m","name":"x","steps":[]}]}`, "no steps"},
		{"macros", `{"macros":[{"id":"m","name":"","steps":[{"type":"delay","ms":5}]}]}`, "name"},
		{"mouseJiggler", `{"enabled":true,"mode":"sideways"}`, "relative or absolute"},
		{"keyJiggler", `{"key":"a"}`, "unknown jiggler key"},
		{"oled", `{"sleepSeconds":5}`, "sleep delay"},
		{"oled", `{"brightness":1}`, "brightness"},
		{"hdmi", `{"idleTimeoutMinutes":-1}`, "idle timeout"},
		{"roomMic", `{"gain":99}`, "gain"},
		{"notifications", `{"loginThreshold":5000}`, "login threshold"},
		{"notifications", `{"events":{"made.up":{"enabled":true}}}`, "unknown event"},
		{"redfish", `{"enabled":"yes"}`, "wrong type"},
		{"ipmi", `[1,2]`, "not a JSON object"},
	} {
		t.Run(tc.section+" "+tc.wantErr, func(t *testing.T) {
			b := newBoard(t)
			b.writeConfig(baseConfig())
			doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"`+tc.section+`":`+tc.value+`}}`))
			preview := b.dev.Preview(doc)
			if len(preview.Sections) != 1 || preview.Sections[0].Status != StatusInvalid {
				t.Fatalf("preview %+v", preview.Sections)
			}
			if !strings.Contains(preview.Sections[0].Error, tc.wantErr) {
				t.Fatalf("error %q, want it to mention %q", preview.Sections[0].Error, tc.wantErr)
			}
		})
	}
}

func TestPowerActionsNeedThePowerLED(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())

	alone := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"watchdog":{"enabled":true,"action":"power"}}}`))
	if _, err := b.dev.Import(alone, nil); err == nil || !strings.Contains(err.Error(), "power LED") {
		t.Fatalf("err = %v", err)
	}

	// The same file that says the LED is connected is enough.
	withLED := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{
		"watchdog":{"enabled":true,"action":"power"},
		"powerLed":{"powerLed":true}}}`))
	if _, err := b.dev.Import(withLED, nil); err != nil {
		t.Fatal(err)
	}

	nut := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{
		"powerLed":{"powerLed":false},
		"nut":{"enabled":true,"host":"ups.lan","ups":"ups"}}}`))
	if _, err := b.dev.Import(nut, nil); err == nil || !strings.Contains(err.Error(), "power LED") {
		t.Fatalf("err = %v", err)
	}
}

func TestPlainVNCAuthNeedsAPasswordOnTheBoard(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"vnc":{"enabled":true,"vncAuth":true}}}`))

	if _, err := b.dev.Import(doc, nil); err == nil || !strings.Contains(err.Error(), "VNC password") {
		t.Fatalf("err = %v", err)
	}
	b.write(vncPasswordFile, "secret")
	if _, err := b.dev.Import(doc, nil); err != nil {
		t.Fatal(err)
	}
}

func TestFailedWritePutsTheFilesBack(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	b.write(keyJigglerFile, "off\nf15\n")
	beforeConf := b.read(config.ConfigurationFile)

	calls := 0
	rename = func(from, to string) error {
		calls++
		if calls == 3 {
			return errors.New("disk full")
		}
		return os.Rename(from, to)
	}
	t.Cleanup(func() { rename = os.Rename })

	// key-jiggler, oled_contrast, oled_sleep, server.yaml: the third rename
	// fails, after two files are in place.
	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{
		"keyJiggler":{"enabled":true,"key":"shift"},
		"oled":{"sleepSeconds":60,"brightness":100},
		"logging":{"level":"debug"}}}`))
	if _, err := b.dev.Import(doc, nil); err == nil {
		t.Fatal("the import reported success")
	}

	if got := b.read(keyJigglerFile); got != "off\nf15\n" {
		t.Errorf("key-jiggler = %q, not put back", got)
	}
	if b.exists(oledBrightnessFile) || b.exists(oledSleepFile) {
		t.Error("an OLED file was left behind")
	}
	if b.read(config.ConfigurationFile) != beforeConf {
		t.Error("server.yaml changed")
	}
	entries, _ := os.ReadDir(filepath.Join(b.root, "etc", "kvm"))
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".settings-import-") {
			t.Errorf("temporary file %s left behind", e.Name())
		}
	}
}

func TestDamagedFileCanBeRepaired(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	b.write(macrosFile, "{not json")

	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"macros":{"macros":[]}}}`))
	preview := b.dev.Preview(doc)
	if preview.Sections[0].Status == StatusInvalid || len(preview.Sections[0].Warnings) == 0 {
		t.Fatalf("preview %+v", preview.Sections[0])
	}
	if _, err := b.dev.Import(doc, nil); err != nil {
		t.Fatal(err)
	}
	if got := b.read(macrosFile); got != `{"macros":[]}` {
		t.Fatalf("macros.json = %q", got)
	}
}

func TestDefaultTitleRemovesTheFile(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	b.write(webTitleFile, "Old")
	b.write(mouseJigglerFile, "relative")

	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{
		"webTitle":{"title":"IronKVM"},
		"mouseJiggler":{"enabled":false}}}`))
	if _, err := b.dev.Import(doc, nil); err != nil {
		t.Fatal(err)
	}
	if b.exists(webTitleFile) || b.exists(mouseJigglerFile) {
		t.Fatal("a file that means the default was kept")
	}
}

func TestKeyboardLayoutImportsItsID(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	b.write(keyboardLayoutFile, "fr")

	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"keyboardLayout":{"layout":" PT-BR "}}}`))
	if _, err := b.dev.Import(doc, nil); err != nil {
		t.Fatal(err)
	}
	if got := b.read(keyboardLayoutFile); got != "pt-br" {
		t.Fatalf("layout file = %q, want pt-br", got)
	}

	doc = parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"keyboardLayout":{"layout":""}}}`))
	if _, err := b.dev.Import(doc, nil); err != nil {
		t.Fatal(err)
	}
	if b.exists(keyboardLayoutFile) {
		t.Fatal("the default layout kept the file")
	}
}

func TestPreviewSummarisesLongLists(t *testing.T) {
	b := newBoard(t)
	b.writeConfig(baseConfig())
	var macros []string
	for i := 0; i < 20; i++ {
		macros = append(macros, fmt.Sprintf(`{"id":"m%d","name":"Macro %d","steps":[{"type":"delay","ms":10}]}`, i, i))
	}
	doc := parse(t, []byte(`{"format":"ironkvm-settings","version":1,"sections":{"macros":{"macros":[`+strings.Join(macros, ",")+`]}}}`))

	s := b.dev.Preview(doc).Sections[0]
	if s.Status != StatusChange {
		t.Fatalf("%+v", s)
	}
	if got, ok := s.Changes[0].To.(ListSummary); !ok || got.Items != 20 {
		t.Fatalf("to = %#v", s.Changes[0].To)
	}
}

func TestFileName(t *testing.T) {
	now := time.Date(2026, 10, 26, 9, 5, 0, 0, time.UTC)
	if got := FileName("ironkvm-settings", "kvm/../x y", now, "json"); got != "ironkvm-settings-kvm-..-x-y-20261026-090500.json" {
		t.Fatal(got)
	}
	if got := FileName("p", "", now, "json"); !strings.HasPrefix(got, "p-board-") {
		t.Fatal(got)
	}
}
