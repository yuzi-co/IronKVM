import { useEffect, useState } from 'react';
import { Button, Input, Segmented } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/network.ts';
import type { IPv6Mode, IPv6State } from '@/api/network.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { Box, Section } from '../components/section.tsx';
import { isValidIPv6 } from './ip.ts';

const DEFAULT_PREFIX = 64;

const scopeKeys = {
  link: 'settings.network.ipv6.scopeLink',
  ula: 'settings.network.ipv6.scopeUla',
  global: 'settings.network.ipv6.scopeGlobal'
} as const;

// Ipv6 sets how the wired interface gets its IPv6 addresses. The IPv4 address
// is not touched by anything on this page. The device saves a setting only
// after it applied, and puts the previous one back when applying fails. The
// page shows only on a server with the ipv6-settings feature.
export const IPv6 = () => {
  const { t } = useTranslation();

  const [state, setState] = useState<IPv6State | null>(null);
  const [mode, setMode] = useState<IPv6Mode>('slaac');
  const [address, setAddress] = useState('');
  const [prefix, setPrefix] = useState(String(DEFAULT_PREFIX));
  const [gateway, setGateway] = useState('');
  const [isSaving, setIsSaving] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  function load() {
    api
      .getIPv6()
      .then((rsp) => {
        if (rsp.code !== 0) {
          setError(rsp.msg || t('settings.network.ipv6.loadFailed'));
          return;
        }
        const data = rsp.data as IPv6State;
        setState(data);
        setMode(data.mode);
        setAddress(data.address);
        setPrefix(String(data.prefix || DEFAULT_PREFIX));
        setGateway(data.gateway);
      })
      .catch((err) => setError(describeFailure(err, t('settings.network.ipv6.loadFailed'))));
  }

  useEffect(load, []); // eslint-disable-line react-hooks/exhaustive-deps

  const prefixNumber = /^\d+$/.test(prefix.trim()) ? Number(prefix.trim()) : NaN;
  const addressBad = mode === 'static' && !isValidIPv6(address.trim());
  const prefixBad = mode === 'static' && !(prefixNumber >= 16 && prefixNumber <= 126);
  const gatewayBad = mode === 'static' && gateway.trim() !== '' && !isValidIPv6(gateway.trim());
  const invalid = addressBad || prefixBad || gatewayBad;

  const isChanged =
    !!state &&
    (mode !== state.mode ||
      (mode === 'static' &&
        (address.trim() !== state.address ||
          prefixNumber !== state.prefix ||
          gateway.trim() !== state.gateway)));

  // A page open over IPv6 may only add DHCPv6 to SLAAC: the device refuses
  // the rest, since it would cut this connection. Say so before asking.
  const overIPv6 = !!state && state.client.includes(':');
  const blockedByIPv6 = overIPv6 && isChanged && !(state?.mode === 'slaac' && mode === 'dhcpv6');

  async function save() {
    if (isSaving || invalid || !isChanged) return;
    setError('');
    setMessage('');
    setIsSaving(true);
    try {
      const rsp = await api.setIPv6(
        mode,
        mode === 'static' ? address.trim() : '',
        mode === 'static' ? prefixNumber : 0,
        mode === 'static' ? gateway.trim() : ''
      );
      if (rsp.code !== 0) {
        setError(
          rsp.msg
            ? `${t('settings.network.ipv6.saveFailed')}: ${rsp.msg}`
            : t('settings.network.ipv6.saveFailed')
        );
        return;
      }
      setMessage(t('settings.network.ipv6.saved'));
      load();
    } catch (err) {
      setError(describeFailure(err, t('settings.network.ipv6.saveFailed')));
    } finally {
      setIsSaving(false);
    }
  }

  const modes: { value: IPv6Mode; label: string; disabled?: boolean }[] = [
    { value: 'slaac', label: t('settings.network.ipv6.slaac') },
    {
      value: 'dhcpv6',
      label: t('settings.network.ipv6.dhcpv6'),
      disabled: !state?.dhcpv6Available
    },
    { value: 'static', label: t('settings.network.ipv6.static') },
    { value: 'off', label: t('settings.network.ipv6.off') }
  ];

  return (
    <Section
      title={t('settings.network.ipv6.title')}
      description={t('settings.network.ipv6.description')}
    >
      {state && !state.supported ? (
        <span className="text-xs text-yellow-400/80">{t('settings.network.ipv6.unsupported')}</span>
      ) : (
        <>
          <Segmented
            value={mode}
            options={modes}
            disabled={!state || isSaving}
            onChange={(value) => {
              setMode(value);
              setError('');
              setMessage('');
            }}
          />

          {state && !state.dhcpv6Available && (
            <span className="text-xs text-neutral-500">{t('settings.network.ipv6.noDhcpv6')}</span>
          )}

          {mode === 'static' && (
            <div className="flex flex-col space-y-2">
              <label className="flex flex-col space-y-1">
                <span className="text-xs text-neutral-400">
                  {t('settings.network.ipv6.address')}
                </span>
                <Input
                  value={address}
                  status={addressBad && address !== '' ? 'error' : undefined}
                  placeholder="2001:db8::50"
                  className="font-mono"
                  spellCheck={false}
                  onChange={(e) => setAddress(e.target.value)}
                />
              </label>
              <label className="flex flex-col space-y-1">
                <span className="text-xs text-neutral-400">
                  {t('settings.network.ipv6.prefix')}
                </span>
                <Input
                  value={prefix}
                  status={prefixBad ? 'error' : undefined}
                  placeholder="64"
                  className="w-24 font-mono"
                  onChange={(e) => setPrefix(e.target.value)}
                />
              </label>
              <label className="flex flex-col space-y-1">
                <span className="text-xs text-neutral-400">
                  {t('settings.network.ipv6.router')}
                </span>
                <Input
                  value={gateway}
                  status={gatewayBad ? 'error' : undefined}
                  placeholder="fe80::1"
                  className="font-mono"
                  spellCheck={false}
                  onChange={(e) => setGateway(e.target.value)}
                />
              </label>
              {(addressBad || prefixBad || gatewayBad) && address !== '' && (
                <span className="text-xs text-red-400">{t('settings.network.ipv6.invalid')}</span>
              )}
            </div>
          )}

          {mode === 'off' && (
            <span className="text-xs text-neutral-500">{t('settings.network.ipv6.offNote')}</span>
          )}

          <span className="text-xs text-neutral-500">{t('settings.network.ipv6.keepsIpv4')}</span>

          {blockedByIPv6 && (
            <span className="text-xs text-yellow-400/80">{t('settings.network.ipv6.useIpv4')}</span>
          )}

          <div className="flex flex-wrap gap-2">
            <Button
              type={isChanged ? 'primary' : 'default'}
              loading={isSaving}
              disabled={!state || !isChanged || invalid || blockedByIPv6}
              onClick={save}
            >
              {t('settings.network.ipv6.apply')}
            </Button>
          </div>

          {error && <span className="text-xs text-red-400">{error}</span>}
          {message && <span className="text-xs text-green-500">{message}</span>}

          {state && (
            <Box>
              <div className="flex justify-between gap-3">
                <span className="text-neutral-300">{t('settings.network.ipv6.current')}</span>
                {state.live.disabled && (
                  <span className="text-neutral-500">{t('settings.network.ipv6.disabled')}</span>
                )}
              </div>
              {state.live.addresses.length === 0 && !state.live.disabled && (
                <span className="text-neutral-500">{t('settings.network.ipv6.noAddress')}</span>
              )}
              {state.live.addresses.map((a) => (
                <div key={a.address} className="flex justify-between gap-3">
                  <span className="font-mono break-all text-neutral-300">
                    {a.address}/{a.prefix}
                  </span>
                  <span className="shrink-0 text-neutral-500">{t(scopeKeys[a.scope])}</span>
                </div>
              ))}
              {state.live.gateway && (
                <div className="flex justify-between gap-3">
                  <span className="text-neutral-300">{t('settings.network.ipv6.router')}</span>
                  <span className="font-mono text-neutral-500">{state.live.gateway}</span>
                </div>
              )}
            </Box>
          )}
        </>
      )}
    </Section>
  );
};
