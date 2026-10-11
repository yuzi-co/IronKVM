package hostlog

import (
	"strings"
	"sync"
	"time"
)

// Entry is one log line as the ring keeps it and the page shows it.
type Entry struct {
	Seq      uint64    `json:"seq"`
	Time     time.Time `json:"time"`
	Source   string    `json:"source"`
	Proto    string    `json:"proto"`
	Facility int       `json:"facility"`
	Severity int       `json:"severity"`
	Host     string    `json:"host,omitempty"`
	Tag      string    `json:"tag,omitempty"`
	Msg      string    `json:"msg"`
}

// entryOverhead stands for what an entry costs besides its strings, so a
// flood of empty lines still fills the cap.
const entryOverhead = 96

func (e Entry) size() int {
	return entryOverhead + len(e.Source) + len(e.Proto) + len(e.Host) + len(e.Tag) + len(e.Msg)
}

// Filter selects entries.
type Filter struct {
	// Words must all appear in the host, tag or message, in any case.
	Words []string
	// MaxSeverity keeps entries at this severity or more urgent (0 is
	// emergency, 7 is debug). Negative keeps every one. An entry with no
	// severity always passes.
	MaxSeverity int
	// Source keeps entries from this address. Empty keeps every one.
	Source string
	// After keeps entries newer than this sequence number.
	After uint64
}

// ParseWords splits a search string into the words Filter wants.
func ParseWords(q string) []string {
	return strings.Fields(strings.ToLower(q))
}

// Match reports whether the entry passes the filter.
func (f Filter) Match(e Entry) bool {
	if e.Seq <= f.After {
		return false
	}
	if f.MaxSeverity >= 0 && e.Severity >= 0 && e.Severity > f.MaxSeverity {
		return false
	}
	if f.Source != "" && e.Source != f.Source {
		return false
	}
	if len(f.Words) > 0 {
		hay := strings.ToLower(e.Host + " " + e.Tag + " " + e.Msg)
		for _, w := range f.Words {
			if !strings.Contains(hay, w) {
				return false
			}
		}
	}
	return true
}

// subscriberBuffer is how far a live viewer may fall behind before it loses
// lines.
const subscriberBuffer = 512

// Ring holds the newest entries up to a size cap in bytes.
type Ring struct {
	mu sync.Mutex
	// q[head:] are the live entries.
	q     []Entry
	head  int
	bytes int
	cap   int
	next  uint64

	subs    map[int]chan Entry
	nextSub int
}

// NewRing returns a ring that holds up to capBytes.
func NewRing(capBytes int) *Ring {
	return &Ring{cap: capBytes, next: 1, subs: map[int]chan Entry{}}
}

// SetCap changes the cap and drops the oldest entries to fit.
func (r *Ring) SetCap(capBytes int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.cap = capBytes
	r.evictLocked()
}

func (r *Ring) evictLocked() {
	for r.bytes > r.cap && r.head < len(r.q) {
		r.bytes -= r.q[r.head].size()
		r.q[r.head] = Entry{}
		r.head++
	}
	// Compact once the dead front is as long as the live part, so the cost
	// of the copy is spread over as many lines as it took to build it.
	if r.head > 0 && r.head >= len(r.q)-r.head {
		r.q = append([]Entry(nil), r.q[r.head:]...)
		r.head = 0
	}
}

// Add stamps the entry with the next sequence number, stores it and hands it
// to the live viewers. It returns the stored entry.
func (r *Ring) Add(e Entry) Entry {
	r.mu.Lock()
	defer r.mu.Unlock()

	e.Seq = r.next
	r.next++
	r.put(e)
	for _, ch := range r.subs {
		select {
		case ch <- e:
		default:
			// A viewer that cannot keep up loses lines; the page can
			// fetch what it missed by sequence number.
		}
	}
	return e
}

// Load puts entries read back from disk in, oldest first, before anything
// else arrives. They are numbered afresh: sequence numbers only order the
// entries of one run.
func (r *Ring) Load(entries []Entry) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, e := range entries {
		e.Seq = r.next
		r.next++
		r.put(e)
	}
}

func (r *Ring) put(e Entry) {
	r.q = append(r.q, e)
	r.bytes += e.size()
	r.evictLocked()
}

// Query returns up to limit entries that pass the filter, oldest first. When
// more match than fit, the newest are kept.
func (r *Ring) Query(f Filter, limit int) []Entry {
	r.mu.Lock()
	defer r.mu.Unlock()

	var out []Entry
	for i := len(r.q) - 1; i >= r.head && (limit <= 0 || len(out) < limit); i-- {
		if f.Match(r.q[i]) {
			out = append(out, r.q[i])
		}
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return out
}

// Subscribe returns a channel of the entries added from now on, and a
// function that ends the subscription.
func (r *Ring) Subscribe() (<-chan Entry, func()) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ch := make(chan Entry, subscriberBuffer)
	id := r.nextSub
	r.nextSub++
	r.subs[id] = ch

	var once sync.Once
	return ch, func() {
		once.Do(func() {
			r.mu.Lock()
			delete(r.subs, id)
			r.mu.Unlock()
		})
	}
}

// Clear forgets every entry. Sequence numbers go on.
func (r *Ring) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.q, r.head, r.bytes = nil, 0, 0
}

// Stats returns how many entries the ring holds and how many bytes they count
// for.
func (r *Ring) Stats() (count, bytes int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return len(r.q) - r.head, r.bytes
}
