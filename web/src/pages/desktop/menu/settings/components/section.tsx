import type { ReactNode } from 'react';
import clsx from 'clsx';

type SectionHeaderProps = {
  title: ReactNode;
  description?: ReactNode;
  // Beside the title, such as a switch.
  action?: ReactNode;
  className?: string;
};

// SectionHeader is the one heading style of the settings pages: the title in
// text-sm font-medium, an optional line saying what the section is for, and
// an optional control on the right.
export const SectionHeader = ({ title, description, action, className }: SectionHeaderProps) => (
  <div className={clsx('flex items-center justify-between gap-4', className)}>
    <div className="flex min-w-0 flex-col space-y-1">
      <span className="text-sm font-medium">{title}</span>
      {description && <span className="text-xs text-neutral-500">{description}</span>}
    </div>
    {action}
  </div>
);

type SectionProps = Omit<SectionHeaderProps, 'className'> & {
  // Wider gaps between the rows below the header.
  loose?: boolean;
  children?: ReactNode;
};

// Section is one block of a settings page: a header and its content below.
export const Section = ({ loose, children, ...header }: SectionProps) => (
  <section className={clsx('flex flex-col', loose ? 'space-y-6' : 'space-y-3')}>
    <SectionHeader {...header} />
    {children}
  </section>
);

// Box is the bordered panel the settings pages put values in.
export const Box = ({ children }: { children: ReactNode }) => (
  <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
    {children}
  </div>
);
