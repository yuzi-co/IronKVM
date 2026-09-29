import { useSetAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import { picoclawChatOpenAtom } from '@/jotai/picoclaw.ts';
import { Robot } from '@/components/icons/robot.tsx';
import { menuRowClassName } from '@/components/menu-item.tsx';

// Picoclaw is a row of the Tools menu that opens and closes the chat.
export const Picoclaw = () => {
  const { t } = useTranslation();
  const setIsChatOpen = useSetAtom(picoclawChatOpenAtom);

  return (
    <button
      type="button"
      className={menuRowClassName}
      onClick={() => setIsChatOpen((open) => !open)}
    >
      <Robot size={18} />
      <span className="text-sm select-none">{t('picoclaw.title')}</span>
    </button>
  );
};
