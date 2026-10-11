import { useCallback, useEffect, useState } from 'react';
import { Button, Divider, Input, InputNumber, message, Switch, Tag } from 'antd';
import { RefreshCwIcon, SendIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/notify.ts';
import type {
  NotifyChannel,
  NotifyEventType,
  NotifySave,
  NotifySettings,
  NotifyState
} from '@/api/notify.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import { StaleNote, StatusTag } from '../components/status-tag.tsx';

// The page lists what was sent lately; a minute is often enough.
const statePollMs = 30 * 1000;

const maxCooldown = 86400;

// A secret the page has typed but not saved. Undefined means "keep what is
// stored", an empty string means "remove it".
type Edits = { webhookUrl?: string; ntfyToken?: string; botToken?: string };

const eventKeys: Record<NotifyEventType, string> = {
  'signal.lost': 'signalLost',
  'signal.restored': 'signalRestored',
  'watchdog.action': 'watchdogAction',
  'login.failures': 'loginFailures',
  'atx.press': 'atxPress',
  'failsafe.entered': 'failsafeEntered',
  'hostlog.match': 'hostLogMatch'
};

const channels: NotifyChannel[] = ['webhook', 'ntfy', 'telegram'];

function formatTime(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

function toSave(s: NotifySettings, edits: Edits): NotifySave {
  return {
    enabled: s.enabled,
    snapshot: s.snapshot,
    events: s.events,
    loginThreshold: s.loginThreshold,
    loginWindowSeconds: s.loginWindowSeconds,
    webhook: {
      enabled: s.webhook.enabled,
      url: edits.webhookUrl,
      includeSnapshot: s.webhook.includeSnapshot
    },
    ntfy: {
      enabled: s.ntfy.enabled,
      server: s.ntfy.server,
      topic: s.ntfy.topic,
      token: edits.ntfyToken
    },
    telegram: { enabled: s.telegram.enabled, botToken: edits.botToken, chatId: s.telegram.chatId }
  };
}

// Notifications: where alerts go and which events send them. A token is
// never read back: the page only learns that one is saved.
export const Notifications = () => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<NotifySettings | null>(null);
  const [draft, setDraft] = useState<NotifySettings | null>(null);
  const [edits, setEdits] = useState<Edits>({});
  const [state, setState] = useState<NotifyState | null>(null);
  const [isStale, setIsStale] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [testing, setTesting] = useState<NotifyChannel | 'all' | null>(null);

  const getSettings = useCallback(() => {
    api
      .getNotifySettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.notifications.failed')));
          return;
        }
        setSettings(rsp.data);
        setDraft(rsp.data);
        setEdits({});
      })
      .catch((err) => message.error(describeFailure(err, t('settings.notifications.failed'))));
  }, [t]);

  const getState = useCallback(() => {
    api
      .getNotifyState()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setIsStale(true);
          return;
        }
        setState(rsp.data);
        setIsStale(false);
      })
      .catch(() => setIsStale(true));
  }, []);

  useEffect(() => {
    getSettings();
    getState();
    return pollWhileVisible(getState, statePollMs);
  }, [getSettings, getState]);

  function save(next: NotifySettings, withEdits: Edits) {
    setIsSaving(true);
    api
      .setNotifySettings(toSave(next, withEdits))
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.notifications.failed')));
          return;
        }
        setSettings(rsp.data);
        setDraft(rsp.data);
        setEdits({});
        message.success(t('settings.notifications.saved'));
        getState();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.notifications.failed'))))
      .finally(() => setIsSaving(false));
  }

  // The master switch saves at once, with the settings as last saved.
  function setEnabled(enabled: boolean) {
    if (!settings) return;
    save({ ...settings, enabled }, {});
  }

  function update(change: Partial<NotifySettings>) {
    setDraft((current) => (current ? { ...current, ...change } : current));
  }

  function updateRule(
    type: NotifyEventType,
    change: Partial<{ enabled: boolean; cooldownSeconds: number }>
  ) {
    setDraft((current) =>
      current
        ? {
            ...current,
            events: { ...current.events, [type]: { ...current.events[type], ...change } }
          }
        : current
    );
  }

  const isDirty =
    !!draft &&
    !!settings &&
    (JSON.stringify({ ...draft, enabled: settings.enabled }) !== JSON.stringify(settings) ||
      Object.values(edits).some((v) => v !== undefined));

  function sendTest(channel: NotifyChannel | '') {
    setTesting(channel === '' ? 'all' : channel);
    api
      .sendNotifyTest(channel)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.notifications.testFailed')));
          return;
        }
        const results = rsp.data as Record<string, string>;
        for (const [name, result] of Object.entries(results)) {
          if (result === 'ok') {
            message.success(
              t('settings.notifications.testOk', { channel: t(`settings.notifications.${name}`) })
            );
          } else {
            message.error(
              t('settings.notifications.testChannelFailed', {
                channel: t(`settings.notifications.${name}`),
                error: result
              })
            );
          }
        }
        getState();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.notifications.testFailed'))))
      .finally(() => setTesting(null));
  }

  // Secrets: what the box shows, and whether the channel can be tested.
  function secretPlaceholder(isSet: boolean, edit: string | undefined, hint?: string) {
    if (edit === '') return t('settings.notifications.willRemove');
    if (isSet)
      return hint
        ? t('settings.notifications.savedHint', { hint })
        : t('settings.notifications.savedSecret');
    return '';
  }

  function secretRow(
    label: string,
    isSet: boolean,
    edit: string | undefined,
    onEdit: (value: string | undefined) => void,
    placeholder: string,
    hint?: string
  ) {
    return (
      <div className="flex items-center justify-between gap-3">
        <span className="text-sm">{label}</span>
        <div className="flex items-center gap-2">
          <Input.Password
            style={{ width: 220 }}
            autoComplete="off"
            value={edit ?? ''}
            placeholder={secretPlaceholder(isSet, edit, hint) || placeholder}
            onChange={(e) => onEdit(e.target.value === '' ? undefined : e.target.value)}
          />
          {isSet && edit !== '' && (
            <Button size="small" type="text" danger onClick={() => onEdit('')}>
              {t('settings.notifications.remove')}
            </Button>
          )}
        </div>
      </div>
    );
  }

  function testButton(channel: NotifyChannel, ready: boolean) {
    return (
      <Button
        size="small"
        icon={<SendIcon size={14} />}
        loading={testing === channel}
        disabled={!ready || isDirty || testing !== null}
        title={isDirty ? t('settings.notifications.saveFirst') : undefined}
        onClick={() => sendTest(channel)}
      >
        {t('settings.notifications.sendTest')}
      </Button>
    );
  }

  const channelReady = settings
    ? {
        webhook: settings.webhook.urlSet,
        ntfy: settings.ntfy.topic !== '',
        telegram: settings.telegram.botTokenSet && settings.telegram.chatId !== ''
      }
    : { webhook: false, ntfy: false, telegram: false };

  return (
    <>
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.notifications.title')}</span>
        {settings && <StatusTag running={settings.enabled} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.notifications.service')}</span>
            <span className="text-xs text-neutral-500">
              {t('settings.notifications.serviceDesc')}
            </span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        {draft && (
          <>
            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.notifications.snapshot')}</span>
                <span className="text-xs text-neutral-500">
                  {t('settings.notifications.snapshotDesc')}
                </span>
              </div>
              <Switch checked={draft.snapshot} onChange={(snapshot) => update({ snapshot })} />
            </div>

            <div className="flex flex-col space-y-3">
              <span className="text-sm font-medium">{t('settings.notifications.events')}</span>
              <span className="text-xs text-neutral-500">
                {t('settings.notifications.eventsDesc')}
              </span>
              <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
                {api.NOTIFY_EVENTS.map((type, index) => (
                  <div
                    key={type}
                    className={
                      index > 0
                        ? 'flex flex-col gap-2 border-t border-neutral-800 px-4 py-3'
                        : 'flex flex-col gap-2 px-4 py-3'
                    }
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div className="flex items-center gap-3">
                        <Switch
                          size="small"
                          checked={draft.events[type].enabled}
                          onChange={(enabled) => updateRule(type, { enabled })}
                        />
                        <span className="text-sm">
                          {t(`settings.notifications.event.${eventKeys[type]}`)}
                        </span>
                      </div>
                      <InputNumber
                        style={{ width: 150 }}
                        min={0}
                        max={maxCooldown}
                        precision={0}
                        value={draft.events[type].cooldownSeconds}
                        addonAfter={t('settings.notifications.seconds')}
                        title={t('settings.notifications.cooldown')}
                        onChange={(value) => updateRule(type, { cooldownSeconds: value ?? 0 })}
                      />
                    </div>
                    {type === 'login.failures' && (
                      <div className="flex flex-wrap items-center gap-3 ps-10 text-xs text-neutral-400">
                        <InputNumber
                          size="small"
                          style={{ width: 90 }}
                          min={1}
                          max={1000}
                          precision={0}
                          value={draft.loginThreshold}
                          onChange={(value) => update({ loginThreshold: value ?? 1 })}
                        />
                        <span>{t('settings.notifications.loginWithin')}</span>
                        <InputNumber
                          size="small"
                          style={{ width: 130 }}
                          min={10}
                          max={86400}
                          precision={0}
                          value={draft.loginWindowSeconds}
                          addonAfter={t('settings.notifications.seconds')}
                          onChange={(value) => update({ loginWindowSeconds: value ?? 10 })}
                        />
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </div>

            <div className="flex flex-col space-y-3">
              <span className="text-sm font-medium">{t('settings.notifications.channels')}</span>

              <div className="flex flex-col space-y-3 rounded-xl border border-neutral-700/50 bg-neutral-800/40 p-4">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <span className="text-sm">{t('settings.notifications.webhook')}</span>
                    <span className="text-xs text-neutral-500">
                      {t('settings.notifications.webhookDesc')}
                    </span>
                  </div>
                  <Switch
                    checked={draft.webhook.enabled}
                    onChange={(enabled) => update({ webhook: { ...draft.webhook, enabled } })}
                  />
                </div>
                {secretRow(
                  t('settings.notifications.webhookUrl'),
                  draft.webhook.urlSet,
                  edits.webhookUrl,
                  (webhookUrl) => setEdits((e) => ({ ...e, webhookUrl })),
                  'https://example.com/hook',
                  draft.webhook.urlHint
                )}
                <div className="flex items-center justify-between">
                  <span className="text-sm">{t('settings.notifications.webhookSnapshot')}</span>
                  <Switch
                    size="small"
                    checked={draft.webhook.includeSnapshot}
                    onChange={(includeSnapshot) =>
                      update({ webhook: { ...draft.webhook, includeSnapshot } })
                    }
                  />
                </div>
                <div className="flex justify-end">
                  {testButton('webhook', channelReady.webhook)}
                </div>
              </div>

              <div className="flex flex-col space-y-3 rounded-xl border border-neutral-700/50 bg-neutral-800/40 p-4">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <span className="text-sm">{t('settings.notifications.ntfy')}</span>
                    <span className="text-xs text-neutral-500">
                      {t('settings.notifications.ntfyDesc')}
                    </span>
                  </div>
                  <Switch
                    checked={draft.ntfy.enabled}
                    onChange={(enabled) => update({ ntfy: { ...draft.ntfy, enabled } })}
                  />
                </div>
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm">{t('settings.notifications.ntfyServer')}</span>
                  <Input
                    style={{ width: 220 }}
                    value={draft.ntfy.server}
                    placeholder="https://ntfy.sh"
                    onChange={(e) => update({ ntfy: { ...draft.ntfy, server: e.target.value } })}
                  />
                </div>
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm">{t('settings.notifications.ntfyTopic')}</span>
                  <Input
                    style={{ width: 220 }}
                    value={draft.ntfy.topic}
                    onChange={(e) =>
                      update({ ntfy: { ...draft.ntfy, topic: e.target.value.trim() } })
                    }
                  />
                </div>
                {secretRow(
                  t('settings.notifications.ntfyToken'),
                  draft.ntfy.tokenSet,
                  edits.ntfyToken,
                  (ntfyToken) => setEdits((e) => ({ ...e, ntfyToken })),
                  'tk_...'
                )}
                <div className="flex justify-end">{testButton('ntfy', channelReady.ntfy)}</div>
              </div>

              <div className="flex flex-col space-y-3 rounded-xl border border-neutral-700/50 bg-neutral-800/40 p-4">
                <div className="flex items-center justify-between">
                  <div className="flex flex-col">
                    <span className="text-sm">{t('settings.notifications.telegram')}</span>
                    <span className="text-xs text-neutral-500">
                      {t('settings.notifications.telegramDesc')}
                    </span>
                  </div>
                  <Switch
                    checked={draft.telegram.enabled}
                    onChange={(enabled) => update({ telegram: { ...draft.telegram, enabled } })}
                  />
                </div>
                {secretRow(
                  t('settings.notifications.botToken'),
                  draft.telegram.botTokenSet,
                  edits.botToken,
                  (botToken) => setEdits((e) => ({ ...e, botToken })),
                  '123456:ABC...'
                )}
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm">{t('settings.notifications.chatId')}</span>
                  <Input
                    style={{ width: 220 }}
                    value={draft.telegram.chatId}
                    placeholder="123456789"
                    onChange={(e) =>
                      update({ telegram: { ...draft.telegram, chatId: e.target.value.trim() } })
                    }
                  />
                </div>
                <div className="flex justify-end">
                  {testButton('telegram', channelReady.telegram)}
                </div>
              </div>
            </div>

            <div className="flex justify-end">
              <Button
                type="primary"
                disabled={!isDirty}
                loading={isSaving}
                onClick={() => save(draft, edits)}
              >
                {t('settings.notifications.save')}
              </Button>
            </div>
          </>
        )}

        <div className={`flex flex-col space-y-2 ${isStale ? 'opacity-60' : ''}`}>
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">{t('settings.notifications.recent')}</span>
            <Button
              type="text"
              size="small"
              className="text-neutral-400 hover:text-white"
              icon={<RefreshCwIcon size={15} />}
              title={t('settings.notifications.refresh')}
              onClick={getState}
            />
          </div>
          {isStale && <StaleNote />}

          {state && channels.some((c) => state.channels[c]) && (
            <div className="flex flex-wrap gap-2">
              {channels.map((c) => {
                const result = state.channels[c];
                if (!result) return null;
                return (
                  <Tag key={c} color={result.ok ? 'green' : 'red'} title={result.error}>
                    {t(`settings.notifications.${c}`)}
                    {': '}
                    {result.ok
                      ? t('settings.notifications.lastOk')
                      : (result.error ?? t('settings.notifications.lastFailed'))}
                  </Tag>
                );
              })}
            </div>
          )}

          {!state || state.recent.length === 0 ? (
            <span className="text-xs text-neutral-500">{t('settings.notifications.noRecent')}</span>
          ) : (
            <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
              {state.recent.map((entry, index) => (
                <div
                  key={`${entry.time}-${index}`}
                  className={
                    index > 0
                      ? 'flex items-center justify-between gap-3 border-t border-neutral-800 px-4 py-2'
                      : 'flex items-center justify-between gap-3 px-4 py-2'
                  }
                >
                  <div className="flex min-w-0 flex-col">
                    <span className="truncate text-sm text-neutral-300">{entry.title}</span>
                    <span className="text-xs text-neutral-500">{formatTime(entry.time)}</span>
                  </div>
                  {entry.suppressed ? (
                    <Tag className="me-0">{t('settings.notifications.heldBack')}</Tag>
                  ) : (
                    <div className="flex flex-wrap justify-end gap-1">
                      {Object.entries(entry.channels ?? {}).map(([name, result]) => (
                        <Tag
                          key={name}
                          className="me-0"
                          color={result === 'ok' ? 'green' : 'red'}
                          title={result}
                        >
                          {t(`settings.notifications.${name}`)}
                        </Tag>
                      ))}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          <div className="flex justify-end">
            <Button
              size="small"
              loading={testing === 'all'}
              disabled={isDirty || testing !== null}
              onClick={() => sendTest('')}
            >
              {t('settings.notifications.sendTestAll')}
            </Button>
          </div>
        </div>
      </div>
    </>
  );
};
