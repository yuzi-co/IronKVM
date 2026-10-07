import { Slider, Switch, Tooltip } from 'antd';
import clsx from 'clsx';
import { useAtom, useAtomValue } from 'jotai';
import { MicIcon, Volume2Icon, VolumeXIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { switchRoomMic } from '@/lib/room-mic.ts';
import { audioMutedAtom, audioStateAtom, hasAudioAtom } from '@/jotai/audio.ts';
import {
  roomMicListeningAtom,
  roomMicMutedAtom,
  roomMicSwitchAtom,
  roomMicVolumeAtom
} from '@/jotai/room-mic.ts';
import { MenuItem } from '@/components/menu-item.tsx';

import { useRoomMicControls } from './use-room-mic.ts';

// HostSpeaker is the host's audio as it always was: one button that mutes and
// unmutes it.
const HostSpeaker = () => {
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

// MuteButton is a small mute toggle for one row of the audio menu.
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

// AudioMenu holds the host's audio and the room microphone, each with its own
// mute, so either can be heard without the other.
const AudioMenu = ({ hasHostAudio }: { hasHostAudio: boolean }) => {
  const { t } = useTranslation();
  const [isHostMuted, setIsHostMuted] = useAtom(audioMutedAtom);
  const isHostIdle = useAtomValue(audioStateAtom) === 'idle';
  const [roomSwitch, setRoomSwitch] = useAtom(roomMicSwitchAtom);
  const isListening = useAtomValue(roomMicListeningAtom);
  const [isRoomMuted, setIsRoomMuted] = useAtom(roomMicMutedAtom);
  const [roomVolume, setRoomVolume] = useAtom(roomMicVolumeAtom);

  return (
    <div className="flex w-[260px] flex-col space-y-3 py-1">
      {hasHostAudio && (
        <div className="flex flex-col space-y-1">
          <div className="flex items-center justify-between">
            <span className="text-sm text-neutral-300">{t('speaker.host')}</span>
            <MuteButton muted={isHostMuted} onToggle={() => setIsHostMuted(!isHostMuted)} />
          </div>
          {isHostIdle && <span className="text-xs text-neutral-500">{t('speaker.hostIdle')}</span>}
        </div>
      )}

      <div className="flex flex-col space-y-2">
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
            <MuteButton muted={isRoomMuted} onToggle={() => setIsRoomMuted(!isRoomMuted)} />
            <Slider
              className="flex-1"
              aria-label={t('speaker.roomVolume')}
              min={0}
              max={100}
              step={5}
              value={Math.round(roomVolume * 100)}
              tooltip={{ formatter: (value) => `${value}%` }}
              onChange={(value) => setRoomVolume(value / 100)}
            />
          </div>
        )}

        <span className="text-xs text-neutral-500">{t('speaker.roomHint')}</span>
      </div>
    </div>
  );
};

export const Speaker = () => {
  const { t } = useTranslation();
  const hasHostAudio = useAtomValue(hasAudioAtom);
  const isHostMuted = useAtomValue(audioMutedAtom);
  const roomControls = useRoomMicControls();

  // Without the microphone on offer the entry stays the host's mute button.
  if (!roomControls) {
    return hasHostAudio ? <HostSpeaker /> : null;
  }

  let icon = <MicIcon size={18} />;
  if (hasHostAudio) {
    icon = isHostMuted ? <VolumeXIcon size={18} /> : <Volume2Icon size={18} />;
  }

  return (
    <MenuItem
      title={t('speaker.title')}
      icon={icon}
      content={<AudioMenu hasHostAudio={hasHostAudio} />}
    />
  );
};
