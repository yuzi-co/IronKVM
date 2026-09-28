import { Popover } from 'antd';
import { useAtom } from 'jotai';
import { CheckIcon, SquareDashedMousePointerIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { applyMouseMode } from '@/lib/mouse-mode.ts';
import { mouseModeAtom } from '@/jotai/mouse.ts';
import { useTouchAvailable } from '@/hooks/useTouchAvailable.ts';

export const MouseMode = () => {
  const { t } = useTranslation();

  const [mouseMode, setMouseMode] = useAtom(mouseModeAtom);
  const touchAvailable = useTouchAvailable();

  const mouseModes = [
    { name: t('mouse.absolute'), value: 'absolute' },
    { name: t('mouse.relative'), value: 'relative' }
  ];
  // Touch is offered only when the gadget declares the touch screen, which
  // takes /boot/usb.touch on the device.
  if (touchAvailable) {
    mouseModes.push({ name: t('mouse.touch'), value: 'touch' });
  }

  function shortName() {
    if (mouseMode === 'relative') return t('mouse.relativeShort');
    if (mouseMode === 'touch' && touchAvailable) return t('mouse.touchShort');
    return t('mouse.absoluteShort');
  }

  function updateMouseMode(mode: string) {
    applyMouseMode(mode, setMouseMode);
  }

  const content = (
    <>
      {mouseModes.map((mode) => (
        <div
          key={mode.value}
          className="flex cursor-pointer items-center space-x-1 rounded py-1.5 pl-2 pr-5 hover:bg-neutral-700/70"
          onClick={() => updateMouseMode(mode.value)}
        >
          <div className="flex h-[16px] w-[16px] items-end text-blue-500">
            {mode.value === mouseMode && <CheckIcon strokeWidth={3} size={16} />}
          </div>
          <span>{mode.name}</span>
        </div>
      ))}
    </>
  );

  return (
    <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
      <div className="flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700/70">
        <SquareDashedMousePointerIcon size={18} />
        <span>{t('mouse.mode')}</span>
        {/* The mode is kept in the browser, so a relative mode chosen once
            outlives the reason for it. Showing it here keeps that visible. */}
        <span className="ml-auto pl-3 text-xs text-neutral-500">{shortName()}</span>
      </div>
    </Popover>
  );
};
