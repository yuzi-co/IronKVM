import clsx from 'clsx';
import { useTranslation } from 'react-i18next';

type PowerLedStateProps = {
  isPowerOn: boolean;
  connected: boolean;
};

// PowerLedState is the read-only line for the host's power LED as the board
// sees it. Without the LED header wired the line reads "off" whatever the
// host does, so the state is shown as unknown rather than off. The power menu
// shows this; the settings pages show it beside the wiring switch.
export const PowerLedState = ({ isPowerOn, connected }: PowerLedStateProps) => {
  const { t } = useTranslation();

  let state = t('power.ledUnknown');
  if (connected) {
    state = isPowerOn ? t('power.ledOn') : t('power.ledOff');
  }

  return (
    <div className="flex items-center justify-between">
      <span className="text-xs text-neutral-400">{t('power.led')}</span>
      <div className="flex items-center space-x-1.5">
        <span
          className={clsx(
            'h-2 w-2 rounded-full',
            !connected && 'border border-neutral-500',
            connected && isPowerOn && 'bg-green-600',
            connected && !isPowerOn && 'bg-neutral-600'
          )}
        />
        <span className="text-xs text-neutral-300">{state}</span>
      </div>
    </div>
  );
};
