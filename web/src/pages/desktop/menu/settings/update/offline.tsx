import { useRef, useState } from 'react';
import { Button, Input, Progress } from 'antd';
import { ExternalLinkIcon, FileArchiveIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/application.ts';
import { describeFailure } from '@/lib/feedback.ts';
import {
  reloadAfterRestart,
  SERVER_RESTART_DOWN_MS,
  SERVER_RESTART_UP_MS
} from '@/lib/wait-server.ts';

interface UpdateProps {
  status: string;
  setStatus: (status: string) => void;
  setIsLocked: (isClosable: boolean) => void;
  setErrMsg: (msg: string) => void;
}

// Both products, and the same shape the server enforces. IronKVM ships its
// own packages and an official Sipeed package stays installable, because
// being able to return to the official firmware is the reason the rename
// stayed shallow. This check ran in the browser and refused every IronKVM
// release before the server ever saw the file.
function validateFilename(filename: string) {
  const regex: RegExp = /^(?:nanokvm|ironkvm)_\d+\.\d+\.\d+\.tar\.gz$/;
  return regex.test(filename);
}

// A refusal the server explained, as opposed to a request that failed.
class UpdateRefused extends Error {}

// The checksum comes first and the file second: the upload starts from the
// button after both, so a checksum typed after choosing the file is not left
// out of a request already sent.
export const Offline = ({ status, setStatus, setIsLocked, setErrMsg }: UpdateProps) => {
  const { t } = useTranslation();

  const inputRef = useRef<HTMLInputElement | null>(null);
  const [sha256Checksum, setSha256Checksum] = useState('');
  const [file, setFile] = useState<File | null>(null);
  // null while no upload runs.
  const [progress, setProgress] = useState<number | null>(null);

  const isBusy = status === 'loading' || status === 'updating';

  function handleClick() {
    inputRef.current?.click();
  }

  function handleFileChange(e: React.ChangeEvent<HTMLInputElement>) {
    const chosen = e.target.files?.[0];
    e.target.value = '';
    if (!chosen) return;

    if (!validateFilename(chosen.name)) {
      setFile(null);
      setStatus('failed');
      setErrMsg(t('settings.update.offline.invalidName'));
      return;
    }

    setErrMsg('');
    setFile(chosen);
  }

  function upload() {
    if (!file || isBusy) return;

    const checksum = sha256Checksum.trim();
    if (checksum && !/^[a-fA-F0-9]{64}$/.test(checksum)) {
      setStatus('failed');
      setErrMsg(t('settings.update.offline.invalidChecksum'));
      return;
    }

    setIsLocked(true);
    setStatus('updating');
    setErrMsg('');
    setProgress(0);

    const formData = new FormData();
    formData.append('file', file);

    api
      .offlineUpdate(formData, checksum, setProgress)
      .then((rsp) => {
        // The proxy may return 502 after the update stops the old server.
        if (rsp.status === 502) return;

        const body = rsp.body;
        if (!rsp.ok || !body || body.code !== 0) {
          const message = body?.msg?.includes('sha256 checksum mismatch')
            ? t('settings.update.offline.checksumMismatch')
            : body?.msg || t('settings.update.offline.updateFailed');
          throw new UpdateRefused(message);
        }
      })
      .then(() => {
        // Installed: the server restarts on its own. The settings stay locked
        // until the page reloads onto the new version.
        reloadAfterRestart(SERVER_RESTART_DOWN_MS, SERVER_RESTART_UP_MS);
      })
      .catch((error: unknown) => {
        setIsLocked(false);
        setStatus('failed');
        setProgress(null);
        setErrMsg(
          error instanceof UpdateRefused
            ? error.message
            : describeFailure(error, t('settings.update.offline.updateFailed'))
        );
      });
  }

  return (
    <>
      <div className="mt-8 flex flex-col gap-3">
        <div className="flex flex-col space-y-1">
          <div className="flex items-center space-x-2">
            <span>{t('settings.update.offline.title')}</span>

            <a
              className="flex items-center text-neutral-500 hover:text-blue-500"
              href="https://github.com/yuzi-co/IronKVM/releases"
              target="_blank"
            >
              <ExternalLinkIcon size={15} />
            </a>
          </div>

          <span className="text-xs text-neutral-500">{t('settings.update.offline.desc')}</span>
        </div>

        <Input
          value={sha256Checksum}
          maxLength={64}
          disabled={isBusy}
          placeholder={t('settings.update.offline.checksumPlaceholder')}
          onChange={(event) => setSha256Checksum(event.target.value)}
        />

        <div className="flex items-center justify-between gap-3">
          <input
            id="file-upload"
            ref={inputRef}
            type="file"
            accept=".tar.gz"
            onChange={handleFileChange}
            className="hidden"
          />
          <Button disabled={isBusy} icon={<FileArchiveIcon size={15} />} onClick={handleClick}>
            {t('settings.update.offline.chooseFile')}
          </Button>
          <span className="min-w-0 flex-1 truncate font-mono text-xs text-neutral-400">
            {file ? file.name : t('settings.update.offline.noFile')}
          </span>
          <Button type="primary" disabled={!file || isBusy} onClick={upload}>
            {t('settings.update.offline.upload')}
          </Button>
        </div>

        {progress !== null && (
          <div className="flex flex-col">
            <Progress percent={progress} size="small" />
            {progress >= 100 && (
              <span className="text-xs text-neutral-500">
                {t('settings.update.offline.installing')}
              </span>
            )}
          </div>
        )}
      </div>
    </>
  );
};
