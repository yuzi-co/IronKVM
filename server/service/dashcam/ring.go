package dashcam

import "time"

// Sample is one access unit in the ring: the encoder's output as the tap
// handed it on, split into NALUs. The NALUs are the capture loop's own
// memory, shared with every other reader, so the ring holds them without a
// copy and nobody may write to them.
type Sample struct {
	Codec uint8
	Key   bool
	NALUs [][]byte
	// PTS is the tap's clock: the capture time, measured from the tap's
	// start.
	PTS time.Duration
}

// sampleOverhead is what one sample costs beside its payload: the Sample
// itself and its share of the ring's slice, about 64 bytes, and a slice header
// for each NALU. The ring counts it, so a stream of tiny frames is bounded by
// the same number as a stream of large ones.
const (
	sampleOverhead = 64
	naluOverhead   = 24
)

// cost is what a sample holds in memory.
func (s Sample) cost() int {
	n := sampleOverhead
	for _, nalu := range s.NALUs {
		n += naluOverhead + len(nalu)
	}
	return n
}

// Ring keeps the most recent access units, oldest first, and always starts at
// a keyframe: a clip cut from it must open on a picture a player can decode.
//
// It is bounded twice. By time, it keeps whole GOPs as long as the GOPs after
// the first still cover the window, so it holds between the window and the
// window plus one GOP. By memory, it drops whole GOPs from the front until it
// fits in maxBytes; a single GOP that does not fit is dropped as a whole, and
// the ring waits for the next keyframe (Waiting answers true until then).
//
// A Ring is not safe for concurrent use. The service holds its mutex around
// every call.
type Ring struct {
	window   time.Duration
	maxBytes int

	samples []Sample
	// keys are the indexes in samples of the keyframes, in order. keys[0]
	// is always 0 while the ring holds anything.
	keys  []int
	bytes int
}

// NewRing returns an empty ring that keeps window and at most maxBytes.
func NewRing(window time.Duration, maxBytes int) *Ring {
	return &Ring{window: window, maxBytes: maxBytes}
}

// Push adds an access unit. A delta frame is refused while the ring is
// empty, since nothing before it could be decoded; Push reports whether it
// kept the sample. A sample whose PTS goes back starts the ring again: the
// tap's clock started over, and the old samples are on another timeline.
func (r *Ring) Push(s Sample) bool {
	if n := len(r.samples); n > 0 && s.PTS < r.samples[n-1].PTS {
		r.Reset()
	}
	if len(r.samples) == 0 && !s.Key {
		return false
	}

	if s.Key {
		r.keys = append(r.keys, len(r.samples))
	}
	r.samples = append(r.samples, s)
	r.bytes += s.cost()
	r.trim()
	return len(r.samples) > 0
}

// trim drops whole GOPs from the front while the ring is over either bound.
func (r *Ring) trim() {
	for len(r.keys) > 0 {
		newest := r.samples[len(r.samples)-1].PTS
		switch {
		case r.bytes > r.maxBytes && len(r.keys) == 1:
			// One GOP bigger than the budget: there is no keyframe to
			// start from that leaves less, so it goes as a whole.
			r.Reset()
			return
		case r.bytes > r.maxBytes:
			r.dropFirstGOP()
		case len(r.keys) > 1 && newest-r.samples[r.keys[1]].PTS >= r.window:
			// The GOPs after the first already cover the window.
			r.dropFirstGOP()
		default:
			return
		}
	}
}

func (r *Ring) dropFirstGOP() {
	cut := r.keys[1]
	for i := range cut {
		r.bytes -= r.samples[i].cost()
		// Clear the entry, so the slice's array no longer holds the
		// payload and the collector can take it.
		r.samples[i] = Sample{}
	}
	r.samples = r.samples[cut:]
	r.keys = r.keys[1:]
	for i := range r.keys {
		r.keys[i] -= cut
	}
}

// Reset empties the ring.
func (r *Ring) Reset() {
	clear(r.samples)
	r.samples = r.samples[:0]
	r.keys = r.keys[:0]
	r.bytes = 0
}

// SetBounds changes the window and the memory budget, and trims to them.
func (r *Ring) SetBounds(window time.Duration, maxBytes int) {
	r.window, r.maxBytes = window, maxBytes
	r.trim()
}

// Waiting reports whether the ring is empty and so waits for a keyframe.
func (r *Ring) Waiting() bool {
	return len(r.samples) == 0
}

// Len is how many access units the ring holds.
func (r *Ring) Len() int {
	return len(r.samples)
}

// Bytes is what the ring holds in memory, by the count Push keeps.
func (r *Ring) Bytes() int {
	return r.bytes
}

// Span is the time from the first sample to the last.
func (r *Ring) Span() time.Duration {
	if len(r.samples) == 0 {
		return 0
	}
	return r.samples[len(r.samples)-1].PTS - r.samples[0].PTS
}

// Snapshot copies the ring's samples into a new slice. The payloads are
// shared, not copied: they are never written to.
func (r *Ring) Snapshot() []Sample {
	out := make([]Sample, len(r.samples))
	copy(out, r.samples)
	return out
}
