import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Form, Input, message, Modal, QRCode, Tag } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/auth.ts';
import type { TotpSetup, TotpState } from '@/api/auth.ts';
import { encrypt } from '@/lib/encrypt.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { CopyBlock, CopyRow } from '../components/copy-button.tsx';
import { Section } from '../components/section.tsx';

type Mode = 'enable' | 'disable' | 'regenerate';
type Step = 'password' | 'scan' | 'codes';

type PasswordValues = { password: string };
type CodeValues = { code: string };

// downloadCodes saves the recovery codes as a text file, the one moment they
// can be kept: the device stores only their hashes.
function downloadCodes(codes: string[], title: string) {
  const text = `${title}\n\n${codes.join('\n')}\n`;
  const url = URL.createObjectURL(new Blob([text], { type: 'text/plain' }));
  const link = document.createElement('a');
  link.href = url;
  link.download = `ironkvm-recovery-codes-${window.location.hostname}.txt`;
  link.click();
  URL.revokeObjectURL(url);
}

// TwoFactor turns TOTP on and off for the signed-in account and replaces its
// recovery codes. Every change asks for the password again, so a session left
// open cannot turn the second factor off or read new codes.
export const TwoFactor = () => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [passwordForm] = Form.useForm<PasswordValues>();
  const [codeForm] = Form.useForm<CodeValues>();

  const [state, setState] = useState<TotpState | null>(null);
  const [mode, setMode] = useState<Mode | null>(null);
  const [step, setStep] = useState<Step>('password');
  const [setup, setSetup] = useState<TotpSetup | null>(null);
  const [codes, setCodes] = useState<string[]>([]);
  const [isBusy, setIsBusy] = useState(false);

  const load = useCallback(
    () =>
      api
        .getTotp()
        .then((rsp) => {
          if (rsp.code !== 0) throw rsp;
          setState({ enabled: !!rsp.data?.enabled, recoveryCodes: rsp.data?.recoveryCodes ?? 0 });
        })
        .catch((err) => {
          messageApi.error(describeFailure(err, t('totp.settings.loadFailed')));
        }),
    [messageApi, t]
  );

  useEffect(() => {
    load();
  }, [load]);

  function open(next: Mode) {
    passwordForm.resetFields();
    codeForm.resetFields();
    setSetup(null);
    setCodes([]);
    setStep('password');
    setMode(next);
  }

  function close() {
    // The secret and the codes live only in this state.
    setMode(null);
    setSetup(null);
    setCodes([]);
    load();
  }

  async function submitPassword(values: PasswordValues) {
    setIsBusy(true);
    try {
      const password = encrypt(values.password);
      if (mode === 'enable') {
        const rsp = await api.setupTotp(password);
        if (rsp.code !== 0) throw rsp;
        setSetup(rsp.data as TotpSetup);
        setStep('scan');
      } else if (mode === 'disable') {
        const rsp = await api.disableTotp(password);
        if (rsp.code !== 0) throw rsp;
        messageApi.success(t('totp.settings.disabled'));
        close();
      } else if (mode === 'regenerate') {
        const rsp = await api.regenerateRecoveryCodes(password);
        if (rsp.code !== 0) throw rsp;
        setCodes(rsp.data?.recoveryCodes ?? []);
        setStep('codes');
      }
    } catch (err) {
      const rsp = err as { code?: number };
      messageApi.error(
        rsp?.code === -2
          ? t('totp.settings.wrongPassword')
          : describeFailure(err, t('totp.settings.failed'))
      );
    } finally {
      setIsBusy(false);
    }
  }

  async function submitCode(values: CodeValues) {
    setIsBusy(true);
    try {
      const rsp = await api.enableTotp(values.code.trim());
      if (rsp.code === -3) {
        // The secret expired before it was confirmed: start over.
        messageApi.error(t('totp.settings.setupExpired'));
        setStep('password');
        setSetup(null);
        return;
      }
      if (rsp.code !== 0) throw rsp;
      setCodes(rsp.data?.recoveryCodes ?? []);
      setStep('codes');
    } catch (err) {
      const rsp = err as { code?: number };
      messageApi.error(
        rsp?.code === -2
          ? t('totp.login.invalidCode')
          : describeFailure(err, t('totp.settings.failed'))
      );
    } finally {
      setIsBusy(false);
    }
  }

  const title =
    mode === 'disable'
      ? t('totp.settings.disableTitle')
      : mode === 'regenerate'
        ? t('totp.settings.regenerateTitle')
        : t('totp.settings.enableTitle');

  return (
    <>
      {contextHolder}
      <Section
        title={t('totp.settings.title')}
        description={t('totp.settings.description')}
        action={
          state?.enabled ? (
            <Tag color="green">{t('totp.settings.on')}</Tag>
          ) : (
            <Tag>{t('totp.settings.off')}</Tag>
          )
        }
      >
        <div className="text-xs text-neutral-500">{t('totp.settings.scope')}</div>

        {state && !state.enabled && (
          <div>
            <Button type="primary" onClick={() => open('enable')}>
              {t('totp.settings.enable')}
            </Button>
          </div>
        )}

        {state?.enabled && (
          <>
            <div className="text-sm text-neutral-400">
              {t('totp.settings.codesLeft', { count: state.recoveryCodes })}
            </div>
            <div className="flex flex-wrap gap-2">
              <Button onClick={() => open('regenerate')}>{t('totp.settings.regenerate')}</Button>
              <Button danger onClick={() => open('disable')}>
                {t('totp.settings.disable')}
              </Button>
            </div>
          </>
        )}
      </Section>

      <Modal
        title={title}
        open={mode !== null}
        footer={null}
        destroyOnHidden
        maskClosable={step !== 'codes'}
        onCancel={close}
      >
        {step === 'password' && (
          <Form<PasswordValues> form={passwordForm} layout="vertical" onFinish={submitPassword}>
            <Form.Item
              name="password"
              label={t('totp.settings.confirmPassword')}
              rules={[{ required: true, message: t('auth.noEmptyPassword') }]}
            >
              <Input.Password autoComplete="current-password" autoFocus />
            </Form.Item>
            <Button
              type="primary"
              danger={mode === 'disable'}
              htmlType="submit"
              loading={isBusy}
              className="w-full"
            >
              {mode === 'disable' ? t('totp.settings.disable') : t('totp.settings.continue')}
            </Button>
          </Form>
        )}

        {step === 'scan' && setup && (
          <div className="flex flex-col space-y-4">
            <span className="text-sm text-neutral-400">{t('totp.settings.scan')}</span>
            <div className="flex justify-center">
              {/* Black on white whatever the theme: some scanners do not read inverted codes. */}
              <QRCode value={setup.uri} size={200} color="#000000" bgColor="#ffffff" />
            </div>
            <CopyRow label={t('totp.settings.secret')} value={setup.secret} />
            <Form<CodeValues> form={codeForm} layout="vertical" onFinish={submitCode}>
              <Form.Item
                name="code"
                label={t('totp.settings.enterCode')}
                rules={[{ required: true, message: t('totp.login.noEmptyCode') }]}
              >
                <Input autoComplete="one-time-code" inputMode="numeric" maxLength={8} autoFocus />
              </Form.Item>
              <Button type="primary" htmlType="submit" loading={isBusy} className="w-full">
                {t('totp.settings.verifyAndEnable')}
              </Button>
            </Form>
          </div>
        )}

        {step === 'codes' && (
          <div className="flex flex-col space-y-4">
            <Alert type="warning" showIcon message={t('totp.settings.codesOnce')} />
            <CopyBlock text={codes.join('\n')} copyLabel={t('totp.settings.recoveryCodes')} />
            <div className="flex flex-wrap justify-end gap-2">
              <Button onClick={() => downloadCodes(codes, t('totp.settings.recoveryCodes'))}>
                {t('totp.settings.download')}
              </Button>
              <Button type="primary" onClick={close}>
                {t('totp.settings.saved')}
              </Button>
            </div>
          </div>
        )}
      </Modal>
    </>
  );
};
