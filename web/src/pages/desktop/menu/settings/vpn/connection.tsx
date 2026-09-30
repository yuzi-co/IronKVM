import { useState } from 'react';
import { Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';

import { SectionHeader } from '../components/section.tsx';
import type { Status, VpnInfo } from './types.ts';
import { isSwitchedOn } from './view.ts';

type ConnectionProps = {
  vpn: VpnInfo;
  status: Status;
  blocked: boolean;
  onChange: () => void;
  onError: (msg: string) => void;
};

// Connection is the page's one switch. On starts the daemon when it is not
// running and joins the network; off leaves the network and stops the
// daemon, which frees its memory. The server does both steps.
export const Connection = ({ vpn, status, blocked, onChange, onError }: ConnectionProps) => {
  const { t } = useTranslation();

  // While a request runs, and until the status it changed arrives, the
  // switch shows where it is going rather than where it was.
  const [pending, setPending] = useState<boolean>();
  const [isLoading, setIsLoading] = useState(false);

  // A new status from the parent ends the wait. Done during render rather
  // than in an effect, so the switch never paints the old value against it.
  const [prevStatus, setPrevStatus] = useState(status);
  if (status !== prevStatus) {
    setPrevStatus(status);
    if (!isLoading) setPending(undefined);
  }

  const checked = pending ?? isSwitchedOn(status.state);

  function toggle(next: boolean) {
    if (isLoading) return;
    setIsLoading(true);
    setPending(next);
    onError('');

    (next ? vpn.api.connect() : vpn.api.disconnect())
      .then((rsp) => {
        if (rsp.code !== 0) {
          setPending(undefined);
          onError(describeFailure(rsp));
        }
      })
      .catch((err) => {
        setPending(undefined);
        onError(describeFailure(err));
      })
      .finally(() => {
        setIsLoading(false);
        onChange();
      });
  }

  return (
    <SectionHeader
      title={t('settings.vpn.connected')}
      description={t('settings.vpn.connectedDesc', { name: vpn.title })}
      action={
        <Switch
          checked={checked}
          loading={isLoading}
          disabled={blocked && !checked}
          onChange={toggle}
        />
      }
    />
  );
};
