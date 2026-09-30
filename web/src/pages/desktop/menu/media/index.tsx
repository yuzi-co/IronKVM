import { useEffect, useState } from 'react';
import { Button, Divider } from 'antd';
import { useSetAtom } from 'jotai';
import { DiscIcon, SettingsIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import * as ventoyApi from '@/api/ventoy.ts';
import type { VentoyStatus } from '@/api/ventoy.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { menuCloseSignalAtom, settingsOpenRequestAtom } from '@/jotai/settings.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';
import { MenuItem } from '@/components/menu-item.tsx';
import { StatusDot } from '@/components/status-dot.tsx';

import { Drives } from '../image/drives.tsx';
import { Images } from '../image/images.tsx';
import { TipsButton, TipsModal } from '../image/tips.tsx';
import { ventoyUsable } from '../image/ventoy-status.ts';
import { Ventoy } from '../image/ventoy.tsx';
import { driveWarnings } from '../image/warnings.ts';
import { AddImage } from './transfer.tsx';
import { useImageTransfer } from './use-image-transfer.ts';

const DRIVES_POLL_MS = 5000;
// While the menu is closed the list is read only for the icon's light, so
// slowly: a read is a few configfs files and a stat.
const DRIVES_IDLE_POLL_MS = 15000;

// Media is one place for the images the virtual drives boot from, during a
// session: what is mounted now, the library on the device, the Ventoy set, and
// adding an image. It replaced separate Image and Download buttons, which both
// worked on the same library. Setup lives in Settings: Ventoy on the Virtual
// media page, the netboot.xyz ISO on the Network boot page.
export const Media = () => {
  const { t } = useTranslation();
  const requestSettings = useSetAtom(settingsOpenRequestAtom);
  const requestMenuClose = useSetAtom(menuCloseSignalAtom);

  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isTipsOpen, setIsTipsOpen] = useState(false);
  const [drives, setDrives] = useState<api.Drive[]>([]);
  const [diskRo, setDiskRo] = useState(false);
  const [images, setImages] = useState<string[]>([]);
  const [ventoy, setVentoy] = useState<VentoyStatus | null>(null);
  const transfer = useImageTransfer();

  const isMounted = drives.some((drive) => !!drive.file);
  const hasWarning = drives.some((drive) => driveWarnings(drive).length > 0);

  const refreshDrives = useStableCallback(() => {
    api
      .getDrives()
      .then((rsp) => {
        if (rsp.code !== 0) return;

        const list: api.Drive[] = rsp.data?.drives ?? [];
        // A poll that finds nothing new keeps the old array, so effects keyed on
        // the list (the Ventoy status) do not run every few seconds.
        setDrives((prev) => (JSON.stringify(prev) === JSON.stringify(list) ? prev : list));

        // A loaded disk shows its real flag. An empty one keeps the operator's
        // choice for the next insert.
        const disk = list.find((drive) => drive.id === 'disk');
        if (disk?.file) {
          setDiskRo(disk.ro);
        }
      })
      // A failed poll keeps the last list; the next one tries again.
      .catch(() => {});
  });

  // The drive list is read for the menu icon's light, and again on every open
  // and every few seconds while the menu is open, since a script, another tab
  // or the host itself can change what a drive holds.
  useEffect(() => {
    refreshDrives();
    return pollWhileVisible(refreshDrives, isMenuOpen ? DRIVES_POLL_MS : DRIVES_IDLE_POLL_MS);
  }, [isMenuOpen, refreshDrives]);

  useKeyboardLock('media-menu', isMenuOpen);

  // The Ventoy status follows the drive list while the menu is open, since an
  // eject from the drive list takes the Ventoy disk out too.
  useEffect(() => {
    if (!isMenuOpen) return;

    ventoyApi
      .getVentoyStatus()
      .then((rsp) => {
        if (rsp.code === 0) setVentoy(rsp.data);
      })
      .catch(() => setVentoy(null));
  }, [isMenuOpen, drives]);

  // Images on the Ventoy disk are in use while that disk is in a drive, so they
  // cannot be deleted, as an image in a drive cannot.
  const ventoyInUse = ventoy?.inDrive ? (ventoy.images ?? []) : [];

  const heading = 'text-xs font-medium tracking-wide text-neutral-500 uppercase select-none';

  // The light on the icon: blue while an image is in a drive, amber when the
  // menu has a warning about it.
  let title = t('menu.media');
  if (isMounted) {
    title = `${title}: ${hasWarning ? t('image.driveWarning') : t('image.driveLoaded')}`;
  }

  // The link closes the menu and opens the Virtual media settings page.
  function openMediaSettings() {
    requestMenuClose((signal) => signal + 1);
    requestSettings('media');
  }

  // The tips dialog sits outside the popover, so the menu closes before it
  // shows rather than staying open behind it.
  function openTips() {
    requestMenuClose((signal) => signal + 1);
    setIsTipsOpen(true);
  }

  // MenuItem keeps the submenu count; this follows the open state for the
  // polling, the keyboard lock and the transfer status.
  const handleOpenChange = useStableCallback((open: boolean) => {
    setIsMenuOpen(open);
    transfer.handleOpenChange(open);
  });

  const icon = (
    <div className="relative flex h-[18px] w-[18px]">
      <DiscIcon size={18} />
      {isMounted && <StatusDot tone={hasWarning ? 'warning' : 'active'} />}
    </div>
  );

  const divider = { margin: '10px 0' };

  // The page scrolls inside the popover, so a short window still reaches the
  // settings link at the bottom.
  const content = (
    <div className="flex max-h-[calc(100dvh-96px)] w-[420px] max-w-[calc(100vw-32px)] flex-col overflow-y-auto">
      <div className="flex items-center space-x-1 px-1">
        <span className="text-base font-bold text-neutral-300">{t('menu.media')}</span>
        <TipsButton onClick={openTips} />
      </div>

      <Divider style={divider} />

      <div className="flex flex-col space-y-3 px-1">
        <span className={heading}>{t('menu.mediaMounted')}</span>
        <Drives
          drives={drives}
          diskRo={diskRo}
          setDiskRo={setDiskRo}
          onDrivesChanged={refreshDrives}
        />

        <Divider style={{ margin: '4px 0 0 0' }} />

        <span className={heading}>{t('menu.mediaLibrary')}</span>
        <Images
          isOpen={isMenuOpen}
          drives={drives}
          diskRo={diskRo}
          inUse={ventoyInUse}
          onImagesChanged={setImages}
          onDrivesChanged={refreshDrives}
        />
        {/* Ventoy shows once it is installed; Settings sets it up. */}
        {ventoyUsable(ventoy) && (
          <Ventoy
            status={ventoy}
            images={images}
            onStatusChanged={setVentoy}
            onDrivesChanged={refreshDrives}
          />
        )}
        <AddImage transfer={transfer} />
      </div>

      <Divider style={{ margin: '4px 0 6px 0' }} />
      <Button
        type="link"
        size="small"
        className="self-start px-1 text-neutral-400"
        icon={<SettingsIcon size={14} />}
        onClick={openMediaSettings}
      >
        {t('menu.mediaSettings')}
      </Button>
    </div>
  );

  return (
    <>
      <MenuItem title={title} icon={icon} content={content} onOpenChange={handleOpenChange} />
      <TipsModal open={isTipsOpen} onClose={() => setIsTipsOpen(false)} />
    </>
  );
};
