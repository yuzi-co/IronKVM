// The encoder's codec as the browser sees it. libkvm numbers them 1 for H.264
// and 2 for H.265; the server reports that number in the screen settings.
const codecNames: Record<number, string> = { 1: 'H.264', 2: 'H.265' };

// codecName names a codec number, falling back to H.264 for anything the
// browser does not know, which is also what the server falls back to.
export function codecName(codec: number | null | undefined): string {
  return (codec !== null && codec !== undefined && codecNames[codec]) || codecNames[1];
}

// videoModeLabel is the Video Mode menu's label for a delivery path. Both
// H.264 paths carry whichever codec the encoder runs, so the label names that
// codec rather than always saying H.264 (#84).
export function videoModeLabel(mode: string, codec: number | null | undefined): string {
  switch (mode) {
    case 'direct':
      return `${codecName(codec)} (Direct)`;
    case 'h264':
      return `${codecName(codec)} (WebRTC)`;
    case 'mjpeg':
      return 'MJPEG';
    default:
      return mode;
  }
}

// codecChangeNeedsNewSession says whether a change of the codec setting makes
// the open WebRTC session useless. The track's codec is fixed when the session
// negotiates, so a session opened at one codec gets no frames at another, and
// the picture freezes until a new session is negotiated (#83). A first read
// of the setting (from nothing) is not a change.
export function codecChangeNeedsNewSession(
  previous: number | null | undefined,
  next: number | null | undefined
): boolean {
  if (previous === null || previous === undefined) return false;
  if (next === null || next === undefined) return false;
  return previous !== next;
}

// The signalling event the server sends when the encoder runs another codec
// than the session negotiated.
export const CODEC_CHANGED_EVENT = 'codec-changed';
