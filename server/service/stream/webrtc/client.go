package webrtc

import (
	"NanoKVM-Server/common"
	"encoding/json"
	"sync"

	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/stream/audio"
	"NanoKVM-Server/service/stream/framequeue"

	"github.com/gorilla/websocket"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v4"
	log "github.com/sirupsen/logrus"
)

func NewClient(ws *websocket.Conn, videoConn *webrtc.PeerConnection) *Client {
	return &Client{
		ws:        ws,
		video:     videoConn,
		mutex:     sync.Mutex{},
		queue:     framequeue.New[[]*rtp.Packet](framequeue.DefaultMaxFrames, framequeue.DefaultMaxDelay),
		done:      make(chan struct{}),
		audioSlot: stream.NewFrameSlot[[]*rtp.Packet](),
		audioDone: make(chan struct{}),
	}
}

// enqueue offers a frame to this client and never blocks.
//
// The writer can be well behind for a moment: a 1080p keyframe is about a
// hundred packets and takes it 90 to 150 ms. The frames that arrive meanwhile
// wait in the queue rather than being dropped, which would cost the viewer the
// rest of the GOP. Only a client that stays behind loses frames, and then up to
// the next keyframe, since an H.264 or H.265 stream with a hole in it stays
// broken until one arrives. See framequeue.Queue.
func (c *Client) enqueue(packets []*rtp.Packet, isKeyFrame bool) {
	c.queue.Put(packets, isKeyFrame)
}

// enqueueAudio offers a frame to this client and never blocks.
//
// Unlike video, audio has no keyframe to recover from a gap, so the newest
// frame replaces whatever is pending rather than being dropped.
func (c *Client) enqueueAudio(packets []*rtp.Packet) {
	c.audioSlot.Replace(packets)
}

// hasAudioTrack reports whether this client negotiated audio. A client that
// connected while the gadget had no capture card did not.
func (c *Client) hasAudioTrack() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	return c.track != nil && c.track.audio != nil
}

// writeAudio drains the audio slot until it is closed or the connection fails.
func (c *Client) writeAudio() {
	defer close(c.audioDone)

	for {
		packets, ok := c.audioSlot.Take()
		if !ok {
			return
		}

		c.mutex.Lock()
		track := c.track
		c.mutex.Unlock()

		if track == nil || track.audio == nil {
			continue
		}

		if err := track.writeAudioPackets(packets); err != nil {
			log.Debugf("audio write to %s failed: %s", c.ws.RemoteAddr(), err)

			c.Close()

			return
		}
	}
}

// write drains the queue until it is closed or the connection fails. It is the
// only goroutine that writes to this client's track, so the capture loop never
// waits on a viewer.
func (c *Client) write() {
	defer close(c.done)

	for {
		packets, ok := c.queue.Take()
		if !ok {
			return
		}

		c.mutex.Lock()
		track := c.track
		c.mutex.Unlock()

		if track == nil {
			continue
		}

		if err := track.writePackets(packets); err != nil {
			log.Debugf("h264 write to %s failed: %s", c.ws.RemoteAddr(), err)

			// Unblock the reader so the handler tears this client down.
			c.Close()

			return
		}
	}
}

// startWriters starts write and writeAudio at most once for this client's
// lifetime, no matter how many times AddClient runs for it.
func (c *Client) startWriters() {
	c.writersOnce.Do(func() {
		go c.write()
		go c.writeAudio()
	})
}

// stop releases the writer and waits for it to let go of the connection.
func (c *Client) stop() {
	c.queue.Close()
	<-c.done

	c.audioSlot.Close()
	<-c.audioDone

	if dropped := c.queue.Dropped(); dropped > 0 {
		log.Debugf("video client dropped %d frames in %d resyncs", dropped, c.queue.Resyncs())
	}

	if dropped := c.audioSlot.Dropped(); dropped > 0 {
		log.Debugf("audio client dropped %d frames", dropped)
	}
}

func (c *Client) Close() {
	if c.video != nil {
		if err := c.video.Close(); err != nil {
			log.Debugf("failed to close video peer connection: %s", err)
		}
	}

	if c.ws != nil {
		if err := c.ws.Close(); err != nil {
			log.Debugf("failed to close websocket: %s", err)
		}
	}
}

func (c *Client) WriteMessage(event string, data string) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	message := &Message{
		Event: event,
		Data:  data,
	}

	if err := c.ws.WriteJSON(message); err != nil {
		log.Errorf("failed to send message %s: %v", event, err)
		return err
	}

	log.Debugf("sent message %s", event)
	return nil
}

func (c *Client) ReadMessage() (*Message, error) {
	_, raw, err := c.ws.ReadMessage()
	if err != nil {
		log.Errorf("failed to read message: %v", err)
		return nil, err
	}

	var message Message
	if err := json.Unmarshal(raw, &message); err != nil {
		log.Errorf("failed to unmarshal message: %v", err)
		return nil, nil
	}

	return &message, nil
}

func (c *Client) AddTrack() error {
	// The codec is read once, here, and remembered. The track declares it to
	// the peer and the answer is built from that, so this session carries this
	// codec for as long as it lives even if the setting changes underneath.
	c.codec = common.GetScreen().Snapshot().Codec

	// video track
	videoTrack, err := webrtc.NewTrackLocalStaticRTP(
		webrtc.RTPCodecCapability{MimeType: mimeTypeForCodec(c.codec)},
		"video",
		"pion-video",
	)
	if err != nil {
		log.Errorf("failed to create video track: %s", err)
		return err
	}

	videoSender, err := c.video.AddTrack(videoTrack)
	if err != nil {
		log.Errorf("failed to add video track: %s", err)
		return err
	}
	go startRTCPReader(videoSender)

	track := &Track{video: videoTrack}

	// The card comes and goes with the settings switch, so this is decided per
	// connection rather than once at start.
	if audio.Available() {
		audioTrack, err := webrtc.NewTrackLocalStaticRTP(
			webrtc.RTPCodecCapability{
				MimeType:  webrtc.MimeTypeOpus,
				ClockRate: audioClockRate,
				Channels:  audio.Channels,
			},
			"audio",
			"pion-audio",
		)
		if err != nil {
			log.Errorf("failed to create audio track: %s", err)
			return err
		}

		audioSender, err := c.video.AddTrack(audioTrack)
		if err != nil {
			log.Errorf("failed to add audio track: %s", err)
			return err
		}
		go startRTCPReader(audioSender)

		track.audio = audioTrack
	}

	c.mutex.Lock()
	c.track = track
	c.mutex.Unlock()

	return nil
}

func startRTCPReader(sender *webrtc.RTPSender) {
	rtcpBuf := make([]byte, 1500)
	for {
		if _, _, err := sender.Read(rtcpBuf); err != nil {
			log.Debugf("RTCP reader error: %v", err)
			return
		}
	}
}
