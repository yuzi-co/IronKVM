import { useSetAtom } from 'jotai';
import { ClipboardIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { pasteDialogAtom } from '@/jotai/paste.ts';

// The dialog itself is mounted with the desktop, so the paste shortcut can open
// it while this menu is closed.
export const Paste = () => {
  const { t } = useTranslation();
  const setDialog = useSetAtom(pasteDialogAtom);

  return (
    <button
      type="button"
      className="flex w-full cursor-pointer items-center space-x-2 rounded p-0 py-1 pr-5 pl-2 text-left select-none hover:bg-neutral-700/70"
      onClick={() => setDialog({ open: true, notice: '', check: null })}
    >
      <ClipboardIcon size={18} />
      <span>{t('keyboard.paste')}</span>
    </button>
  );
};
