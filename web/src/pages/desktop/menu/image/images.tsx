import { useEffect, useState, type MouseEvent as ReactMouseEvent } from 'react';
import { Button, Modal, notification, Tooltip, Typography } from 'antd';
import clsx from 'clsx';
import {
  ArrowBigDownDashIcon,
  ArrowBigUpDashIcon,
  DiscIcon,
  HardDriveIcon,
  LoaderCircleIcon,
  PackageIcon,
  PackageSearchIcon,
  Trash2Icon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

const imageUpdatedEvent = 'nanokvm:image-updated';

type ImagesProps = {
  isOpen: boolean;
  drives: api.Drive[];
  diskRo: boolean;
  inUse: string[];
  onImagesChanged: (images: string[]) => void;
  onDrivesChanged: () => void;
};

// defaultDrive picks where a click inserts an image: an ISO into the CD drive
// and anything else into the disk, falling back to whichever drive exists.
function defaultDrive(image: string, available: api.DriveId[]): api.DriveId {
  const preferred: api.DriveId = image.toLowerCase().endsWith('.iso') ? 'cdrom' : 'disk';
  return available.includes(preferred) ? preferred : available[0];
}

export const Images = ({
  isOpen,
  drives,
  diskRo,
  inUse,
  onImagesChanged,
  onDrivesChanged
}: ImagesProps) => {
  const { t } = useTranslation();
  const [notify, contextHolder] = notification.useNotification();

  const [isLoading, setIsLoading] = useState(false);
  const [images, setImages] = useState<string[]>([]);
  const [busyImage, setBusyImage] = useState('');
  const [targets, setTargets] = useState<Record<string, api.DriveId>>({});
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [selectedImage, setSelectedImage] = useState('');
  const [deletingImage, setDeletingImage] = useState('');

  const available = drives.map((drive) => drive.id);

  function loadedIn(image: string): api.DriveId | undefined {
    return drives.find((drive) => drive.file === image)?.id;
  }

  // An image is locked against deletion while a drive holds it, directly or
  // through the Ventoy disk.
  function isLocked(image: string) {
    return !!loadedIn(image) || inUse.includes(image);
  }

  function targetOf(image: string): api.DriveId {
    const chosen = targets[image];
    return chosen && available.includes(chosen) ? chosen : defaultDrive(image, available);
  }

  // get image list
  //
  // force skips the in-flight guard. The effect below passes it: an update event
  // can arrive while an earlier request is still out, and that request may have
  // been answered before the change it announces.
  const getImages = useStableCallback((force = false) => {
    if (isLoading && !force) return;
    setIsLoading(true);

    api
      .getImages()
      .then((rsp) => {
        if (rsp.code !== 0) {
          return;
        }

        const files: string[] = rsp.data?.files?.length > 0 ? rsp.data.files : [];
        setImages(files);
        onImagesChanged(files);
        onDrivesChanged();
      })
      .finally(() => {
        setIsLoading(false);
      });
  });

  useEffect(() => {
    if (!isOpen) return;

    getImages(true);

    const handleImageUpdated = () => {
      getImages(true);
    };
    window.addEventListener(imageUpdatedEvent, handleImageUpdated);

    return () => {
      window.removeEventListener(imageUpdatedEvent, handleImageUpdated);
    };
  }, [isOpen, getImages]);

  // flip the drive a click will insert into, when both drives exist
  function toggleTarget(e: ReactMouseEvent, image: string) {
    e.stopPropagation();
    if (available.length < 2) return;

    const next: api.DriveId = targetOf(image) === 'cdrom' ? 'disk' : 'cdrom';
    setTargets((prev) => ({ ...prev, [image]: next }));
  }

  // eject the image from the drive holding it, or insert it into its target
  function insertOrEject(image: string) {
    if (busyImage || available.length === 0) return;
    setBusyImage(image);

    const loaded = loadedIn(image);
    const target = targetOf(image);
    const request = loaded
      ? api.ejectDrive(loaded)
      : api.insertDrive(target, image, target === 'disk' ? diskRo : true);

    request
      .then((rsp) => {
        if (rsp.code !== 0) {
          openNotification(!!loaded, rsp.msg);
        }
      })
      .finally(() => {
        setBusyImage('');
        onDrivesChanged();
      });
  }

  // show delete image modal
  function showDeleteModal(e: any, image: string) {
    e.stopPropagation();

    const isDeleting = deletingImage !== '';

    if (isLocked(image) || isDeleting) {
      return;
    }

    setSelectedImage(image);
    setIsModalOpen(true);
  }

  // delete image
  function deleteImage() {
    if (!selectedImage || !!deletingImage) return;
    setDeletingImage(selectedImage);

    setIsModalOpen(false);

    api
      .deleteImage(selectedImage)
      .then((rsp) => {
        if (rsp.code !== 0) {
          notify.open({ message: t('image.deleteFailed'), description: rsp.msg, duration: 10 });
          return;
        }

        getImages();

        setSelectedImage('');
      })
      .finally(() => {
        setDeletingImage('');
      });
  }

  // show insert/eject failed notification
  function openNotification(isEject: boolean, description: string) {
    notify.open({
      message: t(isEject ? 'image.ejectFailed' : 'image.insertFailed'),
      description,
      duration: 10
    });
  }

  // loading
  if (isLoading) {
    return (
      <div className="flex items-center justify-center space-x-2 py-5 text-neutral-400">
        <LoaderCircleIcon className="animate-spin" size={18} />
        <span className="text-sm">{t('image.loading')}</span>
      </div>
    );
  }

  // empty image
  if (images.length === 0) {
    return (
      <div className="flex items-center justify-center space-x-2 py-5 text-neutral-500">
        <PackageSearchIcon size={18} />
        <span className="text-sm">{t('image.empty')}</span>
      </div>
    );
  }

  return (
    <>
      <div className="flex max-h-[400px] flex-col overflow-y-auto pb-2">
        {images.map((image) => {
          const loaded = loadedIn(image);
          const drive = loaded ?? targetOf(image);
          const DriveIcon = drive === 'cdrom' ? DiscIcon : HardDriveIcon;
          const driveName = t(`image.${drive}`);

          return (
            <div
              key={image}
              className={clsx(
                'group flex cursor-pointer items-center space-x-1 rounded px-1 py-2 select-none hover:bg-neutral-700/70',
                loaded && 'text-blue-500'
              )}
              onClick={() => insertOrEject(image)}
            >
              <div className="flex h-[24px] w-[24px] items-center justify-center">
                {busyImage === image ? (
                  <LoaderCircleIcon className="animate-spin" size={18} />
                ) : (
                  <PackageIcon size={18} />
                )}
              </div>

              <div className="flex-1 truncate">{image.replace(/^.*[\\/]/, '')}</div>

              {available.length > 0 && (
                <Tooltip
                  title={
                    loaded
                      ? t('image.loadedIn', { drive: driveName })
                      : t('image.insertInto', { drive: driveName })
                  }
                  mouseEnterDelay={0.6}
                >
                  <div
                    className={clsx(
                      'flex h-[24px] w-[24px] items-center justify-center rounded',
                      !loaded && available.length > 1 && 'hover:bg-neutral-500/50'
                    )}
                    onClick={(e) => (loaded ? e.stopPropagation() : toggleTarget(e, image))}
                  >
                    <DriveIcon size={16} />
                  </div>
                </Tooltip>
              )}

              <div className="flex h-[24px] w-[24px] items-center justify-center rounded">
                {loaded ? (
                  <ArrowBigDownDashIcon
                    size={22}
                    className="hidden text-red-500 group-hover:block"
                  />
                ) : (
                  <ArrowBigUpDashIcon
                    size={22}
                    className="hidden text-blue-500 group-hover:block"
                  />
                )}
              </div>

              <div
                className={clsx(
                  'flex h-[24px] w-[24px] items-center justify-center rounded hover:bg-neutral-500/50',
                  isLocked(image)
                    ? 'cursor-not-allowed text-neutral-500'
                    : 'text-neutral-300 hover:text-red-500'
                )}
                onClick={(e) => showDeleteModal(e, image)}
              >
                {deletingImage === image ? (
                  <LoaderCircleIcon className="animate-spin text-red-500" size={16} />
                ) : (
                  <Trash2Icon size={16} />
                )}
              </div>
            </div>
          );
        })}
      </div>

      <Modal
        title={t('image.attention')}
        open={isModalOpen}
        width={520}
        footer={null}
        onCancel={() => setIsModalOpen(false)}
      >
        <div className="flex flex-col items-center pb-10">
          <p>{t('image.deleteConfirm')}</p>
          <Typography.Text code>{selectedImage}</Typography.Text>
        </div>

        <div className="flex justify-center space-x-3 pb-3">
          <Button type="primary" danger onClick={deleteImage}>
            {t('image.okBtn')}
          </Button>
          <Button onClick={() => setIsModalOpen(false)}>{t('image.cancelBtn')}</Button>
        </div>
      </Modal>

      {contextHolder}
    </>
  );
};
