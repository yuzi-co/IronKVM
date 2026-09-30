import { useEffect, useState, type MouseEvent as ReactMouseEvent } from 'react';
import { Button, notification, Popconfirm, Tooltip, Typography } from 'antd';
import clsx from 'clsx';
import {
  ArrowBigDownDashIcon,
  ArrowBigUpDashIcon,
  DiscIcon,
  HardDriveIcon,
  LoaderCircleIcon,
  PackageIcon,
  PackageSearchIcon,
  Trash2Icon,
  TriangleAlertIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import { formatBytes } from '@/lib/health.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { CD_MAX_BYTES, imageWarnings } from './warnings.ts';

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
  const [loadFailed, setLoadFailed] = useState(false);
  const [images, setImages] = useState<string[]>([]);
  const [sizes, setSizes] = useState<Record<string, number>>({});
  const [busyImage, setBusyImage] = useState('');
  const [targets, setTargets] = useState<Record<string, api.DriveId>>({});
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
  // A failed first try is retried once after a second, quietly: the list is
  // often asked for right as the server restarts, and a Retry button for
  // that is noise.
  const getImages = useStableCallback((force = false) => {
    if (isLoading && !force) return;
    setIsLoading(true);

    const fetchList = () =>
      api.getImages().then((rsp) => (rsp.code === 0 ? rsp : Promise.reject(rsp)));

    fetchList()
      .catch(() => new Promise((resolve) => setTimeout(resolve, 1000)).then(fetchList))
      .then((rsp) => {
        setLoadFailed(false);
        const files: string[] = rsp.data?.files?.length > 0 ? rsp.data.files : [];
        setImages(files);
        setSizes(rsp.data?.sizes ?? {});
        onImagesChanged(files);
        onDrivesChanged();
      })
      .catch(() => setLoadFailed(true))
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
      .catch((err) => openNotification(!!loaded, err?.message ?? ''))
      .finally(() => {
        setBusyImage('');
        onDrivesChanged();
      });
  }

  // delete image, once its confirmation is accepted
  function deleteImage(image: string) {
    if (isLocked(image) || !!deletingImage) return;
    setDeletingImage(image);

    api
      .deleteImage(image)
      .then((rsp) => {
        if (rsp.code !== 0) {
          notify.open({ message: t('image.deleteFailed'), description: rsp.msg, duration: 10 });
          return;
        }

        getImages();
      })
      .catch((err) => {
        notify.open({ message: t('image.deleteFailed'), description: err?.message, duration: 10 });
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

  // A list that could not be read is not an empty one.
  if (loadFailed) {
    return (
      <div className="flex items-center justify-center space-x-2 py-5 text-red-500">
        <span className="text-sm">{t('image.loadFailed')}</span>
        <Button size="small" onClick={() => getImages()}>
          {t('image.retry')}
        </Button>
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
          // Warnings for the drive a click would insert into; a loaded image's
          // warnings show under its drive above.
          const warnings = loaded ? [] : imageWarnings(sizes[image], drive);
          const warningText = warnings
            .map((warning) =>
              t(`image.warning.${warning}`, {
                max: formatBytes(CD_MAX_BYTES),
                size: formatBytes(sizes[image] ?? 0)
              })
            )
            .join(' ');

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

              {warnings.length > 0 && (
                <Tooltip title={warningText}>
                  <TriangleAlertIcon
                    size={16}
                    aria-label={warningText}
                    className="shrink-0 text-amber-400"
                  />
                </Tooltip>
              )}

              {available.length > 0 && (
                <Tooltip
                  title={
                    loaded
                      ? t('image.loadedIn', { drive: driveName })
                      : t('image.insertInto', { drive: driveName })
                  }
                  mouseEnterDelay={0.6}
                >
                  <button
                    type="button"
                    aria-label={
                      loaded
                        ? t('image.loadedIn', { drive: driveName })
                        : t('image.insertInto', { drive: driveName })
                    }
                    className={clsx(
                      'flex h-[24px] w-[24px] items-center justify-center rounded p-0',
                      !loaded && available.length > 1 && 'hover:bg-neutral-500/50'
                    )}
                    onClick={(e) => (loaded ? e.stopPropagation() : toggleTarget(e, image))}
                  >
                    <DriveIcon size={16} />
                  </button>
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

              {/* The confirmation is a popover rather than a dialog, so it
                  counts as part of the Media menu and does not close it. The
                  wrapper keeps clicks on the button and in the confirmation,
                  which bubble through React's tree, from reaching the row and
                  inserting the image. */}
              <div onClick={(e) => e.stopPropagation()}>
                <Popconfirm
                  placement="left"
                  title={t('image.attention')}
                  description={
                    <div className="flex max-w-[260px] flex-col">
                      <span>{t('image.deleteConfirm')}</span>
                      <Typography.Text code className="break-all">
                        {image}
                      </Typography.Text>
                    </div>
                  }
                  okText={t('image.okBtn')}
                  cancelText={t('image.cancelBtn')}
                  okButtonProps={{ danger: true }}
                  disabled={isLocked(image) || deletingImage !== ''}
                  onConfirm={() => deleteImage(image)}
                  color="#404040"
                >
                  <Tooltip
                    title={isLocked(image) ? t('image.inUse') : t('image.delete')}
                    mouseEnterDelay={0.6}
                  >
                    <button
                      type="button"
                      aria-label={isLocked(image) ? t('image.inUse') : t('image.delete')}
                      aria-disabled={isLocked(image)}
                      className={clsx(
                        'flex h-[24px] w-[24px] items-center justify-center rounded p-0 hover:bg-neutral-500/50',
                        isLocked(image)
                          ? 'cursor-not-allowed text-neutral-500'
                          : 'text-neutral-300 hover:text-red-500'
                      )}
                    >
                      {deletingImage === image ? (
                        <LoaderCircleIcon className="animate-spin text-red-500" size={16} />
                      ) : (
                        <Trash2Icon size={15} />
                      )}
                    </button>
                  </Tooltip>
                </Popconfirm>
              </div>
            </div>
          );
        })}
      </div>

      {contextHolder}
    </>
  );
};
