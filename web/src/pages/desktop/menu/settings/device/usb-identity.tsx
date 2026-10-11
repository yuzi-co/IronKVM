import { useEffect, useState } from 'react';
import { Button, Input, Popconfirm, Select } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/virtual-device.ts';
import type { UsbIdentityState } from '@/api/virtual-device.ts';
import { describeFailure } from '@/lib/feedback.ts';
import {
  identityErrors,
  sameIdentity,
  type UsbIdentity as Identity,
  type UsbIdentityPresetName
} from '@/lib/usb-identity.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { ChangedMark } from './usb-network.tsx';

const PRESET_ORDER: UsbIdentityPresetName[] = [
  'default',
  'logitech',
  'microsoft',
  'dell',
  'custom'
];

// The USB identity row of the USB section: the vendor and product IDs and the
// strings the host sees. It applies on its own, apart from the section's
// Apply, because it has its own rebuild and its own warning: the host sees a
// different device afterwards, and may load drivers for it.
export const UsbIdentity = () => {
  const { t } = useTranslation();

  const [board, setBoard] = useState<UsbIdentityState | null>(null);
  const [preset, setPreset] = useState<UsbIdentityPresetName>('default');
  const [draft, setDraft] = useState<Identity | null>(null);
  const [applying, setApplying] = useState(false);
  const [failure, setFailure] = useState('');

  function show(state: UsbIdentityState) {
    setBoard(state);
    setPreset(state.preset);
    setDraft(state.identity);
  }

  const load = useStableCallback(async () => {
    try {
      const rsp = await api.getUsbIdentity();
      if (rsp.code !== 0) {
        setFailure(describeFailure(rsp, t('settings.device.usbIdentity.error')));
        return;
      }
      show(rsp.data);
    } catch (err) {
      setFailure(describeFailure(err, t('settings.device.usbIdentity.error')));
    }
  });

  useEffect(() => {
    load();
  }, [load]);

  if (!board || !draft) {
    return failure ? <span className="text-xs text-red-500">{failure}</span> : null;
  }

  const errors = identityErrors(preset, draft);
  const valid = Object.keys(errors).length === 0;
  const changed = preset !== board.preset || !sameIdentity(draft, board.identity);
  const isCustom = preset === 'custom';
  // Nothing saved: restoring would change nothing and rebuild nothing.
  const isDefault = board.preset === 'default' && board.identity.serial === '';

  function choosePreset(next: UsbIdentityPresetName) {
    setPreset(next);
    const named = board?.presets.find((p) => p.name === next);
    if (named && draft) {
      // A named preset brings its own IDs and strings; the serial stays.
      setDraft({ ...named, serial: draft.serial });
    }
  }

  function edit(field: keyof Identity, value: string) {
    if (draft) setDraft({ ...draft, [field]: value });
  }

  async function save(nextPreset: UsbIdentityPresetName, identity: Identity) {
    if (applying) return;
    setApplying(true);
    setFailure('');

    try {
      const rsp = await api.setUsbIdentity(nextPreset, identity);
      if (rsp.code !== 0) {
        setFailure(describeFailure(rsp));
        await load();
        return;
      }
      show(rsp.data);
    } catch (err) {
      // The rebuild can take the request with it. Read back what the board
      // has now.
      setFailure(describeFailure(err, t('settings.device.usbIdentity.error')));
      await load();
    } finally {
      setApplying(false);
    }
  }

  const live = board.live;
  const fields: { key: keyof Identity; label: string; mono?: boolean }[] = [
    { key: 'vendorId', label: t('settings.device.usbIdentity.vendorId'), mono: true },
    { key: 'productId', label: t('settings.device.usbIdentity.productId'), mono: true },
    { key: 'manufacturer', label: t('settings.device.usbIdentity.manufacturer') },
    { key: 'product', label: t('settings.device.usbIdentity.product') }
  ];

  return (
    <div className="flex flex-col space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-col space-y-1">
          <span>
            {t('settings.device.usbIdentity.title')}
            {changed && <ChangedMark />}
          </span>
          <span className="text-xs text-neutral-500">
            {t('settings.device.usbIdentity.description')}
          </span>
          {live.vendorId && (
            <span className="text-xs text-neutral-500">
              {t('settings.device.usbIdentity.live')}&nbsp;
              <span className="font-mono">
                {live.vendorId}:{live.productId}
              </span>
              &nbsp;{live.manufacturer} {live.product}
            </span>
          )}
          {board.bootOverride && (
            <span className="text-xs text-amber-500">
              {t('settings.device.usbIdentity.bootOverride')}
            </span>
          )}
        </div>

        <Select<UsbIdentityPresetName>
          className="w-60 max-w-full"
          value={preset}
          disabled={applying}
          onChange={choosePreset}
          options={PRESET_ORDER.map((name) => ({
            value: name,
            label: t(`settings.device.usbIdentity.presets.${name}`)
          }))}
        />
      </div>

      {fields.map((field) => (
        <div key={field.key} className="flex flex-wrap items-center justify-between gap-3">
          <span>{field.label}</span>
          <Input
            className={`w-60 max-w-full ${field.mono ? 'font-mono' : ''}`}
            value={draft[field.key]}
            disabled={applying || !isCustom}
            status={errors[field.key] ? 'error' : undefined}
            maxLength={field.mono ? 6 : 64}
            onChange={(e) => edit(field.key, e.target.value)}
          />
        </div>
      ))}

      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-col space-y-1">
          <span>{t('settings.device.usbIdentity.serial')}</span>
          <span className="text-xs text-neutral-500">
            {t('settings.device.usbIdentity.serialDesc')}
          </span>
        </div>
        <Input
          className="w-60 max-w-full font-mono"
          value={draft.serial}
          disabled={applying}
          status={errors.serial ? 'error' : undefined}
          placeholder={live.serial || t('settings.device.usbIdentity.serialPlaceholder')}
          maxLength={64}
          onChange={(e) => edit('serial', e.target.value)}
        />
      </div>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <span className="text-xs text-amber-500">
          {!valid ? t('settings.device.usbIdentity.invalid') : ''}
        </span>

        <div className="flex shrink-0 items-center space-x-2">
          <Popconfirm
            title={t('settings.device.usbIdentity.restoreConfirm')}
            description={
              <div className="max-w-[280px]">{t('settings.device.usbIdentity.warning')}</div>
            }
            okText={t('settings.device.okBtn')}
            cancelText={t('settings.device.cancelBtn')}
            onConfirm={() => {
              const named = board.presets.find((p) => p.name === 'default');
              if (named) save('default', { ...named, serial: '' });
            }}
            disabled={applying || isDefault}
          >
            <Button size="small" disabled={applying || isDefault}>
              {t('settings.device.usbIdentity.restore')}
            </Button>
          </Popconfirm>

          {changed && (
            <Button size="small" disabled={applying} onClick={() => show(board)}>
              {t('settings.device.usbApply.discard')}
            </Button>
          )}

          <Popconfirm
            title={t('settings.device.usbIdentity.confirm')}
            description={
              <div className="max-w-[280px]">{t('settings.device.usbIdentity.warning')}</div>
            }
            okText={t('settings.device.okBtn')}
            cancelText={t('settings.device.cancelBtn')}
            onConfirm={() => save(preset, draft)}
            disabled={!valid || !changed || applying}
          >
            <Button type="primary" size="small" loading={applying} disabled={!valid || !changed}>
              {t('settings.device.usbIdentity.apply')}
            </Button>
          </Popconfirm>
        </div>
      </div>

      {failure && <span className="text-xs text-red-500">{failure}</span>}
    </div>
  );
};
