import { PowerIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { PowerButton } from './button.tsx';
import { press } from './press.ts';

type PowerShortProps = {
  showConfirm: boolean;
  isLoading: boolean;
  setIsLoading: (loading: boolean) => void;
};

export const PowerShort = ({ showConfirm, isLoading, setIsLoading }: PowerShortProps) => {
  const { t } = useTranslation();

  function power() {
    if (isLoading) return;
    press('power', 800, t, setIsLoading);
  }

  return (
    <PowerButton confirm={showConfirm ? t('power.powerConfirm') : null} onPress={power}>
      <PowerIcon size={16} />
      <span>{t('power.powerShort')}</span>
    </PowerButton>
  );
};
