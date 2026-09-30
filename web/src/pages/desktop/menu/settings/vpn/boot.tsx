import { useState } from 'react';
import { Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';

import { SectionHeader } from '../components/section.tsx';
import type { VpnInfo } from './types.ts';

type BootProps = {
  vpn: VpnInfo;
  enabled: boolean;
  blocked: boolean;
  onChange: () => void;
  onError: (msg: string) => void;
};

// Boot is connect at boot. The boot script starts the daemon, and the server
// brings the network up once it answers. It is its own switch: connecting and
// disconnecting do not change it.
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
          onError(describeFailure(rsp));
          return;
        }
        onChange();
      })
      .catch((err) => onError(describeFailure(err)))
      .finally(() => setIsLoading(false));
  }

  return (
    <SectionHeader
      title={t('settings.vpn.connectAtBoot')}
      description={t('settings.vpn.connectAtBootDesc', { name: vpn.title })}
      action={
        <Switch
          checked={enabled}
          loading={isLoading}
          disabled={blocked && !enabled}
          onChange={toggle}
        />
      }
    />
  );
};
