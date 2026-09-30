import { ReactElement, useEffect, useState } from 'react';
import { Alert, Button, Form, Input } from 'antd';
import { LockIcon, UserIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router';

import * as api from '@/api/auth.ts';
import { encrypt } from '@/lib/encrypt.ts';
import { Head } from '@/components/head.tsx';

import { loginNotice, returnTo } from './return-to.ts';
import { Tips } from './tips.tsx';

export const Login = (): ReactElement => {
  const navigate = useNavigate();
  const location = useLocation();
  const { t } = useTranslation();

  // Where to go once signed in, and what to say above the form.
  const target = returnTo(location.state);
  const notice = loginNotice(location.state);

  const [isLoading, setIsloading] = useState(false);
  const [msg, setMsg] = useState('');

  useEffect(() => {
    api
      .getAccount()
      .then((rsp) => {
        if (rsp.code === 0 && rsp.data?.username) {
          navigate(target, { replace: true });
        }
      })
      .catch(() => {});
  }, [navigate, target]);

  useEffect(() => {
    if (msg) {
      setTimeout(() => setMsg(''), 3000);
    }
  }, [msg]);

  function login(values: any) {
    if (isLoading) return;
    setIsloading(true);

    const username = values.username;
    const password = encrypt(values.password);

    api
      .login(username, password)
      .then((rsp: any) => {
        if (rsp.code !== 0) {
          let errorMsg = t('auth.error');
          if (rsp.code === -2) errorMsg = t('auth.invalidUser');
          else if (rsp.code === -5) errorMsg = t('auth.locked');
          else if (rsp.code === -4) errorMsg = t('auth.globalLocked');

          setMsg(errorMsg);
          return;
        }

        // A browser that discards the session cookie sends the operator back
        // here with no explanation, forever. That happens when the page is
        // served over plain http while the server still believes it is on
        // https: the cookie it sets carries Secure, and the browser throws it
        // away. The cookie is HttpOnly, so the page cannot look at it. Ask for
        // the account instead, which is the same question one step further on.
        return api
          .getAccount()
          .then((account: any) => {
            if (account.code !== 0 || !account.data?.username) {
              setMsg(t('auth.cookieRejected'));
              return;
            }

            setMsg('');
            navigate(target, { replace: true });
          })
          .catch(() => {
            setMsg(t('auth.cookieRejected'));
          });
      })
      .catch(() => {
        setMsg(t('auth.error'));
      })
      .finally(() => {
        setIsloading(false);
      });
  }

  return (
    <>
      <Head title={t('head.login')} />

      <div className="flex h-screen w-screen flex-col items-center justify-center">
        <Form
          style={{ minWidth: 300, maxWidth: 500 }}
          initialValues={{ remember: true }}
          onFinish={login}
        >
          <div className="flex flex-col items-center justify-center pb-4">
            <img
              id="logo"
              src="/ironkvm.ico"
              alt="IronKVM"
              onClick={(evt) => {
                evt.preventDefault();
                (evt.target as HTMLImageElement).classList.add('animate-spin');
                setTimeout(() => {
                  (evt.target as HTMLImageElement).classList.remove('animate-spin');
                }, 1000);
              }}
            />
          </div>
          {notice === 'passwordChanged' && (
            <Alert className="mb-4!" type="success" showIcon message={t('auth.passwordChanged')} />
          )}

          <Form.Item
            name="username"
            rules={[{ required: true, message: t('auth.noEmptyUsername'), min: 1 }]}
          >
            <Input prefix={<UserIcon size={15} />} placeholder={t('auth.placeholderUsername')} />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[{ required: true, message: t('auth.noEmptyPassword'), min: 1 }]}
          >
            <Input
              prefix={<LockIcon size={15} />}
              type="password"
              placeholder={t('auth.placeholderPassword')}
            />
          </Form.Item>

          <div className="pb-1 text-red-500">{msg}</div>

          <Form.Item>
            <Button type="primary" htmlType="submit" className="w-full" loading={isLoading}>
              {t('auth.loginButtonText')}
            </Button>
          </Form.Item>

          <div className="flex justify-end pb-4 text-sm">
            <Tips />
          </div>
        </Form>
      </div>
    </>
  );
};
