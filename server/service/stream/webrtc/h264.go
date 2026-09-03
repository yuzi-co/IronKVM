package webrtc

import (
	"NanoKVM-Server/config"
	"NanoKVM-Server/middleware"
	"encoding/json"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/pion/dtls/v3"
	"github.com/pion/interceptor"
	"github.com/pion/interceptor/pkg/nack"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

// SDP plus ICE candidates stay well below this.
const maxSignalingSize = 256 * 1024

// nackResponderSize is how many sent packets are kept for retransmission, per
// track and per client. pion defaults to 1024 and that is far more history
// than a retransmission can use: a packet the decoder has already passed is
// worth nothing however faithfully it arrives.
//
// The window that matters is the browser's NACK interval plus the round trip.
// pion's generator runs on a 100ms interval, and 1080p at 30 frames and
// 4.3 Mbit/s fills about 450 packets a second at the 1200 byte MTU below. So
// 100ms of history is about 45 packets, and 256 covers well over half a second
// of round trip - more than any link this board is reachable over.
//
// The size is also a memory bound, which is the reason to pick it rather than
// take the default. A full buffer holds up to size * MTU per track, so 256 caps
// video at about 300KB per viewer where 1024 would allow 1.2MB. The value has
// to be one of pion's powers of two.
const nackResponderSize = 256

var (
	upgrader = websocket.Upgrader{
		WriteBufferSize: 256 * 1024,
		CheckOrigin:     middleware.SameOrigin,
	}
	globalManager *WebRTCManager
	managerOnce   sync.Once
)

func getManager() *WebRTCManager {
	managerOnce.Do(func() {
		globalManager = NewWebRTCManager()
	})
	return globalManager
}

func Connect(c *gin.Context) {
	// create WebSocket connection
	wsConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Errorf("failed to create h264 websocket: %s", err)
		return
	}
	stopSessionWatcher := middleware.WatchWebSocket(c.Request.Context(), wsConn)
	defer stopSessionWatcher()
	defer func() {
		_ = wsConn.Close()
		log.Debugf("h264 websocket disconnected: %s", c.ClientIP())
	}()
	log.Debugf("h264 websocket connected: %s", c.ClientIP())

	var zeroTime time.Time
	_ = wsConn.SetReadDeadline(zeroTime)
	// Signaling messages carry SDP and ICE candidates, nothing larger.
	wsConn.SetReadLimit(maxSignalingSize)

	// create video connection
	iceServers := createICEServers()

	mediaEngine, err := createMediaEngine()
	if err != nil {
		log.Errorf("failed to create h264 media engine: %s", err)
		return
	}

	videoConn, err := createPeerConnection(iceServers, mediaEngine)
	if err != nil {
		log.Errorf("failed to create h264 video peer connection: %s", err)
		return
	}
	defer func() {
		_ = videoConn.Close()
		log.Debugf("h264 video peer disconnected: %s", c.ClientIP())
	}()

	// create client
	client := NewClient(wsConn, videoConn)
	if err := client.AddTrack(); err != nil {
		log.Errorf("failed to add track: %s", err)
		return
	}

	// handle signaling
	signalingHandler := NewSignalingHandler(client)
	defer signalingHandler.Close()
	signalingHandler.RegisterCallbacks()
	if err := sendICEServers(client, iceServers); err != nil {
		log.Errorf("failed to send ICE servers: %s", err)
		return
	}

	// read and wait
	for {
		message, err := client.ReadMessage()
		if err != nil {
			return
		}
		if message != nil {
			if err := signalingHandler.HandleMessage(message); err != nil {
				log.Errorf("failed to handle signaling message: %s", err)
			}
		}
	}
}

func createICEServers() []webrtc.ICEServer {
	var iceServers []webrtc.ICEServer

	conf := config.GetInstance()

	if conf.Stun != "" && conf.Stun != "disable" {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs: []string{"stun:" + conf.Stun},
		})
	}

	if conf.Turn.TurnAddr != "" && conf.Turn.TurnUser != "" && conf.Turn.TurnCred != "" {
		iceServers = append(iceServers, webrtc.ICEServer{
			URLs:       []string{"turn:" + conf.Turn.TurnAddr},
			Username:   conf.Turn.TurnUser,
			Credential: conf.Turn.TurnCred,
		})
	}

	return iceServers
}

