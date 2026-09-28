import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Alert, Button, Divider, Input, message, Modal, Switch, Tag, Tooltip } from 'antd';
import { CheckIcon, CopyIcon, DicesIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/ipmi.ts';
import type { IpmiSettings, IpmiUser } from '@/api/ipmi.ts';
import { writeClipboardText } from '@/lib/clipboard.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getHostname } from '@/lib/service.ts';

import { PowerLedSetting } from '../../power/power-led-setting.tsx';
import { CopyBlock } from '../components/copy-button.tsx';
import { StatusTag } from '../components/status-tag.tsx';

// The port ipmitool assumes when -p is not given.
const defaultPort = 623;

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
  const { account } = useAuth();
  const [modal, contextHolder] = Modal.useModal();
  const [settings, setSettings] = useState<IpmiSettings | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const [editing, setEditing] = useState<IpmiUser | null>(null);
  const [password, setPassword] = useState('');
  const [isPasswordCopied, setIsPasswordCopied] = useState(false);
  const [isSavingPassword, setIsSavingPassword] = useState(false);

  const port = settings?.port ?? defaultPort;
  const portFlag = port !== defaultPort ? ` -p ${port}` : '';
  const example = `ipmitool -I lanplus -H ${getHostname()}${portFlag} -U ${account.username} -P <password> power status`;
  // An account IPMI can log in with: enabled, a name that fits, a password.
  const canLogIn = (settings?.users ?? []).some((u) => u.enabled && u.nameFits && u.hasPassword);
  const error = password ? passwordError(password) : '';

  const getSettings = useCallback(() => {
    api
      .getIpmiSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.ipmi.failed')));
          return;
        }
        setSettings(rsp.data);
      })
      .catch((err) => message.error(describeFailure(err, t('settings.ipmi.failed'))));
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
          message.error(describeFailure(rsp, t('settings.ipmi.failed')));
          return;
        }
        setSettings((current) => (current ? { ...current, enabled } : current));
        message.success(t(enabled ? 'feedback.enabled' : 'feedback.disabled', { name: 'IPMI' }));
      })
      .catch((err) => message.error(describeFailure(err, t('settings.ipmi.failed'))))
      .finally(() => setIsSaving(false));
  }

  // The copied mark stays until the password changes: it is what tells the
  // operator the password they are about to save is on the clipboard.
  async function copyPassword() {
    try {
      await writeClipboardText(password);
      setIsPasswordCopied(true);
    } catch {
      message.error(t('settings.ipmi.copyFailed'));
    }
  }

  function changePassword(value: string) {
    setPassword(value);
    setIsPasswordCopied(false);
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
          message.error(describeFailure(rsp, t('settings.ipmi.failed')));
          return;
        }
        message.success(t('settings.ipmi.saved'));
        setEditing(null);
        setPassword('');
        getSettings();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.ipmi.failed'))))
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
            message.error(describeFailure(rsp, t('settings.ipmi.failed')));
          }
        } catch (err) {
          message.error(describeFailure(err, t('settings.ipmi.failed')));
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
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.ipmi.title')}</span>
        {settings && <StatusTag running={settings.enabled} />}
      </div>
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

        <PowerLedSetting />

        {settings?.enabled && (
          <div className="flex flex-col space-y-3">
            {!canLogIn && <Alert type="error" showIcon message={t('settings.ipmi.noLogin')} />}

            <CopyBlock title={t('settings.ipmi.example')} text={example} />

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
              onChange={(e) => changePassword(e.target.value)}
            />
            <Tooltip title={t('settings.ipmi.generate')}>
              <Button
                icon={<DicesIcon size={15} />}
                aria-label={t('settings.ipmi.generate')}
                onClick={() => changePassword(generatePassword())}
              />
            </Tooltip>
            <Button
              type={isPasswordCopied ? 'default' : 'primary'}
              icon={
                isPasswordCopied ? (
                  <CheckIcon size={15} className="text-green-500" />
                ) : (
                  <CopyIcon size={15} />
                )
              }
              disabled={!password}
              onClick={copyPassword}
            >
              {isPasswordCopied ? t('common.copied') : t('settings.ipmi.copy')}
            </Button>
          </div>
          {error && <span className="text-xs text-red-500">{t(error)}</span>}
          <Alert
            type={isPasswordCopied ? 'success' : 'warning'}
            showIcon
            message={t('settings.ipmi.copyBeforeSave')}
          />
        </div>
      </Modal>
    </>
  );
};
