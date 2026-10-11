import { useEffect, useRef, useState } from 'react';
import clsx from 'clsx';
import { useAtomValue, useSetAtom } from 'jotai';

import { stopFrameDetect } from '@/api/stream.ts';
import { getBaseUrl } from '@/lib/service.ts';
import { mouseStyleAtom } from '@/jotai/mouse.ts';
import { resolutionAtom, streamInfoVisibleAtom, streamStatsAtom } from '@/jotai/screen.ts';

import { emptyStats } from '../status-bar/model.ts';
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
  const statsVisible = useAtomValue(streamInfoVisibleAtom);
  const setStreamStats = useSetAtom(streamStatsAtom);
  const imageRef = useRef<HTMLImageElement | null>(null);
  // Whether the stream has shown a picture since it was last asked for.
  const hasLoaded = useRef(false);
  const hasErrorRef = useRef(hasError);
  useEffect(() => {
    hasErrorRef.current = hasError;
  }, [hasError]);

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

  // The browser tells a page almost nothing of a multipart image: Chrome
  // fires load for the first part only, so there is no frame count, no
  // bitrate, and no capture time to measure latency against. The bar shows
  // the picture's size and the board's own capture rate.
  useEffect(() => {
    if (!statsVisible) return;
    const timer = setInterval(() => {
      const image = imageRef.current;
      const stats = emptyStats(
        'mjpeg',
        hasErrorRef.current ? 'reconnecting' : hasLoaded.current ? 'connected' : 'connecting'
      );
      stats.codec = 'MJPEG';
      stats.width = image?.naturalWidth ?? 0;
      stats.height = image?.naturalHeight ?? 0;
      stats.transport = window.location.protocol === 'https:' ? 'HTTPS' : 'HTTP';
      setStreamStats(stats);
    }, 1000);
    return () => {
      clearInterval(timer);
      setStreamStats(null);
    };
  }, [statsVisible, setStreamStats]);

  // A new resolution asks for the stream again from a clean state. The state
  // half is done during render, so no frame is painted with the old error flag
  // against the new resolution. The first render needs neither step: the
  // stream it builds is already the first request.
  const [prevResolution, setPrevResolution] = useState(resolution);
  if (resolution !== prevResolution) {
    setPrevResolution(resolution);
    setHasError(false);
    setStreamNonce((current) => current + 1);
  }

  useEffect(() => {
    // Pause frame detection for a while so the new stream gets a first frame.
    // The board ignores this when detection is off.
    stopFrameDetect(10).catch(() => {});

    clearRetry();
    retryDelay.current = retryDelayMs;
  }, [resolution]);

  return (
    <ScreenViewport>
      <img
        ref={imageRef}
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
          hasLoaded.current = true;
        }}
        onError={() => {
          hasLoaded.current = false;
          setHasError(true);
          scheduleRetry();
        }}
        alt="screen"
      />
    </ScreenViewport>
  );
};
