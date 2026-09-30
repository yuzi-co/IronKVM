import { useEffect, useState } from 'react';
import { Button, Popover, Tooltip } from 'antd';
import clsx from 'clsx';
import { useSetAtom } from 'jotai';
import {
  HardDriveIcon,
  MonitorXIcon,
  ShieldOffIcon,
  ThermometerIcon,
  TriangleAlertIcon
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { getHealth } from '@/api/vm.ts';
import { formatBytes, healthAlerts, worstSeverity } from '@/lib/health.ts';
import type { Alert, Health } from '@/lib/health.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { settingsOpenRequestAtom } from '@/jotai/settings.ts';

import { useStreamState } from '../screen/use-stream-state.ts';

// The board's readings change slowly, and every one is a file read or a
// statfs on the server, so half a minute is often enough.
const HEALTH_POLL_MS = 30000;

const alertIcon: Record<Alert['kind'], LucideIcon> = {
  temperature: ThermometerIcon,
  storage: HardDriveIcon,
  vpn: ShieldOffIcon,
  stream: MonitorXIcon
};

// useHealth reads the board's health every half minute while the tab is
// visible. A failed read keeps the last answer.
function useHealth(): Health | null {
  const [health, setHealth] = useState<Health | null>(null);

  useEffect(() => {
    let active = true;

    function read() {
      getHealth()
        .then((rsp) => {
          if (active && rsp.code === 0 && rsp.data) setHealth(rsp.data as Health);
        })
        .catch(() => {});
    }

    read();
    const stopPoll = pollWhileVisible(read, HEALTH_POLL_MS);
    return () => {
      active = false;
      stopPoll();
    };
  }, []);

  return health;
}

// Alerts is the icon at the left of the bar that shows only while something
// is wrong, as PiKVM's health icons do: the SoC running hot, the image
// storage nearly full, a VPN set to start at boot that is down, or a failed
// video stream. Its popover says what each one means in plain words.
export const Alerts = () => {
  const { t } = useTranslation();
  const health = useHealth();
  const stream = useStreamState();
  const requestSettings = useSetAtom(settingsOpenRequestAtom);

  const alerts = healthAlerts(health, stream);
  const worst = worstSeverity(alerts);
  if (!worst) return null;

  function describe(alert: Alert) {
    switch (alert.kind) {
      case 'temperature':
        return t(`alerts.temperature.${alert.severity}`, { celsius: Math.round(alert.celsius) });
      case 'storage':
        return t(`alerts.storage.${alert.severity}`, {
          path: alert.path,
          available: formatBytes(alert.available),
          total: formatBytes(alert.total)
        });
      case 'vpn':
        return t('alerts.vpn', { name: alert.title });
      case 'stream':
        return t('alerts.stream');
    }
  }

  const content = (
    <div className="flex max-w-[320px] flex-col space-y-3">
      <span className="text-base font-bold text-neutral-300">{t('alerts.title')}</span>
      {alerts.map((alert) => {
        const Icon = alertIcon[alert.kind];
        return (
          <div
            key={`${alert.kind}-${'title' in alert ? alert.title : ''}`}
            className="flex space-x-2"
          >
            <Icon
              size={16}
              className={clsx(
                'mt-[2px] shrink-0',
                alert.severity === 'critical' ? 'text-red-500' : 'text-amber-400'
              )}
            />
            <div className="flex flex-col items-start text-sm text-neutral-300">
              <span>{describe(alert)}</span>
              {alert.kind === 'vpn' && (
                <Button
                  type="link"
                  size="small"
                  className="px-0"
                  onClick={() => requestSettings('vpn')}
                >
                  {t('alerts.openVpn')}
                </Button>
              )}
            </div>
          </div>
        );
      })}
    </div>
  );

  return (
    <Popover content={content} arrow={false} trigger="click" placement="bottomLeft">
      <Tooltip title={t('alerts.title')} placement="bottom" mouseEnterDelay={0.6}>
        <button
          type="button"
          aria-label={t('alerts.title')}
          className={clsx(
            'flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded p-0 hover:bg-neutral-700/80',
            worst === 'critical' ? 'text-red-500' : 'text-amber-400'
          )}
        >
          <TriangleAlertIcon size={18} />
        </button>
      </Tooltip>
    </Popover>
  );
};