type clientICEServer struct {
	URLs       []string    `json:"urls"`
	Username   string      `json:"username,omitempty"`
	Credential interface{} `json:"credential,omitempty"`
}

func sendICEServers(client *Client, iceServers []webrtc.ICEServer) error {
	clientServers := make([]clientICEServer, 0, len(iceServers))
	for _, server := range iceServers {
		clientServers = append(clientServers, clientICEServer{
			URLs:       server.URLs,
			Username:   server.Username,
			Credential: server.Credential,
		})
	}

	data, err := json.Marshal(clientServers)
	if err != nil {
		return err
	}

	return client.WriteMessage("ice-servers", string(data))
}

func createMediaEngine() (*webrtc.MediaEngine, error) {
	mediaEngine := &webrtc.MediaEngine{}

	if err := mediaEngine.RegisterDefaultCodecs(); err != nil {
		log.Errorf("failed to register default codecs: %s", err)
		return nil, err
	}

	if err := mediaEngine.RegisterHeaderExtension(
		webrtc.RTPHeaderExtensionCapability{URI: playoutDelayExtensionURI},
		webrtc.RTPCodecTypeVideo,
	); err != nil {
		log.Errorf("failed to register header extension: %s", err)
		return nil, err
	}

	return mediaEngine, nil
}

// createInterceptorRegistry builds the RTCP behaviour this server needs.
//
// pion registers nothing by default: an API built without a registry sends no
// sender reports and answers no retransmission request, which is what this
// package did until now. Both matter here, and for different reasons.
//
// Sender reports carry the mapping from RTP timestamp to wall clock. Without
// them a browser has no common clock for the video and audio tracks, so it
// cannot hold them in sync and cannot measure the round trip either.
//
// NACK is the only repair this server can offer. A lost packet corrupts the
// rest of its frame, and nothing here can force the encoder to emit a keyframe
// early, so an unrepaired loss stays on screen until the next GOP boundary: a
// whole second at the default 30 frames and GOP 30. Retransmission is what
// keeps that from being the normal outcome of a single dropped datagram.
func createInterceptorRegistry() (*interceptor.Registry, error) {
	registry := &interceptor.Registry{}

	if err := webrtc.ConfigureRTCPReports(registry); err != nil {
		return nil, err
	}

	// No RegisterFeedback call is needed. RegisterDefaultCodecs already puts
	// `nack`, `nack pli`, `ccm fir`, `goog-remb` and `transport-cc` on every
	// video codec, so the browser has been able to send retransmission requests
	// all along and there was simply nothing here to answer them.
	//
	// One of those defaults is a promise this server cannot keep. `nack pli`
	// asks the sender for a keyframe now, and `libkvm` exposes `set_h264_gop`
	// and no IDR request, so a picture loss indication is answered with silence
	// and the viewer waits for the next GOP boundary. Removing it from the
	// answer would mean hand-registering the codecs instead of taking the
	// defaults; producing a keyframe on demand would be the better answer, and
	// neither is done here.
	//
	// Only the responder is registered. The generator half of ConfigureNack
	// acts on inbound streams, and this server has no OnTrack handler and never
	// receives media.
	responder, err := nack.NewResponderInterceptor(nack.ResponderSize(nackResponderSize))
	if err != nil {
		return nil, err
	}

	registry.Add(responder)

	return registry, nil
}

func createPeerConnection(iceServers []webrtc.ICEServer, mediaEngine *webrtc.MediaEngine) (*webrtc.PeerConnection, error) {
	settingEngine := webrtc.SettingEngine{}
	settingEngine.SetSRTPProtectionProfiles(
		dtls.SRTP_AEAD_AES_128_GCM,
		dtls.SRTP_AES128_CM_HMAC_SHA1_80,
	)

	apiOptions := []func(api *webrtc.API){
		webrtc.WithSettingEngine(settingEngine),
	}
	if mediaEngine != nil {
		registry, err := createInterceptorRegistry()
		if err != nil {
			log.Errorf("failed to create interceptor registry: %s", err)
			return nil, err
		}

		apiOptions = append(apiOptions,
			webrtc.WithMediaEngine(mediaEngine),
			webrtc.WithInterceptorRegistry(registry),
		)
	}

	api := webrtc.NewAPI(apiOptions...)

	return api.NewPeerConnection(webrtc.Configuration{
		ICEServers:   iceServers,
		SDPSemantics: webrtc.SDPSemanticsUnifiedPlan,
	})
}
