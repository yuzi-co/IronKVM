package common

// CodecSupported reports whether the capture library can deliver a codec, by
// libkvm's numbering: 0 MJPEG, CodecH264, CodecH265.
//
// Sipeed's library encodes all three. libkvm-v4l2, the library for the
// mainline kernel, has no H.265 encoder behind it and answers -2 to every
// H.265 read. A board restored from the vendor image with H.265 chosen then
// asked for H.265 on every frame, in direct and in WebRTC mode alike, and the
// web UI showed no video with nothing in the log after the first line, since
// a repeated read failure is logged once.
//
// libkvm-v4l2 answers through kvmv_codec_supported, which is not part of
// kvm_vision.h and which Sipeed's library does not have. It is looked up at
// run time, so one server binary works with either library; without it, all
// three codecs are assumed, which is what Sipeed's library does.
func CodecSupported(codec uint8) bool {
	return codecSupported(codec)
}

// codecSupported is a variable so a test can stand in for a library that
// lacks a codec. Off-device it is the stub's answer, which is Sipeed's.
var codecSupported = libraryCodecSupported

// vendorCodecSupported is the answer for a library without
// kvmv_codec_supported: Sipeed's encodes MJPEG, H.264 and H.265.
func vendorCodecSupported(codec uint8) bool {
	return codec <= CodecH265
}
