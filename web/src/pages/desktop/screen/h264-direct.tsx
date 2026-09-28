import { useEffect, useRef, useState } from 'react';
import { notification, Spin } from 'antd';
import clsx from 'clsx';
import { useAtomValue, useSetAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import { getBaseUrl } from '@/lib/service.ts';
import { audioMutedAtom, hasAudioAtom } from '@/jotai/audio.ts';
import { mouseStyleAtom } from '@/jotai/mouse';

import { DirectAudioPlayer } from './direct-audio.ts';
import DirectWorker from './direct.worker.ts?worker';
import { ScreenViewport } from './viewport.tsx';

const CONNECTION_FAILED_NOTIFICATION_KEY = 'direct_connection_failed';
// How long the stream may show no picture before the page says so, the same
// grace the WebRTC view gives.
const CONNECTION_GRACE_MS = 5 * 1000;

type WorkerEvent = {
  type?: string;
  state?: 'open' | 'closed';
  width?: number;
  height?: number;
  seq?: number;
  data?: ArrayBuffer;
};

export const H264Direct = () => {
  const { t } = useTranslation();
  const mouseStyle = useAtomValue(mouseStyleAtom);
  const isMuted = useAtomValue(audioMutedAtom);
  const setHasAudio = useSetAtom(hasAudioAtom);

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const workerRef = useRef<Worker | null>(null);
  const playerRef = useRef<DirectAudioPlayer | null>(null);
  const isMutedRef = useRef(isMuted);

  // Without WebCodecs nothing will ever draw, so there is nothing to wait for.
  const [isLoading, setIsLoading] = useState(() => !!window.VideoDecoder);
  const [notificationApi, contextHolder] = notification.useNotification();
  // The effect below must not depend on t: a language change would restart
  // the stream.
  const translationRef = useRef(t);
  useEffect(() => {
    translationRef.current = t;
  }, [t]);

  useEffect(() => {
    if (!window.VideoDecoder) {
      console.log('Error: WebCodecs API not supported.');
      return;
    }
    if (!canvasRef.current) {
      return;
    }

    const worker = new DirectWorker();
    workerRef.current = worker;

    const player = DirectAudioPlayer.supported() ? new DirectAudioPlayer() : null;
    playerRef.current = player;
    player?.setMuted(isMutedRef.current);
    let reportedAudio = false;

    // The worker reconnects by itself. The page shows a spinner until a frame
    // is drawn, and a notice when none has been for the grace period, which
    // is what the WebRTC view does too.
    let failTimer: ReturnType<typeof setTimeout> | null = null;
    const watchForFailure = () => {
      if (failTimer) return;
      failTimer = setTimeout(() => {
        failTimer = null;
        setIsLoading(false);
        notificationApi.error({
          key: CONNECTION_FAILED_NOTIFICATION_KEY,
          message: translationRef.current('screen.directConnectionFailed'),
          description: translationRef.current('screen.webrtcConnectionFailed.description'),
          placement: 'topRight',
          closable: false,
          duration: false
        });
      }, CONNECTION_GRACE_MS);
    };
    watchForFailure();

    const offscreen = canvasRef.current.transferControlToOffscreen();
    const url = `${getBaseUrl('ws')}/api/stream/h264/direct`;
    worker.onmessage = (event: MessageEvent<WorkerEvent>) => {
      const { type, state, width, height, seq, data } = event.data;

      if (type === 'playing') {
        if (failTimer) {
          clearTimeout(failTimer);
          failTimer = null;
        }
        setIsLoading(false);
        notificationApi.destroy(CONNECTION_FAILED_NOTIFICATION_KEY);
        return;
      }

      if (type === 'socket') {
        if (state === 'closed') {
          setIsLoading(true);
          watchForFailure();
        }
        return;
      }

      // The server sends audio only when the board has the USB audio gadget,
      // so the first frame is the signal that there is sound to unmute, the
      // same signal the WebRTC track gives.
      if (type === 'audio' && player && seq !== undefined && data) {
        if (!reportedAudio) {
          reportedAudio = true;
          setHasAudio(true);
        }
        player.push(seq, data);
        return;
      }

      if (type !== 'frame-size' || !width || !height || !canvasRef.current) {
        return;
      }

      canvasRef.current.dataset.mediaWidth = String(width);
      canvasRef.current.dataset.mediaHeight = String(height);
    };
    worker.postMessage({ type: 'h264', canvas: offscreen, url, audio: player !== null }, [
      offscreen
    ]);

    return () => {
      if (failTimer) clearTimeout(failTimer);
      notificationApi.destroy(CONNECTION_FAILED_NOTIFICATION_KEY);
      worker.postMessage({ type: 'stop' });
      worker.terminate();
      player?.close();
      playerRef.current = null;
      setHasAudio(false);
    };
  }, [setHasAudio, notificationApi]);

  // The unmute click is the user act the browser requires before it plays
  // sound, so the player makes its AudioContext here and not before.
  useEffect(() => {
    isMutedRef.current = isMuted;
    playerRef.current?.setMuted(isMuted);
  }, [isMuted]);

  return (
    <div className="relative h-full min-h-0 w-full min-w-0 overflow-hidden">
      {contextHolder}

      <ScreenViewport>
        <canvas
          id="screen"
          ref={canvasRef}
          className={clsx('block touch-none select-none', mouseStyle)}
        ></canvas>
      </ScreenViewport>

      {isLoading && (
        <div className="pointer-events-none absolute inset-0 z-50 flex items-center justify-center bg-black/30 backdrop-blur-[2px] transition-all duration-300">
          <Spin size="large" />
        </div>
      )}
    </div>
  );
};
