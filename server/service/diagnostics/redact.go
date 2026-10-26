package diagnostics

import (
	"encoding/json"
	"net/url"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"NanoKVM-Server/config"
)

// Redacted replaces a secret in the bundle.
const Redacted = "[REDACTED]"

// minSecretLength is the shortest known value the redactor replaces. A
// shorter one, such as a two-digit chat ID, matches too much ordinary text
// for its replacement to say anything, and the patterns still apply.
const minSecretLength = 6

// secretFiles hold a secret in their whole content, or JSON whose secret
// fields secretKeyName picks. The redactor reads them only to learn what to
// remove; none of them is copied into the bundle.
var secretFiles = []string{
	"/etc/kvm/pwd",
	"/etc/kvm/.jwt_secret",
	"/etc/kvm/totp.key",
	"/etc/kvm/api_keys.json",
	"/etc/kvm/share_links.json",
	"/etc/kvm/mcp.json",
	"/etc/kvm/.picoclaw_internal_token",
	"/etc/kvm/first-boot-password",
	"/boot/ironkvm-first-password.txt",
	"/etc/kvm/wifi.ssid",
	"/etc/kvm/wifi.pass",
	"/etc/kvm/vnc.passwd",
	"/etc/kvm/nut.passwd",
	"/etc/kvm/ipmi.key",
	"/etc/kvm/notify.json",
	"/etc/kvm/application-update.json",
	"/etc/kvm/server.key",
	"/etc/kvm/uploaded.key",
}

// secretKeyName picks the fields of a JSON secret file whose values are
// secret. Dates and counts in the same files are left alone, so they do not
// turn every timestamp in the logs into a redaction.
var secretKeyName = regexp.MustCompile(`(?i)pass|secret|token|hash|key|cred|recovery|url|topic|chat|server`)

// patterns catch what has the shape of a secret, whether or not the board
// knows the value. Each one keeps the text that says what was removed.
var patterns = []struct {
	re   *regexp.Regexp
	with string
}{
	{regexp.MustCompile(`(?s)-----BEGIN [A-Z0-9 ]*PRIVATE KEY-----.*?(-----END [A-Z0-9 ]*PRIVATE KEY-----|\z)`), Redacted},
	// A JSON web token: three base64url parts, the first two JSON objects.
	{regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}\.[A-Za-z0-9_-]{8,}`), Redacted},
	{regexp.MustCompile(`(?i)\b(bearer|basic)\s+[A-Za-z0-9._~+/=-]{8,}`), "$1 " + Redacted},
	// The user and password of a URL.
	{regexp.MustCompile(`(?i)\b([a-z][a-z0-9+.-]*://)[^/\s:@]+:[^/\s@]+@`), "${1}" + Redacted + "@"},
	// name=value, name: value and "name": "value" for names that say secret.
	{regexp.MustCompile(`(?i)([a-z0-9_.-]*(?:password|passwd|passphrase|secret|token|api[_-]?key|auth[_-]?key|psk|credential|private[_-]?key)[a-z0-9_.-]*)("?\s*[=:]\s*"?)([^\s"',;&]+)`), "${1}${2}" + Redacted},
	// A Telegram bot token.
	{regexp.MustCompile(`\b[0-9]{6,12}:[A-Za-z0-9_-]{30,}`), Redacted},
	// The prefixes of the board's API keys, MCP keys and Tailscale keys.
	{regexp.MustCompile(`\bnkvm_[A-Za-z0-9_-]{8,}`), Redacted},
	{regexp.MustCompile(`\bnag_mcp_[A-Za-z0-9_-]{8,}`), Redacted},
	{regexp.MustCompile(`\btskey-[A-Za-z0-9-]{8,}`), Redacted},
}

// Redactor removes secrets from text: first every secret value it was given,
// then everything that matches a pattern.
type Redactor struct {
	values []string
}

// NewRedactor returns a redactor for the given values, ignoring the short
// ones.
func NewRedactor(values []string) *Redactor {
	seen := map[string]bool{}
	var keep []string
	for _, v := range withURLPasswords(values) {
		v = strings.TrimSpace(v)
		if len(v) < minSecretLength || seen[v] {
			continue
		}
		seen[v] = true
		keep = append(keep, v)
	}
	// The longest first, so a secret that contains another is removed whole.
	sort.Slice(keep, func(i, j int) bool { return len(keep[i]) > len(keep[j]) })
	return &Redactor{values: keep}
}

// Redact returns text without the secrets.
func (r *Redactor) Redact(text string) string {
	for _, v := range r.values {
		text = strings.ReplaceAll(text, v, Redacted)
	}
	for _, p := range patterns {
		text = p.re.ReplaceAllString(text, p.with)
	}
	return text
}

// knownSecrets reads every secret the board keeps that the redactor can find:
// the secret fields of server.yaml and the secret files beside it.
func (c *Collector) knownSecrets() []string {
	var out []string

	if data, err := c.readFile(config.ConfigurationFile); err == nil {
		var conf config.Config
		if yaml.Unmarshal(data, &conf) == nil {
			out = append(out, conf.JWT.SecretKey, conf.Turn.TurnUser, conf.Turn.TurnCred, conf.Proxy)
		}
	}

	for _, path := range secretFiles {
		data, err := c.readFile(path)
		if err != nil || len(data) == 0 {
			continue
		}
		var parsed any
		if json.Unmarshal(data, &parsed) == nil {
			if _, isObject := parsed.(map[string]any); isObject {
				out = append(out, secretStrings(parsed, false)...)
				continue
			}
			if _, isList := parsed.([]any); isList {
				out = append(out, secretStrings(parsed, false)...)
				continue
			}
		}
		text := string(data)
		out = append(out, text)
		for _, line := range strings.Split(text, "\n") {
			out = append(out, line)
		}
	}
	return out
}

// secretStrings returns the strings in v that sit under a key secretKeyName
// matches, however deep.
func secretStrings(v any, secret bool) []string {
	var out []string
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			out = append(out, secretStrings(child, secret || secretKeyName.MatchString(k))...)
		}
	case []any:
		for _, child := range t {
			out = append(out, secretStrings(child, secret)...)
		}
	case string:
		if secret {
			out = append(out, t)
		}
	}
	return out
}

// withURLPasswords adds, for each value that is a URL with a password, the
// password alone, so a log line that prints only the password loses it too.
func withURLPasswords(values []string) []string {
	out := append([]string(nil), values...)
	for _, v := range values {
		u, err := url.Parse(strings.TrimSpace(v))
		if err != nil || u.User == nil {
			continue
		}
		if password, ok := u.User.Password(); ok {
			out = append(out, password)
		}
	}
	return out
}
