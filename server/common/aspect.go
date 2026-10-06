package common

// How a source whose shape differs from the chosen resolution is fitted to it
// (ironkvm-dist#37). A 1920x1200 or 1024x768 host screen asked for at 1080 or
// 720 lines used to be stretched to 16:9: 1920x1200 came out as 1920x1080,
// 10% squashed. The browser scales whatever picture arrives, so the stream
// can simply keep the source's shape.
//
// The numbers are what /kvmapp/kvm/aspect holds. Keep is 0 so that a missing
// or zeroed file means the default.
const (
	// AspectKeep sends the source's shape at the chosen height, or at the
	// source's own height when that is smaller: 1920x1200 at 1080 lines is
	// 1728x1080.
	AspectKeep = 0
	// AspectStretch stretches the source to the chosen resolution, which is
	// what Sipeed's library always does.
	AspectStretch = 1
)

// AspectSupported reports whether the capture library can keep the source's
// aspect ratio. libkvm-v4l2 (slot B, mainline kernel) can; Sipeed's library
// cannot, and the setting has no effect with it.
func AspectSupported() bool {
	return aspectSupported()
}

// aspectSupported and setLibraryKeepAspect are variables so a test can stand
// in for either library.
var (
	aspectSupported      = libraryAspectSupported
	setLibraryKeepAspect = librarySetKeepAspect
)

func validAspect(value int) bool {
	return value == AspectKeep || value == AspectStretch
}

// applyAspect hands the setting to the library. The library rebuilds the
// encoder on its next read when the setting changed, so this is cheap to call
// with an unchanged value.
func applyAspect(aspect uint8) {
	setLibraryKeepAspect(aspect != AspectStretch)
}
