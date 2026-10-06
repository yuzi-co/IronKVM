package vm

import (
	"NanoKVM-Server/common"
	"fmt"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/proto"
)

// GetScreen reports the capture settings the server holds.
//
// The browser needs this to draw the menu. Without it the menu had to remember
// its own choices and push them on load, so two browsers disagreed with each
// other and both of them could disagree with the board.
//
// CheckScreen runs first because the answer has to be the values a stream would
// use. The settings files are plain text on the card, and a menu that offered a
// value the capture loop repairs would tick an option the board is not running.
func (s *Service) GetScreen(c *gin.Context) {
	var rsp proto.Response

	common.CheckScreen()
	values := common.GetScreen().Snapshot()

	rsp.OkRspWithData(c, &proto.GetScreenRsp{
		Width:   values.Width,
		Height:  values.Height,
		Quality: values.Quality,
		BitRate: values.BitRate,
		FPS:     values.FPS,
		GOP:     values.GOP,
		Codec:   values.Codec,

		Aspect:          values.Aspect,
		AspectSupported: common.AspectSupported(),
	})
}

func (s *Service) SetScreen(c *gin.Context) {
	var req proto.SetScreenReq
	var rsp proto.Response

	err := proto.ParseFormRequest(c, &req)
	if err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	// Refuse a codec the capture library cannot deliver (H.265 with
	// libkvm-v4l2) before it reaches the card. common.SetScreen would ignore
	// it, so answering OK would tell the menu a change was made that was not.
	if req.Type == "codec" && (req.Value == common.CodecH264 || req.Value == common.CodecH265) &&
		!common.CodecSupported(uint8(req.Value)) {
		rsp.ErrRsp(c, -3, "codec not supported by this capture library")
		return
	}

	// The aspect choice exists only with a library that can keep the shape
	// (libkvm-v4l2); Sipeed's always stretches.
	if req.Type == "aspect" {
		if req.Value != common.AspectKeep && req.Value != common.AspectStretch {
			rsp.ErrRsp(c, -1, "invalid arguments")
			return
		}
		if !common.AspectSupported() {
			rsp.ErrRsp(c, -3, "aspect ratio not supported by this capture library")
			return
		}
	}

	switch req.Type {
	case "type":
		data := "h264"
		if req.Value == 0 {
			data = "mjpeg"
		}
		err = writeScreen("type", data)

	default:
		data := strconv.Itoa(req.Value)
		err = writeScreen(req.Type, data)
	}

	if err != nil {
		rsp.ErrRsp(c, -2, "update screen failed")
		return
	}

	common.SetScreen(req.Type, req.Value)

	log.Debugf("update screen: %+v", req)
	rsp.OkRsp(c)
}

func writeScreen(key string, value string) error {
	// The same map the restore at startup reads, so a setting cannot be stored
	// in one place and looked for in another.
	file, ok := common.ScreenFileMap[key]
	if !ok {
		return fmt.Errorf("invalid argument %s", key)
	}

	err := os.WriteFile(file, []byte(value), 0o666)
	if err != nil {
		log.Errorf("write kvm %s failed: %s", file, err)
		return err
	}

	return nil
}
