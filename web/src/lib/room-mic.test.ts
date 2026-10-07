import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  applyRoomMicPush,
  clampGain,
  nextRoomMicSwitch,
  parseRoomMicStatus,
  roomMicAdminState,
  roomMicControlBytes,
  roomMicPushFromBytes,
  roomMicPushFromSignal,
  showAudioMenu,
  showRoomMicControls,
  showRoomMicIndicator,
  switchRoomMic,
  unknownRoomMicStatus,
  type RoomMicPush,
  type RoomMicStatus
} from './room-mic.ts';

const slotB: RoomMicStatus = {
  available: true,
  allowed: true,
  gain: 18,
  live: false,
  listeners: []
};

test('slot A, without the card, shows nothing to viewers', () => {
  const slotA = { ...slotB, available: false };
  assert.equal(showRoomMicControls(slotA, true), false);
  assert.equal(showRoomMicIndicator({ ...slotA, live: true }), false);
  assert.equal(roomMicAdminState(slotA), 'unavailable');
});

test('with the setting off the menu has no microphone', () => {
  assert.equal(showRoomMicControls({ ...slotB, allowed: false }, true), false);
});

test('allowed, the menu offers it only over a connection that carries it', () => {
  assert.equal(showRoomMicControls(slotB, true), true);
  assert.equal(showRoomMicControls(slotB, false), false);
});

test('the indicator shows to every viewer while it is live, and only then', () => {
  assert.equal(showRoomMicIndicator(slotB), false);
  assert.equal(showRoomMicIndicator({ ...slotB, live: true }), true);
  // A viewer who is not listening, or not even allowed to, still sees it.
  assert.equal(showRoomMicIndicator({ ...slotB, allowed: false, live: true }), true);
});

test('the speaker entry shows for host audio or the microphone', () => {
  assert.equal(showAudioMenu(false, false), false);
  assert.equal(showAudioMenu(true, false), true);
  assert.equal(showAudioMenu(false, true), true);
});

test('the admin setting waits for the status, then knows the kernel', () => {
  assert.equal(roomMicAdminState(null), 'loading');
  assert.equal(roomMicAdminState(slotB), 'ready');
});

test('a status the page cannot read hides the feature', () => {
  assert.deepEqual(parseRoomMicStatus(undefined), unknownRoomMicStatus);
  assert.deepEqual(parseRoomMicStatus('nope'), unknownRoomMicStatus);
  assert.deepEqual(
    parseRoomMicStatus({
      available: true,
      allowed: true,
      gain: 12,
      live: true,
      listeners: ['a', 3]
    }),
    { available: true, allowed: true, gain: 12, live: true, listeners: ['a'] }
  );
});

test('a WebRTC state reads its fields', () => {
  assert.deepEqual(
    roomMicPushFromSignal('{"live":true,"listening":false,"listeners":["bob"],"allowed":true}'),
    { live: true, listening: false, allowed: true, listeners: ['bob'], error: '' }
  );
  assert.equal(roomMicPushFromSignal('{'), null);
  assert.equal(roomMicPushFromSignal(undefined), null);
  assert.equal(roomMicPushFromSignal('{"error":"not-allowed"}')?.error, 'not-allowed');
  assert.equal(roomMicPushFromSignal('{"error":"other"}')?.error, '');
});

test('a direct state message is five bytes', () => {
  assert.deepEqual(roomMicPushFromBytes(new Uint8Array([0x13, 1, 0, 1, 0])), {
    live: true,
    listening: false,
    allowed: true,
    error: ''
  });
  assert.equal(roomMicPushFromBytes(new Uint8Array([0x13, 0, 0, 0, 1]))?.error, 'not-allowed');
  assert.equal(roomMicPushFromBytes(new Uint8Array([0x13, 0, 0, 0, 2]))?.error, 'unavailable');
  assert.equal(roomMicPushFromBytes(new Uint8Array([0x13, 1, 1])), null);
  assert.equal(roomMicPushFromBytes(new Uint8Array([0x11, 1, 1, 1, 0])), null);
});

test('the direct switch is the marker and one byte', () => {
  assert.deepEqual([...roomMicControlBytes(true)], [4, 1]);
  assert.deepEqual([...roomMicControlBytes(false)], [4, 0]);
});

test('a push turns the indicator on for everyone at once', () => {
  const push: RoomMicPush = {
    live: true,
    listening: false,
    allowed: true,
    listeners: ['alice'],
    error: ''
  };
  const next = applyRoomMicPush({ ...slotB, available: false }, push);
  assert.equal(showRoomMicIndicator(next), true);
  assert.deepEqual(next.listeners, ['alice']);
});

const push = (p: Partial<RoomMicPush>): RoomMicPush => ({
  live: false,
  listening: false,
  allowed: true,
  error: '',
  ...p
});

test('switching on waits out a stale "not listening"', () => {
  let s = switchRoomMic(true);
  s = nextRoomMicSwitch(s, push({}));
  assert.deepEqual(s, { wanted: true, awaiting: true });
  s = nextRoomMicSwitch(s, push({ live: true, listening: true }));
  assert.deepEqual(s, { wanted: true, awaiting: false });
});

test('a refusal turns the switch off', () => {
  const s = nextRoomMicSwitch(switchRoomMic(true), push({ error: 'not-allowed' }));
  assert.deepEqual(s, { wanted: false, awaiting: false });
});

test('the server ending the listening turns the switch off', () => {
  let s = nextRoomMicSwitch(switchRoomMic(true), push({ live: true, listening: true }));
  s = nextRoomMicSwitch(s, push({ allowed: false }));
  assert.deepEqual(s, { wanted: false, awaiting: false });
});

test('someone else opening it does not switch this viewer on', () => {
  const s = nextRoomMicSwitch(switchRoomMic(false), push({ live: true }));
  assert.deepEqual(s, { wanted: false, awaiting: false });
});

test('the gain stays in the codec range', () => {
  assert.equal(clampGain(30), 24);
  assert.equal(clampGain(-2), 0);
  assert.equal(clampGain(17.6), 18);
  assert.equal(clampGain(Number.NaN), 18);
});
