package vm

import (
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/vm/jiggler"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

const (
	jigglerMethodMouse = "mouse"
	jigglerMethodKey   = "key"
)

func (s *Service) GetMouseJiggler(c *gin.Context) {
	var rsp proto.Response

	mouseJiggler := jiggler.GetJiggler()

	key := mouseJiggler.GetKey()
	method := jigglerMethodMouse
	if key != "" {
		method = jigglerMethodKey
	}

	data := &proto.GetMouseJigglerRsp{
		Enabled: mouseJiggler.IsEnabled(),
		Mode:    mouseJiggler.GetMode(),
		Method:  method,
		Key:     key,
	}

	rsp.OkRspWithData(c, data)
}

// jigglerKey turns a request's method and key into the jiggler's key, empty
// for the mouse. A request that names no method keeps the current one.
func jigglerKey(method string, key string, current string) string {
	switch method {
	case jigglerMethodMouse:
		return ""
	case jigglerMethodKey:
		if key == "" {
			return jiggler.DefaultKey
		}
		return key
	default:
		return current
	}
}

func (s *Service) SetMouseJiggler(c *gin.Context) {
	var req proto.SetMouseJigglerReq
	var rsp proto.Response

	err := proto.ParseFormRequest(c, &req)
	if err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	mouseJiggler := jiggler.GetJiggler()
	key := jigglerKey(req.Method, req.Key, mouseJiggler.GetKey())

	if req.Enabled {
		err = mouseJiggler.Enable(req.Mode, key)
	} else {
		err = mouseJiggler.Disable()
		if err == nil {
			err = mouseJiggler.SetKey(key)
		}
	}

	if err != nil {
		rsp.ErrRsp(c, -2, "operation failed")
		return
	}

	rsp.OkRsp(c)
	log.Debugf("set mouse jiggler: %t, key %q", req.Enabled, key)
}
