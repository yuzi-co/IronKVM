import { ReactNode } from 'react';
import { EllipsisIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { MenuItem } from '@/components/menu-item.tsx';

// On a narrow screen the menu bar holds only the entries used most, and the
// rest move into this popover. Each entry keeps its own popover or dialog,
// which opens from here as it would from the bar.
export const More = ({ children }: { children: ReactNode }) => {
  const { t } = useTranslation();

  const content = (
    <div className="flex max-w-[calc(100vw-48px)] flex-wrap items-center gap-1">{children}</div>
  );

  return <MenuItem title={t('menu.more')} icon={<EllipsisIcon size={18} />} content={content} />;
};
