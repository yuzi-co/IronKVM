import { useState } from 'react';
import { Button, Card, Result } from 'antd';
import { DownloadIcon, InfoIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';

import type { VpnInfo } from './types.ts';

type InstallProps = {
  vpn: VpnInfo;
  blocked: boolean;
  setIsLocked: (isLocked: boolean) => void;
  onSuccess: () => void;
  onError: (msg: string) => void;
};

export const Install = ({ vpn, blocked, setIsLocked, onSuccess, onError }: InstallProps) => {
  const { t } = useTranslation();

  const [state, setState] = useState<'' | 'installing' | 'failed'>('');

  function install() {
    if (state === 'installing' || blocked) return;
    setState('installing');
    setIsLocked(true);
    onError('');

    vpn.api
      .install()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setState('failed');
          onError(rsp.msg);
          return;
        }

        setState('');
        onSuccess();
      })
      .catch((err) => {
        setState('failed');
        onError(describeFailure(err, t('settings.vpn.installFailed')));
      })
      .finally(() => {
        setIsLocked(false);
      });
  }

  if (state === 'failed') {
    return (
      <Result
        status="warning"
        title={t('settings.vpn.installFailed')}
        icon={<InfoIcon size={72} />}
        extra={
          <div className="flex flex-col items-center space-y-4">
            <Button onClick={() => setState('')}>{t('settings.vpn.retry')}</Button>
            {vpn.installHelp}
          </div>
        }
      />
    );
  }

  return (
    <Card>
      <Result
        icon={<DownloadIcon size={72} />}
        subTitle={t('settings.vpn.notInstall', { name: vpn.title })}
        extra={
          <Button
            key="install"
            type="primary"
            loading={state === 'installing'}
            disabled={blocked}
            onClick={install}
          >
            {state === 'installing' ? t('settings.vpn.installing') : t('settings.vpn.install')}
          </Button>
        }
      />
    </Card>
  );
};
