import { Tooltip } from 'antd';
import {
  PauseIcon,
  SkipBackIcon,
  SkipForwardIcon,
  SquareIcon,
  Volume1Icon,
  Volume2Icon,
  VolumeXIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import { sendKey } from '@/api/hid.ts';
import { useExtendedKeys } from '@/hooks/useExtendedKeys.ts';

// Consumer page usages (HID Usage Tables, section 15).
const keys = [
  { usage: 0xe2, label: 'keyboard.mediaKeys.mute', Icon: VolumeXIcon },
  { usage: 0xea, label: 'keyboard.mediaKeys.volumeDown', Icon: Volume1Icon },
  { usage: 0xe9, label: 'keyboard.mediaKeys.volumeUp', Icon: Volume2Icon },
  { usage: 0xb6, label: 'keyboard.mediaKeys.previous', Icon: SkipBackIcon },
  { usage: 0xcd, label: 'keyboard.mediaKeys.playPause', Icon: PauseIcon },
  { usage: 0xb5, label: 'keyboard.mediaKeys.next', Icon: SkipForwardIcon },
  { usage: 0xb7, label: 'keyboard.mediaKeys.stop', Icon: SquareIcon }
];

export const MediaKeys = () => {
  const { t } = useTranslation();
  const available = useExtendedKeys();

  if (!available) return null;

  return (
    <div className="px-3 py-1.5">
      <div className="pb-1 text-xs text-neutral-400">{t('keyboard.mediaKeys.title')}</div>
      <div className="flex items-center space-x-1">
        {keys.map(({ usage, label, Icon }) => (
          <Tooltip key={usage} title={t(label)} placement="bottom">
            <div
              className="flex h-7 w-7 cursor-pointer items-center justify-center rounded hover:bg-neutral-700/70"
              onClick={() => sendKey('consumer', usage)}
            >
              <Icon size={16} />
            </div>
          </Tooltip>
        ))}
      </div>
    </div>
  );
};
