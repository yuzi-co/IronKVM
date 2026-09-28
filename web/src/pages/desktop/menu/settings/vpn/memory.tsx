import { Progress } from 'antd';
import { useTranslation } from 'react-i18next';

import { formatBytes } from './format.ts';
import type { Memory } from './types.ts';

type MemoryBarsProps = {
  memory: Memory;
};

// MemoryBars puts the daemon and the whole addons group against the group's
// memory.high, where the kernel starts to throttle, and names memory.max,
// where it kills.
export const MemoryBars = ({ memory }: MemoryBarsProps) => {
  const { t } = useTranslation();

  const limit = memory.groupHigh || memory.groupMax;
  const percent = (value: number) => (limit > 0 ? Math.min(100, (value / limit) * 100) : 0);
  const pressed = memory.groupHigh > 0 && memory.groupCurrent >= memory.groupHigh * 0.9;
  const limits = [
    memory.groupHigh > 0 ? t('settings.vpn.high', { size: formatBytes(memory.groupHigh) }) : '',
    memory.groupMax > 0 ? t('settings.vpn.max', { size: formatBytes(memory.groupMax) }) : ''
  ].filter(Boolean);

  return (
    <div className="flex flex-col space-y-2">
      <span>{t('settings.vpn.memory')}</span>

      <Bar
        label={t('settings.vpn.daemonRss')}
        value={memory.daemonRss}
        percent={percent(memory.daemonRss)}
        hasLimit={limit > 0}
      />

      {limit > 0 ? (
        <>
          <Bar
            label={t('settings.vpn.group')}
            value={memory.groupCurrent}
            percent={percent(memory.groupCurrent)}
            warn={pressed}
            hasLimit
          />
          <span className="text-xs text-neutral-500">{limits.join(', ')}</span>
        </>
      ) : (
        <span className="text-xs text-neutral-500">{t('settings.vpn.noGroup')}</span>
      )}
    </div>
  );
};

type BarProps = {
  label: string;
  value: number;
  percent: number;
  warn?: boolean;
  hasLimit: boolean;
};

const Bar = ({ label, value, percent, warn, hasLimit }: BarProps) => (
  <div className="flex flex-col">
    <div className="flex justify-between text-sm">
      <span className="text-neutral-400">{label}</span>
      <span className="text-neutral-300">{formatBytes(value)}</span>
    </div>
    {hasLimit && (
      <Progress
        percent={percent}
        showInfo={false}
        size="small"
        strokeColor={warn ? '#f59e0b' : undefined}
      />
    )}
  </div>
);
