package webrtc

import (
	"NanoKVM-Server/service/stream/audio"

	log "github.com/sirupsen/logrus"
)

const (
	// audioPayloadType is a placeholder. Opus has no static assignment, and
	// pion rewrites it per binding from what the peer negotiated, the same as
	// video.
	audioPayloadType = 0
	audioSSRC        = 0x1234ABCE
)

// hasAudioListener reports whether any connected client negotiated an audio
// track. A viewer that connected before the settings switch was thrown has
// none, and capturing for it would burn a process nobody can hear.
//
// It reads the client snapshot rather than the map, so it does not need the
// manager lock and cannot invert the lock order against Client.mutex.
func (m *WebRTCManager) hasAudioListener() bool {
	for _, client := range m.getClients() {
		if client.hasAudioTrack() {
			return true
		}
	}

	return false
}

// clearAudioStream forgets a subscription whose capture has ended on its own,
// so that a later viewer starts a fresh one. It ignores a subscription that has
// already been replaced.
func (m *WebRTCManager) clearAudioStream(sub *audio.Subscription) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.audioSub != sub {
		return
	}

	m.audioSub = nil
	m.audioSending = false
	m.audioState = audio.StateUnknown
}

// StartAudioStream begins capture if a client can hear it and the gadget has a
// capture card. Availability is checked here, not cached at start, because the
// settings switch rebuilds the gadget while the server runs.
func (m *WebRTCManager) StartAudioStream() {
	if !m.hasAudioListener() || !audio.Available() {
		return
	}

	m.mutex.Lock()
	if m.audioSending {
		// Capture is already running for someone else. The viewer that just
		// connected has not heard its state, and the state may not change
		// again for as long as the host stays idle.
		state := m.audioState
		m.mutex.Unlock()
		m.announceAudioState(state)
		return
	}
	if len(m.clients) == 0 {
		m.mutex.Unlock()
		return
	}

	// The hub starts capture if nobody else holds it, and shares it if a
	// direct viewer already does.
	sub := m.audioHub.Subscribe()
	if sub == nil {
		m.mutex.Unlock()
		return
	}
	m.audioSub = sub
	m.audioSending = true
	m.mutex.Unlock()

	go m.sendAudioStream(sub)

	log.Debugf("start sending opus stream")
}

// StopAudioCapture stops capture whatever the client count, and kills the
// arecord child with it. main.go calls it from dispose() so the child does not
// outlive the server.
//
// An orphaned arecord cannot be counted on to die on its own. It notices the
// closed pipe only when it writes, and a read wedged in the driver (see
// audio.stopTimeout) may never return to write. While the host plays nothing
// it does exit within a second, but that is no help while the host plays and
// the read is stuck. The orphan holds hw:UAC1Gadget,0
// exclusively, so the replacement server's arecord fails with "Device or
// resource busy", spends its whole restart budget in about three seconds, and
// gives up. Audio then stays dead until somebody logs in and kills the orphan.
//
// It is a package function rather than a method because dispose() has no
// manager to hand it. getManager() returns the same singleton that signalling
// uses.
func StopAudioCapture() {
	getManager().stopAudioStream()
	audio.Shared.StopAll()
}

// stopAudioStream ends capture and forgets the stream. The caller decides
// whether stopping is the right thing to do.
func (m *WebRTCManager) stopAudioStream() {
	m.mutex.Lock()

	if !m.audioSending {
		m.mutex.Unlock()
		return
	}

	sub := m.audioSub
	m.audioSub = nil
	m.audioSending = false
	m.audioState = audio.StateUnknown
	m.mutex.Unlock()

	// Close runs outside the lock. As the last listener it stops capture and
	// waits on the capture goroutine, and no other caller may be held up
	// behind that.
	if sub != nil {
		sub.Close()
	}

	log.Debugf("stop sending opus stream")
}

// stopAudioStreamIfIdle stops capture once the last listener has gone.
//
// The condition mirrors StartAudioStream, which starts only when a client can
// hear the result. Stopping on an empty client map instead would keep arecord,
// the encoder and the packetizer running for a viewer that negotiated no audio
// track, which is what a viewer that connected before the switch was thrown
// has for its whole life.
//
// This has to kill the child process rather than wait for the loop to notice.
// While the host plays, arecord blocks in a read for as long as it plays, so
// the loop does not tick and would not see that nobody is listening. While the
// host plays nothing, the loop would retry every few seconds forever.
//
// hasAudioListener reads the atomic client snapshot and takes no lock, so
// calling it before m.mutex cannot invert the lock order.
func (m *WebRTCManager) stopAudioStreamIfIdle() {
	if m.hasAudioListener() {
		return
	}

	m.stopAudioStream()
}

