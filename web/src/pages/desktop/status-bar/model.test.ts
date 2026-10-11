import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  captureLatency,
  clockEstimate,
  codecFromMime,
  formatBitrate,
  formatFps,
  formatLatency,
  formatRefresh,
  frameAgeMs,
  lossPct,
  NTP_UNIX_OFFSET_MS,
  rate,
  refreshIntervalMs,
  webrtcSample,
  webrtcStats,
  withBoardPart,
  type WebrtcSample
} from './model.ts';

test('a rate needs two readings, time between them, and a counter that went up', () => {
  assert.equal(rate(130, 100, 1), 30);
  assert.equal(rate(130, 100, 2), 15);
  assert.equal(rate(130, undefined, 1), null);
  assert.equal(rate(130, 100, 0), null);
  // A new session starts its counters again.
  assert.equal(rate(5, 100, 1), null);
});

// A trimmed getStats() report from Chrome, as entries.
function report(frames: number, bytes: number, processing: number, lost = 0, received = 1000) {
  return [
    { id: 'T1', type: 'transport', selectedCandidatePairId: 'CP1' },
    {
      id: 'CP0',
      type: 'candidate-pair',
      nominated: true,
      currentRoundTripTime: 0.5,
      localCandidateId: 'L0'
    },
    { id: 'CP1', type: 'candidate-pair', currentRoundTripTime: 0.004, localCandidateId: 'L1' },
    { id: 'L0', type: 'local-candidate', protocol: 'tcp', candidateType: 'relay' },
    { id: 'L1', type: 'local-candidate', protocol: 'udp', candidateType: 'host' },
    { id: 'C1', type: 'codec', mimeType: 'video/H264' },
    { id: 'A1', type: 'inbound-rtp', kind: 'audio', framesDecoded: 0 },
    {
      id: 'V1',
      type: 'inbound-rtp',
      kind: 'video',
      codecId: 'C1',
      bytesReceived: bytes,
      packetsReceived: received,
      packetsLost: lost,
      framesReceived: frames + 1,
      framesDecoded: frames,
      framesDropped: 2,
      frameWidth: 1920,
      frameHeight: 1080,
      totalProcessingDelay: processing,
      jitterBufferDelay: 0,
      jitterBufferEmittedCount: frames,
      totalDecodeTime: 0
    }
  ];
}

test('a WebRTC sample reads the video entry and the pair the transport selected', () => {
  const sample = webrtcSample(report(300, 1_000_000, 6), 1000);
  assert.equal(sample.framesDecoded, 300);
  assert.equal(sample.mimeType, 'video/H264');
  assert.equal(sample.rttS, 0.004);
  assert.equal(sample.protocol, 'udp');
  assert.equal(sample.candidateType, 'host');
});

test('WebRTC rates, bitrate and the receive-side estimate over one second', () => {
  const before = webrtcSample(report(300, 1_000_000, 6), 1000);
  const after = webrtcSample(report(330, 1_500_000, 6.6), 2000);
  const stats = webrtcStats(after, before, 'connected');
  assert.equal(stats.codec, 'H.264');
  assert.equal(stats.width, 1920);
  assert.equal(stats.fpsDecoded, 30);
  assert.equal(stats.fpsDropped, 0);
  assert.equal(stats.kbps, 4000);
  assert.equal(stats.transport, 'UDP host');
  // 0.6 s over 30 frames is 20 ms, and half of a 4 ms round trip is 2.
  assert.equal(stats.latency?.kind, 'estimate');
  assert.ok(Math.abs((stats.latency?.ms ?? 0) - 22) < 1e-9);
});

test('the estimate falls back to jitter buffer plus decode time', () => {
  const before: WebrtcSample = {
    at: 0,
    framesDecoded: 0,
    jitterBufferDelay: 0,
    jitterBufferEmittedCount: 0,
    totalDecodeTime: 0
  };
  const after: WebrtcSample = {
    at: 1000,
    framesDecoded: 10,
    jitterBufferDelay: 0.1,
    jitterBufferEmittedCount: 10,
    totalDecodeTime: 0.05
  };
  const stats = webrtcStats(after, before, 'connected');
  assert.ok(Math.abs((stats.latency?.ms ?? 0) - 15) < 1e-9);
});

test('a first sample has no rates and no estimate', () => {
  const stats = webrtcStats(webrtcSample(report(300, 1, 6), 0), null, 'connected');
  assert.equal(stats.fpsDecoded, null);
  assert.equal(stats.kbps, null);
  assert.equal(stats.latency, null);
});

test('loss is the interval share of packets that never came', () => {
  const at = (lost: number, received: number): WebrtcSample => ({
    at: 0,
    packetsLost: lost,
    packetsReceived: received
  });
  assert.equal(lossPct(at(10, 1990), at(0, 1000)), 1);
  assert.equal(lossPct(at(0, 1000), at(0, 1000)), null);
  // A packet first counted lost and then received.
  assert.equal(lossPct(at(4, 1100), at(5, 1000)), 0);
});

