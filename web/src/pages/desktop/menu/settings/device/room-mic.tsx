import { useState } from 'react';
import { Slider, Switch } from 'antd';
import { useAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/room-mic.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import {
  clampGain,
  parseRoomMicStatus,
  ROOM_MIC_MAX_GAIN,
  ROOM_MIC_MIN_GAIN,
  roomMicAdminState
} from '@/lib/room-mic.ts';
import { roomMicStatusAtom } from '@/jotai/room-mic.ts';

import { Section } from '../components/section.tsx';

// RoomMic is the administrator's lock on the board's microphone. It is off by
// default. Allowed, each viewer may switch it on in the audio menu, and every
// viewer sees when it is on. A kernel without the onboard card (slot A) says
// so instead.
export const RoomMic = () => {
  const { t } = useTranslation();
  const [status, setStatus] = useAtom(roomMicStatusAtom);
  const [isSaving, setIsSaving] = useState(false);
  // The slider's value while it is dragged; the server is asked on release.
  const [draftGain, setDraftGain] = useState<number | null>(null);

  const state = roomMicAdminState(status);

  function save(settings: { allowed?: boolean; gain?: number }) {
    if (isSaving) return;
    setIsSaving(true);

    api
      .setRoomMic(settings)
      .then((rsp) => {
        if (!showResult(rsp)) return;
        setStatus(parseRoomMicStatus(rsp.data));
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        setIsSaving(false);
        setDraftGain(null);
      });
  }

  const gain = draftGain ?? status?.gain ?? 0;

  return (
    <Section
      loose
      title={t('settings.device.sections.roomMic')}
      description={t('settings.device.roomMic.description')}
    >
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1">
          <span>{t('settings.device.roomMic.allow')}</span>
          {state === 'unavailable' && (
            <span className="text-xs text-neutral-500">
              {t('settings.device.roomMic.unavailable')}
            </span>
          )}
        </div>

        <Switch
          checked={state === 'ready' && !!status?.allowed}
          disabled={state !== 'ready'}
          loading={state === 'loading' || isSaving}
          onChange={(allowed) => save({ allowed })}
        />
      </div>

      {state === 'ready' && status?.allowed && (
        <div className="flex flex-col space-y-1">
          <div className="flex items-center justify-between">
            <span>{t('settings.device.roomMic.gain')}</span>
            <span className="text-xs text-neutral-500">{gain * 2} dB</span>
          </div>
          <Slider
            min={ROOM_MIC_MIN_GAIN}
            max={ROOM_MIC_MAX_GAIN}
            step={1}
            value={gain}
            disabled={isSaving}
            tooltip={{ formatter: (value) => `${(value ?? 0) * 2} dB` }}
            onChange={(value) => setDraftGain(clampGain(value))}
            onChangeComplete={(value) => save({ gain: clampGain(value) })}
          />
        </div>
      )}
    </Section>
  );
};
