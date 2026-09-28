import { useEffect } from 'react';
import { useAtomValue, useSetAtom } from 'jotai';

import { hidModeAtom, refreshHidModeAtom } from '@/jotai/hid.ts';
import type { HidMode } from '@/jotai/hid.ts';

// useHidMode reads the shared HID mode, and fetches it if nothing has yet.
// It returns null until the first answer.
export function useHidMode(): HidMode | null {
  const hidMode = useAtomValue(hidModeAtom);
  const refresh = useSetAtom(refreshHidModeAtom);

  useEffect(() => {
    if (hidMode === null) {
      refresh();
    }
  }, [hidMode, refresh]);

  return hidMode;
}
