package webrtc

import (
	"sync"
	"testing"

	"golang.org/x/sys/unix"
)

func TestKeyNiceSetting(t *testing.T) {
	cases := map[string]int{"": defaultKeyNice, "off": 0, "0": 0, "-5": -5, "-40": -20, "junk": defaultKeyNice}
	for v, want := range cases {
		t.Setenv(keyNiceEnv, v)
		if got := keyNice(); got != want {
			t.Errorf("%s=%q gives %d, want %d", keyNiceEnv, v, got, want)
		}
	}
}

// withNice replaces the priority calls for one test.
func withNice(t *testing.T, set func(int, int) error) {
	t.Helper()

	originalSet, originalGet := setThreadNice, threadNice
	keyBoostOff.Store(false)
	t.Cleanup(func() {
		setThreadNice, threadNice = originalSet, originalGet
		keyBoostOff.Store(false)
	})
	setThreadNice = set
	threadNice = func(int) (int, error) { return 0, nil }
}

func TestAKeyframeIsWrittenAtTheRaisedPriorityAndPutBack(t *testing.T) {
	var mutex sync.Mutex
	var set []int
	withNice(t, func(_ int, nice int) error {
		mutex.Lock()
		defer mutex.Unlock()
		set = append(set, nice)

		return nil
	})
	t.Setenv(keyNiceEnv, "")

	done := make(chan struct{})
	go func() {
		defer close(done)
		b := newKeyBoost()
		b.raise()
		b.lower()
	}()
	<-done

	if len(set) != 2 || set[0] != defaultKeyNice || set[1] != 0 {
		t.Fatalf("priorities set %v, want [%d 0]", set, defaultKeyNice)
	}
}

// Where the server may not raise its priority, it says so once and writes on.
func TestARefusedPriorityIsTriedOnceForEveryWriter(t *testing.T) {
	calls := 0
	withNice(t, func(int, int) error {
		calls++

		return unix.EPERM
	})
	t.Setenv(keyNiceEnv, "")

	done := make(chan struct{})
	go func() {
		defer close(done)
		first := newKeyBoost()
		first.raise()
		first.lower()
		first.raise()

		second := newKeyBoost()
		second.raise()
	}()
	<-done

	if calls != 1 {
		t.Fatalf("setpriority called %d times, want once", calls)
	}
	if !keyBoostOff.Load() {
		t.Fatal("the boost is still on after a refusal")
	}
}
