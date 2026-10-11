// What the stream status bar shows, and the arithmetic behind it, kept free of
// the DOM and of '@/' imports so node's test runner can load it.
//
// Every number comes from what the browser measured itself, apart from the
// source mode and the board's capture rate (use-source-info.ts). Latency is
// only ever one of three things, and the bar says which:
//
//   capture  capture to display, from the abs-capture-time RTP header
//            extension (WebRTC, when the server sends it), corrected by the
//            offset between the board's clock and this page's where the
//            server answers the clock request
//   encoder  the frame leaving the encoder to it being drawn, against the
//            board's clock (H.264 direct, clock-synchronised over the socket)
//   estimate a sum of measured parts: the board's own share (capture to the
//            network, which the server measures), half the round trip, the
//            time from a frame's first packet to its decode, and half a
//            display refresh. From a server that does not measure its share,
//            only what the browser sees, which is a lower bound.

import {
  midpointMs,
  quickestReading,
  roundTripMs,
  type WallClockOffset
} from '../../../lib/clock-offset.ts';

export type StreamMode = 'h264' | 'direct' | 'mjpeg';

export type ConnectionState = 'connecting' | 'connected' | 'reconnecting' | 'failed' | 'paused';

export type LatencyKind = 'capture' | 'encoder' | 'estimate';

export type Latency = {
  ms: number;
  kind: LatencyKind;
  // The clock error the number may carry either way, when known.
  errorMs?: number;
  // A capture time read against the board's clock with no offset to this
  // browser's: right only when both clocks keep the same time.
  clockAssumed?: boolean;
  // How much of it had passed when the frame arrived in this browser, where
  // that is known (the direct path). The rest is decoding and drawing here.
  arrivalMs?: number;
  // The measured offset the capture time was corrected by: the board's clock
  // minus this page's.
  clockOffsetMs?: number;
  // What an estimate is made of. Without boardMs it is only what the browser
  // sees, and ms leaves out displayMs too, as it always did.
  parts?: LatencyParts;
};

export type LatencyParts = {
  // Half the round trip.
  networkMs: number;
  // From a frame's first packet to its decode: buffering and decoding.
  receiveMs: number;
  // Half a display refresh, the average wait for the next one.
  displayMs?: number;
  // The board's median from capture to handing the frame to the network.
  boardMs?: number;
};

export type StreamStats = {
  mode: StreamMode;
  connection: ConnectionState;
  codec: string | null;
  // The picture as it arrives, 0 while unknown.
  width: number;
  height: number;
  fpsReceived: number | null;
  fpsDecoded: number | null;
  // Frames the browser received but did not show, per second.
  fpsDropped: number | null;
  kbps: number | null;
  latency: Latency | null;
  rttMs: number | null;
  // Packets lost in the last interval, as a share of those expected.
  lossPct: number | null;
  // How the media travels, e.g. "UDP host" or "TCP relay".
  transport: string | null;
};

export function emptyStats(mode: StreamMode, connection: ConnectionState): StreamStats {
  return {
    mode,
    connection,
    codec: null,
    width: 0,
    height: 0,
    fpsReceived: null,
    fpsDecoded: null,
    fpsDropped: null,
    kbps: null,
    latency: null,
    rttMs: null,
    lossPct: null,
    transport: null
  };
}

// ---- rates -----------------------------------------------------------------

// rate turns two readings of a counter into a per-second rate. A counter that
// went backwards was reset (a new session), which says nothing about the rate.
export function rate(
  current: number | undefined,
  previous: number | undefined,
  seconds: number
): number | null {
  if (current === undefined || previous === undefined || !(seconds > 0)) return null;
  const delta = current - previous;
  if (delta < 0) return null;
  return delta / seconds;
}

// ---- WebRTC ----------------------------------------------------------------

// The parts of one getStats() report the bar reads: the video inbound-rtp
// entry, its codec, and the selected candidate pair with its local candidate.
export type WebrtcSample = {
  // Milliseconds, the report's own clock.
  at: number;
  bytesReceived?: number;
  packetsReceived?: number;
  packetsLost?: number;
  framesReceived?: number;
  framesDecoded?: number;
  framesDropped?: number;
  frameWidth?: number;
  frameHeight?: number;
  jitterBufferDelay?: number;
  jitterBufferEmittedCount?: number;
  totalDecodeTime?: number;
  totalProcessingDelay?: number;
  mimeType?: string;
  rttS?: number;
  protocol?: string;
  candidateType?: string;
};

type StatsEntry = Record<string, unknown> & { type?: string; id?: string };

const num = (value: unknown) =>
  typeof value === 'number' && Number.isFinite(value) ? value : undefined;
const str = (value: unknown) => (typeof value === 'string' ? value : undefined);

