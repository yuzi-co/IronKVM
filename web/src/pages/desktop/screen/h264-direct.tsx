import { useEffect, useRef } from 'react';
import clsx from 'clsx';
import { useAtomValue, useSetAtom } from 'jotai';

import { getBaseUrl } from '@/lib/service.ts';
import { audioMutedAtom, hasAudioAtom } from '@/jotai/audio.ts';
import { mouseStyleAtom } from '@/jotai/mouse';

import { DirectAudioPlayer } from './direct-audio.ts';
import DirectWorker from './direct.worker.ts?worker';
import { ScreenViewport } from './viewport.tsx';

type WorkerEvent = {
  type?: string;
  width?: number;
  height?: number;
  seq?: number;
  data?: ArrayBuffer;
};

export const H264Direct = () => {
  const mouseStyle = useAtomValue(mouseStyleAtom);
  const isMuted = useAtomValue(audioMutedAtom);
  const setHasAudio = useSetAtom(hasAudioAtom);

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const workerRef = useRef<Worker | null>(null);
  const playerRef = useRef<DirectAudioPlayer | null>(null);
  const isMutedRef = useRef(isMuted);

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

    const offscreen = canvasRef.current.transferControlToOffscreen();
    const url = `${getBaseUrl('ws')}/api/stream/h264/direct`;
    worker.onmessage = (event: MessageEvent<WorkerEvent>) => {
      const { type, width, height, seq, data } = event.data;

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
      worker.postMessage({ type: 'stop' });
      worker.terminate();
      player?.close();
      playerRef.current = null;
      setHasAudio(false);
    };
  }, [setHasAudio]);

  // The unmute click is the user act the browser requires before it plays
  // sound, so the player makes its AudioContext here and not before.
  useEffect(() => {
    isMutedRef.current = isMuted;
    playerRef.current?.setMuted(isMuted);
  }, [isMuted]);

  return (
    <ScreenViewport>
      <canvas
        id="screen"
        ref={canvasRef}
        className={clsx('block touch-none select-none', mouseStyle)}
      ></canvas>
    </ScreenViewport>
  );
};
