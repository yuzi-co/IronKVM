import { useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { message, Switch, Tooltip } from 'antd';
import clsx from 'clsx';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm';

type PowerLedProps = {
  isPowerOn: boolean;
  connected: boolean;
  setConnected: (connected: boolean) => void;
};

// The host's power LED as the board sees it, and the switch that says whether
// the LED header is wired at all. Without it the line reads "off" whatever the
// host does, so the state is shown as unknown rather than off.
export const PowerLed = ({ isPowerOn, connected, setConnected }: PowerLedProps) => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const [isSaving, setIsSaving] = useState(false);

  async function update(value: boolean) {
    setIsSaving(true);
    try {
      const rsp = await api.setPowerLed(value);
      if (rsp.code !== 0) {
        message.error(t('power.ledConnectedFailed'));
        return;
      }
      setConnected(value);
    } catch (err) {
      console.log(err);
      message.error(t('power.ledConnectedFailed'));
    } finally {
      setIsSaving(false);
    }
  }

  let state = t('power.ledUnknown');
  if (connected) {
    state = isPowerOn ? t('power.ledOn') : t('power.ledOff');
  }

  return (
    <div className="flex flex-col space-y-2 px-1">
      <div className="flex items-center justify-between">
        <span className="text-xs text-neutral-400">{t('power.led')}</span>
        <div className="flex items-center space-x-1.5">
          <span
            className={clsx(
              'h-2 w-2 rounded-full',
              !connected && 'border border-neutral-500',
              connected && isPowerOn && 'bg-green-600',
              connected && !isPowerOn && 'bg-neutral-600'
            )}
          />
          <span className="text-xs text-neutral-300">{state}</span>
        </div>
      </div>

      <Tooltip title={t('power.ledConnectedTip')} placement="right">
        <div className="flex items-center justify-between">
          <span className="text-xs text-neutral-400">{t('power.ledConnected')}</span>
          <Switch
            size="small"
            checked={connected}
            loading={isSaving}
            disabled={account.role !== 'admin'}
            onChange={update}
          />
        </div>
      </Tooltip>
    </div>
  );
};
