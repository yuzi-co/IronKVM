import type { ReactNode } from 'react';

type MenuActionProps = {
  description: string;
  children: ReactNode;
};

// MenuAction frames a control written for a toolbar menu, a full-width row
// that names itself and opens its own dialog, so it sits in a settings page
// the way PowerLedSetting does: in a bordered box, with a line saying what it
// is for. The control itself is used as it is.
export const MenuAction = ({ description, children }: MenuActionProps) => (
  <div className="flex flex-col space-y-1.5">
    <div className="rounded-lg border border-neutral-700/60 p-1">{children}</div>
    <span className="px-1 text-xs text-neutral-500">{description}</span>
  </div>
);
