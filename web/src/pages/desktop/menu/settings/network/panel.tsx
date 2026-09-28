import type { ReactNode } from 'react';

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
      <div className="px-4 pt-3 pb-1.5">
        <div className="font-semibold text-neutral-100">{title}</div>
        {description && (
          <div className="mt-0.5 text-xs leading-snug text-neutral-500">{description}</div>
        )}
      </div>
      <div>{children}</div>
    </div>
  );
};
