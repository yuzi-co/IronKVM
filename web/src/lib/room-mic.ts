// The room microphone: the board's own microphone, which hears the room the
// KVM sits in. Only the mainline kernel has it (the ALSA card sg2002onboard).
//
// An administrator allows it for the device, and then each viewer switches it
// on for themselves. Every viewer is shown when it is live, whoever switched
// it on. These functions hold the rules; the components only draw them.

export type RoomMicStatus = {
  // available is false on a kernel without the onboard card.
  available: boolean;
  // allowed is the administrator's setting.
  allowed: boolean;
  // gain is the capture gain, 0 to 24 in 2 dB steps.
  gain: number;
  // live is true while the microphone is open for anyone.
  live: boolean;
  // listenerCount is how many accounts have it on.
  listenerCount: number;
  // listeners names who has it on. The server tells administrators only;
  // for everyone else it is empty.
  listeners: string[];
};

export const ROOM_MIC_MIN_GAIN = 0;
export const ROOM_MIC_MAX_GAIN = 24;
export const ROOM_MIC_DEFAULT_GAIN = 18;

export const unknownRoomMicStatus: RoomMicStatus = {
  available: false,
  allowed: false,
  gain: ROOM_MIC_DEFAULT_GAIN,
  live: false,
  listenerCount: 0,
  listeners: []
};

function asStrings(value: unknown): string[] {
  return Array.isArray(value) ? value.filter((v): v is string => typeof v === 'string') : [];
}

// parseRoomMicStatus reads GET /api/room-mic. Anything unexpected reads as
// not available, which hides the feature.
export function parseRoomMicStatus(data: unknown): RoomMicStatus {
  if (!data || typeof data !== 'object') return unknownRoomMicStatus;
  const d = data as Record<string, unknown>;

  return {
    available: d.available === true,
    allowed: d.allowed === true,
    gain: typeof d.gain === 'number' ? d.gain : ROOM_MIC_DEFAULT_GAIN,
    live: d.live === true,
    listenerCount: asCount(d.listenerCount),
    listeners: asStrings(d.listeners)
  };
}

function asCount(value: unknown): number {
  return typeof value === 'number' && Number.isFinite(value) && value > 0 ? Math.floor(value) : 0;
}

// What the video connection says about the microphone, for this viewer.
export type RoomMicPush = {
  live: boolean;
  listening: boolean;
  allowed: boolean;
  // WebRTC says who listens (to an administrator) and how many; H.264
  // direct says neither.
  listenerCount?: number;
  listeners?: string[];
  error: '' | 'not-allowed' | 'unavailable';
};

function asError(value: unknown): RoomMicPush['error'] {
  return value === 'not-allowed' || value === 'unavailable' ? value : '';
}

// roomMicPushFromSignal reads WebRTC's 'room-mic' signalling event.
export function roomMicPushFromSignal(data: string | undefined): RoomMicPush | null {
  if (!data) return null;
  try {
    const d = JSON.parse(data) as Record<string, unknown>;
    return {
      live: d.live === true,
      listening: d.listening === true,
      allowed: d.allowed === true,
      listenerCount: asCount(d.listenerCount),
      listeners: asStrings(d.listeners),
      error: asError(d.error)
    };
  } catch {
    return null;
  }
}

// The H.264 direct websocket's messages for the microphone.
export const ROOM_MIC_CONTROL = 0x04;
export const ROOM_AUDIO_MESSAGE = 0x12;
export const ROOM_STATE_MESSAGE = 0x13;
const ROOM_STATE_SIZE = 5;

// roomMicPushFromBytes reads a direct state message: the marker, live,
// listening, allowed, and why the last switch-on failed.
export function roomMicPushFromBytes(bytes: Uint8Array): RoomMicPush | null {
  if (bytes.length !== ROOM_STATE_SIZE || bytes[0] !== ROOM_STATE_MESSAGE) return null;

  const errors: RoomMicPush['error'][] = ['', 'not-allowed', 'unavailable'];
  return {
    live: bytes[1] === 1,
    listening: bytes[2] === 1,
    allowed: bytes[3] === 1,
    error: errors[bytes[4]] ?? ''
  };
}

// roomMicControlBytes is the switch a direct viewer sends.
export function roomMicControlBytes(on: boolean): Uint8Array<ArrayBuffer> {
  return new Uint8Array([ROOM_MIC_CONTROL, on ? 1 : 0]);
}

