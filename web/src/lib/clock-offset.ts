// The offset between this page's clock and the board's, measured NTP-style:
// the page notes when it asks for the board's clock and when the answer comes
// back, and takes the board to have read its clock in the middle of that round
// trip. That guess is wrong by at most half the round trip, either way, so of
// several readings the one with the shortest round trip is kept.
//
// H.264 Direct asks over its own socket (direct.worker.ts, the board's
// monotonic clock in microseconds); WebRTC asks over the signalling socket
// (the board's wall clock, which its abs-capture-time stamps are on). Both
// pick their reading here.
//
// This module runs under node --test as well as in the browser, so it imports
// nothing.

// One round trip: the page asked at sentMs and had the answer at receivedMs,
// both on the page's clock.
export type ClockReading = { sentMs: number; receivedMs: number };

export function roundTripMs(reading: ClockReading): number {
  return reading.receivedMs - reading.sentMs;
}

export function midpointMs(reading: ClockReading): number {
  return (reading.sentMs + reading.receivedMs) / 2;
}

// quickestReading is the reading with the shortest round trip, as NTP picks
// its sample, or null when there is none. A negative round trip is a clock
// that moved under the reading, and is skipped.
export function quickestReading<T extends ClockReading>(readings: readonly T[]): T | null {
  let best: T | null = null;
  for (const reading of readings) {
    const rtt = roundTripMs(reading);
    if (!(rtt >= 0)) continue;
    if (!best || rtt < roundTripMs(best)) best = reading;
  }
  return best;
}

// ---- the board's wall clock (WebRTC) ---------------------------------------

// A reading of the board's wall clock: Unix milliseconds on the board, between
// sentMs and receivedMs on the page's clock (see pageClockMs).
export type WallClockReading = ClockReading & { boardMs: number };

export type WallClockOffset = {
  // board Unix ms = page clock ms + offsetMs
  offsetMs: number;
  // How far off the offset may be, either way: half the chosen round trip.
  errorMs: number;
};

// wallClockOffset is the offset the quickest reading gives, or null.
export function wallClockOffset(readings: readonly WallClockReading[]): WallClockOffset | null {
  const best = quickestReading(readings);
  if (!best) return null;
  return { offsetMs: best.boardMs - midpointMs(best), errorMs: roundTripMs(best) / 2 };
}

// pageClockMs is the page's clock as Unix milliseconds: the clock that
// RTCRtpSynchronizationSource.timestamp is defined on (W3C WebRTC, "the
// timestamp is defined as performance.timeOrigin + performance.now()").
// Unlike Date.now() it does not jump when the system clock is set while the
// page is open, so readings taken minutes apart still compare.
export function pageClockMs(): number {
  return performance.timeOrigin + performance.now();
}

// ---- sampling ----------------------------------------------------------------

export type ClockSamplerOptions = {
  // Readings per round, one after another.
  burst: number;
  // Between the starts of two rounds.
  everyMs: number;
  // How long a reading stays eligible. Two rounds' worth: long enough that
  // one slow round does not lose the estimate, short enough that a clock that
  // drifts or is stepped is followed within a minute.
  keepMs: number;
  // An answer later than this is given up on, and the round goes on.
  timeoutMs: number;
  // When the socket cannot send, the round is tried again this much later.
  retryMs: number;
  now: () => number;
  setTimer: (fn: () => void, ms: number) => unknown;
  clearTimer: (timer: unknown) => void;
};

export const DEFAULT_SAMPLER_OPTIONS: ClockSamplerOptions = {
  burst: 5,
  everyMs: 30_000,
  keepMs: 65_000,
  timeoutMs: 2_000,
  retryMs: 1_000,
  now: pageClockMs,
  setTimer: (fn, ms) => setTimeout(fn, ms),
  clearTimer: (timer) => clearTimeout(timer as ReturnType<typeof setTimeout>)
};

// ClockSampler asks for the board's wall clock in rounds of a few readings,
// one round at start and one every half minute, and keeps the quickest recent
// reading. send asks once, with an id the answer must carry back, and answers
// whether the request went out; the caller hands each answer to onReply.
export class ClockSampler {
  private readonly opts: ClockSamplerOptions;
  private readonly send: (id: number) => boolean;
  private readings: (WallClockReading & { at: number })[] = [];
  private pendingSentMs: number | null = null;
  private pendingId = 0;
  private left = 0;
  private timer: unknown = null;
  private running = false;

  constructor(send: (id: number) => boolean, options: Partial<ClockSamplerOptions> = {}) {
    this.send = send;
    this.opts = { ...DEFAULT_SAMPLER_OPTIONS, ...options };
  }

  // start begins a round now, unless the sampler already runs.
  start() {
    if (this.running) return;
    this.running = true;
    this.round();
  }

  stop() {
    this.running = false;
    this.pendingSentMs = null;
    this.left = 0;
    this.cancelTimer();
  }

  // onReply takes the board's clock, Unix milliseconds, as the answer to the
  // request id. An answer nobody waits for is dropped: its request timed out,
  // and its round trip is not known.
  onReply(boardMs: number, id: number) {
    if (this.pendingSentMs === null || id !== this.pendingId) return;
    if (!Number.isFinite(boardMs)) return;
    const receivedMs = this.opts.now();
    this.readings.push({ sentMs: this.pendingSentMs, receivedMs, boardMs, at: receivedMs });
    this.pendingSentMs = null;
    this.cancelTimer();
    this.next();
  }

  // estimate is the offset from the quickest reading still kept, or null.
  estimate(): WallClockOffset | null {
    const now = this.opts.now();
    this.readings = this.readings.filter((r) => now - r.at <= this.opts.keepMs);
    return wallClockOffset(this.readings);
  }

  private round() {
    if (!this.running) return;
    this.left = this.opts.burst;
    this.next();
  }

  // next asks once more in this round, or waits for the next round.
  private next() {
    if (!this.running) return;
    if (this.left <= 0) {
      this.schedule(() => this.round(), this.opts.everyMs);
      return;
    }
    const id = this.pendingId + 1;
    const sentMs = this.opts.now();
    if (!this.send(id)) {
      this.schedule(() => this.next(), this.opts.retryMs);
      return;
    }
    this.left--;
    this.pendingId = id;
    this.pendingSentMs = sentMs;
    this.schedule(() => {
      this.pendingSentMs = null;
      this.next();
    }, this.opts.timeoutMs);
  }

  private schedule(fn: () => void, ms: number) {
    this.cancelTimer();
    this.timer = this.opts.setTimer(() => {
      this.timer = null;
      fn();
    }, ms);
  }

  private cancelTimer() {
    if (this.timer !== null) {
      this.opts.clearTimer(this.timer);
      this.timer = null;
    }
  }
}

// ---- the WebRTC signalling socket ----------------------------------------------

// The signalling event that asks for the board's wall clock (server feature
// webrtc-clock). The request's data is the sampler's id; the answer's is the
// board's Unix microseconds, a space, and the id.
export const CLOCK_EVENT = 'clock';

// parseClockReply reads the answer's data, or null for anything else.
export function parseClockReply(data: unknown): { boardMs: number; id: number } | null {
  if (typeof data !== 'string') return null;
  const match = /^(\d{1,17}) (\d{1,15})$/.exec(data);
  if (!match) return null;
  return { boardMs: Number(match[1]) / 1000, id: Number(match[2]) };
}
