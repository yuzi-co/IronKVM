package authn

import (
	"errors"
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

const (
	// FirstBootPasswordFile holds the password a freshly flashed IronKVM card
	// generated for the device owner. The image's first boot writes it, next to
	// the account file, and the OLED page and the card's boot partition show it.
	// It is plain text because its readers are a person with the card and the
	// display, and it is removed once the owner has chosen a password.
	FirstBootPasswordFile = "/etc/kvm/first-boot-password"

	// FirstBootCardFile is the copy on the card's boot partition (FAT, the
	// first partition), the one a person holding the card can read from any
	// computer.
	FirstBootCardFile = "/boot/ironkvm-first-password.txt"
)

// ProvisionFirstBoot gives a board with no account file the password its first
// boot generated, in place of admin/admin. It reports whether it did.
//
// It never replaces an account. A board that has an account file keeps it
// whatever the password file says, and so does a board with no password file:
// that one is an upgrade of an unmodified vendor image, still using
// admin/admin, and changing its credentials on an update would lock its owner
// out. The owner is still made to choose a password at the first login.
func (s *Store) ProvisionFirstBoot(passwordFile string) (bool, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, err := os.Stat(s.path); err == nil {
		return false, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return false, err
	}

	data, err := os.ReadFile(passwordFile)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	password := strings.TrimSpace(string(data))
	if err = ValidatePassword(password); err != nil {
		return false, fmt.Errorf("first-boot password file: %w", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false, err
	}
	tokenVersion, err := newTokenVersion()
	if err != nil {
		return false, err
	}
	db := &database{
		Version: currentFileVersion,
		Users: []User{{
			Username:           defaultUsername,
			PasswordHash:       string(hash),
			Role:               RoleAdmin,
			Enabled:            true,
			TokenVersion:       tokenVersion,
			MustChangePassword: true,
			SystemAccount:      true,
		}},
	}
	if err = s.saveLocked(db); err != nil {
		return false, err
	}
	return true, nil
}

// FirstBootPending reports whether the device owner still has to choose a
// password, which is when the first-boot password is still the live one.
func (s *Store) FirstBootPending() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	db, _, err := s.readLocked()
	if err != nil || db == nil {
		return false
	}
	for _, user := range db.Users {
		if user.SystemAccount {
			return user.MustChangePassword
		}
	}
	return false
}

// ClearFirstBoot removes the files that show the first-boot password. They are
// removed when the password stops being the live one, so a card or a display
// never offers a password that no longer works. A file that is already gone is
// fine; any other failure is returned for the caller to log.
func ClearFirstBoot(paths ...string) error {
	var first error
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) && first == nil {
			first = err
		}
	}
	return first
}

// SettleFirstBoot is run when the server starts. It gives a fresh board its
// generated password, and removes the files that show it from a board that no
// longer needs them.
func SettleFirstBoot() {
	provisioned, err := DefaultStore.ProvisionFirstBoot(FirstBootPasswordFile)
	if err != nil {
		// Never the password itself. The board keeps admin/admin with the
		// forced change, which is what it did before.
		log.Errorf("first-boot password not applied: %s", err)
		return
	}
	if provisioned {
		log.Info("account created with the first-boot password")
		return
	}
	if !DefaultStore.FirstBootPending() {
		if err = ClearFirstBoot(FirstBootPasswordFile, FirstBootCardFile); err != nil {
			log.Warnf("remove the first-boot password files: %s", err)
		}
	}
}
