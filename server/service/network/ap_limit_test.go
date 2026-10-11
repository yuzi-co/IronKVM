package network

import (
	"fmt"
	"testing"
	"time"
)

type fakeClock struct{ t time.Time }

func (c *fakeClock) now() time.Time          { return c.t }
func (c *fakeClock) advance(d time.Duration) { c.t = c.t.Add(d) }

func newTestLimiter() (*apLimiter, *fakeClock) {
	clock := &fakeClock{t: time.Date(2026, 10, 11, 12, 0, 0, 0, time.UTC)}
	return newAPLimiter(clock.now), clock
}

// fail runs one wrong attempt from ip, which must be allowed to start.
func fail(t *testing.T, l *apLimiter, ip string) {
	t.Helper()
	if retry, ok := l.begin(ip); !ok {
		t.Fatalf("attempt from %s refused, retry in %s", ip, retry)
	}
	l.end(ip, false)
	l.release(ip)
}

func TestTheFirstFailuresCostNothing(t *testing.T) {
	l, _ := newTestLimiter()
	for range apFreeFailures {
		fail(t, l, "192.168.4.2")
	}
	if retry, ok := l.begin("192.168.4.2"); ok || retry != apFirstWait {
		t.Fatalf("after %d failures: ok=%v retry=%s, want a wait of %s", apFreeFailures, ok, retry, apFirstWait)
	}
}

func TestTheWaitDoublesUpToTheCeiling(t *testing.T) {
	l, clock := newTestLimiter()
	ip := "192.168.4.2"
	for range apFreeFailures {
		fail(t, l, ip)
	}

	want := apFirstWait
	for range 12 {
		retry, ok := l.begin(ip)
		if ok || retry != want {
			t.Fatalf("ok=%v retry=%s, want a wait of %s", ok, retry, want)
		}
		clock.advance(retry)
		fail(t, l, ip)
		want = min(want*2, apMaxWait)
	}
	if want != apMaxWait {
		t.Fatalf("the wait never reached the ceiling: %s", want)
	}
}

func TestOneAttemptAtATimePerAddress(t *testing.T) {
	l, _ := newTestLimiter()
	if _, ok := l.begin("192.168.4.2"); !ok {
		t.Fatal("the first attempt was refused")
	}
	if _, ok := l.begin("192.168.4.2"); ok {
		t.Fatal("a second attempt from the same address ran alongside the first")
	}
	if _, ok := l.begin("192.168.4.3"); !ok {
		t.Fatal("another address was held up by the first")
	}
	l.end("192.168.4.2", false)
	if _, ok := l.begin("192.168.4.2"); ok {
		t.Fatal("an attempt started before the failed one was released")
	}
	l.release("192.168.4.2")
	if _, ok := l.begin("192.168.4.2"); !ok {
		t.Fatal("a free failure held up the next attempt")
	}
}

func TestASuccessClearsTheRecord(t *testing.T) {
	l, _ := newTestLimiter()
	ip := "192.168.4.2"
	for range apFreeFailures - 1 {
		fail(t, l, ip)
	}
	if _, ok := l.begin(ip); !ok {
		t.Fatal("refused before the limit")
	}
	l.end(ip, true)
	l.release(ip)

	for range apFreeFailures {
		fail(t, l, ip)
	}
}

func TestManyAddressesShareOneLimit(t *testing.T) {
	l, clock := newTestLimiter()
	for i := range apGlobalFreeFailures {
		fail(t, l, fmt.Sprintf("192.168.4.%d", 10+i))
	}
	retry, ok := l.begin("192.168.4.200")
	if ok || retry != apFirstWait {
		t.Fatalf("a fresh address after %d failures elsewhere: ok=%v retry=%s", apGlobalFreeFailures, ok, retry)
	}
	clock.advance(retry)
	if _, ok := l.begin("192.168.4.200"); !ok {
		t.Fatal("the shared wait did not end")
	}
}

func TestTheSharedWaitHasALowerCeiling(t *testing.T) {
	if got := wait(apGlobalFreeFailures+50, apGlobalFreeFailures, apFirstWait, apGlobalMaxWait); got != apGlobalMaxWait {
		t.Fatalf("got %s", got)
	}
}

func TestQuietRecordsAreForgotten(t *testing.T) {
	l, clock := newTestLimiter()
	ip := "192.168.4.2"
	for range apFreeFailures + 2 {
		retry, ok := l.begin(ip)
		if !ok {
			clock.advance(retry)
		} else {
			l.end(ip, false)
			l.release(ip)
		}
	}
	clock.advance(apForget + time.Minute)
	for range apFreeFailures {
		fail(t, l, ip)
	}
}

func TestTheTableIsBounded(t *testing.T) {
	l, clock := newTestLimiter()
	for i := range apMaxClients * 3 {
		ip := fmt.Sprintf("10.%d.%d.%d", i>>16&0xff, i>>8&0xff, i&0xff)
		retry, ok := l.begin(ip)
		if !ok {
			// The shared record is waiting; let it pass.
			clock.advance(retry)
			if _, ok = l.begin(ip); !ok {
				t.Fatalf("attempt %d refused after the wait", i)
			}
		}
		l.end(ip, false)
		l.release(ip)
	}
	if len(l.clients) > apMaxClients {
		t.Fatalf("the table holds %d records, more than %d", len(l.clients), apMaxClients)
	}
}

func TestAFullTableOfOpenAttemptsRefusesANewAddress(t *testing.T) {
	l, _ := newTestLimiter()
	for i := range apMaxClients {
		if _, ok := l.begin(fmt.Sprintf("10.0.%d.%d", i>>8, i&0xff)); !ok {
			t.Fatalf("attempt %d refused", i)
		}
	}
	if _, ok := l.begin("10.9.9.9"); ok {
		t.Fatal("a new address was let in past a table of open attempts")
	}
}
