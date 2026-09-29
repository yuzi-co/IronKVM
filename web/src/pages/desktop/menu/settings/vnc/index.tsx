import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Divider, Input, InputNumber, message, Modal, Switch, Tag } from 'antd';
import { RefreshCwIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/vnc.ts';
import type { VncSettings, VncState } from '@/api/vnc.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getHostname } from '@/lib/service.ts';

import { CopyRow } from '../components/copy-button.tsx';
import { StaleNote, StatusTag } from '../components/status-tag.tsx';

// The session state changes when a client connects, which nothing announces.
const statePollMs = 5 * 1000;

// The server refuses anything outside these.
const maxFrameRate = 60;
const minPasswordLength = 6;
const maxPasswordLength = 8;

type Draft = {
  port: number;
  maxFps: number;
  vncAuth: boolean;
  password: string;
};

function formatTime(value?: string) {
  if (!value) return '';
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleString();
}

// The VNC server: its switch and settings, the listener, and the open session
// with a button that ends it.
export const Vnc = () => {
  const { t } = useTranslation();
  const [modal, contextHolder] = Modal.useModal();
  const [settings, setSettings] = useState<VncSettings | null>(null);
  const [draft, setDraft] = useState<Draft | null>(null);
  const [state, setState] = useState<VncState | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const [isLoadingState, setIsLoadingState] = useState(false);
  // Set when the last poll failed, so the state shown may be old.
  const [isStale, setIsStale] = useState(false);

  // keepDraft leaves the fields being edited alone. The switch saves only
  // enabled, and reading the settings back must not throw away the rest.
  const getSettings = useCallback(
    (keepDraft = false) => {
      api
        .getVncSettings()
        .then((rsp) => {
          if (rsp.code !== 0) {
            message.error(t('settings.vnc.failed'));
            return;
          }
          const next: VncSettings = rsp.data;
          setSettings(next);
          setDraft((current) =>
            keepDraft && current
              ? current
              : { port: next.port, maxFps: next.maxFps, vncAuth: next.vncAuth, password: '' }
          );
        })
        .catch((err) => message.error(describeFailure(err, t('settings.vnc.failed'))));
    },
    [t]
  );

  // getState clears the loading flag; refreshState is what sets it.
  const getState = useCallback(() => {
    api
      .getVncState()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setIsStale(true);
          return;
        }
        setState(rsp.data);
        setIsStale(false);
      })
      .catch(() => setIsStale(true))
      .finally(() => setIsLoadingState(false));
  }, []);

  useEffect(() => {
    getSettings();
    getState();

    const timer = window.setInterval(getState, statePollMs);
    return () => window.clearInterval(timer);
  }, [getSettings, getState]);

  function refreshState() {
    setIsLoadingState(true);
    getState();
  }

  function save(request: api.VncSettingsRequest, keepDraft = false) {
    setIsSaving(true);
    api
      .setVncSettings(request)
      .then((rsp) => {
        // -3 means the settings were saved but the server cannot listen, so
        // the page shows what the server holds now in that case as well.
        if (rsp.code === 0 || rsp.code === -3) {
          getSettings(keepDraft);
        }
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.vnc.failed')));
          return;
        }
        message.success(t('settings.vnc.saved'));
      })
      .catch((err) => message.error(describeFailure(err, t('settings.vnc.failed'))))
      .finally(() => {
        setIsSaving(false);
        getState();
      });
  }

  // The switch saves at once, with the settings as they were last saved.
  function setEnabled(enabled: boolean) {
    if (!settings) return;
    save(
      {
        enabled,
        port: settings.port,
        maxFps: settings.maxFps,
        vncAuth: settings.vncAuth
      },
      true
    );
  }

  function update<K extends keyof Draft>(key: K, value: Draft[K]) {
    setDraft((current) => (current ? { ...current, [key]: value } : current));
  }

  const password = draft?.password ?? '';
  const isPasswordValid =
    password === '' ||
    (password.length >= minPasswordLength && password.length <= maxPasswordLength);
  // Plain authentication needs a password, either the one set or a new one.
  const hasPassword = password !== '' || !!settings?.passwordSet;
  const isValid = isPasswordValid && (!draft?.vncAuth || hasPassword);
  const isDirty =
    !!draft &&
    !!settings &&
    (draft.port !== settings.port ||
      draft.maxFps !== settings.maxFps ||
      draft.vncAuth !== settings.vncAuth ||
      password !== '');

  function saveDraft() {
    if (!draft || !settings || !isValid) return;
    save({
      enabled: settings.enabled,
      port: draft.port,
      maxFps: draft.maxFps,
      vncAuth: draft.vncAuth,
      password: password || undefined
    });
  }

  function disconnect() {
    modal.confirm({
      title: t('settings.vnc.disconnectConfirmTitle'),
      content: (
        <span className="text-sm text-neutral-400">{t('settings.vnc.disconnectConfirmDesc')}</span>
      ),
      okText: t('settings.vnc.okBtn'),
      cancelText: t('settings.vnc.cancelBtn'),
      onOk: async () => {
        try {
          const rsp = await api.disconnectVnc();
          if (rsp.code !== 0) {
            message.error(describeFailure(rsp, t('settings.vnc.failed')));
          }
        } catch (err) {
          message.error(describeFailure(err, t('settings.vnc.failed')));
        } finally {
          getState();
        }
      }
    });
  }

  // What a client connects to.
  const host = getHostname();
  const vncPort = settings?.port ?? 5900;
  const target = `${host}:${vncPort}`;

  const session = state?.session;
  const sessionRows: [string, string][] = session
    ? [
        [t('settings.vnc.client'), session.client],
        ...(session.user ? ([[t('settings.vnc.user'), session.user]] as [string, string][]) : []),
        [
          t('settings.vnc.method'),
          session.method === 'vnc' ? t('settings.vnc.methodVnc') : t('settings.vnc.methodVencrypt')
        ],
        [t('settings.vnc.since'), formatTime(session.since)],
        [t('settings.vnc.resolution'), `${session.width} × ${session.height}`],
        [t('settings.vnc.framesSent'), String(session.framesSent)]
      ]
    : [];

  return (
    <>
      {contextHolder}
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.vnc.title')}</span>
        {state && <StatusTag running={state.listening} stale={isStale} />}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex items-center justify-between">
          <div className="flex flex-col space-y-1 pr-4">
            <span className="text-sm font-medium">{t('settings.vnc.service')}</span>
            <span className="text-xs text-neutral-500">{t('settings.vnc.serviceDesc')}</span>
          </div>
          <Switch
            checked={settings?.enabled ?? false}
            loading={!settings || isSaving}
            onChange={setEnabled}
          />
        </div>

        <span className="text-xs text-neutral-500">{t('settings.vnc.credentials')}</span>

        {settings?.enabled && (
          <div className="flex flex-col space-y-3">
            <div className="rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
              <CopyRow label={t('settings.vnc.address')} value={target} />
            </div>
            <span className="text-xs text-neutral-500">{t('settings.vnc.certHint')}</span>
          </div>
        )}

        {draft && (
          <div className="flex flex-col space-y-4">
            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.vnc.port')}</span>
                <span className="text-xs text-neutral-500">{t('settings.vnc.portDesc')}</span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={65535}
                precision={0}
                value={draft.port}
                onChange={(value) => update('port', value ?? 5900)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.vnc.maxFps')}</span>
                <span className="text-xs text-neutral-500">{t('settings.vnc.maxFpsDesc')}</span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={1}
                max={maxFrameRate}
                precision={0}
                value={draft.maxFps}
                addonAfter={t('screen.fps')}
                onChange={(value) => update('maxFps', value ?? 1)}
              />
            </div>

            <div className="flex items-center justify-between">
              <div className="flex flex-col space-y-1 pr-4">
                <span className="text-sm">{t('settings.vnc.vncAuth')}</span>
                <span className="text-xs text-neutral-500">{t('settings.vnc.vncAuthDesc')}</span>
              </div>
              <Switch checked={draft.vncAuth} onChange={(value) => update('vncAuth', value)} />
            </div>

            {draft.vncAuth && (
              <>
                <Alert type="warning" showIcon message={t('settings.vnc.vncAuthWarning')} />

                <div className="flex items-center justify-between">
                  <div className="flex flex-col space-y-1 pr-4">
                    <span className="text-sm">{t('settings.vnc.password')}</span>
                    {settings?.passwordSet && (
                      <span className="text-xs text-neutral-500">
                        {t('settings.vnc.passwordSet')}
                      </span>
                    )}
                  </div>
                  <Input.Password
                    style={{ width: 150 }}
                    maxLength={maxPasswordLength}
                    autoComplete="new-password"
                    value={draft.password}
                    status={isPasswordValid && hasPassword ? undefined : 'error'}
                    onChange={(e) => update('password', e.target.value)}
                  />
                </div>
                {(!isPasswordValid || !hasPassword) && (
                  <span className="text-xs text-red-500">{t('settings.vnc.passwordInvalid')}</span>
                )}
              </>
            )}

            <div className="flex justify-end">
              <Button
                type="primary"
                disabled={!isDirty || !isValid}
                loading={isSaving}
                onClick={saveDraft}
              >
                {t('settings.vnc.save')}
              </Button>
            </div>
          </div>
        )}

        {state && (
          <div className={`flex flex-col space-y-2 ${isStale ? 'opacity-60' : ''}`}>
            <div className="flex items-center justify-between">
              <span className="text-sm font-medium">{t('settings.vnc.state')}</span>
              <div className="flex items-center space-x-1">
                <Tag color={state.listening ? 'green' : 'default'}>
                  {state.listening
                    ? t('settings.vnc.listening', { port: state.port })
                    : t('settings.vnc.notListening')}
                </Tag>
                <Button
                  type="text"
                  size="small"
                  className="text-neutral-400 hover:text-white"
                  loading={isLoadingState}
                  icon={<RefreshCwIcon size={15} />}
                  title={t('settings.vnc.refresh')}
                  onClick={refreshState}
                />
              </div>
            </div>

            {isStale && <StaleNote />}
            {state.error && <Alert type="error" showIcon message={state.error} />}

            {session ? (
              <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
                {sessionRows.map(([label, value], index) => (
                  <div
                    key={label}
                    className={
                      index > 0
                        ? 'flex items-center justify-between border-t border-neutral-800 px-4 py-2'
                        : 'flex items-center justify-between px-4 py-2'
                    }
                  >
                    <span className="text-sm text-neutral-400">{label}</span>
                    <span className="truncate pl-4 text-sm text-neutral-300">{value}</span>
                  </div>
                ))}
                <div className="flex justify-end border-t border-neutral-800 px-4 py-2">
                  <Button size="small" danger onClick={disconnect}>
                    {t('settings.vnc.disconnect')}
                  </Button>
                </div>
              </div>
            ) : (
              <span className="text-xs text-neutral-500">{t('settings.vnc.noSession')}</span>
            )}

            {state.lastError && (
              <span className="text-xs text-neutral-500">
                {t('settings.vnc.lastError', { error: state.lastError })}
              </span>
            )}
          </div>
        )}
      </div>
    </>
  );
};
