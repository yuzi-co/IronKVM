import type { ReactNode } from 'react';
import { Popconfirm } from 'antd';
import { useTranslation } from 'react-i18next';

type PowerButtonProps = {
  // The question to ask first, or null to act on the click.
  confirm: string | null;
  onPress: () => void;
  children: ReactNode;
};

// PowerButton is one row of the power menu.
export const PowerButton = ({ confirm, onPress, children }: PowerButtonProps) => {
  const { t } = useTranslation();

  const row = (
    <button
      type="button"
      className="flex w-full cursor-pointer items-center space-x-2 rounded p-0 px-3 py-1.5 text-left select-none hover:bg-neutral-700/70"
      onClick={confirm ? undefined : onPress}
    >
      {children}
    </button>
  );

  if (!confirm) return row;

  return (
    <Popconfirm
      placement="bottomLeft"
      title={confirm}
      okText={t('power.okBtn')}
      cancelText={t('power.cancelBtn')}
      onConfirm={onPress}
      color="#404040"
    >
      {row}
    </Popconfirm>
  );
};
