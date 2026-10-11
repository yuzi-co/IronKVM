// Package events is the server's event bus: a place for the services that see
// something happen (the signal drops, the watchdog presses, a login fails) to
// say so without knowing who listens. Notifications subscribe to it. The
// package imports nothing else from the server, so any service may publish.
package events

import (
	"sync"
	"time"
)

// The event types. A type is also the key the notifier rate limits under.
const (
	// SignalLost and SignalRestored report the HDMI signal.
	SignalLost     = "signal.lost"
	SignalRestored = "signal.restored"
	// WatchdogAction is a press by the host watchdog.
	WatchdogAction = "watchdog.action"
	// LoginFailure is one failed login. It is raw: the notifier turns a run
	// of them into LoginFailures.
	LoginFailure = "login.failure"
	// LoginFailures is a run of failed logins over the owner's threshold.
	LoginFailures = "login.failures"
	// ATXPress is a press of the power or reset button, from whoever asked.
	ATXPress = "atx.press"
	// FailsafeEntered means the server started without video after a crash
	// loop.
	FailsafeEntered = "failsafe.entered"
	// HostLogMatch is a line from the host's log that matched a phrase the
	// owner asked to be told about.
	HostLogMatch = "hostlog.match"
)

// Event is one thing that happened. Fields carry the details a channel may
// show; they never carry a secret.
type Event struct {
	Type    string            `json:"type"`
	Time    time.Time         `json:"time"`
	Title   string            `json:"title"`
	Message string            `json:"message,omitempty"`
	Fields  map[string]string `json:"fields,omitempty"`

	// JPEG is a picture the publisher already has, such as the watchdog's
	// screenshot from before its press. Without one the notifier takes a
	// snapshot when it sends.
	JPEG []byte `json:"-"`
	// NoSnapshot says the notifier must not take a snapshot: there is
	// nothing to see, as when the signal is gone.
	NoSnapshot bool `json:"-"`
}

// subscriberQueue is how many events one subscriber may fall behind by. A
// slower one loses the oldest events rather than holding up the publisher,
// which may be the login handler.
const subscriberQueue = 64

type subscriber struct {
	ch   chan Event
	done chan struct{}
}

// Bus delivers each published event to every subscriber, in order, on the
// subscriber's own goroutine.
type Bus struct {
	mu   sync.Mutex
	subs map[int]*subscriber
	next int
}

// NewBus returns an empty bus.
func NewBus() *Bus {
	return &Bus{subs: map[int]*subscriber{}}
}

// Default is the bus the server's services share.
var Default = NewBus()

// Subscribe calls fn for each event published from now on, one at a time. The
// returned function stops the subscription and waits for a call in progress.
func (b *Bus) Subscribe(fn func(Event)) (cancel func()) {
	s := &subscriber{ch: make(chan Event, subscriberQueue), done: make(chan struct{})}

	b.mu.Lock()
	id := b.next
	b.next++
	b.subs[id] = s
	b.mu.Unlock()

	finished := make(chan struct{})
	go func() {
		defer close(finished)
		for {
			select {
			case ev := <-s.ch:
				fn(ev)
			case <-s.done:
				return
			}
		}
	}()

	var once sync.Once
	return func() {
		once.Do(func() {
			b.mu.Lock()
			delete(b.subs, id)
			b.mu.Unlock()
			close(s.done)
			<-finished
		})
	}
}

// Publish hands the event to every subscriber and returns at once. A
// subscriber whose queue is full drops its oldest event for this one.
func (b *Bus) Publish(ev Event) {
	if ev.Time.IsZero() {
		ev.Time = time.Now()
	}

	b.mu.Lock()
	subs := make([]*subscriber, 0, len(b.subs))
	for _, s := range b.subs {
		subs = append(subs, s)
	}
	b.mu.Unlock()

	for _, s := range subs {
		for {
			select {
			case s.ch <- ev:
			default:
				select {
				case <-s.ch:
				default:
				}
				continue
			}
			break
		}
	}
}

// Subscribe subscribes to the default bus.
func Subscribe(fn func(Event)) func() { return Default.Subscribe(fn) }

// Publish publishes on the default bus.
func Publish(ev Event) { Default.Publish(ev) }
