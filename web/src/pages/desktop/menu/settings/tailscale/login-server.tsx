import { useState } from 'react';
import { Button, Input, message } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/extensions/tailscale.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { SectionHeader } from '../components/section.tsx';
import { ErrorDetail } from '../vpn/error-detail.tsx';
import { LoginUrl } from '../vpn/login-url.tsx';
import type { Status } from '../vpn/types.ts';
import { isLoginServer, sameLoginServer } from './login-server.ts';

type LoginServerProps = {
  status: Status;
  blocked: boolean;
  onChange: () => void;
};

// LoginServer is the control server Tailscale logs in to: Tailscale's own when
// empty, or a Headscale server. Changing it needs a new login, so it is locked
// while the device is logged in. The device puts the old server back when the
// new one does not answer.
export const LoginServer = ({ status, blocked, onChange }: LoginServerProps) => {
  const { t } = useTranslation();

  const current = status.loginServer ?? '';
  const [value, setValue] = useState(current);
  const [isSaving, setIsSaving] = useState(false);
  const [loginUrl, setLoginUrl] = useState('');
  const [errMsg, setErrMsg] = useState('');

  // Follow the device when what it uses changes, as after a save or a
  // rollback. Adjusted during render rather than in an effect, so the field
  // never shows the old value for a frame.
  const [seen, setSeen] = useState(current);
  if (current !== seen) {
    setSeen(current);
    setValue(current);
  }

  const loggedIn = status.state === 'running' || status.state === 'stopped';
  const valid = isLoginServer(value);
  const changed = !sameLoginServer(value, current);

  function save() {
    if (isSaving || !valid || !changed) return;
    setIsSaving(true);
    setErrMsg('');

    api
      .setLoginServer(value.trim())
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(describeFailure(rsp));
          return;
        }
        message.success(t('settings.tailscale.loginServerSaved'));
        if (rsp.data?.url) setLoginUrl(rsp.data.url);
        onChange();
      })
      .catch((err) => setErrMsg(describeFailure(err)))
      .finally(() => setIsSaving(false));
  }

  return (
    <div className="flex flex-col space-y-3">
      <SectionHeader
        title={t('settings.tailscale.loginServer')}
        description={t('settings.tailscale.loginServerDesc')}
      />

      <div className="flex items-center gap-2">
        <Input
          value={value}
          placeholder={t('settings.tailscale.loginServerDefault')}
          status={valid ? undefined : 'error'}
          disabled={loggedIn || blocked || isSaving}
          onChange={(e) => setValue(e.target.value)}
          onPressEnter={save}
          allowClear
        />
        <Button
          type="primary"
          loading={isSaving}
          disabled={loggedIn || blocked || !valid || !changed}
          onClick={save}
        >
          {t('common.save')}
        </Button>
      </div>

      {!valid && (
        <span className="text-xs text-red-400">{t('settings.tailscale.loginServerInvalid')}</span>
      )}
      {loggedIn && (
        <span className="text-xs text-neutral-500">
          {t('settings.tailscale.loginServerLocked')}
        </span>
      )}

      {loginUrl && (
        <LoginUrl
          url={loginUrl}
          period={t('settings.tailscale.urlPeriod')}
          getStatus={api.getStatus}
          onSuccess={() => {
            setLoginUrl('');
            onChange();
          }}
        />
      )}

      <ErrorDetail message={errMsg} />
    </div>
  );
};
