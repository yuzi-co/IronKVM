import { atom } from 'jotai';

import { getHidMode } from '@/api/hid.ts';

export type HidMode = {
  mode: 'normal' | 'hid-only';
  // the gadget declares the Consumer and System Control reports
  extendedKeys: boolean;
  // the gadget declares the touch screen (/boot/usb.touch)
  touch: boolean;
  // /boot/disable_hid leaves the keyboard and both pointers out of the gadget
  hidDisabled: boolean;
};

// The HID mode as GET /api/hid/mode last reported it, null until the first
// answer. Every menu reads this one copy, so a mode switch made in one menu
// reaches the others when the switch refreshes it.
export const hidModeAtom = atom<HidMode | null>(null);

let pending: Promise<void> | null = null;

// Fetches the HID mode into hidModeAtom. Callers that ask while a fetch is
// running share it rather than start another. After a mode switch that fetch
// may predate the switch, so a caller passes after=true to fetch once more
// when it ends.
export const refreshHidModeAtom = atom(null, (_get, set, after?: boolean) => {
  const fetchHidMode = () =>
    getHidMode()
      .then((rsp) => {
        if (rsp.code !== 0) {
          console.log(rsp.msg);
          return;
        }

        set(hidModeAtom, {
          mode: rsp.data.mode,
          extendedKeys: rsp.data.extendedKeys === true,
          touch: rsp.data.touch === true,
          hidDisabled: rsp.data.hidDisabled === true
        });
      })
      .catch((err) => {
        console.log(err);
      });

  if (pending && !after) return pending;

  const current: Promise<void> = (pending ? pending.then(fetchHidMode) : fetchHidMode()).finally(
    () => {
      if (pending === current) pending = null;
    }
  );
  pending = current;

  return current;
});
