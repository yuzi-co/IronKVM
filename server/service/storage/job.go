// Ported from sipeed/NanoKVM (server/service/storage/job.go, commit 5572e1de,
// PR #957). The slot is the package's rather than a Service's, so every
// caller of the storage package shares it.
// Copyright (c) Sipeed. Licensed under the GNU General Public License v3.0.

package storage

import (
	"context"
	"errors"
	"sync"

	"NanoKVM-Server/proto"
	"NanoKVM-Server/utils"
)

type jobState string

const (
	jobStateIdle       jobState = "idle"
	jobStateInProgress jobState = "in_progress"
	jobStateSuccess    jobState = "success"
	jobStateFailed     jobState = "failed"
)

// maxQueuedJobs bounds the queue: the transfers hold their HTTP connections
// open while they wait, so an unbounded queue would let one client hoard them.
const maxQueuedJobs = 8

var errTooManyJobs = errors.New("too many transfers are already queued")

// jobSlot runs one file transfer into an image at a time and reports its
// progress. The board has one core and a slow card; two uploads at once
// would only take longer each.
type jobSlot struct {
	mu      sync.Mutex
	done    chan struct{}
	state   jobState
	name    string
	loaded  int64
	total   int64
	message string
	waiters []chan struct{}
}

// transfers is the slot every upload into an image takes.
var transfers = &jobSlot{}

// begin claims the slot, queueing behind whatever is running instead of
// refusing: the user asked for a file to be written, and waiting their turn is
// what they expect. The queue is first come, first served, and a request
// whose client goes away leaves it.
func (j *jobSlot) begin(ctx context.Context, name string) (chan struct{}, error) {
	for {
		j.mu.Lock()

		if j.done == nil {
			done := make(chan struct{})
			j.done = done
			j.state = jobStateInProgress
			j.name = name
			j.loaded = 0
			j.total = 0
			j.message = ""
			j.mu.Unlock()

			return done, nil
		}

		if len(j.waiters) >= maxQueuedJobs {
			j.mu.Unlock()
			return nil, errTooManyJobs
		}

		waiter := make(chan struct{})
		j.waiters = append(j.waiters, waiter)
		j.mu.Unlock()

		select {
		case <-waiter:
			// Our turn; the loop claims the slot.
		case <-ctx.Done():
			j.abandon(waiter)
			return nil, ctx.Err()
		}
	}
}

// abandon drops a waiter whose client went away. A waiter that is still
// queued only has to leave the queue; one that was already handed the turn
// (finish closed its channel) has to pass it on, or the transfers behind it
// would wait until their own clients gave up too.
func (j *jobSlot) abandon(waiter chan struct{}) {
	j.mu.Lock()
	defer j.mu.Unlock()

	for i, queued := range j.waiters {
		if queued == waiter {
			j.waiters = append(j.waiters[:i], j.waiters[i+1:]...)
			return
		}
	}

	if j.done == nil && len(j.waiters) > 0 {
		close(j.waiters[0])
		j.waiters = j.waiters[1:]
	}
}

// progress records how far the job holding done has come. A finished job's
// late report is dropped.
func (j *jobSlot) progress(done chan struct{}, loaded, total int64) {
	j.mu.Lock()
	defer j.mu.Unlock()

	if j.done == done {
		j.loaded = loaded
		j.total = total
	}
}

// finish ends the job holding done and hands the slot to the first waiter.
func (j *jobSlot) finish(done chan struct{}, state jobState, message string) {
	j.mu.Lock()
	if j.done != done {
		j.mu.Unlock()
		return
	}

	j.done = nil
	j.state = state
	j.loaded = 0
	j.total = 0
	j.message = message

	if len(j.waiters) > 0 {
		close(j.waiters[0])
		j.waiters = j.waiters[1:]
	}
	j.mu.Unlock()

	close(done)
}

// active returns what the running job writes, if one runs.
func (j *jobSlot) active() (string, bool) {
	j.mu.Lock()
	defer j.mu.Unlock()

	return j.name, j.done != nil
}

func (j *jobSlot) status() proto.FileJobRsp {
	j.mu.Lock()
	defer j.mu.Unlock()

	state := j.state
	if state == "" {
		state = jobStateIdle
	}
	return proto.FileJobRsp{
		Status:     string(state),
		Name:       j.name,
		Percentage: utils.FormatPercentage(j.loaded, j.total),
		Loaded:     j.loaded,
		Total:      j.total,
		Message:    j.message,
		Queued:     len(j.waiters),
	}
}
