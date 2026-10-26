import { useEffect, useRef, useState } from 'react';
import { Button, Modal, notification, Popconfirm, Progress, Switch, Tooltip } from 'antd';
import clsx from 'clsx';
import {
  ArrowUpIcon,
  DownloadIcon,
  FileIcon,
  FolderIcon,
  LoaderCircleIcon,
  RefreshCwIcon,
  Trash2Icon,
  UploadIcon
} from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/storage.ts';
import { useKeyboardLock } from '@/hooks/useKeyboardLock.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

import { breadcrumbs, fileSize, joinPath, parentPath } from './file-paths.ts';

type FilesModalProps = {
  // The image to browse, by its path under /data; '' keeps the dialog closed.
  image: string;
  onClose: () => void;
};

type LoadError = { code: number; msg: string };

// FilesModal browses the files inside an image, and exchanges files with the
// host through it: download what the host left there, upload what it should
// find. It opens read-only. Changes need the switch, and the server allows
// them only while no drive serves the image, since the host and the board
// writing one filesystem corrupts it.
export const FilesModal = ({ image, onClose }: FilesModalProps) => {
  const { t } = useTranslation();
  const [notify, contextHolder] = notification.useNotification();
  const open = image !== '';
  useKeyboardLock('image-files', open);

  const [dir, setDir] = useState('/');
  const [listing, setListing] = useState<api.ImageFiles | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<LoadError | null>(null);
  const [editing, setEditing] = useState(false);
  const [uploadPercent, setUploadPercent] = useState(-1);
  const [deleting, setDeleting] = useState('');
  const fileInput = useRef<HTMLInputElement>(null);
  // The answer to the latest listing request only, so a slow answer for a
  // directory already left does not replace the one now shown.
  const request = useRef(0);

  const load = useStableCallback((path: string) => {
    const id = ++request.current;
    setLoading(true);
    api
      .listFiles(image, path)
      .then((rsp) => {
        if (id !== request.current) return;
        if (rsp.code !== 0) {
          setError({ code: rsp.code, msg: rsp.msg });
          return;
        }
        setError(null);
        setListing(rsp.data as api.ImageFiles);
        setDir(path);
      })
      .catch((err) => {
        if (id === request.current) setError({ code: -2, msg: err?.message ?? '' });
      })
      .finally(() => {
        if (id === request.current) setLoading(false);
      });
  });

  // Every image opens at its root, read-only: the caller keys the dialog by
  // the image, so a new one starts from fresh state.
  useEffect(() => {
    if (open) load('/');
  }, [open, load]);

  const writable = !!listing?.writable;
  const canChange = editing && writable && uploadPercent < 0;

  function readOnlyReason() {
    if (!listing) return '';
    if (listing.drive) {
      return t('image.files.inDrive', { drive: t(`image.${listing.drive}`) });
    }
    if (!listing.writable) return t('image.files.readOnlyFs', { fs: listing.fs });
    return '';
  }

  function upload(selected: File | undefined) {
    if (!selected || !canChange) return;
    setUploadPercent(0);
    api
      .uploadFile(image, dir, selected, setUploadPercent)
      .then((rsp) => {
        if (rsp.code !== 0) {
          notify.open({
            message: t('image.files.uploadFailed'),
            description: rsp.msg,
            duration: 10
          });
        }
      })
      .catch((err) => {
        notify.open({
          message: t('image.files.uploadFailed'),
          description: err?.message,
          duration: 10
        });
      })
      .finally(() => {
        setUploadPercent(-1);
        if (fileInput.current) fileInput.current.value = '';
        load(dir);
      });
  }

  function remove(entry: api.ImageFile) {
    if (!canChange || deleting) return;
    const path = joinPath(dir, entry.name);
    setDeleting(path);
    api
      .deleteFile(image, path)
      .then((rsp) => {
        if (rsp.code !== 0) {
          notify.open({
            message: t('image.files.deleteFailed'),
            description: rsp.msg,
            duration: 10
          });
        }
      })
      .catch((err) => {
        notify.open({
          message: t('image.files.deleteFailed'),
          description: err?.message,
          duration: 10
        });
      })
      .finally(() => {
        setDeleting('');
        load(dir);
      });
  }

  function errorText(err: LoadError) {
    switch (err.code) {
      case api.FILES_NOT_BROWSABLE:
        return t('image.files.notBrowsable');
      case api.FILES_BUSY:
        return t('image.files.busy');
      default:
        return t('image.files.loadFailed');
    }
  }

  const name = image.replace(/^.*[\\/]/, '');
  const reason = readOnlyReason();

  const title = (
    <div className="flex min-w-0 flex-col pr-6">
      <span className="truncate">{name}</span>
      {listing && (
        <span className="text-xs font-normal text-neutral-500">
          {listing.fs.toUpperCase()} ·{' '}
          {t('image.files.free', { free: fileSize(listing.free), total: fileSize(listing.total) })}
        </span>
      )}
    </div>
  );

  return (
    <Modal title={title} open={open} width={640} footer={null} centered onCancel={onClose}>
      {contextHolder}

      {/* The toolbar: where we are, and what can be done here. */}
      <div className="flex flex-wrap items-center gap-2 pb-2">
        <nav className="flex min-w-0 flex-1 flex-wrap items-center text-sm" aria-label="path">
          {breadcrumbs(dir).map((crumb, i) => (
            <span key={crumb.path} className="flex items-center">
              {i > 0 && <span className="px-1 text-neutral-500">/</span>}
              <button
                type="button"
                className="cursor-pointer truncate rounded px-1 text-neutral-300 hover:bg-neutral-700 hover:text-white"
                onClick={() => load(crumb.path)}
              >
                {crumb.name || t('image.files.root')}
              </button>
            </span>
          ))}
        </nav>

        <Tooltip title={t('image.files.up')}>
          <Button
            size="small"
            aria-label={t('image.files.up')}
            icon={<ArrowUpIcon size={14} />}
            disabled={dir === '/' || loading}
            onClick={() => load(parentPath(dir))}
          />
        </Tooltip>
        <Tooltip title={t('image.files.refresh')}>
          <Button
            size="small"
            aria-label={t('image.files.refresh')}
            icon={<RefreshCwIcon size={14} className={clsx(loading && 'animate-spin')} />}
            onClick={() => load(dir)}
          />
        </Tooltip>
      </div>

      <div className="flex flex-wrap items-center gap-3 pb-3">
        <Tooltip title={reason}>
          <span className="flex items-center gap-2 text-sm text-neutral-300">
            <Switch
              size="small"
              checked={editing && writable}
              disabled={!writable}
              onChange={setEditing}
            />
            {t('image.files.allowChanges')}
          </span>
        </Tooltip>

        {canChange && (
          <>
            <input
              ref={fileInput}
              type="file"
              className="hidden"
              onChange={(e) => upload(e.target.files?.[0])}
            />
            <Button
              size="small"
              icon={<UploadIcon size={14} />}
              onClick={() => fileInput.current?.click()}
            >
              {t('image.files.upload')}
            </Button>
          </>
        )}
        {reason && <span className="text-xs text-neutral-500">{reason}</span>}
      </div>

      {uploadPercent >= 0 && <Progress percent={uploadPercent} size="small" />}
      {editing && writable && (
        <div className="pb-2 text-xs text-amber-400">{t('image.files.changesHint')}</div>
      )}

      {error ? (
        <div className="py-6 text-center text-sm text-red-500">
          <div>{errorText(error)}</div>
          {error.msg && <div className="pt-1 text-xs break-all text-neutral-500">{error.msg}</div>}
        </div>
      ) : !listing ? (
        <div className="flex items-center justify-center py-6 text-neutral-400">
          <LoaderCircleIcon className="animate-spin" size={18} />
        </div>
      ) : listing.entries.length === 0 ? (
        <div className="py-6 text-center text-sm text-neutral-500">{t('image.files.empty')}</div>
      ) : (
        <div className="flex max-h-[50vh] flex-col overflow-y-auto">
          {listing.entries.map((entry) => {
            const path = joinPath(dir, entry.name);
            return (
              <div
                key={entry.name}
                className={clsx(
                  'group flex items-center space-x-2 rounded px-1 py-1.5 hover:bg-neutral-700/70',
                  entry.isDir && 'cursor-pointer'
                )}
                onClick={() => entry.isDir && load(path)}
              >
                {entry.isDir ? (
                  <FolderIcon size={16} className="shrink-0 text-amber-300" />
                ) : (
                  <FileIcon size={16} className="shrink-0 text-neutral-400" />
                )}
                <span className="min-w-0 flex-1 truncate text-sm" title={entry.name}>
                  {entry.name}
                </span>
                <span className="hidden w-[150px] shrink-0 text-right text-xs text-neutral-500 sm:block">
                  {new Date(entry.modifiedAt).toLocaleString()}
                </span>
                <span className="w-[70px] shrink-0 text-right text-xs text-neutral-400">
                  {entry.isDir ? '' : fileSize(entry.size)}
                </span>

                <div className="flex w-[52px] shrink-0 justify-end space-x-1">
                  {!entry.isDir && (
                    <Tooltip title={t('image.files.download')} mouseEnterDelay={0.6}>
                      <a
                        href={api.fileDownloadUrl(image, path)}
                        download={entry.name}
                        aria-label={t('image.files.download')}
                        className="flex h-[22px] w-[22px] items-center justify-center rounded text-neutral-300 hover:bg-neutral-500/50 hover:text-blue-400"
                        onClick={(e) => e.stopPropagation()}
                      >
                        <DownloadIcon size={14} />
                      </a>
                    </Tooltip>
                  )}
                  {canChange && (
                    <div onClick={(e) => e.stopPropagation()}>
                      <Popconfirm
                        title={t('image.attention')}
                        description={t(
                          entry.isDir
                            ? 'image.files.deleteDirConfirm'
                            : 'image.files.deleteConfirm',
                          { name: entry.name }
                        )}
                        okText={t('image.okBtn')}
                        cancelText={t('image.cancelBtn')}
                        okButtonProps={{ danger: true }}
                        onConfirm={() => remove(entry)}
                      >
                        <button
                          type="button"
                          aria-label={t('image.files.delete')}
                          className="flex h-[22px] w-[22px] cursor-pointer items-center justify-center rounded p-0 text-neutral-300 hover:bg-neutral-500/50 hover:text-red-500"
                        >
                          {deleting === path ? (
                            <LoaderCircleIcon className="animate-spin text-red-500" size={14} />
                          ) : (
                            <Trash2Icon size={14} />
                          )}
                        </button>
                      </Popconfirm>
                    </div>
                  )}
                </div>
              </div>
            );
          })}
          {listing.truncated && (
            <div className="pt-2 text-xs text-amber-400">
              {t('image.files.truncated', { shown: listing.entries.length })}
            </div>
          )}
        </div>
      )}
    </Modal>
  );
};
