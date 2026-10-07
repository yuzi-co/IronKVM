package webrtc

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	log "github.com/sirupsen/logrus"
)

// keyNiceEnv sets the nice value a video writer's thread runs at while it
// writes a keyframe. Unset, defaultKeyNice; "0" or "off" leaves the priority
// alone.
//
// Trial 63 (#72): a 1080p keyframe is 80 to 120 RTP packets, and on slot B the
// writer took 126 to 138 ms to send one, a third of it lost to switching with
// the capture loop on the one core. At nice -10 for the keyframe alone the
// write took about a third less, slot B's p95 delay fell in every case
// measured, and neither slot lost frame rate. The same value for every frame
// cost slot A 1 to 5 fps, and -20 cost slot B frames, so it is -10 and only
// for keyframes.
const keyNiceEnv = "KVM_WEBRTC_KEY_NICE"

const defaultKeyNice = -10

// keyNice is the configured value, 0 for none.
func keyNice() int {
	v := strings.TrimSpace(os.Getenv(keyNiceEnv))
	switch v {
	case "":
		return defaultKeyNice
	case "off":
		return 0
	}

	nice, err := strconv.Atoi(v)
	if err != nil {
		log.Warnf("webrtc: %s=%q is not a number; using %d", keyNiceEnv, v, defaultKeyNice)
		return defaultKeyNice
	}

	return min(max(nice, -20), 19)
}

// keyBoostOff is set after the first refusal, for every writer: a server that
// may not raise its priority will not be allowed on the next keyframe either.
var (
	keyBoostOff     atomic.Bool
	keyBoostRefused sync.Once
)

func keyBoostFailed(nice int, err error) {
	keyBoostOff.Store(true)
	keyBoostRefused.Do(func() {
		log.Infof("webrtc: cannot give the keyframe writer nice %d (%v); keyframes are written at normal priority", nice, err)
	})
}
