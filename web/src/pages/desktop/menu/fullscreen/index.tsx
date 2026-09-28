import { useEffect, useState } from 'react';
import { Tooltip } from 'antd';
import { MaximizeIcon, MinimizeIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

export const Fullscreen = () => {
  const { t } = useTranslation();
  const [isFullscreen, setIsFullscreen] = useState(false);

  useEffect(() => {
    const onFullscreenChange = () => {
      setIsFullscreen(!!document.fullscreenElement);
    };

    onFullscreenChange();

    document.addEventListener('fullscreenchange', onFullscreenChange);

    return () => {
      document.removeEventListener('fullscreenchange', onFullscreenChange);
    };
  }, []);

  function handleFullscreen() {
    if (!document.fullscreenElement) {
      const element = document.documentElement;
      element.requestFullscreen();

      // @ts-expect-error - https://developer.mozilla.org/en-US/docs/Web/API/Keyboard/lock
      navigator.keyboard?.lock();
    } else {
      document.exitFullscreen();

      // @ts-expect-error - https://developer.mozilla.org/en-US/docs/Web/API/Keyboard/unlock
      navigator.keyboard?.unlock();
    }
  }

  // Phones used to lose the button to a width breakpoint, though Android
  // browsers go fullscreen fine. What decides it is whether the browser can:
  // Safari on iPhone has no Fullscreen API for a page at all.
  if (!document.fullscreenEnabled) {
    return null;
  }

  return (
    <Tooltip title={t('fullscreen.toggle')} placement="bottom" mouseEnterDelay={0.6}>
      <button
        type="button"
        aria-label={t('fullscreen.toggle')}
        className="flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded p-0 text-neutral-300 hover:bg-neutral-700/80 hover:text-white"
        onClick={handleFullscreen}
      >
        {isFullscreen ? <MinimizeIcon size={18} /> : <MaximizeIcon size={18} />}
      </button>
    </Tooltip>
  );
};
