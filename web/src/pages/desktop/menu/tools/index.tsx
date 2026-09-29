import { ReactNode } from 'react';
import { WrenchIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { MenuItem } from '@/components/menu-item.tsx';

// Tools gathers the entries used now and then rather than every session:
// scripts, Wake on LAN and PicoClaw. Each is a row that keeps its own popover
// or panel. The menu bar decides which rows this account may see.
export const Tools = ({ children }: { children: ReactNode }) => {
  const { t } = useTranslation();

  const content = <div className="flex min-w-[180px] flex-col space-y-1">{children}</div>;

  return <MenuItem title={t('menu.tools')} icon={<WrenchIcon size={18} />} content={content} />;
};
