// Package audit keeps the record of who did what on the board: logins, the
// control a user takes of the host, power presses, virtual media, settings
// changes, macro runs and the calls an MCP client or the on-board agent makes.
//
// The record lives in a few small files under /data/audit. /data is the large
// user partition (it holds the images and the macros), where a growing file
// does not eat into the system partition, and /etc/kvm is for settings that
// both firmware slots read, which a log has no business in. The files are
// plain JSON lines, one event per line, so either slot can read what the
// other wrote and an export is a copy.
//
// Flash wear is the constraint that shapes the writer. Events wait in memory
// and go to disk together, at most every flushInterval and not at all while
// nothing happens, with no fsync. The total is capped: the active file rolls
// over at segmentBytes and only segments files are kept, so the log never
// grows past about a megabyte however busy the board is.
package audit

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// Dir is where the log lives.
	Dir = "/data/audit"

	activeName = "audit.log"

	// segmentBytes is the size at which the active file rolls over, and
	// segments is how many files are kept, the active one included. Together
	// they cap the log at segmentBytes*segments.
	segmentBytes = 256 << 10
	segments     = 4

	// flushInterval is the longest an event waits in memory. flushBatch is
	// how many may wait before they are written at once.
	flushInterval = 5 * time.Second
	flushBatch    = 32

	// pendingMax bounds the memory held while the disk will not take writes.
	pendingMax = 1024

	// coalesceWindow is how long a repeated failure is folded into the
	// previous entry. A password guesser would otherwise fill the log with
	// copies of one line and push the real history out of the rotation.
	coalesceWindow = time.Minute

	maxActor  = 64
	maxTarget = 256
	maxDetail = 256
)

// Outcomes.
const (
	OutcomeOK     = "ok"
	OutcomeFail   = "fail"
	OutcomeDenied = "denied"
)

// Categories.
const (
	CategoryAuth     = "auth"     // logins, logouts, accounts, keys, second factor
	CategoryControl  = "control"  // HID control sessions, paste, extra keys
	CategoryPower    = "power"    // ATX buttons, Wake-on-LAN, reboot
	CategoryMedia    = "media"    // virtual media and image downloads
	CategorySettings = "settings" // any other change to the board's settings
	CategoryMacro    = "macro"    // macros and auto-tap
	CategorySystem   = "system"   // updates and the like
	CategoryAgent    = "agent"    // MCP clients and the on-board agent
)

// Event is one audited action.
type Event struct {
	Time time.Time `json:"time"`
	// Actor is the account, or the name of the agent or protocol that acted.
	Actor string `json:"actor,omitempty"`
	// Via is how the actor got in: session, apikey, open (authentication is
	// off), mcp, agent, ipmi, redfish, watchdog or macro.
	Via string `json:"via,omitempty"`
	// Key is the name of the API key, when one was used.
	Key string `json:"key,omitempty"`
	// Session is the chat session an on-board agent acted for, so that two
	// conversations with the same agent can be told apart.
	Session string `json:"session,omitempty"`
	// Role is the actor's role: admin, user or viewer.
	Role     string `json:"role,omitempty"`
	IP       string `json:"ip,omitempty"`
	Category string `json:"category"`
	Action   string `json:"action"`
	Target   string `json:"target,omitempty"`
	Outcome  string `json:"outcome"`
	// Tier is the capability tier an MCP or agent call needed.
	Tier   string `json:"tier,omitempty"`
	Detail string `json:"detail,omitempty"`
	// Repeat is how many times the same failure happened inside the
	// coalescing window, when more than once.
	Repeat int `json:"repeat,omitempty"`
}

// Store writes and reads the log.
type Store struct {
	dir string

	mu      sync.Mutex
	pending []Event
	timer   *time.Timer
	size    int64
	sized   bool
	sink    func(Event)
}

// NewStore returns a store that keeps its files in dir.
func NewStore(dir string) *Store {
	return &Store{dir: dir}
}

