import { useState } from 'react';
import { Slider } from 'antd';
import { CirclePowerIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { PowerButton } from './button.tsx';
import { MAX_PRESS_SECONDS, press } from './press.ts';

// Most boards force the power off after four seconds of holding the button.
const DEFAULT_POWER_LONG_DURATION_SECONDS = 5;

type PowerLongProps = {
  isLoading: boolean;
  setIsLoading: (loading: boolean) => void;
};

// A long press always asks, whatever the confirmation switch says: it cuts the
// host's power without letting it shut down.
export const PowerLong = ({ isLoading, setIsLoading }: PowerLongProps) => {
  const { t } = useTranslation();

  const [duration, setDuration] = useState(DEFAULT_POWER_LONG_DURATION_SECONDS);

  function power() {
    if (isLoading) return;
    press('power', duration * 1000, t, setIsLoading);
  }

  return (
    <>
      <PowerButton
        confirm={t('power.powerLongConfirm', { seconds: duration })}
        onPress={power}
        description={t('power.powerLongDesc')}
      >
        <CirclePowerIcon size={16} />
        <span>{t('power.powerLong')}</span>
        <div className="flex h-full items-start text-xs text-neutral-500">{`${duration}s`}</div>
      </PowerButton>

      <div className="px-3">
        <Slider
          defaultValue={DEFAULT_POWER_LONG_DURATION_SECONDS}
          min={1}
          max={MAX_PRESS_SECONDS}
          tooltip={{ placement: 'bottom' }}
          onChange={setDuration}
        />
      </div>
    </>
  );
};
