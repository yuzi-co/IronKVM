package webrtc

import "sync/atomic"

// What the viewers' retransmission requests asked for, and what the viewers'
// queues gave up, since the server started (ironkvm-dist #91). /api/metrics
// reports them.
var (
	// nackResent counts NACKed packets this viewer was sent and the responder
	// still holds: it sends them again.
	nackResent atomic.Uint64
	// nackNotSent counts NACKed packets of frames this viewer's queue gave
	// up: never sent to it, so nothing to resend.
	nackNotSent atomic.Uint64
	// nackExpired counts NACKed packets older than the responder keeps
	// (nackResponderSize).
	nackExpired atomic.Uint64

	// closedDropped is the frames given up by the queues of viewers that
	// have gone. DroppedFrames adds the live viewers' own.
	closedDropped atomic.Uint64
)

// NackCounts returns the NACKed packets by outcome: resent, never sent to that
// viewer, and too old to resend.
func NackCounts() (resent, notSent, expired uint64) {
	return nackResent.Load(), nackNotSent.Load(), nackExpired.Load()
}

// DroppedFrames is the number of video frames the viewers' queues have given
// up because the viewer was behind.
func DroppedFrames() uint64 {
	total := closedDropped.Load()
	for _, client := range getManager().getClients() {
		total += client.queue.Dropped()
	}

	return total
}
