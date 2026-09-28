import clsx from 'clsx';
import { useSetAtom } from 'jotai';
import { KeyboardIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { isKeyboardOpenAtom } from '@/jotai/keyboard.ts';

export const VirtualKeyboard = () => {
  const { t } = useTranslation();
  const setIsKeyboardOpen = useSetAtom(isKeyboardOpenAtom);

  return (
    <button
      type="button"
      className={clsx(
        'flex w-full cursor-pointer items-center space-x-2 rounded p-0 py-1 pr-5 pl-2 text-left select-none hover:bg-neutral-700/70'
      )}
      onClick={() => setIsKeyboardOpen((o) => !o)}
    >
      <KeyboardIcon size={18} />
      <span>{t('keyboard.virtual')}</span>
    </button>
  );
};