// webrtcSample reads a getStats() report, given as its entries.
export function webrtcSample(entries: Iterable<StatsEntry>, at: number): WebrtcSample {
  const byId = new Map<string, StatsEntry>();
  let inbound: StatsEntry | undefined;
  let pair: StatsEntry | undefined;
  let transport: StatsEntry | undefined;

  for (const entry of entries) {
    if (entry.id) byId.set(entry.id, entry);
    if (entry.type === 'inbound-rtp' && entry.kind === 'video') inbound = entry;
    if (entry.type === 'transport') transport = entry;
    if (entry.type === 'candidate-pair' && (entry.selected === true || entry.nominated === true)) {
      if (!pair || entry.selected === true) pair = entry;
    }
  }

  // The transport names the pair in use; Chrome marks it nowhere else.
  const selectedId = str(transport?.selectedCandidatePairId);
  if (selectedId && byId.has(selectedId)) pair = byId.get(selectedId);

  const sample: WebrtcSample = { at };
  if (inbound) {
    sample.bytesReceived = num(inbound.bytesReceived);
    sample.packetsReceived = num(inbound.packetsReceived);
    sample.packetsLost = num(inbound.packetsLost);
    sample.framesReceived = num(inbound.framesReceived);
    sample.framesDecoded = num(inbound.framesDecoded);
    sample.framesDropped = num(inbound.framesDropped);
    sample.frameWidth = num(inbound.frameWidth);
    sample.frameHeight = num(inbound.frameHeight);
    sample.jitterBufferDelay = num(inbound.jitterBufferDelay);
    sample.jitterBufferEmittedCount = num(inbound.jitterBufferEmittedCount);
    sample.totalDecodeTime = num(inbound.totalDecodeTime);
    sample.totalProcessingDelay = num(inbound.totalProcessingDelay);
    const codec = byId.get(str(inbound.codecId) ?? '');
    sample.mimeType = str(codec?.mimeType);
  }
  if (pair) {
    sample.rttS = num(pair.currentRoundTripTime);
    const local = byId.get(str(pair.localCandidateId) ?? '');
    sample.protocol = str(local?.protocol);
    sample.candidateType = str(local?.candidateType);
  }
  return sample;
}

// codecFromMime names the codec of a "video/H264" style MIME type.
export function codecFromMime(mime: string | undefined): string | null {
  if (!mime) return null;
  const name = mime.split('/')[1]?.toUpperCase() ?? '';
  if (name === 'H264') return 'H.264';
  if (name === 'H265' || name === 'HEVC') return 'H.265';
  return name || null;
}

// perFrameMs is the average of a total-seconds counter over the frames counted
// in the same interval, in milliseconds.
function perFrameMs(
  total: number | undefined,
  totalBefore: number | undefined,
  frames: number | undefined,
  framesBefore: number | undefined
): number | null {
  if (total === undefined || totalBefore === undefined) return null;
  if (frames === undefined || framesBefore === undefined) return null;
  const count = frames - framesBefore;
  const seconds = total - totalBefore;
  if (count <= 0 || seconds < 0) return null;
  return (seconds / count) * 1000;
}

// webrtcEstimate is the part of the latency the browser sees for itself: half
// the round trip for the way here, then from a frame's first packet to its
// decode (Chrome's totalProcessingDelay, which takes in the jitter buffer), or
// failing that the jitter buffer delay and the decode time added up. Null
// until a whole interval has frames in it.
export function webrtcEstimate(current: WebrtcSample, previous: WebrtcSample): number | null {
  const parts = webrtcEstimateParts(current, previous);
  return parts === null ? null : parts.receiveMs + parts.networkMs;
}

// webrtcEstimateParts is webrtcEstimate before the two parts are added.
export function webrtcEstimateParts(
  current: WebrtcSample,
  previous: WebrtcSample
): LatencyParts | null {
  let receive = perFrameMs(
    current.totalProcessingDelay,
    previous.totalProcessingDelay,
    current.framesDecoded,
    previous.framesDecoded
  );
  if (receive === null) {
    const buffer = perFrameMs(
      current.jitterBufferDelay,
      previous.jitterBufferDelay,
      current.jitterBufferEmittedCount,
      previous.jitterBufferEmittedCount
    );
    const decode = perFrameMs(
      current.totalDecodeTime,
      previous.totalDecodeTime,
      current.framesDecoded,
      previous.framesDecoded
    );
    if (buffer === null || decode === null) return null;
    receive = buffer + decode;
  }
  const way = current.rttS !== undefined ? (current.rttS * 1000) / 2 : 0;
  return { networkMs: way, receiveMs: receive };
}

