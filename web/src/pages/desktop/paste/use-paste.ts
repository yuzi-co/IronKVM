import { message } from 'antd';
import { useAtomValue, useSetAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/hid.ts';
import { readClipboardText } from '@/lib/clipboard.ts';
import {
  pasteDelayAtom,
  pasteDialogAtom,
  pasteLayoutAtom,
  pasteStatusAtom,
  pasteTextAtom
} from '@/jotai/paste.ts';

// The server's limit, in code points.
export const maxPasteLength = 20000;

// The shortcut that types the browser's clipboard on the host. It is taken
// from the keys sent to the host, so it has to be one a host rarely needs.
export const pasteShortcutLabel = 'Ctrl+Alt+Shift+V';

export function isPasteShortcut(event: KeyboardEvent) {
  return event.code === 'KeyV' && event.ctrlKey && event.altKey && event.shiftKey && !event.metaKey;
}

export function pasteLength(text: string) {
  return Array.from(text).length;
}

// formatDuration writes milliseconds as m:ss, which reads the same in every
// language.
export function formatDuration(ms: number) {
  const seconds = Math.max(1, Math.round(ms / 1000));
  return `${Math.floor(seconds / 60)}:${String(seconds % 60).padStart(2, '0')}`;
}

export function usePaste() {
  const { t } = useTranslation();

  const layout = useAtomValue(pasteLayoutAtom);
  const delay = useAtomValue(pasteDelayAtom);
  const setDialog = useSetAtom(pasteDialogAtom);
  const setText = useSetAtom(pasteTextAtom);
  const setStatus = useSetAtom(pasteStatusAtom);

  // start types text on the host and resolves true once the server is typing
  // it. A text the layout cannot wholly type opens the dialog on the server's
  // report of each character, unless skipUntypeable says to type the rest.
  async function start(text: string, skipUntypeable = false) {
    if (pasteLength(text) > maxPasteLength) {
      setText(text);
      setDialog({
        open: true,
        notice: t('keyboard.pasting.tooLong', { max: maxPasteLength }),
        check: null
      });
      return false;
    }

    try {
      const rsp = await api.paste(text, layout, delay, skipUntypeable);
      switch (rsp.code) {
        case 0:
          setStatus(rsp.data);
          return true;
        case -4:
          setText(text);
          setDialog({ open: true, notice: '', check: rsp.data ?? null });
          return false;
        case -3:
          // Another paste is typing. Showing it gives the user its cancel.
          if (rsp.data) setStatus(rsp.data);
          message.error(t('keyboard.pasting.inProgress'));
          return false;
        case -2:
          message.error(t('keyboard.pasting.tooLong', { max: maxPasteLength }));
          return false;
        default:
          message.error(rsp.msg || t('keyboard.pasting.failed'));
          return false;
      }
    } catch {
      message.error(t('keyboard.pasting.failed'));
      return false;
    }
  }

  // pasteClipboard types the browser's clipboard on the host. A clipboard the
  // browser will not read opens the dialog instead, with the reason, so the
  // text can be pasted into it by hand.
  async function pasteClipboard() {
    const result = await readClipboardText();
    if (!result.ok) {
      const notice =
        result.reason === 'unavailable'
          ? t('keyboard.pasting.clipboardUnavailable')
          : result.reason === 'denied'
            ? t('keyboard.clipboardPermissionDenied')
            : t('keyboard.clipboardReadError');
      setDialog({ open: true, notice, check: null });
      return;
    }

    if (!result.text) {
      message.info(t('keyboard.pasting.clipboardEmpty'));
      return;
    }

    await start(result.text);
  }

  return { start, pasteClipboard };
}
