import { useState } from 'react';
import { PauseCircleOutlined } from '@ant-design/icons';
import { Button, Card, Result } from 'antd';
import { useTranslation } from 'react-i18next';

import type { VpnInfo } from './types.ts';

type RunProps = {
  vpn: VpnInfo;
  blocked: boolean;
  onSuccess: () => void;
  onError: (msg: string) => void;
};

export const Run = ({ vpn, blocked, onSuccess, onError }: RunProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);

  function run() {
    if (isLoading || blocked) return;
    setIsLoading(true);
    onError('');

    vpn.api
      .start()
      .then((rsp) => {
        if (rsp.code !== 0) {
          onError(rsp.msg);
          return;
        }
        onSuccess();
      })
      .catch((err) => onError(err?.message || 'Failed to start'))
      .finally(() => setIsLoading(false));
  }

  return (
    <Card>
      <Result
        icon={<PauseCircleOutlined />}
        subTitle={t('settings.vpn.notRunning', { name: vpn.title })}
        extra={
          <Button key="run" type="primary" loading={isLoading} disabled={blocked} onClick={run}>
            {t('settings.vpn.run')}
          </Button>
        }
      />
    </Card>
  );
};
