import { useCallback, useEffect, useState } from 'react';
import { Button, message, Modal, Tag } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/auth.ts';
import type { LoginSession } from '@/api/auth.ts';
import { describeFailure } from '@/lib/feedback.ts';

function formatTime(seconds: number) {
  return new Date(seconds * 1000).toLocaleString();
}

type SessionsProps = {
  // own lists the signed-in account's sessions; all lists every account's,
  // for an administrator, under the user list.
  scope: 'own' | 'all';
};

// Sessions lists login sessions and revokes them. A revoked session's next
// request is refused and its open connections close. The session this page
// runs in is marked and has no revoke button: signing out is the way to end it.
export const Sessions = ({ scope }: SessionsProps) => {
  const { t } = useTranslation();
  const [messageApi, messageHolder] = message.useMessage();
  const [modal, modalHolder] = Modal.useModal();

  const [sessions, setSessions] = useState<LoginSession[]>([]);
  const [isLoading, setIsLoading] = useState(true);

  // A promise chain rather than async/await, for the same reason as the users
  // list: the effect lint cannot tell that a finally after an await runs later.
  const loadSessions = useCallback(
    () =>
      (scope === 'all' ? api.getAllSessions() : api.getSessions())
        .then((rsp) => {
          if (rsp.code !== 0) throw rsp;
          const listed: LoginSession[] = Array.isArray(rsp.data?.sessions) ? rsp.data.sessions : [];
          setSessions(listed);
        })
        .catch((err) => {
          messageApi.error(describeFailure(err, t('settings.account.sessions.loadFailed')));
        })
        .finally(() => {
          setIsLoading(false);
        }),
    [messageApi, scope, t]
  );

  useEffect(() => {
    loadSessions();
  }, [loadSessions]);

  function revoke(session: LoginSession) {
    modal.confirm({
      title: t('settings.account.sessions.revokeConfirm'),
      okText: t('settings.account.sessions.revoke'),
      okButtonProps: { danger: true },
      cancelText: t('settings.account.cancelBtn'),
      onOk: async () => {
        try {
          const rsp = await api.revokeSession(session.id);
          if (rsp.code !== 0) throw rsp;
          messageApi.success(t('settings.account.sessions.revoked'));
        } catch (err) {
          messageApi.error(describeFailure(err, t('settings.account.sessions.revokeFailed')));
        }
        await loadSessions();
      }
    });
  }

  function revokeOthers() {
    modal.confirm({
      title: t('settings.account.sessions.revokeOthersConfirm'),
      okText: t('settings.account.sessions.revokeOthers'),
      okButtonProps: { danger: true },
      cancelText: t('settings.account.cancelBtn'),
      onOk: async () => {
        try {
          const rsp = await api.revokeOtherSessions();
          if (rsp.code !== 0) throw rsp;
          messageApi.success(t('settings.account.sessions.revoked'));
        } catch (err) {
          messageApi.error(describeFailure(err, t('settings.account.sessions.revokeFailed')));
        }
        await loadSessions();
      }
    });
  }

  const hasOthers = sessions.some((session) => !session.current);

  return (
    <>
      {messageHolder}
      {modalHolder}
      <div className="mb-1 flex items-center justify-between gap-3">
        <span className="text-base">
          {scope === 'all'
            ? t('settings.account.sessions.allTitle')
            : t('settings.account.sessions.title')}
        </span>
        {scope === 'own' && (
          <Button danger size="small" disabled={isLoading || !hasOthers} onClick={revokeOthers}>
            {t('settings.account.sessions.revokeOthers')}
          </Button>
        )}
      </div>
      <div className="mb-4 text-sm text-neutral-400">
        {scope === 'all'
          ? t('settings.account.sessions.allDesc')
          : t('settings.account.sessions.desc')}
      </div>

      <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
        {sessions.length === 0 ? (
          <div className="px-4 py-3.5 text-sm text-neutral-500">
            {isLoading ? '...' : t('settings.account.sessions.empty')}
          </div>
        ) : (
          sessions.map((session, index) => (
            <div
              key={session.id}
              className={`flex items-center justify-between gap-3 px-4 py-3 ${
                index > 0 ? 'border-t border-neutral-800' : ''
              }`}
            >
              <div className="flex min-w-0 flex-col">
                <span className="flex min-w-0 items-center gap-2 text-sm font-medium">
                  <span className="truncate">
                    {scope === 'all' && `${session.username} · `}
                    {session.client || t('settings.account.sessions.unknownClient')}
                    {session.ip && ` · ${session.ip}`}
                  </span>
                  {session.current && (
                    <Tag color="green" className="!me-0">
                      {t('settings.account.sessions.current')}
                    </Tag>
                  )}
                </span>
                <span className="truncate text-xs text-neutral-500">
                  {t('settings.account.sessions.lastUsed', {
                    time: formatTime(session.lastUsedAt)
                  })}
                  {' · '}
                  {t('settings.account.sessions.created', { time: formatTime(session.createdAt) })}
                </span>
              </div>
              {!session.current && (
                <Button danger size="small" disabled={isLoading} onClick={() => revoke(session)}>
                  {t('settings.account.sessions.revoke')}
                </Button>
              )}
            </div>
          ))
        )}
      </div>
    </>
  );
};
