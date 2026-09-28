import { useEffect } from 'react';
import { useSetAtom } from 'jotai';

import { keyboardLockAtom } from '@/jotai/keyboard.ts';

// useKeyboardLock stops key forwarding to the host while active is true, so
// typing and Esc in a dialog stay in the browser. The lock is released when
// active turns false or the component unmounts, so a dialog that disappears
// without its close handler cannot leave the keyboard dead.
export function useKeyboardLock(source: string, active: boolean) {
  const setKeyboardLock = useSetAtom(keyboardLockAtom);

  useEffect(() => {
    if (!active) return;

    setKeyboardLock({ source, locked: true });
    return () => setKeyboardLock({ source, locked: false });
  }, [source, active, setKeyboardLock]);
}
