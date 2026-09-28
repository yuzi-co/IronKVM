import { useEffect, useRef, useState } from 'react';
import { Divider } from 'antd';
import { LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { usePoll } from '../components/use-poll.ts';
import { Boot } from './boot.tsx';
import { Device } from './device.tsx';
import { ErrorDetail } from './error-detail.tsx';
import { Header } from './header.tsx';
import { Install } from './install.tsx';
import { Notice } from './notice.tsx';
import { Run } from './run.tsx';
import type { Status, VpnInfo } from './types.ts';

const statusPollMs = 10 * 1000;

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
  const isPolling = useRef(false);

  // quiet asks without the loading screen: the poll must not blank the page
  // every few seconds, nor wipe an error the operator is reading.
  const fetchStatus = useStableCallback((quiet: boolean) => {
    // A poll never overlaps another request; an action's refresh always runs.
    if (isLoading || (quiet && isPolling.current)) return;
    if (quiet) isPolling.current = true;
    else setIsLoading(true);

    vpn.api
      .getStatus()
      .then((rsp) => {
        if (rsp.code !== 0) {
          if (!quiet) setErrMsg(describeFailure(rsp));
          return;
        }
        setStatus(rsp.data);
      })
      .catch((err) => {
        if (!quiet) setErrMsg(describeFailure(err));
      })
      .finally(() => {
        if (quiet) isPolling.current = false;
        else setIsLoading(false);
      });
  });
  const getStatus = useStableCallback(() => fetchStatus(false));

  useEffect(() => {
    getStatus();
  }, [getStatus]);

  // The address, peers and connection change on their own, so keep them
  // current while the page is open. Not while installing: that has its own
  // progress and holds the page.
  usePoll(() => fetchStatus(true), statusPollMs, !!status && status.state !== 'notInstall');

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
