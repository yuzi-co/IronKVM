import { RotateCwIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { PowerButton } from './button.tsx';
import { press } from './press.ts';

type ResetProps = {
  isLoading: boolean;
  setIsLoading: (loading: boolean) => void;
};

// Reset always asks, whatever the confirmation switch says: it throws away
// everything the host has not saved.
export const Reset = ({ isLoading, setIsLoading }: ResetProps) => {
  const { t } = useTranslation();

  function reset() {
    if (isLoading) return;
    press('reset', 800, t, setIsLoading);
  }

  return (
    <PowerButton
      confirm={t('power.resetConfirm')}
      onPress={reset}
      description={t('power.resetDesc')}
    >
      <RotateCwIcon size={18} />
      <span>{t('power.reset')}</span>
    </PowerButton>
  );
};
