import { useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { message, Switch, Tooltip } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm';

import { PowerLedState } from './power-led-state.tsx';

type PowerLedProps = {
  isPowerOn: boolean;
  connected: boolean;
  setConnected: (connected: boolean) => void;
};

// The host's power LED as the board sees it, and the switch that says whether
// the LED header is wired at all. The switch is a setting; the power menu shows
// only the state line.
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

  return (
    <div className="flex flex-col space-y-2 px-1">
      <PowerLedState isPowerOn={isPowerOn} connected={connected} />

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
