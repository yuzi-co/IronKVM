import { Divider, Popconfirm, Tooltip } from 'antd';
import { MoonIcon, PowerOffIcon, SunIcon } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { sendKey } from '@/api/hid.ts';
import { useExtendedKeys } from '@/hooks/useExtendedKeys.ts';

type HostPowerProps = {
  showConfirm: boolean;
};

// Generic Desktop System Control usages.
const SYSTEM_POWER_DOWN = 0x81;
const SYSTEM_SLEEP = 0x82;
const SYSTEM_WAKE_UP = 0x83;

export const HostPower = ({ showConfirm }: HostPowerProps) => {
  const { t } = useTranslation();
  const available = useExtendedKeys();

  if (!available) return null;

  const item = (Icon: LucideIcon, label: string, usage: number, confirm?: string) => {
    const row = (
      <div
        className="flex cursor-pointer items-center space-x-2 rounded px-3 py-1.5 select-none hover:bg-neutral-700/70"
        onClick={showConfirm && confirm ? undefined : () => sendKey('system', usage)}
      >
        <Icon size={16} />
        <span>{label}</span>
      </div>
    );

    if (!showConfirm || !confirm) return row;

    return (
      <Popconfirm
        placement="bottomLeft"
        title={confirm}
        okText={t('power.okBtn')}
        cancelText={t('power.cancelBtn')}
        onConfirm={() => sendKey('system', usage)}
        color="#404040"
      >
        {row}
      </Popconfirm>
    );
  };

  return (
    <>
      <Divider style={{ margin: '10px 0' }} />
      <Tooltip title={t('power.hostOsTip')} placement="right">
        <div className="px-1 pb-1 text-xs text-neutral-400">{t('power.hostOs')}</div>
      </Tooltip>
      <div className="flex flex-col space-y-1">
        {item(MoonIcon, t('power.sleep'), SYSTEM_SLEEP, t('power.sleepConfirm'))}
        {item(SunIcon, t('power.wake'), SYSTEM_WAKE_UP)}
        {item(PowerOffIcon, t('power.powerDown'), SYSTEM_POWER_DOWN, t('power.powerDownConfirm'))}
      </div>
    </>
  );
};
