import { forwardRef } from 'react';
import type { ComponentPropsWithoutRef, ReactNode } from 'react';
import clsx from 'clsx';

import { menuRowClassName } from '@/components/menu-item.tsx';

type MenuRowProps = Omit<ComponentPropsWithoutRef<'button'>, 'children'> & {
  icon: ReactNode;
  label: ReactNode;
  // After the label, such as a help icon.
  extra?: ReactNode;
  // At the far end of the row, such as an arrow for a link.
  end?: ReactNode;
};

// MenuRow is one row of the VPN page's menu, drawn like a toolbar menu's row.
// It forwards its ref and props so Popconfirm and Tooltip can attach to it.
export const MenuRow = forwardRef<HTMLButtonElement, MenuRowProps>(
  ({ icon, label, extra, end, className, ...props }, ref) => (
    <button
      ref={ref}
      type="button"
      className={clsx(
        menuRowClassName,
        'disabled:cursor-not-allowed disabled:opacity-50',
        className
      )}
      {...props}
    >
      {icon}
      <span className="text-sm select-none">{label}</span>
      {extra}
      {end && <span className="flex flex-1 justify-end">{end}</span>}
    </button>
  )
);
MenuRow.displayName = 'MenuRow';
