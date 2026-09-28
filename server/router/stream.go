package router

import (
	"NanoKVM-Server/middleware"
	kvmcapture "NanoKVM-Server/service/mcp/capture/kvm"
	"NanoKVM-Server/service/stream/direct"
	"NanoKVM-Server/service/stream/mjpeg"
	"NanoKVM-Server/service/stream/screenshot"
	"NanoKVM-Server/service/stream/webrtc"

	"github.com/gin-gonic/gin"
)

func streamRouter(r *gin.Engine) {
	api := r.Group("/api").Use(middleware.CheckToken())

	api.GET("/stream/mjpeg", mjpeg.Connect)                      // mjpeg stream
	api.POST("/stream/mjpeg/detect", mjpeg.UpdateFrameDetect)    // update frame detect
	api.POST("/stream/mjpeg/detect/stop", mjpeg.StopFrameDetect) // temporary stop frame detect

	api.GET("/stream/h264", webrtc.Connect)        // h264 stream (webrtc)
	api.GET("/stream/h264/direct", direct.Connect) // h264 stream (http)

	// One full frame as a JPEG, for the web UI's text recognition. Anyone who
	// can watch the stream can already see the same picture.
	api.GET("/stream/screenshot", screenshot.Handler(kvmcapture.New()))
}
