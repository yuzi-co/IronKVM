package notify

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"NanoKVM-Server/service/events"
	"NanoKVM-Server/utils/atomicfile"
)

// File holds the notification settings, tokens included, so it is written
// 0600. It lives under /etc/kvm, which both slots share: keep the format
// backward compatible, and ignore keys you do not know.
const File = "/etc/kvm/notify.json"

// Types are the events the owner can switch on and off, in the order the page
// lists them.
var Types = []string{
	events.SignalLost,
	events.SignalRestored,
	events.WatchdogAction,
	events.LoginFailures,
	events.ATXPress,
	events.FailsafeEntered,
	events.HostLogMatch,
}

// Rule is what the owner says about one event type.
type Rule struct {
	Enabled bool `json:"enabled"`
	// CooldownSeconds is the least time between two notifications of this
	// type. Events inside it are counted and reported with the next one.
	CooldownSeconds int `json:"cooldownSeconds"`
}

// Webhook posts the event as JSON.
type Webhook struct {
	Enabled bool `json:"enabled"`
	// URL is a secret: a webhook address often carries its token.
	URL string `json:"url,omitempty"`
	// IncludeSnapshot adds the picture to the JSON as base64.
	IncludeSnapshot bool `json:"includeSnapshot,omitempty"`
}

// Ntfy publishes to a topic on ntfy.sh or a server of the owner's.
type Ntfy struct {
	Enabled bool   `json:"enabled"`
	Server  string `json:"server,omitempty"`
	Topic   string `json:"topic,omitempty"`
	// Token is a secret, for a server that needs a login.
	Token string `json:"token,omitempty"`
}

// Telegram sends through a bot.
type Telegram struct {
	Enabled bool `json:"enabled"`
	// BotToken is a secret.
	BotToken string `json:"botToken,omitempty"`
	ChatID   string `json:"chatId,omitempty"`
}

// Settings are the notification settings as the file keeps them.
type Settings struct {
	// Enabled is the master switch. Off by default.
	Enabled bool `json:"enabled"`
	// Snapshot attaches a JPEG of the screen where the channel can carry one.
	Snapshot bool `json:"snapshot"`

	Events map[string]Rule `json:"events,omitempty"`

	// LoginThreshold failed logins inside LoginWindowSeconds raise one
	// login.failures event.
	LoginThreshold     int `json:"loginThreshold,omitempty"`
	LoginWindowSeconds int `json:"loginWindowSeconds,omitempty"`

	Webhook  Webhook  `json:"webhook"`
	Ntfy     Ntfy     `json:"ntfy"`
	Telegram Telegram `json:"telegram"`
}

// The defaults, for a field the file does not carry.
const (
	DefaultCooldownSeconds    = 60
	DefaultLoginThreshold     = 5
	DefaultLoginWindowSeconds = 600
	DefaultNtfyServer         = "https://ntfy.sh"

	maxCooldownSeconds = 86400
)

// WithDefaults returns the settings with every missing value filled in. The
// Events map is a copy, so the result may be changed freely.
func (s Settings) WithDefaults() Settings {
	rules := make(map[string]Rule, len(Types))
	for _, t := range Types {
		r, ok := s.Events[t]
		if !ok {
			r = Rule{Enabled: true, CooldownSeconds: defaultCooldown(t)}
		}
		rules[t] = r
	}
	s.Events = rules

	if s.LoginThreshold <= 0 {
		s.LoginThreshold = DefaultLoginThreshold
	}
	if s.LoginWindowSeconds <= 0 {
		s.LoginWindowSeconds = DefaultLoginWindowSeconds
	}
	if s.Ntfy.Server == "" {
		s.Ntfy.Server = DefaultNtfyServer
	}
	return s
}

func defaultCooldown(t string) int {
	switch t {
	case events.LoginFailures:
		return 300
	case events.ATXPress:
		return 10
	}
	return DefaultCooldownSeconds
}

var (
	ntfyTopicRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)
	botTokenRe  = regexp.MustCompile(`^[0-9]{3,20}:[A-Za-z0-9_-]{20,80}$`)
	chatIDRe    = regexp.MustCompile(`^(-?[0-9]{1,20}|@[A-Za-z][A-Za-z0-9_]{4,31})$`)
)

