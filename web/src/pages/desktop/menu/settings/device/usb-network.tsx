import { useEffect, useState } from 'react';
import { Button, Input, Popconfirm, Select, Tooltip } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/virtual-device.ts';
import type {
  UsbNetworkMode,
  UsbNetwork as UsbNetworkState,
  VirtualDevices
} from '@/api/virtual-device.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { CopyButton } from '../components/copy-button.tsx';

type Props = {
  devices: VirtualDevices;
  onChanged: () => Promise<void>;
};

// The shape only. The server holds the real rules (private, /24 to /30, the
// network address) and says which one a subnet breaks.
const SUBNET_SHAPE = /^\d{1,3}(\.\d{1,3}){3}\/\d{1,2}$/;

export const UsbNetwork = ({ devices, onChanged }: Props) => {
  const { t } = useTranslation();

  const [state, setState] = useState<UsbNetworkState | null>(null);
  const [mode, setMode] = useState<UsbNetworkMode>('off');
  const [subnet, setSubnet] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const load = useStableCallback(async () => {
    try {
      const rsp = await api.getUsbNetwork();
      if (rsp.code !== 0) {
        setError(describeFailure(rsp, t('settings.device.endpoints.error')));
        return;
      }

      setState(rsp.data);
      setMode(rsp.data.mode);
      setSubnet(rsp.data.subnet);
    } catch (err) {
      setError(describeFailure(err, t('settings.device.endpoints.error')));
    }
  });

  // Whether the network fits depends on every other switch in the panel, so
  // reload whenever the panel's own numbers change.
  useEffect(() => {
    load();
  }, [load, devices]);

  if (!state) return null;

  const cost = devices.network.cost;
  const dirty = mode !== state.mode || subnet.trim() !== state.subnet;
  const subnetValid = SUBNET_SHAPE.test(subnet.trim());
  // Turning the link on from off is the only change the budget can refuse.
  // A switch between NCM and ECM costs the same endpoints.
  const blocked = state.mode === 'off' && !state.fits;

  async function apply() {
    if (loading || mode === 'rndis') return;
    setLoading(true);
    setError('');

    try {
      const rsp = await api.setUsbNetwork(mode, subnet.trim());
      if (rsp.code !== 0) {
        // The server owns the rules and the numbers, so show its sentence.
        setError(describeFailure(rsp));
        return;
      }

      setState(rsp.data);
      setMode(rsp.data.mode);
      setSubnet(rsp.data.subnet);
      await onChanged();
    } catch (err) {
      // Applying rebuilds the USB gadget, so the request can go missing while
      // it does. Say so rather than leave the old values with no word.
      setError(describeFailure(err, t('settings.device.endpoints.error')));
    } finally {
      setLoading(false);
    }
  }

  const options: { value: UsbNetworkMode; label: string; disabled?: boolean }[] = [
    { value: 'off', label: t('settings.device.usbNetwork.off') },
    { value: 'ncm', label: t('settings.device.usbNetwork.ncm'), disabled: blocked },
    { value: 'ecm', label: t('settings.device.usbNetwork.ecm'), disabled: blocked }
  ];
  if (state.mode === 'rndis') {
    options.push({ value: 'rndis', label: t('settings.device.usbNetwork.rndis'), disabled: true });
  }

  return (
    <div className="flex flex-col space-y-3">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex flex-col space-y-1">
          <span>{t('settings.device.network')}</span>
          <span className="text-xs text-neutral-500">
            {t('settings.device.usbNetwork.description')}
          </span>

          {state.mode !== 'off' && !state.active && (
            <span className="text-xs text-amber-500">
              {t('settings.device.endpoints.inactive')}
            </span>
          )}

          {state.mode === 'rndis' && (
            <span className="text-xs text-amber-500">
              {t('settings.device.usbNetwork.rndisNote')}
            </span>
          )}

          {blocked && <span className="text-xs text-amber-500">{state.refusal}</span>}
        </div>

        <div className="flex flex-wrap items-center gap-3">
          <span id="endpoint-cost-network" className="text-xs text-neutral-500">
            {state.mode !== 'off'
              ? t('settings.device.endpoints.cost', { cost })
              : t('settings.device.endpoints.needs', { cost })}
          </span>

          <Tooltip title={blocked ? t('settings.device.endpoints.full') : ''}>
            <Select<UsbNetworkMode>
              className="w-60 max-w-full"
              value={mode}
              options={options}
              disabled={loading}
              onChange={setMode}
              aria-describedby="endpoint-cost-network"
            />
          </Tooltip>
        </div>
      </div>

      {mode !== 'off' && (
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="flex flex-col space-y-1">
            <span>{t('settings.device.usbNetwork.subnet')}</span>
            <span className="text-xs text-neutral-500">
              {t('settings.device.usbNetwork.subnetDesc')}
            </span>
            {state.mode !== 'off' && (
              <span className="flex flex-wrap items-center gap-x-3 text-xs text-neutral-500">
                <span className="flex items-center">
                  {t('settings.device.usbNetwork.boardAddress')}&nbsp;
                  <span className="font-mono select-all">{state.board}</span>
                  <CopyButton
                    text={state.board}
                    label={t('settings.device.usbNetwork.boardAddress')}
                  />
                </span>
                <span className="flex items-center">
                  {t('settings.device.usbNetwork.hostAddress')}&nbsp;
                  <span className="font-mono select-all">{state.host}</span>
                  <CopyButton
                    text={state.host}
                    label={t('settings.device.usbNetwork.hostAddress')}
                  />
                </span>
              </span>
            )}
          </div>

          <Input
            className="w-60 max-w-full"
            value={subnet}
            status={subnetValid ? undefined : 'error'}
            disabled={loading}
            placeholder="172.31.255.0/30"
            onChange={(e) => setSubnet(e.target.value)}
          />
        </div>
      )}

      {dirty && (
        <div className="flex items-center justify-between space-x-5">
          <span className="text-xs text-amber-500">
            {subnetValid
              ? t('settings.device.usbNetwork.reenumerate')
              : t('settings.device.usbNetwork.invalidSubnet')}
          </span>

          <Popconfirm
            title={t('settings.device.usbNetwork.confirm')}
            description={t('settings.device.usbNetwork.reenumerate')}
            okText={t('settings.device.okBtn')}
            cancelText={t('settings.device.cancelBtn')}
            onConfirm={apply}
            disabled={!subnetValid || mode === 'rndis'}
          >
            <Button
              type="primary"
              size="small"
              loading={loading}
              disabled={!subnetValid || mode === 'rndis'}
            >
              {t('settings.device.usbNetwork.apply')}
            </Button>
          </Popconfirm>
        </div>
      )}

      {error && <span className="text-xs text-red-500">{error}</span>}
    </div>
  );
};
