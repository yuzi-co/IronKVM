// What the stream status bar shows, and the arithmetic behind it, kept free of
// the DOM and of '@/' imports so node's test runner can load it.
//
// Every number comes from what the browser measured itself, apart from the
// source mode and the board's capture rate (use-source-info.ts). Latency is
// only ever one of three things, and the bar says which:
//
//   capture  capture to display, from the abs-capture-time RTP header
//            extension (WebRTC, when the server sends it)
//   encoder  the frame leaving the encoder to it being drawn, against the
//            board's clock (H.264 direct, clock-synchronised over the socket)
//   estimate what the browser can see of the path: half the round trip plus
//            the time from a frame's first packet to its decode. It leaves out
//            capture and encoding on the board, so it is a lower bound.

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
  return receive + way;
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
// passes it on. captureTimestamp is the frame's capture time on the sender's
// NTP clock, and senderCaptureTimeOffset what the browser estimates that clock
// to be off its own (from the sender reports). timestamp is when the frame
// went to the track, on the Unix clock. Without the offset the two clocks are
// taken to agree, and the answer says so. Null when there is no capture time,
// or the arithmetic gives no plausible answer.
export function captureLatency(source: SyncSource | undefined): Latency | null {
  if (!source) return null;
  const { timestamp, captureTimestamp, senderCaptureTimeOffset } = source;
  if (typeof timestamp !== 'number' || typeof captureTimestamp !== 'number') return null;
  if (!(captureTimestamp > 0)) return null;
  const hasOffset = typeof senderCaptureTimeOffset === 'number';
  const captured =
    captureTimestamp + (hasOffset ? senderCaptureTimeOffset : 0) - NTP_UNIX_OFFSET_MS;
  const ms = timestamp - captured;
  if (!(ms >= 0 && ms <= MAX_PLAUSIBLE_LATENCY_MS)) return null;
  return hasOffset ? { ms, kind: 'capture' } : { ms, kind: 'capture', clockAssumed: true };
}

// transportLabel says how the media travels: "UDP host", "TCP relay".
export function transportLabel(sample: WebrtcSample): string | null {
  const protocol = sample.protocol?.toUpperCase();
  const type = sample.candidateType;
  if (!protocol && !type) return null;
  return [protocol, type].filter(Boolean).join(' ');
}

// webrtcStats is what the bar shows for a WebRTC session, from two samples a
// second or so apart and the video source's capture time if it has one.
export function webrtcStats(
  current: WebrtcSample,
  previous: WebrtcSample | null,
  connection: ConnectionState,
  source?: SyncSource
): StreamStats {
  const stats = emptyStats('h264', connection);
  stats.codec = codecFromMime(current.mimeType);
  stats.width = current.frameWidth ?? 0;
  stats.height = current.frameHeight ?? 0;
  stats.rttMs = current.rttS !== undefined ? current.rttS * 1000 : null;
  stats.transport = transportLabel(current);
  stats.latency = captureLatency(source);

  if (!previous) return stats;

  const seconds = (current.at - previous.at) / 1000;
  stats.fpsReceived = rate(current.framesReceived, previous.framesReceived, seconds);
  stats.fpsDecoded = rate(current.framesDecoded, previous.framesDecoded, seconds);
  stats.fpsDropped = rate(current.framesDropped, previous.framesDropped, seconds);
  const bytes = rate(current.bytesReceived, previous.bytesReceived, seconds);
  stats.kbps = bytes === null ? null : (bytes * 8) / 1000;
  stats.lossPct = lossPct(current, previous);
  if (!stats.latency) {
    const estimate = webrtcEstimate(current, previous);
    if (estimate !== null) stats.latency = { ms: estimate, kind: 'estimate' };
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
  let best: ClockSample | null = null;
  for (const sample of samples) {
    const rtt = sample.receivedMs - sample.sentMs;
    if (rtt < 0) continue;
    if (!best || rtt < best.receivedMs - best.sentMs) best = sample;
  }
  if (!best) return null;
  const middleMs = (best.sentMs + best.receivedMs) / 2;
  return {
    offsetUs: best.boardUs - middleMs * 1000,
    errorMs: (best.receivedMs - best.sentMs) / 2
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
