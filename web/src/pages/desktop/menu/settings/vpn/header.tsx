import { useEffect, useState } from 'react';
import { Popconfirm, Popover } from 'antd';
import {
  CircleArrowUpIcon,
  CircleStopIcon,
  EllipsisIcon,
  LoaderIcon,
  RotateCwIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';
import semver from 'semver';

import { Swap } from './swap.tsx';
import type { Rsp, State, UpdateInfo, VpnInfo } from './types.ts';
import { Uninstall } from './uninstall.tsx';

type HeaderProps = {
  vpn: VpnInfo;
  state: State | undefined;
  setIsLocked: (isLocked: boolean) => void;
  onChange: () => void;
  onError: (msg: string) => void;
};

type Loading = '' | 'restarting' | 'stopping' | 'updating';

function isNewer(latest: string, current: string) {
  if (!latest || !current) return false;
  if (semver.valid(latest) && semver.valid(current)) return semver.gt(latest, current);
  return latest !== current;
}

export const Header = ({ vpn, state, setIsLocked, onChange, onError }: HeaderProps) => {
  const { t } = useTranslation();

  const [loading, setLoading] = useState<Loading>('');
  const [update, setUpdate] = useState<UpdateInfo>();
  const installed = !!state && state !== 'notInstall';

  // The server caches the answer for an hour, so asking on every visit is cheap.
  useEffect(() => {
    if (!installed) return;
    vpn.api
      .getUpdate()
      .then((rsp) => {
        if (rsp.code === 0) setUpdate(rsp.data);
      })
      .catch(() => {});
  }, [installed, vpn.api]);

  const hasUpdate = installed && !!update && isNewer(update.latest, update.current);

  function act(kind: Loading, request: () => Promise<Rsp>, lock = false) {
    if (loading !== '') return;
    setLoading(kind);
    if (lock) setIsLocked(true);
    onError('');

    request()
      .then((rsp) => {
        if (rsp.code !== 0) onError(rsp.msg);
      })
      .catch((err) => onError(err?.message || 'Request failed'))
      .finally(() => {
        setLoading('');
        if (lock) setIsLocked(false);
        onChange();
      });
  }

  function runUpdate() {
    act(
      'updating',
      () =>
        vpn.api.update().then((rsp) => {
          if (rsp.code === 0) setUpdate(undefined);
          return rsp;
        }),
      true
    );
  }

  return (
    <div className="flex items-center justify-between">
      <span className="text-base">{vpn.title}</span>

      <div className="flex items-center space-x-2">
        {hasUpdate && update && (
          <Popconfirm
            title={t('settings.vpn.update', { name: vpn.title, version: update.latest })}
            description={t('settings.vpn.updateDesc')}
            onConfirm={runUpdate}
            okText={t('settings.vpn.okBtn')}
            cancelText={t('settings.vpn.cancelBtn')}
            placement="bottom"
            disabled={loading !== ''}
          >
            <div className="flex cursor-pointer rounded p-1 text-blue-500 hover:bg-neutral-600 hover:text-blue-500/80">
              {loading === 'updating' ? (
                <LoaderIcon className="animate-spin" size={18} />
              ) : (
                <CircleArrowUpIcon size={18} />
              )}
            </div>
          </Popconfirm>
        )}

        {state && ['notLogin', 'stopped', 'running'].includes(state) && (
          <>
            <Popconfirm
              title={t('settings.vpn.restart', { name: vpn.title })}
              onConfirm={() => act('restarting', vpn.api.restart)}
              okText={t('settings.vpn.okBtn')}
              cancelText={t('settings.vpn.cancelBtn')}
              placement="bottom"
              disabled={loading !== ''}
            >
              <div className="flex cursor-pointer rounded p-1 text-green-500 hover:bg-neutral-600 hover:text-green-500/80">
                {loading === 'restarting' ? (
                  <LoaderIcon className="animate-spin" size={18} />
                ) : (
                  <RotateCwIcon size={18} />
                )}
              </div>
            </Popconfirm>

            <Popconfirm
              title={t('settings.vpn.stop', { name: vpn.title })}
              description={t('settings.vpn.stopDesc')}
              onConfirm={() => act('stopping', vpn.api.stop)}
              okText={t('settings.vpn.okBtn')}
              cancelText={t('settings.vpn.cancelBtn')}
              placement="bottom"
              disabled={loading !== ''}
            >
              <div className="flex cursor-pointer rounded p-1 text-red-500 hover:bg-neutral-600 hover:text-red-500/80">
                {loading === 'stopping' ? (
                  <LoaderIcon className="animate-spin" size={18} />
                ) : (
                  <CircleStopIcon size={18} />
                )}
              </div>
            </Popconfirm>
          </>
        )}

        {installed && (
          <Popover
            content={
              <div className="flex min-w-[250px] flex-col">
                <Swap />
                <Uninstall vpn={vpn} onSuccess={onChange} />
              </div>
            }
            placement="bottom"
            trigger="click"
          >
            <div className="flex cursor-pointer rounded p-1 text-white hover:bg-neutral-700/50">
              <EllipsisIcon size={18} />
            </div>
          </Popover>
        )}
      </div>
    </div>
  );
};
