import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Divider, Input, message, Modal, Switch, Tag } from 'antd';
import { CheckIcon, CopyIcon, DicesIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/ipmi.ts';
import type { IpmiSettings, IpmiUser } from '@/api/ipmi.ts';
import { writeClipboardText } from '@/lib/clipboard.ts';
import { getHostname } from '@/lib/service.ts';

const minPasswordLength = 12;
const maxPasswordLength = 20;

// Letters and digits without the ones that are easy to misread, plus a few
// symbols that need no quoting in a shell.
const passwordAlphabet = 'ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789-_.+=';

// generatePassword draws a password of the IPMI maximum length. Bytes that
// would bias the draw are thrown away.
function generatePassword() {
  const limit = 256 - (256 % passwordAlphabet.length);
  let out = '';
  while (out.length < maxPasswordLength) {
    const bytes = new Uint8Array(32);
    crypto.getRandomValues(bytes);
    for (const b of bytes) {
      if (b < limit && out.length < maxPasswordLength) {
        out += passwordAlphabet[b % passwordAlphabet.length];
      }
    }
  }
  return out;
}

function passwordError(password: string) {
  if (password.length < minPasswordLength || password.length > maxPasswordLength) {
    return 'settings.ipmi.passwordLength';
  }
  if (!/^[\x20-\x7e]*$/.test(password)) {
    return 'settings.ipmi.passwordChars';
  }
  return '';
}

// The IPMI service: its switch, the warning that goes with it, what the
// power LED allows, and the IPMI password of each account.
export const Ipmi = () => {
  const { t } = useTranslation();
  const [modal, contextHolder] = Modal.useModal();
  const [settings, setSettings] = useState<IpmiSettings | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isCopied, setIsCopied] = useState(false);

  const [editing, setEditing] = useState<IpmiUser | null>(null);
  const [password, setPassword] = useState('');
  const [isPasswordCopied, setIsPasswordCopied] = useState(false);
  const [isSavingPassword, setIsSavingPassword] = useState(false);

  const example = `ipmitool -I lanplus -H ${getHostname()} -U <user> -P <password> power status`;
  const error = password ? passwordError(password) : '';

  const getSettings = useCallback(() => {
    api
      .getIpmiSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(rsp.msg || t('settings.ipmi.failed'));
          return;
        }
        setSettings(rsp.data);
      })
      .catch(() => message.error(t('settings.ipmi.failed')));
  }, [t]);

  useEffect(() => {
    getSettings();
  }, [getSettings]);

  function setEnabled(enabled: boolean) {
    setIsSaving(true);
    api
      .setIpmiEnabled(enabled)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(rsp.msg || t('settings.ipmi.failed'));
          return;
        }
        setSettings((current) => (current ? { ...current, enabled } : current));
      })
      .catch(() => message.error(t('settings.ipmi.failed')))
      .finally(() => setIsSaving(false));
  }

  async function copy(text: string, setCopied: (copied: boolean) => void) {
    try {
      await writeClipboardText(text);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 2000);
    } catch {
      message.error(t('settings.ipmi.copyFailed'));
    }
  }

  function openPassword(user: IpmiUser) {
    setEditing(user);
    setPassword(generatePassword());
    setIsPasswordCopied(false);
  }

  function savePassword() {
    if (!editing || passwordError(password)) return;
    setIsSavingPassword(true);
    api
      .setIpmiPassword(editing.username, password)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(rsp.msg || t('settings.ipmi.failed'));
          return;
        }
        message.success(t('settings.ipmi.saved'));
        setEditing(null);
        setPassword('');
        getSettings();
      })
      .catch(() => message.error(t('settings.ipmi.failed')))
      .finally(() => setIsSavingPassword(false));
  }

  function removePassword(user: IpmiUser) {
    modal.confirm({
      title: t('settings.ipmi.removeConfirmTitle', { user: user.username }),
      content: (
        <span className="text-sm text-neutral-400">{t('settings.ipmi.removeConfirmDesc')}</span>
      ),
      okText: t('settings.ipmi.okBtn'),
      cancelText: t('settings.ipmi.cancelBtn'),
      onOk: async () => {
        try {
          const rsp = await api.clearIpmiPassword(user.username);
          if (rsp.code !== 0) {
            message.error(rsp.msg || t('settings.ipmi.failed'));
          }
        } catch {
          message.error(t('settings.ipmi.failed'));
        } finally {
          getSettings();
        }
      }
    });
  }

  function userState(user: IpmiUser) {
    if (!user.nameFits) return t('settings.ipmi.nameTooLong');
    if (!user.enabled) return t('settings.ipmi.accountDisabled');
    return user.hasPassword ? t('settings.ipmi.passwordSet') : t('settings.ipmi.passwordNotSet');
  }

  return (
    <>
      {contextHolder}
      <div className="text-base">{t('settings.ipmi.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <Alert type="warning" showIcon message={t('settings.ipmi.warning')} />

        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.ipmi.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.ipmi.serviceDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        {settings?.enabled && (
          <div className="flex flex-col space-y-3">
            <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3.5">
              <span className="text-sm font-medium text-neutral-400">
                {t('settings.ipmi.example')}
              </span>
              <div className="flex min-w-0 items-center justify-between gap-2">
                <span className="min-w-0 flex-1 font-mono text-sm break-all text-neutral-300 select-all">
                  {example}
                </span>
                <Button
                  type="text"
                  size="small"
                  className="text-neutral-400 hover:text-white"
                  icon={
                    isCopied ? (
                      <CheckIcon size={15} className="text-green-500" />
                    ) : (
                      <CopyIcon size={15} />
                    )
                  }
                  onClick={() => copy(example, setIsCopied)}
                />
              </div>
            </div>

            {settings.powerLed ? (
              <Alert type="success" showIcon message={t('settings.ipmi.ledOn')} />
            ) : (
              <Alert type="warning" showIcon message={t('settings.ipmi.ledOff')} />
            )}
          </div>
        )}

        <div className="flex flex-col space-y-2">
          <span className="text-sm font-medium">{t('settings.ipmi.accounts')}</span>
          <span className="text-xs text-neutral-500">{t('settings.ipmi.accountsDesc')}</span>

          <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
            {(settings?.users ?? []).map((user, index) => (
              <div
                key={user.username}
                className={`flex items-center justify-between gap-3 px-4 py-3 ${
                  index > 0 ? 'border-t border-neutral-800' : ''
                }`}
              >
                <div className="flex min-w-0 flex-col space-y-0.5">
                  <div className="flex min-w-0 items-center gap-2">
                    <span className="truncate text-sm text-neutral-300">{user.username}</span>
                    <Tag className="me-0">{t(`settings.account.roles.${user.role}`)}</Tag>
                  </div>
                  <span className="text-xs text-neutral-500">{userState(user)}</span>
                </div>
                {user.nameFits && (
                  <div className="flex shrink-0 gap-2">
                    <Button size="small" onClick={() => openPassword(user)}>
                      {user.hasPassword
                        ? t('settings.ipmi.changePassword')
                        : t('settings.ipmi.setPassword')}
                    </Button>
                    {user.hasPassword && (
                      <Button size="small" danger onClick={() => removePassword(user)}>
                        {t('settings.ipmi.remove')}
                      </Button>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      </div>

      <Modal
        title={t('settings.ipmi.passwordTitle', { user: editing?.username ?? '' })}
        open={!!editing}
        destroyOnHidden
        maskClosable={false}
        okText={t('settings.ipmi.save')}
        cancelText={t('settings.ipmi.cancelBtn')}
        okButtonProps={{ disabled: !password || !!error, loading: isSavingPassword }}
        onOk={savePassword}
        onCancel={() => setEditing(null)}
      >
        <div className="flex flex-col space-y-3">
          <span className="text-sm text-neutral-400">{t('settings.ipmi.passwordDesc')}</span>
          <div className="flex items-center gap-2">
            <Input
              className="font-mono"
              autoComplete="off"
              maxLength={maxPasswordLength}
              placeholder={t('settings.ipmi.passwordPlaceholder')}
              status={error ? 'error' : undefined}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <Button
              icon={<DicesIcon size={15} />}
              title={t('settings.ipmi.generate')}
              onClick={() => setPassword(generatePassword())}
            />
            <Button
              icon={
                isPasswordCopied ? (
                  <CheckIcon size={15} className="text-green-500" />
                ) : (
                  <CopyIcon size={15} />
                )
              }
              title={t('settings.ipmi.copy')}
              disabled={!password}
              onClick={() => copy(password, setIsPasswordCopied)}
            />
          </div>
          {error && <span className="text-xs text-red-500">{t(error)}</span>}
        </div>
      </Modal>
    </>
  );
};
