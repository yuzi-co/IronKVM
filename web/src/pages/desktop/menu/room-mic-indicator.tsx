import { useAuth } from '@/contexts/auth.ts';
import { Tooltip } from 'antd';
import { useAtomValue } from 'jotai';
import { MicIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { roomMicIndicatorTitle, showRoomMicIndicator } from '@/lib/room-mic.ts';
import { roomMicStatusAtom } from '@/jotai/room-mic.ts';

// RoomMicIndicator is shown to every viewer while the board's microphone is
// open, whoever switched it on. The microphone has no light of its own, so
// this is the only sign anyone has that the room is being heard. Only an
// administrator is told who is listening.
export const RoomMicIndicator = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const status = useAtomValue(roomMicStatusAtom);

  if (!status || !showRoomMicIndicator(status)) return null;

  const tip = roomMicIndicatorTitle(status, account.role === 'admin');
  const title =
    tip.key === 'speaker.roomLiveBy'
      ? t('speaker.roomLiveBy', { names: tip.names })
      : t('speaker.roomLive');

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
