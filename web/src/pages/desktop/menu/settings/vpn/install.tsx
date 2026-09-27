import { useState } from 'react';
import { DownloadOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { Button, Card, Result } from 'antd';
import { useTranslation } from 'react-i18next';

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
        onError(err?.message || 'Install failed');
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
        icon={<InfoCircleOutlined />}
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
        icon={<DownloadOutlined />}
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
