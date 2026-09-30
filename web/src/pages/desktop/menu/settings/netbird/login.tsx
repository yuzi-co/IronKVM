import { useState } from 'react';
import { Button, Card, Divider, Input } from 'antd';
import { KeyRoundIcon, LogInIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/netbird.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { ErrorDetail } from '../vpn/error-detail.tsx';
import { LoginUrl } from '../vpn/login-url.tsx';

type LoginProps = {
  onSuccess: () => void;
};

// Login joins with a setup key, or logs in with SSO. The key stays in this
// form's state and goes to the server once; nothing stores it.
export const Login = ({ onSuccess }: LoginProps) => {
  const { t } = useTranslation();

  const [setupKey, setSetupKey] = useState('');
  const [loading, setLoading] = useState<'' | 'key' | 'sso'>('');
  const [loginUrl, setLoginUrl] = useState('');
  const [errMsg, setErrMsg] = useState('');

  function join() {
    const key = setupKey.trim();
    if (loading !== '' || !key) return;
    setLoading('key');
    setErrMsg('');

    api
      .login(key)
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }
        setSetupKey('');
        onSuccess();
      })
      .catch((err) => setErrMsg(describeFailure(err, t('settings.netbird.joinFailed'))))
      .finally(() => setLoading(''));
  }

  function sso() {
    if (loading !== '') return;
    setLoading('sso');
    setErrMsg('');

    api
      .login()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }

        const url = rsp.data?.url;
        if (!url) {
          onSuccess();
          return;
        }

        setLoginUrl(url);
        setTimeout(() => setLoginUrl(''), 10 * 60 * 1000);
      })
      .catch((err) => setErrMsg(describeFailure(err, t('settings.vpn.loginFailed'))))
      .finally(() => setLoading(''));
  }

  return (
    <div className="flex flex-col items-center justify-center space-y-8 pt-5">
      <Card>{t('settings.netbird.notLogin')}</Card>

      {loginUrl === '' ? (
        <div className="flex w-full max-w-[420px] flex-col space-y-4">
          <span>{t('settings.netbird.setupKey')}</span>
          <div className="flex space-x-2">
            <Input.Password
              value={setupKey}
              placeholder={t('settings.netbird.setupKeyPlaceholder')}
              autoComplete="off"
              onChange={(e) => setSetupKey(e.target.value)}
              onPressEnter={join}
            />
            <Button
              type="primary"
              icon={<KeyRoundIcon size={15} />}
              loading={loading === 'key'}
              disabled={!setupKey.trim() || loading === 'sso'}
              onClick={join}
            >
              {t('settings.netbird.join')}
            </Button>
          </div>

          <Divider plain>{t('settings.netbird.or')}</Divider>

          <Button
            size="large"
            shape="round"
            icon={<LogInIcon size={15} />}
            loading={loading === 'sso'}
            disabled={loading === 'key'}
            onClick={sso}
          >
            {t('settings.netbird.sso')}
          </Button>
        </div>
      ) : (
        <LoginUrl
          url={loginUrl}
          period={t('settings.netbird.urlPeriod')}
          getStatus={api.getStatus}
          onSuccess={onSuccess}
        />
      )}

      <ErrorDetail message={errMsg} />
    </div>
  );
};
