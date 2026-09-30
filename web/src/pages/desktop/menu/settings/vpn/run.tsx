import { useState } from 'react';
import { Button, Card, Result } from 'antd';
import { CirclePauseIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';

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
          onError(describeFailure(rsp));
          return;
        }
        onSuccess();
      })
      .catch((err) => onError(describeFailure(err)))
      .finally(() => setIsLoading(false));
  }

  return (
    <Card>
      <Result
        icon={<CirclePauseIcon size={72} />}
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
