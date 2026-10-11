package network

import (
	"sync"
	"time"
)

// apLimiter bounds guesses at the AP password. In AP mode the setup hotspot
// is open to anyone in radio range, and the two routes that check the
// password need no session, so without it the password could be guessed as
// fast as the board answers.
//
// It is always on, unlike the web login's lockout, which the owner enables
// in server.yaml. It never locks an address out for good: after a few free
// failures each further one doubles the wait before the next attempt, up to
// a ceiling, and a success clears the record.
//
// It counts per address, and it allows one attempt per address at a time, so
// parallel requests do not multiply the rate. A client in range can take
// several addresses on the hotspot's subnet, so all failures also count
// towards one shared record with a larger allowance and a shorter ceiling.
// While that one is waiting nobody can try, the owner included; that is the
// price of a password that cannot be ground through from many addresses.
type apLimiter struct {
	mu      sync.Mutex
	clients map[string]*apRecord
	global  apRecord
	now     func() time.Time
}

type apRecord struct {
	failures int
	last     time.Time // the last failure
	until    time.Time // no attempt before this
	busy     bool      // an attempt is being checked
}

const (
	// apFreeFailures are the failures an address has before it must wait
	// between attempts.
	apFreeFailures = 3
	// apGlobalFreeFailures are the failures all addresses share before
	// every one of them must wait.
	apGlobalFreeFailures = 20

	apFirstWait     = 2 * time.Second
	apMaxWait       = 5 * time.Minute
	apGlobalMaxWait = time.Minute

	// apForget is how long a record lasts after its last failure.
	apForget = 30 * time.Minute

	// apMaxClients bounds the table. A hotspot has a few clients.
	apMaxClients = 256
)

var apAttempts = newAPLimiter(time.Now)

func newAPLimiter(now func() time.Time) *apLimiter {
	return &apLimiter{clients: make(map[string]*apRecord), now: now}
}

// wait is how long a record with this many failures must wait before the
// next attempt: nothing until it has used the free ones, then first, doubling
// with each further failure, up to ceiling.
func wait(failures, free int, first, ceiling time.Duration) time.Duration {
	if failures < free {
		return 0
	}
	d := first
	for i := free; i < failures && d < ceiling; i++ {
		d *= 2
	}
	return min(d, ceiling)
}

// begin starts an attempt from ip. When it may not start, it returns how long
// to wait before the next one and false. Every attempt that starts must be
// recorded with end and then closed with release.
func (l *apLimiter) begin(ip string) (time.Duration, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	forget(&l.global, now)
	if now.Before(l.global.until) {
		return l.global.until.Sub(now), false
	}

	rec := l.clients[ip]
	if rec != nil {
		forget(rec, now)
		if rec.busy {
			return apFirstWait, false
		}
		if now.Before(rec.until) {
			return rec.until.Sub(now), false
		}
	} else {
		if len(l.clients) >= apMaxClients && !l.evictLocked(now) {
			return apFirstWait, false
		}
		rec = &apRecord{}
		l.clients[ip] = rec
	}

	rec.busy = true
	return 0, true
}

// end records how an attempt went. A failure is counted at once, so that
// the shared record sees a burst from many addresses while it is still
// running, but the attempt stays open until release.
func (l *apLimiter) end(ip string, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	rec := l.clients[ip]
	if rec == nil {
		return
	}
	if ok {
		delete(l.clients, ip)
		l.global = apRecord{}
		return
	}

	now := l.now()
	rec.failures++
	rec.last = now
	rec.until = now.Add(wait(rec.failures, apFreeFailures, apFirstWait, apMaxWait))

	l.global.failures++
	l.global.last = now
	l.global.until = now.Add(wait(l.global.failures, apGlobalFreeFailures, apFirstWait, apGlobalMaxWait))
}

// forget clears a record that has been quiet for apForget.
func forget(rec *apRecord, now time.Time) {
	if !rec.busy && rec.failures > 0 && now.Sub(rec.last) > apForget {
		*rec = apRecord{}
	}
}

// evictLocked frees a slot in a full table: the record whose last failure is
// oldest among those not being checked. Dropping it costs its address its
// count, but the shared record still holds that address's failures.
func (l *apLimiter) evictLocked(now time.Time) bool {
	var oldestIP string
	var oldest time.Time
	for ip, rec := range l.clients {
		if rec.busy {
			continue
		}
		if !now.Before(rec.until) && now.Sub(rec.last) > apForget {
			delete(l.clients, ip)
			return true
		}
		if oldestIP == "" || rec.last.Before(oldest) {
			oldestIP, oldest = ip, rec.last
		}
	}
	if oldestIP == "" {
		return false
	}
	delete(l.clients, oldestIP)
	return true
}

// release closes an attempt. Another may start from the same address.
func (l *apLimiter) release(ip string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if rec := l.clients[ip]; rec != nil {
		rec.busy = false
	}
}
