import type { ReactNode } from 'react';
import { Tooltip } from 'antd';
import clsx from 'clsx';
import type { TFunction } from 'i18next';
import { useAtomValue } from 'jotai';
import { useTranslation } from 'react-i18next';

import { CAPTURE_NO_SIGNAL } from '@/lib/health.ts';
import { codecName } from '@/lib/stream-codec.ts';
import {
  captureStatusAtom,
  screenSettingsAtom,
  streamInfoVisibleAtom,
  streamPausedAtom,
  streamStatsAtom,
  videoModeAtom
} from '@/jotai/screen.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

import {
  formatBitrate,
  formatFps,
  formatLatency,
  formatRefresh,
  withBoardPart,
  type ConnectionState,
  type Latency,
  type StreamStats
} from './model.ts';
import { useSourceInfo } from './use-source-info.ts';

type Tone = 'ok' | 'warning' | 'error' | 'off';

const toneClass: Record<Tone, string> = {
  ok: 'bg-green-500',
  warning: 'bg-amber-400',
  error: 'bg-red-500',
  off: 'bg-neutral-500'
};

const connectionTone: Record<ConnectionState, Tone> = {
  connected: 'ok',
  connecting: 'warning',
  reconnecting: 'warning',
  failed: 'error',
  paused: 'off'
};

const modeNames: Record<string, string> = { h264: 'WebRTC', direct: 'Direct', mjpeg: 'MJPEG' };

// One field of the bar. The less important ones hide on a phone, where the
// bar is one line that scrolls sideways rather than wrapping.
const Item = ({
  title,
  wide,
  children
}: {
  title?: string;
  wide?: boolean;
  children: ReactNode;
}) => (
  <span title={title} className={clsx('shrink-0', wide && 'hidden sm:inline')}>
    {children}
  </span>
);

const Muted = ({ children }: { children: ReactNode }) => (
  <span className="text-neutral-500">{children}</span>
);

const size = (width: number, height: number) => `${width}×${height}`;

// StatusBar shows what the stream is doing under the picture, as PiKVM's
// stream window does in its title: the connection, the source and the stream,
// the frame rate, the bitrate and the latency. It is off unless switched on
// in the Screen menu, and the views measure nothing while it is off.
export const StatusBar = () => {
  const visible = useAtomValue(streamInfoVisibleAtom);
  if (!visible) return null;
  return <Bar />;
};

const Bar = () => {
  const { t } = useTranslation();
  const stats = useAtomValue(streamStatsAtom);
  const videoMode = useAtomValue(videoModeAtom);
  const paused = useAtomValue(streamPausedAtom);
  const capture = useAtomValue(captureStatusAtom);
  const settings = useAtomValue(screenSettingsAtom);
  const source = useSourceInfo();
  const boardLatency = useServerFeature('info-latency') === true;

  const noSignal = capture?.result === CAPTURE_NO_SIGNAL;
  const connection: ConnectionState = paused ? 'paused' : (stats?.connection ?? 'connecting');
  const tone: Tone = noSignal ? 'warning' : connectionTone[connection];
  const state = noSignal ? t('screen.streamInfo.noSignal') : t(`screen.streamInfo.${connection}`);

  const codec =
    stats?.codec ?? (videoMode === 'mjpeg' ? 'MJPEG' : settings ? codecName(settings.codec) : null);
  const mode = modeNames[videoMode] ?? videoMode;

  return (
    <div
      className="box-border flex h-6 w-full shrink-0 [scrollbar-width:none] items-center gap-x-3 overflow-x-auto border-t border-neutral-800 bg-neutral-900 px-3 font-mono text-[11px] leading-none whitespace-nowrap text-neutral-300"
      role="status"
      aria-label={t('screen.streamInfo.title')}
    >
      <Item>
        <span
          aria-hidden="true"
          className={clsx('mr-1.5 inline-block h-[7px] w-[7px] rounded-full', toneClass[tone])}
        />
        {state}
      </Item>

      <Item>{mode === 'MJPEG' || !codec || codec === 'MJPEG' ? mode : `${mode} ${codec}`}</Item>

      {source && source.width > 0 && !noSignal && (
        <Item title={t('screen.streamInfo.sourceTip')}>
          <Muted>{t('screen.streamInfo.source')} </Muted>
          {size(source.width, source.height)}
          {source.refresh > 0 && <> @ {formatRefresh(source.refresh)}</>}
        </Item>
      )}

      {stats && stats.width > 0 && (
        <Item title={t('screen.streamInfo.streamTip')} wide={sameSize(stats, source)}>
          <Muted>{t('screen.streamInfo.stream')} </Muted>
          {size(stats.width, stats.height)}
        </Item>
      )}

      {stats && <FrameRate stats={stats} captureFps={source?.captureFps} />}

      {stats?.kbps !== null && stats?.kbps !== undefined && (
        <Item title={t('screen.streamInfo.bitrateTip')}>{formatBitrate(stats.kbps)}</Item>
      )}

      {stats?.latency && (
        <LatencyItem
          latency={withBoardPart(
            directFromCapture(stats.latency, source?.captureStamps),
            boardLatency ? source?.boardLatency?.webrtc : undefined
          )}
        />
      )}

      {stats?.rttMs !== null && stats?.rttMs !== undefined && (
        <Item title={t('screen.streamInfo.rttTip')} wide>
          <Muted>RTT </Muted>
          {Math.round(stats.rttMs)} ms
        </Item>
      )}

      {stats?.lossPct !== null && stats?.lossPct !== undefined && (
        <Item title={t('screen.streamInfo.lossTip')}>
          <Muted>{t('screen.streamInfo.loss')} </Muted>
          <span className={stats.lossPct >= 1 ? 'text-amber-400' : undefined}>
            {stats.lossPct.toFixed(1)}%
          </span>
        </Item>
      )}

      {stats?.transport && (
        <Item title={t('screen.streamInfo.transportTip')} wide>
          <Muted>{stats.transport}</Muted>
        </Item>
      )}
    </div>
  );
};

