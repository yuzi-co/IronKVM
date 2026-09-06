import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { Button, Input, Modal, Segmented } from 'antd';
import { CheckIcon, ExternalLinkIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/network.ts';
import type { EthernetMode } from '@/api/network.ts';

type EthernetLive = {
  interface?: string;
  address?: string;
  prefix?: number;
  netmask?: string;
  gateway?: string;
};

type EthernetTrial = {
  token: string;
  mode: EthernetMode;
  address: string;
  prefix: number;
  gateway: string;
  remainingSeconds: number;
};

type EthernetState = {
  mode: EthernetMode;
  address: string;
  prefix: number;
  gateway: string;
  live: EthernetLive;
  trial?: EthernetTrial;
};

function isValidIPv4(value: string) {
  const parts = value.split('.');
  if (parts.length !== 4) return false;

  return parts.every((part) => {
    if (!/^\d+$/.test(part)) return false;
    if (part.length > 1 && part.startsWith('0')) return false;

    const number = Number(part);
    return number >= 0 && number <= 255;
  });
}

// The field accepts either form, because a subnet is written as a dotted mask
// on one platform and as a prefix length on the next, and a person reading it
// off another machine should not have to convert it first.
function parsePrefix(value: string): number | null {
  const trimmed = value.trim().replace(/^\//, '');
  if (trimmed === '') return null;

  if (/^\d+$/.test(trimmed)) {
    const prefix = Number(trimmed);
    return prefix >= 1 && prefix <= 30 ? prefix : null;
  }

  if (!isValidIPv4(trimmed)) return null;

  // A mask is valid only when it is a run of ones followed by a run of zeros.
  const bits = trimmed
    .split('.')
    .map((part) => Number(part).toString(2).padStart(8, '0'))
    .join('');
  if (!/^1*0*$/.test(bits)) return null;

  const prefix = bits.indexOf('0') === -1 ? 32 : bits.indexOf('0');
  return prefix >= 1 && prefix <= 30 ? prefix : null;
}

function maskOf(prefix: number) {
  const bytes = [0, 0, 0, 0].map((_, index) => {
    const bits = Math.min(Math.max(prefix - index * 8, 0), 8);
    return 256 - Math.pow(2, 8 - bits);
  });

  return bytes.join('.');
}

const Panel = ({
  title,
  description,
  children
}: {
  title: string;
  description?: string;
  children: ReactNode;
}) => {
  return (
    <div className="overflow-hidden rounded-xl bg-neutral-800/50">
      <div className="px-4 pt-3 pb-1.5">
        <div className="font-semibold text-neutral-100">{title}</div>
        {description && (
          <div className="mt-0.5 text-xs leading-snug text-neutral-500">{description}</div>
        )}
      </div>
      <div>{children}</div>
    </div>
  );
};

const Row = ({
  label,
  children,
  isLast = false
}: {
  label: string;
  children: ReactNode;
  isLast?: boolean;
}) => {
  return (
    <div className="px-4">
      <div
        className={`flex min-h-[44px] items-center justify-between gap-3 ${
          isLast ? '' : 'border-b border-neutral-700/50'
        }`}
      >
        <span className="shrink-0 text-sm text-neutral-300">{label}</span>
        {children}
      </div>
    </div>
  );
};

const ReadOnlyValue = ({ value }: { value?: string }) => (
  <span className="max-w-[330px] text-right text-sm break-all text-neutral-500">
    {value || '-'}
  </span>
);

export const Ethernet = () => {
  const { t } = useTranslation();

  const [mode, setMode] = useState<EthernetMode>('dhcp');
  const [savedMode, setSavedMode] = useState<EthernetMode>('dhcp');
  const [address, setAddress] = useState('');
  const [mask, setMask] = useState('');
  const [gateway, setGateway] = useState('');
  const [saved, setSaved] = useState({ address: '', mask: '', gateway: '' });
  const [live, setLive] = useState<EthernetLive>({});

  const [trial, setTrial] = useState<EthernetTrial | null>(null);
  const [remaining, setRemaining] = useState(0);

  const [isLoading, setIsLoading] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [isAsking, setIsAsking] = useState(false);
  const [message, setMessage] = useState('');
  const [error, setError] = useState('');

  const countdown = useRef<ReturnType<typeof setInterval>>(undefined);

  useEffect(() => {
    getEthernet();

    return () => clearInterval(countdown.current);
  }, []);

  // The device counts down on its own. This only keeps the number on screen
  // moving between polls, so a stale page cannot claim there is more time left
  // than there is.
  useEffect(() => {
    clearInterval(countdown.current);
    if (!trial) return;

    countdown.current = setInterval(() => {
      setRemaining((value) => {
        if (value <= 1) {
          clearInterval(countdown.current);
          setTrial(null);
          getEthernet(false);
          return 0;
        }

        return value - 1;
      });
    }, 1000);

    return () => clearInterval(countdown.current);
  }, [trial]);

  async function getEthernet(showLoading = true) {
    if (showLoading) setIsLoading(true);

    try {
      const rsp = await api.getEthernet();
      if (rsp.code !== 0) {
        setError(rsp.msg);
        return;
      }

      const data = rsp.data as EthernetState;

      setSavedMode(data.mode);
      setMode(data.mode);
      setAddress(data.address || '');
      setMask(data.prefix ? maskOf(data.prefix) : '');
      setGateway(data.gateway || '');
      setSaved({
        address: data.address || '',
        mask: data.prefix ? maskOf(data.prefix) : '',
        gateway: data.gateway || ''
      });
      setLive(data.live || {});

      setTrial(data.trial ?? null);
      setRemaining(data.trial?.remainingSeconds ?? 0);
    } catch (err) {
      console.log(err);
    } finally {
      if (showLoading) setIsLoading(false);
    }
  }

  function validate() {
    if (mode === 'dhcp') return '';

    if (!address.trim()) return t('settings.network.ethernet.addressRequired');
    if (!isValidIPv4(address.trim())) return t('settings.network.ethernet.invalidAddress');
    if (!mask.trim()) return t('settings.network.ethernet.maskRequired');
    if (parsePrefix(mask) === null) return t('settings.network.ethernet.invalidMask');
    if (gateway.trim() && !isValidIPv4(gateway.trim()))
      return t('settings.network.ethernet.invalidRouter');

    return '';
  }

  async function apply() {
    setIsAsking(false);
    setMessage('');
    setError('');
    setIsSaving(true);

    try {
      const prefix = mode === 'static' ? (parsePrefix(mask) ?? 0) : 0;
      const rsp = await api.setEthernet(mode, address.trim(), prefix, gateway.trim());

      if (rsp.code !== 0) {
        setError(rsp.msg || t('settings.network.ethernet.applyFailed'));
        return;
      }

      // No token means the device found nothing to change and left the
      // interface alone.
      if (!rsp.data?.token) {
        await getEthernet(false);
        return;
      }

      setTrial({
        token: rsp.data.token,
        mode,
        address: rsp.data.address || address.trim(),
        prefix,
        gateway: gateway.trim(),
        remainingSeconds: rsp.data.trialSeconds
      });
      setRemaining(rsp.data.trialSeconds);
    } catch (err) {
      // The interface goes down while this request is in flight, so a network
      // error here is the expected outcome rather than a failure. The trial is
      // running on the device either way, and the page below says what to do.
      console.log(err);
      setTrial({
        token: '',
        mode,
        address: address.trim(),
        prefix: parsePrefix(mask) ?? 0,
        gateway: gateway.trim(),
        remainingSeconds: 0
      });
    } finally {
      setIsSaving(false);
    }
  }

  async function keep() {
    if (!trial?.token) return;

    setMessage('');
    setError('');

    try {
      const rsp = await api.confirmEthernet(trial.token);
      if (rsp.code !== 0) {
        setError(rsp.msg || t('settings.network.ethernet.trialKeepFailed'));
        return;
      }

      setTrial(null);
      setMessage(t('settings.network.ethernet.trialKept'));
      await getEthernet(false);
    } catch (err) {
      console.log(err);
      setError(t('settings.network.ethernet.trialKeepFailed'));
    }
  }

  const invalid = validate();
  const hasChanges =
    mode !== savedMode ||
    (mode === 'static' &&
      (address.trim() !== saved.address ||
        parsePrefix(mask) !== parsePrefix(saved.mask) ||
        gateway.trim() !== saved.gateway));

  const newAddressUrl = trial?.address
    ? `${window.location.protocol}//${trial.address}${window.location.port ? `:${window.location.port}` : ''}/`
    : '';

  const statusText = error || message || (hasChanges ? t('settings.network.ethernet.unsaved') : '');
  const statusColor = error ? 'text-red-400' : message ? 'text-green-400' : 'text-yellow-400/80';

  return (
    <div className="flex flex-col space-y-5">
      <div className="flex items-center justify-between">
        <div className="flex flex-col space-y-1">
          <span>{t('settings.network.ethernet.title')}</span>
          <span className="text-xs text-neutral-500">
            {t('settings.network.ethernet.description')}
          </span>
        </div>

        <Segmented
          disabled={isLoading || isSaving || !!trial}
          value={mode}
          onChange={(value) => {
            setMode(value as EthernetMode);
            setMessage('');
            setError('');
          }}
          options={[
            { label: t('settings.network.ethernet.dhcp'), value: 'dhcp' },
            { label: t('settings.network.ethernet.manual'), value: 'static' }
          ]}
        />
      </div>

      <Panel title={t('settings.network.ethernet.networkDetails')}>
        <Row label={t('settings.network.ethernet.interface')}>
          <ReadOnlyValue value={live.interface} />
        </Row>

        {mode === 'static' ? (
          <>
            <Row label={t('settings.network.ethernet.ipAddress')}>
              <Input
                className="max-w-[220px]"
                value={address}
                disabled={!!trial}
                placeholder="0.0.0.0"
                onChange={(e) => {
                  setAddress(e.target.value);
                  setMessage('');
                  setError('');
                }}
                status={address.trim() && !isValidIPv4(address.trim()) ? 'error' : undefined}
              />
            </Row>
            <Row label={t('settings.network.ethernet.subnetMask')}>
              <Input
                className="max-w-[220px]"
                value={mask}
                disabled={!!trial}
                placeholder="255.255.255.0"
                onChange={(e) => {
                  setMask(e.target.value);
                  setMessage('');
                  setError('');
                }}
                status={mask.trim() && parsePrefix(mask) === null ? 'error' : undefined}
              />
            </Row>
            <Row label={t('settings.network.ethernet.router')} isLast>
              <Input
                className="max-w-[220px]"
                value={gateway}
                disabled={!!trial}
                placeholder="0.0.0.0"
                onChange={(e) => {
                  setGateway(e.target.value);
                  setMessage('');
                  setError('');
                }}
                status={gateway.trim() && !isValidIPv4(gateway.trim()) ? 'error' : undefined}
              />
            </Row>
          </>
        ) : (
          <>
            <Row label={t('settings.network.ethernet.ipAddress')}>
              <ReadOnlyValue value={live.address} />
            </Row>
            <Row label={t('settings.network.ethernet.subnetMask')}>
              <ReadOnlyValue value={live.netmask} />
            </Row>
            <Row label={t('settings.network.ethernet.router')} isLast>
              <ReadOnlyValue value={live.gateway} />
            </Row>
          </>
        )}
      </Panel>

      {trial && (
        <div className="rounded-xl border border-yellow-500/40 bg-yellow-500/5 px-4 py-3">
          <div className="font-semibold text-yellow-300">
            {t('settings.network.ethernet.trialTitle')}
          </div>
          <div className="mt-1 text-sm text-neutral-300">
            {trial.mode === 'dhcp'
              ? t('settings.network.ethernet.trialDhcp')
              : t('settings.network.ethernet.trialStatic', { address: trial.address })}
          </div>
          <div className="mt-1 text-xs text-neutral-400">
            {t('settings.network.ethernet.trialInstruction', { seconds: remaining })}
          </div>

          <div className="mt-3 flex items-center gap-2">
            {newAddressUrl && (
              <Button
                icon={<ExternalLinkIcon size={14} />}
                href={newAddressUrl}
                target="_blank"
                rel="noreferrer"
              >
                {t('settings.network.ethernet.trialOpen')}
              </Button>
            )}
            <Button type="primary" disabled={!trial.token} onClick={keep}>
              {t('settings.network.ethernet.trialKeep')}
            </Button>
          </div>
        </div>
      )}

      {(hasChanges || statusText) && !trial && (
        <div className="flex items-center justify-between">
          <span className={`text-xs ${statusColor}`}>{invalid || statusText}</span>

          <Button
            type={hasChanges ? 'primary' : 'default'}
            icon={message ? <CheckIcon size={14} /> : undefined}
            loading={isSaving}
            disabled={isLoading || !hasChanges || !!invalid}
            onClick={() => setIsAsking(true)}
          >
            {t('settings.network.ethernet.save')}
          </Button>
        </div>
      )}

      <Modal
        open={isAsking}
        title={t('settings.network.ethernet.applyTitle')}
        okText={t('settings.network.ethernet.applyConfirm')}
        cancelText={t('settings.network.ethernet.applyCancel')}
        onOk={apply}
        onCancel={() => setIsAsking(false)}
      >
        <p className="text-sm text-neutral-300">
          {t('settings.network.ethernet.applyWarning', { seconds: 180 })}
        </p>
      </Modal>
    </div>
  );
};
