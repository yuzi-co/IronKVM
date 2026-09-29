import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Divider, message, Modal, Switch, Tag } from 'antd';
import { RefreshCwIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/netboot.ts';
import type { NetbootStatus } from '@/api/netboot.ts';
import { UPDATE_PATHS } from '@/api/updates.ts';
import { getUsbNetwork } from '@/api/virtual-device.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { UpstreamUpdate } from '@/components/upstream-update.tsx';

import { CopyRow } from '../components/copy-button.tsx';
import { usePoll } from '../components/use-poll.ts';
import { ErrorDetail } from '../vpn/error-detail.tsx';

// Leases, boots and the log change while a host boots, which takes seconds.
const statusPollMs = 5 * 1000;

type NetbootProps = {
  setIsLocked: (isLocked: boolean) => void;
};

function formatTime(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

// Network boot of the host: the add-on (dnsmasq and the boot files), the USB
// link side, proxy DHCP on the LAN, and what the host fetched.
export const Netboot = ({ setIsLocked }: NetbootProps) => {
  const { t } = useTranslation();
  const [modal, contextHolder] = Modal.useModal();

  const [status, setStatus] = useState<NetbootStatus | null>(null);
  const [busy, setBusy] = useState<'' | 'install' | 'uninstall' | 'usb' | 'lan'>('');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState('');
  // Why the USB network link cannot be turned on, when the board refuses it.
  const [linkRefusal, setLinkRefusal] = useState('');

  // load clears the loading flag; refresh is what sets it. A quiet load is
  // the poll, which says nothing when it fails: the next one tries again.
  const load = useCallback(
    (quiet = false) => {
      api
        .getNetbootStatus()
        .then((rsp) => {
          if (rsp.code !== 0) {
            if (!quiet) message.error(describeFailure(rsp, t('settings.netboot.failed')));
            return;
          }
          setStatus(rsp.data);
        })
        .catch((err) => {
          if (!quiet) message.error(describeFailure(err, t('settings.netboot.failed')));
        })
        .finally(() => setIsLoading(false));
    },
    [t]
  );

  useEffect(() => {
    load();
  }, [load]);

  usePoll(
    () => {
      if (!busy) load(true);
    },
    statusPollMs,
    !!status && (status.usb || status.lan)
  );

  function refresh() {
    setIsLoading(true);
    load();
  }

  // run sends one change and shows the state the server answers with.
  function run(kind: typeof busy, request: () => ReturnType<typeof api.getNetbootStatus>) {
    setBusy(kind);
    setError('');
    if (kind === 'install') setIsLocked(true);

    request()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(describeFailure(rsp, t('settings.netboot.failed')));
          load();
          return;
        }
        setStatus(rsp.data);
      })
      .catch((err) => {
        // The change may have landed before the answer was lost, so the page
        // asks for the state rather than keep showing the old one.
        setError(describeFailure(err, t('settings.netboot.failed')));
        load();
      })
      .finally(() => {
        setBusy('');
        if (kind === 'install') setIsLocked(false);
      });
  }

  function setUsb(usb: boolean) {
    if (!status) return;
    run('usb', () => api.setNetbootSettings(usb, status.lan));
  }

  function setLan(lan: boolean) {
    if (!status) return;
    if (!lan) {
      run('lan', () => api.setNetbootSettings(status.usb, false));
      return;
    }

    modal.confirm({
      title: t('settings.netboot.lanConfirm'),
      content: <span className="text-sm text-neutral-400">{t('settings.netboot.lanWarning')}</span>,
      okText: t('settings.netboot.okBtn'),
      cancelText: t('settings.netboot.cancelBtn'),
      onOk: () => run('lan', () => api.setNetbootSettings(status.usb, true))
    });
  }

  function uninstall() {
    modal.confirm({
      title: t('settings.netboot.uninstallConfirm'),
      okText: t('settings.netboot.okBtn'),
      cancelText: t('settings.netboot.cancelBtn'),
      okButtonProps: { danger: true },
      onOk: () => run('uninstall', api.uninstallNetboot)
    });
  }

  const runningTag = (running: boolean) =>
    running ? (
      <Tag color="green">{t('settings.netboot.running')}</Tag>
    ) : (
      <Tag>{t('settings.netboot.stopped')}</Tag>
    );

  const linkOn = !!status && status.link.mode !== 'off' && status.link.mode !== '';
  const needsLink = !!status?.usb && !linkOn;

  // With the link off, ask the board whether it could be turned on. If not,
  // its reason (the USB endpoints) is what the operator has to fix first.
  useEffect(() => {
    if (!needsLink) return;
    let active = true;
    getUsbNetwork()
      .then((rsp) => {
        if (active && rsp.code === 0 && rsp.data.mode === 'off' && !rsp.data.fits) {
          setLinkRefusal(rsp.data.refusal || '');
        }
      })
      .catch(() => {});
    return () => {
      active = false;
      setLinkRefusal('');
    };
  }, [needsLink]);
  const images = status?.images ?? [];
  const leases = status?.leases ?? [];
  const boots = status?.boots ?? [];

  return (
    <>
      {contextHolder}
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.netboot.title')}</span>
        <Button
          type="text"
          size="small"
          className="text-neutral-400 hover:text-white"
          loading={isLoading}
          icon={<RefreshCwIcon size={15} />}
          title={t('settings.netboot.refresh')}
          onClick={refresh}
        />
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <span className="text-xs text-neutral-500">{t('settings.netboot.description')}</span>

        {/* The add-on */}
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.netboot.addon')}</span>
            <span className="text-xs text-neutral-500">{t('settings.netboot.addonDesc')}</span>
            {status?.installed && status.version && (
              <span className="font-mono text-xs text-neutral-500">dnsmasq {status.version}</span>
            )}
          </div>
          {status?.installed ? (
            <Button danger loading={busy === 'uninstall'} disabled={!!busy} onClick={uninstall}>
              {t('settings.netboot.uninstall')}
            </Button>
          ) : (
            <Button
              type="primary"
              loading={busy === 'install'}
              disabled={!status || !status.onData || !!busy}
              onClick={() => run('install', api.installNetboot)}
            >
              {t('settings.netboot.install')}
            </Button>
          )}
        </div>
        {status?.installed && (
          <UpstreamUpdate
            path={UPDATE_PATHS.netbootxyz}
            name="netboot.xyz"
            onUpdated={() => load(true)}
          />
        )}
        {status && !status.onData && !status.installed && (
          <Alert type="info" showIcon message={t('settings.netboot.needsData')} />
        )}
        {busy === 'install' && (
          <Alert type="info" showIcon message={t('settings.netboot.installing')} />
        )}
        <ErrorDetail message={error} />

        {/* The USB link */}
        <div className="flex flex-col space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex flex-col space-y-1 pr-4">
              <span className="text-sm font-medium">{t('settings.netboot.usb')}</span>
              <span className="text-xs text-neutral-500">{t('settings.netboot.usbDesc')}</span>
            </div>
            <Switch
              checked={status?.usb ?? false}
              loading={!status || busy === 'usb'}
              disabled={!status?.installed || (!!busy && busy !== 'usb')}
              onChange={setUsb}
            />
          </div>

          {needsLink && (
            <Alert
              type="warning"
              showIcon
              message={t('settings.netboot.linkOff')}
              description={linkRefusal || undefined}
            />
          )}

          {status?.usb && linkOn && (
            <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3.5 text-sm">
              <div className="flex items-center justify-between">
                <span className="text-neutral-400">dnsmasq</span>
                {runningTag(status.usbRunning)}
              </div>
              <CopyRow
                label={t('settings.netboot.menuUrl')}
                value={status.menuUrl}
                display={status.menuUrl || '-'}
              />
              <div className="flex flex-wrap items-center justify-between gap-x-3">
                <span className="text-neutral-400">{t('settings.netboot.leases')}</span>
                <span className="min-w-0 font-mono text-xs break-all text-neutral-300">
                  {leases.length === 0
                    ? t('settings.netboot.noLeases')
                    : leases
                        .map((lease) =>
                          [lease.ip, lease.mac, lease.hostname].filter(Boolean).join(' ')
                        )
                        .join(', ')}
                </span>
              </div>
            </div>
          )}

          {status?.usb && (
            <span className="text-xs text-neutral-500">{t('settings.netboot.netbootxyzNote')}</span>
          )}
        </div>

        {/* The LAN */}
        <div className="flex flex-col space-y-3">
          <div className="flex items-center justify-between">
            <div className="flex flex-col space-y-1 pr-4">
              <span className="text-sm font-medium">{t('settings.netboot.lan')}</span>
              <span className="text-xs text-neutral-500">{t('settings.netboot.lanDesc')}</span>
            </div>
            <Switch
              checked={status?.lan ?? false}
              loading={!status || busy === 'lan'}
              disabled={!status?.installed || (!!busy && busy !== 'lan')}
              onChange={setLan}
            />
          </div>

          {status?.lan && (
            <>
              <Alert type="warning" showIcon message={t('settings.netboot.lanWarning')} />
              <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3.5 text-sm">
                <div className="flex items-center justify-between">
                  <span className="text-neutral-400">dnsmasq</span>
                  {runningTag(status.lanRunning)}
                </div>
                <div className="flex items-center justify-between">
                  <span className="text-neutral-400">{t('settings.netboot.lanInterface')}</span>
                  <span className="font-mono text-xs text-neutral-300">
                    {status.lanInterface ? `${status.lanInterface} ${status.lanNetwork}` : '-'}
                  </span>
                </div>
              </div>
              {status.lanError && <Alert type="error" showIcon message={status.lanError} />}
            </>
          )}
        </div>

        {/* What the host sees and fetched */}
        {status?.installed && (
          <>
            <div className="flex flex-col space-y-2">
              <span className="text-sm font-medium">{t('settings.netboot.images')}</span>
              {images.length === 0 ? (
                <span className="text-xs text-neutral-500">{t('settings.netboot.noImages')}</span>
              ) : (
                <div className="flex flex-wrap gap-y-2">
                  {images.map((image) => (
                    <Tag key={image} className="font-mono">
                      {image}
                    </Tag>
                  ))}
                </div>
              )}
            </div>

            <div className="flex flex-col space-y-2">
              <span className="text-sm font-medium">{t('settings.netboot.boots')}</span>
              {boots.length === 0 ? (
                <span className="text-xs text-neutral-500">{t('settings.netboot.noBoots')}</span>
              ) : (
                <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
                  {boots.map((boot, index) => (
                    <div
                      key={`${boot.time}-${index}`}
                      className={
                        index > 0
                          ? 'flex flex-wrap justify-between gap-x-3 border-t border-neutral-800 px-4 py-2 text-xs'
                          : 'flex flex-wrap justify-between gap-x-3 px-4 py-2 text-xs'
                      }
                    >
                      <span className="min-w-0 font-mono break-all text-neutral-300">
                        {boot.what}
                      </span>
                      <span className="text-neutral-500">
                        {boot.client} · {formatTime(boot.time)}
                      </span>
                    </div>
                  ))}
                </div>
              )}
            </div>

            {[...(status.usbLog ?? []), ...(status.lanLog ?? [])].length > 0 && (
              <div className="flex flex-col space-y-2">
                <span className="text-sm font-medium">{t('settings.netboot.log')}</span>
                <pre className="max-h-[200px] w-full overflow-auto rounded bg-neutral-800/60 p-3 font-mono text-xs break-words whitespace-pre-wrap text-neutral-400">
                  {[...(status.usbLog ?? []), ...(status.lanLog ?? [])].join('\n')}
                </pre>
              </div>
            )}
          </>
        )}
      </div>
    </>
  );
};
