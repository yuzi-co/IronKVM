package mjpeg

import (
	"sync"
	"time"

	"NanoKVM-Server/common"
	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const FrameDetectInterval uint8 = 60

const (
	defaultPauseDuration = 10 * time.Second
	maxPauseDuration     = 5 * time.Minute
)

// setFrameDetect is a variable so the pause logic can be tested without the
// capture hardware.
var setFrameDetect = func(frames uint8) {
	common.GetKvmVision().SetFrameDetect(frames)
}

// pauseDuration is how long detection stays off for a request. The value comes
// from the client, and detection is what notices the screen has changed, so an
// unbounded one switches the feature off for as long as the caller likes.
func pauseDuration(seconds int) time.Duration {
	if seconds <= 0 {
		return defaultPauseDuration
	}

	duration := time.Duration(seconds) * time.Second
	if duration > maxPauseDuration || duration < 0 {
		return maxPauseDuration
	}

	return duration
}

var (
	pauseMutex sync.Mutex
	pauseTimer *time.Timer
	pauseUntil time.Time

	// detectEnabled is the operator's setting. It is board-wide, so it lives
	// here rather than in one browser, and a pause returns to it rather than
	// to on. The capture library starts with detection off.
	detectEnabled bool
)

func framesFor(enabled bool) uint8 {
	if enabled {
		return FrameDetectInterval
	}
	return 0
}

// frameDetectEnabled reports the operator's setting.
func frameDetectEnabled() bool {
	pauseMutex.Lock()
	defer pauseMutex.Unlock()

	return detectEnabled
}

// setFrameDetectEnabled records the setting and applies it at once. It
// overrides a pause still in flight, so the pause's timer must not come along
// later and undo it.
func setFrameDetectEnabled(enabled bool) {
	pauseMutex.Lock()
	defer pauseMutex.Unlock()

	stopPauseLocked()
	detectEnabled = enabled
	setFrameDetect(framesFor(enabled))
}

// pauseFrameDetect switches detection off and schedules it back on, without
// holding the request open for the duration.
//
// Overlapping requests extend the pause but never shorten it: a caller that
// asked for a minute must not have it cut to a second by whoever asks next.
func pauseFrameDetect(duration time.Duration) {
	pauseMutex.Lock()
	defer pauseMutex.Unlock()

	// Detection that is off has nothing to pause.
	if !detectEnabled {
		return
	}

	deadline := time.Now().Add(duration)
	if !deadline.After(pauseUntil) {
		// Already paused at least this long, and already switched off.
		return
	}

	pauseUntil = deadline
	setFrameDetect(0)

	if pauseTimer != nil {
		pauseTimer.Stop()
	}
	pauseTimer = time.AfterFunc(duration, resumeFrameDetect)
}

func resumeFrameDetect() {
	pauseMutex.Lock()
	defer pauseMutex.Unlock()

	// A later request extended the pause after this timer was armed; the timer
	// it armed is the one that gets to resume.
	if time.Now().Before(pauseUntil) {
		return
	}

	pauseTimer = nil
	pauseUntil = time.Time{}
	setFrameDetect(framesFor(detectEnabled))
}

func stopPauseLocked() {
	if pauseTimer != nil {
		pauseTimer.Stop()
		pauseTimer = nil
	}
	pauseUntil = time.Time{}
}

// resetFrameDetectPause drops any pause in flight, for tests.
func resetFrameDetectPause() {
	pauseMutex.Lock()
	defer pauseMutex.Unlock()

	stopPauseLocked()
}

// GetFrameDetect answers with the current setting, so every browser shows the
// board's state instead of what it last chose itself.
func GetFrameDetect(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, &proto.GetFrameDetectRsp{Enabled: frameDetectEnabled()})
}

func UpdateFrameDetect(c *gin.Context) {
	var req proto.UpdateFrameDetectReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}

	setFrameDetectEnabled(req.Enabled)

	rsp.OkRsp(c)
	log.Debugf("update frame detect: %t", req.Enabled)
}

func StopFrameDetect(c *gin.Context) {
	var req proto.StopFrameDetectReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid parameters")
		return
	}

	pauseFrameDetect(pauseDuration(req.Duration))

	rsp.OkRsp(c)
}
