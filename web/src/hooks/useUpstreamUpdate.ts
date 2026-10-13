import { useEffect, useRef, useState } from 'react';
import { message } from 'antd';
import type { TFunction } from 'i18next';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/updates.ts';
import type { UpdateStatus } from '@/api/updates.ts';
import { showFailure } from '@/lib/feedback.ts';
import type { UpdateView } from '@/lib/upstream-update.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { useStableCallback } from '@/hooks/useStableCallback.ts';

// A running update is followed closely: it takes seconds to a minute.
const runningPollMs = 1000;

// useUpstreamUpdate follows the update status of one component at path, and
// asks the server to check for or start an update. check and start resolve
// to false when the server refused or could not be reached; the failure has
// been shown by then.
export function useUpstreamUpdate(path: string, name: string, onUpdated?: () => void) {
  const { t } = useTranslation();
  const [status, setStatus] = useState<UpdateStatus | null>(null);
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
    return pollWhileVisible(load, runningPollMs);
  }, [running, load]);

  // force asks GitHub even when the server's last answer is fresh.
  const check = useStableCallback((force: boolean = true): Promise<boolean> =>
    api
      .checkUpdate(path, force)
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          return false;
        }
        accept(rsp.data);
        return true;
      })
      .catch((err) => {
        showFailure(err);
        return false;
      })
  );

  const start = useStableCallback((): Promise<boolean> =>
    api
      .startUpdate(path)
      .then((rsp) => {
        if (rsp.code !== 0) {
          showFailure(rsp);
          load();
          return false;
        }
        accept(rsp.data);
        return true;
      })
      .catch((err) => {
        showFailure(err);
        load();
        return false;
      })
  );

  return { status, load, check, start };
}

// updateNote is the line under a component's row for its status, and its
// colour class; an empty note shows nothing.
export function updateNote(t: TFunction, status: UpdateStatus, view: UpdateView) {
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
  return { note, noteClass };
}
