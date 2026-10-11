import type { WallClockOffset } from '@/lib/clock-offset.ts';

import {
  emptyStats,
  refreshIntervalMs,
  webrtcSample,
  webrtcStats,
  type ConnectionState,
  type StreamStats,
  type SyncSource,
  type WebrtcSample
} from './model.ts';

const SAMPLE_MS = 1000;

// The animation frame gaps the display's refresh is read from: two seconds'
// worth at 60 Hz.
const REFRESH_GAPS = 120;

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

// watchRefresh follows the display's refresh interval from the gaps between
// animation frames, which the estimate's display part is half of. Returns the
// reader and the stop function.
function watchRefresh(): [() => number | null, () => void] {
  const gaps: number[] = [];
  let last: number | null = null;
  let frame = 0;
  const step = (now: number) => {
    if (last !== null) {
      gaps.push(now - last);
      if (gaps.length > REFRESH_GAPS) gaps.shift();
    }
    last = now;
    frame = requestAnimationFrame(step);
  };
  frame = requestAnimationFrame(step);
  return [() => refreshIntervalMs(gaps), () => cancelAnimationFrame(frame)];
}

// watchWebrtc reads the browser's own statistics for the current session once
// a second and publishes what the status bar shows. getStats() is answered
// inside the browser; the board's clock comes through getClock, from the
// signalling socket, where the server answers for it. The peer is read
// through getPeer because a reconnect replaces it. Returns the stop function.
export function watchWebrtc(
  getPeer: () => RTCPeerConnection | null,
  publish: (stats: StreamStats) => void,
  getClock: () => WallClockOffset | null = () => null
): () => void {
  let previous: WebrtcSample | null = null;
  let previousPeer: RTCPeerConnection | null = null;
  let busy = false;
  let stopped = false;
  const [refreshMs, stopRefresh] = watchRefresh();

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
      publish(
        webrtcStats(sample, previous, connectionOf(peer), {
          source: videoSource(peer),
          clock: getClock(),
          refreshMs: refreshMs()
        })
      );
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
    stopRefresh();
  };
}
