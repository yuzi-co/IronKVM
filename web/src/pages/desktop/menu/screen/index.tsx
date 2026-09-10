import { useEffect } from 'react';
import { useAtom, useAtomValue } from 'jotai';
import { MonitorIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { updateScreen } from '@/api/vm';
import { ScreenSettings } from '@/types';
import { defaultScreenSettings, screenSettingsAtom, videoModeAtom } from '@/jotai/screen.ts';
import { MenuItem } from '@/components/menu-item.tsx';

import { Codec } from './codec.tsx';
import { getQualityMap, getScreenType } from './constants.ts';
import { Fps } from './fps';
import { FrameDetect } from './frame-detect';
import { Gop } from './gop.tsx';
import { Quality } from './quality';
import { Reset } from './reset.tsx';
import { Resolution } from './resolution';
import { Scale } from './scale';
import { VideoMode } from './video-mode.tsx';

// The quality item offers four steps rather than a value, and what those steps
// mean depends on the delivery path: a JPEG quality on MJPEG, a bitrate on both
// H.264 paths. The server reports the two separately because it cannot know
// which one the viewer is asking about.
function qualityKey(videoMode: string, settings: ScreenSettings): number {
  const qualityMap = getQualityMap(videoMode);
  if (!qualityMap) return 2;

  const value = videoMode === 'mjpeg' ? settings.quality : settings.bitRate;
  for (const [key, mapped] of qualityMap) {
    if (mapped === value) return key;
  }

  return 2;
}

// The menu shows what the server holds, and it holds nothing of its own.
//
// Every item here is a device setting: there is one encoder on the board, so a
// change takes effect for every viewer at once. The menu used to keep its own
// copy in localStorage and push it on load, which had all three of the failures
// that arrangement always has. A second browser saw its own settings rather
// than the board's; a setting the server restored from the card was overwritten
// by whatever this browser last remembered; and a refresh reset the menu to a
// value the board was not running, so the tick and the picture disagreed. The
// last of those was the sharp one: the menu ticked H.264 while the encoder sent
// H.265, and clicking H.264 did nothing because the menu believed it was
// already there.
//
// Only videoMode stays in localStorage, because it is a real per-viewer choice:
// which delivery path this browser wants, not what the encoder does.
//
// The read itself happens in the desktop, once, and lands in an atom. Absolute
// mouse positioning needs the capture resolution whether or not anyone opens
// this menu, so asking here as well would be a second request for the same
// answer. Nothing is mirrored into component state: every item below is drawn
// from the atom and writes back to it, so there is one copy of each setting in
// the browser and it is the one the server reported.
export const Screen = () => {
  const { t } = useTranslation();

  const videoMode = useAtomValue(videoModeAtom);
  const [settings, setSettings] = useAtom(screenSettingsAtom);
  const isMjpeg = videoMode === 'mjpeg';

  const current = settings ?? defaultScreenSettings;

  // The capture mode follows this viewer's delivery path, so it is the one
  // setting the browser still pushes. Two viewers on different paths do fight
  // over it, which is inherent to having one encoder and is not new here.
  useEffect(() => {
    const screenType = getScreenType(videoMode);
    if (screenType === null) return;

    updateScreen('type', screenType);
  }, [videoMode]);

  function apply(change: Partial<ScreenSettings>) {
    setSettings((previous) => ({ ...(previous ?? defaultScreenSettings), ...change }));
  }

  function setQuality(key: number) {
    const value = getQualityMap(videoMode)?.get(key);
    if (value === undefined) return;

    apply(isMjpeg ? { quality: value } : { bitRate: value });
  }

  const content = (
    <div className="flex flex-col space-y-1">
      <VideoMode />
      <Resolution />
      <Quality quality={qualityKey(videoMode, current)} setQuality={setQuality} />
      <Fps fps={current.fps} setFps={(fps) => apply({ fps })} />
      <Scale />
      {!isMjpeg && <Gop gop={current.gop} setGop={(gop) => apply({ gop })} />}
      {!isMjpeg && (
        <Codec codec={current.codec} setCodec={(codec) => apply({ codec })} videoMode={videoMode} />
      )}
      {isMjpeg && <FrameDetect />}
      <Reset />
    </div>
  );

  return <MenuItem title={t('screen.title')} icon={<MonitorIcon size={18} />} content={content} />;
};
