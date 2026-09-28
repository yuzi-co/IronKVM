package watchdog

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// LogDir holds the action log and its screenshots. It is on /data and not on
// /kvmapp: the log is written only when the watchdog acts, a few times an hour
// at most, and it must survive a reboot, which is the point of reading it.
const LogDir = "/data/ironkvm/watchdog"

const (
	logFile = "log.json"
	// logLimit is how many actions the log keeps. An entry that falls off the
	// end takes its screenshot with it.
	logLimit = 50
)

// Entry is one action the watchdog took.
type Entry struct {
	// ID is the time of the action in Unix nanoseconds, which also names the
	// screenshot.
	ID     string    `json:"id"`
	Time   time.Time `json:"time"`
	Action string    `json:"action"`
	// Reason is ReasonFrozen or ReasonNoSignal.
	Reason string `json:"reason"`
	// StuckSeconds is how long the host had shown no sign of life.
	StuckSeconds int `json:"stuckSeconds"`
	// Screenshot is the path of the screenshot taken before the press. Empty
	// when there was no picture to take.
	Screenshot string `json:"screenshot,omitempty"`
	// Error is why the press failed. Empty when it did not.
	Error string `json:"error,omitempty"`
}

// Log is the action log, kept in a directory as a JSON file and one JPEG per
// entry.
type Log struct {
	dir string

	mu      sync.Mutex
	entries []Entry // oldest first
}

// OpenLog loads the log kept in dir. A missing or unreadable file gives an
// empty log; the directory is created on the first action.
func OpenLog(dir string) *Log {
	l := &Log{dir: dir}

	data, err := os.ReadFile(filepath.Join(dir, logFile))
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Errorf("watchdog: read the action log: %s", err)
		}
		return l
	}
	if err := json.Unmarshal(data, &l.entries); err != nil {
		log.Errorf("watchdog: the action log is not valid, starting a new one: %s", err)
		l.entries = nil
	}
	return l
}

// Add records e, and jpeg as its screenshot when it is not empty. The entry is
// kept in memory even when the disk refuses it, so the page still shows it.
func (l *Log) Add(e Entry, jpeg []byte) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	var errs []error
	if err := os.MkdirAll(l.dir, 0o755); err != nil {
		errs = append(errs, err)
	} else if len(jpeg) > 0 {
		path := filepath.Join(l.dir, e.ID+".jpg")
		if err := os.WriteFile(path, jpeg, 0o644); err != nil {
			errs = append(errs, err)
		} else {
			e.Screenshot = path
		}
	}

	l.entries = append(l.entries, e)
	if drop := len(l.entries) - logLimit; drop > 0 {
		for _, old := range l.entries[:drop] {
			if old.Screenshot != "" {
				if err := os.Remove(old.Screenshot); err != nil && !errors.Is(err, os.ErrNotExist) {
					errs = append(errs, err)
				}
			}
		}
		l.entries = slices.Clone(l.entries[drop:])
	}

	if err := l.saveLocked(); err != nil {
		errs = append(errs, err)
	}
	return errors.Join(errs...)
}

// saveLocked writes the log by rename, so a power cut leaves the old file or
// the new one and never half of one.
func (l *Log) saveLocked() error {
	data, err := json.MarshalIndent(l.entries, "", "  ")
	if err != nil {
		return err
	}

	tmp, err := os.CreateTemp(l.dir, ".log-*.json")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(l.dir, logFile))
}

// Entries returns the log, newest first.
func (l *Log) Entries() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()

	out := slices.Clone(l.entries)
	slices.Reverse(out)
	return out
}

// Screenshot returns the path of the screenshot for the entry id. It answers
// only for an entry in the log that has one, and it builds the path from the
// ID rather than trusting the one stored, so a request cannot reach any other
// file.
func (l *Log) Screenshot(id string) (string, bool) {
	if _, err := strconv.ParseUint(id, 10, 64); err != nil {
		return "", false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for _, e := range l.entries {
		if e.ID == id && e.Screenshot != "" {
			return filepath.Join(l.dir, e.ID+".jpg"), true
		}
	}
	return "", false
}