// withBoardPart completes a WebRTC estimate with the board's share of the
// latency (GET /api/stream/info boardLatency.webrtc) and the display's: the
// sum of measured parts. Anything else, or no board share, is left as it is.
export function withBoardPart(latency: Latency, boardMs: number | undefined): Latency {
  if (latency.kind !== 'estimate' || !latency.parts) return latency;
  if (typeof boardMs !== 'number' || !Number.isFinite(boardMs) || boardMs < 0) return latency;
  const parts = { ...latency.parts, boardMs };
  const ms = boardMs + parts.networkMs + parts.receiveMs + (parts.displayMs ?? 0);
  return { ...latency, ms, parts };
}

// refreshIntervalMs is the display's refresh interval from the gaps between
// requestAnimationFrame callbacks: their median, so that a frame the page
// missed (a gap of two) does not count. Gaps outside 4 to 100 ms (250 Hz to
// 10 Hz) are a page that was busy or hidden. Null without enough gaps.
export function refreshIntervalMs(gaps: readonly number[]): number | null {
  const usable = gaps.filter((g) => g >= 4 && g <= 100).sort((a, b) => a - b);
  if (usable.length < 10) return null;
  const middle = usable.length >> 1;
  return usable.length % 2 ? usable[middle] : (usable[middle - 1] + usable[middle]) / 2;
}

// lossPct is the share of the interval's packets that never came.
export function lossPct(current: WebrtcSample, previous: WebrtcSample): number | null {
  const { packetsReceived: received, packetsLost: lost } = current;
  const { packetsReceived: receivedBefore, packetsLost: lostBefore } = previous;
  if (received === undefined || lost === undefined) return null;
  if (receivedBefore === undefined || lostBefore === undefined) return null;
  const gotten = received - receivedBefore;
  // Lost can go down: a late packet counted lost arrives after all.
  const missing = Math.max(0, lost - lostBefore);
  if (gotten < 0 || gotten + missing === 0) return null;
  return (missing / (gotten + missing)) * 100;
}

// The NTP epoch (1900) is this many milliseconds before the Unix one (1970).
export const NTP_UNIX_OFFSET_MS = 2208988800000;

// An answer outside this range is a clock that is not what the arithmetic
// assumed, not a latency.
const MAX_PLAUSIBLE_LATENCY_MS = 10_000;

// What RTCRtpReceiver.getSynchronizationSources() gives for a source.
export type SyncSource = {
  timestamp?: number;
  captureTimestamp?: number;
  senderCaptureTimeOffset?: number;
};

// captureLatency reads the abs-capture-time RTP header extension as Chrome
// passes it on. captureTimestamp is the frame's capture time on the board's
// wall clock (NTP epoch), and timestamp when the frame went to the track, on
// this page's clock (performance.timeOrigin + performance.now(), Unix epoch).
//
// With clock, the offset between the two clocks measured over the signalling
// socket, the capture time is moved onto the page's clock by it, and the
// answer carries the offset's error. senderCaptureTimeOffset is not used
// then: the board sends 0 for it, and anything else is the browser's own
// guess at the same offset, which would count it twice.
//
// Without clock the arithmetic is the old one, which takes the two clocks to
// agree (senderCaptureTimeOffset is the board's 0 unless the browser filled
// in a guess), and the answer says so. Null when there is no capture time, or
// the arithmetic gives no plausible answer.
export function captureLatency(
  source: SyncSource | undefined,
  clock?: WallClockOffset | null
): Latency | null {
  if (!source) return null;
  const { timestamp, captureTimestamp, senderCaptureTimeOffset } = source;
  if (typeof timestamp !== 'number' || typeof captureTimestamp !== 'number') return null;
  if (!(captureTimestamp > 0)) return null;

  if (clock) {
    const captured = captureTimestamp - NTP_UNIX_OFFSET_MS - clock.offsetMs;
    const ms = timestamp - captured;
    // Within the offset's error of zero is a fast frame, not a broken clock.
    if (!(ms >= -clock.errorMs && ms <= MAX_PLAUSIBLE_LATENCY_MS)) return null;
    return {
      ms: Math.max(0, ms),
      kind: 'capture',
      errorMs: clock.errorMs,
      clockOffsetMs: clock.offsetMs
    };
  }

  const offset = typeof senderCaptureTimeOffset === 'number' ? senderCaptureTimeOffset : 0;
  const ms = timestamp - (captureTimestamp + offset - NTP_UNIX_OFFSET_MS);
  if (!(ms >= 0 && ms <= MAX_PLAUSIBLE_LATENCY_MS)) return null;
  return { ms, kind: 'capture', clockAssumed: true };
}

// transportLabel says how the media travels: "UDP host", "TCP relay".
export function transportLabel(sample: WebrtcSample): string | null {
  const protocol = sample.protocol?.toUpperCase();
  const type = sample.candidateType;
  if (!protocol && !type) return null;
  return [protocol, type].filter(Boolean).join(' ');
}