test('without a measured offset, the old arithmetic, taken on trust', () => {
  const capturedUnix = 1_700_000_000_000;
  const source = {
    timestamp: capturedUnix + 48,
    captureTimestamp: capturedUnix + NTP_UNIX_OFFSET_MS + 5,
    senderCaptureTimeOffset: -5
  };
  // The board's 0, or a guess the browser filled in, says nothing about this
  // page's clock: still assumed.
  assert.deepEqual(captureLatency(source), { ms: 48, kind: 'capture', clockAssumed: true });
});

// The board's clock is right, and the page's is skewMs ahead of it: at the
// moment the board reads boardMs, the page reads boardMs + skewMs. A frame captured at capturedBoardMs reaches the track latencyMs later, which
// the page stamps on its own clock.
function skewedSource(capturedBoardMs: number, latencyMs: number, skewMs: number) {
  return {
    timestamp: capturedBoardMs + latencyMs + skewMs,
    captureTimestamp: capturedBoardMs + NTP_UNIX_OFFSET_MS,
    senderCaptureTimeOffset: 0
  };
}

for (const skewMs of [500, -500, 420]) {
  test(`a page clock ${skewMs} ms off is corrected by the measured offset`, () => {
    const captured = 1_700_000_000_000;
    const source = skewedSource(captured, 37, skewMs);

    // Uncorrected, the skew is all there is to see, or nothing at all.
    const assumed = captureLatency(source);
    if (skewMs > 0) assert.ok(assumed && Math.abs(assumed.ms - (37 + skewMs)) < 1e-6);
    else assert.equal(assumed, null);

    // board = page + offset, so the offset is -skew, measured within 3 ms.
    const clock = { offsetMs: -skewMs + 2, errorMs: 3 };
    const latency = captureLatency(source, clock);
    assert.ok(latency);
    assert.equal(latency.kind, 'capture');
    assert.equal(latency.clockAssumed, undefined);
    assert.equal(latency.errorMs, 3);
    assert.equal(latency.clockOffsetMs, -skewMs + 2);
    // Off by the offset's own error, 2 ms, inside its bound.
    assert.ok(Math.abs(latency.ms - 37) <= 3, `got ${latency.ms}`);
    assert.equal(formatLatency(latency), '39 ±3 ms');
  });
}

test('with a measured offset, senderCaptureTimeOffset is not counted twice', () => {
  const captured = 1_700_000_000_000;
  const source = { ...skewedSource(captured, 40, 420), senderCaptureTimeOffset: 418 };
  const latency = captureLatency(source, { offsetMs: -420, errorMs: 1 });
  assert.equal(latency?.ms, 40);
});

test('a corrected latency just under zero is a fast frame within the error', () => {
  const captured = 1_700_000_000_000;
  const source = skewedSource(captured, 2, 0);
  // The offset is 4 ms out, which the error bound of 5 allows.
  assert.equal(captureLatency(source, { offsetMs: -4, errorMs: 5 })?.ms, 0);
  assert.equal(captureLatency(source, { offsetMs: -40, errorMs: 5 }), null);
});

test('a capture time without an offset is taken on trust, and says so', () => {
  const capturedUnix = 1_700_000_000_000;
  const latency = captureLatency({
    timestamp: capturedUnix + 30,
    captureTimestamp: capturedUnix + NTP_UNIX_OFFSET_MS
  });
  assert.deepEqual(latency, { ms: 30, kind: 'capture', clockAssumed: true });
  assert.equal(formatLatency(latency!), '~30 ms');
});

test('no capture time, or an impossible one, is no latency', () => {
  assert.equal(captureLatency(undefined), null);
  assert.equal(captureLatency({ timestamp: 1000 }), null);
  const unix = 1_700_000_000_000;
  // A board clock an hour ahead.
  assert.equal(
    captureLatency({ timestamp: unix, captureTimestamp: unix + NTP_UNIX_OFFSET_MS + 3_600_000 }),
    null
  );
});

test('a capture latency replaces the estimate', () => {
  const unix = 1_700_000_000_000;
  const before = webrtcSample(report(300, 1_000_000, 6), 1000);
  const after = webrtcSample(report(330, 1_500_000, 6.6), 2000);
  const source = {
    timestamp: unix + 60,
    captureTimestamp: unix + NTP_UNIX_OFFSET_MS,
    senderCaptureTimeOffset: 0
  };
  assert.deepEqual(webrtcStats(after, before, 'connected', { source }).latency, {
    ms: 60,
    kind: 'capture',
    clockAssumed: true
  });
  assert.deepEqual(
    webrtcStats(after, before, 'connected', { source, clock: { offsetMs: 0, errorMs: 1.5 } })
      .latency,
    { ms: 60, kind: 'capture', errorMs: 1.5, clockOffsetMs: 0 }
  );
});

