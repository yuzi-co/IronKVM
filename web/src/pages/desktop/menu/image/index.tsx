import { useEffect, useState } from 'react';
import { Divider, Modal, Tooltip } from 'antd';
import clsx from 'clsx';
import { useSetAtom } from 'jotai';
import { DiscIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import { submenuOpenCountAtom } from '@/jotai/settings.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { Drives } from './drives.tsx';
import { Images } from './images.tsx';
import { Tips } from './tips.tsx';

export const Image = () => {
  const { t } = useTranslation();
  const setSubmenuOpenCount = useSetAtom(submenuOpenCountAtom);

  const [isModalOpen, setIsModalOpen] = useState(false);
  const [drives, setDrives] = useState<api.Drive[]>([]);
  const [diskRo, setDiskRo] = useState(false);

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
            onDrivesChanged={refreshDrives}
          />
        </div>
      </Modal>
    </>
  );
};
