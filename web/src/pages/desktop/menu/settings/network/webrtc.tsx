import { useEffect, useState } from 'react';
import { Button, Input } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/network.ts';
import type { WebRTCForwardState } from '@/api/network.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { Section } from '../components/section.tsx';
import { parseUdpPort } from './webrtc-port.ts';

// WebRTCForward sets the fixed UDP port and the public address that let a
// router's port forward carry WebRTC. Both are off by default, and the page
// shows only on a server with the webrtc-port-forward feature.
export const WebRTCForward = () => {
  const { t } = useTranslation();

  const [state, setState] = useState<WebRTCForwardState | null>(null);
  const [port, setPort] = useState('');
  const [address, setAddress] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  function load() {
    api
      .getWebRTCForward()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(rsp.msg || t('settings.network.webrtc.loadFailed'));
          return;
        }
        const data = rsp.data as WebRTCForwardState;
        setState(data);
        setPort(data.udpPort ? String(data.udpPort) : '');
        setAddress(data.publicAddress);
      })
      .catch((err) => setError(describeFailure(err, t('settings.network.webrtc.loadFailed'))));
  }

  useEffect(load, []); // eslint-disable-line react-hooks/exhaustive-deps

  const parsedPort = parseUdpPort(port);
  const portBad = parsedPort === null;
  const needsPort = !portBad && parsedPort === 0 && address.trim() !== '';
  const invalid = portBad || needsPort;

  const isChanged =
    !!state && (parsedPort !== state.udpPort || address.trim() !== state.publicAddress);

  async function save() {
    if (isSaving || invalid || !isChanged) return;
    setError('');
    setMessage('');
    setIsSaving(true);
    try {
      const rsp = await api.setWebRTCForward(parsedPort ?? 0, address.trim());
      if (rsp.code !== 0) {
        setError(
          rsp.msg
            ? `${t('settings.network.webrtc.saveFailed')}: ${rsp.msg}`
            : t('settings.network.webrtc.saveFailed')
        );
        return;
      }
      setMessage(t('settings.network.webrtc.saved'));
      load();
    } catch (err) {
      setError(describeFailure(err, t('settings.network.webrtc.saveFailed')));
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <Section
      title={t('settings.network.webrtc.title')}
      description={t('settings.network.webrtc.description')}
    >
      <span className="text-xs text-yellow-400/80">{t('settings.network.webrtc.warning')}</span>

      <label className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('settings.network.webrtc.udpPort')}</span>
        <Input
          value={port}
          status={portBad || needsPort ? 'error' : undefined}
          placeholder="58000"
          className="w-32 font-mono"
          disabled={!state || isSaving}
          inputMode="numeric"
          onChange={(e) => {
            setPort(e.target.value);
            setError('');
            setMessage('');
          }}
        />
        <span className="text-xs text-neutral-500">{t('settings.network.webrtc.udpPortHint')}</span>
      </label>

      <label className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">
          {t('settings.network.webrtc.publicAddress')}
        </span>
        <Input
          value={address}
          placeholder="kvm.example.com"
          className="font-mono"
          disabled={!state || isSaving}
          spellCheck={false}
          autoComplete="off"
          onChange={(e) => {
            setAddress(e.target.value);
            setError('');
            setMessage('');
          }}
        />
        <span className="text-xs text-neutral-500">
          {t('settings.network.webrtc.publicAddressHint')}
        </span>
      </label>

      {portBad && (
        <span className="text-xs text-red-400">{t('settings.network.webrtc.invalidPort')}</span>
      )}
      {needsPort && (
        <span className="text-xs text-red-400">{t('settings.network.webrtc.needsPort')}</span>
      )}

      {state && (
        <div className="flex flex-col space-y-1 text-xs text-neutral-500">
          <span>
            {state.activePort
              ? t('settings.network.webrtc.active', { port: state.activePort })
              : t('settings.network.webrtc.inactive')}
          </span>
          {state.resolved.length > 0 && (
            <span className="font-mono break-all">
              {t('settings.network.webrtc.offered', { addresses: state.resolved.join(', ') })}
            </span>
          )}
          {state.activePort > 0 && <span>{t('settings.network.webrtc.cost')}</span>}
        </div>
      )}

      <div className="flex flex-wrap gap-2">
        <Button
          type={isChanged ? 'primary' : 'default'}
          loading={isSaving}
          disabled={!state || !isChanged || invalid}
          onClick={save}
        >
          {t('settings.network.webrtc.apply')}
        </Button>
      </div>

      {error && <span className="text-xs text-red-400">{error}</span>}
      {message && <span className="text-xs text-green-500">{message}</span>}
    </Section>
  );
};
