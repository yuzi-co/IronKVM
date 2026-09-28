import { useState } from 'react';
import { Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import type { VpnInfo } from './types.ts';

type BootProps = {
  vpn: VpnInfo;
  enabled: boolean;
  blocked: boolean;
  onChange: () => void;
  onError: (msg: string) => void;
};

// Boot is start at boot. It is its own switch: starting and stopping the
// daemon no longer change it.
export const Boot = ({ vpn, enabled, blocked, onChange, onError }: BootProps) => {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  function toggle(next: boolean) {
    if (isLoading) return;
    setIsLoading(true);
    onError('');

    vpn.api
      .setBoot(next)
      .then((rsp) => {
        if (rsp.code !== 0) {
          onError(rsp.msg);
          return;
        }
        onChange();
      })
      .catch((err) => onError(err?.message || 'Failed to set start at boot'))
      .finally(() => setIsLoading(false));
  }

  return (
    <div className="flex items-center justify-between">
      <div className="flex flex-col">
        <span>{t('settings.vpn.boot')}</span>
        <span className="text-xs text-neutral-500">
          {t('settings.vpn.bootDesc', { name: vpn.title })}
        </span>
      </div>
      <Switch
        checked={enabled}
        loading={isLoading}
        disabled={blocked && !enabled}
        onChange={toggle}
      />
    </div>
  );
};
