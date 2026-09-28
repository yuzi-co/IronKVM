import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';

import { getPort } from '@/lib/service.ts';

import { CopyButton } from '../components/copy-button.tsx';
import { formatUptime } from './format.ts';
import type { Status } from './types.ts';

type DetailsProps = {
  status: Status;
};

// kvmUrl is this page at the VPN address: the same scheme and port, which the
// web server answers on every interface.
function kvmUrl(ip: string) {
  const protocol = window.location.protocol;
  const port = getPort();
  const isDefault =
    (protocol === 'https:' && port === '443') || (protocol === 'http:' && port === '80');
  const host = ip.includes(':') ? `[${ip}]` : ip;
  return `${protocol}//${host}${isDefault ? '' : `:${port}`}`;
}

export const Details = ({ status }: DetailsProps) => {
  const { t } = useTranslation();

  const copyable = (value: string, label: string) =>
    value ? (
      <span className="flex items-center gap-1">
        <span className="font-mono select-all">{value}</span>
        <CopyButton text={value} label={label} />
      </span>
    ) : (
      '-'
    );

  const url = status.ip ? kvmUrl(status.ip) : '';

  const rows: [string, ReactNode][] = [
    [
      t('settings.vpn.control'),
      status.control ? (
        <span className="text-green-500">{t('settings.vpn.connected')}</span>
      ) : (
        <span className="text-red-400">{t('settings.vpn.disconnected')}</span>
      )
    ],
    [t('settings.vpn.deviceName'), copyable(status.name, t('settings.vpn.deviceName'))],
    [t('settings.vpn.deviceIP'), copyable(status.ip, t('settings.vpn.deviceIP'))],
    [
      t('settings.vpn.kvmUrl'),
      url ? (
        <span className="flex items-center gap-1">
          <a href={url} target="_blank" rel="noreferrer" className="font-mono">
            {url}
          </a>
          <CopyButton text={url} label={t('settings.vpn.kvmUrl')} />
        </span>
      ) : (
        '-'
      )
    ],
    [t('settings.vpn.account'), status.account || '-'],
    [t('settings.vpn.version'), status.version || '-'],
    [t('settings.vpn.uptime'), formatUptime(status.uptimeSec)]
  ];

  return (
    <div className="flex flex-col space-y-3">
      {rows.map(([label, value]) => (
        <div key={label} className="flex min-h-[24px] items-center justify-between gap-3">
          <span>{label}</span>
          <span className="min-w-0 text-right break-all text-neutral-300">{value}</span>
        </div>
      ))}
    </div>
  );
};
