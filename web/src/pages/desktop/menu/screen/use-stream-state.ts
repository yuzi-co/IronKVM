import { atom, useAtomValue } from 'jotai';

import { streamState, type StreamState } from '@/lib/health.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { captureStatusAtom, streamPausedAtom } from '@/jotai/screen.ts';

import { isMediaReady } from '../../screen/geometry.ts';

const PICTURE_CHECK_MS = 2000;

// isPictureShown says whether the video element has a picture to show, in
// whichever of the three modes it is: a video, an image or a canvas.
function isPictureShown() {
  const screen = document.getElementById('screen');
  return !!screen && isMediaReady(screen);
}

// pictureShownAtom is one DOM check for every reader (the Screen menu and the
// alert icon), made while anything reads it and the tab is visible. A timer
// rather than a MutationObserver, since readiness also follows the video's
// readyState and the image's load, which change no attribute.
const pictureShownAtom = atom(false);
pictureShownAtom.onMount = (set) => {
  const check = () => set(isPictureShown());
  check();
  return pollWhileVisible(check, PICTURE_CHECK_MS);
};

// useStreamState is the picture as the Screen dot and the alert icon show it.
// A failure comes from the capture status the board pushes; "ok" needs a
// picture on the page as well, since the board sends no status while all is
// well. The page check is a DOM read every two seconds.
export function useStreamState(): StreamState {
  const report = useAtomValue(captureStatusAtom);
  const paused = useAtomValue(streamPausedAtom);
  const pictureShown = useAtomValue(pictureShownAtom);

  if (paused) return 'unknown';
  return streamState(report, pictureShown);
}
