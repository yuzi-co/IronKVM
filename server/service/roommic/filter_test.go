package roommic

import (
	"encoding/binary"
	"math"
	"testing"
)

const testRate = 48000.0

// sine is one second of a tone as mono S16_LE.
func sine(freq float64, amplitude float64) []byte {
	pcm := make([]byte, int(testRate)*2)
	for i := 0; i < int(testRate); i++ {
		v := amplitude * math.Sin(2*math.Pi*freq*float64(i)/testRate)
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(int16(v)))
	}
	return pcm
}

// rms of the second half, after the filter has settled.
func rms(pcm []byte) float64 {
	var sum float64
	n := len(pcm) / 2
	for i := n / 2; i < n; i++ {
		v := float64(int16(binary.LittleEndian.Uint16(pcm[i*2:])))
		sum += v * v
	}
	return math.Sqrt(sum / float64(n-n/2))
}

func gainDB(freq float64) float64 {
	in := sine(freq, 10000)
	before := rms(in)

	filter := NewHighPass(testRate, HighPassCutoff)
	// In 20 ms chunks, the way the stream hands them over, so the state
	// carried between calls is exercised too.
	for off := 0; off < len(in); off += 1920 {
		filter.Process(in[off : off+1920])
	}

	return 20 * math.Log10(rms(in)/before)
}

func TestTheFilterCutsHum(t *testing.T) {
	if g := gainDB(50); g > -25 {
		t.Fatalf("50 Hz is down %.1f dB, want at least 25", -g)
	}
	if g := gainDB(30); g > -40 {
		t.Fatalf("30 Hz is down %.1f dB, want at least 40", -g)
	}
}

func TestTheFilterKeepsTheVoiceBand(t *testing.T) {
	for _, freq := range []float64{300, 1000, 3000, 8000} {
		if g := gainDB(freq); math.Abs(g) > 1 {
			t.Errorf("%v Hz changed by %.2f dB, want within 1", freq, g)
		}
	}
}

func TestTheCutoffIsWhereItSays(t *testing.T) {
	// A Butterworth is 3 dB down at its cutoff.
	if g := gainDB(HighPassCutoff); g < -4 || g > -2 {
		t.Fatalf("%v Hz is at %.2f dB, want about -3", HighPassCutoff, g)
	}
}

func TestTheFilterRemovesAnOffset(t *testing.T) {
	pcm := make([]byte, int(testRate)*2)
	for i := 0; i < int(testRate); i++ {
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(int16(3000)))
	}

	NewHighPass(testRate, HighPassCutoff).Process(pcm)

	if got := rms(pcm); got > 1 {
		t.Fatalf("a constant offset leaves %.1f RMS, want none", got)
	}
}

// A full-scale square wave rings past full scale through a high-pass. The
// output has to saturate rather than wrap around.
func TestTheFilterClipsRatherThanWraps(t *testing.T) {
	pcm := make([]byte, 4800*2)
	for i := 0; i < 4800; i++ {
		v := int16(math.MaxInt16)
		if (i/240)%2 == 1 {
			v = math.MinInt16
		}
		binary.LittleEndian.PutUint16(pcm[i*2:], uint16(v))
	}

	NewHighPass(testRate, HighPassCutoff).Process(pcm)

	// Right after the first edge down the input is at full negative scale and
	// the output must be strongly negative too, not wrapped to positive.
	sample := int16(binary.LittleEndian.Uint16(pcm[241*2:]))
	if sample > -20000 {
		t.Fatalf("after a falling edge the output is %d, want strongly negative", sample)
	}
}

func TestAnOddByteIsLeftAlone(t *testing.T) {
	pcm := []byte{0, 0, 0x7f}
	NewHighPass(testRate, HighPassCutoff).Process(pcm)
	if pcm[2] != 0x7f {
		t.Fatal("the trailing byte changed")
	}
}
