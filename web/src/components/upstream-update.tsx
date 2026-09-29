import { useEffect, useRef, useState } from 'react';
import { Button, message, Popconfirm, Progress } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/updates.ts';
import type { UpdateStatus } from '@/api/updates.ts';
import { showFailure } from '@/lib/feedback.ts';
import { updateView } from '@/lib/upstream-update.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

// A running update is followed closely: it takes seconds to a minute.
const runningPollMs = 1000;

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
  const [status, setStatus] = useState<UpdateStatus | null>(null);
  const [busy, setBusy] = useState<'' | 'check' | 'update'>('');
  const lastJob = useRef('');

  // accept shows a status, and reports the end of an update this page saw run.
  const accept = useStableCallback((st: UpdateStatus) => {
    const was = lastJob.current;
    lastJob.current = st.job.state;
    setStatus(st);
    if (was === 'running' && st.job.state === 'done') {
      message.success(t('upstream.done', { name, version: st.job.version }));
      onUpdated?.();
    }
    if (was === 'running' && st.job.state === 'failed') {
      onUpdated?.();
    }
  });

  const load = useStableCallback(() => {
    api
      .getUpdate(path)
      .then((rsp) => {
        if (rsp.code === 0) accept(rsp.data);
      })
      // The row stays as it was; the next poll or visit asks again.
      .catch(() => {});
  });

  useEffect(() => {
    load();
  }, [path, load]);

  const running = status?.job.state === 'running';
  useEffect(() => {
    if (!running) return;
    const timer = window.setInterval(load, runningPollMs);
    return () => window.clearInterval(timer);
  }, [running, load]);

  function check() {
    setBusy('check');
    api
      .checkUpdate(path)
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return;
        }
        accept(rsp.data);
      })
      .catch((err) => showFailure(err))
      .finally(() => setBusy(''));
  }

  function update() {
    setBusy('update');
    api
      .startUpdate(path)
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          load();
          return;
        }
        accept(rsp.data);
      })
      .catch((err) => {
        showFailure(err);
        load();
      })
      .finally(() => setBusy(''));
  }

  if (!status || !status.installed) return null;

  const view = updateView(status);
  const job = status.job;

  let note = '';
  let noteClass = 'text-neutral-500';
  switch (view) {
    case 'upToDate':
      note = t('upstream.upToDate');
      break;
    case 'checkFailed':
      note = t('upstream.checkFailed', { error: status.checkError });
      noteClass = 'text-amber-500';
      break;
    case 'unverifiable':
      note = t('upstream.unverifiable', { version: status.latest, reason: status.unverifiable });
      noteClass = 'text-amber-500';
      break;
    case 'available':
      if (status.inUse) {
        note = t('upstream.inUse', { reason: status.inUse });
        noteClass = 'text-amber-500';
      }
      break;
    case 'running':
      note = t('upstream.running', { version: job.version });
      break;
  }
  if (view !== 'running' && job.state === 'failed') {
    note = t('upstream.failed', { error: job.error });
    noteClass = 'text-red-500';
  }

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
            onConfirm={update}
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
          <Button size="small" loading={busy === 'check'} disabled={!!busy} onClick={check}>
            {t('upstream.check')}
          </Button>
        ) : null}
      </div>
      {note && <span className={`text-xs break-words ${noteClass}`}>{note}</span>}
      {view === 'running' && <Progress percent={job.progress} size="small" status="active" />}
    </div>
  );
};
