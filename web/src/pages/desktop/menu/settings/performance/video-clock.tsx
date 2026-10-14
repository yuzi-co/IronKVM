import { useEffect, useState } from 'react';
import { Select, Tooltip } from 'antd';
import { CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { showFailure } from '@/lib/feedback.ts';

type Rates = { codec: number; vpss: number; bus: number };

type VideoClockState = {
  profile: string;
  options: { name: string; rates: Rates }[];
  rates: Rates;
  measured: boolean;
  pending: boolean;
  temperature: number;
};

// While a change waits for the video to start again, the rates are read again
// this often, so the line shows when the new profile took effect.
const PENDING_POLL_MS = 5000;

// VideoClock chooses the clock profile of the codec core, the scaler and the
// codec bus. Only the mainline kernel (slot B) has it; the settings page shows
// it only when the server advertises the video-clock feature.
export const VideoClock = () => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [state, setState] = useState<VideoClockState | null>(null);

  useEffect(() => {
    getVideoClock();
  }, []);

  useEffect(() => {
    if (!state?.pending) return;
    const timer = setTimeout(() => getVideoClock(true), PENDING_POLL_MS);
    return () => clearTimeout(timer);
  }, [state]);

  function getVideoClock(quiet = false) {
    if (!quiet) setIsLoading(true);

    api
      .getVideoClock()
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }
        setState(rsp.data);
      })
      .catch((err) => console.log(err))
      .finally(() => {
        if (!quiet) setIsLoading(false);
      });
  }

  function update(value: string) {
    if (isLoading) return;
    setIsLoading(true);

    api
      .setVideoClock(value)
      .then((rsp) => {
        if (rsp.code !== 0) showFailure(rsp);
      })
      .catch((err) => showFailure(err))
      .finally(() => {
        // The kernel applies it at once when the video is idle, so the rates
        // read back may already be the new ones.
        setIsLoading(false);
        getVideoClock();
      });
  }

  const options = (state?.options ?? []).map((o) => ({
    value: o.name,
    label: t(`settings.device.videoClock.${o.name}`, { defaultValue: o.name })
  }));

  function statusText(): string {
    if (!state) return '';

    const parts: string[] = [];
    if (state.measured) {
      parts.push(t('settings.device.videoClock.rates', state.rates));
    }
    if (state.temperature > 0) {
      parts.push(`${state.temperature.toFixed(1)}°C`);
    }
    return parts.join(' · ');
  }

  const text = statusText();

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col space-y-1">
        <div className="flex items-center space-x-2">
          <span>{t('settings.device.videoClock.title')}</span>

          <Tooltip
            title={t('settings.device.videoClock.tip')}
            className="cursor-pointer"
            placement="right"
            styles={{ root: { maxWidth: '400px' } }}
          >
            <CircleHelpIcon className="text-neutral-500" size={14} />
          </Tooltip>
        </div>

        <span className="text-xs text-neutral-500">
          {t('settings.device.videoClock.description')}
        </span>

        {state && text && (
          <span
            className={`w-fit cursor-default text-xs ${state.pending ? 'text-amber-500' : 'text-neutral-400'}`}
          >
            {text}
            {state.pending && ` · ${t('settings.device.videoClock.pending')}`}
          </span>
        )}
      </div>

      <Select
        style={{ width: 150 }}
        value={state?.profile}
        options={options}
        loading={isLoading}
        disabled={!state || options.length === 0}
        onChange={update}
      />
    </div>
  );
};
