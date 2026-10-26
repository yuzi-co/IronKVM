import { useCallback, useEffect, useState } from 'react';
import { Divider, message, Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/whep.ts';
import type { WhepSettings } from '@/api/whep.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getBaseUrl } from '@/lib/service.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import { CopyBlock, CopyRow } from '../components/copy-button.tsx';
import { gstreamerCommand, obsGstreamerPipeline, whepEndpoint } from './commands.ts';

// The session count changes when a player connects, which nothing announces.
const pollMs = 5 * 1000;

// WHEP (ironkvm-dist#145): the switch, the endpoint a player is given, and
// commands that play it.
export const Whep = () => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<WhepSettings | null>(null);
  const [isSaving, setIsSaving] = useState(false);

  const load = useCallback(() => {
    api
      .getWhepSettings()
      .then((rsp) => {
        if (rsp.code !== 0) return;
        setSettings(rsp.data as WhepSettings);
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    load();
    return pollWhileVisible(load, pollMs);
  }, [load]);

  function setEnabled(enabled: boolean) {
    setIsSaving(true);
    api
      .setWhepSettings(enabled)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.whep.failed')));
          return;
        }
        setSettings(rsp.data as WhepSettings);
        message.success(t('settings.whep.saved'));
      })
      .catch((err) => message.error(describeFailure(err, t('settings.whep.failed'))))
      .finally(() => setIsSaving(false));
  }

  const endpoint = whepEndpoint(getBaseUrl('http'), settings?.path ?? '');

  return (
    <>
      <div className="text-base">WHEP</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.whep.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.whep.serviceDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        <span className="text-xs text-neutral-500">{t('settings.whep.token')}</span>

        {settings?.enabled && (
          <>
            <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
              <CopyRow label={t('settings.whep.endpoint')} value={endpoint} />
              <div className="flex justify-between gap-3">
                <span className="text-neutral-400">{t('settings.whep.sessions')}</span>
                <span className="text-neutral-300">
                  {settings.sessions} / {settings.maxSessions}
                </span>
              </div>
            </div>

            <CopyBlock title={t('settings.whep.gstreamer')} text={gstreamerCommand(endpoint)} />
            <CopyBlock title={t('settings.whep.obs')} text={obsGstreamerPipeline(endpoint)} />
            <span className="text-xs text-neutral-500">{t('settings.whep.clientsNote')}</span>
          </>
        )}
      </div>
    </>
  );
};
