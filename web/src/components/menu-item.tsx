import { ReactNode, useEffect, useRef, useState } from 'react';
import { Popover, Tooltip } from 'antd';
import { useAtomValue, useSetAtom } from 'jotai';
import { useMediaQuery } from 'react-responsive';

import { menuCloseSignalAtom, submenuOpenCountAtom } from '@/jotai/settings.ts';

type MenuItemProps = {
  title: string;
  icon: ReactNode;
  content: ReactNode;
  className?: string;
  fresh?: boolean;
  // asRow draws the entry as a labelled row inside another menu (Tools)
  // rather than as an icon on the bar. Its popover then opens to the side.
  asRow?: boolean;
  onOpenChange?: (open: boolean) => void;
};

// The row look shared by entries that sit inside a menu rather than on the bar.
export const menuRowClassName =
  'box-border flex h-[30px] w-full cursor-pointer items-center space-x-2 rounded p-0 px-3 text-left text-neutral-300 hover:bg-neutral-700/70';

export const MenuItem = ({
  title,
  icon,
  content,
  className,
  fresh,
  asRow,
  onOpenChange
}: MenuItemProps) => {
  const isBigScreen = useMediaQuery({ minWidth: 640 });
  const setSubmenuOpenCount = useSetAtom(submenuOpenCountAtom);
  const menuCloseSignal = useAtomValue(menuCloseSignalAtom);

  const [isPopoverOpen, setIsPopoverOpen] = useState(false);
  const [isTooltipOpen, setIsTooltipOpen] = useState(false);
  const handledCloseSignalRef = useRef(0);

  useEffect(() => {
    if (
      menuCloseSignal === 0 ||
      menuCloseSignal === handledCloseSignalRef.current ||
      !isPopoverOpen
    ) {
      return;
    }

    handledCloseSignalRef.current = menuCloseSignal;
    setIsPopoverOpen(false);
    setIsTooltipOpen(false);
    setSubmenuOpenCount((count) => Math.max(0, count - 1));
    onOpenChange?.(false);
  }, [isPopoverOpen, menuCloseSignal, onOpenChange, setSubmenuOpenCount]);

  function togglePopover(open: boolean) {
    setIsTooltipOpen(false);
    setIsPopoverOpen(open);

    // Update global submenu count
    setSubmenuOpenCount((count) => (open ? count + 1 : Math.max(0, count - 1)));

    if (onOpenChange) {
      onOpenChange(open);
    }
  }

  function toggleTooltip(open: boolean) {
    if (isPopoverOpen) {
      return;
    }
    setIsTooltipOpen(open);
  }

  let placement: 'bottomLeft' | 'bottom' | 'rightTop' = isBigScreen ? 'bottomLeft' : 'bottom';
  if (asRow && isBigScreen) placement = 'rightTop';

  return (
    <Popover
      content={content}
      arrow={false}
      trigger="click"
      placement={placement}
      open={isPopoverOpen}
      onOpenChange={togglePopover}
      fresh={!!fresh}
    >
      {asRow ? (
        <button type="button" aria-expanded={isPopoverOpen} className={menuRowClassName}>
          {icon}
          <span className="text-sm select-none">{title}</span>
        </button>
      ) : (
        <Tooltip
          title={title}
          mouseEnterDelay={0.6}
          placement="bottom"
          open={isTooltipOpen}
          onOpenChange={toggleTooltip}
        >
          <button
            type="button"
            aria-label={title}
            aria-expanded={isPopoverOpen}
            className={
              className
                ? className
                : 'flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded p-0 text-neutral-300 hover:bg-neutral-700/80 hover:text-white'
            }
          >
            {icon}
          </button>
        </Tooltip>
      )}
    </Popover>
  );
};