// SetSink sets a function called with every recorded event, outside the
// store's lock. It carries the syslog forward.
func (s *Store) SetSink(sink func(Event)) {
	s.mu.Lock()
	s.sink = sink
	s.mu.Unlock()
}

// Record adds an event. It never blocks on the disk and never fails the
// caller: an audit write that went wrong is logged and the action goes on.
func (s *Store) Record(e Event) {
	e = clean(e)

	s.mu.Lock()
	if n := len(s.pending); n > 0 && e.Outcome != OutcomeOK && repeats(s.pending[n-1], e) {
		last := &s.pending[n-1]
		if last.Repeat == 0 {
			last.Repeat = 1
		}
		last.Repeat++
		s.mu.Unlock()
		return
	}

	s.pending = append(s.pending, e)
	switch {
	case len(s.pending) >= flushBatch:
		s.flushLocked()
	case s.timer == nil:
		s.timer = time.AfterFunc(flushInterval, s.Flush)
	}
	sink := s.sink
	s.mu.Unlock()

	if sink != nil {
		sink(e)
	}
}

// Flush writes the waiting events to disk.
func (s *Store) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flushLocked()
}

func (s *Store) flushLocked() {
	if s.timer != nil {
		s.timer.Stop()
		s.timer = nil
	}
	if len(s.pending) == 0 {
		return
	}

	var buf bytes.Buffer
	for _, e := range s.pending {
		line, err := json.Marshal(e)
		if err != nil {
			continue
		}
		buf.Write(line)
		buf.WriteByte('\n')
	}

	if err := s.appendLocked(buf.Bytes()); err != nil {
		log.Warnf("audit log write failed: %s", err)
		// Keep the events for the next try, up to a bound.
		if len(s.pending) > pendingMax {
			s.pending = append([]Event(nil), s.pending[len(s.pending)-pendingMax:]...)
		}
		if s.timer == nil {
			s.timer = time.AfterFunc(flushInterval, s.Flush)
		}
		return
	}
	s.pending = s.pending[:0]
}

func (s *Store) appendLocked(data []byte) error {
	if err := os.MkdirAll(s.dir, 0o700); err != nil {
		return err
	}
	active := filepath.Join(s.dir, activeName)

	if !s.sized {
		if info, err := os.Stat(active); err == nil {
			s.size = info.Size()
		}
		s.sized = true
	}
	if s.size > 0 && s.size+int64(len(data)) > segmentBytes {
		s.rotateLocked()
	}

	file, err := os.OpenFile(active, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	n, err := file.Write(data)
	s.size += int64(n)
	if cerr := file.Close(); err == nil {
		err = cerr
	}
	return err
}

// rotateLocked shifts audit.log to audit.1.log, that to audit.2.log and so
// on, dropping the oldest.
func (s *Store) rotateLocked() {
	_ = os.Remove(s.segmentPath(segments - 1))
	for i := segments - 2; i >= 1; i-- {
		_ = os.Rename(s.segmentPath(i), s.segmentPath(i+1))
	}
	_ = os.Rename(filepath.Join(s.dir, activeName), s.segmentPath(1))
	s.size = 0
}

func (s *Store) segmentPath(i int) string {
	if i == 0 {
		return filepath.Join(s.dir, activeName)
	}
	return filepath.Join(s.dir, fmt.Sprintf("audit.%d.log", i))
}

// Query selects events. Zero fields match everything.
type Query struct {
	Category string
	Outcome  string
	// Actor matches the account or key name, case-insensitively, as a part.
	Actor string
	// IP matches a part of the source address.
	IP string
	// Text matches a part of the action, target or detail.
	Text   string
	Since  time.Time
	Until  time.Time
	Offset int
	Limit  int
}

// Page is a window onto the matching events, newest first.
type Page struct {
	Events []Event `json:"events"`
	// Total is how many events match, not how many are in this page.
	Total int `json:"total"`
}

// Query returns the matching events, newest first.
func (s *Store) Query(q Query) (Page, error) {
	matches, err := s.collect(q)
	if err != nil {
		return Page{}, err
	}

	total := len(matches)
	limit := q.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	start := q.Offset
	if start < 0 {
		start = 0
	}
	if start > total {
		start = total
	}
	end := start + limit
	if end > total {
		end = total
	}
	return Page{Events: matches[start:end], Total: total}, nil
}

// collect returns every matching event, newest first.
func (s *Store) collect(q Query) ([]Event, error) {
	s.Flush()

	s.mu.Lock()
	defer s.mu.Unlock()

	var out []Event
	// Oldest file first, so the whole list is in time order before it is
	// reversed.
	for i := segments - 1; i >= 0; i-- {
		file, err := os.Open(s.segmentPath(i))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return nil, err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 0, 4096), 64<<10)
		for scanner.Scan() {
			var e Event
			if json.Unmarshal(scanner.Bytes(), &e) != nil {
				continue
			}
			if q.matches(e) {
				out = append(out, e)
			}
		}
		_ = file.Close()
	}

	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out, nil
}

