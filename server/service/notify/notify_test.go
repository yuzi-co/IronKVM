package notify

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/service/events"
)

const (
	testBotToken = "123456789:AAEhBOweik6ad6PsWb9gUcPhjnJmTcXSkwq"
	testNtfyTok  = "tk_supersecrettokenvalue"
	testHookPath = "/hooks/SECRETPATHVALUE"
)

var testJPEG = []byte{0xff, 0xd8, 0xff, 0xe0, 'J', 'F', 'I', 'F', 0xff, 0xd9}

type clock struct {
	mu sync.Mutex
	t  time.Time
}

func (c *clock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *clock) advance(d time.Duration) {
	c.mu.Lock()
	c.t = c.t.Add(d)
	c.mu.Unlock()
}

type request struct {
	Method string
	Path   string
	Header http.Header
	Body   []byte
}

type sink struct {
	srv  *httptest.Server
	mu   sync.Mutex
	reqs []request
	code int
	body string
}

func newSink(t *testing.T) *sink {
	t.Helper()
	s := &sink{code: 200}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		s.mu.Lock()
		s.reqs = append(s.reqs, request{r.Method, r.URL.Path, r.Header.Clone(), body})
		code, reply := s.code, s.body
		s.mu.Unlock()
		w.WriteHeader(code)
		_, _ = w.Write([]byte(reply))
	}))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *sink) requests() []request {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]request(nil), s.reqs...)
}

type fixture struct {
	svc   *Service
	store *Store
	bus   *events.Bus
	clk   *clock
	shots int
	mu    sync.Mutex
}

func newFixture(t *testing.T, set Settings) *fixture {
	t.Helper()
	f := &fixture{
		store: NewStore(filepath.Join(t.TempDir(), "notify.json")),
		bus:   events.NewBus(),
		clk:   &clock{t: time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)},
	}
	if _, err := f.store.Update(func(s *Settings) error { *s = set; return nil }); err != nil {
		t.Fatalf("seed the settings: %s", err)
	}
	f.svc = New(Deps{
		Store: f.store,
		Bus:   f.bus,
		Snapshot: func(context.Context) ([]byte, error) {
			f.mu.Lock()
			f.shots++
			f.mu.Unlock()
			return testJPEG, nil
		},
		Device: func() string { return "kvm1" },
		Now:    f.clk.now,
	})
	return f
}

func (f *fixture) snapshots() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.shots
}

func TestDefaultsAreOffAndAllEventsOn(t *testing.T) {
	s := Settings{}.WithDefaults()
	if s.Enabled || s.Webhook.Enabled || s.Ntfy.Enabled || s.Telegram.Enabled {
		t.Fatal("something is on by default")
	}
	for _, typ := range Types {
		r, ok := s.Events[typ]
		if !ok || !r.Enabled || r.CooldownSeconds <= 0 {
			t.Errorf("%s: rule %+v", typ, r)
		}
	}
	if s.LoginThreshold != DefaultLoginThreshold || s.Ntfy.Server != DefaultNtfyServer {
		t.Errorf("defaults: %+v", s)
	}
}

