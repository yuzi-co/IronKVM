import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  applyRoomMicPush,
  audioEntries,
  clampGain,
  nextRoomMicSwitch,
  parseRoomMicStatus,
  roomMicAdminState,
  roomMicControlBytes,
  roomMicIndicatorTitle,
  roomMicPushFromBytes,
  roomMicPushFromSignal,
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
  listenerCount: 0,
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

test('the speaker entry is host audio only and follows Preferences', () => {
  const entries = (speakerEnabled: boolean, hasHostAudio: boolean) =>
    audioEntries({ speakerEnabled, hasHostAudio, status: slotB, transportOffersRoom: true });
  assert.equal(entries(true, true).speaker, true);
  assert.equal(entries(true, false).speaker, false);
  assert.equal(entries(false, true).speaker, false);
});

test('the microphone has its own entry, whatever Preferences does to the speaker', () => {
  for (const speakerEnabled of [true, false]) {
    for (const hasHostAudio of [true, false]) {
      const entries = audioEntries({
        speakerEnabled,
        hasHostAudio,
        status: slotB,
        transportOffersRoom: true
      });
      assert.equal(entries.roomMic, true, `speaker ${speakerEnabled}, host audio ${hasHostAudio}`);
    }
  }
});

test('the microphone entry needs the card, the setting and a transport with audio', () => {
  const roomMic = (status: RoomMicStatus, transportOffersRoom: boolean) =>
    audioEntries({ speakerEnabled: true, hasHostAudio: true, status, transportOffersRoom }).roomMic;
  assert.equal(roomMic(slotB, true), true);
  assert.equal(roomMic({ ...slotB, available: false }, true), false);
  assert.equal(roomMic({ ...slotB, allowed: false }, true), false);
  // MJPEG carries no audio, so it never offers the microphone.
  assert.equal(roomMic(slotB, false), false);
});

test('the indicator names the listeners to an administrator only', () => {
  const live = { ...slotB, live: true, listenerCount: 2, listeners: ['alice', 'bob'] };
  assert.deepEqual(roomMicIndicatorTitle(live, true), {
    key: 'speaker.roomLiveBy',
    names: 'alice, bob'
  });
  assert.deepEqual(roomMicIndicatorTitle(live, false), { key: 'speaker.roomLive' });
  // What the server sends a non-admin: the count, no names.
  const userView = { ...live, listeners: [] };
  assert.deepEqual(roomMicIndicatorTitle(userView, false), { key: 'speaker.roomLive' });
  assert.deepEqual(roomMicIndicatorTitle(userView, true), { key: 'speaker.roomLive' });
});

test('a status without names keeps the count', () => {
  const status = parseRoomMicStatus({
    available: true,
    allowed: true,
    gain: 18,
    live: true,
    listenerCount: 2
  });
  assert.equal(status.listenerCount, 2);
  assert.deepEqual(status.listeners, []);
});

test('a direct push keeps the polled listeners while live and clears them after', () => {
  const live = { ...slotB, live: true, listenerCount: 1, listeners: ['alice'] };
  const stillLive = applyRoomMicPush(
    live,
    roomMicPushFromBytes(new Uint8Array([0x13, 1, 0, 1, 0]))!
  );
  assert.equal(stillLive.listenerCount, 1);
  assert.deepEqual(stillLive.listeners, ['alice']);
  const off = applyRoomMicPush(live, roomMicPushFromBytes(new Uint8Array([0x13, 0, 0, 1, 0]))!);
  assert.equal(off.listenerCount, 0);
  assert.deepEqual(off.listeners, []);
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
      listenerCount: 1,
      listeners: ['a', 3]
    }),
    { available: true, allowed: true, gain: 12, live: true, listenerCount: 1, listeners: ['a'] }
  );
});

test('a WebRTC state reads its fields', () => {
  assert.deepEqual(
    roomMicPushFromSignal(
      '{"live":true,"listening":false,"listenerCount":1,"listeners":["bob"],"allowed":true}'
    ),
    { live: true, listening: false, allowed: true, listenerCount: 1, listeners: ['bob'], error: '' }
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
