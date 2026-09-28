import type { ReactNode } from 'react';
import { message, Popconfirm } from 'antd';
import type { TFunction } from 'i18next';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';

// The longest press the server carries out. server/service/vm/gpio.go caps a
// press at this and holds the line no longer, whatever the request asks for.
export const MAX_PRESS_SECONDS = 10;

// press holds a front panel line and tells the operator whether the board did
// it. The loading flag covers the whole press, so a second click cannot start
// another one.
export function press(
  type: 'reset' | 'power',
  durationMs: number,
  t: TFunction,
  setIsLoading: (loading: boolean) => void
) {
  setIsLoading(true);

  api
    .setGpio(type, durationMs)
    .then((rsp) => {
      if (rsp.code !== 0) {
        message.error(rsp.msg || t('power.failed'));
        return;
      }
      message.success(t('power.done'));
    })
    .catch(() => message.error(t('power.failed')))
    .finally(() => setIsLoading(false));
}

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
    <div
      className="flex cursor-pointer select-none items-center space-x-2 rounded px-3 py-1.5 hover:bg-neutral-700/70"
      onClick={confirm ? undefined : onPress}
    >
      {children}
    </div>
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
