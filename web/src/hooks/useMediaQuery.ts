import { useSyncExternalStore } from 'react';

interface WidthRange {
  minWidth?: number;
  maxWidth?: number;
}

// Whether the viewport width is inside the range, in CSS pixels, kept current
// as the window resizes. Both bounds are inclusive, as in a CSS media query.
export function useMediaQuery({ minWidth, maxWidth }: WidthRange): boolean {
  const conditions = [];
  if (minWidth !== undefined) conditions.push(`(min-width: ${minWidth}px)`);
  if (maxWidth !== undefined) conditions.push(`(max-width: ${maxWidth}px)`);
  const query = conditions.join(' and ') || 'all';

  return useSyncExternalStore(
    (onChange) => {
      const list = window.matchMedia(query);
      list.addEventListener('change', onChange);
      return () => list.removeEventListener('change', onChange);
    },
    () => window.matchMedia(query).matches
  );
}
