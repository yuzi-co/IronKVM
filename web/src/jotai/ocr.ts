import { atom } from 'jotai';

// Set while the user drags a rectangle over the screen to read the text in it.
// The screen menu sets it, and the overlay that takes the drag clears it.
export const ocrSelectingAtom = atom(false);
