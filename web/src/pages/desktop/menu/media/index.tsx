import { useEffect, useState } from 'react';
import { Divider, Modal, Tooltip } from 'antd';
import clsx from 'clsx';
import { useSetAtom } from 'jotai';
import { DiscIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import * as ventoyApi from '@/api/ventoy.ts';
import type { VentoyStatus } from '@/api/ventoy.ts';
import { submenuOpenCountAtom } from '@/jotai/settings.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { Drives } from '../image/drives.tsx';
import { Images } from '../image/images.tsx';
import { Tips } from '../image/tips.tsx';
import { Ventoy } from '../image/ventoy.tsx';
import { BootMenu, LibraryTransfer } from './transfer.tsx';
import { useImageTransfer } from './use-image-transfer.ts';

const DRIVES_POLL_MS = 5000;

// Media is one place for the images the virtual drives boot from: what is
// mounted now, the library on the device and the ways to add to it, and the
// boot menu download. It replaced separate Image and Download buttons, which
// both worked on the same library.
export const Media = () => {
  const { t } = useTranslation();
  const setSubmenuOpenCount = useSetAtom(submenuOpenCountAtom);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [drives, setDrives] = useState<api.Drive[]>([]);
  const [diskRo, setDiskRo] = useState(false);
  const [images, setImages] = useState<string[]>([]);
  const [ventoy, setVentoy] = useState<VentoyStatus | null>(null);
  const transfer = useImageTransfer();

  const isMounted = drives.some((drive) => !!drive.file);

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

  // The drive list is read once for the menu icon, then again on every open
  // and every few seconds while the dialog is up, since a script, another tab
  // or the host itself can change what a drive holds.
  useEffect(() => {
    refreshDrives();
  }, [refreshDrives]);

  useEffect(() => {
    if (!isModalOpen) return;

    refreshDrives();
    const timer = window.setInterval(refreshDrives, DRIVES_POLL_MS);
    return () => window.clearInterval(timer);
  }, [isModalOpen, refreshDrives]);

  useKeyboardLock('media-modal', isModalOpen);

  // The Ventoy status follows the drive list while the modal is open, since an
  // eject from the drive list takes the Ventoy disk out too.
  useEffect(() => {
    if (!isModalOpen) return;

    ventoyApi
      .getVentoyStatus()
      .then((rsp) => {
        if (rsp.code === 0) setVentoy(rsp.data);
      })
      .catch(() => setVentoy(null));
  }, [isModalOpen, drives]);

  // Images on the Ventoy disk are in use while that disk is in a drive, so they
  // cannot be deleted, as an image in a drive cannot.
  const ventoyInUse = ventoy?.inDrive ? (ventoy.images ?? []) : [];

  const heading = 'text-xs font-medium tracking-wide text-neutral-500 uppercase select-none';

  function toggleModal(open: boolean) {
    setIsModalOpen(open);
    transfer.handleOpenChange(open);
    setSubmenuOpenCount((count) => (open ? count + 1 : Math.max(0, count - 1)));
  }

  return (
    <>
      <Tooltip title={t('menu.media')} placement="bottom" mouseEnterDelay={0.6}>
        <button
          type="button"
          aria-label={t('menu.media')}
          className={clsx(
            'flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded p-0 hover:bg-neutral-700',
            isMounted ? 'text-blue-500' : 'text-neutral-300 hover:text-white'
          )}
          onClick={() => toggleModal(true)}
        >
          <DiscIcon size={18} />
        </button>
      </Tooltip>

      <Modal open={isModalOpen} footer={null} onCancel={() => toggleModal(false)}>
        <div className="flex items-center space-x-1">
          <span className="text-xl font-bold">{t('menu.media')}</span>
          <Tips />
        </div>

        <Divider style={{ margin: '24px 0 16px 0' }} />

        <div className="flex flex-col space-y-4">
          <span className={heading}>{t('menu.mediaMounted')}</span>
          <Drives
            drives={drives}
            diskRo={diskRo}
            setDiskRo={setDiskRo}
            onDrivesChanged={refreshDrives}
          />

          {/* The status stays empty for an account that may not use Ventoy. */}
          {ventoy && (
            <Ventoy
              status={ventoy}
              images={images}
              onStatusChanged={setVentoy}
              onDrivesChanged={refreshDrives}
            />
          )}

          <Divider style={{ margin: '8px 0 0 0' }} />

          <span className={heading}>{t('menu.mediaLibrary')}</span>
          <Images
            isOpen={isModalOpen}
            drives={drives}
            diskRo={diskRo}
            inUse={ventoyInUse}
            onImagesChanged={setImages}
            onDrivesChanged={refreshDrives}
          />
          <LibraryTransfer transfer={transfer} />

          {transfer.diskEnabled && (
            <>
              <Divider style={{ margin: '8px 0 0 0' }} />

              <span className={heading}>{t('menu.mediaBoot')}</span>
              <BootMenu transfer={transfer} />
            </>
          )}
        </div>
      </Modal>
    </>
  );
};
