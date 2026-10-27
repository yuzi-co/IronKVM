import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Divider, InputNumber, message, Switch, Tag } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/streaming.ts';
import type { RtspSettings } from '@/api/streaming.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getHostname } from '@/lib/service.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import { CopyRow } from '../components/copy-button.tsx';
import { StaleNote, StatusTag } from '../components/status-tag.tsx';
import { formatTime } from './format.ts';
import { Rows, SettingRow } from './rows.tsx';
import { rtspUrl } from './urls.ts';

// The players change when one connects, which nothing announces.
const pollMs = 5 * 1000;

const defaultPort = 8554;
const defaultUdpPort = 8000;
const defaultTlsPort = 8322;

type Draft = { port: number; udp: boolean; udpPort: number; tlsPort: number };

function draftOf(settings: RtspSettings): Draft {
  return {
    port: settings.port,
    udp: settings.udp,
    udpPort: settings.udpPort,
    tlsPort: settings.tlsPort
  };
}

// The RTSP server (ironkvm-dist#142): a switch for RTSPS and one for plain
// RTSP, their ports, the addresses a player is given, and the players
// connected.
export const Rtsp = () => {
  const { t } = useTranslation();
  const [settings, setSettings] = useState<RtspSettings | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  // Set when the last poll failed, so the players shown may be old.
  const [isStale, setIsStale] = useState(false);

  // keepDraft leaves the fields being edited alone: a poll must not throw
  // them away.
  const load = useCallback((keepDraft = true) => {
    api
      .getRtspSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setIsStale(true);
          return;
        }
        const next: RtspSettings = rsp.data;
        setSettings(next);
        setIsStale(false);
        setDraft((current) => (keepDraft && current ? current : draftOf(next)));
      })
      .catch(() => setIsStale(true));
  }, []);

  useEffect(() => {
    load(false);
    return pollWhileVisible(() => load(true), pollMs);
  }, [load]);

  function save(request: api.RtspSettingsRequest, keepDraft: boolean) {
    setIsSaving(true);
    api
      .setRtspSettings(request)
      .then((rsp) => {
        // -3 means the settings were saved but the server cannot listen.
        if (rsp.code === 0 || rsp.code === -3) load(keepDraft);
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.rtsp.failed')));
          return;
        }
        message.success(t('settings.rtsp.saved'));
      })
      .catch((err) => message.error(describeFailure(err, t('settings.rtsp.failed'))))
      .finally(() => setIsSaving(false));
  }

  // A switch saves at once, with the other settings as they were last saved.
  function setSwitch(key: 'enabled' | 'tls', value: boolean) {
    if (!settings) return;
    save(
      { enabled: settings.enabled, tls: settings.tls, ...draftOf(settings), [key]: value },
      true
    );
  }

  function update<K extends keyof Draft>(key: K, value: Draft[K]) {
    setDraft((current) => (current ? { ...current, [key]: value } : current));
  }

  const isDirty =
    !!draft &&
    !!settings &&
    (draft.port !== settings.port ||
      draft.udp !== settings.udp ||
      draft.udpPort !== settings.udpPort ||
      draft.tlsPort !== settings.tlsPort);

  function saveDraft() {
    if (!draft || !settings) return;
    save({ enabled: settings.enabled, tls: settings.tls, ...draft }, false);
  }

  const host = getHostname();
  const address = settings ? rtspUrl(host, settings.port, settings.path) : '';
  const tlsAddress = settings
    ? rtspUrl(host, settings.tlsPort, settings.path, undefined, true)
    : '';

  // RTSPS comes first once the board has a certificate to serve. Without
  // one, plain RTSP is what works today, and RTSPS says how to get there.
  const preferTls = settings?.certificate ?? true;

  const tlsSwitch = (
    <div key="tls" className="flex flex-col space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1 pr-4">
          <span className="text-sm font-medium">
            {t('settings.rtsp.tls')}
            {settings?.certificate && (
              <Tag color="green" className="ml-2">
                {t('settings.rtsp.recommended')}
              </Tag>
            )}
          </span>
          <span className="text-xs text-neutral-500">{t('settings.rtsp.tlsDesc')}</span>
          {settings && !settings.certificate && (
            <span className="text-xs text-neutral-500">{t('settings.rtsp.noCertificate')}</span>
          )}
        </div>
        <Switch
          checked={settings?.tls ?? false}
          loading={!settings || isSaving}
          onChange={(value) => setSwitch('tls', value)}
        />
      </div>
      {settings?.tls && (
        <>
          <div className="rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
            <CopyRow label={t('settings.rtsp.tlsAddress')} value={tlsAddress} />
          </div>
          <span className="text-xs text-neutral-500">{t('settings.rtsp.tlsHint')}</span>
        </>
      )}
    </div>
  );

  const plainSwitch = (
    <div key="plain" className="flex flex-col space-y-3">
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1 pr-4">
          <span className="text-sm font-medium">{t('settings.rtsp.plain')}</span>
          <span className="text-xs text-neutral-500">{t('settings.rtsp.plainDesc')}</span>
        </div>
        <Switch
          checked={settings?.enabled ?? false}
          loading={!settings || isSaving}
          onChange={(value) => setSwitch('enabled', value)}
        />
      </div>
      <Alert type="warning" showIcon message={t('settings.rtsp.cleartext')} />
      {settings?.enabled && (
        <div className="rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
          <CopyRow label={t('settings.rtsp.address')} value={address} />
        </div>
      )}
    </div>
  );

  return (
    <>
      <div className="flex items-center justify-between">
        <span className="text-base">RTSP</span>
        {settings && <StatusTag running={settings.listening} stale={isStale} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex flex-col space-y-1">
          <span className="text-sm font-medium">{t('settings.rtsp.service')}</span>
          <span className="text-xs text-neutral-500">{t('settings.rtsp.serviceDesc')}</span>
        </div>

        <span className="text-xs text-neutral-500">{t('settings.rtsp.credentials')}</span>

        {preferTls ? [tlsSwitch, plainSwitch] : [plainSwitch, tlsSwitch]}

        {(settings?.enabled || settings?.tls) && (
          <span className="text-xs text-neutral-500">{t('settings.rtsp.addressHint')}</span>
        )}

        {draft && (
          <div className="flex flex-col space-y-4">
            <SettingRow label={t('settings.rtsp.tlsPort')} desc={t('settings.rtsp.tlsPortDesc')}>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={65535}
                precision={0}
                value={draft.tlsPort}
                onChange={(value) => update('tlsPort', value ?? defaultTlsPort)}
              />
            </SettingRow>

            <SettingRow label={t('settings.rtsp.port')} desc={t('settings.rtsp.portDesc')}>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={65535}
                precision={0}
                value={draft.port}
                onChange={(value) => update('port', value ?? defaultPort)}
              />
            </SettingRow>

            <SettingRow
              label={t('settings.rtsp.udp')}
              desc={`${t('settings.rtsp.udpDesc')} ${t('settings.rtsp.udpPlainOnly')}`}
            >
              <Switch checked={draft.udp} onChange={(value) => update('udp', value)} />
            </SettingRow>

            {draft.udp && (
              <SettingRow label={t('settings.rtsp.udpPort')} desc={t('settings.rtsp.udpPortDesc')}>
                <InputNumber
                  style={{ width: 150 }}
                  min={1}
                  max={65534}
                  precision={0}
                  value={draft.udpPort}
                  onChange={(value) => update('udpPort', value ?? defaultUdpPort)}
                />
              </SettingRow>
            )}

            <div className="flex justify-end">
              <Button type="primary" disabled={!isDirty} loading={isSaving} onClick={saveDraft}>
                {t('settings.rtsp.save')}
              </Button>
            </div>
          </div>
        )}

        {settings && (
          <div className={`flex flex-col space-y-2 ${isStale ? 'opacity-60' : ''}`}>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{t('settings.rtsp.players')}</span>
              <Tag color={settings.listening ? 'green' : 'default'}>
                {settings.clients.length} / {settings.maxClients}
              </Tag>
            </div>
            {isStale && <StaleNote />}
            {settings.error && <Alert type="error" showIcon message={settings.error} />}

            {settings.clients.length === 0 ? (
              <span className="text-xs text-neutral-500">{t('settings.rtsp.noPlayers')}</span>
            ) : (
              settings.clients.map((client) => (
                <Rows
                  key={client.address}
                  rows={[
                    [t('settings.rtsp.client'), client.address],
                    ...(client.user
                      ? ([[t('settings.rtsp.user'), client.user]] as [string, string][])
                      : []),
                    [
                      t('settings.rtsp.state'),
                      client.playing ? t('settings.rtsp.playing') : t('settings.rtsp.connecting')
                    ],
                    ...(client.transport
                      ? ([
                          [
                            t('settings.rtsp.transport'),
                            // RTSPS is TCP inside TLS.
                            client.secure
                              ? `${client.transport.toUpperCase()}, TLS`
                              : client.transport.toUpperCase()
                          ]
                        ] as [string, string][])
                      : []),
                    [t('settings.rtsp.since'), formatTime(client.since)]
                  ]}
                />
              ))
            )}
            <span className="text-xs text-neutral-500">{t('settings.rtsp.limits')}</span>
          </div>
        )}
      </div>
    </>
  );
};
