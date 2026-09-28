// Package screenshot serves one full frame of the host screen to the web UI.
//
// The browser reads text out of the frame itself (OCR). It asks the board for
// the frame rather than drawing the video element to a canvas, because the
// same request then works on every delivery path: the MJPEG image, the WebRTC
// video and the H.264 direct canvas each need their own capture code, and the
// frame the board sends does not depend on how this browser scales or crops
// the picture.
package screenshot

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"NanoKVM-Server/proto"
	mcpservice "NanoKVM-Server/service/mcp"
)

// quality is the highest the snapshotter allows. Text survives JPEG better as
// the quality rises, and the MCP tool that set the ceiling has the same need.
const quality = 60

// Handler answers with the frame as image/jpeg, or with the usual error
// envelope when there is no frame to give.
//
// It takes the frame through the same snapshotter as the MCP screenshot tool
// and the watchdog: at the stream's current size, so a viewer's stream is never
// resized, and under the capture lease, so capture resumes if the idle timer
// stopped it.
func Handler(snapshotter mcpservice.Snapshotter) gin.HandlerFunc {
	return func(c *gin.Context) {
		var rsp proto.Response

		snap, err := snapshotter.Capture(c.Request.Context(), mcpservice.SnapshotRequest{Quality: quality})
		if err != nil {
			rsp.ErrRsp(c, -1, err.Error())
			return
		}
		if !snap.OK {
			message := snap.Message
			if message == "" {
				message = "failed to capture screenshot"
			}
			rsp.ErrRsp(c, -2, message)
			return
		}

		// Each request is a new picture of a screen that keeps changing.
		c.Header("Cache-Control", "no-store")
		c.Data(http.StatusOK, "image/jpeg", snap.JPEG)
	}
}
