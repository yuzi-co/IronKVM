import { useSetAtom } from 'jotai';
import { ClipboardIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { pasteDialogAtom } from '@/jotai/paste.ts';
import { menuRowClassName } from '@/components/menu-item.tsx';

// The dialog itself is mounted with the desktop, so the paste shortcut can open
// it while this menu is closed.
export const Paste = () => {
  const { t } = useTranslation();
  const setDialog = useSetAtom(pasteDialogAtom);

  return (
    <button
      type="button"
      className={menuRowClassName}
      onClick={() => setDialog({ open: true, notice: '', check: null })}
    >
      <ClipboardIcon size={18} />
      <span className="text-sm select-none">{t('keyboard.paste')}</span>
    </button>
  );
};
