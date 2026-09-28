import { useState } from 'react';
import { UserSwitchOutlined } from '@ant-design/icons';
import { Button, Card } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/tailscale.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { ErrorDetail } from '../vpn/error-detail.tsx';
import { LoginUrl } from '../vpn/login-url.tsx';

type LoginProps = {
  onSuccess: () => void;
};

export const Login = ({ onSuccess }: LoginProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [loginUrl, setLoginUrl] = useState('');
  const [errMsg, setErrMsg] = useState('');

  function login() {
    if (isLoading) return;
    setIsLoading(true);
    setErrMsg('');

    api
      .login()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }

        const url = rsp.data.url;
        if (!url) {
          onSuccess();
          return;
        }

        setLoginUrl(url);
        setTimeout(() => setLoginUrl(''), 10 * 60 * 1000);
      })
      .catch((err) => {
        setErrMsg(describeFailure(err, t('settings.vpn.loginFailed')));
      })
      .finally(() => {
        setIsLoading(false);
      });
  }

  return (
    <div className="flex flex-col items-center justify-center space-y-10">
      <Card>{t('settings.tailscale.notLogin')}</Card>

      {loginUrl === '' ? (
        <Button
          type="primary"
          size="large"
          shape="round"
          icon={<UserSwitchOutlined />}
          loading={isLoading}
          onClick={login}
        >
          {t('settings.tailscale.login')}
        </Button>
      ) : (
        <LoginUrl
          url={loginUrl}
          period={t('settings.tailscale.urlPeriod')}
          getStatus={api.getStatus}
          onSuccess={onSuccess}
        />
      )}

      <ErrorDetail message={errMsg} />
    </div>
  );
};
