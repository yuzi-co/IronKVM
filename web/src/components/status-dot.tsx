import clsx from 'clsx';

export type DotTone = 'ok' | 'off' | 'active' | 'warning' | 'error';

const toneClass: Record<DotTone, string> = {
  ok: 'bg-green-500',
  off: 'bg-neutral-500',
  active: 'bg-blue-500',
  warning: 'bg-amber-400',
  error: 'bg-red-500'
};

type StatusDotProps = {
  tone: DotTone;
  // Where the dot sits on the icon: the top right corner, or below it for a
  // second light.
  position?: 'top' | 'bottom';
};

// StatusDot is a small light on a toolbar icon, as PiKVM puts LEDs in its
// menu titles. It says nothing on its own: the caller leaves it out when the
// state is unknown, and the menu's tooltip or popover names what it shows.
export const StatusDot = ({ tone, position = 'top' }: StatusDotProps) => (
  <span
    aria-hidden="true"
    className={clsx(
      'pointer-events-none absolute right-[-3px] h-[7px] w-[7px] rounded-full ring-2 ring-neutral-800',
      position === 'top' ? 'top-[-3px]' : 'bottom-[-3px]',
      toneClass[tone]
    )}
  />
);
