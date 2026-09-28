import { Popover } from 'antd';
import { useAtom } from 'jotai';
import { ArrowDownUpIcon, CheckIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as storage from '@/lib/localstorage.ts';
import { scrollDirectionAtom } from '@/jotai/mouse.ts';

export const Direction = () => {
  const { t } = useTranslation();

  const [scrollDirection, setScrollDirection] = useAtom(scrollDirectionAtom);

  const directions = [
    { name: t('mouse.scrollUp'), value: -1 },
    { name: t('mouse.scrollDown'), value: 1 }
  ];

  function update(direction: number): void {
    const value = Number(direction);

    setScrollDirection(value);
    storage.setMouseScrollDirection(value);
  }

  const content = (
    <>
      {directions.map((direction) => (
        <button
          type="button"
          key={direction.value}
          className="flex w-full cursor-pointer items-center space-x-1 rounded p-0 py-1.5 pr-5 pl-2 text-left hover:bg-neutral-700/70"
          onClick={() => update(direction.value)}
        >
          <div className="flex h-[16px] w-[16px] items-end text-blue-500">
            {direction.value === scrollDirection && <CheckIcon strokeWidth={3} size={16} />}
          </div>
          <span>{direction.name}</span>
        </button>
      ))}
    </>
  );

  return (
    <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
      <div className="flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700/70">
        <ArrowDownUpIcon size={18} />
        <span>{t('mouse.direction')}</span>
      </div>
    </Popover>
  );
};
