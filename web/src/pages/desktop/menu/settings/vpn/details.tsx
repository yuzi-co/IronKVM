import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

import { formatUptime } from './format.ts';
import type { Status } from './types.ts';

type DetailsProps = {
  status: Status;
};

export const Details = ({ status }: DetailsProps) => {
  const { t } = useTranslation();

  const rows: [string, ReactNode][] = [
    [
      t('settings.vpn.control'),
      status.control ? (
        <span className="text-green-500">{t('settings.vpn.connected')}</span>
      ) : (
        <span className="text-red-400">{t('settings.vpn.disconnected')}</span>
      )
    ],
    [t('settings.vpn.deviceName'), status.name || '-'],
    [t('settings.vpn.deviceIP'), status.ip || '-'],
    [t('settings.vpn.account'), status.account || '-'],
    [t('settings.vpn.version'), status.version || '-'],
    [t('settings.vpn.uptime'), formatUptime(status.uptimeSec)]
  ];

  return (
    <div className="flex flex-col space-y-3">
      {rows.map(([label, value]) => (
        <div key={label} className="flex justify-between">
          <span>{label}</span>
          <span className="text-neutral-300">{value}</span>
        </div>
      ))}
    </div>
  );
};
