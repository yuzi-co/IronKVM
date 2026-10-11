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

// The WebRTC port forward setting, under /api/network/webrtc. This is its own
// package because the stream package it configures is imported by middleware,
// which service/network imports. The mechanism is
// in service/stream/webrtc/portforward.go.

// saveWebRTCMutex serialises saves, which read and rewrite server.yaml.
var saveWebRTCMutex sync.Mutex

// saveWebRTC writes the setting to server.yaml. A var for the tests.
var saveWebRTC = func(settings config.WebRTC) error {
	conf, err := config.Read()
	if err != nil {
		return err
	}
	conf.WebRTC = settings

	return config.Write(conf)
}

func (s *Service) Get(c *gin.Context) {
	var rsp proto.Response

	rsp.OkRspWithData(c, webrtcForwardState(webrtc.PortForwardSettings()))
}

func (s *Service) Set(c *gin.Context) {
	var req proto.SetWebRTCForwardReq
	var rsp proto.Response

	if err := proto.ParseFormRequest(c, &req); err != nil {
		rsp.ErrRsp(c, -1, "invalid arguments")
		return
	}

	settings := config.WebRTC{
		UDPPort:       req.UDPPort,
		PublicAddress: strings.TrimSpace(req.PublicAddress),
	}

	conf := config.GetInstance()
	if err := webrtc.ValidatePortForward(settings, conf.Port.Http, conf.Port.Https); err != nil {
		rsp.ErrRsp(c, -2, err.Error())
		return
	}

	saveWebRTCMutex.Lock()
	defer saveWebRTCMutex.Unlock()

	// The port is opened before the setting is saved, so that a port that is
	// taken is refused here and not saved as a setting that cannot work.
	if err := webrtc.BindPortForward(settings); err != nil {
		rsp.ErrRsp(c, -4, err.Error())
		return
	}
	if err := saveWebRTC(settings); err != nil {
		log.Errorf("webrtc: failed to save the port forward setting: %s", err)
		rsp.ErrRsp(c, -3, "failed to save the setting")
		return
	}
	conf.WebRTC = settings
	webrtc.SetPortForward(settings)

	log.Infof("webrtc: port forward set: udpPort=%d publicAddress=%q", settings.UDPPort, settings.PublicAddress)
	rsp.OkRspWithData(c, webrtcForwardState(settings))
}

func webrtcForwardState(settings config.WebRTC) *proto.GetWebRTCForwardRsp {
	resolved := webrtc.ResolvePublicAddress(settings.PublicAddress)
	if resolved == nil {
		resolved = []string{}
	}

	return &proto.GetWebRTCForwardRsp{
		UDPPort:       settings.UDPPort,
		PublicAddress: settings.PublicAddress,
		ActivePort:    webrtc.ActivePort(),
		Resolved:      resolved,
	}
}

// Service holds the handlers.
type Service struct{}

func NewService() *Service {
	return &Service{}
}
