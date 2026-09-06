import { Popover, Tooltip } from 'antd';
import { useAtomValue } from 'jotai';
import { CheckIcon, TvMinimalPlayIcon, Volume2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { setVideoMode as setCookie } from '@/lib/localstorage.ts';
import { videoModeAtom } from '@/jotai/screen.ts';

const videoModes = [
  { key: 'direct', name: 'H.264 (Direct)' },
  { key: 'h264', name: 'H.264 (WebRTC)' },
  { key: 'mjpeg', name: 'MJPEG' }
];

// Neither fact changes while the page is open, so both are read once rather
// than in an effect that renders the component a second time.
const isDirectSupported = window.location.protocol === 'https:' && !!window.VideoDecoder;

export const VideoMode = () => {
  const { t } = useTranslation();
  const videoMode = useAtomValue(videoModeAtom);

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
            <span>H.264 (Direct)</span>
          </div>
        </Tooltip>
      )}

      {videoModes.map(
        (mode) =>
          (isDirectSupported || mode.key !== 'direct') && (
            <div
              key={mode.key}
              className="flex cursor-pointer items-center rounded py-1.5 pr-5 pl-1 select-none hover:bg-neutral-700/70"
              onClick={() => update(mode.key)}
            >
              <div className="flex h-[14px] w-[20px] items-end text-blue-500">
                {mode.key === videoMode && <CheckIcon size={15} />}
              </div>
              <span>{mode.name}</span>
            </div>
          )
      )}

      {/* The speaker control only appears once a WebRTC audio track arrives,
          so on the other two modes there is nothing to click and nothing to
          say why. The server has one caller of audio.NewStream and it is in
          the WebRTC path: MJPEG is a multipart response and Direct is a
          websocket whose nine-byte frame header has no room to say what a
          message holds, so neither can carry a second stream. */}
      <div className="mt-1 flex max-w-[210px] items-start space-x-1.5 border-t border-neutral-700 pt-1.5 pr-5 pl-1 text-xs text-neutral-500">
        <Volume2Icon className="mt-[2px] shrink-0" size={12} />
        <span className="select-none">{t('screen.videoAudioNote')}</span>
      </div>
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
