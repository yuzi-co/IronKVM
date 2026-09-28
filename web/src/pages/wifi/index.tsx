import { useEffect, useState } from 'react';
import { CheckOutlined, KeyOutlined, LockOutlined, WifiOutlined } from '@ant-design/icons';
import { Button, Form, Input } from 'antd';
import { useTranslation } from 'react-i18next';
import { useSearchParams } from 'react-router';

import * as api from '@/api/network.ts';
import { wifiCredentialsError } from '@/lib/wifi.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';
import { Head } from '@/components/head.tsx';

type State = '' | 'loading' | 'success' | 'failed' | 'denied' | 'lost' | 'done';
type VerifyState = '' | 'failed' | 'denied';

export const Wifi = () => {
  const { t } = useTranslation();
  const [searchParams] = useSearchParams();

  const [state, setState] = useState<State>('');
  const [apPassword, setApPassword] = useState<string>('');
  const [isAuthenticated, setIsAuthenticated] = useState<boolean>(false);
  const [verifying, setVerifying] = useState<boolean>(false);
  const [verifyState, setVerifyState] = useState<VerifyState>('');

  const verifyPassword = useStableCallback(async (password: string) => {
    if (verifying) return;
    setVerifying(true);
    setVerifyState('');

    try {
      const rsp = await api.verifyApLogin(password);
      if (rsp?.code === 0) {
        setApPassword(password);
        setIsAuthenticated(true);
      } else {
        setVerifyState(rsp?.code === -1 ? 'denied' : 'failed');
      }
    } catch (err) {
      console.error(err);
      setVerifyState('failed');
    }
    setVerifying(false);
  });

  // A password in the link is tried once, when the page opens. The parameter
  // is read from the first render so that a later change to the query string
  // does not submit it again.
  const [linkPassword] = useState(() => searchParams.get('p') || searchParams.get('P'));

  useEffect(() => {
    if (linkPassword) {
      verifyPassword(linkPassword);
    }
  }, [linkPassword, verifyPassword]);

  async function onVerifyFinish(values: any) {
    if (!values.apPassword) return;
    await verifyPassword(values.apPassword);
  }

  async function connect(values: any) {
    const ssid: string = values.ssid ?? '';
    const password: string = values.password ?? '';
    if (wifiCredentialsError(ssid, password) !== '') return;

    if (state === 'loading') return;
    setState('loading');

    try {
      const rsp = await api.connectWifiNoAuth(ssid, password, apPassword);

      switch (rsp?.code) {
        case 0:
          setState('success');
          return;
        case -1:
        case -4:
          setState('denied');
          return;
        default:
          setState('failed');
          return;
      }
    } catch (err) {
      // Joining the network closes the setup hotspot this page talks through,
      // and that can happen before the answer arrives. A lost connection is
      // therefore not a failure, but it is not a success either.
      console.log(err);
      setState('lost');
    }
  }

  // Finished: the board is on the new network and this page has nothing left
  // to talk to. A tab the operator opened cannot always be closed by script,
  // so the page also says what to do next.
  function finish() {
    setState('done');
    window.close();
  }

  if (!isAuthenticated) {
    return (
      <>
        <Head title={t('head.wifi')} />

        <div className="flex h-screen w-screen flex-col items-center justify-center">
          <Form
            style={{ minWidth: 300, maxWidth: 500 }}
            initialValues={{ remember: true }}
            onFinish={onVerifyFinish}
          >
            <div className="flex flex-col space-y-1 pb-5">
              <span className="text-center text-2xl font-semibold text-red-500">
                {t('wifi.ap.authTitle')}
              </span>
              <span className="text-center text-neutral-400">{t('wifi.ap.authDescription')}</span>
            </div>

            <Form.Item name="apPassword">
              <Input.Password prefix={<KeyOutlined />} placeholder={t('wifi.ap.passPlaceholder')} />
            </Form.Item>

            <Form.Item>
              <Button className="w-full" htmlType="submit" type="primary" loading={verifying}>
                {t('wifi.ap.verifyBtn')}
              </Button>
            </Form.Item>
          </Form>

          <div className="flex max-w-[500px] justify-center px-5 pt-3 md:px-10">
            {verifyState === 'failed' && (
              <span className="text-sm text-red-500">{t('wifi.ap.authFailed')}</span>
            )}
            {verifyState === 'denied' && (
              <span className="text-sm text-red-500">{t('wifi.invalidMode')}</span>
            )}
          </div>
        </div>
      </>
    );
  }

  return (
    <>
      <Head title={t('head.wifi')} />

      <div className="flex h-screen w-screen flex-col items-center justify-center">
        <Form
          style={{ minWidth: 300, maxWidth: 500 }}
          initialValues={{ remember: true }}
          onFinish={connect}
        >
          <div className="flex flex-col space-y-1 pb-5">
            <span className="text-center text-2xl font-semibold text-neutral-100">
              {t('wifi.title')}
            </span>
            <span className="text-center text-neutral-400">{t('wifi.description')}</span>
          </div>

          <Form.Item
            name="ssid"
            rules={[
              {
                validator: (_, value) =>
                  wifiCredentialsError(value ?? '', '') === 'ssid'
                    ? Promise.reject(new Error(t('wifi.ssidRequired')))
                    : Promise.resolve()
              }
            ]}
          >
            <Input prefix={<WifiOutlined />} placeholder="SSID" />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              {
                validator: (_, value) =>
                  wifiCredentialsError('x', value ?? '') === 'password'
                    ? Promise.reject(new Error(t('wifi.passwordLength')))
                    : Promise.resolve()
              }
            ]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder={t('wifi.passwordOptional')} />
          </Form.Item>

          <Form.Item>
            {state === 'success' || state === 'done' ? (
              <Button
                className="w-full"
                type="primary"
                icon={<CheckOutlined />}
                disabled={state === 'done'}
                onClick={finish}
              >
                {t('wifi.finishBtn')}
              </Button>
            ) : (
              <Button
                className="w-full"
                htmlType="submit"
                type="primary"
                loading={state === 'loading'}
              >
                {t('wifi.confirmBtn')}
              </Button>
            )}
          </Form.Item>
        </Form>

        <div className="flex max-w-[500px] justify-center px-5 pt-3 md:px-10">
          {state === 'success' && (
            <span className="text-sm text-green-500">{t('wifi.success')}</span>
          )}

          {state === 'done' && <span className="text-sm text-green-500">{t('wifi.done')}</span>}
          {state === 'lost' && <span className="text-sm text-amber-500">{t('wifi.lost')}</span>}
          {state === 'failed' && <span className="text-sm text-red-500">{t('wifi.failed')} </span>}
          {state === 'denied' && (
            <div className="flex flex-col items-center space-y-5">
              <span className="text-sm text-red-500">{t('wifi.invalidMode')} </span>
            </div>
          )}
        </div>
      </div>
    </>
  );
};
