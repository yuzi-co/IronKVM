import { useEffect, useRef, useState } from 'react';
import { Button, Popconfirm } from 'antd';
import { FileArchiveIcon, RefreshCwIcon, Undo2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/ui.ts';
import type { UiLatest, UiRelease, UiStatus } from '@/api/ui.ts';
import { showFailure, showResult } from '@/lib/feedback.ts';
import { formatRange } from '@/lib/server-contract.ts';
import { pickUpload, uiOffer, type UploadPick } from '@/lib/ui-package.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

// The new UI is served as soon as the server switches to it; the page reloads
// onto it. A short pause lets the success message be read.
const RELOAD_DELAY_MS = 1500;

type Busy = '' | 'check' | 'update' | 'upload' | 'rollback';

// UiPackage is the web UI's own update: the UI is installed, and rolled back,
// without touching the server. Shown only by a server that has the feature.
export const UiPackage = ({ disabled }: { disabled: boolean }) => {
  const supported = useServerFeature('ui-update');
  if (!supported) return null;
  return <UiPackageSection disabled={disabled} />;
};

const UiPackageSection = ({ disabled }: { disabled: boolean }) => {
  const { t } = useTranslation();
  const inputRef = useRef<HTMLInputElement | null>(null);
  const [status, setStatus] = useState<UiStatus | null>(null);
  const [latest, setLatest] = useState<UiLatest | null>(null);
  const [checkError, setCheckError] = useState('');
  const [busy, setBusy] = useState<Busy>('');
  const [pick, setPick] = useState<UploadPick<File> | null>(null);

  const loadStatus = useStableCallback(() => {
    api
      .getUiStatus()
      .then((rsp) => {
        if (rsp.code === 0) setStatus(rsp.data as UiStatus);
      })
      .catch(() => {});
  });

  useEffect(() => {
    loadStatus();
  }, [loadStatus]);

  const blocked = disabled || busy !== '';

  function check() {
    setBusy('check');
    setCheckError('');
    api
      .getUiLatest()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setLatest(null);
          setCheckError(t('settings.update.ui.noPackage'));
          return;
        }
        setLatest(rsp.data as UiLatest);
      })
      .catch((err) => showFailure(err))
      .finally(() => setBusy(''));
  }

  // After an install or a rollback the server serves another UI. This page is
  // the old one, so it reloads.
  function done(rsp: { code: number; msg?: string }, success: string) {
    if (!showResult(rsp, { success })) {
      loadStatus();
      return;
    }
    setTimeout(() => window.location.reload(), RELOAD_DELAY_MS);
  }

  function update() {
    setBusy('update');
    api
      .updateUi()
      .then((rsp) => done(rsp, t('settings.update.ui.installed')))
      .catch((err) => showFailure(err))
      .finally(() => setBusy(''));
  }

  function rollback() {
    setBusy('rollback');
    api
      .rollbackUi()
      .then((rsp) => done(rsp, t('settings.update.ui.rolledBack')))
      .catch((err) => showFailure(err))
      .finally(() => setBusy(''));
  }

  function choose(e: React.ChangeEvent<HTMLInputElement>) {
    const files = Array.from(e.target.files ?? []);
    e.target.value = '';
    setPick(files.length ? pickUpload(files) : null);
  }

  function upload() {
    if (!pick?.ok) return;
    setBusy('upload');
    api
      .uploadUi(pick.pkg, pick.signature)
      .then((rsp) => done(rsp, t('settings.update.ui.installed')))
      .catch((err) => showFailure(err))
      .finally(() => {
        setBusy('');
        setPick(null);
      });
  }

  const describe = (r: UiRelease | undefined) => {
    if (!r) return '';
    if (r.bundled) {
      return r.version
        ? t('settings.update.ui.bundledVersion', { version: r.version })
        : t('settings.update.ui.bundled');
    }
    return r.version || r.id;
  };

  const offer = uiOffer(latest);

  return (
    <div className="mt-8 flex flex-col gap-3">
      <div className="flex items-start justify-between gap-3">
        <div className="flex min-w-0 flex-col space-y-1">
          <span>{t('settings.update.ui.title')}</span>
          <span className="text-xs text-neutral-500">{t('settings.update.ui.desc')}</span>
          {status && (
            <span className="text-xs text-neutral-400">
              {t('settings.update.ui.served', {
                version: describe(status.active),
                api: status.serverApi
              })}
            </span>
          )}
          {status?.signatureCheck === 'off' && (
            <span className="text-xs text-amber-400">{t('settings.update.ui.checkOff')}</span>
          )}
          {status?.signatureCheck === 'no-keys' && (
            <span className="text-xs text-amber-400">{t('settings.update.ui.noKeys')}</span>
          )}
        </div>
        <Button
          icon={<RefreshCwIcon size={15} />}
          loading={busy === 'check'}
          disabled={blocked}
          onClick={check}
        >
          {t('settings.update.ui.check')}
        </Button>
      </div>

      {checkError && <span className="text-xs text-neutral-500">{checkError}</span>}
      {offer.kind === 'current' && (
        <span className="text-xs text-neutral-500">
          {t('settings.update.ui.isCurrent', { version: offer.version })}
        </span>
      )}
      {offer.kind === 'incompatible' && (
        <span className="text-xs text-amber-400">
          {t('settings.update.ui.incompatible', {
            version: offer.version,
            api: formatRange(offer.api),
            server: status?.serverApi ?? '?'
          })}
        </span>
      )}
      {offer.kind === 'available' && (
        <Popconfirm
          title={t('settings.update.ui.installTo', { version: offer.version })}
          description={<div className="max-w-[320px]">{t('settings.update.ui.installDesc')}</div>}
          okText={t('settings.update.confirm')}
          cancelText={t('settings.update.cancel')}
          disabled={blocked}
          onConfirm={update}
        >
          <Button type="primary" loading={busy === 'update'} disabled={blocked}>
            {t('settings.update.ui.installTo', { version: offer.version })}
          </Button>
        </Popconfirm>
      )}

      <div className="flex items-center justify-between gap-3">
        <input
          ref={inputRef}
          type="file"
          multiple
          accept=".tar.gz,.minisig"
          onChange={choose}
          className="hidden"
        />
        <Button
          disabled={blocked}
          icon={<FileArchiveIcon size={15} />}
          onClick={() => inputRef.current?.click()}
        >
          {t('settings.update.ui.chooseFiles')}
        </Button>
        <span className="min-w-0 flex-1 truncate font-mono text-xs text-neutral-400">
          {!pick && t('settings.update.ui.chooseHint')}
          {pick?.ok &&
            [
              pick.pkg.name,
              pick.signature ? pick.signature.name : t('settings.update.ui.noSignature')
            ].join(' + ')}
          {pick && !pick.ok && (
            <span className="text-amber-400">{t(`settings.update.ui.pick.${pick.error}`)}</span>
          )}
        </span>
        <Button
          type="primary"
          loading={busy === 'upload'}
          disabled={blocked || !pick?.ok}
          onClick={upload}
        >
          {t('settings.update.ui.install')}
        </Button>
      </div>

      {status?.canRollback && (
        <div className="flex items-center justify-between gap-3">
          <span className="min-w-0 text-xs text-neutral-500">
            {t('settings.update.ui.rollbackDesc', { version: describe(status.previous) })}
          </span>
          <Popconfirm
            title={t('settings.update.ui.rollbackConfirm', { version: describe(status.previous) })}
            okText={t('settings.update.confirm')}
            cancelText={t('settings.update.cancel')}
            disabled={blocked}
            onConfirm={rollback}
          >
            <Button icon={<Undo2Icon size={15} />} loading={busy === 'rollback'} disabled={blocked}>
              {t('settings.update.ui.rollback')}
            </Button>
          </Popconfirm>
        </div>
      )}
    </div>
  );
};
