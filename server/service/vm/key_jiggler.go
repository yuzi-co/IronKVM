package vm

import (
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/vm/jiggler"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// GetKeyJiggler reports whether the key jiggler is on and the key it presses.
func (s *Service) GetKeyJiggler(c *gin.Context) {
	var rsp proto.Response

	enabled, key := jiggler.GetJiggler().KeyJiggler()

	rsp.OkRspWithData(c, &proto.GetKeyJigglerRsp{
		Enabled: enabled,
		Key:     key,
	})
}

// SetKeyJiggler turns the key jiggler on or off and, when the request names
// one, changes its key. The key is saved either way, so it can be picked while
// the key jiggler is off.
func (s *Service) SetKeyJiggler(c *gin.Context) {
	var req proto.SetKeyJigglerReq
	var rsp proto.Response

	err := proto.ParseFormRequest(c, &req)
	if err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	keyJiggler := jiggler.GetJiggler()
	key := req.Key
	if key == "" {
		_, key = keyJiggler.KeyJiggler()
	}

	if err := keyJiggler.SetKeyJiggler(req.Enabled, key); err != nil {
		rsp.ErrRsp(c, -2, "operation failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("set key jiggler: %t, key %q", req.Enabled, key)
}
