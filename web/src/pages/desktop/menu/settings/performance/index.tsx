import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { CpuFreq } from './cpu-freq.tsx';
import { Memory } from './memory.tsx';
import { Swap } from './swap.tsx';
import { Zram } from './zram.tsx';

// The board's memory at a glance, then its CPU and memory tuning.
export const Performance = () => {
  const { t } = useTranslation();

  return (
    <>
      <div className="text-base">{t('settings.performance.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <Memory />
        <CpuFreq />
        <Zram />
        <Swap />
      </div>
    </>
  );
};
