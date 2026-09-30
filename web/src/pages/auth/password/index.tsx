import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Button, Card, Form, Input } from 'antd';
import { LockIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';
import { useNavigate } from 'react-router';

import * as api from '@/api/auth.ts';
import { notifyAuthExpired } from '@/lib/auth-events.ts';
import { encrypt } from '@/lib/encrypt.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { Head } from '@/components/head.tsx';

// This code is specific to POST /api/auth/password. The backend uses it when
// the authenticated user cannot verify their current password.
const invalidCurrentPasswordCode = -3;

export const Password = () => {
  const { t } = useTranslation();
  const [msg, setMsg] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const navigate = useNavigate();
  const { account } = useAuth();

  useEffect(() => {
    if (msg) {
      setTimeout(() => setMsg(''), 3000);
    }
  }, [msg]);

  function changePassword(values: any) {
    if (isLoading) return;
    if (values.password !== values.password2) {
      setMsg(t('auth.differentPassword'));
      return;
    }
    const currentPassword = encrypt(values.currentPassword);
    const password = encrypt(values.password);

    setIsLoading(true);
    api
      .changePassword(currentPassword, password)
      .then((rsp: any) => {
        if (rsp.code !== 0) {
          // -4 (sign-in unavailable) and -5 (the new password was refused)
          // carry the server's reason, which is the only thing that says what
          // to change.
          setMsg(
            rsp.code === invalidCurrentPasswordCode
              ? t('auth.invalidCurrentPassword')
              : describeFailure(rsp, t('auth.error'))
          );
          return;
        }

        notifyAuthExpired();
        navigate('/auth/login', { replace: true, state: { notice: 'passwordChanged' } });
      })
      .catch((err) => {
        setMsg(describeFailure(err, t('auth.error')));
      })
      .finally(() => setIsLoading(false));
  }

  function cancel() {
    window.location.replace('/');
  }

  return (
    <>
      <Head title={t('head.changePassword')} />

      <div className="flex h-screen w-screen flex-col items-center justify-center space-y-5">
        <h2 className="text-xl font-semibold text-neutral-100">{t('auth.changePassword')}</h2>

        <Form
          style={{ minWidth: 300, maxWidth: 500 }}
          initialValues={{ remember: true }}
          onFinish={changePassword}
        >
          <Form.Item
            name="currentPassword"
            rules={[{ required: true, message: t('auth.noEmptyPassword') }]}
          >
            <Input
              prefix={<LockIcon size={15} />}
              type="password"
              autoComplete="current-password"
              placeholder={t('auth.placeholderCurrentPassword')}
            />
          </Form.Item>

          <Form.Item
            name="password"
            rules={[
              { required: true, message: t('auth.noEmptyPassword') },
              { min: 8, max: 72, message: t('auth.passwordLength') }
            ]}
          >
            <Input
              prefix={<LockIcon size={15} />}
              type="password"
              autoComplete="new-password"
              placeholder={t('auth.placeholderPassword')}
            />
          </Form.Item>

          <Form.Item
            name="password2"
            rules={[
              { required: true, message: t('auth.noEmptyPassword') },
              { min: 8, max: 72, message: t('auth.passwordLength') }
            ]}
          >
            <Input
              prefix={<LockIcon size={15} />}
              type="password"
              autoComplete="new-password"
              placeholder={t('auth.placeholderPassword2')}
            />
          </Form.Item>

          <span className="text-red-500">{msg}</span>
          <Form.Item>
            <div className="flex w-full space-x-2">
              <Button type="primary" htmlType="submit" className="w-1/2" loading={isLoading}>
                {t('auth.ok')}
              </Button>
              <Button className="w-1/2" onClick={cancel}>
                {t('auth.cancel')}
              </Button>
            </div>
          </Form.Item>
        </Form>

        {/* Only the device owner's password change also sets root's, which
            is the SSH and console login. Any other account changes its web
            password alone, and the card would say otherwise. */}
        {account.systemAccount && (
          <Card>
            <div className="flex w-full max-w-[450px] flex-col">
              <div>{t('auth.tips.change1')}</div>
              <ul className="list-outside list-decimal">
                <li>{t('auth.tips.change2')}</li>
                <li>{t('auth.tips.change3')}</li>
              </ul>
              <div className="text-red-500">{t('auth.tips.change4')}</div>
            </div>
          </Card>
        )}
      </div>
    </>
  );
};
