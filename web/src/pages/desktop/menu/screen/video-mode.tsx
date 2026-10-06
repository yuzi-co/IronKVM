import { Popover, Tooltip } from 'antd';
import { useAtomValue } from 'jotai';
import { CheckIcon, TvMinimalPlayIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { setVideoMode as setCookie } from '@/lib/localstorage.ts';
import { videoModeLabel } from '@/lib/stream-codec.ts';
import { screenSettingsAtom, videoModeAtom } from '@/jotai/screen.ts';

// The two H.264 paths carry whichever codec the encoder runs, so their labels
// name that codec (#84). The keys stay as they are: they are stored per viewer.
const videoModes = ['direct', 'h264', 'mjpeg'];

// Neither fact changes while the page is open, so both are read once rather
// than in an effect that renders the component a second time.
const isDirectSupported = window.location.protocol === 'https:' && !!window.VideoDecoder;

export const VideoMode = () => {
  const { t } = useTranslation();
  const videoMode = useAtomValue(videoModeAtom);
  const codec = useAtomValue(screenSettingsAtom)?.codec;

  function update(mode: string) {
    if (mode === videoMode) return;

    setCookie(mode);

    // reload after changing video mode
    setTimeout(() => {
      window.location.reload();
    }, 500);
  }

  const content = (
    <>
      {!isDirectSupported && (
        <Tooltip
          title={t('screen.videoDirectTips')}
          placement="right"
          styles={{ root: { maxWidth: '270px' } }}
        >
          <div className="flex cursor-not-allowed items-center rounded py-1.5 pr-5 pl-1 text-neutral-500 select-none hover:bg-neutral-700/70">
            <div className="flex h-[14px] w-[20px] items-end text-blue-500"></div>
            <span>{videoModeLabel('direct', codec)}</span>
          </div>
        </Tooltip>
      )}

      {videoModes.map(
        (mode) =>
          (isDirectSupported || mode !== 'direct') && (
            <button
              type="button"
              key={mode}
              className="flex w-full cursor-pointer items-center rounded p-0 py-1.5 pr-5 pl-1 text-left select-none hover:bg-neutral-700/70"
              onClick={() => update(mode)}
            >
              <div className="flex h-[14px] w-[20px] items-end text-blue-500">
                {mode === videoMode && <CheckIcon size={14} />}
              </div>
              <span>{videoModeLabel(mode, codec)}</span>
            </button>
          )
      )}
    </>
  );

  return (
    <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
      <div className="flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700/70">
        <TvMinimalPlayIcon size={18} />
        <span className="text-sm select-none">{t('screen.video')}</span>
      </div>
    </Popover>
  );
};
