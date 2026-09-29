import { atom } from 'jotai';

// menu bar disabled items
export const menuDisabledItemsAtom = atom<string[]>([]);

// track how many submenus are currently open
export const submenuOpenCountAtom = atom(0);

// signal all toolbar menus to close
export const menuCloseSignalAtom = atom(0);

// web title
export const webTitleAtom = atom('');

// menu display mode: 'off' | 'auto' | 'always'
export const menuDisplayModeAtom = atom<string>('auto');

// show the remote keyboard lock-status indicator beside the menu bar
export const keyboardLedStatusVisibleAtom = atom(true);

// A request, from elsewhere in the UI, to open Settings on a given tab: the
// Media dialog's settings link uses it. The Settings button answers it and
// clears it. A tab this account cannot see opens the default one instead.
export const settingsOpenRequestAtom = atom<string | null>(null);

// Whether the Virtual Disk (the USB mass storage the Media dialog mounts into)
// is switched on. null until the menu bar or Settings has asked. Media hides
// from the bar while it is off, and comes back when Settings turns it on.
export const virtualDiskEnabledAtom = atom<boolean | null>(null);
