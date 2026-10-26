import { ReactElement, useEffect, useState } from 'react';
import { Alert, Button, Form, Input } from 'antd';
import { KeyRoundIcon, LockIcon, LogInIcon, UserIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useLocation, useNavigate } from 'react-router';

import * as api from '@/api/auth.ts';
import { encrypt } from '@/lib/encrypt.ts';
import { oidcReason } from '@/lib/oidc.ts';
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
  // Why the server's OIDC callback sent the browser back here, if it did.
  const [oidcMessage] = useState(() => {
    const reason = oidcReason(location.search);
    return reason ? t(`auth.oidc.errors.${reason}`) : '';
  });
  const [msg, setMsg] = useState(oidcMessage);
  // Set while the password was right and the account still owes a TOTP code.
  const [challenge, setChallenge] = useState<string | null>(null);
  // The OIDC provider's button, when the server has one configured.
  const [oidc, setOidc] = useState<api.OidcInfo | null>(null);
  // Set when the operator unfolds a password form the settings fold away.
  const [showPassword, setShowPassword] = useState(false);

  useEffect(() => {
    api
      .getOidcInfo()
      .then((rsp) => {
        if (rsp.code === 0 && rsp.data?.enabled) setOidc(rsp.data);
      })
      // An older server has no OIDC: no button.
      .catch(() => {});
  }, []);

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
    // A refused OIDC login stays on screen: the operator left this page
    // and came back, and may not be looking yet.
    if (msg && msg !== oidcMessage) {
      setTimeout(() => setMsg(''), 3000);
    }
  }, [msg, oidcMessage]);

  // The login leaves the page: the server redirects to the provider.
  function signInWithOidc() {
    window.location.assign(api.OIDC_LOGIN_URL);
  }

  const passwordFolded = !!oidc?.hidePasswordLogin && !showPassword;

  // A session cookie was just set. A browser that discards it sends the
  // operator back here with no explanation, forever. That happens when the
  // page is served over plain http while the server still believes it is on
  // https: the cookie it sets carries Secure, and the browser throws it away.
  // The cookie is HttpOnly, so the page cannot look at it. Ask for the account
  // instead, which is the same question one step further on.
  function finishLogin() {
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
  }

  function login(values: any) {
    if (isLoading) return;
    setIsloading(true);

    const username = values.username;
    const password = encrypt(values.password);

    api
      .login(username, password)
      .then((rsp: any) => {
        if (rsp.code === api.LOGIN_TOTP_REQUIRED && rsp.data?.challenge) {
          setMsg('');
          setChallenge(rsp.data.challenge);
          return;
        }
        if (rsp.code !== 0) {
          let errorMsg = t('auth.error');
          if (rsp.code === -2) errorMsg = t('auth.invalidUser');
          else if (rsp.code === -5) errorMsg = t('auth.locked');
          else if (rsp.code === -4) errorMsg = t('auth.globalLocked');

          setMsg(errorMsg);
          return;
        }

        return finishLogin();
      })
      .catch(() => {
        setMsg(t('auth.error'));
      })
      .finally(() => {
        setIsloading(false);
      });
  }

  function submitCode(values: { code: string }) {
    if (isLoading || !challenge) return;
    setIsloading(true);

    api
      .loginTotp(challenge, values.code.trim())
      .then((rsp: any) => {
        if (rsp.code === api.LOGIN_TOTP_EXPIRED) {
          setChallenge(null);
          setMsg(t('totp.login.expired'));
          return;
        }
        if (rsp.code !== 0) {
          let errorMsg = t('auth.error');
          if (rsp.code === -2) errorMsg = t('totp.login.invalidCode');
          else if (rsp.code === -5) errorMsg = t('totp.login.locked');
          else if (rsp.code === -4) errorMsg = t('auth.globalLocked');

          setMsg(errorMsg);
          return;
        }

        return finishLogin();
      })
      .catch(() => {
        setMsg(t('auth.error'));
      })
      .finally(() => {
        setIsloading(false);
      });
  }

  function backToPassword() {
    setChallenge(null);
    setMsg('');
  }

  return (
    <>
      <Head title={t('head.login')} />

      <div className="flex h-screen w-screen flex-col items-center justify-center">
        {challenge ? (
          <Form key="code" style={{ minWidth: 300, maxWidth: 500 }} onFinish={submitCode}>
            <div className="flex flex-col items-center justify-center pb-4">
              <img id="logo" src="/ironkvm.ico" alt="IronKVM" />
            </div>
            <div className="pb-1 text-base">{t('totp.login.title')}</div>
            <div className="pb-4 text-sm text-neutral-500">{t('totp.login.description')}</div>

            <Form.Item
              name="code"
              rules={[{ required: true, message: t('totp.login.noEmptyCode') }]}
            >
              <Input
                prefix={<KeyRoundIcon size={15} />}
                autoFocus
                autoComplete="one-time-code"
                maxLength={64}
                placeholder={t('totp.login.placeholder')}
              />
            </Form.Item>

            <div className="pb-1 text-red-500">{msg}</div>

            <Form.Item>
              <Button type="primary" htmlType="submit" className="w-full" loading={isLoading}>
                {t('totp.login.verify')}
              </Button>
            </Form.Item>

            <div className="flex justify-end pb-4 text-sm">
              <Button type="link" size="small" className="px-0!" onClick={backToPassword}>
                {t('totp.login.back')}
              </Button>
            </div>
          </Form>
        ) : (
          <Form
            key="password"
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
              <Alert
                className="mb-4"
                type="success"
                showIcon
                message={t('auth.passwordChanged')}
              />
            )}

            {oidc && (
              <div className="pb-4">
                <Button
                  type={passwordFolded ? 'primary' : 'default'}
                  className="w-full"
                  icon={<LogInIcon size={15} />}
                  onClick={signInWithOidc}
                >
                  {t('auth.oidc.signIn', { name: oidc.providerName })}
                </Button>
              </div>
            )}

            {passwordFolded ? (
              <>
                <div className="pb-1 text-red-500">{msg}</div>
                <div className="flex justify-center pb-4 text-sm">
                  <Button type="link" size="small" onClick={() => setShowPassword(true)}>
                    {t('auth.oidc.usePassword')}
                  </Button>
                </div>
              </>
            ) : (
              <>
                <Form.Item
                  name="username"
                  rules={[{ required: true, message: t('auth.noEmptyUsername'), min: 1 }]}
                >
                  <Input
                    prefix={<UserIcon size={15} />}
                    placeholder={t('auth.placeholderUsername')}
                  />
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
              </>
            )}

            <div className="flex justify-end pb-4 text-sm">
              <Tips />
            </div>
          </Form>
        )}
      </div>
    </>
  );
};
