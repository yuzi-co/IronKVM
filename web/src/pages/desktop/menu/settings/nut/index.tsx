import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Divider, Input, InputNumber, message, Switch, Tag } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/nut.ts';
import type { NutEvent, NutSettings, NutSettingsRequest, NutState, NutStatus } from '@/api/nut.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import { PowerLedSetting } from '../../power/power-led-setting.tsx';
import { StaleNote, StatusTag } from '../components/status-tag.tsx';

// The client polls the UPS every five seconds.
const statePollMs = 10 * 1000;

const maxMinutes = 1440;
const minDelaySeconds = 10;
const maxDelaySeconds = 3600;

// A loose check: the server has the final word.
const hostPattern = /^([A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?|[0-9A-Fa-f:.]*:[0-9A-Fa-f:.]*)$/;

const statusColors: Record<NutStatus, string> = {
  off: 'default',
  unconfigured: 'default',
  connecting: 'blue',
  ok: 'green',
  unreachable: 'red'
};

function formatTime(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

// The NUT client: its switch and settings, what the UPS and the host are doing
// now, and the last things the client did.
export const Nut = () => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<NutSettings | null>(null);
  const [draft, setDraft] = useState<NutSettings | null>(null);
  // A new password, or null to leave the stored one alone.
  const [password, setPassword] = useState('');
  const [state, setState] = useState<NutState | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  // Set when the last poll failed, so the state shown may be old.
  const [isStale, setIsStale] = useState(false);

  const getSettings = useCallback(() => {
    api
      .getNutSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(t('settings.nut.failed'));
          return;
        }
        setSettings(rsp.data);
        setDraft(rsp.data);
      })
      .catch((err) => message.error(describeFailure(err, t('settings.nut.failed'))));
  }, [t]);

  const getState = useCallback(() => {
    api
      .getNutState()
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

  function save(next: NutSettings, extra: { password?: string; clearPassword?: boolean } = {}) {
    const { passwordSet, ...body } = next;
    const request: NutSettingsRequest = { ...body, ...extra };

    setIsSaving(true);
    api
      .setNutSettings(request)
      .then((rsp) => {
        if (rsp.code === -3) {
          // The server refuses to enable the client without the power LED.
          message.error(t('settings.nut.needsLed'));
          return;
        }
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.nut.failed')));
          return;
        }
        const passwordNow = extra.clearPassword ? false : extra.password ? true : passwordSet;
        setSettings({ ...next, passwordSet: passwordNow });
        setDraft((current) =>
          current ? { ...current, enabled: next.enabled, passwordSet: passwordNow } : next
        );
        setPassword('');
        message.success(t('settings.nut.saved'));
        getState();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.nut.failed'))))
      .finally(() => setIsSaving(false));
  }

  // The switch saves at once, with the settings as they were last saved.
  function setEnabled(enabled: boolean) {
    if (!settings) return;
    save({ ...settings, enabled });
  }

  function update<K extends keyof NutSettings>(key: K, value: NutSettings[K]) {
    setDraft((current) => (current ? { ...current, [key]: value } : current));
  }

  const host = draft?.host.trim() ?? '';
  const isHostValid = host === '' || hostPattern.test(host);
  const isDirty =
    !!draft &&
    !!settings &&
    (host !== settings.host ||
      draft.port !== settings.port ||
      draft.ups.trim() !== settings.ups ||
      draft.username.trim() !== settings.username ||
      password !== '' ||
      draft.onBatteryMinutes !== settings.onBatteryMinutes ||
      draft.forceOffMinutes !== settings.forceOffMinutes ||
      draft.powerOnRestore !== settings.powerOnRestore ||
      draft.powerOnMinCharge !== settings.powerOnMinCharge ||
      draft.powerOnDelaySeconds !== settings.powerOnDelaySeconds);

  function saveDraft() {
    if (!draft || !settings || !isHostValid) return;
    save(
      {
        ...draft,
        enabled: settings.enabled,
        host,
        ups: draft.ups.trim(),
        username: draft.username.trim()
      },
      password ? { password } : {}
    );
  }

  function removePassword() {
    if (!settings) return;
    save(settings, { clearPassword: true });
  }

  function formatDuration(seconds: number) {
    return t('settings.nut.duration', {
      minutes: Math.floor(seconds / 60),
      seconds: seconds % 60
    });
  }

  function loginText(s: NutState) {
    if (s.loggedIn) return t('settings.nut.loginSecondary');
    if (s.loginError) return t('settings.nut.loginRefused', { code: s.loginError });
    return t('settings.nut.loginReadOnly');
  }

  function eventText(e: NutEvent) {
    const base = t(`settings.nut.event.${e.kind}`);
    if (!e.detail) return base;
    if (e.kind === 'shutdown') return `${base} (${t(`settings.nut.reason.${e.detail}`)})`;
    return `${base}: ${e.detail}`;
  }

  const stateRows: [string, string][] = state
    ? [
        [t('settings.nut.upsStateLabel'), t(`settings.nut.upsState.${state.ups}`)],
        ...(state.charge !== undefined
          ? ([[t('settings.nut.charge'), `${state.charge}%`]] as [string, string][])
          : []),
        ...(state.runtimeSeconds !== undefined
          ? ([[t('settings.nut.runtime'), formatDuration(state.runtimeSeconds)]] as [
              string,
              string
            ][])
          : []),
        ...(state.onBatterySeconds !== undefined
          ? ([[t('settings.nut.onBatteryFor'), formatDuration(state.onBatterySeconds)]] as [
              string,
              string
            ][])
          : []),
        [t('settings.nut.hostPower'), t(`settings.nut.hostState.${state.host}`)],
        [t('settings.nut.login'), loginText(state)],
        [t('settings.nut.activity'), t(`settings.nut.phase.${state.phase}`)],
        ...(state.powerOnInSeconds !== undefined
          ? ([[t('settings.nut.powerOnIn'), formatDuration(state.powerOnInSeconds)]] as [
              string,
              string
            ][])
          : []),
        [t('settings.nut.lastPoll'), formatTime(state.lastPoll) || t('settings.nut.never')]
      ]
    : [];

  return (
    <>
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.nut.title')}</span>
        {settings && <StatusTag running={settings.enabled} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <PowerLedSetting />

        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.nut.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.nut.serviceDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        <Alert type="warning" showIcon message={t('settings.nut.notice')} />

        {draft && (
          <div className="flex flex-col space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.nut.host')}</span>
                <span className="text-xs text-neutral-500">{t('settings.nut.hostDesc')}</span>
              </div>
              <Input
                style={{ width: 180 }}
                value={draft.host}
                status={isHostValid ? undefined : 'error'}
                placeholder="192.168.1.5"
                onChange={(e) => update('host', e.target.value)}
              />
            </div>
            {!isHostValid && (
              <span className="text-xs text-red-500">{t('settings.nut.hostInvalid')}</span>
            )}

            <div className="flex items-center justify-between">
              <span className="text-sm">{t('settings.nut.port')}</span>
              <InputNumber
                style={{ width: 180 }}
                min={1}
                max={65535}
                precision={0}
                value={draft.port}
                onChange={(value) => update('port', value ?? 3493)}
              />
            </div>

            <div className="flex items-center justify-between">
              <span className="text-sm">{t('settings.nut.upsName')}</span>
              <Input
                style={{ width: 180 }}
                maxLength={64}
                value={draft.ups}
                placeholder="ups"
                onChange={(e) => update('ups', e.target.value)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.nut.username')}</span>
                <span className="text-xs text-neutral-500">
                  {t('settings.nut.credentialsDesc')}
                </span>
              </div>
              <Input
                style={{ width: 180 }}
                maxLength={64}
                autoComplete="off"
                value={draft.username}
                onChange={(e) => update('username', e.target.value)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.nut.password')}</span>
                {settings?.passwordSet && (
                  <span className="text-xs text-neutral-500">{t('settings.nut.passwordSet')}</span>
                )}
              </div>
              <Input.Password
                style={{ width: 180 }}
                maxLength={128}
                autoComplete="new-password"
                value={password}
                placeholder={settings?.passwordSet ? t('settings.nut.passwordKeep') : undefined}
                onChange={(e) => setPassword(e.target.value.replace(/\s/g, ''))}
              />
            </div>
            {settings?.passwordSet && (
              <div className="flex justify-end">
                <Button size="small" danger loading={isSaving} onClick={removePassword}>
                  {t('settings.nut.passwordClear')}
                </Button>
              </div>
            )}

            <Divider className="my-0 opacity-50" />

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.nut.onBattery')}</span>
                <span className="text-xs text-neutral-500">{t('settings.nut.onBatteryDesc')}</span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={0}
                max={maxMinutes}
                precision={0}
                value={draft.onBatteryMinutes}
                addonAfter={t('settings.nut.minutes')}
                onChange={(value) => update('onBatteryMinutes', value ?? 0)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.nut.forceOff')}</span>
                <span className="text-xs text-neutral-500">{t('settings.nut.forceOffDesc')}</span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={0}
                max={maxMinutes}
                precision={0}
                value={draft.forceOffMinutes}
                addonAfter={t('settings.nut.minutes')}
                onChange={(value) => update('forceOffMinutes', value ?? 0)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.nut.powerOn')}</span>
                <span className="text-xs text-neutral-500">{t('settings.nut.powerOnDesc')}</span>
              </div>
              <Switch
                checked={draft.powerOnRestore}
                onChange={(value) => update('powerOnRestore', value)}
              />
            </div>

            {draft.powerOnRestore && (
              <>
                <div className="flex items-center justify-between">
                  <div className="flex flex-col space-y-1 pr-4">
                    <span className="text-sm">{t('settings.nut.minCharge')}</span>
                    <span className="text-xs text-neutral-500">
                      {t('settings.nut.minChargeDesc')}
                    </span>
                  </div>
                  <InputNumber
                    style={{ width: 150 }}
                    min={0}
                    max={100}
                    precision={0}
                    value={draft.powerOnMinCharge}
                    addonAfter="%"
                    onChange={(value) => update('powerOnMinCharge', value ?? 0)}
                  />
                </div>

                <div className="flex items-center justify-between">
                  <div className="flex flex-col space-y-1 pr-4">
                    <span className="text-sm">{t('settings.nut.delay')}</span>
                    <span className="text-xs text-neutral-500">{t('settings.nut.delayDesc')}</span>
                  </div>
                  <InputNumber
                    style={{ width: 150 }}
                    min={minDelaySeconds}
                    max={maxDelaySeconds}
                    precision={0}
                    value={draft.powerOnDelaySeconds}
                    addonAfter={t('settings.nut.secondsUnit')}
                    onChange={(value) => update('powerOnDelaySeconds', value ?? minDelaySeconds)}
                  />
                </div>
              </>
            )}

            <div className="flex justify-end">
              <Button
                type="primary"
                disabled={!isDirty || !isHostValid}
                loading={isSaving}
                onClick={saveDraft}
              >
                {t('settings.nut.save')}
              </Button>
            </div>
          </div>
        )}

        {state && (
          <div className={`flex flex-col space-y-2 ${isStale ? 'opacity-60' : ''}`}>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{t('settings.nut.state')}</span>
              <Tag color={statusColors[state.status]}>
                {t(`settings.nut.status.${state.status}`)}
              </Tag>
            </div>
            {isStale && <StaleNote />}
            {state.phase === 'needsLed' && (
              <Alert type="info" showIcon message={t('settings.nut.ledHint')} />
            )}
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

        {state && (
          <div className="flex flex-col space-y-2">
            <span className="text-sm font-medium">{t('settings.nut.events')}</span>
            {state.events.length === 0 ? (
              <span className="text-xs text-neutral-500">{t('settings.nut.noEvents')}</span>
            ) : (
              <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
                {state.events.map((e, index) => (
                  <div
                    key={`${e.time}-${e.kind}-${index}`}
                    className={
                      index > 0
                        ? 'flex items-start justify-between gap-3 border-t border-neutral-800 px-4 py-2'
                        : 'flex items-start justify-between gap-3 px-4 py-2'
                    }
                  >
                    <span
                      className={
                        e.kind === 'pressFailed' || e.kind === 'unreachable'
                          ? 'text-sm text-red-500'
                          : 'text-sm text-neutral-300'
                      }
                    >
                      {eventText(e)}
                    </span>
                    <span className="shrink-0 text-xs text-neutral-500">{formatTime(e.time)}</span>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </>
  );
};
