import { Divider, message, Popconfirm, Tooltip } from 'antd';
import clsx from 'clsx';
import { ArrowBigUpIcon, MoonIcon, PowerOffIcon, SunIcon } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { KeyboardReport } from '@/lib/keyboard.ts';
import { client, MessageEvent } from '@/lib/websocket.ts';
import { useExtendedKeys, useSendKey } from '@/hooks/useExtendedKeys.ts';

type HostPowerProps = {
  showConfirm: boolean;
};

// Generic Desktop System Control usages.
const SYSTEM_POWER_DOWN = 0x81;
const SYSTEM_SLEEP = 0x82;
const SYSTEM_WAKE_UP = 0x83;

export const HostPower = ({ showConfirm }: HostPowerProps) => {
  const { t } = useTranslation();
  const state = useExtendedKeys();
  const sendKey = useSendKey();

  if (state === 'hidden') return null;

  const disabled = state === 'disabled';

  // A suspended host often ignores System Control Wake Up from the device it
  // suspended, and wakes on an ordinary key instead. Shift is the key: on a
  // host that is already awake it types nothing and triggers nothing.
  function wakeWithKey() {
    const keyboard = new KeyboardReport();
    const sent = [keyboard.keyDown('ShiftLeft'), keyboard.reset()].every((report) =>
      client.send(new Uint8Array([MessageEvent.Keyboard, ...report]))
    );
    if (!sent) {
      message.error(t('input.keyFailed'));
    }
  }

  const item = (Icon: LucideIcon, label: string, action: () => void, confirm?: string) => {
    const row = (
      <button
        type="button"
        aria-disabled={disabled}
        className={clsx(
          'flex w-full items-center space-x-2 rounded p-0 px-3 py-1.5 text-left select-none',
          disabled
            ? 'cursor-not-allowed text-neutral-500'
            : 'cursor-pointer hover:bg-neutral-700/70'
        )}
        onClick={disabled || (showConfirm && confirm) ? undefined : action}
      >
        <Icon size={16} />
        <span>{label}</span>
      </button>
    );

    if (disabled || !showConfirm || !confirm) return row;

    return (
      <Popconfirm
        placement="bottomLeft"
        title={confirm}
        okText={t('power.okBtn')}
        cancelText={t('power.cancelBtn')}
        onConfirm={action}
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
        {item(
          MoonIcon,
          t('power.sleep'),
          () => sendKey('system', SYSTEM_SLEEP),
          t('power.sleepConfirm')
        )}
        {item(SunIcon, t('power.wake'), () => sendKey('system', SYSTEM_WAKE_UP))}
        {item(ArrowBigUpIcon, t('power.wakeKey'), wakeWithKey)}
        {item(
          PowerOffIcon,
          t('power.powerDown'),
          () => sendKey('system', SYSTEM_POWER_DOWN),
          t('power.powerDownConfirm')
        )}
      </div>
      <div className="max-w-[240px] px-1 pt-1 text-xs text-neutral-500">
        {disabled ? t('input.hidDisabled') : t('power.wakeTip')}
      </div>
    </>
  );
};
