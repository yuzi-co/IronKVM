import { message } from 'antd';
import { useTranslation } from 'react-i18next';

import { sendKey } from '@/api/hid.ts';
import { VIEW_ONLY_CODE } from '@/lib/view-only.ts';
import { useHidMode } from '@/hooks/useHidMode.ts';

export type ExtendedKeysState = 'hidden' | 'disabled' | 'available';

// The media and power keys exist only when the USB gadget declares them, which
// normal mode does and hid-only mode does not. The server reads that from the
// gadget, so the buttons are hidden rather than left to fail.
//
// With /boot/disable_hid the gadget has no HID at all. The buttons are then
// shown disabled, so the owner learns why they do nothing instead of wondering
// where they went.
export function useExtendedKeys(): ExtendedKeysState {
  const hidMode = useHidMode();

  if (!hidMode || hidMode.mode !== 'normal') return 'hidden';
  if (hidMode.hidDisabled) return 'disabled';
  return hidMode.extendedKeys ? 'available' : 'hidden';
}

// useSendKey returns a function that presses one Consumer or System Control
// key and says so when the board refuses it or cannot be reached.
export function useSendKey() {
  const { t } = useTranslation();

  return async (page: 'consumer' | 'system', usage: number) => {
    try {
      const rsp = await sendKey(page, usage);
      if (rsp.code !== 0) {
        console.log(rsp.msg);
        message.error(rsp.code === VIEW_ONLY_CODE ? rsp.msg : t('input.keyFailed'));
      }
    } catch (err) {
      console.log(err);
      message.error(t('input.keyFailed'));
    }
  };
}