// applyRoomMicPush folds a pushed state into the polled one, so the indicator
// changes as soon as the connection says so rather than at the next poll. A
// push only comes on a kernel with the card.
export function applyRoomMicPush(status: RoomMicStatus, push: RoomMicPush): RoomMicStatus {
  return {
    ...status,
    available: true,
    allowed: push.allowed,
    live: push.live,
    // A push without the count or the names (H.264 direct) keeps the polled
    // ones while the microphone stays live, and clears them once it is not.
    listenerCount: push.listenerCount ?? (push.live ? status.listenerCount : 0),
    listeners: push.listeners ?? (push.live ? status.listeners : [])
  };
}

// The viewer's switch and whether an answer to it is still on its way.
export type RoomMicSwitch = { wanted: boolean; awaiting: boolean };

// switchRoomMic records the viewer flipping the switch. Switching on waits
// for the server to say it opened.
export function switchRoomMic(on: boolean): RoomMicSwitch {
  return { wanted: on, awaiting: on };
}

// nextRoomMicSwitch follows the server. A refusal turns the switch off. So
// does the server ending this viewer's listening by itself (an administrator
// disallowed the microphone, or it failed). A state sent before the server saw
// the switch says "not listening" too, so while an answer is awaited that is
// not taken as an end.
export function nextRoomMicSwitch(current: RoomMicSwitch, push: RoomMicPush): RoomMicSwitch {
  if (push.error) return { wanted: false, awaiting: false };
  if (push.listening) return { wanted: current.wanted, awaiting: false };
  if (current.wanted && !current.awaiting) return { wanted: false, awaiting: false };
  return current;
}

// showRoomMicControls decides whether the microphone's own toolbar entry
// shows: the kernel has it, an administrator allowed it, and the video
// connection in use can carry it (MJPEG cannot).
export function showRoomMicControls(status: RoomMicStatus, transportOffersRoom: boolean): boolean {
  return status.available && status.allowed && transportOffersRoom;
}

// showRoomMicIndicator decides whether the red indicator shows. It shows to
// every viewer while the microphone is open, whether or not they listen.
export function showRoomMicIndicator(status: RoomMicStatus): boolean {
  return status.available && status.live;
}

// The toolbar's two audio entries: the speaker for the host's audio and the
// room microphone's own.
export type AudioEntries = { speaker: boolean; roomMic: boolean };

// audioEntries decides which audio entries are on the bar. The speaker needs
// host audio and can be hidden in Preferences. The microphone's entry follows
// showRoomMicControls alone: hiding the speaker does not hide it.
export function audioEntries(options: {
  speakerEnabled: boolean;
  hasHostAudio: boolean;
  status: RoomMicStatus;
  transportOffersRoom: boolean;
}): AudioEntries {
  return {
    speaker: options.speakerEnabled && options.hasHostAudio,
    roomMic: showRoomMicControls(options.status, options.transportOffersRoom)
  };
}

// The indicator's tooltip, as a translation key and its values.
export type RoomMicIndicatorTitle =
  { key: 'speaker.roomLiveBy'; names: string } | { key: 'speaker.roomLive' };

// roomMicIndicatorTitle says who listens to an administrator and just "Mic
// live" to everyone else. The server sends the names to administrators only;
// the role check keeps the rule here too.
export function roomMicIndicatorTitle(
  status: RoomMicStatus,
  isAdmin: boolean
): RoomMicIndicatorTitle {
  const names = status.listeners.join(', ');
  if (isAdmin && names) return { key: 'speaker.roomLiveBy', names };
  return { key: 'speaker.roomLive' };
}

// roomMicAdminState is what the administrator's setting shows: the switch, or
// a line saying this kernel has no microphone.
export function roomMicAdminState(
  status: RoomMicStatus | null
): 'loading' | 'unavailable' | 'ready' {
  if (!status) return 'loading';
  return status.available ? 'ready' : 'unavailable';
}

// clampGain keeps a slider value inside the codec's range.
export function clampGain(gain: number): number {
  if (!Number.isFinite(gain)) return ROOM_MIC_DEFAULT_GAIN;
  return Math.min(ROOM_MIC_MAX_GAIN, Math.max(ROOM_MIC_MIN_GAIN, Math.round(gain)));
}
