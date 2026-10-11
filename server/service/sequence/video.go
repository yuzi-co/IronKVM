package sequence

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"
)

// The files the capture library keeps the HDMI source's size in, the same
// ones GET /api/stream/info reads. Variables so the tests can point them
// elsewhere.
var (
	videoWidthFile  = "/kvmapp/kvm/width"
	videoHeightFile = "/kvmapp/kvm/height"
)

// defaultVideoPoll is how often a run that stops on a video change reads the
// mode. Two small reads four times a second cost nothing next to the capture,
// and a firmware menu stays up far longer than a quarter of a second.
const defaultVideoPoll = 250 * time.Millisecond

type videoMode struct {
	width, height int
}

func readVideoMode() videoMode {
	return videoMode{width: readPositive(videoWidthFile), height: readPositive(videoHeightFile)}
}

// watchVideo ends the run when the source's mode changes from the one it had
// when the run started: the host switched from its desktop to the firmware's
// screen, or from no signal to any. A mode that reads as zero, a capture that
// has not published one yet, is not a change.
//
// This only sees the size. A firmware menu drawn at the size of the boot logo
// before it is not seen, which is why the option is off by default.
func (m *Manager) watchVideo(ctx context.Context, cancel context.CancelCauseFunc) {
	first := m.videoMode()
	ticker := time.NewTicker(m.videoPoll)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		mode := m.videoMode()
		if mode.width == 0 || mode.height == 0 || mode == first {
			continue
		}
		cancel(errVideo)
		return
	}
}

func readPositive(path string) int {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil || value <= 0 {
		return 0
	}
	return value
}
