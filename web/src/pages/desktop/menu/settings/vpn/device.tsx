import clsx from 'clsx';
import { useTranslation } from 'react-i18next';

import { Section } from '../components/section.tsx';
import { Details } from './details.tsx';
import { formatBytes, formatUptime } from './format.ts';
import type { Status } from './types.ts';
import { memoryUse } from './view.ts';

type DeviceProps = {
  status: Status;
};

// Device is this KVM on the VPN: its details, then one quiet line with the
// daemon's version, uptime and memory.
export const Device = ({ status }: DeviceProps) => {
  const { t } = useTranslation();

  const memory = memoryUse(status.memory);
  const memoryText = memory.limit
    ? t('settings.vpn.memoryOf', {
        used: formatBytes(memory.used),
        limit: formatBytes(memory.limit)
      })
    : formatBytes(memory.used);

  return (
    <Section title={t('settings.vpn.thisDevice')}>
      <Details status={status} />

      <div className="flex flex-wrap gap-x-3 text-xs text-neutral-500">
        <span>{`${t('settings.vpn.version')} ${status.version || '-'}`}</span>
        <span>{`${t('settings.vpn.uptime')} ${formatUptime(status.uptimeSec)}`}</span>
        <span
          className={clsx(memory.pressed && 'text-amber-500')}
          title={memory.pressed ? t('settings.vpn.memoryPressed') : undefined}
        >
          {`${t('settings.vpn.memory')} ${memoryText}`}
        </span>
      </div>
    </Section>
  );
};
