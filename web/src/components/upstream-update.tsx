import { useState } from 'react';
import { Button, Popconfirm, Progress } from 'antd';
import { useTranslation } from 'react-i18next';

import { updateView } from '@/lib/upstream-update.ts';
import { updateNote, useUpstreamUpdate } from '@/hooks/useUpstreamUpdate.ts';

type UpstreamUpdateProps = {
  // The component's update endpoint, one of api.UPDATE_PATHS.
  path: string;
  // What the row and the confirmation call the component, such as "Ventoy".
  name: string;
  // Runs when an update this page saw start has ended, done or failed.
  onUpdated?: () => void;
};

// UpstreamUpdate shows the installed version of a downloaded component and lets
// the owner check GitHub for a newer release and update to it. Nothing is
// checked or updated without a click.
export const UpstreamUpdate = ({ path, name, onUpdated }: UpstreamUpdateProps) => {
  const { t } = useTranslation();
  const { status, check, start } = useUpstreamUpdate(path, name, onUpdated);
  const [busy, setBusy] = useState<'' | 'check' | 'update'>('');

  function run(what: 'check' | 'update', call: () => Promise<boolean>) {
    setBusy(what);
    void call().finally(() => setBusy(''));
  }

  if (!status || !status.installed) return null;

  const view = updateView(status);
  const job = status.job;
  const { note, noteClass } = updateNote(t, status, view);

  return (
    <div className="flex flex-col space-y-1">
      <div className="flex items-center justify-between gap-2">
        <span className="min-w-0 truncate font-mono text-xs text-neutral-500">
          {name} {status.version}
          {status.version === status.pinned && ` (${t('upstream.builtIn')})`}
        </span>
        {view === 'available' ? (
          <Popconfirm
            title={t('upstream.confirm', { name, version: status.latest })}
            description={<div className="max-w-[260px]">{t('upstream.confirmDesc')}</div>}
            okText={t('upstream.ok')}
            cancelText={t('common.cancel')}
            disabled={!!status.inUse}
            onConfirm={() => run('update', start)}
          >
            <Button
              type="primary"
              size="small"
              loading={busy === 'update'}
              disabled={!!busy || !!status.inUse}
            >
              {t('upstream.updateTo', { version: status.latest })}
            </Button>
          </Popconfirm>
        ) : view !== 'running' ? (
          <Button
            size="small"
            loading={busy === 'check'}
            disabled={!!busy}
            onClick={() => run('check', () => check())}
          >
            {t('upstream.check')}
          </Button>
        ) : null}
      </div>
      {note && <span className={`text-xs break-words ${noteClass}`}>{note}</span>}
      {view === 'running' && <Progress percent={job.progress} size="small" status="active" />}
    </div>
  );
};
