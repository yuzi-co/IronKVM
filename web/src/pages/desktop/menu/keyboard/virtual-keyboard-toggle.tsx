import { Tooltip } from 'antd';
import clsx from 'clsx';
import { useAtom } from 'jotai';
import { KeyboardIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { isKeyboardOpenAtom } from '@/jotai/keyboard.ts';

// VirtualKeyboardToggle opens and closes the on-screen keyboard in one tap.
// The narrow bar shows it beside the Keyboard menu: on a phone the on-screen
// keyboard is the only way to type on the host, and two taps through a menu
// each time is one too many.
export const VirtualKeyboardToggle = () => {
  const { t } = useTranslation();
  const [isKeyboardOpen, setIsKeyboardOpen] = useAtom(isKeyboardOpenAtom);

  return (
    <Tooltip title={t('keyboard.onScreen')} placement="bottom" mouseEnterDelay={0.6}>
      <button
        type="button"
        aria-label={t('keyboard.onScreen')}
        aria-pressed={isKeyboardOpen}
        className={clsx(
          'flex size-9 cursor-pointer items-center justify-center rounded p-0 hover:bg-neutral-700/80 hover:text-white',
          isKeyboardOpen ? 'text-blue-400' : 'text-neutral-300'
        )}
        onClick={() => setIsKeyboardOpen((open) => !open)}
      >
        <KeyboardIcon size={18} />
      </button>
    </Tooltip>
  );
};