// maxSecret bounds a token or URL a request may carry.
const maxSecret = 2048

func validHTTPURL(raw string) error {
	if len(raw) > maxSecret {
		return errors.New("the address is too long")
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return errors.New("the address must be an http or https URL")
	}
	return nil
}

// Validate checks the settings, defaults filled in. It never repeats a secret
// in its error.
func (s Settings) Validate() error {
	for t, r := range s.Events {
		known := false
		for _, k := range Types {
			known = known || k == t
		}
		if !known {
			return fmt.Errorf("unknown event type %q", t)
		}
		if r.CooldownSeconds < 0 || r.CooldownSeconds > maxCooldownSeconds {
			return errors.New("the cooldown must be 0 to 86400 seconds")
		}
	}
	if s.LoginThreshold < 1 || s.LoginThreshold > 1000 {
		return errors.New("the login threshold must be 1 to 1000")
	}
	if s.LoginWindowSeconds < 10 || s.LoginWindowSeconds > 86400 {
		return errors.New("the login window must be 10 to 86400 seconds")
	}

	if s.Webhook.URL != "" {
		if err := validHTTPURL(s.Webhook.URL); err != nil {
			return fmt.Errorf("webhook: %w", err)
		}
	}
	if s.Webhook.Enabled && s.Webhook.URL == "" {
		return errors.New("webhook: set the address first")
	}

	if err := validHTTPURL(s.Ntfy.Server); err != nil {
		return fmt.Errorf("ntfy server: %w", err)
	}
	if s.Ntfy.Topic != "" && !ntfyTopicRe.MatchString(s.Ntfy.Topic) {
		return errors.New("ntfy: the topic may hold letters, digits, - and _ (up to 64)")
	}
	if len(s.Ntfy.Token) > maxSecret {
		return errors.New("ntfy: the token is too long")
	}
	if s.Ntfy.Enabled && s.Ntfy.Topic == "" {
		return errors.New("ntfy: set the topic first")
	}

	if s.Telegram.BotToken != "" && !botTokenRe.MatchString(s.Telegram.BotToken) {
		return errors.New("telegram: that does not look like a bot token")
	}
	if s.Telegram.ChatID != "" && !chatIDRe.MatchString(s.Telegram.ChatID) {
		return errors.New("telegram: the chat is a number or @name")
	}
	if s.Telegram.Enabled && (s.Telegram.BotToken == "" || s.Telegram.ChatID == "") {
		return errors.New("telegram: set the bot token and the chat first")
	}
	return nil
}

// Store reads and writes the settings file.
type Store struct {
	path string
	mu   sync.Mutex
}

// NewStore returns a store on path.
func NewStore(path string) *Store { return &Store{path: path} }

// Load reads the settings with the defaults filled in. A missing file is the
// defaults: everything off.
func (st *Store) Load() (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	return st.loadLocked()
}

func (st *Store) loadLocked() (Settings, error) {
	data, err := os.ReadFile(st.path)
	if errors.Is(err, os.ErrNotExist) {
		return Settings{}.WithDefaults(), nil
	}
	if err != nil {
		return Settings{}, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return Settings{}, errors.New("the notification settings file is not valid JSON")
	}
	return s.WithDefaults(), nil
}

// Update loads the settings, lets fn change them, validates and saves the
// result, and returns it. fn runs under the store's lock.
func (st *Store) Update(fn func(*Settings) error) (Settings, error) {
	st.mu.Lock()
	defer st.mu.Unlock()

	s, err := st.loadLocked()
	if err != nil {
		// A damaged file must not stop the owner from writing a good one.
		s = Settings{}.WithDefaults()
	}
	if err := fn(&s); err != nil {
		return Settings{}, err
	}
	s = s.WithDefaults()
	if err := s.Validate(); err != nil {
		return Settings{}, err
	}
	if err := st.save(s); err != nil {
		return Settings{}, err
	}
	return s, nil
}

func (st *Store) save(s Settings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	dir := filepath.Dir(st.path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return atomicfile.WriteFile(st.path, data, 0o600)
}

// hint shows where a secret URL points without the part that is secret: the
// scheme and host.
func hint(raw string) string {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return ""
	}
	return strings.ToLower(u.Scheme) + "://" + u.Host + "/..."
}
