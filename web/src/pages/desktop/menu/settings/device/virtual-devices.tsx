import { useEffect, useState } from 'react';
import { Button, Popconfirm, Progress, Switch, Tooltip } from 'antd';
import { useSetAtom } from 'jotai';
import { Volume2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/virtual-device.ts';
import type {
  UsbApplied,
  UsbNetwork as UsbNetworkState,
  VirtualDeviceName,
  VirtualDevices as VirtualDevicesState
} from '@/api/virtual-device.ts';
import { describeFailure } from '@/lib/feedback.ts';
import {
  applyRequest,
  boardDraft,
  changedDevices,
  draftUsed,
  fitsInDraft,
  subnetValid,
  type UsbDraft
} from '@/lib/usb-draft.ts';
import { virtualDiskEnabledAtom } from '@/jotai/settings.ts';
import { useHidMode } from '@/hooks/useHidMode.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { ChangedMark, UsbNetwork } from './usb-network.tsx';

// Every control in the USB section edits a draft. Nothing reaches the board
// until Apply, which sends the whole set in one request so the gadget is
// rebuilt, and the host loses its keyboard, once. Leaving the tab or closing
// Settings drops an unapplied draft, as the other settings forms do.
export const VirtualDevices = () => {
  const { t } = useTranslation();

  const isHidOnlyMode = useHidMode()?.mode === 'hid-only';
  const [devices, setDevices] = useState<VirtualDevicesState | null>(null);
  const [network, setNetwork] = useState<UsbNetworkState | null>(null);
  const [draft, setDraft] = useState<UsbDraft | null>(null);
  const setIsVirtualDiskEnabled = useSetAtom(virtualDiskEnabledAtom);
  const [applying, setApplying] = useState(false);
  const [refusal, setRefusal] = useState('');

  function showBoard(nextDevices: VirtualDevicesState, nextNetwork: UsbNetworkState) {
    setDevices(nextDevices);
    setNetwork(nextNetwork);
    setIsVirtualDiskEnabled(nextDevices.disk.enabled);
  }

  // keepDraft leaves the draft alone, for a reload after a failed apply: the
  // operator still sees what they asked for against what the board now has.
  const load = useStableCallback(async (keepDraft = false) => {
    try {
      const [devicesRsp, networkRsp] = await Promise.all([
        api.getVirtualDevice(),
        api.getUsbNetwork()
      ]);
      for (const rsp of [devicesRsp, networkRsp]) {
        if (rsp.code !== 0) {
          setRefusal(describeFailure(rsp, t('settings.device.endpoints.error')));
          return;
        }
      }

      showBoard(devicesRsp.data, networkRsp.data);
      if (!keepDraft) setDraft(boardDraft(devicesRsp.data, networkRsp.data));
    } catch (err) {
      setRefusal(describeFailure(err, t('settings.device.endpoints.error')));
    }
  });

  useEffect(() => {
    load();
  }, [load]);

  async function apply() {
    if (applying || !draft) return;
    setApplying(true);
    setRefusal('');

    try {
      const rsp = await api.applyUsbDevices(applyRequest(draft));
      if (rsp.code !== 0) {
        // The server owns the numbers, so show its sentence rather than
        // recomputing the budget here and risking a different answer. A
        // failed rebuild may have changed part of the set, so read it back.
        setRefusal(describeFailure(rsp));
        await load(true);
        return;
      }

      const applied = rsp.data as UsbApplied;
      showBoard(applied, applied.usbNetwork);
      setDraft(boardDraft(applied, applied.usbNetwork));
    } catch (err) {
      // Applying restarts the USB gadget, so the request can go missing while
      // it rebuilds. Say so, and read back what the board has now.
      setRefusal(describeFailure(err, t('settings.device.endpoints.error')));
      await load(true);
    } finally {
      setApplying(false);
    }
  }

  if (isHidOnlyMode) {
    return (
      <div className="flex items-center justify-between space-x-10">
        <div className="flex flex-col space-y-1">
          <span>{t('settings.device.hidOnly')}</span>
          <span className="text-xs text-neutral-500">{t('settings.device.hidOnlyDesc')}</span>
        </div>

        <Switch checked={true} disabled={true} />
      </div>
    );
  }

  if (!devices || !network || !draft) {
    return refusal ? <span className="text-xs text-red-500">{refusal}</span> : null;
  }

  const board = boardDraft(devices, network);
  const changed = changedDevices(draft, board);
  const used = draftUsed(draft, devices);
  const free = devices.total - used;
  const validSubnet = subnetValid(draft);

  function row(device: Exclude<VirtualDeviceName, 'network'>) {
    if (!devices || !draft) return null;

    const state = devices[device];
    const on = draft[device];
    const fits = fitsInDraft(draft, devices, device);
    const isChanged = changed.includes(device);

    return (
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1">
          <span>
            {t(`settings.device.${device}`)}
            {isChanged && <ChangedMark />}
          </span>
          <span className="text-xs text-neutral-500">{t(`settings.device.${device}Desc`)}</span>

          {device === 'console' && (
            <span className="text-xs text-amber-500">{t('settings.device.consoleTip')}</span>
          )}

          {/* Audio reaches the browser on both H.264 paths and not on MJPEG.
              WebRTC carries it as an Opus track. Direct carries it on the
              video websocket, as messages whose first byte is 0x10, which a
              video message never starts with; the browser asks for them with
              ?audio=1 and plays them through WebCodecs. MJPEG is a multipart
              response with no room for a second stream.

              Said whatever this browser is on, because the switch is a device
              setting: it presents a sound card to the host for every viewer,
              and the one flipping it may not be the one listening. */}
          {device === 'audio' && (
            <span className="flex items-start space-x-1.5 text-xs text-amber-500">
              <Volume2Icon className="mt-[2px] shrink-0" size={12} />
              <span>{t('settings.device.audioNote')}</span>
            </span>
          )}

          {state.enabled && !state.active && !isChanged && (
            <span className="text-xs text-amber-500">
              {t('settings.device.endpoints.inactive')}
            </span>
          )}
        </div>

        <div className="flex items-center space-x-3">
          <span id={`endpoint-cost-${device}`} className="text-xs text-neutral-500">
            {on
              ? t('settings.device.endpoints.cost', { cost: state.cost })
              : t('settings.device.endpoints.needs', { cost: state.cost })}
          </span>

          {/* antd clones the Tooltip child directly, and a disabled native
              <button> suppresses mouse events, so the tooltip on the Switch
              itself would never open. Wrapping it in a span that stays
              enabled gives the Tooltip a target that still receives hover. */}
          <Tooltip title={fits ? '' : t('settings.device.endpoints.full')}>
            <span className="inline-block">
              <Switch
                checked={on}
                disabled={!fits || applying}
                onChange={(next) => setDraft({ ...draft, [device]: next })}
                aria-describedby={`endpoint-cost-${device}`}
              />
            </span>
          </Tooltip>
        </div>
      </div>
    );
  }

  return (
    <>
      <div className="flex flex-col space-y-1">
        <div className="flex items-center justify-between">
          <span>{t('settings.device.endpoints.title')}</span>
          <span className="text-xs text-neutral-500">
            {t('settings.device.endpoints.used', { used, total: devices.total })}
          </span>
        </div>

        {/* free goes negative on a board that carries more markers than the
            controller fits, and that is precisely the board the warning is
            for. Comparing against 0 exactly left the bar its default colour
            there, at a percent antd clamps to 100. */}
        <Progress
          percent={(used / devices.total) * 100}
          showInfo={false}
          size="small"
          strokeColor={free <= 0 ? '#f59e0b' : undefined}
        />

        <span className="text-xs text-neutral-500">{t('settings.device.endpoints.explain')}</span>

        {/* Said before anything is switched, so the operator learns which
            devices go together from the list rather than from a refusal. */}
        {devices.fits?.length > 0 && (
          <span className="text-xs text-neutral-500">
            {t('settings.device.endpoints.fitTogether', {
              sets: devices.fits
                .map((set) => set.map((name) => t(`settings.device.${name}`)).join(' + '))
                .join('; ')
            })}
          </span>
        )}
      </div>

      {row('console')}
      {row('disk')}
      <UsbNetwork
        board={network}
        cost={devices.network.cost}
        mode={draft.network}
        subnet={draft.subnet}
        fits={fitsInDraft(draft, devices, 'network')}
        changed={changed.includes('network')}
        disabled={applying}
        onChange={(mode, subnet) => setDraft({ ...draft, network: mode, subnet })}
      />
      {row('audio')}

      {changed.length > 0 && (
        <div className="flex items-center justify-between space-x-5">
          <span className="text-xs text-amber-500">
            {validSubnet
              ? t('settings.device.usbApply.pending')
              : t('settings.device.usbNetwork.invalidSubnet')}
          </span>

          <div className="flex shrink-0 items-center space-x-2">
            <Button size="small" disabled={applying} onClick={() => setDraft(board)}>
              {t('settings.device.usbApply.discard')}
            </Button>

            {/* One rebuild for the whole draft, so one question. */}
            <Popconfirm
              title={t('settings.device.usbNetwork.confirm')}
              description={
                <div className="max-w-[280px]">{t('settings.device.usbNetwork.reenumerate')}</div>
              }
              okText={t('settings.device.okBtn')}
              cancelText={t('settings.device.cancelBtn')}
              onConfirm={apply}
              disabled={!validSubnet || applying}
            >
              <Button type="primary" size="small" loading={applying} disabled={!validSubnet}>
                {t('settings.device.usbNetwork.apply')}
              </Button>
            </Popconfirm>
          </div>
        </div>
      )}

      {refusal && <span className="text-xs text-red-500">{refusal}</span>}
    </>
  );
};
