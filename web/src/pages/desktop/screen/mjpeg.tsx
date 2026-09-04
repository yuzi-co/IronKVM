import { useEffect, useRef, useState } from 'react';
import clsx from 'clsx';
import { useAtomValue } from 'jotai';

import { stopFrameDetect } from '@/api/stream.ts';
import { getFrameDetect } from '@/lib/localstorage.ts';
import { getBaseUrl } from '@/lib/service.ts';
import { mouseStyleAtom } from '@/jotai/mouse.ts';
import { resolutionAtom } from '@/jotai/screen.ts';

import { ScreenViewport } from './viewport.tsx';

// How long to wait before asking for the stream again, and the ceiling that
// backing off reaches. A device that is genuinely gone is retried every five
// seconds rather than continuously.
const retryDelayMs = 1000;
const maxRetryDelayMs = 5000;

export const Mjpeg = () => {
  const resolution = useAtomValue(resolutionAtom);
  const mouseStyle = useAtomValue(mouseStyleAtom);
  const [hasError, setHasError] = useState(false);
  const [streamNonce, setStreamNonce] = useState(0);
  const retryTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const retryDelay = useRef(retryDelayMs);

  const streamURL = `${getBaseUrl('http')}/api/stream/mjpeg`;
  const streamSrc = hasError ? undefined : `${streamURL}?v=${streamNonce}`;

  function clearRetry() {
    if (retryTimer.current) {
      clearTimeout(retryTimer.current);
      retryTimer.current = null;
    }
  }

  // The response ends whenever the server stops having frames for long enough,
  // and losing the HDMI signal for half a minute is enough. This used to set
  // hasError and stop there, and the only thing that ever cleared it again was
  // a resolution change: the signal coming back did nothing, because nothing
  // asked for the stream a second time. The view stayed blank until the mode
  // was switched or the page was reloaded, both of which remount this. WebRTC
  // has always reconnected on its own.
  function scheduleRetry() {
    if (retryTimer.current) {
      return;
    }

    retryTimer.current = setTimeout(() => {
      retryTimer.current = null;
      retryDelay.current = Math.min(retryDelay.current * 2, maxRetryDelayMs);
      setHasError(false);
      setStreamNonce((current) => current + 1);
    }, retryDelay.current);
  }

  useEffect(() => clearRetry, []);

  useEffect(() => {
    // stop frame detect for a while
    const enabled = getFrameDetect();
    if (enabled) {
      stopFrameDetect(10);
    }

    clearRetry();
    retryDelay.current = retryDelayMs;
    setHasError(false);
    setStreamNonce((current) => current + 1);
  }, [resolution]);

  return (
    <ScreenViewport>
      <img
        id="screen"
        className={clsx('block touch-none select-none', mouseStyle)}
        style={{
          visibility: hasError ? 'hidden' : 'visible'
        }}
        src={streamSrc}
        // A multipart stream reports a load per part, so this is the signal
        // that the stream is serving again. It writes a ref rather than state,
        // so a frame arriving does not re-render anything.
        onLoad={() => {
          retryDelay.current = retryDelayMs;
        }}
        onError={() => {
          setHasError(true);
          scheduleRetry();
        }}
        alt="screen"
      />
    </ScreenViewport>
  );
};
