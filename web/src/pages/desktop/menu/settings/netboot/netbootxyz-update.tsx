import { useEffect, useState } from 'react';
import { Button, Popconfirm, Progress } from 'antd';
import { useTranslation } from 'react-i18next';

import type { UpdateStatus } from '@/api/updates.ts';
import { UPDATE_PATHS } from '@/api/updates.ts';
import { updateView } from '@/lib/upstream-update.ts';
import { updateNote, useUpstreamUpdate } from '@/hooks/useUpstreamUpdate.ts';

import { imageUpdatedEvent } from '../../media/use-image-transfer.ts';

type NetbootXYZUpdateProps = {
  // Whether the add-on, and so the boot files, is installed; the files'
  // status is read again when it changes.
  addonInstalled: boolean;
  // Runs when an update of the boot files has ended.
  onFilesUpdated: () => void;
};

// NetbootXYZUpdate is the one update control for both things this page takes
// from netboot.xyz's releases: the boot files the add-on serves, and the boot
// menu ISO for the virtual CD. The server updates each on its own path, but
// both come from the same release, so the page checks once and updates
// whichever of the two is installed and behind.
export const NetbootXYZUpdate = ({ addonInstalled, onFilesUpdated }: NetbootXYZUpdateProps) => {
  const { t } = useTranslation();
  const files = useUpstreamUpdate(
    UPDATE_PATHS.netbootxyz,
    t('settings.netboot.updateFiles'),
    onFilesUpdated
  );
  const iso = useUpstreamUpdate(UPDATE_PATHS.bootMenu, t('settings.netboot.updateIso'), () =>
    window.dispatchEvent(new Event(imageUpdatedEvent))
  );
  const [busy, setBusy] = useState<'' | 'check' | 'update'>('');

  const reloadFiles = files.load;
  useEffect(() => {
    reloadFiles();
  }, [addonInstalled, reloadFiles]);

  // A download from the boot menu section adds the ISO.
  const reloadIso = iso.load;
  useEffect(() => {
    window.addEventListener(imageUpdatedEvent, reloadIso);
    return () => window.removeEventListener(imageUpdatedEvent, reloadIso);
  }, [reloadIso]);

  const parts = [
    { key: 'files', label: t('settings.netboot.updateFiles'), ...files },
    { key: 'iso', label: t('settings.netboot.updateIso'), ...iso }
  ].filter((p): p is typeof p & { status: UpdateStatus } => !!p.status?.installed);
  if (parts.length === 0) return null;

  const views = parts.map((p) => updateView(p.status));
  const running = views.includes('running');
  const behind = parts.filter((_, i) => views[i] === 'available');
  const ready = behind.filter((p) => !p.status.inUse);
  const latest = behind[0]?.status.latest ?? '';

  // One question to GitHub: the first check asks it, and the second gets the
  // server's cached answer, as both name the same repository.
  function check() {
    setBusy('check');
    void (async () => {
      const [first, ...rest] = parts;
      if (await first.check(true)) {
        for (const p of rest) await p.check(false);
      }
    })().finally(() => setBusy(''));
  }

  function update() {
    setBusy('update');
    void Promise.all(ready.map((p) => p.start())).finally(() => setBusy(''));
  }

  return (
    <div className="flex flex-col space-y-1">
      <div className="flex items-center justify-between gap-2">
        <span className="min-w-0 truncate text-xs text-neutral-500">netboot.xyz</span>
        {behind.length > 0 && !running ? (
          <Popconfirm
            title={t('upstream.confirm', { name: 'netboot.xyz', version: latest })}
            description={
              <div className="flex max-w-[260px] flex-col space-y-1">
                <span>{ready.map((p) => p.label).join(', ')}</span>
                <span>{t('upstream.confirmDesc')}</span>
              </div>
            }
            okText={t('upstream.ok')}
            cancelText={t('common.cancel')}
            disabled={ready.length === 0}
            onConfirm={update}
          >
            <Button
              type="primary"
              size="small"
              loading={busy === 'update'}
              disabled={!!busy || ready.length === 0}
            >
              {t('upstream.updateTo', { version: latest })}
            </Button>
          </Popconfirm>
        ) : !running ? (
          <Button size="small" loading={busy === 'check'} disabled={!!busy} onClick={check}>
            {t('upstream.check')}
          </Button>
        ) : null}
      </div>
      {parts.map((p, i) => {
        const { note, noteClass } = updateNote(t, p.status, views[i]);
        return (
          <div key={p.key} className="flex flex-col space-y-1">
            <div className="flex flex-wrap items-center justify-between gap-x-3">
              <span className="text-xs text-neutral-500">{p.label}</span>
              <span className="font-mono text-xs text-neutral-500">
                {p.status.version}
                {p.status.version === p.status.pinned && ` (${t('upstream.builtIn')})`}
              </span>
            </div>
            {note && <span className={`text-xs break-words ${noteClass}`}>{note}</span>}
            {views[i] === 'running' && (
              <Progress percent={p.status.job.progress} size="small" status="active" />
            )}
          </div>
        );
      })}
    </div>
  );
};
