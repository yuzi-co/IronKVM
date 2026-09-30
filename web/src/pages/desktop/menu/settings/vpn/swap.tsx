import { Tooltip } from 'antd';
import { ArrowRightIcon, CircleHelpIcon, MemoryStickIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { useSettingsNav } from '../nav-context.ts';
import { MenuRow } from './menu-row.tsx';

// Swap is one setting for the whole board, set under Performance. The daemon
// is what most often runs short of memory, so this menu points there rather
// than keeping a second switch that could disagree with the first.
export const Swap = () => {
  const { t } = useTranslation();
  const { openTab } = useSettingsNav();

  return (
    <MenuRow
      icon={<MemoryStickIcon size={18} />}
      label={t('settings.vpn.swap.title')}
      extra={
        <Tooltip
          title={t('settings.vpn.swap.tip')}
          placement="top"
          styles={{ root: { maxWidth: '400px' } }}
        >
          <CircleHelpIcon className="text-neutral-500" size={14} />
        </Tooltip>
      }
      end={<ArrowRightIcon size={15} className="text-neutral-500" />}
      onClick={() => openTab('performance')}
    />
  );
};