func TestValidateRefusesWhatCannotWork(t *testing.T) {
	good := Settings{}.WithDefaults()
	if err := good.Validate(); err != nil {
		t.Fatalf("the defaults do not validate: %s", err)
	}

	cases := map[string]func(*Settings){
		"webhook on without an address": func(s *Settings) { s.Webhook.Enabled = true },
		"webhook not http":              func(s *Settings) { s.Webhook.URL = "ftp://x/y" },
		"ntfy on without a topic":       func(s *Settings) { s.Ntfy.Enabled = true },
		"ntfy topic with a slash":       func(s *Settings) { s.Ntfy.Topic = "a/b" },
		"telegram token malformed":      func(s *Settings) { s.Telegram.BotToken = "nonsense" },
		"telegram on without a chat": func(s *Settings) {
			s.Telegram.Enabled, s.Telegram.BotToken = true, testBotToken
		},
		"telegram chat malformed": func(s *Settings) { s.Telegram.ChatID = "not a chat" },
		"cooldown too long": func(s *Settings) {
			s.Events[events.SignalLost] = Rule{Enabled: true, CooldownSeconds: 999999}
		},
		"unknown event":          func(s *Settings) { s.Events["nope"] = Rule{} },
		"login threshold zero":   func(s *Settings) { s.LoginThreshold = 0 },
		"login window too short": func(s *Settings) { s.LoginWindowSeconds = 1 },
	}
	for name, mutate := range cases {
		s := Settings{}.WithDefaults()
		mutate(&s)
		if err := s.Validate(); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
}

func TestTheValidationErrorNeverRepeatsASecret(t *testing.T) {
	s := Settings{}.WithDefaults()
	s.Webhook.URL = "ftp://host" + testHookPath
	err := s.Validate()
	if err == nil || strings.Contains(err.Error(), "SECRETPATHVALUE") {
		t.Fatalf("error %v", err)
	}
}

func TestSettingsAreWrittenPrivateAndReadBack(t *testing.T) {
	path := filepath.Join(t.TempDir(), "sub", "notify.json")
	st := NewStore(path)
	_, err := st.Update(func(s *Settings) error {
		s.Enabled = true
		s.Telegram = Telegram{Enabled: true, BotToken: testBotToken, ChatID: "42"}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if mode := info.Mode().Perm(); mode&0o077 != 0 {
		t.Errorf("the settings file is %v, want owner only", mode)
	}
	got, err := st.Load()
	if err != nil || got.Telegram.BotToken != testBotToken || !got.Enabled {
		t.Fatalf("read back %+v, %v", got, err)
	}
}

func TestAnUnreadableFileDoesNotBlockASave(t *testing.T) {
	path := filepath.Join(t.TempDir(), "notify.json")
	if err := os.WriteFile(path, []byte("{broken"), 0o600); err != nil {
		t.Fatal(err)
	}
	st := NewStore(path)
	if _, err := st.Load(); err == nil {
		t.Fatal("a broken file loaded")
	}
	if _, err := st.Update(func(s *Settings) error { s.Enabled = true; return nil }); err != nil {
		t.Fatalf("save over a broken file: %s", err)
	}
}

func TestTheViewHidesEverySecret(t *testing.T) {
	s := Settings{}.WithDefaults()
	s.Webhook = Webhook{Enabled: true, URL: "https://hooks.example.com" + testHookPath}
	s.Ntfy = Ntfy{Enabled: true, Server: "https://ntfy.sh", Topic: "kvm", Token: testNtfyTok}
	s.Telegram = Telegram{Enabled: true, BotToken: testBotToken, ChatID: "42"}

	data, err := json.Marshal(viewOf(s))
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{testBotToken, testNtfyTok, "SECRETPATHVALUE", "AAEhBOweik"} {
		if strings.Contains(string(data), secret) {
			t.Errorf("the view carries %q: %s", secret, data)
		}
	}
	var v View
	_ = json.Unmarshal(data, &v)
	if !v.Webhook.URLSet || !v.Ntfy.TokenSet || !v.Telegram.BotTokenSet {
		t.Errorf("the view does not say the secrets are set: %s", data)
	}
	if v.Webhook.URLHint != "https://hooks.example.com/..." {
		t.Errorf("hint %q", v.Webhook.URLHint)
	}
}

func TestASaveWithoutASecretKeepsIt(t *testing.T) {
	cur := Settings{}.WithDefaults()
	cur.Ntfy.Token = testNtfyTok
	cur.Webhook.URL = "https://h.example.com/x"

	var req setRequest
	req.Ntfy.Server = "https://ntfy.sh"
	req.apply(&cur)
	if cur.Ntfy.Token != testNtfyTok || cur.Webhook.URL != "https://h.example.com/x" {
		t.Fatalf("a secret was lost: %+v", cur)
	}

	empty := ""
	req.Ntfy.Token = &empty
	req.apply(&cur)
	if cur.Ntfy.Token != "" {
		t.Error("an empty string did not clear the token")
	}
}

func TestWebhookPostsJSONAndTheSnapshotOnlyWhenAsked(t *testing.T) {
	sk := newSink(t)
	ev := events.Event{Type: events.WatchdogAction, Time: time.Unix(0, 0), Title: "T", Message: "M",
		Fields: map[string]string{"a": "b"}}
	m := message{ev: ev, device: "kvm1", jpeg: testJPEG}

	if err := sendWebhook(context.Background(), http.DefaultClient, Webhook{URL: sk.srv.URL + "/h"}, m); err != nil {
		t.Fatal(err)
	}
	if err := sendWebhook(context.Background(), http.DefaultClient, Webhook{URL: sk.srv.URL + "/h", IncludeSnapshot: true}, m); err != nil {
		t.Fatal(err)
	}

	reqs := sk.requests()
	if len(reqs) != 2 || reqs[0].Method != http.MethodPost || reqs[0].Header.Get("Content-Type") != "application/json" {
		t.Fatalf("requests: %+v", reqs)
	}
	var plain, with map[string]any
	_ = json.Unmarshal(reqs[0].Body, &plain)
	_ = json.Unmarshal(reqs[1].Body, &with)
	if plain["type"] != events.WatchdogAction || plain["device"] != "kvm1" || plain["title"] != "T" {
		t.Errorf("body %v", plain)
	}
	if _, ok := plain["snapshot"]; ok {
		t.Error("the snapshot went out without being asked for")
	}
	raw, _ := base64.StdEncoding.DecodeString(with["snapshot"].(string))
	if !bytes.Equal(raw, testJPEG) {
		t.Error("the snapshot is not the picture")
	}
}

func TestNtfySendsThePictureAsTheBodyAndTheTextInHeaders(t *testing.T) {
	sk := newSink(t)
	m := message{
		ev:     events.Event{Type: events.SignalRestored, Title: "kvm1: Signal back \u2713", Message: "line", Fields: map[string]string{"k": "v"}},
		device: "kvm1", jpeg: testJPEG,
	}
	err := sendNtfy(context.Background(), http.DefaultClient, Ntfy{Server: sk.srv.URL + "/", Topic: "my-topic", Token: testNtfyTok}, m)
	if err != nil {
		t.Fatal(err)
	}
	r := sk.requests()[0]
	if r.Path != "/my-topic" || r.Method != http.MethodPost {
		t.Errorf("%s %s", r.Method, r.Path)
	}
	if !bytes.Equal(r.Body, testJPEG) || r.Header.Get("Content-Type") != "image/jpeg" || r.Header.Get("Filename") == "" {
		t.Errorf("the picture is not the attachment: %v", r.Header)
	}
	if r.Header.Get("Authorization") != "Bearer "+testNtfyTok {
		t.Error("the token was not sent")
	}
	title, err := new(mime.WordDecoder).DecodeHeader(r.Header.Get("Title"))
	if err != nil || title != "kvm1: Signal back \u2713" {
		t.Errorf("title %q, %v", title, err)
	}
	if strings.ContainsAny(r.Header.Get("Message"), "\r\n") || r.Header.Get("Message") == "" {
		t.Errorf("message header %q", r.Header.Get("Message"))
	}
}

func TestNtfyWithoutAPictureSendsTheTextAsTheBody(t *testing.T) {
	sk := newSink(t)
	m := message{ev: events.Event{Type: events.ATXPress, Title: "T", Message: "body text"}}
	if err := sendNtfy(context.Background(), http.DefaultClient, Ntfy{Server: sk.srv.URL, Topic: "t"}, m); err != nil {
		t.Fatal(err)
	}
	if got := string(sk.requests()[0].Body); !strings.Contains(got, "body text") {
		t.Errorf("body %q", got)
	}
}

func TestTelegramSendsAPhotoWithACaptionOrAMessage(t *testing.T) {
	sk := newSink(t)
	old := telegramBase
	telegramBase = sk.srv.URL
	t.Cleanup(func() { telegramBase = old })

	tg := Telegram{BotToken: testBotToken, ChatID: "42"}
	m := message{ev: events.Event{Type: events.ATXPress, Title: "Hello", Message: "m"}, jpeg: testJPEG}
	if err := sendTelegram(context.Background(), http.DefaultClient, tg, m); err != nil {
		t.Fatal(err)
	}
	m.jpeg = nil
	if err := sendTelegram(context.Background(), http.DefaultClient, tg, m); err != nil {
		t.Fatal(err)
	}

	reqs := sk.requests()
	if reqs[0].Path != "/bot"+testBotToken+"/sendPhoto" {
		t.Fatalf("path %s", reqs[0].Path)
	}
	mt, params, _ := mime.ParseMediaType(reqs[0].Header.Get("Content-Type"))
	if mt != "multipart/form-data" {
		t.Fatalf("content type %s", mt)
	}
	form := multipart.NewReader(bytes.NewReader(reqs[0].Body), params["boundary"])
	seen := map[string][]byte{}
	for {
		p, err := form.NextPart()
		if err != nil {
			break
		}
		seen[p.FormName()], _ = io.ReadAll(p)
	}
	if string(seen["chat_id"]) != "42" || !strings.HasPrefix(string(seen["caption"]), "Hello") || !bytes.Equal(seen["photo"], testJPEG) {
		t.Errorf("form %v", seen)
	}

	if reqs[1].Path != "/bot"+testBotToken+"/sendMessage" {
		t.Fatalf("path %s", reqs[1].Path)
	}
	var j map[string]string
	_ = json.Unmarshal(reqs[1].Body, &j)
	if j["chat_id"] != "42" || !strings.HasPrefix(j["text"], "Hello") {
		t.Errorf("message %v", j)
	}
}

func TestAFailureNamesTheReasonButNeverTheSecret(t *testing.T) {
	sk := newSink(t)
	sk.code, sk.body = 400, `{"ok":false,"description":"Bad Request: chat not found"}`
	old := telegramBase
	telegramBase = sk.srv.URL
	t.Cleanup(func() { telegramBase = old })

	err := sendTelegram(context.Background(), http.DefaultClient, Telegram{BotToken: testBotToken, ChatID: "42"},
		message{ev: events.Event{Title: "x"}})
	if err == nil || !strings.Contains(err.Error(), "chat not found") || strings.Contains(err.Error(), testBotToken) {
		t.Fatalf("error %v", err)
	}
}

func TestATransportErrorLeavesOutTheURL(t *testing.T) {
	sk := newSink(t)
	url := sk.srv.URL + testHookPath
	sk.srv.Close()

	err := sendWebhook(context.Background(), http.DefaultClient, Webhook{URL: url},
		message{ev: events.Event{Title: "x"}})
	if err == nil {
		t.Fatal("no error from a closed server")
	}
	if strings.Contains(err.Error(), "SECRETPATHVALUE") || strings.Contains(err.Error(), "http://") {
		t.Errorf("the error carries the address: %v", err)
	}

	scrubbed := scrub(errors.New("Post https://api.telegram.org/bot"+testBotToken+"/x: boom"), testBotToken)
	if strings.Contains(scrubbed.Error(), testBotToken) {
		t.Errorf("scrub left the token: %v", scrubbed)
	}
}

func TestNothingIsSentWhileTheMasterSwitchIsOff(t *testing.T) {
	sk := newSink(t)
	set := Settings{Webhook: Webhook{Enabled: true, URL: sk.srv.URL + "/h"}}
	f := newFixture(t, set)

	f.svc.handle(events.Event{Type: events.ATXPress, Title: "x"})
	if n := len(sk.requests()); n != 0 {
		t.Fatalf("%d requests with the master switch off", n)
	}
}

func TestEachTypeHasItsOwnCooldownAndTheSuppressedAreCounted(t *testing.T) {
	sk := newSink(t)
	set := Settings{Enabled: true, Webhook: Webhook{Enabled: true, URL: sk.srv.URL + "/h"}}
	f := newFixture(t, set)

	send := func(typ string) { f.svc.handle(events.Event{Type: typ, Title: typ, NoSnapshot: true}) }
	send(events.SignalLost)
	send(events.SignalLost) // inside the cooldown
	send(events.SignalLost)
	send(events.ATXPress) // another type: its own limit
	if n := len(sk.requests()); n != 2 {
		t.Fatalf("%d requests, want 2", n)
	}

	f.clk.advance(2 * time.Minute)
	send(events.SignalLost)
	reqs := sk.requests()
	if len(reqs) != 3 {
		t.Fatalf("%d requests after the cooldown, want 3", len(reqs))
	}
	var body struct{ Fields map[string]string }
	_ = json.Unmarshal(reqs[2].Body, &body)
	if body.Fields["suppressed"] != "2" {
		t.Errorf("fields %v, want 2 suppressed", body.Fields)
	}

	suppressed := 0
	for _, e := range f.svc.State().Recent {
		if e.Suppressed {
			suppressed++
		}
	}
	if suppressed != 2 {
		t.Errorf("the page lists %d suppressed, want 2", suppressed)
	}
}

func TestADisabledTypeIsNotSent(t *testing.T) {
	sk := newSink(t)
	set := Settings{Enabled: true, Webhook: Webhook{Enabled: true, URL: sk.srv.URL + "/h"},
		Events: map[string]Rule{events.ATXPress: {Enabled: false, CooldownSeconds: 1}}}
	f := newFixture(t, set)

	f.svc.handle(events.Event{Type: events.ATXPress, Title: "x"})
	f.svc.handle(events.Event{Type: "something.else", Title: "x"})
	if n := len(sk.requests()); n != 0 {
		t.Fatalf("%d requests", n)
	}
}

func TestTheSnapshotIsTakenOnlyWhenWanted(t *testing.T) {
	sk := newSink(t)
	set := Settings{Enabled: true, Snapshot: true, Webhook: Webhook{Enabled: true, URL: sk.srv.URL + "/h"}}
	f := newFixture(t, set)

	f.svc.handle(events.Event{Type: events.SignalLost, Title: "no picture", NoSnapshot: true})
	if f.snapshots() != 0 {
		t.Error("a snapshot was taken for an event that has nothing to see")
	}
	f.svc.handle(events.Event{Type: events.ATXPress, Title: "has picture"})
	if f.snapshots() != 1 {
		t.Errorf("%d snapshots, want 1", f.snapshots())
	}
	f.svc.handle(events.Event{Type: events.WatchdogAction, Title: "has its own", JPEG: testJPEG})
	if f.snapshots() != 1 {
		t.Error("a second snapshot was taken for an event that carried one")
	}

	if _, err := f.store.Update(func(s *Settings) error { s.Snapshot = false; return nil }); err != nil {
		t.Fatal(err)
	}
	f.clk.advance(time.Hour)
	f.svc.handle(events.Event{Type: events.ATXPress, Title: "off"})
	if f.snapshots() != 1 {
		t.Error("a snapshot was taken with snapshots off")
	}
}

func TestFailedLoginsRaiseOneEventAtTheThreshold(t *testing.T) {
	sk := newSink(t)
	set := Settings{Enabled: true, LoginThreshold: 3, LoginWindowSeconds: 60,
		Webhook: Webhook{Enabled: true, URL: sk.srv.URL + "/h"}}
	f := newFixture(t, set)

	fail := func() {
		f.svc.handle(events.Event{Type: events.LoginFailure, Fields: map[string]string{"source": "10.0.0.9"}})
	}
	fail()
	fail()
	if n := len(sk.requests()); n != 0 {
		t.Fatalf("%d requests below the threshold", n)
	}
	fail()
	reqs := sk.requests()
	if len(reqs) != 1 {
		t.Fatalf("%d requests at the threshold, want 1", len(reqs))
	}
	var body struct {
		Type   string
		Fields map[string]string
	}
	_ = json.Unmarshal(reqs[0].Body, &body)
	if body.Type != events.LoginFailures || body.Fields["count"] != "3" || body.Fields["lastSource"] != "10.0.0.9" {
		t.Errorf("body %+v", body)
	}
}

func TestOldLoginFailuresFallOutOfTheWindow(t *testing.T) {
	sk := newSink(t)
	set := Settings{Enabled: true, LoginThreshold: 2, LoginWindowSeconds: 60,
		Webhook: Webhook{Enabled: true, URL: sk.srv.URL + "/h"}}
	f := newFixture(t, set)

	f.svc.handle(events.Event{Type: events.LoginFailure})
	f.clk.advance(2 * time.Minute)
	f.svc.handle(events.Event{Type: events.LoginFailure})
	if n := len(sk.requests()); n != 0 {
		t.Fatalf("%d requests for failures a window apart", n)
	}
}

func TestTestSendsOnTheNamedChannelWhateverTheSwitchesSay(t *testing.T) {
	sk := newSink(t)
	set := Settings{Webhook: Webhook{URL: sk.srv.URL + "/h"}, Ntfy: Ntfy{Server: sk.srv.URL, Topic: "t"}}
	f := newFixture(t, set)

	res, err := f.svc.Test(ChannelWebhook)
	if err != nil || res[ChannelWebhook] != "ok" || len(res) != 1 {
		t.Fatalf("%v, %v", res, err)
	}
	if _, err := f.svc.Test(ChannelTelegram); !errors.Is(err, ErrNoChannel) {
		t.Errorf("telegram is not set up, got %v", err)
	}
	if _, err := f.svc.Test("carrier-pigeon"); !errors.Is(err, ErrNoChannel) {
		t.Errorf("unknown channel, got %v", err)
	}
	res, err = f.svc.Test("")
	if err != nil || len(res) != 2 {
		t.Fatalf("all channels: %v, %v", res, err)
	}
	if st := f.svc.State(); !st.Channels[ChannelWebhook].OK {
		t.Errorf("state %+v", st)
	}
}

func TestAFailingChannelDoesNotStopTheOthers(t *testing.T) {
	bad, good := newSink(t), newSink(t)
	bad.code = 500
	set := Settings{Enabled: true,
		Webhook: Webhook{Enabled: true, URL: bad.srv.URL + "/h"},
		Ntfy:    Ntfy{Enabled: true, Server: good.srv.URL, Topic: "t"}}
	f := newFixture(t, set)

	f.svc.handle(events.Event{Type: events.ATXPress, Title: "x", NoSnapshot: true})
	if len(good.requests()) != 1 {
		t.Fatal("ntfy was not tried after the webhook failed")
	}
	st := f.svc.State()
	if st.Channels[ChannelWebhook].OK || st.Channels[ChannelWebhook].Error == "" || !st.Channels[ChannelNtfy].OK {
		t.Errorf("state %+v", st.Channels)
	}
}

type captureHook struct {
	mu   sync.Mutex
	text strings.Builder
}

func (h *captureHook) Levels() []log.Level { return log.AllLevels }
func (h *captureHook) Fire(e *log.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.text.WriteString(e.Message)
	for k, v := range e.Data {
		data, _ := json.Marshal(v)
		h.text.WriteString(k)
		h.text.Write(data)
	}
	return nil
}

func TestNoSecretReachesTheLog(t *testing.T) {
	hook := &captureHook{}
	log.AddHook(hook)

	sk := newSink(t)
	sk.code = 500
	oldBase := telegramBase
	telegramBase = sk.srv.URL
	t.Cleanup(func() { telegramBase = oldBase })

	set := Settings{Enabled: true,
		Webhook:  Webhook{Enabled: true, URL: sk.srv.URL + testHookPath},
		Ntfy:     Ntfy{Enabled: true, Server: sk.srv.URL, Topic: "t", Token: testNtfyTok},
		Telegram: Telegram{Enabled: true, BotToken: testBotToken, ChatID: "42"}}
	f := newFixture(t, set)
	f.svc.handle(events.Event{Type: events.ATXPress, Title: "x", NoSnapshot: true})
	sk.srv.Close()
	f.clk.advance(time.Hour)
	f.svc.handle(events.Event{Type: events.ATXPress, Title: "x", NoSnapshot: true})

	hook.mu.Lock()
	logged := hook.text.String()
	hook.mu.Unlock()
	if !strings.Contains(logged, "notify:") {
		t.Fatalf("nothing was logged, so the check proves nothing: %q", logged)
	}
	for _, secret := range []string{testBotToken, testNtfyTok, "SECRETPATHVALUE"} {
		if strings.Contains(logged, secret) {
			t.Errorf("the log carries %q:\n%s", secret, logged)
		}
	}
}

func TestTheSignalMustHoldForTwoReadsBeforeItIsReported(t *testing.T) {
	sk := newSink(t)
	set := Settings{Enabled: true, Webhook: Webhook{Enabled: true, URL: sk.srv.URL + "/h"}}
	f := newFixture(t, set)

	var mu sync.Mutex
	present := true
	f.svc.deps.Signal = func() (bool, bool) {
		mu.Lock()
		defer mu.Unlock()
		return present, true
	}
	f.svc.deps.SignalInterval = 20 * time.Millisecond

	got := make(chan events.Event, 8)
	cancelSub := f.bus.Subscribe(func(e events.Event) { got <- e })
	defer cancelSub()
	stop := f.svc.run()
	defer stop()

	setPresent := func(p bool) {
		mu.Lock()
		present = p
		mu.Unlock()
	}
	expect := func(typ string) {
		t.Helper()
		select {
		case e := <-got:
			if e.Type != typ {
				t.Fatalf("got %s, want %s", e.Type, typ)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("no %s event", typ)
		}
	}
	quiet := func() {
		t.Helper()
		select {
		case e := <-got:
			t.Fatalf("unexpected %s", e.Type)
		case <-time.After(150 * time.Millisecond):
		}
	}

	time.Sleep(100 * time.Millisecond) // the baseline is taken, silently
	quiet()
	setPresent(false)
	expect(events.SignalLost)
	setPresent(true)
	expect(events.SignalRestored)

	// A flicker: down for less than one interval, so at most one read sees it.
	setPresent(false)
	time.Sleep(5 * time.Millisecond)
	setPresent(true)
	quiet()
}

// Another service that listens for the signal (the dashcam) has it watched
// while notifications are off; nobody listening leaves it unread.
func TestTheSignalIsWatchedForAnotherService(t *testing.T) {
	for _, wanted := range []bool{true, false} {
		f := newFixture(t, Settings{})

		var mu sync.Mutex
		present, reads := true, 0
		f.svc.deps.Signal = func() (bool, bool) {
			mu.Lock()
			defer mu.Unlock()
			reads++
			return present, true
		}
		f.svc.deps.SignalWanted = func() bool { return wanted }
		f.svc.deps.SignalInterval = 20 * time.Millisecond

		got := make(chan events.Event, 8)
		cancelSub := f.bus.Subscribe(func(e events.Event) { got <- e })
		stop := f.svc.run()

		time.Sleep(100 * time.Millisecond)
		mu.Lock()
		present = false
		mu.Unlock()

		select {
		case e := <-got:
			if !wanted || e.Type != events.SignalLost {
				t.Fatalf("wanted %t: got %s", wanted, e.Type)
			}
		case <-time.After(500 * time.Millisecond):
			if wanted {
				t.Fatal("no signal.lost event for the other service")
			}
		}
		stop()
		cancelSub()

		mu.Lock()
		if !wanted && reads != 0 {
			t.Fatalf("the signal was read %d times with nobody listening", reads)
		}
		mu.Unlock()
	}
}
