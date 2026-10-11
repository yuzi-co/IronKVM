import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  ABS_CAPTURE_TIME_URI,
  captureLatencyMs,
  NTP_EPOCH_OFFSET_MS,
  offerAbsCaptureTime
} from './abs-capture-time.ts';

test('turns a stopped abs-capture-time on, leaving the rest', () => {
  let set: { uri: string; direction: string }[] | undefined;
  const transceiver = {
    getHeaderExtensionsToNegotiate: () => [
      { uri: 'urn:ietf:params:rtp-hdrext:toffset', direction: 'sendrecv' },
      { uri: ABS_CAPTURE_TIME_URI, direction: 'stopped' }
    ],
    setHeaderExtensionsToNegotiate: (e: { uri: string; direction: string }[]) => {
      set = e;
    }
  };

  assert.equal(offerAbsCaptureTime(transceiver), true);
  assert.deepEqual(set, [
    { uri: 'urn:ietf:params:rtp-hdrext:toffset', direction: 'sendrecv' },
    { uri: ABS_CAPTURE_TIME_URI, direction: 'recvonly' }
  ]);
});

test('a browser without the API is left alone', () => {
  assert.equal(offerAbsCaptureTime({}), false);
});

test('a browser without the extension is left alone', () => {
  let called = false;
  const transceiver = {
    getHeaderExtensionsToNegotiate: () => [{ uri: 'x', direction: 'sendrecv' }],
    setHeaderExtensionsToNegotiate: () => {
      called = true;
    }
  };
  assert.equal(offerAbsCaptureTime(transceiver), false);
  assert.equal(called, false);
});

test('latency from a synchronization source', () => {
  const capturedUnixMs = 1760000000000;
  assert.equal(
    captureLatencyMs({
      timestamp: capturedUnixMs + 42,
      captureTimestamp: capturedUnixMs + NTP_EPOCH_OFFSET_MS,
      senderCaptureTimeOffset: 0
    }),
    42
  );
  assert.equal(captureLatencyMs({ timestamp: 1 }), undefined);
});
