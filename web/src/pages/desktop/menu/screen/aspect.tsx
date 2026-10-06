import { Popover, Tooltip } from 'antd';
import { CheckIcon, CircleHelpIcon, ProportionsIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { applyScreenSetting } from './update.ts';

// The values the server stores: keep the source's aspect ratio, or stretch it
// to the selected resolution.
export const aspectKeep = 0;
export const aspectStretch = 1;

type AspectProps = {
  aspect: number;
  setAspect: (aspect: number) => void;
};

// Aspect decides how a host screen whose shape differs from the selected
// resolution is sent. Keep sends it in its own shape at the selected height
// (1920x1200 at 1080 lines is 1728x1080); the picture, the mouse and the input
// region follow the stream's own size, so nothing else in the page changes.
export const Aspect = ({ aspect, setAspect }: AspectProps) => {
  const { t } = useTranslation();

  const items = [
    { key: aspectKeep, label: t('screen.aspectKeep') },
    { key: aspectStretch, label: t('screen.aspectStretch') }
  ];

  async function update(value: number) {
    if (value === aspect) return;

    if (!(await applyScreenSetting('aspect', value, t))) return;

    setAspect(value);
  }

  const content = (
    <>
      {items.map((item) => (
        <button
          type="button"
          key={item.key}
          className="flex w-full cursor-pointer items-center rounded p-0 py-1.5 pr-5 pl-1 text-left select-none hover:bg-neutral-700/70"
          onClick={() => update(item.key)}
        >
          <div className="flex h-[14px] w-[20px] items-end text-blue-500">
            {item.key === aspect && <CheckIcon size={14} />}
          </div>
          <span>{item.label}</span>
        </button>
      ))}
    </>
  );

  return (
    <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
      <div className="flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700/70">
        <ProportionsIcon size={18} />
        <span className="text-sm select-none">{t('screen.aspect')}</span>
        <Tooltip
          title={t('screen.aspectTips')}
          placement="right"
          overlayInnerStyle={{ width: '300px' }}
        >
          <CircleHelpIcon className="text-neutral-500" size={14} />
        </Tooltip>
      </div>
    </Popover>
  );
};
