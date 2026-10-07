package roommic

import (
	"encoding/binary"
	"math"
)

// HighPassCutoff is where the filter starts to cut. The room's floor at full
// gain is mostly hum below 500 Hz, its strongest part under 50 Hz, and speech
// has little below 100 Hz worth keeping.
const HighPassCutoff = 120.0

// biquad is one second-order section in transposed direct form II.
type biquad struct {
	b0, b1, b2, a1, a2 float64
	z1, z2             float64
}

func (q *biquad) step(x float64) float64 {
	y := q.b0*x + q.z1
	q.z1 = q.b1*x - q.a1*y + q.z2
	q.z2 = q.b2*x - q.a2*y
	return y
}

// newHighPassSection is the RBJ cookbook high-pass at cutoff with quality q.
func newHighPassSection(sampleRate, cutoff, q float64) biquad {
	w0 := 2 * math.Pi * cutoff / sampleRate
	cos := math.Cos(w0)
	alpha := math.Sin(w0) / (2 * q)
	a0 := 1 + alpha

	return biquad{
		b0: (1 + cos) / 2 / a0,
		b1: -(1 + cos) / a0,
		b2: (1 + cos) / 2 / a0,
		a1: -2 * cos / a0,
		a2: (1 - alpha) / a0,
	}
}

// HighPass is a fourth-order Butterworth high-pass, two sections, for one
// channel of S16_LE. 24 dB an octave takes 50 Hz hum down by about 30 dB at a
// 120 Hz cutoff and leaves the voice band flat.
//
// It keeps state across calls, so one HighPass belongs to one capture.
type HighPass struct {
	sections [2]biquad
}

// NewHighPass builds the filter for a sample rate and cutoff in Hz.
func NewHighPass(sampleRate, cutoff float64) *HighPass {
	// The Q of each section of a fourth-order Butterworth.
	return &HighPass{sections: [2]biquad{
		newHighPassSection(sampleRate, cutoff, 0.54119610),
		newHighPassSection(sampleRate, cutoff, 1.30656296),
	}}
}

// Process filters a chunk of mono S16_LE in place. A trailing odd byte is left
// alone.
func (h *HighPass) Process(pcm []byte) {
	for i := 0; i+1 < len(pcm); i += 2 {
		x := float64(int16(binary.LittleEndian.Uint16(pcm[i:])))
		for s := range h.sections {
			x = h.sections[s].step(x)
		}

		binary.LittleEndian.PutUint16(pcm[i:], uint16(clamp16(x)))
	}
}

func clamp16(x float64) int16 {
	switch {
	case x >= math.MaxInt16:
		return math.MaxInt16
	case x <= math.MinInt16:
		return math.MinInt16
	default:
		return int16(math.Round(x))
	}
}
