package authn

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"NanoKVM-Server/utils/atomicfile"
)

const (
	// ShareFile holds the view-only share links. It sits beside the account
	// file and is written the same way. A server without share links ignores
	// it, and nothing in it is a credential: only the hash of each token.
	ShareFile = "/etc/kvm/share_links.json"

	// MinShareTTL and MaxShareTTL bound how long a link may live.
	MinShareTTL = time.Minute
	MaxShareTTL = 30 * 24 * time.Hour

	// MaxShareLinks bounds the links that exist at once.
	MaxShareLinks = 50

	// SharePrincipalPrefix starts the username of a session that came from a
	// share link. The account pattern has no colon, so no account can have it.
	SharePrincipalPrefix = "share:"

	shareFileVersion = 1
	maxShareLabel    = 64
)

var (
	ErrShareNotFound = errors.New("share link not found")
	ErrShareTTL      = errors.New("share link lifetime is out of range")
	ErrShareLimit    = errors.New("too many share links")
)

// ShareLink is one view-only link. The token itself is shown once, when the
// link is made, and only its SHA-256 is kept.
type ShareLink struct {
	ID        string    `json:"id"`
	Label     string    `json:"label"`
	TokenHash string    `json:"tokenHash"`
	CreatedBy string    `json:"createdBy"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type shareDatabase struct {
	Version int         `json:"version"`
	Links   []ShareLink `json:"links"`
}

// ShareStore keeps the share links in a file.
type ShareStore struct {
	path  string
	now   func() time.Time
	mutex sync.Mutex

	// loaded is the file as last read, with the identity it had. Every request
	// made with a share session asks whether its link is still good, and the
	// session recheck asks every few seconds, so an unchanged file costs a stat.
	loaded    []ShareLink
	loadedFor fileIdentity
	haveLoad  bool
}

type fileIdentity struct {
	size    int64
	modTime time.Time
}

var DefaultShares = NewShareStore(ShareFile)

func NewShareStore(path string) *ShareStore {
	return &ShareStore{path: path, now: time.Now}
}

// SetClock replaces the store's clock. For tests.
func (s *ShareStore) SetClock(now func() time.Time) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.now = now
	s.haveLoad = false
}

func hashShareToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func randomString(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// Create makes a link that works for ttl. The token it returns is the only
// time it exists in the clear.
func (s *ShareStore) Create(label, createdBy string, ttl time.Duration) (*ShareLink, string, error) {
	if ttl < MinShareTTL || ttl > MaxShareTTL {
		return nil, "", ErrShareTTL
	}
	if runes := []rune(label); len(runes) > maxShareLabel {
		label = string(runes[:maxShareLabel])
	}
	token, err := randomString(32)
	if err != nil {
		return nil, "", err
	}
	idBytes := make([]byte, 8)
	if _, err = rand.Read(idBytes); err != nil {
		return nil, "", err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	links, err := s.loadLocked()
	if err != nil {
		return nil, "", err
	}
	now := s.now()
	links = liveLinks(links, now)
	if len(links) >= MaxShareLinks {
		return nil, "", ErrShareLimit
	}
	link := ShareLink{
		ID:        hex.EncodeToString(idBytes),
		Label:     label,
		TokenHash: hashShareToken(token),
		CreatedBy: createdBy,
		CreatedAt: now.UTC(),
		ExpiresAt: now.Add(ttl).UTC(),
	}
	links = append(links, link)
	if err = s.saveLocked(links); err != nil {
		return nil, "", err
	}
	return &link, token, nil
}

// List returns the links that have not expired, soonest to expire first.
func (s *ShareStore) List() ([]ShareLink, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	links, err := s.loadLocked()
	if err != nil {
		return nil, err
	}
	links = liveLinks(links, s.now())
	sort.Slice(links, func(i, j int) bool { return links[i].ExpiresAt.Before(links[j].ExpiresAt) })
	return links, nil
}

// Revoke ends a link at once. The sessions it opened are ended by the caller.
func (s *ShareStore) Revoke(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	links, err := s.loadLocked()
	if err != nil {
		return err
	}
	kept := make([]ShareLink, 0, len(links))
	found := false
	for _, link := range links {
		if link.ID == id {
			found = true
			continue
		}
		kept = append(kept, link)
	}
	if !found {
		return ErrShareNotFound
	}
	return s.saveLocked(liveLinks(kept, s.now()))
}

// Redeem finds the link a token belongs to. Every stored hash is compared, in
// constant time, so the time taken does not say how close a guess was.
func (s *ShareStore) Redeem(token string) (*ShareLink, bool) {
	if token == "" {
		return nil, false
	}
	want := []byte(hashShareToken(token))

	s.mutex.Lock()
	defer s.mutex.Unlock()
	links, err := s.loadLocked()
	if err != nil {
		return nil, false
	}
	now := s.now()
	var match *ShareLink
	for index := range links {
		equal := subtle.ConstantTimeCompare(want, []byte(links[index].TokenHash)) == 1
		if equal && links[index].ExpiresAt.After(now) {
			link := links[index]
			match = &link
		}
	}
	return match, match != nil
}

// Valid reports whether a link still exists and has not expired.
func (s *ShareStore) Valid(id string) (*ShareLink, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	links, err := s.loadLocked()
	if err != nil {
		return nil, false
	}
	now := s.now()
	for _, link := range links {
		if link.ID == id && link.ExpiresAt.After(now) {
			found := link
			return &found, true
		}
	}
	return nil, false
}

func liveLinks(links []ShareLink, now time.Time) []ShareLink {
	out := make([]ShareLink, 0, len(links))
	for _, link := range links {
		if link.ExpiresAt.After(now) {
			out = append(out, link)
		}
	}
	return out
}

func (s *ShareStore) loadLocked() ([]ShareLink, error) {
	info, err := os.Stat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.haveLoad = false
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	identity := fileIdentity{size: info.Size(), modTime: info.ModTime()}
	// A file written a moment ago is read again: two writes inside one tick of
	// the filesystem's clock can leave the same size and time behind.
	if s.haveLoad && s.loadedFor == identity && s.now().Sub(identity.modTime) > 5*time.Second {
		return append([]ShareLink(nil), s.loaded...), nil
	}
	data, err := os.ReadFile(s.path)
	if err != nil {
		return nil, err
	}
	var db shareDatabase
	if err = json.Unmarshal(data, &db); err != nil || db.Version != shareFileVersion {
		// An unreadable file grants nothing: no link in it can be redeemed.
		s.haveLoad = false
		return nil, errors.New("invalid share link file")
	}
	s.loaded = db.Links
	s.loadedFor = identity
	s.haveLoad = true
	return append([]ShareLink(nil), db.Links...), nil
}

func (s *ShareStore) saveLocked(links []ShareLink) error {
	s.haveLoad = false
	data, err := json.MarshalIndent(shareDatabase{Version: shareFileVersion, Links: links}, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	directory := filepath.Dir(s.path)
	if err = os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	return atomicfile.WriteFile(s.path, data, 0o600)
}
