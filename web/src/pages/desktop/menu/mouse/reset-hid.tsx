import { useState } from 'react';
import { message } from 'antd';
import clsx from 'clsx';
import { RefreshCwIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { resetHid } from '@/lib/hid-reset.ts';

export const ResetHid = () => {
  const { t } = useTranslation();

  const [isResetting, setIsResetting] = useState(false);

  async function reset() {
    if (isResetting) return;
    setIsResetting(true);

    const result = await resetHid();
    if (result.ok) {
      message.success(t('mouse.resetHidDone'));
    } else {
      message.error(result.msg || t('mouse.resetHidFailed'));
    }
    setIsResetting(false);
  }

  return (
    <button
      type="button"
      className="flex h-[30px] w-full cursor-pointer items-center space-x-2 rounded p-0 px-3 text-left text-neutral-300 select-none hover:bg-neutral-700/70"
      onClick={reset}
    >
      <RefreshCwIcon className={clsx({ 'animate-spin text-blue-500': isResetting })} size={18} />
      <span>{t('mouse.resetHid')}</span>
    </button>
  );
};
