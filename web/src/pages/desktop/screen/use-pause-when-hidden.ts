import { useEffect, useState } from 'react';
import { useAtomValue, useSetAtom } from 'jotai';

import { pauseWhenHiddenAtom, streamPausedAtom } from '@/jotai/screen.ts';

// A tab hidden this long has its stream stopped. A quick look at another tab
// does not pay for a reconnect, which takes a second or more on WebRTC.
const PAUSE_AFTER_MS = 5000;

// usePauseWhenHidden says whether the stream should be stopped now: the
// viewer asked for it and the tab has been hidden for a while. The desktop
// unmounts the video while it is true, which closes the stream the same way
// for all three video modes, and mounts it again when the tab comes back.
export function usePauseWhenHidden(): boolean {
  const enabled = useAtomValue(pauseWhenHiddenAtom);
  const setStreamPaused = useSetAtom(streamPausedAtom);
  const [hiddenLong, setHiddenLong] = useState(false);

  useEffect(() => {
    if (!enabled) return;

    let timer: ReturnType<typeof setTimeout> | null = null;
    const clear = () => {
      if (timer) clearTimeout(timer);
      timer = null;
    };

    function update() {
      clear();
      if (document.visibilityState === 'hidden') {
        timer = setTimeout(() => setHiddenLong(true), PAUSE_AFTER_MS);
      } else {
        setHiddenLong(false);
      }
    }

    if (document.visibilityState === 'hidden') update();
    document.addEventListener('visibilitychange', update);
    return () => {
      clear();
      document.removeEventListener('visibilitychange', update);
      setHiddenLong(false);
    };
  }, [enabled]);

  const paused = enabled && hiddenLong;

  useEffect(() => {
    setStreamPaused(paused);
  }, [paused, setStreamPaused]);

  return paused;
}
