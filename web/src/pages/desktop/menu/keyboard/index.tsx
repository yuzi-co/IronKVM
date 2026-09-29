import { KeyboardIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { MenuItem } from '@/components/menu-item.tsx';

import { MediaKeys } from './media-keys.tsx';
import { Shortcuts } from './shortcuts';
import { VirtualKeyboard } from './virtual-keyboard.tsx';

// Paste moved to the Text menu. The leader key is set once, so it is a
// setting rather than a menu entry; its component stays in this folder.
export const Keyboard = () => {
  const { t } = useTranslation();

  const content = (
    <div className="flex flex-col space-y-1">
      <VirtualKeyboard />
      <Shortcuts />
      <MediaKeys />
    </div>
  );

  return (
    <MenuItem title={t('keyboard.title')} icon={<KeyboardIcon size={18} />} content={content} />
  );
};
