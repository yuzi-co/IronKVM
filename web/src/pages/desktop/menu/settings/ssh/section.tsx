import type { ReactNode } from 'react';

type SectionProps = {
  title: ReactNode;
  description?: ReactNode;
  // Beside the title, such as a switch.
  action?: ReactNode;
  children?: ReactNode;
};

// Section is one block of the SSH page: a title, a line saying what it is
// for, and its content below.
export const Section = ({ title, description, action, children }: SectionProps) => (
  <section className="flex flex-col space-y-3">
    <div className="flex items-center justify-between gap-4">
      <div className="flex min-w-0 flex-col space-y-1">
        <span className="text-sm font-medium">{title}</span>
        {description && <span className="text-xs text-neutral-500">{description}</span>}
      </div>
      {action}
    </div>
    {children}
  </section>
);

// Box is the bordered panel the settings pages put values in.
export const Box = ({ children }: { children: ReactNode }) => (
  <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
    {children}
  </div>
);
