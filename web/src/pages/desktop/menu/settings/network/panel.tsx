import type { ReactNode } from 'react';

import { SectionHeader } from '../components/section.tsx';

// Panel is a titled card, the frame of the Ethernet and DNS settings.
export const Panel = ({
  title,
  description,
  children
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) => {
  return (
    <div className="overflow-hidden rounded-xl bg-neutral-800/50">
      <SectionHeader title={title} description={description} className="px-4 pt-3 pb-1.5" />
      <div>{children}</div>
    </div>
  );
};
