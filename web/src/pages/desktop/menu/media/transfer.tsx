import { useRef, useState } from 'react';
import { Button, Input, Progress } from 'antd';
import clsx from 'clsx';
import { ChevronDownIcon, ChevronRightIcon, PlusIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import type { ImageTransfer, Origin } from './use-image-transfer.ts';

type TransferProps = { transfer: ImageTransfer };

// TransferStatus shows how the current transfer is going, under the control
// that started it.
export const TransferStatus = ({ transfer, origin }: TransferProps & { origin: Origin }) => {
  const { status, log, uploadPercent } = transfer;
  if (transfer.origin !== origin) return null;

  return (
    <div className="min-h-8 pt-2">
      {status && (
        <div
          className={clsx(
            'text-sm wrap-break-word',
            status === 'failed' || status === 'checksum_failed' ? 'text-red-500' : 'text-green-500'
          )}
        >
          {log}
        </div>
      )}
      {uploadPercent >= 0 && status === 'in_progress' && (
        <Progress percent={uploadPercent} size="small" />
      )}
    </div>
  );
};

// AddImage adds images to the library: an upload of an ISO from this
// computer, or a download from a URL. Both sit behind one button, and the
// checksum behind a further toggle, since most visits only mount what is
// there. The progress shows whether the form is open or not.
export const AddImage = ({ transfer }: TransferProps) => {
  const { t } = useTranslation();
  const { status, isRemoteDownloading, isCancelling, selectedFile, isDragging } = transfer;
  const fileInputRef = useRef<HTMLInputElement>(null);
  const [isOpen, setIsOpen] = useState(false);
  const [showMore, setShowMore] = useState(false);
  const inProgress = status === 'in_progress';
  const canCancel = isRemoteDownloading && inProgress;
  // A running download keeps the form open, so its cancel button stays in reach.
  const open = isOpen || (canCancel && transfer.origin === 'library');
  const moreOpen = showMore || !!transfer.sha256sum;

  if (!transfer.diskEnabled) {
    return <div className="text-red-500">{t('download.disabled')}</div>;
  }

  return (
    <div>
      <Button
        className="self-start"
        icon={<PlusIcon size={16} />}
        aria-expanded={open}
        onClick={() => setIsOpen(!open)}
      >
        {t('menu.mediaAdd')}
      </Button>

      {open && (
        <div className="mt-3 space-y-2">
          <div>
            <div className="mb-1 text-neutral-500">{t('download.inputfile')}</div>
            <div className="flex items-center gap-1">
              <button
                type="button"
                className={clsx(
                  'flex h-10 min-w-0 flex-1 flex-col items-center justify-center rounded-xl border-2 border-solid p-0 transition',
                  isDragging ? 'border-blue-500 bg-neutral-500' : 'border-neutral-600',
                  inProgress
                    ? 'cursor-not-allowed bg-neutral-700 opacity-50'
                    : 'cursor-pointer hover:bg-neutral-500'
                )}
                onDrop={(e) => {
                  if (inProgress) return;
                  e.preventDefault();
                  transfer.setIsDragging(false);
                  transfer.selectFile(e.dataTransfer.files?.[0] ?? null, false);
                }}
                onDragOver={(e) => {
                  if (inProgress) return;
                  e.preventDefault();
                  transfer.setIsDragging(true);
                }}
                onDragLeave={(e) => {
                  if (inProgress) return;
                  e.preventDefault();
                  transfer.setIsDragging(false);
                }}
                onClick={() => {
                  if (inProgress) return;
                  fileInputRef.current?.click();
                }}
              >
                <span className="w-full truncate px-2 text-center text-sm text-neutral-100">
                  {selectedFile ? selectedFile.name : t('download.uploadbox')}
                </span>
              </button>
              {/* A plain input: antd's Input styles beat Tailwind's `hidden`,
                  which left the native picker showing under the drop zone. */}
              <input
                ref={fileInputRef}
                type="file"
                accept=".iso"
                hidden
                onChange={transfer.handleFileChange}
                disabled={inProgress}
              />
              <Button
                type="primary"
                className="h-10 w-16 shrink-0 border-2 px-0"
                onClick={() => transfer.upload(selectedFile)}
                disabled={inProgress || !selectedFile}
              >
                {t('download.ok')}
              </Button>
            </div>
          </div>
          <div>
            <div className="mb-1 text-neutral-500">{t('download.input')}</div>
            <div className="flex items-center gap-1">
              <Input
                value={transfer.input}
                onChange={(e) => transfer.setInput(e.target.value)}
                disabled={inProgress}
                className="min-w-0 flex-1"
              />
              <Button
                type="primary"
                className="h-10 w-16 shrink-0 px-0"
                danger={canCancel}
                onClick={() =>
                  canCancel ? transfer.cancelDownload() : transfer.download(transfer.input)
                }
                disabled={isCancelling || (inProgress && !isRemoteDownloading)}
              >
                {canCancel ? t('download.cancel') : t('download.ok')}
              </Button>
            </div>
          </div>
          <button
            type="button"
            aria-expanded={moreOpen}
            className="flex cursor-pointer items-center space-x-1 p-0 text-xs text-neutral-500 select-none hover:text-neutral-300"
            onClick={() => setShowMore(!moreOpen)}
          >
            {moreOpen ? <ChevronDownIcon size={14} /> : <ChevronRightIcon size={14} />}
            <span>{t('menu.mediaMoreOptions')}</span>
          </button>
          {moreOpen && (
            <div>
              <div className="mb-1 text-neutral-500">{t('download.sha256')}</div>
              <Input
                value={transfer.sha256sum}
                onChange={(e) => transfer.setSha256sum(e.target.value)}
                disabled={inProgress}
                maxLength={64}
                placeholder={t('download.sha256Placeholder')}
              />
            </div>
          )}
        </div>
      )}

      <TransferStatus transfer={transfer} origin="library" />
    </div>
  );
};
