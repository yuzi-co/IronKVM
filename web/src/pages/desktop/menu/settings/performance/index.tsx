import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { useServerFeature } from '@/hooks/useServerFeature.ts';

import { CpuFreq } from './cpu-freq.tsx';
import { Memory } from './memory.tsx';
import { Swap } from './swap.tsx';
import { VideoClock } from './video-clock.tsx';
import { Zram } from './zram.tsx';

// The board's memory at a glance, then its CPU, video clock and memory tuning.
export const Performance = () => {
  const { t } = useTranslation();
  // Only the mainline kernel (slot B) can change the video clocks.
  const hasVideoClock = useServerFeature('video-clock') === true;

  return (
    <>
      <div className="text-base">{t('settings.performance.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <Memory />
        <CpuFreq />
        {hasVideoClock && <VideoClock />}
        <Zram />
        <Swap />
      </div>
    </>
  );
};
