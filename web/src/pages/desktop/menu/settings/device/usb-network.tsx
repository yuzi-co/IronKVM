import { useState } from 'react';
import { Input, Select, Switch, Tooltip } from 'antd';
import { useTranslation } from 'react-i18next';

import type { UsbNetworkMode, UsbNetwork as UsbNetworkState } from '@/api/virtual-device.ts';
import { subnetShapeValid } from '@/lib/usb-draft.ts';

import { CopyButton } from '../components/copy-button.tsx';

// ChangedMark sits beside a row's title while the draft changes that row.
export const ChangedMark = () => {
  const { t } = useTranslation();

  return (
    <span className="ml-2 rounded bg-amber-500/15 px-1.5 py-0.5 text-xs text-amber-500">
      {t('settings.device.usbApply.changed')}
    </span>
  );
};

type Props = {
  // What the board runs now.
  board: UsbNetworkState;
  cost: number;
  // The draft: the mode and subnet Apply would send.
  mode: UsbNetworkMode;
  subnet: string;
  // Whether the link can be switched on beside the rest of the draft.
  fits: boolean;
  changed: boolean;
  disabled: boolean;
  onChange: (mode: UsbNetworkMode, subnet: string) => void;
};

// The USB network row of the USB section. It edits the section's draft and
// applies nothing itself: the section's Apply sends it with every other device.
export const UsbNetwork = ({
  board,
  cost,
  mode,
  subnet,
  fits,
  changed,
  disabled,
  onChange
}: Props) => {
  const { t } = useTranslation();

  // The protocol the switch turns back on with: the last one chosen, the one
  // the board runs, NCM before either.
  const [lastMode, setLastMode] = useState<UsbNetworkMode>(board.mode === 'ecm' ? 'ecm' : 'ncm');

  const validSubnet = subnetShapeValid(subnet);

  function toggle(on: boolean) {
    onChange(on ? lastMode : 'off', subnet);
  }

  function chooseMode(next: UsbNetworkMode) {
    setLastMode(next);
    onChange(next, subnet);
  }

  const options: { value: UsbNetworkMode; label: string; disabled?: boolean }[] = [
    { value: 'ncm', label: t('settings.device.usbNetwork.ncm') },
    { value: 'ecm', label: t('settings.device.usbNetwork.ecm') }
  ];
  if (board.mode === 'rndis') {
    options.push({ value: 'rndis', label: t('settings.device.usbNetwork.rndis'), disabled: true });
  }

  return (
    <div className="flex flex-col space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1">
          <span>
            {t('settings.device.network')}
            {changed && <ChangedMark />}
          </span>
          <span className="text-xs text-neutral-500">
            {t('settings.device.usbNetwork.description')}
          </span>

          {board.mode !== 'off' && !board.active && !changed && (
            <span className="text-xs text-amber-500">
              {t('settings.device.endpoints.inactive')}
            </span>
          )}

          {mode === 'rndis' && (
            <span className="text-xs text-amber-500">
              {t('settings.device.usbNetwork.rndisNote')}
            </span>
          )}
        </div>

        <div className="flex items-center space-x-3">
          <span id="endpoint-cost-network" className="text-xs text-neutral-500">
            {mode !== 'off'
              ? t('settings.device.endpoints.cost', { cost })
              : t('settings.device.endpoints.needs', { cost })}
          </span>

          {/* The span keeps the tooltip working over a disabled switch; see
              the rows in virtual-devices.tsx. */}
          <Tooltip title={fits ? '' : t('settings.device.endpoints.full')}>
            <span className="inline-block">
              <Switch
                checked={mode !== 'off'}
                disabled={disabled || !fits}
                onChange={toggle}
                aria-describedby="endpoint-cost-network"
              />
            </span>
          </Tooltip>
        </div>
      </div>

      {mode !== 'off' && (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <span>{t('settings.device.usbNetwork.mode')}</span>
          <Select<UsbNetworkMode>
            className="w-60 max-w-full"
            value={mode}
            options={options}
            disabled={disabled}
            onChange={chooseMode}
          />
        </div>
      )}

      {mode !== 'off' && (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-col space-y-1">
            <span>{t('settings.device.usbNetwork.subnet')}</span>
            <span className="text-xs text-neutral-500">
              {t('settings.device.usbNetwork.subnetDesc')}
            </span>
            {board.mode !== 'off' && (
              <span className="flex flex-wrap items-center gap-x-3 text-xs text-neutral-500">
                <span className="flex items-center">
                  {t('settings.device.usbNetwork.boardAddress')}&nbsp;
                  <span className="font-mono select-all">{board.board}</span>
                  <CopyButton
                    text={board.board}
                    label={t('settings.device.usbNetwork.boardAddress')}
                  />
                </span>
                <span className="flex items-center">
                  {t('settings.device.usbNetwork.hostAddress')}&nbsp;
                  <span className="font-mono select-all">{board.host}</span>
                  <CopyButton
                    text={board.host}
                    label={t('settings.device.usbNetwork.hostAddress')}
                  />
                </span>
              </span>
            )}
          </div>

          <Input
            className="w-60 max-w-full"
            value={subnet}
            status={validSubnet ? undefined : 'error'}
            disabled={disabled}
            placeholder="172.31.255.0/30"
            onChange={(e) => onChange(mode, e.target.value)}
          />
        </div>
      )}
    </div>
  );
};
