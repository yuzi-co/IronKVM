import { useEffect, useState } from 'react';
import { Button, Popconfirm, Popover, Tooltip } from 'antd';
import {
  CircleArrowUpIcon,
  EllipsisIcon,
  LoaderCircleIcon,
  LogOutIcon,
  RotateCwIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { describeFailure } from '@/lib/feedback.ts';
import { isValidVersion, versionGt } from '@/lib/version.ts';
import { StatusDot } from '@/components/status-dot.tsx';

import { StateTag } from '../components/status-tag.tsx';
import { MenuRow } from './menu-row.tsx';
import { Swap } from './swap.tsx';
import type { Rsp, State, UpdateInfo, VpnInfo } from './types.ts';
import { Uninstall } from './uninstall.tsx';
import { hasDaemon, statusTag, type Tag } from './view.ts';

type HeaderProps = {
  vpn: VpnInfo;
  state: State | undefined;
  // The last status request failed.
  failed: boolean;
  setIsLocked: (isLocked: boolean) => void;
  onChange: () => void;
  onError: (msg: string) => void;
};

type Loading = '' | 'restarting' | 'updating' | 'loggingOut';

function isNewer(latest: string, current: string) {
  if (!latest || !current) return false;
  if (isValidVersion(latest) && isValidVersion(current)) return versionGt(latest, current);
  return latest !== current;
}

const tagColor: Record<Tag, 'green' | 'gold' | 'red' | 'default'> = {
  connected: 'green',
  needsLogin: 'gold',
  off: 'default',
  error: 'red'
};

// Header is the title, the badge saying where the VPN stands, and the menu of
// the actions an operator needs now and then.
export const Header = ({ vpn, state, failed, setIsLocked, onChange, onError }: HeaderProps) => {
  const { t } = useTranslation();

  const [loading, setLoading] = useState<Loading>('');
  const [update, setUpdate] = useState<UpdateInfo>();
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const installed = !!state && state !== 'notInstall';
  const tag = statusTag(state, failed);

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
    setIsMenuOpen(false);
    setLoading(kind);
    if (lock) setIsLocked(true);
    onError('');

    request()
      .then((rsp) => {
        if (rsp.code !== 0) onError(describeFailure(rsp));
      })
      .catch((err) => onError(describeFailure(err)))
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

  const busy = loading !== '';
  const tagLabel: Record<Tag, string> = {
    connected: t('settings.vpn.connected'),
    needsLogin: t('settings.vpn.needsLogin'),
    off: t('common.off'),
    error: t('settings.vpn.error')
  };

  const menu = (
    <div className="flex min-w-[240px] flex-col">
      {hasUpdate && update && (
        <Popconfirm
          title={t('settings.vpn.update', { name: vpn.title, version: update.latest })}
          description={t('settings.vpn.updateDesc')}
          onConfirm={runUpdate}
          okText={t('settings.vpn.okBtn')}
          cancelText={t('settings.vpn.cancelBtn')}
          placement="left"
          disabled={busy}
        >
          <MenuRow
            icon={<CircleArrowUpIcon className="text-blue-500" size={18} />}
            label={t('settings.vpn.updateTip', { version: update.latest })}
            disabled={busy}
          />
        </Popconfirm>
      )}

      {hasDaemon(state) && (
        <Popconfirm
          title={t('settings.vpn.restart', { name: vpn.title })}
          onConfirm={() => act('restarting', vpn.api.restart)}
          okText={t('settings.vpn.okBtn')}
          cancelText={t('settings.vpn.cancelBtn')}
          placement="left"
          disabled={busy}
        >
          <MenuRow
            icon={<RotateCwIcon size={18} />}
            label={t('settings.vpn.restartService')}
            disabled={busy}
          />
        </Popconfirm>
      )}

      {(state === 'stopped' || state === 'running') && (
        <Popconfirm
          title={<div className="max-w-[320px]">{vpn.logoutWarning}</div>}
          onConfirm={() => act('loggingOut', vpn.api.logout)}
          okText={t('settings.vpn.okBtn')}
          cancelText={t('settings.vpn.cancelBtn')}
          placement="left"
          disabled={busy}
        >
          <MenuRow icon={<LogOutIcon size={18} />} label={vpn.logoutLabel} disabled={busy} />
        </Popconfirm>
      )}

      <Swap />
      <Uninstall vpn={vpn} onSuccess={onChange} />
    </div>
  );

  return (
    <div className="flex items-center justify-between">
      <div className="flex items-center space-x-2">
        <span className="text-base">{vpn.title}</span>
        {tag && <StateTag color={tagColor[tag]} label={tagLabel[tag]} />}
      </div>

      {installed && (
        <Popover
          content={menu}
          placement="bottomRight"
          trigger="click"
          open={isMenuOpen}
          onOpenChange={setIsMenuOpen}
        >
          <Tooltip title={t('settings.vpn.moreTip')}>
            <Button
              type="text"
              size="small"
              aria-label={t('settings.vpn.moreTip')}
              icon={
                <span className="relative flex h-[18px] w-[18px]">
                  {busy ? (
                    <LoaderCircleIcon className="animate-spin" size={18} />
                  ) : (
                    <EllipsisIcon size={18} />
                  )}
                  {hasUpdate && !busy && <StatusDot tone="active" />}
                </span>
              }
            />
          </Tooltip>
        </Popover>
      )}
    </div>
  );
};
