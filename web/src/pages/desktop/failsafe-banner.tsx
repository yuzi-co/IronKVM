import { useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { AlertTriangleIcon, LoaderCircleIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vm.ts';
import { describeFailure } from '@/lib/feedback.ts';

// The shape of GET /api/vm/failsafe. Only what the banner reads.
type FailsafeExit = { uptime: number; reason: string; excerpt: string };
type FailsafeState = { active: boolean; reason: string; exits: FailsafeExit[] };

// How long to wait before reloading after a retry. The server restarts in
// place, which takes a few seconds, and a reload any earlier meets a closed
// port and shows the browser's own error page.
const reloadAfterMs = 8000;

// FailsafeBanner says that the server started without video after a video
// crash loop, so the operator does not read the black screen as a cable
// fault. The rest of the UI works as usual: updates, logs, power and the
// terminal. An admin can try video again; a reboot does the same.
export const FailsafeBanner = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const isAdmin = account.role === 'admin';

  const [state, setState] = useState<FailsafeState | null>(null);
  const [showLog, setShowLog] = useState(false);
  const [isRetrying, setIsRetrying] = useState(false);
  const [errMsg, setErrMsg] = useState('');

  useEffect(() => {
    let live = true;
    api
      .getFailsafe()
      .then((rsp: any) => {
        if (live && rsp.code === 0) setState(rsp.data);
      })
      .catch((err) => console.log(err));
    return () => {
      live = false;
    };
  }, []);

  if (!state?.active) return null;

  function retry() {
    if (isRetrying) return;
    setIsRetrying(true);
    setErrMsg('');

    api
      .retryFailsafe()
      .then((rsp: any) => {
        if (rsp.code !== 0) {
          setErrMsg(describeFailure(rsp));
          setIsRetrying(false);
          return;
        }
        setTimeout(() => window.location.reload(), reloadAfterMs);
      })
      .catch((err) => {
        setErrMsg(describeFailure(err));
        setIsRetrying(false);
      });
  }

  const excerpt = state.exits[0]?.excerpt ?? '';

  return (
    <div className="fixed top-14 left-1/2 z-1000 flex w-[min(42rem,calc(100%-2rem))] -translate-x-1/2 flex-col gap-3 rounded-lg border border-amber-500/50 bg-neutral-900/95 p-4 text-sm shadow-xl">
      <div className="flex items-center gap-2.5 font-medium text-amber-400">
        <AlertTriangleIcon size={18} className="shrink-0" />
        <span>{t('failsafe.title')}</span>
      </div>

      <p className="text-neutral-300">{t('failsafe.body')}</p>
      <p className="font-mono text-xs break-words text-neutral-400">{state.reason}</p>

      {excerpt && (
        <div>
          <button
            type="button"
            className="cursor-pointer text-xs text-neutral-400 underline hover:text-neutral-200"
            onClick={() => setShowLog(!showLog)}
          >
            {showLog ? t('failsafe.hideLog') : t('failsafe.showLog')}
          </button>
          {showLog && (
            <pre className="mt-2 max-h-64 overflow-auto rounded bg-black/60 p-2 text-[11px] leading-snug whitespace-pre-wrap text-neutral-300">
              {excerpt}
            </pre>
          )}
        </div>
      )}

      {errMsg && <p className="text-xs text-red-400">{errMsg}</p>}

      {isAdmin && (
        <div className="flex justify-end">
          <button
            type="button"
            className="flex items-center gap-2 rounded bg-amber-600 px-3 py-1.5 text-white hover:bg-amber-500 disabled:cursor-default disabled:bg-amber-600/60"
            disabled={isRetrying}
            onClick={retry}
          >
            {isRetrying && <LoaderCircleIcon size={16} className="shrink-0 animate-spin" />}
            {isRetrying ? t('failsafe.retrying') : t('failsafe.retry')}
          </button>
        </div>
      )}
    </div>
  );
};
