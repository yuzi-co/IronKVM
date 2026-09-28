import { Tooltip } from 'antd';
import clsx from 'clsx';
import { useTranslation } from 'react-i18next';

import { useKeyboardLedStatus } from './use-keyboard-led-status';

type LockIndicatorProps = {
  labelKey: 'numLock' | 'capsLock' | 'scrollLock';
  active: boolean;
  known: boolean;
};

function LockIndicator({ labelKey, active, known }: LockIndicatorProps) {
  const { t } = useTranslation();

  const label = t(`settings.keyboardLedStatus.${labelKey}`);
  const shortLabel = t(`settings.keyboardLedStatus.${labelKey}Short`);
  const state = t(`settings.keyboardLedStatus.${known ? (active ? 'on' : 'off') : 'unknown'}`);
  const indicatorLabel = t('settings.keyboardLedStatus.indicatorLabel', { label, state });

  return (
    <Tooltip title={indicatorLabel} placement="bottom" mouseEnterDelay={0.6}>
      <div
        className="flex h-[11px] items-center gap-1 px-1 text-[10px] leading-[11px] font-medium text-neutral-400"
        aria-label={indicatorLabel}
        role="img"
      >
        <span
          aria-hidden="true"
          className={clsx(
            'flex h-2 w-2 shrink-0 items-center justify-center rounded-full text-[7px] leading-none',
            known
              ? active
                ? 'bg-emerald-400'
                : 'bg-neutral-600'
              : 'border border-dashed border-neutral-500 text-neutral-300'
          )}
        >
          {!known && '?'}
        </span>
        <span className="hidden sm:inline">{shortLabel}</span>
      </div>
    </Tooltip>
  );
}

export function KeyboardLedStatus() {
  const { t } = useTranslation();
  const status = useKeyboardLedStatus();
  const known = status?.known ?? false;

  return (
    <div
      className="flex h-full min-w-[40px] flex-col items-start justify-center rounded bg-neutral-800/80 pr-0.5"
      aria-label={t('settings.keyboardLedStatus.groupLabel')}
      role="group"
    >
      <LockIndicator labelKey="numLock" active={status?.numLock ?? false} known={known} />
      <LockIndicator labelKey="capsLock" active={status?.capsLock ?? false} known={known} />
      <LockIndicator labelKey="scrollLock" active={status?.scrollLock ?? false} known={known} />
    </div>
  );
}
