import { useEffect, useRef, useState } from 'react';
import { Divider } from 'antd';
import { LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { usePoll } from '../components/use-poll.ts';
import { Boot } from './boot.tsx';
import { Connection } from './connection.tsx';
import { Device } from './device.tsx';
import { ErrorDetail } from './error-detail.tsx';
import { Header } from './header.tsx';
import { Install } from './install.tsx';
import { Notice } from './notice.tsx';
import { Peers } from './peers.tsx';
import type { Status, VpnInfo } from './types.ts';
import { hasSettled } from './view.ts';

const statusPollMs = 10 * 1000;

// After the switch is turned on, the status is asked for this often, for at
// most this long, until the address and a peer online show up. The daemon
// answers connect before its peers are known, and the normal poll would leave
// "0 online of 0" on the page for up to its whole interval.
const settlePollMs = 1500;
const settleForMs = 20 * 1000;

type VpnPageProps = {
  vpn: VpnInfo;
  setIsLocked: (isLocked: boolean) => void;
};

// How a status request is made. load blanks the page with a spinner, as on
// the first visit; refresh follows an action and always runs; poll is the
// timer's, which never overlaps another request and never shows an error
// over one the operator is reading.
type Fetch = 'load' | 'refresh' | 'poll';

// VpnPage is the settings page of Tailscale and of NetBird. Each gives it its
// API module and its login form; the rest is the same for both.
export const VpnPage = ({ vpn, setIsLocked }: VpnPageProps) => {
  const { t } = useTranslation();

  const [isLoading, setIsLoading] = useState(false);
  const [status, setStatus] = useState<Status>();
  const [failed, setFailed] = useState(false);
  const [errMsg, setErrMsg] = useState('');
  const isPolling = useRef(false);

  const fetchStatus = useStableCallback((how: Fetch) => {
    if (isLoading || (how === 'poll' && isPolling.current)) return;
    if (how === 'poll') isPolling.current = true;
    if (how === 'load') setIsLoading(true);

    const fail = (msg: string) => {
      setFailed(true);
      if (how !== 'poll') setErrMsg(msg);
    };

    vpn.api
      .getStatus()
      .then((rsp) => {
        if (rsp.code !== 0) {
          fail(describeFailure(rsp));
          return;
        }
        setFailed(false);
        setStatus(rsp.data);
      })
      .catch((err) => fail(describeFailure(err)))
      .finally(() => {
        if (how === 'poll') isPolling.current = false;
        if (how === 'load') setIsLoading(false);
      });
  });
  const getStatus = useStableCallback(() => fetchStatus('load'));
  const refresh = useStableCallback(() => fetchStatus('refresh'));

  useEffect(() => {
    getStatus();
  }, [getStatus]);

  // The address, peers and connection change on their own, so keep them
  // current while the page is open. Not while installing: that has its own
  // progress and holds the page.
  usePoll(() => fetchStatus('poll'), statusPollMs, !!status && status.state !== 'notInstall');

  // When the settling after a connect ends, or 0 when there is none.
  const [settleUntil, setSettleUntil] = useState(0);
  const onConnected = useStableCallback(() => setSettleUntil(Date.now() + settleForMs));
  usePoll(
    () => {
      if (
        Date.now() >= settleUntil ||
        (!failed && hasSettled(status?.state, status?.ip ?? '', status?.peers ?? null))
      ) {
        setSettleUntil(0);
        return;
      }
      fetchStatus('poll');
    },
    settlePollMs,
    settleUntil > 0
  );

  const blocked = !!status?.blockedBy;
  const state = status?.state;
  const installed = !!status && state !== 'notInstall';

  return (
    <>
      <Header
        vpn={vpn}
        state={state}
        failed={failed}
        setIsLocked={setIsLocked}
        onChange={refresh}
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

          {state === 'notInstall' && (
            <Install
              vpn={vpn}
              blocked={blocked}
              setIsLocked={setIsLocked}
              onSuccess={getStatus}
              onError={setErrMsg}
            />
          )}

          {installed && status && (
            <div className="flex flex-col space-y-6 pt-5">
              <Connection
                vpn={vpn}
                status={status}
                blocked={blocked}
                onChange={refresh}
                onConnected={onConnected}
                onError={setErrMsg}
              />
              <Boot
                vpn={vpn}
                enabled={status.bootEnabled}
                blocked={blocked}
                onChange={refresh}
                onError={setErrMsg}
              />

              {state === 'notLogin' && vpn.renderLogin(refresh)}

              {(state === 'stopped' || state === 'running') && (
                <>
                  <Divider className="my-0!" />
                  <Device status={status} />
                  <Peers peers={status.peers ?? []} />
                </>
              )}
            </div>
          )}

          <ErrorDetail message={errMsg} />
        </>
      )}
    </>
  );
};
