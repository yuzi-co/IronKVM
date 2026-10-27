package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	"NanoKVM-Server/authn"
)

// OIDCSettingsFile holds the OpenID Connect provider's settings, client
// secret included, so it is written 0600. The settings export leaves it out
// (service/settingsio is an allowlist and does not name it).
const OIDCSettingsFile = "/etc/kvm/oidc.json"

// The claims an IdP account may be matched to a local account by.
const (
	claimPreferredUsername = "preferred_username"
	claimEmail             = "email"
)

const (
	defaultProviderName = "SSO"
	maxProviderNameRune = 64
	maxSettingLength    = 512
	maxScopes           = 16
)

var (
	oidcSettingsFile = OIDCSettingsFile
	oidcSettingsMu   sync.Mutex
)

// oidcSettings is the file. ClientSecret is never sent to a browser: the
// API answers with oidcSettingsView.
type oidcSettings struct {
	Enabled bool `json:"enabled"`
	// ProviderName is the button's label: "Sign in with <ProviderName>".
	ProviderName string   `json:"providerName"`
	Issuer       string   `json:"issuer"`
	ClientID     string   `json:"clientId"`
	ClientSecret string   `json:"clientSecret,omitempty"`
	Scopes       []string `json:"scopes"`

	// UsernameClaim is the ID token claim matched against the local
	// username: preferred_username, or email, whose local part is the
	// username when the domain is EmailDomain.
	UsernameClaim string `json:"usernameClaim"`
	EmailDomain   string `json:"emailDomain,omitempty"`

	// AutoCreate makes a local account for an IdP user who has none, with
	// the role the groups give or DefaultRole.
	AutoCreate  bool       `json:"autoCreate"`
	DefaultRole authn.Role `json:"defaultRole"`

	// GroupsClaim names the claim that lists the user's groups. When any of
	// the three groups is set, the role comes from them at every login, and
	// a user in none of them is refused.
	GroupsClaim string `json:"groupsClaim,omitempty"`
	AdminGroup  string `json:"adminGroup,omitempty"`
	UserGroup   string `json:"userGroup,omitempty"`
	ViewerGroup string `json:"viewerGroup,omitempty"`

	// HidePasswordLogin folds the password form away behind a link on the
	// login page. The password login itself stays on: it is how an
	// administrator gets in when the IdP is down.
	HidePasswordLogin bool `json:"hidePasswordLogin"`

	// RedirectURL, when set, is the callback address registered at the IdP,
	// in place of the one derived from the request. A reverse proxy that does
	// not pass the Host header or the scheme on needs it.
	RedirectURL string `json:"redirectUrl,omitempty"`
}

func defaultOIDCSettings() oidcSettings {
	return oidcSettings{
		ProviderName:  defaultProviderName,
		Scopes:        []string{"openid", "profile", "email"},
		UsernameClaim: claimPreferredUsername,
		DefaultRole:   authn.RoleViewer,
		GroupsClaim:   "groups",
	}
}

// groupsMapped reports whether the role comes from the IdP's groups.
func (s *oidcSettings) groupsMapped() bool {
	return s.AdminGroup != "" || s.UserGroup != "" || s.ViewerGroup != ""
}

// usable reports whether a login can be started with these settings.
func (s *oidcSettings) usable() bool {
	return s.Enabled && s.Issuer != "" && s.ClientID != ""
}

func loadOIDCSettings() (oidcSettings, error) {
	oidcSettingsMu.Lock()
	defer oidcSettingsMu.Unlock()
	return loadOIDCSettingsLocked()
}

func loadOIDCSettingsLocked() (oidcSettings, error) {
	s := defaultOIDCSettings()
	data, err := os.ReadFile(oidcSettingsFile)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	if err != nil {
		return s, err
	}
	if err = json.Unmarshal(data, &s); err != nil {
		return defaultOIDCSettings(), fmt.Errorf("parse %s: %w", oidcSettingsFile, err)
	}
	return s, nil
}

func saveOIDCSettings(s oidcSettings) error {
	oidcSettingsMu.Lock()
	defer oidcSettingsMu.Unlock()
	return saveOIDCSettingsLocked(s)
}

