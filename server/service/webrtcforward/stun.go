package webrtcforward

import (
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"

	"NanoKVM-Server/config"
	"NanoKVM-Server/proto"
	"NanoKVM-Server/service/stream/webrtc"
)

// The STUN server setting, under /api/network/webrtc/stun (ironkvm-dist#157).
// Each WebRTC session reads it when it starts, so a save applies to the next
// viewer without a restart.

// saveStunMutex serialises saves, which read and rewrite server.yaml.
var saveStunMutex sync.Mutex

// saveStun writes the setting to server.yaml. A var for the tests.
var saveStun = func(stun string) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}
	conf.Stun = stun

	return config.Write(conf)
}

func (s *Service) GetStun(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, stunState(webrtc.CurrentStun()))
}

func (s *Service) SetStun(c *gin.Context) {
	var req proto.SetStunServerReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	stun := strings.TrimSpace(req.Stun)
	if stun == "" {
		stun = config.DefaultStun
	}
	if err := webrtc.ValidateStun(stun); err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	saveStunMutex.Lock()
	defer saveStunMutex.Unlock()

	if err := saveStun(stun); err != nil {
		log.Errorf("webrtc: failed to save the STUN server: %s", err)
		rsp.ErrRsp(c, -3, "failed to save the setting")
		return
	}
	webrtc.SetStun(stun)

	log.Infof("webrtc: STUN server set to %q", stun)
	rsp.OkRspWithData(c, stunState(stun))
}

func stunState(stun string) *proto.StunServerRsp {
	if stun == "" {
		stun = config.DefaultStun
	}

	return &proto.StunServerRsp{Stun: stun, Default: config.DefaultStun}
}
