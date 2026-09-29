import { TypeIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { MenuItem } from '@/components/menu-item.tsx';

import { Ocr } from './ocr.tsx';
import { Paste } from './paste.tsx';

// Text holds both directions of text between this browser and the host:
// typing a paste into the host, and reading text back off its screen.
export const Text = () => {
  const { t } = useTranslation();

  const heading = 'px-3 pt-1 text-xs text-neutral-500 uppercase select-none';

  const content = (
    <div className="flex min-w-[200px] flex-col space-y-1">
      <span className={heading}>{t('menu.textToHost')}</span>
      <Paste />
      <span className={heading}>{t('menu.textFromHost')}</span>
      <Ocr />
    </div>
  );

  return <MenuItem title={t('menu.text')} icon={<TypeIcon size={18} />} content={content} />;
};
