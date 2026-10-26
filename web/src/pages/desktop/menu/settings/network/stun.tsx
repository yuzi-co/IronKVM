import { useEffect, useState } from 'react';
import { Button, Input, Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/network.ts';
import type { StunServerState } from '@/api/network.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { Section } from '../components/section.tsx';
import { isStunServer, STUN_DISABLED, stunToSave } from './stun-server.ts';

// StunServer shows the STUN server WebRTC uses, with a switch to turn it off.
// The browser and the board both contact it for every WebRTC session. Each
// session reads the setting when it starts, so a save needs no restart.
export const StunServer = () => {
  const { t } = useTranslation();

  const [state, setState] = useState<StunServerState | null>(null);
  const [enabled, setEnabled] = useState(true);
  const [server, setServer] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  function show(data: StunServerState) {
    setState(data);
    const off = data.stun === STUN_DISABLED;
    setEnabled(!off);
    setServer(off ? data.default : data.stun);
  }

  function load() {
    api
      .getStunServer()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(rsp.msg || t('settings.network.stun.loadFailed'));
          return;
        }
        show(rsp.data as StunServerState);
      })
      .catch((err) => setError(describeFailure(err, t('settings.network.stun.loadFailed'))));
  }

  useEffect(load, []); // eslint-disable-line react-hooks/exhaustive-deps

  const value = stunToSave(enabled, server);
  const invalid = enabled && !isStunServer(server);
  const isChanged = !!state && value !== state.stun;

  function edited() {
    setError('');
    setMessage('');
  }

  async function save() {
    if (isSaving || invalid || !isChanged) return;
    edited();
    setIsSaving(true);
    try {
      const rsp = await api.setStunServer(value);
      if (rsp.code !== 0) {
        setError(
          rsp.msg
            ? `${t('settings.network.stun.saveFailed')}: ${rsp.msg}`
            : t('settings.network.stun.saveFailed')
        );
        return;
      }
      show(rsp.data as StunServerState);
      setMessage(t('settings.network.stun.saved'));
    } catch (err) {
      setError(describeFailure(err, t('settings.network.stun.saveFailed')));
    } finally {
      setIsSaving(false);
    }
  }

  return (
    <Section
      title={t('settings.network.stun.title')}
      description={t('settings.network.stun.description')}
      action={
        <label className="flex shrink-0 items-center gap-2 text-xs text-neutral-400">
          <span>{t('settings.network.stun.enable')}</span>
          <Switch
            checked={enabled}
            disabled={!state || isSaving}
            onChange={(checked) => {
              setEnabled(checked);
              edited();
            }}
          />
        </label>
      }
    >
      <label className="flex flex-col space-y-1">
        <span className="text-xs text-neutral-400">{t('settings.network.stun.server')}</span>
        <div className="flex flex-wrap gap-2">
          <Input
            value={server}
            status={invalid ? 'error' : undefined}
            placeholder={state?.default}
            className="max-w-80 font-mono"
            disabled={!state || !enabled || isSaving}
            spellCheck={false}
            autoComplete="off"
            onChange={(e) => {
              setServer(e.target.value);
              edited();
            }}
          />
          {state && enabled && server.trim() !== state.default && (
            <Button
              disabled={isSaving}
              onClick={() => {
                setServer(state.default);
                edited();
              }}
            >
              {t('settings.network.stun.useDefault')}
            </Button>
          )}
        </div>
        <span className="text-xs text-neutral-500">
          {t('settings.network.stun.serverHint', { server: state?.default ?? '' })}
        </span>
      </label>

      {invalid && (
        <span className="text-xs text-red-400">{t('settings.network.stun.invalid')}</span>
      )}

      <div className="flex flex-wrap gap-2">
        <Button
          type={isChanged ? 'primary' : 'default'}
          loading={isSaving}
          disabled={!state || !isChanged || invalid}
          onClick={save}
        >
          {t('settings.network.stun.apply')}
        </Button>
      </div>

      {error && <span className="text-xs text-red-400">{error}</span>}
      {message && <span className="text-xs text-green-500">{message}</span>}
    </Section>
  );
};
