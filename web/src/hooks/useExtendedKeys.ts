import { useEffect, useState } from 'react';

import { getHidMode } from '@/api/hid.ts';

// The media and power keys exist only when the USB gadget declares them, which
// normal mode does and hid-only mode does not. The server reads that from the
// gadget, so the buttons are hidden rather than left to fail.
export function useExtendedKeys(): boolean {
  const [available, setAvailable] = useState(false);

  useEffect(() => {
    let active = true;
    getHidMode()
      .then((rsp) => {
        if (active && rsp.code === 0) {
          setAvailable(rsp.data.extendedKeys === true);
        }
      })
      .catch(() => {});
    return () => {
      active = false;
    };
  }, []);

  return available;
}
