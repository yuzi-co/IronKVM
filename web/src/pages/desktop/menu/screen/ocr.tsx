import { useSetAtom } from 'jotai';
import { ScanTextIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { ocrSelectingAtom } from '@/jotai/ocr.ts';
import { menuCloseSignalAtom } from '@/jotai/settings.ts';

// Ocr starts a selection of the screen to read text from. The overlay and the
// result live with the desktop, because the menu closes when the selection
// starts.
export const Ocr = () => {
  const { t } = useTranslation();
  const setSelecting = useSetAtom(ocrSelectingAtom);
  const requestMenuClose = useSetAtom(menuCloseSignalAtom);

  function start() {
    requestMenuClose((signal) => signal + 1);
    setSelecting(true);
  }

  return (
    <button
      type="button"
      className="flex h-[30px] w-full cursor-pointer items-center space-x-2 rounded p-0 px-3 text-left text-neutral-300 hover:bg-neutral-700/70"
      onClick={start}
    >
      <ScanTextIcon size={18} />
      <span className="text-sm select-none">{t('screen.ocr.title')}</span>
    </button>
  );
};