function sameSize(stats: StreamStats, source: { width: number; height: number } | null) {
  return !!source && source.width === stats.width && source.height === stats.height;
}

const FrameRate = ({ stats, captureFps }: { stats: StreamStats; captureFps?: number }) => {
  const { t } = useTranslation();
  const shown = stats.fpsDecoded ?? stats.fpsReceived;
  const board = captureFps !== undefined && captureFps >= 0 ? captureFps : null;
  if (shown === null && board === null) return null;

  const parts = [
    stats.fpsReceived !== null &&
      `${t('screen.streamInfo.received')} ${formatFps(stats.fpsReceived)}`,
    stats.fpsDecoded !== null && `${t('screen.streamInfo.decoded')} ${formatFps(stats.fpsDecoded)}`,
    stats.fpsDropped !== null && `${t('screen.streamInfo.dropped')} ${formatFps(stats.fpsDropped)}`,
    board !== null && `${t('screen.streamInfo.board')} ${board}`
  ].filter(Boolean);

  return (
    <>
      {shown !== null && <Item title={parts.join(' · ')}>{formatFps(shown)} fps</Item>}
      {stats.fpsDropped !== null && stats.fpsDropped >= 0.5 && (
        <Item title={parts.join(' · ')}>
          <span className="text-amber-400">
            {t('screen.streamInfo.dropped')} {formatFps(stats.fpsDropped)}/s
          </span>
        </Item>
      )}
      {board !== null && (
        <Item title={t('screen.streamInfo.boardTip')} wide>
          <Muted>{t('screen.streamInfo.board')} </Muted>
          {board} fps
        </Item>
      )}
    </>
  );
};

// H.264 Direct's age at drawing counts from the capture, not the encoder,
// when the board stamps its frames with their capture time.
function directFromCapture(latency: Latency, captureStamps?: boolean): Latency {
  return latency.kind === 'encoder' && captureStamps ? { ...latency, kind: 'capture' } : latency;
}

const latencyTips: Record<Latency['kind'], string> = {
  capture: 'screen.streamInfo.latencyCapture',
  encoder: 'screen.streamInfo.latencyEncoder',
  estimate: 'screen.streamInfo.latencyEstimate'
};

// The tooltip's first sentence: what kind of number it is, and for an
// estimate the board reported its share of, the parts it adds up.
function latencyTip(latency: Latency, t: TFunction): string {
  const parts = latency.parts;
  if (latency.kind === 'estimate' && parts?.boardMs !== undefined) {
    return t('screen.streamInfo.latencyEstimateSum', {
      board: Math.round(parts.boardMs),
      network: Math.round(parts.networkMs),
      receive: Math.round(parts.receiveMs),
      display: Math.round(parts.displayMs ?? 0)
    });
  }
  return t(latencyTips[latency.kind]);
}

const LatencyItem = ({ latency }: { latency: Latency }) => {
  const { t } = useTranslation();
  const tip = [
    latencyTip(latency, t),
    latency.clockAssumed ? t('screen.streamInfo.latencyClockAssumed') : '',
    latency.clockOffsetMs !== undefined
      ? t('screen.streamInfo.latencyClockOffset', { offset: Math.round(latency.clockOffsetMs) })
      : '',
    latency.arrivalMs !== undefined
      ? t('screen.streamInfo.latencyArrival', { ms: Math.round(latency.arrivalMs) })
      : ''
  ]
    .filter(Boolean)
    .join(' ');

  return (
    <Tooltip title={tip} placement="top" color="#262626">
      <span className="shrink-0 cursor-help">
        <Muted>{t('screen.streamInfo.latency')} </Muted>
        {formatLatency(latency)}
      </span>
    </Tooltip>
  );
};
