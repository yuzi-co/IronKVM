package authn

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"NanoKVM-Server/utils"

	"golang.org/x/crypto/bcrypt"
)

const (
	AccountFile        = "/etc/kvm/pwd"
	currentFileVersion = 1
	defaultUsername    = "admin"
	defaultPassword    = "admin"
)

type Role string

const (
	RoleAdmin Role = "admin"
	RoleUser  Role = "user"
	// RoleViewer sees the video and nothing else: no input, no power, no media,
	// no settings. middleware/viewer_policy.go is the list of what it reaches.
	RoleViewer Role = "viewer"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserExists       = errors.New("username already exists")
	ErrLastAdmin        = errors.New("at least one enabled admin is required")
	ErrSelfModification = errors.New("administrators cannot disable or demote themselves")
	ErrSelfDelete       = errors.New("administrators cannot delete themselves")
	ErrSystemAccount    = errors.New("the device owner account must remain an enabled administrator")
	// ErrOIDCSubjectMismatch is a single sign-on to an account that is linked
	// to another account at the same identity provider.
	ErrOIDCSubjectMismatch = errors.New("the account is linked to another identity provider account")
	// ErrOIDCNotLinked is a single sign-on to an account that is not linked
	// to any identity provider account yet. An administrator links it, or the
	// user confirms the link with the account's password.
	ErrOIDCNotLinked = errors.New("the account is not linked to an identity provider account")

	usernamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,31}$`)
)

type User struct {
	Username           string `json:"username"`
	PasswordHash       string `json:"password"`
	Role               Role   `json:"role"`
	Enabled            bool   `json:"enabled"`
	TokenVersion       uint64 `json:"tokenVersion"`
	MustChangePassword bool   `json:"mustChangePassword,omitempty"`
	SystemAccount      bool   `json:"systemAccount,omitempty"`
	// OnDiskViewOnly is how a viewer is written to the account file, and it is
	// never set on a user in memory. See toDisk.
	OnDiskViewOnly bool `json:"viewOnly,omitempty"`
	// IPMIPassword is the account's IPMI password, sealed by service/ipmi.
	// IPMI's RAKP handshake needs the password itself, which the bcrypt hash
	// above cannot give back, so it is a separate password. Empty means the
	// account cannot log in over IPMI.
	IPMIPassword string `json:"ipmiPassword,omitempty"`
	// TOTPSecret is the account's TOTP seed, sealed by a Keyring. Empty means
	// the web login asks for the password alone.
	TOTPSecret string `json:"totpSecret,omitempty"`
	// RecoveryCodes are the SHA-256 hashes of the recovery codes that have not
	// been used yet. Each one stands in for a TOTP code once.
	RecoveryCodes []string `json:"recoveryCodes,omitempty"`
	// OIDCIssuer and OIDCSubject link the account to one account at an
	// OpenID Connect provider: the issuer and the sub claim of its ID token.
	// An administrator sets them, or the user does by confirming a first
	// single sign-on with the account's password, or a single sign-on that
	// creates the account. Every single sign-on must carry the same sub, so
	// an IdP account that renames itself into this username does not get in.
	// Empty means not linked yet, and a single sign-on is refused. Older
	// servers ignore both fields and drop them when they write the file,
	// which leaves the account to be linked again.
	OIDCIssuer  string `json:"oidcIssuer,omitempty"`
	OIDCSubject string `json:"oidcSubject,omitempty"`
	// NoPassword marks an account made by a single sign-on, whose password
	// is random and known to nobody, until a password is set. Such an
	// account cannot confirm a link with its password; only an administrator
	// can link it again.
	NoPassword bool `json:"noPassword,omitempty"`
}

// TOTPEnabled reports whether the web login asks this account for a second
// factor.
func (u *User) TOTPEnabled() bool {
	return u.TOTPSecret != ""
}

// OIDCLinked reports whether a single sign-on has linked the account to an
// account at an identity provider.
func (u *User) OIDCLinked() bool {
	return u.OIDCSubject != ""
}

type UserInfo struct {
	Username      string `json:"username"`
	Role          Role   `json:"role"`
	Enabled       bool   `json:"enabled"`
	SystemAccount bool   `json:"systemAccount,omitempty"`
	TOTPEnabled   bool   `json:"totpEnabled,omitempty"`
	OIDCLinked    bool   `json:"oidcLinked,omitempty"`
	// OIDCIssuer is the issuer the link was made under.
	OIDCIssuer string `json:"-"`
}

type UserPatch struct {
	Role    *Role
	Enabled *bool
}

type database struct {
	Version        int    `json:"version"`
	Users          []User `json:"users"`
	LegacyUsername string `json:"username,omitempty"`
	LegacyPassword string `json:"password,omitempty"`
}

type legacyAccount struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Store struct {
	path  string
	mutex sync.RWMutex
	// cache is the last account file parsed, with the identity of the file it
	// came from. Every authenticated request and every open session's recheck
	// reads the store, so it pays a stat instead of a read, a parse and a
	// validation when the file has not changed. See readLocked.
	cache atomic.Pointer[cachedDatabase]
}

var DefaultStore = NewStore(AccountFile)

func NewStore(path string) *Store {
	return &Store{path: path}
}

func IsValidRole(role Role) bool {
	return role == RoleAdmin || role == RoleUser || role == RoleViewer
}

func ValidateUsername(username string) error {
	if !usernamePattern.MatchString(username) {
		return errors.New("username must be 1-32 characters and contain only letters, numbers, '.', '_' or '-'")
	}
	return nil
}

func ValidatePassword(password string) error {
	length := len([]byte(password))
	if length < 8 || length > 72 {
		return errors.New("password must be between 8 and 72 bytes")
	}
	return nil
}

func (s *Store) List() ([]UserInfo, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	db, err := s.loadLocked(false)
	if err != nil {
		return nil, err
	}

	users := make([]UserInfo, 0, len(db.Users))
	for _, user := range db.Users {
		users = append(users, UserInfo{
			Username:      user.Username,
			Role:          user.Role,
			Enabled:       user.Enabled,
			SystemAccount: user.SystemAccount,
			TOTPEnabled:   user.TOTPEnabled(),
			OIDCLinked:    user.OIDCLinked(),
			OIDCIssuer:    user.OIDCIssuer,
		})
	}
	return users, nil
}

func (s *Store) Get(username string) (*User, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	db, err := s.loadLocked(false)
	if err != nil {
		return nil, err
	}
	return findUser(db.Users, username)
}

// compareHash is bcrypt's password check. Tests replace it to observe when
// it runs.
var compareHash = bcrypt.CompareHashAndPassword

func (s *Store) Authenticate(username, password string) (*User, bool, error) {
	user, err := s.userForLogin(username)
	if err != nil {
		return nil, false, err
	}
	if user == nil || !user.Enabled {
		return nil, false, nil
	}

	// The password check runs on a copy of the account, outside the lock, so
	// a login does not stall every other reader of the store for the length
	// of a bcrypt.
	if compareHash([]byte(user.PasswordHash), []byte(password)) == nil {
		return user, true, nil
	}

	// Older releases stored a reversibly encrypted password. Upgrade it after
	// the first successful login so the compatibility path is not permanent.
	legacyPassword, decodeErr := utils.DecodeDecrypt(user.PasswordHash)
	if decodeErr != nil || legacyPassword != password {
		return nil, false, nil
	}

	hash, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if hashErr != nil {
		return nil, false, hashErr
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	db, err := s.loadLocked(true)
	if err != nil {
		return nil, false, err
	}
	index := userIndex(db.Users, username)
	// The account changed while the password was checked: the password that
	// matched may no longer be its own.
	if index < 0 || !db.Users[index].Enabled || db.Users[index].PasswordHash != user.PasswordHash {
		return nil, false, nil
	}
	db.Users[index].PasswordHash = string(hash)
	if db.Users[index].TokenVersion == 0 {
		db.Users[index].TokenVersion = 1
	}
	user = cloneUser(&db.Users[index])
	if err = s.saveLocked(db); err != nil {
		return nil, false, err
	}
	return user, true, nil
}

// userForLogin returns a copy of the account, or nil when there is none. A
// current account file is read under the read lock. A missing or legacy one
// is migrated and written first, under the write lock, as before.
func (s *Store) userForLogin(username string) (*User, error) {
	s.mutex.RLock()
	db, current, err := s.readCurrentLocked()
	s.mutex.RUnlock()
	if err != nil {
		return nil, err
	}

	if !current {
		s.mutex.Lock()
		db, err = s.loadLocked(true)
		s.mutex.Unlock()
		if err != nil {
			return nil, err
		}
	}

	user, err := findUser(db.Users, username)
	if err != nil {
		return nil, nil
	}
	return user, nil
}

// readCurrentLocked reads an account file already in the current format.
// current is false when the file is missing or in a legacy format, which
// loadLocked has to migrate.
func (s *Store) readCurrentLocked() (*database, bool, error) {
	db, _, err := s.readLocked()
	if err != nil || db == nil {
		return nil, false, err
	}
	return db, true, nil
}

// readLocked returns a copy of the account file's database when the file is
// in the current format. Otherwise db is nil and data is the file's raw
// content for loadLocked to migrate, or nil when there is no file.
//
// A file that parsed before and still has the same identity (size,
// modification time, and on Linux inode, device and change time) is not read
// again. A shell edit or a reset that removes the file changes that identity,
// so it is still seen on the next call. A file modified in the last few
// seconds is never cached, because two writes inside one tick of the
// filesystem's clock can leave the same size and time behind.
func (s *Store) readLocked() (*database, []byte, error) {
	info, err := os.Stat(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.cache.Store(nil)
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if cached := s.cache.Load(); cached != nil && cached.stamp.equal(stampOf(info)) {
		return cached.db.clone(), nil, nil
	}

	data, info, err := readAccountFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		s.cache.Store(nil)
		return nil, nil, nil
	}
	if err != nil {
		return nil, nil, err
	}

	var db database
	if err = json.Unmarshal(data, &db); err != nil || db.Version == 0 {
		s.cache.Store(nil)
		return nil, data, nil
	}
	fromDisk(&db)
	if err = validateDatabase(&db); err != nil {
		s.cache.Store(nil)
		return nil, nil, err
	}
	stamp := stampOf(info)
	if stamp.settled(time.Now()) {
		s.cache.Store(&cachedDatabase{stamp: stamp, db: db.clone()})
	} else {
		s.cache.Store(nil)
	}
	return &db, nil, nil
}

func (s *Store) ValidateToken(username string, tokenVersion uint64) (*User, error) {
	user, err := s.Get(username)
	if err != nil {
		return nil, err
	}
	if !user.Enabled || user.TokenVersion != tokenVersion {
		return nil, errors.New("session revoked")
	}
	return user, nil
}

func (s *Store) Create(username, password string, role Role) error {
	return s.create(User{Username: username, Role: role}, password)
}

// CreateOIDC creates an account for a single sign-on, already linked to the
// identity provider account it was made for, so that no other login can link
// it first.
func (s *Store) CreateOIDC(username, password string, role Role, issuer, subject string) error {
	if issuer == "" || subject == "" {
		return errors.New("an OIDC account needs an issuer and a subject")
	}
	return s.create(User{Username: username, Role: role, OIDCIssuer: issuer, OIDCSubject: subject, NoPassword: true}, password)
}

// create adds the account template describes, with the password hashed, the
// account enabled and a fresh token version.
func (s *Store) create(template User, password string) error {
	username, role := template.Username, template.Role
	if err := ValidateUsername(username); err != nil {
		return err
	}
	if err := ValidatePassword(password); err != nil {
		return err
	}
	if !IsValidRole(role) {
		return errors.New("invalid role")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	tokenVersion, err := newTokenVersion()
	if err != nil {
		return err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	db, err := s.loadLocked(true)
	if err != nil {
		return err
	}
	if _, err = findUser(db.Users, username); err == nil {
		return ErrUserExists
	}
	user := template
	user.PasswordHash = string(hash)
	user.Enabled = true
	user.TokenVersion = tokenVersion
	db.Users = append(db.Users, user)
	return s.saveLocked(db)
}

func (s *Store) Update(actor, username string, patch UserPatch) (*User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	db, err := s.loadLocked(true)
	if err != nil {
		return nil, err
	}
	index := userIndex(db.Users, username)
	if index < 0 {
		return nil, ErrUserNotFound
	}
	user := db.Users[index]

	if patch.Role != nil && !IsValidRole(*patch.Role) {
		return nil, errors.New("invalid role")
	}
	if actor == username && user.Role == RoleAdmin &&
		((patch.Role != nil && *patch.Role != RoleAdmin) || (patch.Enabled != nil && !*patch.Enabled)) {
		return nil, ErrSelfModification
	}
	if user.SystemAccount &&
		((patch.Role != nil && *patch.Role != RoleAdmin) || (patch.Enabled != nil && !*patch.Enabled)) {
		return nil, ErrSystemAccount
	}

	changed := false
	if patch.Role != nil && user.Role != *patch.Role {
		user.Role = *patch.Role
		changed = true
	}
	if patch.Enabled != nil && user.Enabled != *patch.Enabled {
		user.Enabled = *patch.Enabled
		changed = true
	}
	if !changed {
		return &user, nil
	}

	db.Users[index] = user
	if enabledAdminCount(db.Users) == 0 {
		return nil, ErrLastAdmin
	}
	db.Users[index].TokenVersion++
	if err = s.saveLocked(db); err != nil {
		return nil, err
	}
	return cloneUser(&db.Users[index]), nil
}

func (s *Store) Delete(actor, username string) error {
	if actor == username {
		return ErrSelfDelete
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	db, err := s.loadLocked(true)
	if err != nil {
		return err
	}
	index := userIndex(db.Users, username)
	if index < 0 {
		return ErrUserNotFound
	}
	if db.Users[index].SystemAccount {
		return ErrSystemAccount
	}
	users := append([]User(nil), db.Users[:index]...)
	users = append(users, db.Users[index+1:]...)
	if enabledAdminCount(users) == 0 {
		return ErrLastAdmin
	}
	db.Users = users
	return s.saveLocked(db)
}

func (s *Store) SetPassword(username, password string) (*User, error) {
	return s.SetPasswordAndRun(username, password, nil)
}

// SetPasswordAndRun commits the web password before running the optional
// system-side update. If that update fails, the account record is rolled back
// while the store lock is still held.
func (s *Store) SetPasswordAndRun(username, password string, afterCommit func() error) (*User, error) {
	if err := ValidatePassword(password); err != nil {
		return nil, err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()
	db, err := s.loadLocked(true)
	if err != nil {
		return nil, err
	}
	index := userIndex(db.Users, username)
	if index < 0 {
		return nil, ErrUserNotFound
	}
	previous := *db
	previous.Users = append([]User(nil), db.Users...)
	db.Users[index].PasswordHash = string(hash)
	db.Users[index].MustChangePassword = false
	db.Users[index].NoPassword = false
	db.Users[index].TokenVersion++
	if err = s.saveLocked(db); err != nil {
		return nil, err
	}
	if afterCommit != nil {
		if err = afterCommit(); err != nil {
			if rollbackErr := s.saveLocked(&previous); rollbackErr != nil {
				return nil, fmt.Errorf("system password update failed: %v; account rollback failed: %w", err, rollbackErr)
			}
			return nil, err
		}
	}
	return cloneUser(&db.Users[index]), nil
}

// SetIPMIPassword stores an account's sealed IPMI password, or removes it
// when sealed is empty. The store does not look inside the value. Web
// sessions are not revoked: the web password is unchanged.
func (s *Store) SetIPMIPassword(username, sealed string) (*User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	db, err := s.loadLocked(true)
	if err != nil {
		return nil, err
	}
	index := userIndex(db.Users, username)
	if index < 0 {
		return nil, ErrUserNotFound
	}
	db.Users[index].IPMIPassword = sealed
	if err = s.saveLocked(db); err != nil {
		return nil, err
	}
	return cloneUser(&db.Users[index]), nil
}

// SetTOTP turns on the second login step for an account, or replaces its
// secret and recovery codes. sealed is the secret as a Keyring sealed it;
// recoveryHashes are the hashes of the recovery codes. Sessions are not
// revoked: the password is unchanged, and the session that enrols is the
// one that proved it.
func (s *Store) SetTOTP(username, sealed string, recoveryHashes []string) (*User, error) {
	if sealed == "" {
		return nil, errors.New("empty TOTP secret")
	}
	return s.updateUser(username, func(user *User) bool {
		user.TOTPSecret = sealed
		user.RecoveryCodes = append([]string(nil), recoveryHashes...)
		return true
	})
}

// SetRecoveryCodes replaces the recovery codes of an account that has TOTP
// on.
func (s *Store) SetRecoveryCodes(username string, recoveryHashes []string) (*User, error) {
	return s.updateUser(username, func(user *User) bool {
		if !user.TOTPEnabled() {
			return false
		}
		user.RecoveryCodes = append([]string(nil), recoveryHashes...)
		return true
	})
}

// ClearTOTP turns the second login step off for an account and forgets its
// secret and recovery codes.
func (s *Store) ClearTOTP(username string) (*User, error) {
	return s.updateUser(username, func(user *User) bool {
		changed := user.TOTPSecret != "" || len(user.RecoveryCodes) > 0
		user.TOTPSecret = ""
		user.RecoveryCodes = nil
		return changed
	})
}

// ConsumeRecoveryCode removes one recovery code, by its hash, and reports
// whether the account had it. The check and the removal happen under one
// lock, so a code cannot be spent twice by two requests at once.
func (s *Store) ConsumeRecoveryCode(username, hash string) (bool, error) {
	found := false
	_, err := s.updateUser(username, func(user *User) bool {
		if !user.TOTPEnabled() {
			return false
		}
		kept := make([]string, 0, len(user.RecoveryCodes))
		for _, stored := range user.RecoveryCodes {
			if !found && subtle.ConstantTimeCompare([]byte(stored), []byte(hash)) == 1 {
				found = true
				continue
			}
			kept = append(kept, stored)
		}
		user.RecoveryCodes = kept
		return found
	})
	if err != nil {
		return false, err
	}
	return found, nil
}

// updateUser applies change to an account under the write lock, and saves
// when change reports that it changed something.
func (s *Store) updateUser(username string, change func(*User) bool) (*User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	db, err := s.loadLocked(true)
	if err != nil {
		return nil, err
	}
	index := userIndex(db.Users, username)
	if index < 0 {
		return nil, ErrUserNotFound
	}
	if !change(&db.Users[index]) {
		return cloneUser(&db.Users[index]), nil
	}
	if err = s.saveLocked(db); err != nil {
		return nil, err
	}
	return cloneUser(&db.Users[index]), nil
}

// BindOIDC checks the link of an account a single sign-on is for, and
// returns the account when the link is this identity. It never changes the
// link.
//
// An account with no link is ErrOIDCNotLinked: a single sign-on alone never
// links an existing account (ironkvm-dist#169). See LinkOIDC and ConfirmOIDC.
//
// So is an account linked under another issuer. A sub means nothing outside
// its issuer, and whoever runs the identity provider the settings now name
// could otherwise take any account by its name. After an issuer change each
// account is linked again by an administrator or by its own password.
//
// A link under the same issuer with another subject is ErrOIDCSubjectMismatch.
func (s *Store) BindOIDC(username, issuer, subject string) (*User, error) {
	if issuer == "" || subject == "" {
		return nil, errors.New("an OIDC link needs an issuer and a subject")
	}
	user, err := s.Get(username)
	if err != nil {
		return nil, err
	}
	switch {
	case user.OIDCSubject == "" || user.OIDCIssuer != issuer:
		return nil, ErrOIDCNotLinked
	case user.OIDCSubject != subject:
		return nil, ErrOIDCSubjectMismatch
	}
	return user, nil
}

// ConfirmOIDC links an account that has no link, or one under another issuer,
// for a user who confirmed a single sign-on with the account's password. The
// same link again is no change. Another subject under the same issuer, which
// an administrator may have linked meanwhile, is ErrOIDCSubjectMismatch, and
// the account is left as it is.
func (s *Store) ConfirmOIDC(username, issuer, subject string) (*User, error) {
	if issuer == "" || subject == "" {
		return nil, errors.New("an OIDC link needs an issuer and a subject")
	}
	mismatch := false
	user, err := s.updateUser(username, func(user *User) bool {
		if user.OIDCSubject != "" && user.OIDCIssuer == issuer {
			mismatch = user.OIDCSubject != subject
			return false
		}
		user.OIDCIssuer = issuer
		user.OIDCSubject = subject
		return true
	})
	if err != nil {
		return nil, err
	}
	if mismatch {
		return nil, ErrOIDCSubjectMismatch
	}
	return user, nil
}

// LinkOIDC is an administrator linking an account to an identity provider
// account before its first single sign-on, or to another one. It replaces
// whatever link the account had, and returns that link's issuer and subject.
func (s *Store) LinkOIDC(username, issuer, subject string) (*User, string, string, error) {
	if issuer == "" || subject == "" {
		return nil, "", "", errors.New("an OIDC link needs an issuer and a subject")
	}
	var previousIssuer, previousSubject string
	user, err := s.updateUser(username, func(user *User) bool {
		previousIssuer, previousSubject = user.OIDCIssuer, user.OIDCSubject
		if user.OIDCIssuer == issuer && user.OIDCSubject == subject {
			return false
		}
		user.OIDCIssuer = issuer
		user.OIDCSubject = subject
		return true
	})
	if err != nil {
		return nil, "", "", err
	}
	return user, previousIssuer, previousSubject, nil
}

// ClearOIDC removes an account's link to an identity provider account, for a
// user who was recreated at the provider. The account is then linked again
// like one that never had a link.
func (s *Store) ClearOIDC(username string) (*User, error) {
	return s.updateUser(username, func(user *User) bool {
		changed := user.OIDCIssuer != "" || user.OIDCSubject != ""
		user.OIDCIssuer = ""
		user.OIDCSubject = ""
		return changed
	})
}

func (s *Store) Revoke(username string) (*User, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	db, err := s.loadLocked(true)
	if err != nil {
		return nil, err
	}
	index := userIndex(db.Users, username)
	if index < 0 {
		return nil, ErrUserNotFound
	}
	db.Users[index].TokenVersion++
	if err = s.saveLocked(db); err != nil {
		return nil, err
	}
	return cloneUser(&db.Users[index]), nil
}

func (s *Store) loadLocked(migrate bool) (*database, error) {
	current, data, err := s.readLocked()
	if err != nil {
		return nil, err
	}
	if current != nil {
		return current, nil
	}
	if data == nil {
		db, defaultErr := defaultDatabase()
		if defaultErr != nil {
			return nil, defaultErr
		}
		if migrate {
			if saveErr := s.saveLocked(db); saveErr != nil {
				return nil, saveErr
			}
		}
		return db, nil
	}

	var db database
	var legacy legacyAccount
	if err = json.Unmarshal(data, &legacy); err != nil || legacy.Username == "" || legacy.Password == "" {
		return nil, errors.New("invalid account file")
	}
	if err = ValidateUsername(legacy.Username); err != nil {
		return nil, fmt.Errorf("invalid legacy username: %w", err)
	}
	tokenVersion, err := newTokenVersion()
	if err != nil {
		return nil, err
	}
	db = database{
		Version: currentFileVersion,
		Users: []User{{
			Username:           legacy.Username,
			PasswordHash:       legacy.Password,
			Role:               RoleAdmin,
			Enabled:            true,
			TokenVersion:       tokenVersion,
			MustChangePassword: passwordMatches(legacy.Password, defaultPassword),
			SystemAccount:      true,
		}},
	}
	if migrate {
		if err = s.saveLocked(&db); err != nil {
			return nil, err
		}
	}
	return &db, nil
}

func (s *Store) saveLocked(db *database) error {
	// The next read parses what was written rather than trusting the file's
	// new identity to differ from the cached one.
	defer s.cache.Store(nil)
	if err := validateDatabase(db); err != nil {
		return err
	}
	syncLegacyAccount(db)
	data, err := json.MarshalIndent(toDisk(db), "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	directory := filepath.Dir(s.path)
	if err = os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	temporary, err := os.CreateTemp(directory, ".pwd-*")
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
	if err = os.Rename(temporaryPath, s.path); err != nil {
		return err
	}

	dir, err := os.Open(directory)
	if err != nil {
		return err
	}
	defer func() { _ = dir.Close() }()
	return dir.Sync()
}

// defaultPasswordHash is derived once. loadLocked builds the default database
// on every read while the account file is absent, and bcrypt at the default
// cost is about a second of a 1GHz C906 - paid on the request path, and on the
// exact path someone guessing passwords would be hammering. The hash is only
// kept in memory: the file gets whichever one was current when it was written.
var defaultPasswordHash = sync.OnceValues(func() (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), bcrypt.DefaultCost)
	return string(hash), err
})

func defaultDatabase() (*database, error) {
	hash, err := defaultPasswordHash()
	if err != nil {
		return nil, err
	}
	tokenVersion, err := newTokenVersion()
	if err != nil {
		return nil, err
	}
	return &database{
		Version: currentFileVersion,
		Users: []User{{
			Username:           defaultUsername,
			PasswordHash:       hash,
			Role:               RoleAdmin,
			Enabled:            true,
			TokenVersion:       tokenVersion,
			MustChangePassword: true,
			SystemAccount:      true,
		}},
	}, nil
}

func newTokenVersion() (uint64, error) {
	var data [8]byte
	for {
		if _, err := rand.Read(data[:]); err != nil {
			return 0, err
		}
		version := binary.LittleEndian.Uint64(data[:])
		if version != 0 && version != ^uint64(0) {
			return version, nil
		}
	}
}

func validateDatabase(db *database) error {
	if db.Version != currentFileVersion {
		return fmt.Errorf("unsupported account file version: %d", db.Version)
	}
	if len(db.Users) == 0 {
		return errors.New("account file contains no users")
	}
	seen := make(map[string]struct{}, len(db.Users))
	systemAccounts := 0
	for index := range db.Users {
		user := &db.Users[index]
		if err := ValidateUsername(user.Username); err != nil {
			return err
		}
		if user.PasswordHash == "" || !IsValidRole(user.Role) {
			return errors.New("account file contains an invalid user")
		}
		if _, exists := seen[user.Username]; exists {
			return errors.New("account file contains duplicate usernames")
		}
		seen[user.Username] = struct{}{}
		if user.TokenVersion == 0 {
			user.TokenVersion = 1
		}
		if user.SystemAccount {
			systemAccounts++
			if user.Role != RoleAdmin || !user.Enabled || systemAccounts > 1 {
				return ErrSystemAccount
			}
		}
	}
	if enabledAdminCount(db.Users) == 0 {
		return ErrLastAdmin
	}
	return nil
}

func findUser(users []User, username string) (*User, error) {
	index := userIndex(users, username)
	if index < 0 {
		return nil, ErrUserNotFound
	}
	return cloneUser(&users[index]), nil
}

func userIndex(users []User, username string) int {
	for index := range users {
		if users[index].Username == username {
			return index
		}
	}
	return -1
}

func enabledAdminCount(users []User) int {
	count := 0
	for _, user := range users {
		if user.Role == RoleAdmin && user.Enabled {
			count++
		}
	}
	return count
}

func cloneUser(user *User) *User {
	if user == nil {
		return nil
	}
	copy := *user
	copy.RecoveryCodes = append([]string(nil), user.RecoveryCodes...)
	return &copy
}

func passwordMatches(hash, password string) bool {
	if bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil {
		return true
	}
	legacyPassword, err := utils.DecodeDecrypt(hash)
	return err == nil && legacyPassword == password
}

// Keep a legacy top-level account mirror so downgrading to a single-user
// server still authenticates the original device owner instead of failing
// open or forcing an immediate physical reset.
func syncLegacyAccount(db *database) {
	db.LegacyUsername = ""
	db.LegacyPassword = ""
	var fallback *User
	for _, user := range db.Users {
		if user.SystemAccount {
			db.LegacyUsername = user.Username
			db.LegacyPassword = user.PasswordHash
			return
		}
		if fallback == nil && user.Role == RoleAdmin && user.Enabled {
			copy := user
			fallback = &copy
		}
	}
	if fallback != nil {
		db.LegacyUsername = fallback.Username
		db.LegacyPassword = fallback.PasswordHash
	}
}

// viewerHashPrefix marks the password hash of a viewer on disk.
const viewerHashPrefix = "viewer$"

// toDisk returns the database as it is written. A server that predates the
// viewer role rejects the whole account file over a role it does not know, and
// slot A may run such a server against the same /etc/kvm. So a viewer is
// written as an ordinary user with viewOnly set and a password hash that is
// not a hash. Such a server loads the file and refuses the viewer's login,
// which fails closed, and keeps every other account working.
func toDisk(db *database) *database {
	out := *db
	out.Users = make([]User, len(db.Users))
	for index, user := range db.Users {
		if user.Role == RoleViewer {
			user.Role = RoleUser
			user.OnDiskViewOnly = true
			user.PasswordHash = viewerHashPrefix + user.PasswordHash
		}
		out.Users[index] = user
	}
	return &out
}

// fromDisk undoes toDisk on a database just read.
func fromDisk(db *database) {
	for index := range db.Users {
		user := &db.Users[index]
		if !user.OnDiskViewOnly {
			continue
		}
		user.OnDiskViewOnly = false
		user.Role = RoleViewer
		user.PasswordHash = strings.TrimPrefix(user.PasswordHash, viewerHashPrefix)
	}
}
