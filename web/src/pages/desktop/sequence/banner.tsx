import { useEffect, useRef, useState } from 'react';
import { message } from 'antd';
import { useAtom, useAtomValue, useSetAtom } from 'jotai';
import { CircleIcon, RepeatIcon, SquareIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/sequence.ts';
import type { SequenceStatus } from '@/api/sequence.ts';
import { showFailure } from '@/lib/feedback.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { macroDraftAtom, macroRecordingAtom, sequenceStatusAtom } from '@/jotai/sequence.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

import { discardRecording, finishRecording } from './recording.ts';

// How often the banner reads the run's progress while one runs.
const pollMs = 500;

const failures = new Set(['busy', 'hid_error', 'atx_error']);

const pillClassName =
  'fixed top-3 left-1/2 z-1000 flex -translate-x-1/2 items-center space-x-3 rounded-full border bg-neutral-900/90 px-4 py-1.5 text-xs shadow-lg';

const buttonClassName =
  'flex cursor-pointer items-center space-x-1 rounded px-2 py-0.5 text-neutral-200 hover:bg-neutral-700/70 disabled:cursor-wait';

// SequenceBanner shows an auto-tap or a macro while the server plays it, with
// a button that stops it, and the macro recording while one is made. The run
// may have been started by another tab or before a reload, so the state is
// read from the server when the page opens.
export const SequenceBanner = () => {
  const { t } = useTranslation();
  const autoTap = useServerFeature('hid-autotap');
  const macros = useServerFeature('hid-macros');
  const supported = autoTap === true || macros === true;

  const [status, setStatus] = useAtom(sequenceStatusAtom);
  const recording = useAtomValue(macroRecordingAtom);
  const setDraft = useSetAtom(macroDraftAtom);
  const [isStopping, setIsStopping] = useState(false);
  const seenRunning = useRef<SequenceStatus | null>(null);

  useEffect(() => {
    if (!supported) return;
    let live = true;
    api
      .getSequenceStatus()
      .then((rsp) => live && rsp.code === 0 && setStatus(rsp.data))
      .catch(() => {});
    return () => {
      live = false;
    };
  }, [supported, setStatus]);

  const running = status?.running === true;
  useEffect(() => {
    if (!running) return;
    return pollWhileVisible(() => {
      api
        .getSequenceStatus()
        .then((rsp) => rsp.code === 0 && setStatus(rsp.data))
        .catch(() => {});
    }, pollMs);
  }, [running, setStatus]);

  // Say how a run this tab watched has ended.
  useEffect(() => {
    if (!status) return;
    if (status.running) {
      seenRunning.current = status;
      return;
    }
    if (!seenRunning.current) return;
    seenRunning.current = null;

    const text = t(
      status.kind === 'autotap' ? 'keyboard.sequence.endedAutoTap' : 'keyboard.sequence.endedMacro',
      {
        name: status.name,
        reason: t(`keyboard.sequence.reasons.${status.reason || 'finished'}`)
      }
    );
    if (failures.has(status.reason)) message.error(text);
    else message.info(text);
  }, [status, t]);

  function stop() {
    if (isStopping) return;
    setIsStopping(true);
    api
      .stopSequence()
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        setStatus(rsp.data);
      })
      .catch((err) => showFailure(err))
      .finally(() => setIsStopping(false));
  }

  function finish() {
    setDraft({ name: '', steps: finishRecording() });
  }

  if (recording) {
    return (
      <div className={`${pillClassName} border-red-500/60 text-neutral-200`}>
        <CircleIcon size={10} className="animate-pulse fill-red-500 text-red-500" />
        <span>
          {recording.full
            ? t('keyboard.macro.recordingFull')
            : t('keyboard.macro.recording', { count: recording.count })}
        </span>
        <button type="button" className={buttonClassName} onClick={finish}>
          <SquareIcon size={12} />
          <span>{t('keyboard.macro.finish')}</span>
        </button>
        <button type="button" className={buttonClassName} onClick={discardRecording}>
          <span>{t('keyboard.macro.discard')}</span>
        </button>
      </div>
    );
  }

  if (!status?.running) return null;

  const label =
    status.kind === 'autotap'
      ? t('keyboard.sequence.autoTapRunning', { name: status.name, taps: status.taps })
      : t('keyboard.sequence.macroRunning', {
          name: status.name,
          step: status.step,
          total: status.total
        });

  return (
    <div className={`${pillClassName} border-blue-500/60 text-neutral-200`}>
      <RepeatIcon size={14} className="text-blue-400" />
      <span>{label}</span>
      {status.kind === 'autotap' && (
        <span className="text-neutral-400">
          {t('keyboard.sequence.elapsed', {
            elapsed: Math.floor(status.elapsedMs / 1000),
            timeout: Math.round(status.timeoutMs / 1000)
          })}
        </span>
      )}
      <button type="button" className={buttonClassName} disabled={isStopping} onClick={stop}>
        <SquareIcon size={12} className="fill-current" />
        <span>{t('keyboard.sequence.stop')}</span>
      </button>
    </div>
  );
};