func saveOIDCSettingsLocked(s oidcSettings) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(oidcSettingsFile)
	tmp, err := os.CreateTemp(dir, ".oidc-*.json")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer os.Remove(name)
	// CreateTemp makes the file 0600, so the secret is never readable to
	// others, not even for the moment before the rename.
	if err = tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err = tmp.Write(append(data, '\n')); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	return os.Rename(name, oidcSettingsFile)
}

// useOIDCSettingsFile points the settings at path, for tests.
func useOIDCSettingsFile(path string) (restore func()) {
	oidcSettingsMu.Lock()
	old := oidcSettingsFile
	oidcSettingsFile = path
	oidcSettingsMu.Unlock()
	return func() {
		oidcSettingsMu.Lock()
		oidcSettingsFile = old
		oidcSettingsMu.Unlock()
	}
}

// normalize trims the settings and checks them. An incomplete provider may
// be saved while it is off; turning it on needs the issuer and client ID.
func (s *oidcSettings) normalize() error {
	s.ProviderName = strings.TrimSpace(s.ProviderName)
	s.Issuer = strings.TrimRight(strings.TrimSpace(s.Issuer), "/")
	s.ClientID = strings.TrimSpace(s.ClientID)
	s.UsernameClaim = strings.TrimSpace(s.UsernameClaim)
	s.EmailDomain = strings.ToLower(strings.TrimPrefix(strings.TrimSpace(s.EmailDomain), "@"))
	s.GroupsClaim = strings.TrimSpace(s.GroupsClaim)
	s.AdminGroup = strings.TrimSpace(s.AdminGroup)
	s.UserGroup = strings.TrimSpace(s.UserGroup)
	s.ViewerGroup = strings.TrimSpace(s.ViewerGroup)
	s.RedirectURL = strings.TrimSpace(s.RedirectURL)

	if s.ProviderName == "" {
		s.ProviderName = defaultProviderName
	}
	if utf8.RuneCountInString(s.ProviderName) > maxProviderNameRune {
		return fmt.Errorf("the provider name is at most %d characters", maxProviderNameRune)
	}
	for _, field := range []string{s.ProviderName, s.Issuer, s.ClientID, s.ClientSecret, s.EmailDomain,
		s.GroupsClaim, s.AdminGroup, s.UserGroup, s.ViewerGroup, s.RedirectURL} {
		if len(field) > maxSettingLength {
			return fmt.Errorf("a value is longer than %d bytes", maxSettingLength)
		}
		if strings.IndexFunc(field, unicode.IsControl) >= 0 {
			return errors.New("a value holds a control character")
		}
	}
	if s.Issuer != "" {
		u, err := url.Parse(s.Issuer)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
			return errors.New("the issuer is an http or https URL")
		}
	}
	if s.RedirectURL != "" {
		u, err := url.Parse(s.RedirectURL)
		if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" || u.RawQuery != "" || u.Fragment != "" {
			return errors.New("the redirect URL is an http or https URL")
		}
	}
	if s.Enabled && (s.Issuer == "" || s.ClientID == "") {
		return errors.New("the issuer and client ID are needed to turn OIDC on")
	}

	switch s.UsernameClaim {
	case "", claimPreferredUsername:
		s.UsernameClaim = claimPreferredUsername
	case claimEmail:
		if s.EmailDomain == "" {
			return errors.New("matching by email needs the email domain")
		}
	default:
		return errors.New("the username claim is preferred_username or email")
	}

	if s.DefaultRole == "" {
		s.DefaultRole = authn.RoleViewer
	}
	if !authn.IsValidRole(s.DefaultRole) {
		return errors.New("invalid default role")
	}
	if s.GroupsClaim == "" {
		s.GroupsClaim = "groups"
	}

	scopes := []string{"openid"}
	for _, scope := range s.Scopes {
		for _, part := range strings.Fields(scope) {
			if part == "openid" || contains(scopes, part) {
				continue
			}
			scopes = append(scopes, part)
		}
	}
	if len(scopes) > maxScopes {
		return fmt.Errorf("at most %d scopes", maxScopes)
	}
	s.Scopes = scopes
	return nil
}

func contains(list []string, value string) bool {
	for _, v := range list {
		if v == value {
			return true
		}
	}
	return false
}
