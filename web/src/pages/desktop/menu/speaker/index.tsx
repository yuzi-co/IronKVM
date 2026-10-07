import { Tooltip } from 'antd';
import clsx from 'clsx';
import { useAtom, useAtomValue } from 'jotai';
import { Volume2Icon, VolumeXIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { audioMutedAtom, audioStateAtom } from '@/jotai/audio.ts';

export const Speaker = () => {
  const { t } = useTranslation();
  const [isMuted, setIsMuted] = useAtom(audioMutedAtom);
  // An idle host sends nothing, so unmuting plays silence. Saying why keeps
  // that from looking like broken audio.
  const isHostIdle = useAtomValue(audioStateAtom) === 'idle';

  const action = isMuted ? t('speaker.unmute') : t('speaker.mute');
  const title = isHostIdle ? (
    <div>
      <div className="font-medium">{t('speaker.hostIdle')}</div>
      <div className="text-neutral-300">{t('speaker.hostIdleHint')}</div>
    </div>
  ) : (
    action
  );

  return (
    <Tooltip title={title} placement="bottom" mouseEnterDelay={isHostIdle ? 0.2 : 0.6}>
      <button
        type="button"
        aria-label={isHostIdle ? `${action}. ${t('speaker.hostIdle')}` : action}
        className={clsx(
          'relative flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded p-0 hover:bg-neutral-700/80 hover:text-white',
          isHostIdle ? 'text-neutral-500' : 'text-neutral-300'
        )}
        onClick={() => setIsMuted(!isMuted)}
      >
        {isMuted ? <VolumeXIcon size={18} /> : <Volume2Icon size={18} />}
        {isHostIdle && (
          <span className="absolute top-[3px] right-[3px] h-[6px] w-[6px] rounded-full bg-amber-400" />
        )}
      </button>
    </Tooltip>
  );
};
