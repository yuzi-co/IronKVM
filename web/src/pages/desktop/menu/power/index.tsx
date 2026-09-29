import { useEffect, useState } from 'react';
import { Divider, Switch, Tooltip } from 'antd';
import { LoaderCircleIcon, PowerIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm';
import * as localstorage from '@/lib/localstorage.ts';
import { MenuItem } from '@/components/menu-item.tsx';
import { StatusDot } from '@/components/status-dot.tsx';

import { HostPower } from './host-power.tsx';
import { PowerLedState } from './power-led-state.tsx';
import { PowerLong } from './power-long.tsx';
import { PowerShort } from './power-short.tsx';
import { Reset } from './reset.tsx';

// The "power LED connected" switch is set once per wiring, so it is a
// setting rather than a menu entry. The menu shows the LED's state, and the
// icon carries it as a light, as PiKVM's ATX menu does: green while the
// host's LED is lit, grey while it is dark, and none while the header is not
// wired and the state is unknown. A board with an HDD LED input (only the
// alpha board) adds a second light while the host's disk is busy.
export const Power = () => {
  const { t } = useTranslation();

  const [isPowerOn, setIsPowerOn] = useState(false);
  const [ledConnected, setLedConnected] = useState(false);
  const [hasHdd, setHasHdd] = useState(false);
  const [isHddOn, setIsHddOn] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [showConfirm, setShowConfirm] = useState(() => localstorage.getPowerConfirm());

  useEffect(() => {
    async function getLed() {
      try {
        const rsp = await api.getGpio();
        if (rsp.code === 0) {
          setIsPowerOn(rsp.data.pwr);
          setLedConnected(rsp.data.ledConnected);
          setHasHdd(!!rsp.data.hasHdd);
          setIsHddOn(!!rsp.data.hdd);
        }
      } catch (err) {
        console.log(err);
      }
    }

    getLed();
    const interval = setInterval(getLed, 5000);

    return () => {
      clearInterval(interval);
    };
  }, []);

  function updateShowConfirm(value: boolean) {
    setShowConfirm(value);
    localstorage.setPowerConfirm(value);
  }

  // The HDD LED shares the front panel header with the power LED, so it is
  // trusted only where the power LED is marked wired.
  const showHdd = hasHdd && ledConnected;

  const icon = (
    <div className="relative flex h-[18px] w-[18px]">
      {isLoading ? (
        <LoaderCircleIcon className="animate-spin" size={18} />
      ) : (
        <PowerIcon size={18} />
      )}
      {ledConnected && <StatusDot tone={isPowerOn ? 'ok' : 'off'} />}
      {showHdd && isHddOn && <StatusDot tone="warning" position="bottom" />}
    </div>
  );

  let title = t('power.title');
  if (ledConnected) {
    title = `${title}: ${isPowerOn ? t('power.ledOn') : t('power.ledOff')}`;
  }

  const content = (
    <div className="min-w-[200px]">
      <div className="flex items-center justify-between px-1">
        <span className="text-base font-bold text-neutral-300">{t('power.title')}</span>

        <div className="flex items-center space-x-2">
          <Tooltip title={t('power.showConfirmTip')} placement="right">
            <div className="flex items-center space-x-2">
              <span className="text-xs text-neutral-400">{t('power.showConfirm')}</span>

              <Switch size="small" checked={showConfirm} onChange={updateShowConfirm} />
            </div>
          </Tooltip>
        </div>
      </div>

      <div className="flex flex-col space-y-1 px-1 pt-2">
        <PowerLedState isPowerOn={isPowerOn} connected={ledConnected} />
        {showHdd && (
          <div className="flex items-center justify-between">
            <span className="text-xs text-neutral-400">{t('power.hddLed')}</span>
            <span className="text-xs text-neutral-300">
              {isHddOn ? t('power.hddActive') : t('power.hddIdle')}
            </span>
          </div>
        )}
      </div>

      <Divider style={{ margin: '10px 0 15px 0' }} />

      <div className="flex flex-col space-y-1">
        <Reset isLoading={isLoading} setIsLoading={setIsLoading} />
        <PowerShort showConfirm={showConfirm} isLoading={isLoading} setIsLoading={setIsLoading} />
        <PowerLong isLoading={isLoading} setIsLoading={setIsLoading} />
      </div>

      <HostPower showConfirm={showConfirm} />
    </div>
  );

  return <MenuItem title={title} icon={icon} content={content} />;
};
