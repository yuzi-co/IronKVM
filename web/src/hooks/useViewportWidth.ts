import { useSyncExternalStore } from 'react';

function subscribe(onChange: () => void) {
  window.addEventListener('resize', onChange);
  return () => window.removeEventListener('resize', onChange);
}

// The viewport width in CSS pixels, kept current as the window resizes or a
// phone turns.
export function useViewportWidth(): number {
  return useSyncExternalStore(subscribe, () => window.innerWidth);
}
