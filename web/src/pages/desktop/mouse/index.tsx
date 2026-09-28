import { useAtomValue } from 'jotai';

import { mouseModeAtom } from '@/jotai/mouse.ts';
import { useTouchAvailable } from '@/hooks/useTouchAvailable.ts';

import { Absolute } from './absolute.tsx';
import { Relative } from './relative.tsx';
import { Touch } from './touch.tsx';

export const Mouse = () => {
  const mouseMode = useAtomValue(mouseModeAtom);
  const touchAvailable = useTouchAvailable();

  if (mouseMode === 'relative') return <Relative />;
  // Touch mode is kept in the browser, so it can outlive the gadget that
  // declared it. The absolute mouse takes over until the touch screen is back.
  if (mouseMode === 'touch' && touchAvailable) return <Touch />;
  return <Absolute />;
};
