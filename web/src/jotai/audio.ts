import { atom } from 'jotai';

// Audio starts muted. Browsers refuse to autoplay sound before the user acts,
// so the first unmute has to be a click.
export const audioMutedAtom = atom(true);

// hasAudio follows the sound itself, which is the only availability signal
// there is. The server sends Opus only when the USB audio gadget has a capture
// card: over WebRTC as an audio track, and over H.264 direct as audio messages
// on the video websocket. Either arriving means the operator can hear the host,
// and neither means the feature is off on this device.
//
// The speaker button is gated on this. Without it the button renders on every
// device, and on the great majority - which have no /boot/usb.uac marker -
// clicking unmute sets muted = false on an <audio> with no srcObject and
// produces silence with nothing to explain it.
export const hasAudioAtom = atom(false);
