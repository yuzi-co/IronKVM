// What the board's audio capture is doing, as the server reports it.
//
// 'idle' means the host plays nothing to the KVM's USB sound device, so there
// is nothing to hear; it turns into 'playing' by itself when the host plays.
// 'unknown' is before the server has said anything.
export type AudioState = 'unknown' | 'playing' | 'idle' | 'failing';

// The server's audio.State values in order. H.264 direct sends the number as
// the second byte of a state message.
const states: readonly AudioState[] = ['unknown', 'playing', 'idle', 'failing'];

export function audioStateFromByte(value: number): AudioState {
  return states[value] ?? 'unknown';
}

// WebRTC sends the name in an 'audio-state' signalling event.
export function audioStateFromName(name: unknown): AudioState {
  return states.find((state) => state === name) ?? 'unknown';
}
