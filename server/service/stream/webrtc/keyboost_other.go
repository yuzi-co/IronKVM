//go:build !linux

package webrtc

// keyBoost does nothing off Linux: thread priorities are a Linux matter here.
type keyBoost struct{}

func newKeyBoost() *keyBoost { return &keyBoost{} }

func (b *keyBoost) raise() {}

func (b *keyBoost) lower() {}
