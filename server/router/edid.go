package router

import (
	"encoding/hex"
	"errors"
	"io"
	"net/http"

	"NanoKVM-Server/middleware"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/edid"

	"github.com/gin-gonic/gin"
)

// edidRouter registers the read side of EDID management (ironkvm-dist#95):
// the board's EDID as far as it can be known, the preset library, and a
// decoder for an uploaded file. Nothing here writes an EDID anywhere; that
// is parked until an operator decides on writing the bridge's flash.
func edidRouter(r *gin.Engine) {
	api := r.Group("/api/edid").Use(middleware.CheckToken())

	api.GET("/current", getCurrentEdid)             // what the board presents, if readable
	api.GET("/presets", listEdidPresets)            // the preset library
	api.GET("/presets/:id", getEdidPreset)          // one preset, raw and decoded
	api.GET("/presets/:id/bin", downloadEdidPreset) // one preset as a .bin file
	api.POST("/decode", decodeEdid)                 // decode an upload; nothing is applied
}

// edidRsp is one EDID, as hex and decoded.
type edidRsp struct {
	Hex     string     `json:"hex"`
	Decoded *edid.Info `json:"decoded"`
}

// maxEdidUpload bounds a decode request: the largest EDID as hex text, with
// room for a separator after every digit pair.
const maxEdidUpload = 3 * edid.MaxSize

func getCurrentEdid(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, edid.Current())
}

func listEdidPresets(c *gin.Context) {
	var rsp proto.Response
	rsp.OkRspWithData(c, gin.H{"presets": edid.Presets()})
}

func getEdidPreset(c *gin.Context) {
	var rsp proto.Response

	data, err := edid.PresetData(c.Param("id"))
	if err != nil {
		rsp.ErrRsp(c, -1, err.Error())
		return
	}
	info, err := edid.Parse(data)
	if err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}
	rsp.OkRspWithData(c, edidRsp{Hex: hex.EncodeToString(data), Decoded: info})
}

func downloadEdidPreset(c *gin.Context) {
	id := c.Param("id")
	data, err := edid.PresetData(id)
	if err != nil {
		c.String(http.StatusNotFound, err.Error())
		return
	}
	c.Header("Content-Disposition", `attachment; filename="edid-`+id+`.bin"`)
	c.Header("X-Content-Type-Options", "nosniff")
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// decodeEdid decodes the request body, a binary EDID or its hex text. An
// EDID with problems (a bad checksum, say) is still decoded and answered
// with code 0; valid:false and the problems say what is wrong.
func decodeEdid(c *gin.Context) {
	var rsp proto.Response

	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxEdidUpload+1))
	if err != nil {
		rsp.ErrRsp(c, -1, "failed to read the upload")
		return
	}
	if len(body) > maxEdidUpload {
		rsp.ErrRsp(c, -2, edid.ErrTooLarge.Error())
		return
	}
	data, err := edid.DecodeInput(body)
	if err != nil {
		rsp.ErrRsp(c, -3, err.Error())
		return
	}
	info, err := edid.Parse(data)
	if err != nil {
		code := -3
		if errors.Is(err, edid.ErrTooLarge) {
			code = -2
		}
		rsp.ErrRsp(c, code, err.Error())
		return
	}
	rsp.OkRspWithData(c, edidRsp{Hex: hex.EncodeToString(data), Decoded: info})
}
