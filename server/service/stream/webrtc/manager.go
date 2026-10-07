package webrtc

import (
	"NanoKVM-Server/common"
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/stream/procs"
	"NanoKVM-Server/service/vm"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/rtp/codecs"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

const (
	// rtpMTU keeps a packet inside a normal path MTU once the RTP, UDP and IP
	// headers are added.
	rtpMTU = 1200

	// videoPayloadType and videoSSRC are placeholders: pion rewrites both per
	// binding from what the peer negotiated.
	videoPayloadType = 100
	videoSSRC        = 0x1234ABCD

	// clockRate is the RTP clock for H.264.
	clockRate = 90000
)

func NewWebRTCManager() *WebRTCManager {
	m := &WebRTCManager{
		clients:      make(map[*websocket.Conn]*Client),
		videoSending: false,
		videoPacketizers: map[uint8]rtp.Packetizer{
			common.CodecH264: rtp.NewPacketizer(
				rtpMTU,
				videoPayloadType,
				videoSSRC,
				&codecs.H264Payloader{},
				rtp.NewRandomSequencer(),
				clockRate,
			),
			common.CodecH265: rtp.NewPacketizer(
				rtpMTU,
				videoPayloadType,
				videoSSRC,
				&codecs.H265Payloader{},
				rtp.NewRandomSequencer(),
				clockRate,
			),
		},
		audioPacketizer: rtp.NewPacketizer(
			rtpMTU,
			audioPayloadType,
			audioSSRC,
			&codecs.OpusPayloader{},
			rtp.NewRandomSequencer(),
			audioClockRate,
		),
		audioHub: audio.Shared,
	}
	m.updateClientSnapshotLocked()

	return m
}

// storeClient records the client and reports whether it is new to the
// manager's map. The bool is bookkeeping only now: it does not gate the
// writer start, because a reconnect reuses the same *Client after
// RemoveClient deleted its map entry, and the map alone cannot tell that
// apart from a client the manager has never seen. See Client.startWriters.
func (m *WebRTCManager) storeClient(ws *websocket.Conn, client *Client) (int, uint64, bool) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	_, exists := m.clients[ws]
	m.clients[ws] = client

	count := m.updateClientSnapshotLocked()
	m.viewerVersion++

	return count, m.viewerVersion, !exists
}

func (m *WebRTCManager) AddClient(ws *websocket.Conn, client *Client) {
	count, version, _ := m.storeClient(ws, client)

	// The writer-start guard lives on the Client, not here: ICE reaches
	// Connected and then Completed for one handshake, and can also flap
	// Connected -> Disconnected -> Connected on a blip, and signalling calls
	// AddClient on all of them, sometimes with the same *Client after
	// RemoveClient already closed its slots. Starting the writers twice would
	// put two goroutines on one slot, and the second close of the done
	// channel panics the server.
	client.startWriters()

	vm.UpdateHdmiViewerSnapshot("webrtc", count, version)

	log.Debugf("added client %s, total clients: %d", ws.RemoteAddr(), count)
}

func (m *WebRTCManager) RemoveClient(ws *websocket.Conn) {
	m.mutex.Lock()
	client, exists := m.clients[ws]
	delete(m.clients, ws)
	count := m.updateClientSnapshotLocked()
	m.viewerVersion++
	version := m.viewerVersion
	m.mutex.Unlock()
	vm.UpdateHdmiViewerSnapshot("webrtc", count, version)

	if exists {
		client.stop()
	}

	m.stopAudioStreamIfIdle()

	log.Debugf("removed client %s, total clients: %d", ws.RemoteAddr(), count)
}

func (m *WebRTCManager) GetClientCount() int {
	return len(m.getClients())
}

func (m *WebRTCManager) updateClientSnapshotLocked() int {
	clients := make([]*Client, 0, len(m.clients))
	for _, client := range m.clients {
		clients = append(clients, client)
	}
	m.clientSnapshot.Store(&clients)

	return len(clients)
}

func (m *WebRTCManager) getClients() []*Client {
	clients := m.clientSnapshot.Load()
	if clients == nil {
		return nil
	}

	return *clients
}

func (m *WebRTCManager) StartVideoStream() {
	m.mutex.Lock()
	if m.videoSending || len(m.clients) == 0 {
		m.mutex.Unlock()
		return
	}
	m.videoSending = true
	m.mutex.Unlock()

	go m.sendVideoStream()
	log.Debugf("start sending h264 stream")
}

func (m *WebRTCManager) stopVideoStreamIfIdle() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.clients) > 0 {
		return false
	}

	m.videoSending = false
	return true
}

// moreProcs is procs.Two, a variable so a test can count the calls without
// touching the runtime.
var moreProcs = procs.Two

