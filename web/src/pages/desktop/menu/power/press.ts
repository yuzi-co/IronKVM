import { message } from 'antd';
import type { TFunction } from 'i18next';

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