test('the estimate is a sum of measured parts once the board reports its share', () => {
  const before = webrtcSample(report(300, 1_000_000, 6), 1000);
  const after = webrtcSample(report(330, 1_500_000, 6.6), 2000);
  const stats = webrtcStats(after, before, 'connected', { refreshMs: 1000 / 60 });
  const latency = stats.latency;
  assert.ok(latency && latency.parts);
  // Without the board's share it stays what it was: the browser's part only.
  assert.ok(Math.abs(latency.ms - 22) < 1e-9);
  assert.equal(withBoardPart(latency, undefined), latency);

  const summed = withBoardPart(latency, 24.5);
  assert.equal(summed.kind, 'estimate');
  assert.equal(summed.parts?.boardMs, 24.5);
  // 24.5 on the board, 2 on the way, 20 buffering and decoding, 8.33 display.
  assert.ok(Math.abs(summed.ms - (24.5 + 2 + 20 + 1000 / 120)) < 1e-9, `got ${summed.ms}`);
  assert.equal(formatLatency(summed), '~55 ms');

  // A capture latency is never touched.
  const capture = { ms: 40, kind: 'capture' as const };
  assert.equal(withBoardPart(capture, 24.5), capture);
  assert.equal(withBoardPart(latency, Number.NaN), latency);
});

// One frame's path, with every part known, as the two measurements see it.
// The capture measurement stops at the frame reaching the track (after
// decode); the sum goes on to the display. They differ by the display part.
test('the summed estimate and the capture measurement agree on one timeline', () => {
  const capturedBoardMs = 1_700_000_000_000;
  const boardMs = 24; // capture to the last packet leaving the board
  const oneWayMs = 3; // half of a 6 ms round trip
  const receiveMs = 18; // first packet to decoded, 30 frames
  const refreshMs = 1000 / 60;

  const before = webrtcSample(report(300, 1_000_000, 6), 1000);
  const afterReport = report(330, 1_500_000, 6 + (receiveMs * 30) / 1000).map((entry) =>
    entry.type === 'candidate-pair' ? { ...entry, currentRoundTripTime: 0.006 } : entry
  );
  const after = webrtcSample(afterReport, 2000);
  const estimate = webrtcStats(after, before, 'connected', { refreshMs }).latency;
  assert.ok(estimate);
  const summed = withBoardPart(estimate, boardMs);

  const skewMs = 420;
  const reachedTrack = capturedBoardMs + boardMs + oneWayMs + receiveMs;
  const capture = captureLatency(
    {
      timestamp: reachedTrack + skewMs,
      captureTimestamp: capturedBoardMs + NTP_UNIX_OFFSET_MS,
      senderCaptureTimeOffset: 0
    },
    { offsetMs: -skewMs, errorMs: 3 }
  );
  assert.ok(capture);
  assert.ok(Math.abs(capture.ms - 45) < 1e-6, `capture ${capture.ms}`);
  assert.ok(Math.abs(summed.ms - capture.ms - refreshMs / 2) < 1e-6, `sum ${summed.ms}`);
});

test('the refresh interval is the median gap between animation frames', () => {
  const at60 = Array.from({ length: 30 }, (_, i) => (i % 2 ? 16.5 : 16.8));
  assert.ok(Math.abs((refreshIntervalMs(at60) ?? 0) - 16.65) < 1e-9);
  // A missed frame and a hidden tab do not move it.
  assert.equal(refreshIntervalMs([...Array(20).fill(6.94), 13.9, 2000]), 6.94);
  assert.equal(refreshIntervalMs([16.7, 16.7]), null);
});

test('the clock estimate takes the quickest answer, wrong by half its round trip', () => {
  const clock = clockEstimate([
    { sentMs: 100, receivedMs: 140, boardUs: 5_000_000 },
    { sentMs: 1100, receivedMs: 1102, boardUs: 5_990_000 },
    { sentMs: 2100, receivedMs: 2120, boardUs: 7_000_000 }
  ]);
  assert.ok(clock);
  assert.equal(clock.errorMs, 1);
  // The board said 5.99 s at page time 1101 ms.
  assert.equal(clock.offsetUs, 5_990_000 - 1_101_000);
  // A frame stamped 5.95 s on the board, drawn at page time 1101 ms, is 40 ms old.
  assert.equal(frameAgeMs(5_950_000, 1101, clock), 40);
  assert.equal(clockEstimate([]), null);
});

test('formatting', () => {
  assert.equal(codecFromMime('video/H265'), 'H.265');
  assert.equal(codecFromMime(undefined), null);
  assert.equal(formatBitrate(850), '850 kb/s');
  assert.equal(formatBitrate(4200), '4.2 Mb/s');
  assert.equal(formatBitrate(12_400), '12 Mb/s');
  assert.equal(formatRefresh(60_000), '60 Hz');
  assert.equal(formatRefresh(59_940), '59.94 Hz');
  assert.equal(formatRefresh(59_999), '60 Hz');
  assert.equal(formatFps(29.7), '30');
  assert.equal(formatFps(2.5), '2.5');
  assert.equal(formatFps(0), '0');
  assert.equal(formatLatency({ ms: 41.6, kind: 'encoder', errorMs: 2.2 }), '42 ±3 ms');
  assert.equal(formatLatency({ ms: 41.6, kind: 'encoder', errorMs: 0.3 }), '42 ms');
  assert.equal(formatLatency({ ms: 12, kind: 'estimate' }), '~12 ms');
});
