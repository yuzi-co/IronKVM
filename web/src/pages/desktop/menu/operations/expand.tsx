import { Divider } from 'antd';
import { ChevronRightIcon, GripVerticalIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

interface ExpandProps {
  toggleMenu: (expanded: boolean) => void;
}

export const Expand = ({ toggleMenu }: ExpandProps) => {
  const { t } = useTranslation();

  return (
    <div className="flex items-center rounded-lg bg-neutral-800/50 p-1">
      <strong>
        <div className="flex size-[26px] cursor-move items-center justify-center text-neutral-500 select-none">
          <GripVerticalIcon size={18} />
        </div>
      </strong>

      <Divider type="vertical" style={{ margin: '0 4px' }} />

      <button
        type="button"
        aria-label={t('menu.expand')}
        className="flex size-[26px] cursor-pointer items-center justify-center rounded p-0 text-neutral-500 hover:bg-neutral-800/60 hover:text-white"
        onClick={() => toggleMenu(true)}
      >
        <ChevronRightIcon size={18} />
      </button>
    </div>
  );
};
