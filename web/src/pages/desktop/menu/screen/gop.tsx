import { Popover } from 'antd';
import { CheckIcon, SquareKanbanIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { applyScreenSetting } from './update.ts';

type GopProps = {
  gop: number;
  setGop: (gop: number) => void;
};

const gopList = [
  { key: 10, label: '10' },
  { key: 30, label: '30' },
  { key: 50, label: '50' },
  { key: 100, label: '100' }
];

export const Gop = ({ gop, setGop }: GopProps) => {
  const { t } = useTranslation();

  async function update(value: number) {
    if (value === gop) return;

    if (!(await applyScreenSetting('gop', value, t))) return;

    setGop(value);
  }

  const content = (
    <>
      {gopList.map((item) => (
        <button
          type="button"
          key={item.key}
          className="flex w-full cursor-pointer items-center rounded p-0 py-1 pr-6 pl-1 text-left select-none hover:bg-neutral-700/70"
          onClick={() => update(item.key)}
        >
          <div className="flex h-[14px] w-[20px] items-end text-blue-500">
            {item.key === gop && <CheckIcon size={14} />}
          </div>
          <span>{item.label}</span>
        </button>
      ))}
    </>
  );

  return (
    <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
      <div className="flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700/70">
        <SquareKanbanIcon size={18} />
        <span className="text-sm select-none">GOP</span>
      </div>
    </Popover>
  );
};
