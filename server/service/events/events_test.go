package events

import (
	"sync"
	"testing"
	"time"
)

func collect(t *testing.T, b *Bus) (*[]Event, func(), func(n int)) {
	t.Helper()

	var (
		mu  sync.Mutex
		got []Event
	)
	cancel := b.Subscribe(func(e Event) {
		mu.Lock()
		got = append(got, e)
		mu.Unlock()
	})
	wait := func(n int) {
		t.Helper()
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			mu.Lock()
			have := len(got)
			mu.Unlock()
			if have >= n {
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		t.Fatalf("wanted %d events, got %d", n, len(got))
	}
	return &got, cancel, wait
}

func TestEveryEventReachesEverySubscriberInOrder(t *testing.T) {
	b := NewBus()
	a, cancelA, waitA := collect(t, b)
	c, cancelC, waitC := collect(t, b)
	defer cancelA()
	defer cancelC()

	for _, typ := range []string{SignalLost, SignalRestored, ATXPress} {
		b.Publish(Event{Type: typ})
	}
	waitA(3)
	waitC(3)

	for _, got := range []*[]Event{a, c} {
		for i, want := range []string{SignalLost, SignalRestored, ATXPress} {
			if (*got)[i].Type != want {
				t.Errorf("event %d is %s, want %s", i, (*got)[i].Type, want)
			}
		}
	}
}

func TestPublishStampsATimeAndKeepsOne(t *testing.T) {
	b := NewBus()
	got, cancel, wait := collect(t, b)
	defer cancel()

	set := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	b.Publish(Event{Type: SignalLost})
	b.Publish(Event{Type: SignalLost, Time: set})
	wait(2)

	if (*got)[0].Time.IsZero() {
		t.Error("an event published without a time was not stamped")
	}
	if !(*got)[1].Time.Equal(set) {
		t.Errorf("a given time was replaced: %v", (*got)[1].Time)
	}
}

func TestACancelledSubscriberHearsNothingMore(t *testing.T) {
	b := NewBus()
	got, cancel, wait := collect(t, b)

	b.Publish(Event{Type: SignalLost})
	wait(1)
	cancel()
	cancel() // twice is fine
	b.Publish(Event{Type: SignalRestored})
	time.Sleep(30 * time.Millisecond)

	if len(*got) != 1 {
		t.Fatalf("got %d events after cancel, want 1", len(*got))
	}
}

func TestASlowSubscriberNeverBlocksThePublisher(t *testing.T) {
	b := NewBus()
	release := make(chan struct{})
	var mu sync.Mutex
	var seen []string
	cancel := b.Subscribe(func(e Event) {
		<-release
		mu.Lock()
		seen = append(seen, e.Title)
		mu.Unlock()
	})
	defer cancel()

	done := make(chan struct{})
	go func() {
		for i := 0; i < subscriberQueue*3; i++ {
			b.Publish(Event{Type: ATXPress, Title: "e"})
		}
		b.Publish(Event{Type: ATXPress, Title: "last"})
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Publish blocked on a slow subscriber")
	}
	close(release)

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(seen)
		last := n > 0 && seen[n-1] == "last"
		mu.Unlock()
		if last {
			if n > subscriberQueue+2 {
				t.Errorf("the subscriber saw %d events, more than its queue holds", n)
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the newest event was dropped instead of the oldest")
}
