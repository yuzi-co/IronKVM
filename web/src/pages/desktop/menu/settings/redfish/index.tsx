import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Alert, Button, Divider, message, Modal, Switch, Tag } from 'antd';
import { RefreshCwIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/redfish.ts';
import type { RedfishSession, RedfishSettings } from '@/api/redfish.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getBaseUrl } from '@/lib/service.ts';

import { PowerLedSetting } from '../../power/power-led-setting.tsx';
import { CopyBlock, CopyButton } from '../components/copy-button.tsx';
import { StatusTag } from '../components/status-tag.tsx';

function formatTime(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

// The Redfish service: its switch, where it answers, which power actions it
// offers, and the open sessions.
export const Redfish = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const [modal, contextHolder] = Modal.useModal();
  const [settings, setSettings] = useState<RedfishSettings | null>(null);
  const [sessions, setSessions] = useState<RedfishSession[]>([]);
  const [isSaving, setIsSaving] = useState(false);
  const [isLoadingSessions, setIsLoadingSessions] = useState(false);

  const endpoint = settings ? `${getBaseUrl('http')}${settings.serviceRoot}` : '';
  // -k because the board's certificate is self-signed until one is installed.
  const example = settings
    ? `curl -k -u ${account.username} ${getBaseUrl('http')}${settings.serviceRoot.replace(/\/$/, '')}/Systems/1`
    : '';

  const getSettings = useCallback(() => {
    api
      .getRedfishSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(t('settings.redfish.failed'));
          return;
        }
        setSettings(rsp.data);
      })
      .catch((err) => message.error(describeFailure(err, t('settings.redfish.failed'))));
  }, [t]);

  // getSessions clears the loading flag; refreshSessions is what sets it.
  const getSessions = useCallback(() => {
    api
      .getRedfishSessions()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(t('settings.redfish.failed'));
          return;
        }
        setSessions(rsp.data || []);
      })
      .catch((err) => message.error(describeFailure(err, t('settings.redfish.failed'))))
      .finally(() => setIsLoadingSessions(false));
  }, [t]);

  useEffect(() => {
    getSettings();
    getSessions();
  }, [getSettings, getSessions]);

  function refreshSessions() {
    setIsLoadingSessions(true);
    getSessions();
  }

  function setEnabled(enabled: boolean) {
    setIsSaving(true);
    api
      .setRedfishEnabled(enabled)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.redfish.failed')));
          return;
        }
        setSettings((current) => (current ? { ...current, enabled } : current));
        message.success(t(enabled ? 'feedback.enabled' : 'feedback.disabled', { name: 'Redfish' }));
        getSessions();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.redfish.failed'))))
      .finally(() => setIsSaving(false));
  }

  function endSession(session: RedfishSession) {
    modal.confirm({
      title: t('settings.redfish.endConfirmTitle'),
      content: (
        <span className="text-sm text-neutral-400">{t('settings.redfish.endConfirmDesc')}</span>
      ),
      okText: t('settings.redfish.okBtn'),
      cancelText: t('settings.redfish.cancelBtn'),
      onOk: async () => {
        try {
          const rsp = await api.endRedfishSession(session.id);
          if (rsp.code !== 0) {
            message.error(describeFailure(rsp, t('settings.redfish.failed')));
          }
        } catch (err) {
          message.error(describeFailure(err, t('settings.redfish.failed')));
        } finally {
          getSessions();
        }
      }
    });
  }

  return (
    <>
      {contextHolder}
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.redfish.title')}</span>
        {settings && <StatusTag running={settings.enabled} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.redfish.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.redfish.serviceDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        <PowerLedSetting />

        {settings?.enabled && (
          <>
            <div className="flex flex-col space-y-3">
              <div className="group flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3.5 sm:flex-row sm:items-center sm:justify-between sm:space-y-0">
                <span className="w-28 shrink-0 text-sm font-medium text-neutral-400">
                  {t('settings.redfish.endpoint')}
                </span>
                <div className="flex min-w-0 items-center justify-between gap-2">
                  <span className="min-w-0 flex-1 truncate font-mono text-sm text-neutral-300 select-all">
                    {endpoint}
                  </span>
                  <CopyButton text={endpoint} />
                </div>
              </div>

              {settings.https ? (
                <Alert type="success" showIcon message={t('settings.redfish.httpsOn')} />
              ) : (
                <Alert type="warning" showIcon message={t('settings.redfish.httpsOff')} />
              )}

              <span className="text-xs text-neutral-500">{t('settings.redfish.credentials')}</span>

              <CopyBlock title={t('settings.redfish.example')} text={example} />
            </div>

            <div className="flex flex-col space-y-2">
              <span className="text-sm font-medium">{t('settings.redfish.powerActions')}</span>
              <div className="flex flex-wrap gap-y-2">
                {settings.resetTypes.map((resetType) => (
                  <Tag key={resetType} className="font-mono">
                    {resetType}
                  </Tag>
                ))}
              </div>
              <span className="text-xs text-neutral-500">
                {t('settings.redfish.powerActionsDesc')}
              </span>
            </div>

            <div className="flex flex-col space-y-2">
              <div className="flex items-center justify-between">
                <span className="text-sm font-medium">{t('settings.redfish.sessions')}</span>
                <Button
                  type="text"
                  size="small"
                  className="text-neutral-400 hover:text-white"
                  loading={isLoadingSessions}
                  icon={<RefreshCwIcon size={15} />}
                  title={t('settings.redfish.refresh')}
                  onClick={refreshSessions}
                />
              </div>

              {sessions.length === 0 ? (
                <span className="text-xs text-neutral-500">{t('settings.redfish.noSessions')}</span>
              ) : (
                <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
                  {sessions.map((session, index) => (
                    <div
                      key={session.id}
                      className={
                        index > 0
                          ? 'flex items-center justify-between border-t border-neutral-800 px-4 py-3'
                          : 'flex items-center justify-between px-4 py-3'
                      }
                    >
                      <div className="flex min-w-0 flex-col space-y-0.5">
                        <span className="truncate text-sm text-neutral-300">{session.user}</span>
                        <span className="text-xs text-neutral-500">
                          {t('settings.redfish.created')}: {formatTime(session.createdAt)}
                        </span>
                        <span className="text-xs text-neutral-500">
                          {t('settings.redfish.lastUsed')}: {formatTime(session.lastUsed)}
                        </span>
                      </div>
                      <Button size="small" danger onClick={() => endSession(session)}>
                        {t('settings.redfish.end')}
                      </Button>
                    </div>
                  ))}
                </div>
              )}
            </div>
          </>
        )}
      </div>
    </>
  );
};