// sendAudioStream packetizes each frame once and hands the packets to every
// client, the same way the video loop does.
//
// A gap in the sequence is frames this loop lost in the hub. The packetizer
// advances its RTP timestamp only by what it is handed, so a lost frame would
// cut 20 ms out of the stream and the receiver would drift further from the
// host with every one. SkipSamples moves the clock across the gap, and the
// receiver treats it as loss, which its jitter buffer is built for.
func (m *WebRTCManager) sendAudioStream(sub *audio.Subscription) {
	var next uint64
	started := false

	for frame := range sub.Frames() {
		// A notice carries no audio and no sequence, so it must not move the
		// clock.
		if frame.State != audio.StateUnknown {
			m.setAudioState(sub, frame.State)
			continue
		}

		if started && frame.Seq > next {
			m.audioPacketizer.SkipSamples(uint32(frame.Seq-next) * audio.SamplesPerFrame)
		}
		next = frame.Seq + 1
		started = true

		m.deliverAudioFrame(frame.Data)
	}

	// The channel closed for one of two reasons: the last listener left and
	// stopAudioStreamIfIdle stopped this stream, or Start could not construct
	// the encoder and closed the channel itself before any capture ran. The
	// second case leaves the flag set unless it is cleared here, and no later
	// StartAudioStream could ever run.
	//
	// Clearing the flag does not bring audio back for the viewer that lost it.
	// StartAudioStream has one caller, an ICE state change, so a viewer whose
	// encoder failed to construct stays silent until it reconnects. What this
	// buys is that the next connection starts a fresh stream instead of finding
	// the manager still convinced audio is being sent.
	m.clearAudioStream(sub)
}

// deliverAudioFrame packetizes one frame and hands the packets to every
// client that negotiated audio. Split out from sendAudioStream so the
// packetize-and-fan-out step can be tested directly, without a live capture
// stream behind it.
//
// A client whose track has no audio leg is skipped rather than enqueued and
// dropped: enqueueAudio would still take the slot lock and wake that client's
// writer for a frame it can only discard, fifty times a second, for as long
// as anyone else is listening.
//
// The listener check comes before Packetize for the same reason. Packetize
// allocates a header and a payload slice per packet, and a frame nobody can
// hear is fifty of those a second thrown away on a board with one core.
func (m *WebRTCManager) deliverAudioFrame(frame []byte) {
	clients := m.getClients()

	var listening bool
	for _, client := range clients {
		if client.hasAudioTrack() {
			listening = true
			break
		}
	}

	if !listening {
		return
	}

	// The sample count is per channel and fixed at 20 ms, whatever the encoded
	// packet happens to be long.
	packets := m.audioPacketizer.Packetize(frame, audio.SamplesPerFrame)

	for _, client := range clients {
		if !client.hasAudioTrack() {
			continue
		}

		client.enqueueAudio(packets)
	}
}

// audioStateEvent is the signalling event that tells a viewer what capture is
// doing, so it can say why there is no sound. Its data is State.String().
const audioStateEvent = "audio-state"

// setAudioState records the capture's new state and tells every viewer that
// negotiated audio. A notice from a subscription that has been replaced or
// stopped is ignored.
func (m *WebRTCManager) setAudioState(sub *audio.Subscription, state audio.State) {
	m.mutex.Lock()
	if m.audioSub != sub {
		m.mutex.Unlock()
		return
	}
	m.audioState = state
	m.mutex.Unlock()

	m.announceAudioState(state)
}

// announceAudioState sends a state to every viewer that negotiated audio. It
// runs outside m.mutex, because WriteMessage takes each client's own lock.
func (m *WebRTCManager) announceAudioState(state audio.State) {
	if state == audio.StateUnknown {
		return
	}

	for _, client := range m.getClients() {
		if client.ws == nil || !client.hasAudioTrack() {
			continue
		}

		_ = client.WriteMessage(audioStateEvent, state.String())
	}
}
