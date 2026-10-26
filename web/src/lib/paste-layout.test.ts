import assert from 'node:assert/strict';
import { test } from 'node:test';

import { guessPasteLayout, initialPasteLayout, pasteLayoutIds } from './paste-layout.ts';

test('the browser language is the guess, US when no layout matches', () => {
  assert.equal(guessPasteLayout('de-DE'), 'de');
  assert.equal(guessPasteLayout('en-GB'), 'uk');
  assert.equal(guessPasteLayout('fi-FI'), 'se');
  assert.equal(guessPasteLayout('cs'), 'cz');
  assert.equal(guessPasteLayout('xx'), 'us');
});

test('a layout this browser picked wins', () => {
  assert.equal(initialPasteLayout('fr', { layout: 'de', configured: true }, 'it'), 'fr');
});

test('without one, the layout the owner set on the device is used', () => {
  assert.equal(initialPasteLayout(null, { layout: 'de', configured: true }, 'it'), 'de');
  assert.equal(initialPasteLayout('nonsense', { layout: 'us', configured: true }, 'it'), 'us');
});

test('without either, the browser language is the guess', () => {
  assert.equal(initialPasteLayout(null, null, 'it'), 'it');
  assert.equal(initialPasteLayout(null, { layout: 'us', configured: false }, 'it'), 'it');
  assert.equal(initialPasteLayout(null, { layout: 'xx', configured: true }, 'it'), 'it');
});

test('the layouts match the server', () => {
  assert.equal(pasteLayoutIds.length, 37);
  assert.equal(pasteLayoutIds[0], 'us');
});
