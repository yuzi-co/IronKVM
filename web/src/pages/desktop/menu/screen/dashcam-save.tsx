import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { message, Tooltip } from 'antd';
import { CctvIcon, LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/dashcam.ts';
import type { DashcamPending, DashcamState } from '@/api/dashcam.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';

const rowClass =
  'box-border flex h-[30px] w-full items-center space-x-2 rounded p-0 pr-5 pl-3 text-left text-neutral-300';

// secondsUntil is how long until the clip is written, in whole seconds.
function secondsUntil(pending: DashcamPending) {
  return Math.max(0, Math.ceil((new Date(pending.deadline).getTime() - Date.now()) / 1000));
}

// DashcamSave saves the last minute of video as a clip on the board
// (ironkvm-dist#119). It shows only while the dashcam runs, and never to a
// viewer, whom the server refuses.
export const DashcamSave = () => {
  const { t } = useTranslation();
  const isViewer = useAuth().account.role === 'viewer';
  const hasDashcam = useServerFeature('dashcam') === true;
  const [state, setState] = useState<DashcamState | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  // The menu renders this when it opens, so the state is as fresh as the
  // menu.
  useEffect(() => {
    if (!hasDashcam || isViewer) return;
    api
      .getDashcamState()
      .then((rsp) => {
        if (rsp.code === 0) setState(rsp.data);
      })
      .catch(() => {});
  }, [hasDashcam, isViewer]);

  if (!hasDashcam || isViewer || !state?.enabled || !state.running) return null;

  async function save() {
    if (isSaving) return;
    setIsSaving(true);
    try {
      const rsp = await api.saveDashcamClip();
      if (rsp.code === 0) {
        const pending: DashcamPending = rsp.data;
        message.success(
          pending.joined
            ? t('screen.dashcam.joined')
            : t('screen.dashcam.saving', { seconds: secondsUntil(pending) })
        );
        return;
      }
      switch (rsp.code) {
        case api.DASHCAM_EMPTY:
          message.error(t('screen.dashcam.empty'));
          break;
        case api.DASHCAM_TOO_SOON:
          message.error(t('screen.dashcam.tooSoon'));
          break;
        default:
          message.error(describeFailure(rsp, t('screen.dashcam.failed')));
      }
    } catch (err) {
      message.error(describeFailure(err, t('screen.dashcam.failed')));
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <Tooltip placement="rightTop" title={t('screen.dashcam.tip')} color="#262626" arrow>
      <button
        type="button"
        className={`${rowClass} cursor-pointer hover:bg-neutral-700/70`}
        onClick={save}
      >
        {isSaving ? (
          <LoaderCircleIcon className="animate-spin" size={18} />
        ) : (
          <CctvIcon size={18} />
        )}
        <span className="text-sm select-none">
          {t('screen.dashcam.save', { seconds: state.seconds })}
        </span>
      </button>
    </Tooltip>
  );
};
