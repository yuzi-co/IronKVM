package webrtc

import "unsafe"

func unsafePointer(b *byte) unsafe.Pointer { return unsafe.Pointer(b) }
