import { Divider } from 'antd';
import { useTranslation } from 'react-i18next';

import { CpuFreq } from './cpu-freq.tsx';
import { Swap } from './swap.tsx';
import { Zram } from './zram.tsx';

// The board's CPU and memory tuning. The VPN page links here for swap rather
// than keeping a second switch for the same setting.
export const Performance = () => {
  const { t } = useTranslation();

  return (
    <>
      <div className="text-base">{t('settings.performance.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-8">
        <CpuFreq />
        <Zram />
        <Swap />
      </div>
    </>
  );
};
