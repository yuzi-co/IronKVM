package storage

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAJobReportsItsProgressAndOutcome(t *testing.T) {
	var j jobSlot
	if st := j.status(); st.Status != string(jobStateIdle) {
		t.Fatalf("a fresh slot is %q", st.Status)
	}

	done, err := j.begin(context.Background(), "disk.img: docs/a.bin")
	if err != nil {
		t.Fatal(err)
	}
	j.progress(done, 25, 100)
	st := j.status()
	if st.Status != string(jobStateInProgress) || st.Name != "disk.img: docs/a.bin" ||
		st.Loaded != 25 || st.Total != 100 || st.Percentage != "25.00%" {
		t.Fatalf("status = %+v", st)
	}
	if name, ok := j.active(); !ok || name != "disk.img: docs/a.bin" {
		t.Errorf("active = %q, %v", name, ok)
	}

	j.finish(done, jobStateFailed, "disk full")
	st = j.status()
	if st.Status != string(jobStateFailed) || st.Message != "disk full" || st.Loaded != 0 {
		t.Fatalf("status after the failure = %+v", st)
	}
	if _, ok := j.active(); ok {
		t.Error("the slot is still active")
	}
	select {
	case <-done:
	default:
		t.Error("finish did not close the job's channel")
	}
}

// A stale handle cannot report into, or end, the job that followed it.
func TestAFinishedJobCannotTouchTheNextOne(t *testing.T) {
	var j jobSlot
	first, _ := j.begin(context.Background(), "first")
	j.finish(first, jobStateSuccess, "")

	second, _ := j.begin(context.Background(), "second")
	j.progress(first, 99, 100)
	j.finish(first, jobStateFailed, "late")
	st := j.status()
	if st.Name != "second" || st.Status != string(jobStateInProgress) || st.Loaded != 0 {
		t.Fatalf("status = %+v", st)
	}
	j.finish(second, jobStateSuccess, "")
}

// A second transfer waits its turn rather than failing, and runs once the
// first one ends.
func TestASecondJobQueuesBehindTheFirst(t *testing.T) {
	var j jobSlot
	first, _ := j.begin(context.Background(), "first")

	started := make(chan chan struct{})
	go func() {
		done, err := j.begin(context.Background(), "second")
		if err != nil {
			t.Error(err)
		}
		started <- done
	}()

	select {
	case <-started:
		t.Fatal("the second job started while the first ran")
	case <-time.After(50 * time.Millisecond):
	}

	j.finish(first, jobStateSuccess, "")
	select {
	case done := <-started:
		if st := j.status(); st.Name != "second" {
			t.Errorf("running %q", st.Name)
		}
		j.finish(done, jobStateSuccess, "")
	case <-time.After(5 * time.Second):
		t.Fatal("the second job never started")
	}
}

// A client that gives up while queued leaves the queue, and the turn passes on
// to the one behind it.
func TestAGivenUpWaiterPassesItsTurnOn(t *testing.T) {
	var j jobSlot
	first, _ := j.begin(context.Background(), "first")

	ctx, cancel := context.WithCancel(context.Background())
	gaveUp := make(chan error)
	go func() {
		_, err := j.begin(ctx, "impatient")
		gaveUp <- err
	}()
	waitForWaiters(t, &j, 1)

	third := make(chan chan struct{})
	go func() {
		done, _ := j.begin(context.Background(), "third")
		third <- done
	}()
	waitForWaiters(t, &j, 2)

	cancel()
	if err := <-gaveUp; !errors.Is(err, context.Canceled) {
		t.Fatalf("the impatient waiter got %v", err)
	}

	j.finish(first, jobStateSuccess, "")
	select {
	case done := <-third:
		j.finish(done, jobStateSuccess, "")
	case <-time.After(5 * time.Second):
		t.Fatal("the third job never started")
	}
}

func TestTheQueueIsBounded(t *testing.T) {
	var j jobSlot
	first, _ := j.begin(context.Background(), "first")
	defer j.finish(first, jobStateSuccess, "")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for i := 0; i < maxQueuedJobs; i++ {
		go func() { _, _ = j.begin(ctx, "queued") }()
	}
	waitForWaiters(t, &j, maxQueuedJobs)

	if _, err := j.begin(context.Background(), "one too many"); !errors.Is(err, errTooManyJobs) {
		t.Fatalf("got %v, want errTooManyJobs", err)
	}
}

func waitForWaiters(t *testing.T, j *jobSlot, n int) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for {
		j.mu.Lock()
		got := len(j.waiters)
		j.mu.Unlock()
		if got == n {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("%d waiters, want %d", got, n)
		}
		time.Sleep(time.Millisecond)
	}
}
