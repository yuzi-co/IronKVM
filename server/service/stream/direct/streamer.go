package direct

import (
	"NanoKVM-Server/service/stream"
	"NanoKVM-Server/service/vm"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
)

type Streamer struct {
	mutex          sync.Mutex
	clients        map[*client]struct{}
	clientSnapshot atomic.Pointer[[]*client]
	running        bool
	viewerVersion  uint64
}

func newStreamer() *Streamer {
	s := &Streamer{
		clients: make(map[*client]struct{}),
	}
	s.updateClientSnapshotLocked()

	return s
}

func (s *Streamer) addClient(client *client) {
	client.start()

	s.mutex.Lock()
	s.clients[client] = struct{}{}
	count := s.updateClientSnapshotLocked()
	s.viewerVersion++
	version := s.viewerVersion
	start := !s.running
	if start {
		s.running = true
	}
	s.mutex.Unlock()
	vm.UpdateHdmiViewerSnapshot("direct", count, version)

	if start {
		go s.run()
		log.Debug("h264 stream started")
	}
}

func (s *Streamer) removeClient(client *client) {
	s.mutex.Lock()
	if _, exists := s.clients[client]; !exists {
		s.mutex.Unlock()
		return
	}

	delete(s.clients, client)
	count := s.updateClientSnapshotLocked()
	s.viewerVersion++
	version := s.viewerVersion
	s.mutex.Unlock()
	client.close()
	vm.UpdateHdmiViewerSnapshot("direct", count, version)

	log.Debugf("h264 websocket disconnected, remaining clients: %d", count)
}

func (s *Streamer) updateClientSnapshotLocked() int {
	clients := make([]*client, 0, len(s.clients))
	for client := range s.clients {
		clients = append(clients, client)
	}
	s.clientSnapshot.Store(&clients)

	return len(clients)
}

func (s *Streamer) getClients() []*client {
	clients := s.clientSnapshot.Load()
	if clients == nil {
		return nil
	}

	return *clients
}

// idleCheckInterval is how often the loop asks whether its last viewer has
// gone. The loop used to notice on the capture tick, which it no longer owns.
// Nothing is served in the meantime, so a second of lag costs nothing, and
// stopIfIdle rechecks the client count under the mutex, so a viewer arriving
// inside that second is not stranded.
const idleCheckInterval = time.Second

// run takes frames from the shared capture loop rather than reading the
// encoder itself. WebRTC mode reads the same encoder, and two readers do not
// each get the stream: they divide it, and half a GOP decodes to nothing.
func (s *Streamer) run() {
	// The demand check runs on the capture loop's goroutine. It reads the
	// client snapshot, which is an atomic pointer, so it takes no lock that
	// this goroutine holds.
	subscription := stream.SubscribeH264(func() bool {
		clients := s.getClients()

		return len(clients) > 0 && hasCaptureDemand(clients)
	})
	defer subscription.Close()

	idle := time.NewTicker(idleCheckInterval)
	defer idle.Stop()

	for {
		select {
		case <-idle.C:
			if len(s.getClients()) == 0 && s.stopIfIdle() {
				log.Debug("h264 stream stopped due to no clients")
				return
			}

		case frame, ok := <-subscription.Frames():
			if !ok {
				return
			}

			stream.UpdateCaptureStatus(stream.CaptureModeDirect, frame.Result)
			if frame.Result < 0 || len(frame.Data) == 0 {
				continue
			}

			clients := s.getClients()
			if len(clients) == 0 {
				continue
			}

			outbound := newOutboundFrame(frame.KeyFrame, frame.Timestamp, frame.Data)
			for _, client := range clients {
				client.offer(outbound)
			}
		}
	}
}

func (s *Streamer) stopIfIdle() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.clients) > 0 {
		return false
	}

	s.running = false
	return true
}
