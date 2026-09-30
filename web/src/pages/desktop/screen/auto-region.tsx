import { useEffect } from 'react';
import { useAtomValue, useSetAtom } from 'jotai';

import { autoRegionDelay } from '@/lib/input-region.ts';
import {
  controlRegionModeAtom,
  inputRegionAtom,
  inputRegionSelectingAtom,
  resolutionAtom
} from '@/jotai/screen.ts';

import { detectFrameContent, getMediaSize } from './geometry.ts';

export const AutoRegion = () => {
  const mode = useAtomValue(controlRegionModeAtom);
  const selecting = useAtomValue(inputRegionSelectingAtom);
  const resolution = useAtomValue(resolutionAtom);
  const setInputRegion = useSetAtom(inputRegionAtom);

  useEffect(() => {
    if (mode !== 'auto' || selecting) {
      return;
    }

    let stopped = false;
    let timer: number | undefined;
    let target: Element | null = null;
    let candidate = '';
    let confirmations = 0;
    const canvas = document.createElement('canvas');

    // Each check schedules the next: fast while the region is settling, slow
    // once it has held, and none while the tab is hidden (coming back checks
    // at once).
    const schedule = () => {
      window.clearTimeout(timer);
      timer = undefined;
      if (stopped || document.visibilityState !== 'visible') return;
      timer = window.setTimeout(run, autoRegionDelay(confirmations));
    };
    const run = () => {
      detect();
      schedule();
    };
    const onVisible = () => {
      if (document.visibilityState === 'visible') run();
      else schedule();
    };

    const detect = () => {
      if (stopped) return;

      const screen = document.getElementById('screen');
      if (!screen) return;
      if (screen !== target) {
        target = screen;
        candidate = '';
        confirmations = 0;
      }

      const mediaSize = getMediaSize(screen, resolution);
      if (!mediaSize) return;
      const content = detectFrameContent(screen, mediaSize, canvas);
      const key = [
        mediaSize.width,
        mediaSize.height,
        content.left,
        content.top,
        content.width,
        content.height
      ]
        .map(Math.round)
        .join(':');
      if (candidate === key) {
        confirmations += 1;
      } else {
        candidate = key;
        confirmations = 1;
      }
      if (confirmations === 3) {
        setInputRegion({
          frameWidth: mediaSize.width,
          frameHeight: mediaSize.height,
          left: Math.round(content.left),
          top: Math.round(content.top),
          width: Math.round(content.width),
          height: Math.round(content.height)
        });
      }
    };

    run();
    document.addEventListener('visibilitychange', onVisible);
    return () => {
      stopped = true;
      window.clearTimeout(timer);
      document.removeEventListener('visibilitychange', onVisible);
    };
  }, [mode, resolution, selecting, setInputRegion]);

  useEffect(() => {
    if (mode !== 'auto') setInputRegion(null);
  }, [mode, setInputRegion]);

  return null;
};
