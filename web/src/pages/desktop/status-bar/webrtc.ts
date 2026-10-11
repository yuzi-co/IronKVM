import {
  emptyStats,
  webrtcSample,
  webrtcStats,
  type ConnectionState,
  type StreamStats,
  type SyncSource,
  type WebrtcSample
} from './model.ts';

const SAMPLE_MS = 1000;

function connectionOf(peer: RTCPeerConnection): ConnectionState {
  switch (peer.connectionState) {
    case 'connected':
      return 'connected';
    case 'disconnected':
      return 'reconnecting';
    case 'failed':
    case 'closed':
      return 'failed';
    default:
      return 'connecting';
  }
}

// The video receiver's newest source. Chrome fills in captureTimestamp once
// the server sends the abs-capture-time header extension; until then, and in
// other browsers, it is absent and the bar falls back to its estimate.
function videoSource(peer: RTCPeerConnection): SyncSource | undefined {
  const receiver = peer.getReceivers().find((r) => r.track?.kind === 'video');
  if (!receiver || typeof receiver.getSynchronizationSources !== 'function') return undefined;
  return receiver.getSynchronizationSources()[0] as SyncSource | undefined;
}

// watchWebrtc reads the browser's own statistics for the current session once
// a second and publishes what the status bar shows. It asks the board for
// nothing: getStats() is answered inside the browser. The peer is read
// through getPeer because a reconnect replaces it. Returns the stop function.
export function watchWebrtc(
  getPeer: () => RTCPeerConnection | null,
  publish: (stats: StreamStats) => void
): () => void {
  let previous: WebrtcSample | null = null;
  let previousPeer: RTCPeerConnection | null = null;
  let busy = false;
  let stopped = false;

  const tick = async () => {
    if (busy) return;
    busy = true;
    try {
      const peer = getPeer();
      if (peer !== previousPeer) {
        previous = null;
        previousPeer = peer;
      }
      if (!peer) {
        publish(emptyStats('h264', 'connecting'));
        return;
      }

      const report = await peer.getStats();
      if (stopped) return;
      const sample = webrtcSample(
        report.values() as Iterable<Record<string, unknown>>,
        performance.now()
      );
      publish(webrtcStats(sample, previous, connectionOf(peer), videoSource(peer)));
      previous = sample;
    } catch {
      // A peer closed under the read. The next second reads the new one.
    } finally {
      busy = false;
    }
  };

  void tick();
  const timer = setInterval(() => void tick(), SAMPLE_MS);
  return () => {
    stopped = true;
    clearInterval(timer);
  };
}
