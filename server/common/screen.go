package common

import (
	"os"
	"strconv"
	"strings"
	"sync"
)

// ScreenFileMap names the file each setting is stored in.
//
// These are the operator's choices rather than runtime state, so they stay on
// the card: a board set to 60 frames should still be at 60 frames after a
// reboot. They are written once per change, which is what makes that
// acceptable where `now_fps` and `wifi_state` are not - see the comment about
// tmpfs in `kvmapp/system/init.d/S95nanokvm`.
//
// `gop` is absent on purpose. The API hands it straight to libkvm and stores
// nothing, so there is nothing to restore.
//
// It is a variable so a test can point it at a temporary directory. Nothing on
// the device changes it.
var ScreenFileMap = map[string]string{
	"type":       "/kvmapp/kvm/type",
	"fps":        "/kvmapp/kvm/fps",
	"quality":    "/kvmapp/kvm/qlty",
	"resolution": "/kvmapp/kvm/res",
	"codec":      "/kvmapp/kvm/codec",
}

// The video codecs the encoder implements. These are libkvm's public numbering
// and not mmf's, which runs the other way: libkvm converts.
const (
	CodecH264 = 1
	CodecH265 = 2
)

// defaultScreenValues is what a board serves when it has never been configured.
var defaultScreenValues = ScreenValues{
	Width:   0,
	Height:  0,
	Quality: 80,
	FPS:     30,
	BitRate: 3000,
	GOP:     30,
	Codec:   CodecH264,
}

// ScreenValues is a consistent copy of the capture parameters.
type ScreenValues struct {
	Width   uint16
	Height  uint16
	FPS     int
	Quality uint16
	BitRate uint16
	GOP     uint8
	// Codec applies to both H.264 delivery paths. There is one hardware
	// encoder, so this cannot be a per-viewer choice: changing it rebuilds
	// the VENC channel out from under every viewer at once.
	Codec uint8
}

// Screen holds the capture parameters. HTTP handlers write them while the
// streamer goroutines read them on every frame, so access goes through the
// mutex and readers take a snapshot instead of reading field by field.
type Screen struct {
	mutex  sync.RWMutex
	values ScreenValues
}

var (
	screen     *Screen
	screenOnce sync.Once
)

// ResolutionMap height to width
var ResolutionMap = map[uint16]uint16{
	1080: 1920,
	720:  1280,
	600:  800,
	480:  640,
	0:    0,
}

var QualityMap = map[uint16]bool{
	100: true,
	80:  true,
	60:  true,
	50:  true,
}

var BitRateMap = map[uint16]bool{
	5000: true,
	3000: true,
	2000: true,
	1000: true,
}

func GetScreen() *Screen {
	screenOnce.Do(func() {
		screen = &Screen{values: loadScreenValues()}
	})

	return screen
}

// loadScreenValues seeds the singleton from what the operator last chose.
//
// The settings files outlive the process, and nothing read them back: every
// start served 30 frames at quality 80 whatever the board was set to, until
// somebody opened the UI and changed a setting. The browser kept its own copy
// in localStorage, so the page went on showing 60 while the server sent 30, and
// a second browser saw neither.
//
// A file that is missing, unreadable or not a number leaves its default alone.
// That is the same answer for all three, and none of them is worth a log line
// on a board where an unconfigured setting is the ordinary case.
func loadScreenValues() ScreenValues {
	values := defaultScreenValues

	// Resolution first, so a stored quality that the resolution constrains is
	// applied against the right one. The order also matches the switch below.
	for _, key := range []string{"resolution", "quality", "fps", "codec"} {
		if value, ok := readScreenSetting(key); ok {
			applyScreenValue(&values, key, value)
		}
	}

	return values
}

func readScreenSetting(key string) (int, bool) {
	path, ok := ScreenFileMap[key]
	if !ok {
		return 0, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}

	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}

	return value, true
}

// Snapshot returns the current parameters as a single consistent copy.
func (s *Screen) Snapshot() ScreenValues {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.values
}

func SetScreen(key string, value int) {
	s := GetScreen()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	applyScreenValue(&s.values, key, value)
}

// applyScreenValue is shared by the API and by the restore at startup, so a
// stored setting is read back under exactly the rule that wrote it. The
// overloaded quality field is the reason that matters: one API key carries
// either a JPEG quality or an H.264 bitrate, and both land in the same file.
func applyScreenValue(values *ScreenValues, key string, value int) {
	switch key {
	case "resolution":
		height := uint16(value)
		if width, ok := ResolutionMap[height]; ok {
			values.Width = width
			values.Height = height
		}

	case "quality":
		if value > 100 {
			values.BitRate = uint16(value)
		} else {
			values.Quality = uint16(value)
		}

	case "fps":
		values.FPS = validateFPS(value)

	case "gop":
		values.GOP = uint8(value)

	case "codec":
		// Anything else is left alone rather than stored. A codec libkvm does
		// not implement reaches mmf_add_venc_channel, which answers -1, and
		// before 2026-09-09 that was an uncaught C++ exception across cgo and
		// so the whole server. It is an error return now, and it still has no
		// business getting that far.
		if value == CodecH264 || value == CodecH265 {
			values.Codec = uint8(value)
		}
	}
}

func CheckScreen() {
	s := GetScreen()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := ResolutionMap[s.values.Height]; !ok {
		s.values.Width = 1920
		s.values.Height = 1080
	}

	if _, ok := QualityMap[s.values.Quality]; !ok {
		s.values.Quality = 80
	}

	if _, ok := BitRateMap[s.values.BitRate]; !ok {
		s.values.BitRate = 3000
	}

	s.values.Codec = validateCodec(s.values.Codec)
}

// validateCodec keeps an unusable codec away from the encoder. The settings
// files are plain text on the card and a person can edit them, so the value
// read back is not necessarily one this build knows.
func validateCodec(codec uint8) uint8 {
	if codec == CodecH264 || codec == CodecH265 {
		return codec
	}

	return CodecH264
}

func validateFPS(fps int) int {
	if fps > 60 {
		return 60
	}
	if fps < 10 {
		return 10
	}

	return fps
}
