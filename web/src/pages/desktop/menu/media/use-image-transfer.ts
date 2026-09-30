import { useEffect, useRef, useState } from 'react';
import type { ChangeEvent } from 'react';
import { useTranslation } from 'react-i18next';

import {
  cancelDownloadImage,
  downloadBootMenu,
  downloadImage,
  imageEnabled,
  statusImage,
  uploadImageFile
} from '@/api/download.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

export const imageUpdatedEvent = 'nanokvm:image-updated';

// The server takes ISO 9660 images only; it checks the name and the content.
function isISO(file: File) {
  return file.name.toLowerCase().endsWith('.iso');
}

// Where a transfer was started from. The progress shows next to the control
// that started it: the Media dialog's add image form, or the boot menu button
// on the Network boot settings page.
export type Origin = 'library' | 'boot';

export type ImageTransfer = ReturnType<typeof useImageTransfer>;

// useImageTransfer holds the one image transfer the server runs at a time: a
// download from a URL, the boot menu download, or an upload from this browser.
// The Media menu calls handleOpenChange as it opens and closes; the Network
// boot settings page calls it as it opens.
export function useImageTransfer() {
  const { t } = useTranslation();

  const [input, setInput] = useState('');
  const [sha256sum, setSha256sum] = useState('');
  const [status, setStatus] = useState('');
  const [log, setLog] = useState('');
  const [origin, setOrigin] = useState<Origin>('library');
  const [isCancelling, setIsCancelling] = useState(false);
  const [isRemoteDownloading, setIsRemoteDownloading] = useState(false);
  const [diskEnabled, setDiskEnabled] = useState(false);

  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [isDragging, setIsDragging] = useState(false);
  // -1 while no upload runs; otherwise the share of the file sent so far.
  const [uploadPercent, setUploadPercent] = useState(-1);

  const stopPoll = useRef<(() => void) | undefined>(undefined);
  const pollingGeneration = useRef(0);
  const remoteDownloadActive = useRef(false);
  const fileUploadActive = useRef(false);
  const downloadRequestGeneration = useRef(0);

  useEffect(() => {
    checkDiskEnabled();
  }, []);

  // A page that goes away stops watching the transfer. The Media dialog never
  // does; the Network boot settings page does when another tab is chosen.
  useEffect(() => {
    const generation = pollingGeneration;
    const stop = stopPoll;
    return () => {
      generation.current += 1;
      stop.current?.();
    };
  }, []);

  function checkDiskEnabled() {
    imageEnabled()
      .then((res) => {
        setDiskEnabled(res.data.enabled);
      })
      .catch(() => {
        setDiskEnabled(false);
      });
  }

  function handleOpenChange(open: boolean) {
    if (open) {
      checkDiskEnabled();
      startStatusPolling();
    } else {
      // Keep monitoring an active remote download after the dialog closes so
      // completion can still refresh an already-open image list.
      const transferActive = remoteDownloadActive.current || fileUploadActive.current;
      if (!transferActive) {
        setInput('');
        setSha256sum('');
        setStatus('');
        setLog('');
        setIsCancelling(false);
        setIsRemoteDownloading(false);
        stopStatusPolling();
      }
    }
  }

  function fail(text: string, from: Origin = 'library') {
    setOrigin(from);
    setStatus('failed');
    setLog(text);
  }

  function getValidatedSHA256() {
    const checksum = sha256sum.trim();
    if (checksum && !/^[a-fA-F0-9]{64}$/.test(checksum)) {
      fail(t('download.invalidSHA256'));
      return null;
    }

    return checksum;
  }

  function startStatusPolling() {
    stopStatusPolling();
    const generation = pollingGeneration.current;
    getDownloadStatus(generation);
    stopPoll.current = pollWhileVisible(() => getDownloadStatus(generation), 2500);
  }

  function stopStatusPolling() {
    pollingGeneration.current += 1;
    stopPoll.current?.();
    stopPoll.current = undefined;
  }

  function finishImageTransfer(refreshImages: boolean, uploaded = false) {
    stopStatusPolling();
    remoteDownloadActive.current = false;
    fileUploadActive.current = false;
    setIsRemoteDownloading(false);
    setUploadPercent(-1);
    setStatus('success');
    setLog(t(uploaded ? 'download.uploadSuccess' : 'download.success'));

    if (refreshImages) {
      window.dispatchEvent(new Event(imageUpdatedEvent));
    }
  }

  function getDownloadStatus(generation = pollingGeneration.current) {
    statusImage()
      .then((rsp) => {
        // Ignore a response from a previous polling session. This can happen when
        // a download is started while the initial status request is still pending.
        if (generation !== pollingGeneration.current) return;
        // An upload shows its own progress, measured in the browser, and its
        // own result, from the upload request's answer.
        if (fileUploadActive.current) return;

        if (rsp.data.status) {
          setStatus(rsp.data.status);
          if (rsp.data.status === 'in_progress') {
            const isRemoteDownload = /^https?:\/\//.test(rsp.data.file);
            remoteDownloadActive.current = isRemoteDownload;
            setIsRemoteDownloading(isRemoteDownload);
            setLog(
              rsp.data.percentage
                ? t('download.downloadingPercent', {
                    file: rsp.data.file,
                    percent: rsp.data.percentage
                  })
                : t('download.downloading', { file: rsp.data.file })
            );
            setInput(rsp.data.file);
          }
          if (rsp.data.status === 'checksum_failed') {
            remoteDownloadActive.current = false;
            setIsRemoteDownloading(false);
            setLog(t('download.checksumFailed'));
            stopStatusPolling();
          }
          if (rsp.data.status === 'failed') {
            remoteDownloadActive.current = false;
            setIsRemoteDownloading(false);
            setLog(t('download.failed'));
            stopStatusPolling();
          }
          if (rsp.data.status === 'success') {
            const completedRemoteDownload = remoteDownloadActive.current;
            finishImageTransfer(completedRemoteDownload);
          }
          if (rsp.data.status === 'idle') {
            remoteDownloadActive.current = false;
            setIsRemoteDownloading(false);
            setLog('');
            stopStatusPolling();
          }
        }
      })
      // A missed poll is retried by the next tick.
      .catch(() => {});
  }

  function download(url?: string) {
    if (!url) return;

    const checksum = getValidatedSHA256();
    if (checksum === null) return;

    startRemoteDownload('library', url, () => downloadImage(url, checksum));
  }

  // The boot menu is netboot.xyz's ISO. The server holds its URL and its
  // checksum, and stores it in the image directory for the virtual CD.
  function downloadBootMenuImage() {
    startRemoteDownload('boot', 'netboot.xyz.iso', downloadBootMenu);
  }

  function startRemoteDownload(
    from: Origin,
    label: string,
    request: () => ReturnType<typeof downloadBootMenu>
  ) {
    // Invalidate the status request started when the dialog was opened.
    // Start polling only after the download request has created the server-side
    // download state, otherwise the first response can still be `idle`.
    stopStatusPolling();
    const requestGeneration = ++downloadRequestGeneration.current;
    remoteDownloadActive.current = true;
    setOrigin(from);
    setIsRemoteDownloading(true);
    setStatus('in_progress');
    setLog(t('download.downloading', { file: label }));

    request()
      .then((rsp) => {
        if (requestGeneration !== downloadRequestGeneration.current) return;

        if (rsp.code !== 0) {
          stopStatusPolling();
          remoteDownloadActive.current = false;
          setIsRemoteDownloading(false);
          setStatus('failed');
          setLog(rsp.msg || t('download.failed'));
          return;
        }

        // The boot menu ISO is already there with its pinned checksum, so
        // nothing was started.
        if (rsp.data?.status === 'present') {
          remoteDownloadActive.current = false;
          setIsRemoteDownloading(false);
          setStatus('success');
          setLog(t('download.bootMenuPresent', { file: rsp.data.file }));
          return;
        }

        startStatusPolling();
      })
      .catch(() => {
        if (requestGeneration !== downloadRequestGeneration.current) return;

        stopStatusPolling();
        remoteDownloadActive.current = false;
        setIsRemoteDownloading(false);
        setStatus('failed');
        setLog(t('download.failed'));
      });
  }

  function cancelDownload() {
    if (isCancelling) return;

    downloadRequestGeneration.current += 1;
    setIsCancelling(true);
    cancelDownloadImage()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setLog(rsp.msg || t('download.cancelFailed'));
          if (remoteDownloadActive.current) {
            startStatusPolling();
          }
          return;
        }

        stopStatusPolling();
        remoteDownloadActive.current = false;
        setIsRemoteDownloading(false);
        setStatus('idle');
        setLog('');
      })
      .catch(() => {
        setLog(t('download.cancelFailed'));
      })
      .finally(() => {
        setIsCancelling(false);
      });
  }

  // selectFile takes a file picked or dropped for upload. Picking a file also
  // stops watching a remote download; a drop leaves that alone.
  function selectFile(file: File | null, picked: boolean) {
    if (!file || !isISO(file)) {
      fail(t('download.NoISO'));
      return;
    }
    if (picked) {
      setIsRemoteDownloading(false);
      remoteDownloadActive.current = false;
      stopStatusPolling();
    }
    setOrigin('library');
    setStatus('idle');
    setLog('');
    setSelectedFile(file);
  }

  function handleFileChange(e: ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0] ?? null;
    // Clear the picker so choosing the same file again still fires onChange.
    e.target.value = '';
    selectFile(file, true);
  }

  function upload(file: File | null) {
    if (!file) return;

    if (!isISO(file)) {
      fail(t('download.NoISO'));
      return;
    }

    const checksum = getValidatedSHA256();
    if (checksum === null) return;

    setOrigin('library');
    setStatus('in_progress');
    remoteDownloadActive.current = false;
    fileUploadActive.current = true;
    setIsRemoteDownloading(false);
    setUploadPercent(0);
    setLog(t('download.uploading', { file: file.name }));

    uploadImageFile(file, checksum, setUploadPercent)
      .then((rsp) => {
        if (rsp.code !== 0) {
          throw new Error(uploadError(rsp.msg));
        }

        finishImageTransfer(true, true);
        setSelectedFile(null);
      })
      .catch((error: any) => {
        fileUploadActive.current = false;
        stopStatusPolling();
        setUploadPercent(-1);
        setStatus('failed');
        // A rejected request may still carry the server's answer; a lost one
        // has only axios's English text, which says less than ours.
        if (error?.isAxiosError) {
          setLog(uploadError(error.response?.data?.msg ?? ''));
        } else {
          setLog(error?.message || t('download.uploadFailed'));
        }
      });
  }

  function uploadError(msg: string) {
    return msg === 'sha256 mismatch'
      ? t('download.checksumFailed')
      : msg || t('download.uploadFailed');
  }

  return {
    input,
    setInput,
    sha256sum,
    setSha256sum,
    status,
    log,
    origin,
    isCancelling,
    isRemoteDownloading,
    diskEnabled,
    selectedFile,
    isDragging,
    setIsDragging,
    uploadPercent,
    handleOpenChange,
    download,
    downloadBootMenuImage,
    cancelDownload,
    selectFile,
    handleFileChange,
    upload
  };
}
