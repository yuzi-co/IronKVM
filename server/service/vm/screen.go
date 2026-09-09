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
