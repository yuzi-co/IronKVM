import { atom } from 'jotai';

import { getRoomMic } from '@/api/room-mic.ts';
import {
  applyRoomMicPush,
  parseRoomMicStatus,
  switchRoomMic,
  type RoomMicPush,
  type RoomMicStatus,
  type RoomMicSwitch
} from '@/lib/room-mic.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

// The poll is what tells a viewer on MJPEG, which carries no sound, that the
// microphone is live. WebRTC and H.264 direct push each change at once.
const ROOM_MIC_POLL_MS = 5000;

const baseRoomMicStatusAtom = atom<RoomMicStatus | null>(null);

baseRoomMicStatusAtom.onMount = (set) => {
  let active = true;

  function read() {
    getRoomMic()
      .then((rsp) => {
        if (active && rsp.code === 0) set(parseRoomMicStatus(rsp.data));
      })
      // A failed poll keeps the last state; the next one tries again.
      .catch(() => {});
  }

  read();
  const stop = pollWhileVisible(read, ROOM_MIC_POLL_MS);
  return () => {
    active = false;
    stop();
  };
};

// roomMicStatusAtom is the device's microphone state, null until the first
// answer. Writing a status replaces it (the admin setting does, with the
// server's answer).
export const roomMicStatusAtom = atom(
  (get) => get(baseRoomMicStatusAtom),
  (_get, set, status: RoomMicStatus) => set(baseRoomMicStatusAtom, status)
);

// roomMicPushAtom folds a state pushed by the video connection into the
// status.
export const roomMicPushAtom = atom(null, (get, set, push: RoomMicPush) => {
  const current = get(baseRoomMicStatusAtom);
  if (!current) return;
  set(baseRoomMicStatusAtom, applyRoomMicPush(current, push));
});

// roomMicSwitchAtom is this viewer's switch. It starts off on every page
// load: the microphone is never switched on for a viewer who did not ask.
export const roomMicSwitchAtom = atom<RoomMicSwitch>(switchRoomMic(false));

// roomMicListeningAtom is the server's word that it sends this viewer the
// microphone.
export const roomMicListeningAtom = atom(false);

// roomMicTransportAtom is true while the video connection in use can carry
// the microphone: WebRTC negotiated its track, or H.264 direct sent its state.
export const roomMicTransportAtom = atom(false);

// The microphone's own mute and volume, apart from the host's audio.
export const roomMicMutedAtom = atom(false);
export const roomMicVolumeAtom = atom(1);
