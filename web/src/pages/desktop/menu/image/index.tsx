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
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { Drives } from './drives.tsx';
import { Images } from './images.tsx';
import { Tips } from './tips.tsx';
import { Ventoy } from './ventoy.tsx';

export const Image = () => {
  const { t } = useTranslation();
  const setSubmenuOpenCount = useSetAtom(submenuOpenCountAtom);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [drives, setDrives] = useState<api.Drive[]>([]);
  const [diskRo, setDiskRo] = useState(false);
  const [images, setImages] = useState<string[]>([]);
  const [ventoy, setVentoy] = useState<VentoyStatus | null>(null);

  const isMounted = drives.some((drive) => !!drive.file);

  const refreshDrives = useStableCallback(() => {
    api.getDrives().then((rsp) => {
      if (rsp.code !== 0) return;

      const list: api.Drive[] = rsp.data?.drives ?? [];
      setDrives(list);

      // A loaded disk shows its real flag. An empty one keeps the operator's
      // choice for the next insert.
      const disk = list.find((drive) => drive.id === 'disk');
      if (disk?.file) {
        setDiskRo(disk.ro);
      }
    });
  });

  useEffect(() => {
    refreshDrives();
  }, [refreshDrives]);

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

  function toggleModal(open: boolean) {
    setIsModalOpen(open);
    setSubmenuOpenCount((count) => (open ? count + 1 : Math.max(0, count - 1)));
  }

  return (
    <>
      <Tooltip title={t('image.title')} placement="bottom" mouseEnterDelay={0.6}>
        <div
          className={clsx(
            'flex h-[30px] w-[30px] cursor-pointer items-center justify-center rounded hover:bg-neutral-700',
            isMounted ? 'text-blue-500' : 'text-neutral-300 hover:text-white'
          )}
          onClick={() => toggleModal(true)}
        >
          <DiscIcon size={18} />
        </div>
      </Tooltip>

      <Modal open={isModalOpen} footer={null} onCancel={() => toggleModal(false)}>
        <div className="flex items-center space-x-1">
          <span className="text-xl font-bold">{t('image.title')}</span>
          <Tips />
        </div>

        <Divider style={{ margin: '24px 0' }} />

        <div className="flex flex-col space-y-6">
          <Drives
            drives={drives}
            diskRo={diskRo}
            setDiskRo={setDiskRo}
            onDrivesChanged={refreshDrives}
          />

          <Divider style={{ margin: '24px 0 0 0' }} />

          <Images
            isOpen={isModalOpen}
            drives={drives}
            diskRo={diskRo}
            inUse={ventoyInUse}
            onImagesChanged={setImages}
            onDrivesChanged={refreshDrives}
          />

          {/* The status stays empty for an account that may not use Ventoy. */}
          {ventoy && (
            <>
              <Divider style={{ margin: '24px 0 0 0' }} />

              <Ventoy
                status={ventoy}
                images={images}
                onStatusChanged={setVentoy}
                onDrivesChanged={refreshDrives}
              />
            </>
          )}
        </div>
      </Modal>
    </>
  );
};
