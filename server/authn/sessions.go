package authn

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	// SessionFile holds the login sessions. It sits beside the account file
	// and is written the same way. Nothing in it is a credential: a session's
	// id is the SHA-256 of the secret its cookie carries, and the cookie is
	// signed besides.
	SessionFile = "/etc/kvm/login_sessions.json"

	// MaxSessionsPerUser and MaxSessions bound the file. A login past either
	// bound evicts the session used least recently: the user's own first, and
	// past the overall bound anyone's.
	MaxSessionsPerUser = 20
	MaxSessions        = 200

	// sessionTouchPersist is how stale the last-used time on disk may get. A
	// request updates it in memory; the file is rewritten for it at most this
	// often per session, which keeps a busy session from wearing the flash.
	sessionTouchPersist = 10 * time.Minute

	sessionFileVersion = 1
	maxSessionClient   = 64
	maxSessionIP       = 45
)

var ErrSessionNotFound = errors.New("session not found")

// LoginSession is one sign-in: the cookie a browser got from a login, until
// it expires or is revoked.
type LoginSession struct {
	// ID is the SHA-256 of the session secret in the cookie. It names the
	// session in the list and the API; the secret itself is never stored.
	ID           string    `json:"id"`
	Username     string    `json:"username"`
	TokenVersion uint64    `json:"tokenVersion"`
	CreatedAt    time.Time `json:"createdAt"`
	LastUsedAt   time.Time `json:"lastUsedAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	// IP is the peer address of the latest request. Client is a summary of
	// the user agent that logged in, such as "Firefox on Linux".
	IP     string `json:"ip"`
	Client string `json:"client"`
}

// SessionClient is what a login says about the client that made it.
type SessionClient struct {
	IP        string
	UserAgent string
}

type sessionDatabase struct {
	Version  int            `json:"version"`
	Sessions []LoginSession `json:"sessions"`
}

// SessionStore keeps the login sessions. The file is read once and memory is
// the truth after that: the server is its only writer, and every request asks
// it, so a lookup must not cost a stat.
type SessionStore struct {
	path  string
	now   func() time.Time
	mutex sync.Mutex

	loaded   bool
	sessions map[string]*LoginSession
	// persisted is each session's last-used time as the file has it.
	persisted map[string]time.Time
}

var DefaultSessions = NewSessionStore(SessionFile)

func NewSessionStore(path string) *SessionStore {
	return &SessionStore{path: path, now: time.Now}
}

// SetClock replaces the store's clock. For tests.
func (s *SessionStore) SetClock(now func() time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.now = now
}

// SessionID is the public id of the session a secret belongs to.
func SessionID(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}

// Create records a new session for an account and returns its secret, which
// goes into the cookie and nowhere else.
func (s *SessionStore) Create(username string, tokenVersion uint64, ttl time.Duration, client SessionClient) (string, LoginSession, error) {
	secret, err := randomString(32)
	if err != nil {
		return "", LoginSession{}, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	now := s.now().UTC()
	s.dropExpiredLocked(now)
	s.evictLocked(username)

	session := LoginSession{
		ID:           SessionID(secret),
		Username:     username,
		TokenVersion: tokenVersion,
		CreatedAt:    now,
		LastUsedAt:   now,
		ExpiresAt:    now.Add(ttl),
		IP:           cutString(client.IP, maxSessionIP),
		Client:       SummarizeUserAgent(client.UserAgent),
	}
	s.sessions[session.ID] = &session
	if err = s.saveLocked(); err != nil {
		delete(s.sessions, session.ID)
		return "", LoginSession{}, err
	}
	return secret, session, nil
}

// Touch looks up the session a cookie's secret names, for the account and
// token version the cookie claims, and records the request. It reports false
// for a session that was revoked, expired or never issued.
func (s *SessionStore) Touch(secret, username string, tokenVersion uint64, ip string) (LoginSession, bool) {
	if secret == "" {
		return LoginSession{}, false
	}
	id := SessionID(secret)

	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	session, ok := s.sessions[id]
	now := s.now().UTC()
	if !ok || session.Username != username || session.TokenVersion != tokenVersion || !session.ExpiresAt.After(now) {
		return LoginSession{}, false
	}
	session.LastUsedAt = now
	if ip != "" {
		session.IP = cutString(ip, maxSessionIP)
	}
	if now.Sub(s.persisted[id]) >= sessionTouchPersist {
		// A failed write leaves the newer time in memory to try again with
		// the next one; the session itself is good either way.
		_ = s.saveLocked()
	}
	return *session, true
}

// Valid reports whether a session exists and has not expired.
func (s *SessionStore) Valid(id string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	session, ok := s.sessions[id]
	return ok && session.ExpiresAt.After(s.now())
}

// Get returns a session that has not expired.
func (s *SessionStore) Get(id string) (LoginSession, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	session, ok := s.sessions[id]
	if !ok || !session.ExpiresAt.After(s.now()) {
		return LoginSession{}, false
	}
	return *session, true
}

// List returns a user's live sessions, or everyone's for an empty username,
// the most recently used first.
func (s *SessionStore) List(username string) []LoginSession {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	now := s.now()
	out := make([]LoginSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		if (username == "" || session.Username == username) && session.ExpiresAt.After(now) {
			out = append(out, *session)
		}
	}
	sortSessions(out)
	return out
}

// Revoke ends a session. The requests it has running are ended by the caller.
// The session is gone from memory even when the file cannot be written, and
// the error says it may come back after a restart.
func (s *SessionStore) Revoke(id string) (LoginSession, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	session, ok := s.sessions[id]
	if !ok {
		return LoginSession{}, ErrSessionNotFound
	}
	delete(s.sessions, id)
	return *session, s.saveLocked()
}

// RevokeUser ends every session of a user except keepID, which may be empty,
// and returns the ones it ended.
func (s *SessionStore) RevokeUser(username, keepID string) ([]LoginSession, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	var revoked []LoginSession
	for id, session := range s.sessions {
		if session.Username == username && id != keepID {
			revoked = append(revoked, *session)
			delete(s.sessions, id)
		}
	}
	if len(revoked) == 0 {
		return nil, nil
	}
	return revoked, s.saveLocked()
}

// Prune drops every session keep rejects, and the expired ones.
func (s *SessionStore) Prune(keep func(LoginSession) bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.loadLocked()
	changed := s.dropExpiredLocked(s.now())
	for id, session := range s.sessions {
		if !keep(*session) {
			delete(s.sessions, id)
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return s.saveLocked()
}

func sortSessions(sessions []LoginSession) {
	sort.Slice(sessions, func(i, j int) bool {
		if !sessions[i].LastUsedAt.Equal(sessions[j].LastUsedAt) {
			return sessions[i].LastUsedAt.After(sessions[j].LastUsedAt)
		}
		return sessions[i].ID < sessions[j].ID
	})
}

func (s *SessionStore) dropExpiredLocked(now time.Time) bool {
	changed := false
	for id, session := range s.sessions {
		if !session.ExpiresAt.After(now) {
			delete(s.sessions, id)
			changed = true
		}
	}
	return changed
}

// evictLocked makes room for one more session of username.
func (s *SessionStore) evictLocked(username string) {
	var own []LoginSession
	for _, session := range s.sessions {
		if session.Username == username {
			own = append(own, *session)
		}
	}
	sortSessions(own)
	for len(own) >= MaxSessionsPerUser {
		delete(s.sessions, own[len(own)-1].ID)
		own = own[:len(own)-1]
	}

	if len(s.sessions) < MaxSessions {
		return
	}
	all := make([]LoginSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		all = append(all, *session)
	}
	sortSessions(all)
	for _, session := range all[MaxSessions-1:] {
		delete(s.sessions, session.ID)
	}
}

// loadLocked reads the file the first time it is needed. A file that cannot
// be read or parsed grants nothing: its sessions are gone, and the next write
// replaces it.
func (s *SessionStore) loadLocked() {
	if s.loaded {
		return
	}
	s.loaded = true
	s.sessions = make(map[string]*LoginSession)
	s.persisted = make(map[string]time.Time)

	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	var db sessionDatabase
	if err = json.Unmarshal(data, &db); err != nil || db.Version != sessionFileVersion {
		return
	}
	for index := range db.Sessions {
		session := db.Sessions[index]
		if session.ID == "" || session.Username == "" {
			continue
		}
		s.sessions[session.ID] = &session
		s.persisted[session.ID] = session.LastUsedAt
	}
}

func (s *SessionStore) saveLocked() error {
	sessions := make([]LoginSession, 0, len(s.sessions))
	for _, session := range s.sessions {
		sessions = append(sessions, *session)
	}
	sortSessions(sessions)
	data, err := json.MarshalIndent(sessionDatabase{Version: sessionFileVersion, Sessions: sessions}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err = writeFileAtomic(s.path, ".sessions-*", data); err != nil {
		return err
	}
	s.persisted = make(map[string]time.Time, len(sessions))
	for _, session := range sessions {
		s.persisted[session.ID] = session.LastUsedAt
	}
	return nil
}

// writeFileAtomic writes data to path through a temporary file in the same
// directory, readable by the owner alone.
func writeFileAtomic(path, pattern string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, pattern)
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	defer func() { _ = os.Remove(temporaryPath) }()

	if err = temporary.Chmod(0o600); err == nil {
		_, err = temporary.Write(data)
	}
	if err == nil {
		err = temporary.Sync()
	}
	closeErr := temporary.Close()
	if err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}
	return os.Rename(temporaryPath, path)
}

// SummarizeUserAgent turns a User-Agent header into "Browser on System". An
// agent it does not recognise is kept as sent, cut short and stripped of
// control characters, since the client wrote it.
func SummarizeUserAgent(ua string) string {
	browser := ""
	switch {
	case strings.Contains(ua, "Edg/"), strings.Contains(ua, "EdgA/"), strings.Contains(ua, "EdgiOS/"):
		browser = "Edge"
	case strings.Contains(ua, "OPR/"):
		browser = "Opera"
	case strings.Contains(ua, "Firefox/"), strings.Contains(ua, "FxiOS/"):
		browser = "Firefox"
	case strings.Contains(ua, "Chrome/"), strings.Contains(ua, "CriOS/"):
		browser = "Chrome"
	case strings.Contains(ua, "Safari/"):
		browser = "Safari"
	}

	system := ""
	switch {
	case strings.Contains(ua, "iPhone"), strings.Contains(ua, "iPad"):
		system = "iOS"
	case strings.Contains(ua, "Android"):
		system = "Android"
	case strings.Contains(ua, "Windows"):
		system = "Windows"
	case strings.Contains(ua, "Mac OS X"), strings.Contains(ua, "Macintosh"):
		system = "macOS"
	case strings.Contains(ua, "CrOS"):
		system = "ChromeOS"
	case strings.Contains(ua, "Linux"), strings.Contains(ua, "X11"):
		system = "Linux"
	}

	switch {
	case browser != "" && system != "":
		return browser + " on " + system
	case browser != "":
		return browser
	}
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, ua)
	return cutString(strings.TrimSpace(cleaned), maxSessionClient)
}

func cutString(value string, limit int) string {
	if len(value) <= limit {
		return value
	}
	runes := []rune(value)
	for len(string(runes)) > limit {
		runes = runes[:len(runes)-1]
	}
	return string(runes)
}
