import { useState } from 'react';
import { Button, Divider, Popconfirm, Switch } from 'antd';
import { LogOutIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';

import { Details } from './details.tsx';
import { MemoryBars } from './memory.tsx';
import { Peers } from './peers.tsx';
import type { Status, VpnInfo } from './types.ts';

type DeviceProps = {
  vpn: VpnInfo;
  status: Status;
  onChange: () => void;
  onError: (msg: string) => void;
};

export const Device = ({ vpn, status, onChange, onError }: DeviceProps) => {
  const { t } = useTranslation();

  const [isRunning, setIsRunning] = useState(status.state === 'running');
  const [isUpdating, setIsUpdating] = useState(false);
  const [isLogging, setIsLogging] = useState(false);

  // A new status from the parent replaces whatever the switch last set. This
  // is done during render rather than in an effect, so the switch never paints
  // the old value against the new status.
  const [prevStatus, setPrevStatus] = useState(status);
  if (status !== prevStatus) {
    setPrevStatus(status);
    setIsRunning(status.state === 'running');
  }

  async function toggle() {
    if (isUpdating) return;
    setIsUpdating(true);
    onError('');

    try {
      const rsp = isRunning ? await vpn.api.down() : await vpn.api.up();
      if (rsp.code !== 0) {
        onError(describeFailure(rsp));
        return;
      }
      setIsRunning(!isRunning);
      // Up and down change the peers, the address and the connection, which
      // the page only learns by asking again.
      onChange();
    } catch (err) {
      onError(describeFailure(err));
    } finally {
      setIsUpdating(false);
    }
  }

  function logout() {
    if (isLogging) return;
    setIsLogging(true);
    onError('');

    vpn.api
      .logout()
      .then((rsp) => {
        if (rsp.code !== 0) {
          onError(describeFailure(rsp));
          return;
        }
        onChange();
      })
      .catch((err) => onError(describeFailure(err)))
      .finally(() => setIsLogging(false));
  }

  return (
    <div className="flex flex-col space-y-6 pt-5">
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1 pr-4">
          <span>{t('settings.vpn.connect')}</span>
          <span className="text-xs text-neutral-500">
            {t('settings.vpn.connectDesc', { name: vpn.title })}
          </span>
        </div>
        <Switch checked={isRunning} loading={isUpdating} onClick={toggle} />
      </div>

      <Details status={status} />
      <Divider className="my-0" />
      <Peers peers={status.peers ?? []} />
      <Divider className="my-0" />
      <MemoryBars memory={status.memory} />
      <Divider className="my-0" />

      <div className="flex justify-center pt-3">
        <Popconfirm
          placement="bottom"
          title={<div className="max-w-[320px]">{vpn.logoutWarning}</div>}
          okText={t('settings.vpn.okBtn')}
          cancelText={t('settings.vpn.cancelBtn')}
          onConfirm={logout}
        >
          <Button
            danger
            type="primary"
            size="large"
            shape="round"
            icon={<LogOutIcon size={15} />}
            loading={isLogging}
          >
            {vpn.logoutLabel}
          </Button>
        </Popconfirm>
      </div>
    </div>
  );
};
