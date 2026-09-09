package common

import "testing"

// The codec is one global setting because there is one hardware encoder. Two
// viewers cannot be served different codecs at once, so this is stored and
// restored exactly like the resolution is.

func TestAnUnconfiguredBoardServesH264(t *testing.T) {
	values := defaultScreenValues

	if values.Codec != CodecH264 {
		t.Fatalf("default codec = %d, want H.264 (%d)", values.Codec, CodecH264)
	}
}

func TestTheCodecIsStoredOnTheCard(t *testing.T) {
	file, ok := ScreenFileMap["codec"]
	if !ok {
		t.Fatal("codec has no file in ScreenFileMap, so a choice would not survive a reboot")
	}
	if file != "/kvmapp/kvm/codec" {
		t.Fatalf("codec file = %q, want /kvmapp/kvm/codec", file)
	}
}

func TestSettingTheCodecTakesBothValues(t *testing.T) {
	for _, codec := range []int{CodecH264, CodecH265} {
		values := defaultScreenValues
		applyScreenValue(&values, "codec", codec)

		if int(values.Codec) != codec {
			t.Fatalf("applyScreenValue(codec=%d) gave %d", codec, values.Codec)
		}
	}
}

func TestAnUnknownCodecIsRefusedRatherThanStored(t *testing.T) {
	// A codec libkvm does not implement reaches mmf_add_venc_channel, which
	// answers -1. That used to be std::terminate. It is an error return now,
	// but a value that cannot work should never get as far as the encoder.
	for _, codec := range []int{0, 3, 255, -1} {
		values := defaultScreenValues
		values.Codec = CodecH265
		applyScreenValue(&values, "codec", codec)

		if values.Codec != CodecH265 {
			t.Fatalf("applyScreenValue(codec=%d) changed the codec to %d, want it left alone",
				codec, values.Codec)
		}
	}
}

func TestAnImpossibleStoredCodecIsRepairedToH264(t *testing.T) {
	// validateCodec mirrors validateFPS: a pure rule the restore and the
	// consistency check both use, so a hand-edited file cannot leave the
	// encoder asking for a codec that does not exist.
	for _, stored := range []uint8{0, 3, 9, 255} {
		if got := validateCodec(stored); got != CodecH264 {
			t.Fatalf("validateCodec(%d) = %d, want H.264 (%d)", stored, got, CodecH264)
		}
	}
	for _, stored := range []uint8{CodecH264, CodecH265} {
		if got := validateCodec(stored); got != stored {
			t.Fatalf("validateCodec(%d) = %d, want it left alone", stored, got)
		}
	}
}
