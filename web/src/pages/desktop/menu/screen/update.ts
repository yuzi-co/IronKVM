import { message } from 'antd';
import type { TFunction } from 'i18next';

import { updateScreen } from '@/api/vm.ts';

// applyScreenSetting sends one stream setting and reports a refusal or a lost
// request, so a click that changed nothing does not look like it worked. It
// resolves to whether the board took the new value.
export async function applyScreenSetting(
  type: string,
  value: number,
  t: TFunction
): Promise<boolean> {
  try {
    const rsp = await updateScreen(type, value);
    if (rsp.code !== 0) {
      message.error(rsp.msg ? `${t('screen.updateFailed')}: ${rsp.msg}` : t('screen.updateFailed'));
      return false;
    }
    return true;
  } catch {
    message.error(t('screen.updateFailed'));
    return false;
  }
}
