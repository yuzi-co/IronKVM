//go:build novision

package audio

import "errors"

// errNoEncoder is what a build without the device libraries reports.
//
// The pipeline still builds and is still tested here: every test supplies its
// own Encoder. Only the codec is missing, and a workstation has no UAC1 gadget
// to capture from either, so Available reports false long before this matters.
var errNoEncoder = errors.New("audio: this build has no Opus encoder")

func newOpusEncoder() (Encoder, error) {
	return nil, errNoEncoder
}
