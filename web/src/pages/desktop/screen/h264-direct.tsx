import { useEffect, useRef, useState } from 'react';
import { notification, Spin } from 'antd';
import clsx from 'clsx';
import { useAtom, useAtomValue, useSetAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import { audioStateFromByte } from '@/lib/audio-state.ts';
import { nextRoomMicSwitch, roomMicPushFromBytes } from '@/lib/room-mic.ts';
import { getBaseUrl } from '@/lib/service.ts';
import { audioMutedAtom, audioStateAtom, hasAudioAtom } from '@/jotai/audio.ts';
import { mouseStyleAtom } from '@/jotai/mouse';
import {
  roomMicListeningAtom,
  roomMicMutedAtom,
  roomMicPushAtom,
  roomMicSwitchAtom,
  roomMicTransportAtom,
  roomMicVolumeAtom
} from '@/jotai/room-mic.ts';

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
  value?: number;
  bytes?: Uint8Array;
};

export const H264Direct = () => {
  const { t } = useTranslation();
  const mouseStyle = useAtomValue(mouseStyleAtom);
  const isMuted = useAtomValue(audioMutedAtom);
  const setHasAudio = useSetAtom(hasAudioAtom);
  const setAudioState = useSetAtom(audioStateAtom);
  const [roomSwitch, setRoomSwitch] = useAtom(roomMicSwitchAtom);
  const roomMuted = useAtomValue(roomMicMutedAtom);
  const roomVolume = useAtomValue(roomMicVolumeAtom);
  const setRoomListening = useSetAtom(roomMicListeningAtom);
  const setRoomTransport = useSetAtom(roomMicTransportAtom);
  const pushRoomState = useSetAtom(roomMicPushAtom);

  const canvasRef = useRef<HTMLCanvasElement>(null);
  const workerRef = useRef<Worker | null>(null);
  const playerRef = useRef<DirectAudioPlayer | null>(null);
  const roomPlayerRef = useRef<DirectAudioPlayer | null>(null);
  const roomWantedRef = useRef(roomSwitch.wanted);
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
    // The room microphone is one channel. It plays only while switched on.
    const roomPlayer = DirectAudioPlayer.supported() ? new DirectAudioPlayer(1) : null;
    roomPlayerRef.current = roomPlayer;
    let reportedAudio = false;
    let playing = false;

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
      const { type, state, width, height, seq, data, value, bytes } = event.data;

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
          // The server let go of the microphone with the socket. The worker
          // switches it on again on the next one.
          setRoomListening(false);
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
        // Sound arriving says the host plays, even if the notice saying so
        // was lost on the way.
        if (!playing) {
          playing = true;
          setAudioState('playing');
        }
        player.push(seq, data);
        return;
      }

      // The server says what capture is doing. It sends this only to a viewer
      // that asked for audio on a board that has it, so it also means there is
      // a speaker to show, before the host has played anything.
      if (type === 'audio-state' && player && value !== undefined) {
        if (!reportedAudio) {
          reportedAudio = true;
          setHasAudio(true);
        }
        const next = audioStateFromByte(value);
        playing = next === 'playing';
        setAudioState(next);
        return;
      }

      // The microphone's state comes only on a kernel that has one, so it
      // also says this connection can carry it.
      if (type === 'room-state' && bytes) {
        const push = roomMicPushFromBytes(bytes);
        if (push) {
          setRoomTransport(true);
          pushRoomState(push);
          setRoomListening(push.listening);
          setRoomSwitch((current) => nextRoomMicSwitch(current, push));
        }
        return;
      }

      if (type === 'room-audio' && roomPlayer && seq !== undefined && data) {
        roomPlayer.push(seq, data);
        return;
      }

      if (type !== 'frame-size' || !width || !height || !canvasRef.current) {
        return;
      }

      canvasRef.current.dataset.mediaWidth = String(width);
      canvasRef.current.dataset.mediaHeight = String(height);
    };
    worker.postMessage(
      {
        type: 'h264',
        canvas: offscreen,
        url,
        audio: player !== null,
        room: roomPlayer !== null
      },
      [offscreen]
    );
    if (roomWantedRef.current) {
      worker.postMessage({ type: 'room-mic', on: true });
    }

    return () => {
      if (failTimer) clearTimeout(failTimer);
      notificationApi.destroy(CONNECTION_FAILED_NOTIFICATION_KEY);
      worker.postMessage({ type: 'stop' });
      worker.terminate();
      player?.close();
      playerRef.current = null;
      roomPlayer?.close();
      roomPlayerRef.current = null;
      setHasAudio(false);
      setAudioState('unknown');
      setRoomTransport(false);
      setRoomListening(false);
    };
  }, [
    setHasAudio,
    setAudioState,
    notificationApi,
    setRoomTransport,
    setRoomListening,
    pushRoomState,
    setRoomSwitch
  ]);

  // The switch goes to the server through the worker, which sends it again
  // on every reconnect. The flip is the user act that lets the microphone's
  // player start its AudioContext.
  useEffect(() => {
    roomWantedRef.current = roomSwitch.wanted;
    workerRef.current?.postMessage({ type: 'room-mic', on: roomSwitch.wanted });
    roomPlayerRef.current?.setMuted(roomMuted || !roomSwitch.wanted);
  }, [roomSwitch.wanted, roomMuted]);

  useEffect(() => {
    roomPlayerRef.current?.setVolume(roomVolume);
  }, [roomVolume]);

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
