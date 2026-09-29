import { Tooltip } from 'antd';
import { ArrowRightIcon, CircleHelpIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { useSettingsNav } from '../nav-context.ts';

// Swap is one setting for the whole board, set under Performance. The daemon
// is what most often runs short of memory, so this menu points there rather
// than keeping a second switch that could disagree with the first.
export const Swap = () => {
  const { t } = useTranslation();
  const { openTab } = useSettingsNav();

  return (
    <button
      type="button"
      className="flex h-[40px] w-full cursor-pointer items-center justify-between space-x-6 rounded p-0 px-2 text-left text-neutral-300 hover:bg-neutral-700/70"
      onClick={() => openTab('performance')}
    >
      <div className="flex items-center space-x-1">
        <span>{t('settings.vpn.swap.title')}</span>
        <Tooltip
          title={t('settings.vpn.swap.tip')}
          className="text-neutral-500"
          placement="top"
          styles={{ root: { maxWidth: '400px' } }}
        >
          <CircleHelpIcon size={15} />
        </Tooltip>
      </div>

      <ArrowRightIcon size={15} className="text-neutral-500" />
    </button>
  );
};
