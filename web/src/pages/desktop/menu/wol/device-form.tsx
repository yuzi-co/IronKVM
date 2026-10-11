import { useState } from 'react';
import { Button, Input, Select } from 'antd';
import { useTranslation } from 'react-i18next';

import { saveWolDevice, scanWol } from '@/api/network.ts';
import type { WolNeighbor } from '@/api/network.ts';

import { DeviceForm, toRequest, validateForm } from './device-form.ts';

type Props = {
  initial: DeviceForm;
  onSaved: () => void;
  onCancel: () => void;
};

// DeviceFormView adds a device or edits one. It sits inside the popover rather
// than in a modal: a click in a modal is a click outside the popover, which
// would close it and drop the form.
export const DeviceFormView = ({ initial, onSaved, onCancel }: Props) => {
  const { t } = useTranslation();

  const [form, setForm] = useState<DeviceForm>(initial);
  const [tried, setTried] = useState(false);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);
  const [scanning, setScanning] = useState(false);
  const [neighbors, setNeighbors] = useState<WolNeighbor[] | null>(null);

  const errors = validateForm(form);
  const shown = tried ? errors : {};
  const isEdit = initial.oldMac !== '';

  function update(patch: Partial<DeviceForm>) {
    setForm((f) => ({ ...f, ...patch }));
    setError('');
  }

  async function scan() {
    setScanning(true);
    setError('');
    try {
      const rsp = await scanWol();
      if (rsp.code !== 0) {
        setError(t('wol.scanFailed'));
        return;
      }
      setNeighbors(rsp.data.neighbors as WolNeighbor[]);
    } catch {
      setError(t('wol.scanFailed'));
    } finally {
      setScanning(false);
    }
  }

  function pick(mac: string) {
    const found = neighbors?.find((n) => n.mac === mac);
    if (!found) return;
    update({
      mac: found.mac,
      ip: found.ip,
      name: form.name || found.name
    });
  }

  async function save() {
    setTried(true);
    if (Object.keys(errors).length > 0 || saving) return;

    setSaving(true);
    try {
      const rsp = await saveWolDevice(toRequest(form));
      if (rsp.code === 0) {
        onSaved();
        return;
      }
      if (rsp.code === -4) setError(t('wol.duplicate'));
      else if (rsp.code === -2) setError(t('wol.invalidMac'));
      else setError(rsp.msg ? `${t('wol.saveFailed')}: ${rsp.msg}` : t('wol.saveFailed'));
    } catch {
      setError(t('wol.saveFailed'));
    } finally {
      setSaving(false);
    }
  }

  const invalid = 'text-xs text-red-400';

  return (
    <div className="flex w-[300px] flex-col space-y-3 py-1">
      <div className="text-base font-bold text-neutral-300">
        {isEdit ? t('wol.edit') : t('wol.add')}
      </div>

      <label className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('wol.name')}</span>
        <Input
          value={form.name}
          maxLength={64}
          onChange={(e) => update({ name: e.target.value })}
        />
      </label>

      <div className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('wol.mac')}</span>
        <Input
          value={form.mac}
          status={shown.mac ? 'error' : undefined}
          placeholder="AA:BB:CC:DD:EE:FF"
          onChange={(e) => update({ mac: e.target.value })}
        />
        {shown.mac && <span className={invalid}>{t('wol.invalidMac')}</span>}
        <div className="flex items-center space-x-2">
          <Button size="small" loading={scanning} onClick={scan}>
            {t('wol.scan')}
          </Button>
          {neighbors && neighbors.length === 0 && (
            <span className="text-xs text-neutral-500">{t('wol.scanEmpty')}</span>
          )}
        </div>
        {neighbors && neighbors.length > 0 && (
          <Select
            size="small"
            placeholder={t('wol.scanPick')}
            onChange={pick}
            options={neighbors.map((n) => ({
              value: n.mac,
              label: `${n.name || n.ip} (${n.mac})${n.listed ? ' *' : ''}`
            }))}
          />
        )}
      </div>

      <label className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('wol.ip')}</span>
        <Input
          value={form.ip}
          status={shown.ip ? 'error' : undefined}
          placeholder="192.168.1.10"
          onChange={(e) => update({ ip: e.target.value })}
        />
        {shown.ip && <span className={invalid}>{t('wol.invalidIp')}</span>}
      </label>

      <label className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('wol.broadcast')}</span>
        <Input
          value={form.broadcast}
          status={shown.broadcast ? 'error' : undefined}
          placeholder="192.168.1.0/24"
          onChange={(e) => update({ broadcast: e.target.value })}
        />
        {shown.broadcast && <span className={invalid}>{t('wol.invalidBroadcast')}</span>}
      </label>

      <div className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('wol.port')}</span>
        <Select
          value={form.port}
          onChange={(port) => update({ port })}
          options={[
            { value: 0, label: t('wol.portFrame') },
            { value: 7, label: 'UDP 7' },
            { value: 9, label: 'UDP 9' }
          ]}
        />
      </div>

      <div className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('wol.secureOn')}</span>
        <Input.Password
          value={form.secureOn}
          status={shown.secureOn ? 'error' : undefined}
          autoComplete="new-password"
          placeholder={
            form.hasSecureOn && !form.clearSecureOn ? t('wol.secureOnKept') : 'AA:BB:CC:DD:EE:FF'
          }
          onChange={(e) => update({ secureOn: e.target.value })}
        />
        {shown.secureOn && <span className={invalid}>{t('wol.invalidSecureOn')}</span>}
        {form.hasSecureOn && (
          <Button
            size="small"
            danger={!form.clearSecureOn}
            onClick={() => update({ clearSecureOn: !form.clearSecureOn, secureOn: '' })}
          >
            {form.clearSecureOn ? t('wol.secureOnKeep') : t('wol.secureOnRemove')}
          </Button>
        )}
      </div>

      {error && <span className={invalid}>{error}</span>}

      <div className="flex space-x-2">
        <Button type="primary" loading={saving} onClick={save}>
          {t('wol.save')}
        </Button>
        <Button onClick={onCancel}>{t('wol.cancel')}</Button>
      </div>
    </div>
  );
};
