import { useCallback, useEffect, useState } from 'react';
import {
  Alert,
  Button,
  Divider,
  Image,
  Input,
  InputNumber,
  message,
  Select,
  Switch,
  Tag
} from 'antd';
import { RefreshCwIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/watchdog.ts';
import type {
  WatchdogEntry,
  WatchdogSettings,
  WatchdogState,
  WatchdogStatus
} from '@/api/watchdog.ts';

import { PowerLedSetting } from '../../power/power-led-setting.tsx';

// The detector state changes with every sample, which the server takes every
// 10 seconds.
const statePollMs = 10 * 1000;

// The server refuses anything outside these.
const maxMinutes = 1440;
const maxActionsPerHour = 20;

const ipv4 = /^(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)(\.(25[0-5]|2[0-4]\d|1\d\d|[1-9]?\d)){3}$/;

// A loose check: the server has the final word. An IPv6 address has hex
// digits and colons, and may end in an IPv4 address.
function isIPAddress(value: string) {
  if (ipv4.test(value)) return true;
  return value.includes(':') && /^[0-9a-fA-F:.]+$/.test(value);
}

function formatTime(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

const statusColors: Record<WatchdogStatus, string> = {
  off: 'default',
  watching: 'green',
  hostOff: 'default',
  captureOff: 'orange',
  cooldown: 'blue',
  capped: 'red',
  acting: 'red'
};

// The host watchdog: its switch and settings, what the detector sees now, and
// the actions it took, each with the screenshot saved before the press.
export const Watchdog = () => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<WatchdogSettings | null>(null);
  const [draft, setDraft] = useState<WatchdogSettings | null>(null);
  const [state, setState] = useState<WatchdogState | null>(null);
  const [entries, setEntries] = useState<WatchdogEntry[]>([]);
  const [isSaving, setIsSaving] = useState(false);
  const [isLoadingLog, setIsLoadingLog] = useState(false);

  const getSettings = useCallback(() => {
    api
      .getWatchdogSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(t('settings.watchdog.failed'));
          return;
        }
        setSettings(rsp.data);
        setDraft(rsp.data);
      })
      .catch(() => message.error(t('settings.watchdog.failed')));
  }, [t]);

  const getState = useCallback(() => {
    api
      .getWatchdogState()
      .then((rsp) => {
        if (rsp.code === 0) setState(rsp.data);
      })
      .catch(() => {});
  }, []);

  // getLog clears the loading flag; refreshLog is what sets it.
  const getLog = useCallback(() => {
    api
      .getWatchdogLog()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(t('settings.watchdog.failed'));
          return;
        }
        setEntries(rsp.data || []);
      })
      .catch(() => message.error(t('settings.watchdog.failed')))
      .finally(() => setIsLoadingLog(false));
  }, [t]);

  useEffect(() => {
    getSettings();
    getState();
    getLog();

    const timer = window.setInterval(getState, statePollMs);
    return () => window.clearInterval(timer);
  }, [getSettings, getState, getLog]);

  function refreshLog() {
    setIsLoadingLog(true);
    getState();
    getLog();
  }

  function save(next: WatchdogSettings) {
    setIsSaving(true);
    api
      .setWatchdogSettings(next)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(rsp.msg || t('settings.watchdog.failed'));
          return;
        }
        setSettings(next);
        setDraft((current) => (current ? { ...current, enabled: next.enabled } : next));
        message.success(t('settings.watchdog.saved'));
        getState();
      })
      .catch(() => message.error(t('settings.watchdog.failed')))
      .finally(() => setIsSaving(false));
  }

  // The switch saves at once, with the settings as they were last saved.
  function setEnabled(enabled: boolean) {
    if (!settings) return;
    save({ ...settings, enabled });
  }

  function update<K extends keyof WatchdogSettings>(key: K, value: WatchdogSettings[K]) {
    setDraft((current) => (current ? { ...current, [key]: value } : current));
  }

  const pingHost = draft?.pingHost.trim() ?? '';
  const isPingValid = pingHost === '' || isIPAddress(pingHost);
  const isDirty =
    !!draft &&
    !!settings &&
    (draft.timeoutMinutes !== settings.timeoutMinutes ||
      draft.action !== settings.action ||
      draft.cooldownMinutes !== settings.cooldownMinutes ||
      draft.maxPerHour !== settings.maxPerHour ||
      pingHost !== settings.pingHost);

  function saveDraft() {
    if (!draft || !settings || !isPingValid) return;
    save({ ...draft, enabled: settings.enabled, pingHost });
  }

  function formatDuration(seconds: number) {
    return t('settings.watchdog.duration', {
      minutes: Math.floor(seconds / 60),
      seconds: seconds % 60
    });
  }

  function ledText(s: WatchdogState) {
    if (!s.ledConnected) return t('settings.watchdog.ledNotConnected');
    return s.ledOn ? t('settings.watchdog.on') : t('settings.watchdog.off');
  }

  function pingText(s: WatchdogState) {
    if (!s.pingHost) return t('settings.watchdog.pingNotSet');
    return s.pingOK ? t('settings.watchdog.pingReply') : t('settings.watchdog.pingNoReply');
  }

  const stateRows: [string, string][] = state
    ? [
        [
          t('settings.watchdog.signal'),
          state.signal ? t('settings.watchdog.yes') : t('settings.watchdog.no')
        ],
        [t('settings.watchdog.led'), ledText(state)],
        [t('settings.watchdog.ping'), pingText(state)],
        [
          t('settings.watchdog.lastChange'),
          formatTime(state.lastChange) || t('settings.watchdog.never')
        ],
        ...(state.actsInSeconds !== undefined
          ? ([[t('settings.watchdog.actsIn'), formatDuration(state.actsInSeconds)]] as [
              string,
              string
            ][])
          : []),
        [t('settings.watchdog.actionsLastHour'), String(state.actionsLastHour)]
      ]
    : [];

  return (
    <>
      <div className="text-base">{t('settings.watchdog.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <PowerLedSetting />

        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.watchdog.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.watchdog.serviceDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        <Alert type="warning" showIcon message={t('settings.watchdog.stillWarning')} />
        {state && !state.ledConnected && (
          <Alert type="info" showIcon message={t('settings.watchdog.ledHint')} />
        )}

        {draft && (
          <div className="flex flex-col space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.watchdog.timeout')}</span>
                <span className="text-xs text-neutral-500">
                  {t('settings.watchdog.timeoutDesc')}
                </span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={maxMinutes}
                precision={0}
                value={draft.timeoutMinutes}
                addonAfter={t('settings.watchdog.minutes')}
                onChange={(value) => update('timeoutMinutes', value ?? 1)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.watchdog.action')}</span>
                <span className="text-xs text-neutral-500">
                  {t('settings.watchdog.actionDesc')}
                </span>
              </div>
              <Select
                style={{ width: 150 }}
                value={draft.action}
                options={[
                  { value: 'reset', label: t('settings.watchdog.actionReset') },
                  { value: 'power', label: t('settings.watchdog.actionPower') }
                ]}
                onChange={(value) => update('action', value)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.watchdog.cooldown')}</span>
                <span className="text-xs text-neutral-500">
                  {t('settings.watchdog.cooldownDesc')}
                </span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={maxMinutes}
                precision={0}
                value={draft.cooldownMinutes}
                addonAfter={t('settings.watchdog.minutes')}
                onChange={(value) => update('cooldownMinutes', value ?? 1)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.watchdog.maxPerHour')}</span>
                <span className="text-xs text-neutral-500">
                  {t('settings.watchdog.maxPerHourDesc')}
                </span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={maxActionsPerHour}
                precision={0}
                value={draft.maxPerHour}
                onChange={(value) => update('maxPerHour', value ?? 1)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.watchdog.pingHost')}</span>
                <span className="text-xs text-neutral-500">
                  {t('settings.watchdog.pingHostDesc')}
                </span>
              </div>
              <Input
                style={{ width: 150 }}
                value={draft.pingHost}
                status={isPingValid ? undefined : 'error'}
                placeholder="192.168.1.10"
                onChange={(e) => update('pingHost', e.target.value)}
              />
            </div>
            {!isPingValid && (
              <span className="text-xs text-red-500">{t('settings.watchdog.pingHostInvalid')}</span>
            )}

            <div className="flex justify-end">
              <Button
                type="primary"
                disabled={!isDirty || !isPingValid}
                loading={isSaving}
                onClick={saveDraft}
              >
                {t('settings.watchdog.save')}
              </Button>
            </div>
          </div>
        )}

        {state && (
          <div className="flex flex-col space-y-2">
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{t('settings.watchdog.state')}</span>
              <Tag color={statusColors[state.status]}>
                {t(`settings.watchdog.status.${state.status}`)}
              </Tag>
            </div>
            <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
              {stateRows.map(([label, value], index) => (
                <div
                  key={label}
                  className={
                    index > 0
                      ? 'flex items-center justify-between border-t border-neutral-800 px-4 py-2'
                      : 'flex items-center justify-between px-4 py-2'
                  }
                >
                  <span className="text-sm text-neutral-400">{label}</span>
                  <span className="text-sm text-neutral-300">{value}</span>
                </div>
              ))}
            </div>
          </div>
        )}

        <div className="flex flex-col space-y-2">
          <div className="flex items-center justify-between">
            <span className="text-sm font-medium">{t('settings.watchdog.log')}</span>
            <Button
              type="text"
              size="small"
              className="text-neutral-400 hover:text-white"
              loading={isLoadingLog}
              icon={<RefreshCwIcon size={15} />}
              title={t('settings.watchdog.refresh')}
              onClick={refreshLog}
            />
          </div>

          {entries.length === 0 ? (
            <span className="text-xs text-neutral-500">{t('settings.watchdog.noLog')}</span>
          ) : (
            <Image.PreviewGroup>
              <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
                {entries.map((entry, index) => (
                  <div
                    key={entry.id}
                    className={
                      index > 0
                        ? 'flex items-start justify-between gap-3 border-t border-neutral-800 px-4 py-3'
                        : 'flex items-start justify-between gap-3 px-4 py-3'
                    }
                  >
                    <div className="flex min-w-0 flex-col space-y-0.5">
                      <span className="text-sm text-neutral-300">
                        {entry.action === 'power'
                          ? t('settings.watchdog.actionPower')
                          : t('settings.watchdog.actionReset')}
                        {' · '}
                        {formatTime(entry.time)}
                      </span>
                      <span className="text-xs text-neutral-500">
                        {entry.reason === 'noSignal'
                          ? t('settings.watchdog.reasonNoSignal')
                          : t('settings.watchdog.reasonFrozen')}
                        {', '}
                        {t('settings.watchdog.stuckFor', {
                          duration: formatDuration(entry.stuckSeconds)
                        })}
                      </span>
                      {entry.error && (
                        <span className="text-xs text-red-500">
                          {t('settings.watchdog.pressFailed', { error: entry.error })}
                        </span>
                      )}
                    </div>
                    {entry.screenshot ? (
                      <Image
                        width={120}
                        className="rounded"
                        src={api.watchdogScreenshotUrl(entry.id)}
                        alt={formatTime(entry.time)}
                      />
                    ) : (
                      <span className="shrink-0 text-xs text-neutral-400">
                        {t('settings.watchdog.noScreenshot')}
                      </span>
                    )}
                  </div>
                ))}
              </div>
            </Image.PreviewGroup>
          )}
        </div>
      </div>
    </>
  );
};