func (q Query) matches(e Event) bool {
	if q.Category != "" && e.Category != q.Category {
		return false
	}
	if q.Outcome != "" && e.Outcome != q.Outcome {
		return false
	}
	if q.Actor != "" && !containsFold(e.Actor, q.Actor) && !containsFold(e.Key, q.Actor) &&
		!containsFold(e.Session, q.Actor) {
		return false
	}
	if q.IP != "" && !strings.Contains(e.IP, q.IP) {
		return false
	}
	if q.Text != "" && !containsFold(e.Action, q.Text) && !containsFold(e.Target, q.Text) &&
		!containsFold(e.Detail, q.Text) {
		return false
	}
	if !q.Since.IsZero() && e.Time.Before(q.Since) {
		return false
	}
	if !q.Until.IsZero() && e.Time.After(q.Until) {
		return false
	}
	return true
}

func containsFold(s, part string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(part))
}

// repeats reports whether b is a second copy of a within the window.
func repeats(a, b Event) bool {
	return b.Time.Sub(a.Time) < coalesceWindow &&
		a.Actor == b.Actor && a.Via == b.Via && a.Key == b.Key && a.Session == b.Session &&
		a.Role == b.Role && a.IP == b.IP &&
		a.Category == b.Category && a.Action == b.Action && a.Target == b.Target &&
		a.Outcome == b.Outcome && a.Tier == b.Tier && a.Detail == b.Detail
}

// clean stamps the time and makes every text field safe to store and to
// forward: no control characters, so a hostile user name cannot forge a
// second line of the log, and a bounded length.
func clean(e Event) Event {
	if e.Time.IsZero() {
		e.Time = time.Now()
	}
	e.Time = e.Time.UTC().Truncate(time.Millisecond)
	e.Actor = field(e.Actor, maxActor)
	e.Via = field(e.Via, 16)
	e.Key = field(e.Key, maxActor)
	e.Session = field(e.Session, maxActor)
	e.Role = field(e.Role, 8)
	e.IP = field(e.IP, 45)
	e.Category = field(e.Category, 16)
	e.Action = field(e.Action, 64)
	e.Target = field(e.Target, maxTarget)
	e.Outcome = field(e.Outcome, 8)
	e.Tier = field(e.Tier, 16)
	e.Detail = field(e.Detail, maxDetail)
	if e.Outcome == "" {
		e.Outcome = OutcomeOK
	}
	return e
}

func field(s string, max int) string {
	s = strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return ' '
		}
		return r
	}, s)
	s = strings.TrimSpace(s)
	if len(s) > max {
		// Cut on a rune boundary.
		cut := max
		for cut > 0 && s[cut]&0xC0 == 0x80 {
			cut--
		}
		s = s[:cut]
	}
	return s
}
