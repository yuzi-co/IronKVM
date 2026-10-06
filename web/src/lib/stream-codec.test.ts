import assert from 'node:assert/strict';
import { test } from 'node:test';

import { codecChangeNeedsNewSession, codecName, videoModeLabel } from './stream-codec.ts';

test('each codec number has its name, and an unknown one reads as H.264', () => {
  assert.equal(codecName(1), 'H.264');
  assert.equal(codecName(2), 'H.265');
  assert.equal(codecName(9), 'H.264');
  assert.equal(codecName(null), 'H.264');
  assert.equal(codecName(undefined), 'H.264');
});

test('the two H.264 paths are labelled with the codec the encoder runs', () => {
  assert.equal(videoModeLabel('direct', 1), 'H.264 (Direct)');
  assert.equal(videoModeLabel('h264', 1), 'H.264 (WebRTC)');
  assert.equal(videoModeLabel('direct', 2), 'H.265 (Direct)');
  assert.equal(videoModeLabel('h264', 2), 'H.265 (WebRTC)');
  assert.equal(videoModeLabel('mjpeg', 2), 'MJPEG');
});

test('a WebRTC session is renegotiated when the codec changes, either way', () => {
  assert.equal(codecChangeNeedsNewSession(1, 2), true);
  assert.equal(codecChangeNeedsNewSession(2, 1), true);
});

test('the first read of the setting and an unchanged codec keep the session', () => {
  assert.equal(codecChangeNeedsNewSession(null, 2), false);
  assert.equal(codecChangeNeedsNewSession(undefined, 1), false);
  assert.equal(codecChangeNeedsNewSession(1, 1), false);
  assert.equal(codecChangeNeedsNewSession(2, null), false);
});
