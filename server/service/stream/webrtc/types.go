package webrtc

import (
	"sync"
	"sync/atomic"

	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/audio"

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
	audioStream     *audio.Stream
	audioPacketizer rtp.Packetizer
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

	// slot holds at most one frame for this client. The capture loop hands a
	// frame over and moves on; the writer goroutine takes frames at whatever
	// rate this connection manages.
	slot *stream.FrameSlot[[]*rtp.Packet]
	done chan struct{}

	// audioSlot holds at most one pending audio frame, with its own writer
	// goroutine. Sharing the video slot would drop audio whenever video fell
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

	// waitingForKeyFrame is read and written only by the capture goroutine.
	waitingForKeyFrame bool
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

	// extensionID is negotiated on the websocket goroutine and read on the
	// capture goroutine.
	extensionID atomic.Uint32
}

type Message struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}