// What webrtcStats takes besides the two samples, where known.
export type WebrtcExtras = {
  // The video source's newest capture time.
  source?: SyncSource;
  // The board's wall clock against this page's.
  clock?: WallClockOffset | null;
  // The display's refresh interval (refreshIntervalMs).
  refreshMs?: number | null;
};

// webrtcStats is what the bar shows for a WebRTC session, from two samples a
// second or so apart, the video source's capture time if it has one, and the
// clock offset and the display refresh where known.
export function webrtcStats(
  current: WebrtcSample,
  previous: WebrtcSample | null,
  connection: ConnectionState,
  extras: WebrtcExtras = {}
): StreamStats {
  const stats = emptyStats('h264', connection);
  stats.codec = codecFromMime(current.mimeType);
  stats.width = current.frameWidth ?? 0;
  stats.height = current.frameHeight ?? 0;
  stats.rttMs = current.rttS !== undefined ? current.rttS * 1000 : null;
  stats.transport = transportLabel(current);
  stats.latency = captureLatency(extras.source, extras.clock);

  if (!previous) return stats;

  const seconds = (current.at - previous.at) / 1000;
  stats.fpsReceived = rate(current.framesReceived, previous.framesReceived, seconds);
  stats.fpsDecoded = rate(current.framesDecoded, previous.framesDecoded, seconds);
  stats.fpsDropped = rate(current.framesDropped, previous.framesDropped, seconds);
  const bytes = rate(current.bytesReceived, previous.bytesReceived, seconds);
  stats.kbps = bytes === null ? null : (bytes * 8) / 1000;
  stats.lossPct = lossPct(current, previous);
  if (!stats.latency) {
    const parts = webrtcEstimateParts(current, previous);
    if (parts !== null) {
      if (typeof extras.refreshMs === 'number') parts.displayMs = extras.refreshMs / 2;
      stats.latency = { ms: parts.networkMs + parts.receiveMs, kind: 'estimate', parts };
    }
  }
  return stats;
}

// ---- the direct path's clock ----------------------------------------------

// A clock sample: this page asked the board for its clock at sentMs, the
// answer said boardUs, and it came back at receivedMs (performance.now()).
export type ClockSample = { sentMs: number; receivedMs: number; boardUs: number };

export type ClockEstimate = {
  // board microseconds = page milliseconds * 1000 + offsetUs
  offsetUs: number;
  // How far off the offset may be, either way: half the round trip.
  errorMs: number;
};

// The samples the estimate is chosen from. Ten seconds at one a second: long
// enough for one quick answer among them, short enough to follow drift.
export const CLOCK_WINDOW = 10;

// clockEstimate picks the sample with the shortest round trip, as NTP does.
// The board stamped its clock somewhere inside that round trip, and the
// middle of it is the best guess, wrong by at most half of it.
export function clockEstimate(samples: readonly ClockSample[]): ClockEstimate | null {
  const best = quickestReading(samples);
  if (!best) return null;
  return {
    offsetUs: best.boardUs - midpointMs(best) * 1000,
    errorMs: roundTripMs(best) / 2
  };
}

// frameAgeMs is how long ago, at page time nowMs, the board stamped a frame
// stamped frameUs.
export function frameAgeMs(frameUs: number, nowMs: number, clock: ClockEstimate): number {
  return (nowMs * 1000 + clock.offsetUs - frameUs) / 1000;
}

// ---- formatting ------------------------------------------------------------

export function formatBitrate(kbps: number): string {
  if (kbps >= 1000) return `${(kbps / 1000).toFixed(kbps >= 10_000 ? 0 : 1)} Mb/s`;
  return `${Math.round(kbps)} kb/s`;
}

export function formatRefresh(milliHz: number): string {
  const hz = milliHz / 1000;
  // 60 and 59.94 are both common, so keep two places when they mean something.
  return Number.isInteger(Math.round(hz * 100) / 100)
    ? `${Math.round(hz)} Hz`
    : `${hz.toFixed(2)} Hz`;
}

export function formatFps(value: number): string {
  return value >= 10 ? String(Math.round(value)) : value.toFixed(1).replace(/\.0$/, '');
}

// formatLatency writes a latency with the sign that says how sure it is: "~"
// for an estimate or a clock taken on trust, "±" for a clock error that
// matters (over 1 ms).
export function formatLatency(latency: Latency): string {
  const ms = Math.round(latency.ms);
  const approx = latency.kind === 'estimate' || latency.clockAssumed ? '~' : '';
  const error =
    latency.errorMs !== undefined && latency.errorMs >= 1 ? ` ±${Math.ceil(latency.errorMs)}` : '';
  return `${approx}${ms}${error} ms`;
}
