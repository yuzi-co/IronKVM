import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Divider, Input, InputNumber, message, Switch, Tag } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/streaming.ts';
import type { SrtSettings } from '@/api/streaming.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getHostname } from '@/lib/service.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import { CopyRow } from '../components/copy-button.tsx';
import { StaleNote, StatusTag } from '../components/status-tag.tsx';
import { formatTime } from './format.ts';
import { Rows, SettingRow } from './rows.tsx';
import { srtUrl } from './urls.ts';

// The callers' link figures change every second; this is often enough to
// watch them without loading the board.
const pollMs = 3 * 1000;

const defaultPort = 8890;
const defaultLatency = 120;
// The server refuses anything outside these, and SRT a passphrase outside
// 10 to 79 characters.
const minLatency = 20;
const maxLatency = 8000;
const minPassphrase = 10;
const maxPassphrase = 79;

type Draft = { port: number; latencyMs: number; passphrase: string };

// The SRT listener (ironkvm-dist#143): its switch, port, latency and
// passphrase, the URL a caller is given, and the callers with SRT's view of
// their link.
export const Srt = () => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<SrtSettings | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  // Set when the last poll failed, so the callers shown may be old.
  const [isStale, setIsStale] = useState(false);

  // keepDraft leaves the fields being edited alone: a poll must not throw
  // them away.
  const load = useCallback((keepDraft = true) => {
    api
      .getSrtSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setIsStale(true);
          return;
        }
        const next: SrtSettings = rsp.data;
        setSettings(next);
        setIsStale(false);
        setDraft((current) =>
          keepDraft && current
            ? current
            : { port: next.port, latencyMs: next.latencyMs, passphrase: '' }
        );
      })
      .catch(() => setIsStale(true));
  }, []);

  useEffect(() => {
    load(false);
    return pollWhileVisible(() => load(true), pollMs);
  }, [load]);

  function save(request: api.SrtSettingsRequest, keepDraft: boolean) {
    setIsSaving(true);
    api
      .setSrtSettings(request)
      .then((rsp) => {
        // -3 means the settings were saved but the listener cannot open.
        if (rsp.code === 0 || rsp.code === -3) load(keepDraft);
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.srt.failed')));
          return;
        }
        message.success(t('settings.srt.saved'));
      })
      .catch((err) => message.error(describeFailure(err, t('settings.srt.failed'))))
      .finally(() => setIsSaving(false));
  }

  const passphrase = draft?.passphrase ?? '';
  const isPassphraseValid =
    passphrase === '' || (passphrase.length >= minPassphrase && passphrase.length <= maxPassphrase);
  const hasPassphrase = passphrase !== '' || !!settings?.passphraseSet;

  // The switch saves at once, with the settings as they were last saved. It
  // cannot turn SRT on before there is a passphrase.
  function setEnabled(enabled: boolean) {
    if (!settings) return;
    if (enabled && !settings.passphraseSet) {
      message.error(t('settings.srt.passphraseFirst'));
      return;
    }
    save({ enabled, port: settings.port, latencyMs: settings.latencyMs }, true);
  }

  function update<K extends keyof Draft>(key: K, value: Draft[K]) {
    setDraft((current) => (current ? { ...current, [key]: value } : current));
  }

  const isDirty =
    !!draft &&
    !!settings &&
    (draft.port !== settings.port || draft.latencyMs !== settings.latencyMs || passphrase !== '');

  function saveDraft() {
    if (!draft || !settings || !isPassphraseValid) return;
    save(
      {
        enabled: settings.enabled,
        port: draft.port,
        latencyMs: draft.latencyMs,
        passphrase: passphrase || undefined
      },
      false
    );
  }

  const address = settings ? srtUrl(getHostname(), settings.port) : '';

  return (
    <>
      <div className="flex items-center justify-between">
        <span className="text-base">SRT</span>
        {settings && <StatusTag running={settings.listening} stale={isStale} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.srt.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.srt.serviceDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        <span className="text-xs text-neutral-500">{t('settings.srt.credentials')}</span>

        {settings?.enabled && (
          <div className="flex flex-col space-y-3">
            <div className="rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
              <CopyRow label={t('settings.srt.address')} value={address} />
            </div>
            <span className="text-xs text-neutral-500">{t('settings.srt.addressHint')}</span>
          </div>
        )}

        {draft && (
          <div className="flex flex-col space-y-4">
            <SettingRow
              label={t('settings.srt.passphrase')}
              desc={settings?.passphraseSet ? t('settings.srt.passphraseSet') : undefined}
            >
              <Input.Password
                style={{ width: 200 }}
                maxLength={maxPassphrase}
                autoComplete="new-password"
                value={draft.passphrase}
                status={isPassphraseValid && hasPassphrase ? undefined : 'error'}
                onChange={(e) => update('passphrase', e.target.value)}
              />
            </SettingRow>
            {(!isPassphraseValid || !hasPassphrase) && (
              <span className="text-xs text-red-500">{t('settings.srt.passphraseInvalid')}</span>
            )}

            <SettingRow label={t('settings.srt.port')} desc={t('settings.srt.portDesc')}>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={65535}
                precision={0}
                value={draft.port}
                onChange={(value) => update('port', value ?? defaultPort)}
              />
            </SettingRow>

            <SettingRow label={t('settings.srt.latency')} desc={t('settings.srt.latencyDesc')}>
              <InputNumber
                style={{ width: 150 }}
                min={minLatency}
                max={maxLatency}
                precision={0}
                value={draft.latencyMs}
                addonAfter="ms"
                onChange={(value) => update('latencyMs', value ?? defaultLatency)}
              />
            </SettingRow>

            <div className="flex justify-end">
              <Button
                type="primary"
                disabled={!isDirty || !isPassphraseValid}
                loading={isSaving}
                onClick={saveDraft}
              >
                {t('settings.srt.save')}
              </Button>
            </div>
          </div>
        )}

        {settings && (
          <div className={`flex flex-col space-y-2 ${isStale ? 'opacity-60' : ''}`}>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{t('settings.srt.callers')}</span>
              <Tag color={settings.listening ? 'green' : 'default'}>
                {settings.clients.length} / {settings.maxClients}
              </Tag>
            </div>
            {isStale && <StaleNote />}
            {settings.error && <Alert type="error" showIcon message={settings.error} />}

            {settings.clients.length === 0 ? (
              <span className="text-xs text-neutral-500">{t('settings.srt.noCallers')}</span>
            ) : (
              settings.clients.map((client) => (
                <Rows
                  key={client.address}
                  rows={[
                    [t('settings.srt.client'), client.address],
                    ...(client.user
                      ? ([[t('settings.srt.user'), client.user]] as [string, string][])
                      : []),
                    [t('settings.srt.rtt'), `${client.rttMs.toFixed(1)} ms`],
                    [t('settings.srt.rate'), `${client.sendRateMbps.toFixed(2)} Mbit/s`],
                    [
                      t('settings.srt.packets'),
                      t('settings.srt.packetCounts', {
                        sent: client.sent,
                        lost: client.lost,
                        retransmitted: client.retransmitted,
                        dropped: client.dropped
                      })
                    ],
                    [t('settings.srt.since'), formatTime(client.since)]
                  ]}
                />
              ))
            )}
            <span className="text-xs text-neutral-500">{t('settings.srt.limits')}</span>
          </div>
        )}
      </div>
    </>
  );
};
