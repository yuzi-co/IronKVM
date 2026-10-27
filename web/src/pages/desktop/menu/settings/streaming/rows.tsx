import type { ReactNode } from 'react';

// A bordered list of label and value rows, as the VNC page shows its session.
export const Rows = ({ rows, footer }: { rows: [string, ReactNode][]; footer?: ReactNode }) => (
  <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
    {rows.map(([label, value], index) => (
      <div
        key={label}
        className={
          index > 0
            ? 'flex items-center justify-between border-t border-neutral-800 px-4 py-2'
            : 'flex items-center justify-between px-4 py-2'
        }
      >
        <span className="text-sm text-neutral-400">{label}</span>
        <span className="truncate pl-4 text-sm text-neutral-300">{value}</span>
      </div>
    ))}
    {footer}
  </div>
);

// SettingRow is a setting's name and help beside its control.
export const SettingRow = ({
  label,
  desc,
  children
}: {
  label: string;
  desc?: string;
  children: ReactNode;
}) => (
  <div className="flex items-center justify-between">
    <div className="flex flex-col space-y-1 pr-4">
      <span className="text-sm">{label}</span>
      {desc && <span className="text-xs text-neutral-500">{desc}</span>}
    </div>
    {children}
  </div>
);
