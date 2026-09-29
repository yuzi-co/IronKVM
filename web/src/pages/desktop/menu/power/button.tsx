import type { ReactNode } from 'react';
import { Popconfirm } from 'antd';
import { useTranslation } from 'react-i18next';

type PowerButtonProps = {
  // The question to ask first, or null to act on the click.
  confirm: string | null;
  onPress: () => void;
  // What the press does to the host, in a line under the label, as PiKVM
  // words its ATX menu.
  description?: string;
  children: ReactNode;
};

// PowerButton is one row of the power menu.
export const PowerButton = ({ confirm, onPress, description, children }: PowerButtonProps) => {
  const { t } = useTranslation();

  const row = (
    <button
      type="button"
      className="flex w-full cursor-pointer flex-col rounded p-0 px-3 py-1.5 text-left select-none hover:bg-neutral-700/70"
      onClick={confirm ? undefined : onPress}
    >
      <span className="flex items-center space-x-2">{children}</span>
      {description && (
        <span className="max-w-[240px] pl-6 text-xs text-neutral-500">{description}</span>
      )}
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