// soloProcs holds a second P while the video stream has exactly one viewer.
//
// The capture loop's read is a cgo call that waits 13 to 17 ms for a 1080p
// picture, and with one P the viewer's writer waits through most of it with
// the core idle, then the capture goroutine waits behind the writer. One
// viewer of H.264 at 1080p60 on the mainline slot got 57.5 to 57.9 fps, its
// frames 115 to 125 ms behind the least delayed one at the median; with a
// second P 58.8 to 60.9 fps and 17 to 36 ms. At 30 fps the median went from
// 75 to 6 to 12 ms, the 95th percentile from 245 to 71 to 124 ms. Slot A, on
// the vendor library, gains the same way (ironkvm-dist trial 56).
//
// Not with two viewers or more. The core is then close to busy with the
// writers alone, 0.6 to 0.8 ms of CPU a packet for each viewer (SRTP in
// software, the UDP send, pion), and a second P only adds the runtime's
// spinning and lets the capture loop hand the writers frames they cannot send
// in time: with three viewers the 95th percentile went from 260 to 300 ms to
// 740 to 900 ms for at most three more frames a second. With one P the
// capture loop yields to the writers instead, which holds the delay down.
//
// It changes only when the number of viewers does. update runs on the stream
// goroutine only.
type soloProcs struct {
	release func()
}

func (s *soloProcs) update(viewers int) {
	switch {
	case viewers == 1 && s.release == nil:
		s.release = moreProcs()
	case viewers != 1 && s.release != nil:
		s.release()
		s.release = nil
	}
}

// idleCheckInterval is how often the loop asks whether its last viewer has
// gone. The loop used to notice on the capture tick, which it no longer owns.
const idleCheckInterval = time.Second

// mimeTypeForCodec names what a track carrying this codec must declare.
//
// The value is fixed when the track is created, which is before the peer
// answers, so a session is tied to the codec that was configured when it
// negotiated. The sender below uses the codec each session recorded rather
// than the one currently configured, so a setting changed under a live viewer
// cannot put HEVC into a track that promised H.264.
func mimeTypeForCodec(codec uint8) string {
	if codec == common.CodecH265 {
		return webrtc.MimeTypeH265
	}

	return webrtc.MimeTypeH264
}

// packetizerFor returns the one packetizer for this codec.
//
// It has to be the same one every time: the packetizer carries the RTP
// sequence number, and a fresh one per frame would restart the sequence and
// leave the peer treating the stream as permanently reordered. An unknown
// codec gets the H.264 packetizer, matching mimeTypeForCodec.
func (m *WebRTCManager) packetizerFor(codec uint8) rtp.Packetizer {
	if packetizer, ok := m.videoPacketizers[codec]; ok {
		return packetizer
	}

	return m.videoPacketizers[common.CodecH264]
}

// sendVideoStream takes frames from the shared capture loop rather than
// reading the encoder itself. Direct mode reads the same encoder, and two
// readers do not each get the stream: they divide it between them.
func (m *WebRTCManager) sendVideoStream() {
	subscription := stream.SubscribeH264(func() bool {
		return len(m.getClients()) > 0
	})
	defer subscription.Close()

	idle := time.NewTicker(idleCheckInterval)
	defer idle.Stop()

	var solo soloProcs
	defer solo.update(0)

	// One RTP clock per packetizer, for this run of the stream: a frame left
	// out moves it on, see frameClock.
	clocks := make(map[rtp.Packetizer]*frameClock)

	for {
		select {
		case <-idle.C:
			solo.update(len(m.getClients()))
			if len(m.getClients()) == 0 && m.stopVideoStreamIfIdle() {
				log.Debugf("stop sending h264 stream")
				return
			}

		case frame, ok := <-subscription.Frames():
			if !ok {
				return
			}

			stream.UpdateCaptureStatus(stream.CaptureModeH264, frame.Result)
			if frame.Result < 0 || len(frame.Data) == 0 {
				continue
			}

			clients := m.getClients()
			solo.update(len(clients))
			if len(clients) == 0 {
				continue
			}

			// Packetized once per codec, not once per viewer. Cutting the
			// same frame up again for each viewer copies the whole payload
			// per client, which is real work on a board with one core and no
			// memory to spare.
			//
			// There is normally one codec in play, so this is one packetize
			// as before. Two only happens across a codec change, while
			// sessions negotiated before it are still open.
			samples := uint32(frame.Duration.Seconds() * clockRate)
			captured := common.GetScreen().Snapshot().Codec

			var packets []*rtp.Packet
			for _, client := range clients {
				// A session that negotiated the other codec cannot be served
				// this frame: its description promised something else, and
				// the peer would decode noise. The viewer is told once and
				// reconnects, which renegotiates at the current codec.
				if client.codec != captured {
					client.noteCodecChanged(captured)
					continue
				}
				if packets == nil {
					packetizer := m.packetizerFor(captured)
					clock, ok := clocks[packetizer]
					if !ok {
						clock = &frameClock{}
						clocks[packetizer] = clock
					}
					if skip := clock.skip(frame.Timestamp, frame.Duration, clockRate); skip > 0 {
						packetizer.SkipSamples(skip)
					}
					packets = packetizer.Packetize(frame.Data, samples)
				}

				// Handing the frame over never blocks: a client that is
				// behind drops it and waits for the next keyframe.
				client.enqueue(packets, frame.KeyFrame)
			}
		}
	}
}
