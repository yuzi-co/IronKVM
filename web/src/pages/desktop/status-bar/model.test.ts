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
  webrtcSample,
  webrtcStats,
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

test('abs-capture-time gives capture-to-display latency', () => {
  const capturedUnix = 1_700_000_000_000;
  const source = {
    timestamp: capturedUnix + 48,
    captureTimestamp: capturedUnix + NTP_UNIX_OFFSET_MS + 5,
    senderCaptureTimeOffset: -5
  };
  assert.deepEqual(captureLatency(source), { ms: 48, kind: 'capture' });
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
  const stats = webrtcStats(after, before, 'connected', {
    timestamp: unix + 60,
    captureTimestamp: unix + NTP_UNIX_OFFSET_MS,
    senderCaptureTimeOffset: 0
  });
  assert.deepEqual(stats.latency, { ms: 60, kind: 'capture' });
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
