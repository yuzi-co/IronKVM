import { useEffect, useState } from 'react';
import { Divider } from 'antd';
import { LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { Boot } from './boot.tsx';
import { Device } from './device.tsx';
import { ErrorDetail } from './error-detail.tsx';
import { Header } from './header.tsx';
import { Install } from './install.tsx';
import { Notice } from './notice.tsx';
import { Run } from './run.tsx';
import type { Status, VpnInfo } from './types.ts';

type VpnPageProps = {
  vpn: VpnInfo;
  setIsLocked: (isLocked: boolean) => void;
};

// VpnPage is the settings page of Tailscale and of NetBird. Each gives it its
// API module and its login form; the rest is the same for both.
export const VpnPage = ({ vpn, setIsLocked }: VpnPageProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<Status>();
  const [errMsg, setErrMsg] = useState('');

  const getStatus = useStableCallback(() => {
    if (isLoading) return;
    setIsLoading(true);

    vpn.api
      .getStatus()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setErrMsg(rsp.msg);
          return;
        }
        setStatus(rsp.data);
      })
      .catch((err) => {
        setErrMsg(err?.message || 'Failed to get status');
      })
      .finally(() => {
        setIsLoading(false);
      });
  });

  useEffect(() => {
    getStatus();
  }, [getStatus]);

  const blocked = !!status?.blockedBy;
  const installed = !!status && status.state !== 'notInstall';

  return (
    <>
      <Header
        vpn={vpn}
        state={status?.state}
        setIsLocked={setIsLocked}
        onChange={getStatus}
        onError={setErrMsg}
      />
      <Divider className="opacity-50" />

      {isLoading ? (
        <div className="flex w-full items-center justify-center space-x-2 pt-5 text-neutral-500">
          <LoaderCircleIcon className="animate-spin" size={18} />
          <span>{t('settings.vpn.loading')}</span>
        </div>
      ) : (
        <>
          <Notice blockedBy={status?.blockedBy ?? ''} />

          {status?.state === 'notInstall' && (
            <Install
              vpn={vpn}
              blocked={blocked}
              setIsLocked={setIsLocked}
              onSuccess={getStatus}
              onError={setErrMsg}
            />
          )}

          {status?.state === 'notRunning' && (
            <Run vpn={vpn} blocked={blocked} onSuccess={getStatus} onError={setErrMsg} />
          )}

          {status?.state === 'notLogin' && vpn.renderLogin(getStatus)}

          {(status?.state === 'stopped' || status?.state === 'running') && (
            <Device vpn={vpn} status={status} onChange={getStatus} onError={setErrMsg} />
          )}

          {installed && status && (
            <div className="pt-6">
              <Boot
                vpn={vpn}
                enabled={status.bootEnabled}
                blocked={blocked}
                onChange={getStatus}
                onError={setErrMsg}
              />
            </div>
          )}

          <ErrorDetail message={errMsg} />
        </>
      )}
    </>
  );
};
