package stream

import (
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"NanoKVM-Server/proto"

	"github.com/gin-gonic/gin"
)

// The files the capture library keeps the HDMI source's mode in. Both
// libraries write the size; only libkvm-v4l2 writes the refresh rate, in
// millihertz. Variables so the tests can point them elsewhere.
var (
	sourceWidthFile   = "/kvmapp/kvm/width"
	sourceHeightFile  = "/kvmapp/kvm/height"
	sourceRefreshFile = "/kvmapp/kvm/refresh"
)

// captureStamps says whether the last frame the capture loop read came with
// its capture time (H264Frame.HasTimes).
var captureStamps atomic.Bool

// Info answers GET /api/stream/info, which the web UI's status bar asks
// every two seconds while it shows: the source's mode as the capture library
// last published it, and the rate the capture loop delivers frames at. It
// reads two or three small files and a counter, and starts nothing; a board
// that streams to nobody reports a capture rate of -1.
func Info(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, readStreamInfo())
}

func readStreamInfo() proto.GetStreamInfoRsp {
	info := proto.GetStreamInfoRsp{CaptureFps: -1}

	info.Width = readPositive(sourceWidthFile)
	info.Height = readPositive(sourceHeightFile)
	info.Refresh = readRefresh()

	if fps, started := CurrentFPS(); started {
		info.CaptureFps = int(fps)
	}
	info.CaptureStamps = captureStamps.Load()

	return info
}

// readRefresh reads the refresh rate, if the library that wrote the size
// wrote it too. Sipeed's library writes the size and not the rate, so a rate
// file older than the size file is one an earlier libkvm-v4l2 left behind,
// about a mode that may have changed since.
func readRefresh() int {
	refresh, err := os.Stat(sourceRefreshFile)
	if err != nil {
		return 0
	}
	if width, err := os.Stat(sourceWidthFile); err != nil || refresh.ModTime().Before(width.ModTime()) {
		return 0
	}

	return readPositive(sourceRefreshFile)
}

// readPositive reads a file that holds one positive integer, or answers 0.
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
