package auth

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"

	"NanoKVM-Server/authn/totp"
)

// The limits of the second login step. A code is six digits and a guess is
// right for three of the 10^6 values at once (one step either side of now),
// so the step is safe only while guesses are few:
//
//   - a challenge, the login a correct password started, takes at most
//     challengeAttempts codes and lives for challengeTTL;
//   - an account takes at most userMaxFailures wrong codes in a row, from any
//     address and over any number of challenges, before the code step is
//     closed to it for userLockout;
//   - every wrong code also counts against the address in the password
//     limiter (brute_force.go), so one address cannot work through accounts.
//
// Five wrong codes per five minutes is about 230 days of guessing for an
// even chance, and every guess needs the password first.
const (
	challengeTTL      = 5 * time.Minute
	challengeAttempts = 5
	maxChallenges     = 256

	userMaxFailures = 5
	userLockout     = 5 * time.Minute
	// userFailureWindow is how long a wrong code is remembered when no
	// lockout followed it.
	userFailureWindow = 15 * time.Minute

	enrolmentTTL      = 10 * time.Minute
	enrolmentAttempts = 5
)

// challenge is a login whose password was right and whose code is still
// owed. tokenVersion pins it to the account as it was: a password change or a
// revocation in between invalidates it.
type challenge struct {
	username     string
	tokenVersion uint64
	expires      time.Time
	attempts     int
}

// enrolment is a secret that has been shown to the user but is not in force
// until a code from it is typed back.
type enrolment struct {
	secret   []byte
	expires  time.Time
	attempts int
}

type userFailures struct {
	count      int
	last       time.Time
	lockedTill time.Time
}

// totpGuard holds the state of the second factor that lives in memory only:
// half finished logins, enrolments in progress, the per-account limiter and
// the last step each account used. A restart drops all of it, which costs a
// user a retyped password and nothing else; the one weakening is that a code
// accepted just before a restart could be replayed in the 90 seconds after.
type totpGuard struct {
	mu          sync.Mutex
	now         func() time.Time
	challenges  map[string]*challenge
	enrolments  map[string]*enrolment
	failures    map[string]*userFailures
	lastCounter map[string]uint64
}

func newTOTPGuard() *totpGuard {
	return &totpGuard{
		now:         time.Now,
		challenges:  make(map[string]*challenge),
		enrolments:  make(map[string]*enrolment),
		failures:    make(map[string]*userFailures),
		lastCounter: make(map[string]uint64),
	}
}

var guard = newTOTPGuard()

// issue starts a challenge for an account and returns its token.
func (g *totpGuard) issue(username string, tokenVersion uint64) (string, error) {
	raw := make([]byte, 24)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)

	g.mu.Lock()
	defer g.mu.Unlock()
	now := g.now()
	g.pruneLocked(now)
	if len(g.challenges) >= maxChallenges {
		g.evictOldestChallengeLocked()
	}
	g.challenges[token] = &challenge{
		username:     username,
		tokenVersion: tokenVersion,
		expires:      now.Add(challengeTTL),
	}
	return token, nil
}

// lookup returns a copy of a live challenge.
func (g *totpGuard) lookup(token string) (challenge, bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry, ok := g.challenges[token]
	if !ok {
		return challenge{}, false
	}
	if !g.now().Before(entry.expires) {
		delete(g.challenges, token)
		return challenge{}, false
	}
	return *entry, true
}

// fail counts a wrong code against a challenge and its account, and reports
// whether the account is now locked out of the code step.
func (g *totpGuard) fail(token, username string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if entry, ok := g.challenges[token]; ok {
		entry.attempts++
		if entry.attempts >= challengeAttempts {
			delete(g.challenges, token)
		}
	}
	return g.recordFailureLocked(username)
}

func (g *totpGuard) drop(token string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.challenges, token)
}

// locked reports whether an account's code step is closed.
func (g *totpGuard) locked(username string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry, ok := g.failures[username]
	return ok && g.now().Before(entry.lockedTill)
}

func (g *totpGuard) recordFailureLocked(username string) bool {
	now := g.now()
	entry, ok := g.failures[username]
	if !ok {
		entry = &userFailures{}
		g.failures[username] = entry
	}
	if now.Sub(entry.last) > userFailureWindow || (!entry.lockedTill.IsZero() && !now.Before(entry.lockedTill)) {
		entry.count = 0
		entry.lockedTill = time.Time{}
	}
	entry.count++
	entry.last = now
	if entry.count >= userMaxFailures {
		entry.lockedTill = now.Add(userLockout)
		return true
	}
	return false
}

// succeed clears an account's wrong codes after a login.
func (g *totpGuard) succeed(username string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	delete(g.failures, username)
}

// verify checks a TOTP code for an account and, when it is right, records its
// step, so the same code cannot be used again while it is still valid.
func (g *totpGuard) verify(username string, secret []byte, code string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	counter, ok := totp.Verify(secret, code, g.now(), g.lastCounter[username])
	if !ok {
		return false
	}
	g.lastCounter[username] = counter
	return true
}

// startEnrolment keeps a new secret for an account until it is confirmed.
func (g *totpGuard) startEnrolment(username string, secret []byte) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.enrolments[username] = &enrolment{secret: secret, expires: g.now().Add(enrolmentTTL)}
}

// confirmEnrolment checks a code against the pending secret. It returns the
// secret when the code is right, and drops the enrolment when it is right or
// has been guessed at too often. expired is true when there is no enrolment
// to confirm.
func (g *totpGuard) confirmEnrolment(username, code string) (secret []byte, expired bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry, ok := g.enrolments[username]
	now := g.now()
	if !ok || !now.Before(entry.expires) {
		delete(g.enrolments, username)
		return nil, true
	}
	counter, ok := totp.Verify(entry.secret, code, now, 0)
	if !ok {
		entry.attempts++
		if entry.attempts >= enrolmentAttempts {
			delete(g.enrolments, username)
		}
		return nil, false
	}
	delete(g.enrolments, username)
	// The code that confirmed the enrolment is spent.
	g.lastCounter[username] = counter
	return entry.secret, false
}

// forget drops everything held for an account: its challenges, its pending
// enrolment and its last step. Called when its second factor is turned off or
// reset.
func (g *totpGuard) forget(username string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	for token, entry := range g.challenges {
		if entry.username == username {
			delete(g.challenges, token)
		}
	}
	delete(g.enrolments, username)
	delete(g.lastCounter, username)
	delete(g.failures, username)
}

func (g *totpGuard) pruneLocked(now time.Time) {
	for token, entry := range g.challenges {
		if !now.Before(entry.expires) {
			delete(g.challenges, token)
		}
	}
	for username, entry := range g.enrolments {
		if !now.Before(entry.expires) {
			delete(g.enrolments, username)
		}
	}
	for username, entry := range g.failures {
		if now.Sub(entry.last) > userFailureWindow && !now.Before(entry.lockedTill) {
			delete(g.failures, username)
		}
	}
}

func (g *totpGuard) evictOldestChallengeLocked() {
	oldest := ""
	var oldestExpiry time.Time
	for token, entry := range g.challenges {
		if oldest == "" || entry.expires.Before(oldestExpiry) {
			oldest = token
			oldestExpiry = entry.expires
		}
	}
	delete(g.challenges, oldest)
}
