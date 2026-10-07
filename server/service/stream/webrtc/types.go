package webrtc

import (
	"sync"
	"sync/atomic"

	"NanoKVM-Server/authn"
	"NanoKVM-Server/service/roommic"
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/stream/framequeue"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
)

type WebRTCManager struct {
	clients        map[*websocket.Conn]*Client
	clientSnapshot atomic.Pointer[[]*Client]
	videoSending   bool
	mutex          sync.Mutex
	viewerVersion  uint64

	// videoPacketizers holds one packetizer per codec. A frame is cut into RTP
	// packets once per codec and the packets handed to every client on that
	// codec, rather than each client paying to packetize and copy the same
	// frame again.
	//
	// It is a map rather than a field because the codec is a global setting
	// that can change while sessions are live, and each packetizer has to keep
	// its own RTP sequence across such a change.
	videoPacketizers map[uint8]rtp.Packetizer

	audioSending    bool
	audioSub        *audio.Subscription
	audioPacketizer rtp.Packetizer
	// audioState is the capture's last reported state, which a viewer that
	// joins a running capture is told at once. Guarded by mutex.
	audioState audio.State

	// audioHub is audio.Shared in production. The capture device opens
	// exclusively, so this path and H.264 direct share one capture through it.
	// A field so a test can supply a hub with no arecord behind it.
	audioHub *audio.Hub
}

type Client struct {
	ws    *websocket.Conn
	video *webrtc.PeerConnection
	track *Track
	mutex sync.Mutex

	// codec is what this session's video track told the peer it carries. It is
	// fixed at AddTrack and never changes: the mime type is part of the
	// negotiated description, so a codec change needs a new session rather than
	// a different payload in the old one.
	codec uint8

	// codecChanged sends the viewer one "codec-changed" message once the
	// encoder runs another codec than this session negotiated. The sender skips
	// such a session, and without the message the browser shows a frozen
	// picture until it reloads (#83).
	codecChanged sync.Once

	// queue holds the frames waiting for this client. The capture loop hands a
	// frame over and moves on; the writer goroutine takes frames at whatever
	// rate this connection manages. It is several frames deep so that the
	// frames after a slow keyframe wait instead of being dropped.
	queue *framequeue.Queue[[]*rtp.Packet]
	done  chan struct{}

	// ready closes when the peer connection is up and packets written to the
	// track reach the viewer. The video writer waits for it, or for stopping.
	ready     chan struct{}
	readyOnce sync.Once
	stopping  chan struct{}
	stopOnce  sync.Once

	// audioSlot holds at most one pending audio frame, with its own writer
	// goroutine. Sharing the video queue would drop audio whenever video fell
	// behind, and the two have nothing to do with each other.
	audioSlot *stream.FrameSlot[[]*rtp.Packet]
	audioDone chan struct{}

	// writersOnce starts write and writeAudio at most once for this Client's
	// whole lifetime. It has to live here rather than on the manager's client
	// map: ICE can flap Connected -> Disconnected -> Connected, which drives
	// AddClient -> RemoveClient -> AddClient on the same *Client pointer, and
	// RemoveClient already closed both slots by the time the second AddClient
	// runs. Gating on map membership would restart both writers, and each
	// one's first Take() would return immediately and close an already-closed
	// channel.
	writersOnce sync.Once

	// user names the account behind this session, for the room microphone's
	// log.
	user string

	// role is that account's role. Only an administrator is told who is
	// listening to the room microphone.
	role authn.Role

	// room is this viewer's hold on the room microphone, nil while it is off.
	// Guarded by roomMutex, which also orders the state messages.
	roomMutex sync.Mutex
	room      *roommic.Listener
	// roomError is the reason the last attempt to switch it on failed, sent
	// once with the next state message.
	roomError string
}

func (c *Client) WsConn() *websocket.Conn {
	return c.ws
}

type SignalingHandler struct {
	client         *Client
	mutex          sync.Mutex
	unregisterMode func()
	closed         bool
}

type Track struct {
	video rtpWriter

	// audio is nil when the gadget had no capture card at negotiation time.
	audio rtpWriter

	// room carries the board's own microphone. It is nil on a kernel without
	// the onboard card. It is negotiated whatever the administrator's setting,
	// so allowing the microphone needs no new session; nothing is sent on it
	// until this viewer switches it on.
	room rtpWriter

	// extensionID is negotiated on the websocket goroutine and read on the
	// capture goroutine.
	extensionID atomic.Uint32

	// sent is the newest video sequence number written, with sentValid set
	// once there is one. The writer stores it and the RTCP reader reads it.
	sent atomic.Uint32
	// sentRing holds the latest sequence numbers written, by their low bits
	// (wasSent).
	sentRing [sentRingSize]atomic.Uint32
	// keySent is the first sequence number of the last keyframe written,
	// with sentValid set once there is one (markKey).
	keySent atomic.Uint32
}

type Message struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}
