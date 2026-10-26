import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  formatAudio,
  formatMode,
  formatRate,
  formatTiming,
  hexDump,
  presetDownloadPath,
  timingDetail
} from './edid.ts';
import type { Timing } from './edid.ts';

const cea1080p60: Timing = {
  block: 0,
  pixelClockKHz: 148500,
  hActive: 1920,
  hBlank: 280,
  hFrontPorch: 88,
  hSync: 44,
  vActive: 1080,
  vBlank: 45,
  vFrontPorch: 4,
  vSync: 5,
  widthMm: 476,
  heightMm: 268,
  interlaced: false,
  hSyncPositive: true,
  vSyncPositive: true,
  refreshHz: 60,
  hFreqKHz: 67.5
};

test('a rate keeps the decimals it has, up to two', () => {
  assert.equal(formatRate(60), '60');
  assert.equal(formatRate(59.94), '59.94');
  assert.equal(formatRate(99.998), '100');
  assert.equal(formatRate(59.951), '59.95');
});

test('a timing reads as its mode, with frame lines for an interlaced one', () => {
  assert.equal(formatTiming(cea1080p60), '1920x1080 @ 60 Hz');
  assert.equal(
    formatTiming({ ...cea1080p60, vActive: 540, vBlank: 22, interlaced: true, refreshHz: 50 }),
    '1920x1080i @ 50 Hz'
  );
});

test('the detail line has the clocks, porches and sync polarity', () => {
  assert.equal(timingDetail(cea1080p60), '148.5 MHz · 67.5 kHz · H 88/44/148 · V 4/5/36 · +H +V');
  assert.equal(
    timingDetail({ ...cea1080p60, hSyncPositive: true, vSyncPositive: false }).endsWith('+H −V'),
    true
  );
});

test('a mode from a list', () => {
  assert.equal(formatMode({ width: 640, height: 480, refreshHz: 60 }), '640x480 @ 60 Hz');
  assert.equal(
    formatMode({ width: 1024, height: 768, refreshHz: 87, interlaced: true }),
    '1024x768i @ 87 Hz'
  );
});

test('an audio descriptor in one line', () => {
  assert.equal(
    formatAudio({
      format: 'LPCM',
      channels: 2,
      sampleRatesKHz: [32, 44.1, 48],
      bitDepths: [16, 20]
    }),
    'LPCM, 2 ch, 32/44.1/48 kHz, 16/20 bit'
  );
  assert.equal(
    formatAudio({ format: 'AC-3', channels: 6, sampleRatesKHz: [48], maxBitrateKbps: 640 }),
    'AC-3, 6 ch, 48 kHz, 640 kbit/s'
  );
});

test('a hex dump has an offset and sixteen bytes a line', () => {
  const hex = '00ffffffffffff00' + '4e04013315000000' + '1e23';
  assert.deepEqual(hexDump(hex), [
    '0000  00 ff ff ff ff ff ff 00 4e 04 01 33 15 00 00 00',
    '0010  1e 23'
  ]);
  assert.deepEqual(hexDump(''), []);
});

test('a preset download path escapes the id', () => {
  assert.equal(presetDownloadPath('1080p60'), '/api/edid/presets/1080p60/bin');
  assert.equal(presetDownloadPath('a/b'), '/api/edid/presets/a%2Fb/bin');
});
