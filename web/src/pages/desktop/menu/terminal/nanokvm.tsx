import { SquareTerminalIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export const Nanokvm = () => {
  const { t } = useTranslation();

  function openTerminal() {
    window.open('/#terminal', '_blank');
  }

  return (
    <button
      type="button"
      className="flex h-[28px] w-full cursor-pointer items-center space-x-1 rounded p-0 px-2 py-1 text-left select-none hover:bg-neutral-700/70"
      onClick={openTerminal}
    >
      <SquareTerminalIcon size={14} />
      <span>{t('terminal.nanokvm')}</span>
    </button>
  );
};
