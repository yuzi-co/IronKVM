import { Slider, Switch, Tooltip } from 'antd';
import { useAtom, useAtomValue } from 'jotai';
import { MicIcon, Volume2Icon, VolumeXIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { switchRoomMic } from '@/lib/room-mic.ts';
import {
  roomMicListeningAtom,
  roomMicMutedAtom,
  roomMicSwitchAtom,
  roomMicVolumeAtom
} from '@/jotai/room-mic.ts';
import { MenuItem } from '@/components/menu-item.tsx';

// MuteButton mutes the microphone for this viewer only.
const MuteButton = ({ muted, onToggle }: { muted: boolean; onToggle: () => void }) => {
  const { t } = useTranslation();
  const label = muted ? t('speaker.unmute') : t('speaker.mute');

  return (
    <Tooltip title={label} mouseEnterDelay={0.6}>
      <button
        type="button"
        aria-label={label}
        className="flex h-[26px] w-[26px] cursor-pointer items-center justify-center rounded text-neutral-300 hover:bg-neutral-700/80 hover:text-white"
        onClick={onToggle}
      >
        {muted ? <VolumeXIcon size={16} /> : <Volume2Icon size={16} />}
      </button>
    </Tooltip>
  );
};

// RoomMicControls is the switch, this viewer's own mute and volume.
const RoomMicControls = () => {
  const { t } = useTranslation();
  const [roomSwitch, setRoomSwitch] = useAtom(roomMicSwitchAtom);
  const isListening = useAtomValue(roomMicListeningAtom);
  const [isMuted, setIsMuted] = useAtom(roomMicMutedAtom);
  const [volume, setVolume] = useAtom(roomMicVolumeAtom);

  return (
    <div className="flex w-[260px] flex-col space-y-2 py-1">
      <div className="flex items-center justify-between">
        <span className="flex items-center space-x-2 text-sm text-neutral-300">
          <MicIcon size={16} className={isListening ? 'text-red-500' : undefined} />
          <span>{t('speaker.room')}</span>
        </span>
        <Switch
          size="small"
          aria-label={t('speaker.roomListen')}
          checked={roomSwitch.wanted}
          loading={roomSwitch.awaiting}
          onChange={(on) => setRoomSwitch(switchRoomMic(on))}
        />
      </div>

      {roomSwitch.wanted && (
        <div className="flex items-center space-x-2">
          <MuteButton muted={isMuted} onToggle={() => setIsMuted(!isMuted)} />
          <Slider
            className="flex-1"
            aria-label={t('speaker.roomVolume')}
            min={0}
            max={100}
            step={5}
            value={Math.round(volume * 100)}
            tooltip={{ formatter: (value) => `${value}%` }}
            onChange={(value) => setVolume(value / 100)}
          />
        </div>
      )}

      <span className="text-xs text-neutral-500">{t('speaker.roomHint')}</span>
    </div>
  );
};

// RoomMic is the room microphone's own toolbar entry, apart from the
// speaker, which is the host's audio only. The menu shows it whenever the
// microphone is on offer, whether or not the speaker icon is hidden.
export const RoomMic = () => {
  const { t } = useTranslation();
  const isListening = useAtomValue(roomMicListeningAtom);

  return (
    <MenuItem
      title={t('speaker.room')}
      icon={<MicIcon size={18} className={isListening ? 'text-red-500' : undefined} />}
      content={<RoomMicControls />}
    />
  );
};
