import { Tooltip } from 'antd';
import { useAtomValue } from 'jotai';
import { MicIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { showRoomMicIndicator } from '@/lib/room-mic.ts';
import { roomMicStatusAtom } from '@/jotai/room-mic.ts';

// RoomMicIndicator is shown to every viewer while the board's microphone is
// open, whoever switched it on. The microphone has no light of its own, so
// this is the only sign anyone has that the room is being heard.
export const RoomMicIndicator = () => {
  const { t } = useTranslation();
  const status = useAtomValue(roomMicStatusAtom);

  if (!status || !showRoomMicIndicator(status)) return null;

  const names = status.listeners.join(', ');
  const title = names ? t('speaker.roomLiveBy', { names }) : t('speaker.roomLive');

  return (
    <Tooltip title={title} placement="bottom">
      <div
        role="status"
        aria-label={title}
        className="mr-1 flex h-[24px] shrink-0 items-center space-x-1 rounded bg-red-600 px-2 text-xs font-medium text-white select-none"
      >
        <MicIcon size={14} className="animate-pulse" />
        <span className="whitespace-nowrap">{t('speaker.roomLive')}</span>
      </div>
    </Tooltip>
  );
};
