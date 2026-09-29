import { useAuth } from '@/contexts/auth.ts';
import { Switch } from 'antd';
import clsx from 'clsx';
import { useAtom } from 'jotai';
import {
  DiscIcon,
  FileJsonIcon,
  MaximizeIcon,
  NetworkIcon,
  PowerIcon,
  TerminalSquareIcon,
  TypeIcon,
  Volume2Icon,
  WrenchIcon,
  XIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as ls from '@/lib/localstorage.ts';
import { menuDisabledItemsAtom } from '@/jotai/settings.ts';
import { Robot } from '@/components/icons/robot.tsx';

export const MenuIcons = () => {
  const { t } = useTranslation();
  const { account } = useAuth();

  const [menuDisabledItems, setMenuDisabledItems] = useAtom(menuDisabledItemsAtom);

  // In bar order. Scripts, Wake on LAN and PicoClaw are entries of the Tools
  // menu, so they are indented under it; Tools leaves the bar on its own once
  // all of them are hidden.
  const items = [
    // The speaker also needs an audio track before it appears. This switch
    // hides it on a device that has one, the way every other icon can be
    // hidden.
    { key: 'speaker', icon: <Volume2Icon size={16} /> },
    { key: 'text', icon: <TypeIcon size={16} />, label: 'menu.text' },
    { key: 'media', icon: <DiscIcon size={16} />, label: 'menu.media' },
    { key: 'terminal', icon: <TerminalSquareIcon size={16} /> },
    { key: 'tools', icon: <WrenchIcon size={16} />, label: 'menu.tools' },
    { key: 'script', icon: <FileJsonIcon size={16} />, inTools: true },
    { key: 'wol', icon: <NetworkIcon size={16} />, inTools: true },
    { key: 'picoclaw', icon: <Robot size={16} />, inTools: true },
    { key: 'power', icon: <PowerIcon size={16} /> },
    { key: 'fullscreen', icon: <MaximizeIcon size={16} />, label: 'fullscreen.toggle' },
    { key: 'collapse', icon: <XIcon size={16} />, label: 'menu.collapse' }
  ].filter(
    (item) =>
      account.role === 'admin' || !['media', 'terminal', 'script', 'picoclaw'].includes(item.key)
  );

  function updateItems(key: string) {
    const exist = menuDisabledItems.includes(key);

    const newItems = exist
      ? menuDisabledItems.filter((item) => item !== key)
      : [...menuDisabledItems, key];

    setMenuDisabledItems(newItems);
    ls.setMenuDisabledItems(newItems);
  }

  return (
    <div className="mt-8 flex flex-col space-y-5">
      <div className="flex flex-col">
        <span className="text-neutral-400">{t('settings.appearance.menuBar.icons')}</span>
        <span className="text-xs text-neutral-500">
          {t('settings.appearance.menuBar.iconsDesc')}
        </span>
      </div>

      <div className="mt-5 flex flex-col space-y-5">
        {items.map((item) => (
          <div key={item.key} className="flex items-center justify-between">
            <div
              className={clsx(
                'flex items-center space-x-2 text-neutral-400',
                item.inTools && 'pl-6'
              )}
            >
              {item.icon}
              <span className="text-neutral-300">
                {item.label ? t(item.label) : t(`${item.key}.title`)}
              </span>
            </div>

            <Switch
              value={!menuDisabledItems.includes(item.key)}
              onChange={() => updateItems(item.key)}
            />
          </div>
        ))}
      </div>
    </div>
  );
};
