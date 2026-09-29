import { useCallback, useEffect, useState } from 'react';
import { Alert, Divider, message, Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { describeFailure, showFailure, showResult } from '@/lib/feedback.ts';

import { StatusTag } from '../components/status-tag.tsx';
import { AuthorizedKeys } from './authorized-keys.tsx';
import { Connection } from './connection.tsx';
import { HostKeys } from './host-keys.tsx';
import { KeysOnly } from './keys-only.tsx';
import type { SshKey, SshState } from './types.ts';

// The SSH page: the switch and what sshd is doing, how to connect, the host
// key fingerprints, root's authorized keys and whether passwords still work.
export const Ssh = () => {
  const { t } = useTranslation();
  const [state, setState] = useState<SshState | null>(null);
  const [keys, setKeys] = useState<SshKey[] | null>(null);
  const [isSwitching, setIsSwitching] = useState(false);

  const failed = useCallback(
    (cause: unknown) => message.error(describeFailure(cause, t('settings.ssh.failed'))),
    [t]
  );

  const loadState = useCallback(
    () =>
      api
        .getSSHState()
        .then((rsp) => {
          if (rsp.code !== 0) failed(rsp);
          else setState(rsp.data);
        })
        .catch(failed),
    [failed]
  );

  const loadKeys = useCallback(
    () =>
      api
        .getSSHKeys()
        .then((rsp) => {
          if (rsp.code !== 0) failed(rsp);
          else setKeys(rsp.data?.keys ?? []);
        })
        .catch(failed),
    [failed]
  );

  // Keys and state move together: the key count gates the keys-only switch.
  const reload = useCallback(() => Promise.all([loadState(), loadKeys()]), [loadState, loadKeys]);

  useEffect(() => {
    reload();
  }, [reload]);

  async function toggle(enable: boolean) {
    setIsSwitching(true);
    try {
      const rsp = enable ? await api.enableSSH() : await api.disableSSH();
      const success = t(enable ? 'feedback.enabled' : 'feedback.disabled', { name: 'SSH' });
      showResult(rsp, { success });
    } catch (err) {
      showFailure(err);
    } finally {
      await loadState();
      setIsSwitching(false);
    }
  }

  const rootWarning =
    state?.rootPassword === 'default'
      ? t('settings.ssh.rootDefault')
      : state?.rootPassword === 'empty'
        ? t('settings.ssh.rootEmpty')
        : '';

  return (
    <>
      <div className="flex items-center justify-between">
        <span className="text-base">SSH</span>
        {state && <StatusTag running={state.running} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.ssh.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.ssh.serviceDesc')}</span>
          </div>
          <Switch
            checked={state?.enabled ?? false}
            loading={!state || isSwitching}
            onChange={toggle}
          />
        </div>

        {rootWarning && (
          <Alert
            type="warning"
            showIcon
            message={rootWarning}
            description={t('settings.ssh.rootWarning', {
              account: t('settings.account.title'),
              password: t('settings.account.password')
            })}
          />
        )}

        {state && (
          <>
            <Connection state={state} onChange={loadState} />
            <HostKeys keys={state.hostKeys ?? []} />
            <AuthorizedKeys keys={keys} keysOnly={state.keysOnly} onChange={reload} />
            <KeysOnly state={state} keyCount={keys?.length ?? state.keyCount} onChange={reload} />
          </>
        )}
      </div>
    </>
  );
};
