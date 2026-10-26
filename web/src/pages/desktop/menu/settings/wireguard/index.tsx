import { useEffect, useRef, useState } from 'react';
import type { ReactNode } from 'react';
import { Alert, Button, Checkbox, Divider, Input, InputNumber, message, Modal, Switch } from 'antd';
import { PlusIcon, Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/wireguard.ts';
import type { WireguardState } from '@/api/wireguard.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { CopyBlock, CopyRow } from '../components/copy-button.tsx';
import { Box, Section, SectionHeader } from '../components/section.tsx';
import { StaleNote, StateTag } from '../components/status-tag.tsx';
import { usePoll } from '../components/use-poll.ts';
import { formatBytes } from '../vpn/format.ts';
import {
  draftFromSettings,
  emptyPeer,
  isDefaultRoute,
  peerSnippet,
  requestFromDraft,
  splitList,
  tunnelState
} from './model.ts';
import type { Draft, PeerDraft } from './model.ts';

// The handshake and the byte counts change all the time.
const statusPollMs = 5 * 1000;
// A wg-quick file is a few hundred bytes; the server refuses 64 KiB.
const maxFileSize = 64 * 1024;

type Rsp = { code: number; msg: string; data?: WireguardState };

// Wireguard is the built-in tunnel: no add-on, the kernel's own WireGuard,
// configured by the server.
export const Wireguard = () => {
  const { t } = useTranslation();
  const [modal, contextHolder] = Modal.useModal();
  const [state, setState] = useState<WireguardState>();
  const [draft, setDraft] = useState<Draft>();
  const [warnings, setWarnings] = useState<string[]>([]);
  const [busy, setBusy] = useState('');
  const [isStale, setIsStale] = useState(false);
  const [importOpen, setImportOpen] = useState(false);
  const [importText, setImportText] = useState('');
  const fileInput = useRef<HTMLInputElement>(null);

  // When the state was read, which the handshake's age is measured from.
  const [readAt, setReadAt] = useState(0);

  function apply(data: WireguardState, resetDraft: boolean) {
    setState(data);
    setReadAt(Date.now() / 1000);
    setIsStale(false);
    if (resetDraft) setDraft(draftFromSettings(data.settings));
  }

  function load(resetDraft: boolean) {
    api
      .getState()
      .then((rsp: Rsp) => {
        if (rsp.code !== 0 || !rsp.data) {
          setIsStale(true);
          return;
        }
        apply(rsp.data, resetDraft);
      })
      .catch(() => setIsStale(true));
  }

  useEffect(() => load(true), []); // eslint-disable-line react-hooks/exhaustive-deps
  usePoll(() => load(false), statusPollMs);

  // run makes a change. The answer carries the new state; a refused change
  // shows the server's reason.
  function run(name: string, request: () => Promise<Rsp>, resetDraft = true, done?: string) {
    if (busy) return Promise.resolve(false);
    setBusy(name);
    return request()
      .then((rsp) => {
        if (rsp.code !== 0 || !rsp.data) {
          message.error(describeFailure(rsp, t('settings.wireguard.failed')));
          load(false);
          return false;
        }
        apply(rsp.data, resetDraft);
        setWarnings(rsp.data.warnings ?? []);
        if (done) message.success(done);
        return true;
      })
      .catch((err) => {
        message.error(describeFailure(err, t('settings.wireguard.failed')));
        return false;
      })
      .finally(() => setBusy(''));
  }

  function setUp(on: boolean) {
    run(on ? 'up' : 'down', on ? api.up : api.down, false);
  }

  function regenerate() {
    if (!state?.settings.publicKey) {
      run('key', api.regenerateKey, false);
      return;
    }
    modal.confirm({
      title: t('settings.wireguard.newKeyConfirmTitle'),
      content: (
        <span className="text-sm text-neutral-400">
          {t('settings.wireguard.newKeyConfirmDesc')}
        </span>
      ),
      okText: t('settings.wireguard.newKey'),
      cancelText: t('common.cancel'),
      onOk: () => run('key', api.regenerateKey, false)
    });
  }

  function removeAll() {
    modal.confirm({
      title: t('settings.wireguard.removeConfirmTitle'),
      content: (
        <span className="text-sm text-neutral-400">
          {t('settings.wireguard.removeConfirmDesc')}
        </span>
      ),
      okText: t('common.remove'),
      okButtonProps: { danger: true },
      cancelText: t('common.cancel'),
      onOk: () => run('remove', api.remove)
    });
  }

  function readFile(file: File | undefined) {
    if (!file) return;
    if (file.size > maxFileSize) {
      message.error(t('settings.wireguard.fileTooLarge'));
      return;
    }
    file.text().then(setImportText, () => message.error(t('settings.wireguard.readFailed')));
  }

  function doImport() {
    run('import', () => api.importConfig(importText), true, t('settings.wireguard.imported')).then(
      (ok) => {
        if (!ok) return;
        setImportOpen(false);
        setImportText('');
      }
    );
  }

  function save() {
    if (!draft) return;
    run('save', () => api.save(requestFromDraft(draft)), true, t('settings.wireguard.saved'));
  }

  function updatePeer(index: number, change: Partial<PeerDraft>) {
    setDraft((current) =>
      current
        ? {
            ...current,
            peers: current.peers.map((peer, i) => (i === index ? { ...peer, ...change } : peer))
          }
        : current
    );
  }

  const settings = state?.settings;
  const status = state?.status;
  const tunnel = tunnelState(status, readAt);
  const draftDefaultRoute = !!draft?.peers.some((peer) =>
    isDefaultRoute(splitList(peer.allowedIPs))
  );
  const snippet = settings ? peerSnippet(settings.publicKey, settings.addresses) : '';
  const serverNotes = [...warnings, ...(status?.warnings ?? [])].filter(
    (note, i, all) => all.indexOf(note) === i
  );

  return (
    <>
      {contextHolder}
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.wireguard.title')}</span>
        {state && (
          <StateTag
            color={tunnel === 'connected' ? 'green' : tunnel === 'waiting' ? 'gold' : 'default'}
            label={
              tunnel === 'connected'
                ? t('settings.wireguard.stateConnected')
                : tunnel === 'waiting'
                  ? t('settings.wireguard.stateWaiting')
                  : t('common.off')
            }
            stale={isStale}
          />
        )}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <span className="text-xs text-neutral-500">{t('settings.wireguard.description')}</span>
        {isStale && <StaleNote />}

        <SectionHeader
          title={t('settings.wireguard.tunnel')}
          description={
            settings?.configured
              ? t('settings.wireguard.tunnelDesc')
              : t('settings.wireguard.notConfigured')
          }
          action={
            <Switch
              checked={!!status?.up}
              loading={!state || busy === 'up' || busy === 'down'}
              disabled={!settings?.configured && !status?.up}
              onChange={setUp}
            />
          }
        />
        <SectionHeader
          title={t('settings.wireguard.startAtBoot')}
          description={t('settings.wireguard.startAtBootDesc')}
          action={
            <Switch
              checked={!!settings?.startAtBoot}
              loading={!state || busy === 'boot'}
              disabled={!settings?.configured}
              onChange={(on) => run('boot', () => api.setBoot(on), false)}
            />
          }
        />

        {(settings?.defaultRoute || draftDefaultRoute) && (
          <Alert type="warning" showIcon message={t('settings.wireguard.defaultRouteWarning')} />
        )}
        {status?.error && <Alert type="error" showIcon message={status.error} />}
        {serverNotes.length > 0 && (
          <Alert
            type="info"
            showIcon
            message={
              <div className="flex flex-col">
                {serverNotes.map((note) => (
                  <span key={note}>{note}</span>
                ))}
              </div>
            }
          />
        )}

        {status?.up && status.peers.length > 0 && (
          <Section title={t('settings.wireguard.status')}>
            {status.peers.map((peer) => (
              <Box key={peer.publicKey}>
                <CopyRow label={t('settings.wireguard.publicKey')} value={peer.publicKey} />
                <CopyRow label={t('settings.wireguard.endpoint')} value={peer.endpoint ?? ''} />
                <div className="flex justify-between">
                  <span className="text-neutral-400">
                    {t('settings.wireguard.latestHandshake')}
                  </span>
                  <span className="text-neutral-300">
                    {peer.latestHandshake > 0
                      ? new Date(peer.latestHandshake * 1000).toLocaleString()
                      : t('settings.wireguard.never')}
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-neutral-400">{t('settings.wireguard.transfer')}</span>
                  <span className="text-neutral-300">
                    {t('settings.wireguard.transferValue', {
                      rx: formatBytes(peer.rxBytes),
                      tx: formatBytes(peer.txBytes)
                    })}
                  </span>
                </div>
              </Box>
            ))}
          </Section>
        )}

        <Section
          title={t('settings.wireguard.boardKey')}
          description={t('settings.wireguard.boardKeyDesc')}
          action={
            <Button size="small" loading={busy === 'key'} onClick={regenerate}>
              {settings?.publicKey
                ? t('settings.wireguard.newKey')
                : t('settings.wireguard.generateKey')}
            </Button>
          }
        >
          {settings?.publicKey ? (
            <>
              <Box>
                <CopyRow label={t('settings.wireguard.publicKey')} value={settings.publicKey} />
              </Box>
              {snippet && <CopyBlock title={t('settings.wireguard.peerSide')} text={snippet} />}
            </>
          ) : (
            <span className="text-xs text-neutral-500">{t('settings.wireguard.noKey')}</span>
          )}
        </Section>

        <Section
          title={t('settings.wireguard.configuration')}
          action={
            <Button size="small" onClick={() => setImportOpen(true)}>
              {t('settings.wireguard.import')}
            </Button>
          }
        >
          {draft && (
            <div className="flex flex-col space-y-3">
              <Field
                label={t('settings.wireguard.addresses')}
                hint={t('settings.wireguard.addressesDesc')}
              >
                <Input
                  value={draft.addresses}
                  placeholder="10.8.0.2/24"
                  onChange={(e) => setDraft({ ...draft, addresses: e.target.value })}
                />
              </Field>
              <Field label={t('settings.wireguard.dns')} hint={t('settings.wireguard.dnsDesc')}>
                <Input
                  value={draft.dns}
                  onChange={(e) => setDraft({ ...draft, dns: e.target.value })}
                />
              </Field>
              <div className="flex flex-wrap gap-4">
                <Field
                  label={t('settings.wireguard.listenPort')}
                  hint={t('settings.wireguard.listenPortDesc')}
                >
                  <InputNumber
                    min={0}
                    max={65535}
                    precision={0}
                    value={draft.listenPort}
                    onChange={(value) => setDraft({ ...draft, listenPort: value ?? 0 })}
                  />
                </Field>
                <Field label={t('settings.wireguard.mtu')} hint={t('settings.wireguard.mtuDesc')}>
                  <InputNumber
                    min={0}
                    max={1500}
                    precision={0}
                    value={draft.mtu}
                    onChange={(value) => setDraft({ ...draft, mtu: value ?? 0 })}
                  />
                </Field>
              </div>
              <Field
                label={t('settings.wireguard.privateKey')}
                hint={t('settings.wireguard.privateKeyDesc')}
              >
                <Input.Password
                  autoComplete="off"
                  value={draft.privateKey}
                  placeholder={t('settings.wireguard.privateKeyPlaceholder')}
                  onChange={(e) => setDraft({ ...draft, privateKey: e.target.value })}
                />
              </Field>

              {draft.peers.map((peer, index) => (
                <Box key={index}>
                  <div className="flex items-center justify-between">
                    <span className="font-medium">
                      {t('settings.wireguard.peer', { n: index + 1 })}
                    </span>
                    <Button
                      type="text"
                      size="small"
                      danger
                      icon={<Trash2Icon size={14} />}
                      title={t('common.remove')}
                      onClick={() =>
                        setDraft({ ...draft, peers: draft.peers.filter((_, i) => i !== index) })
                      }
                    />
                  </div>
                  <Field label={t('settings.wireguard.publicKey')}>
                    <Input
                      value={peer.publicKey}
                      onChange={(e) => updatePeer(index, { publicKey: e.target.value })}
                    />
                  </Field>
                  <Field label={t('settings.wireguard.presharedKey')}>
                    <Input.Password
                      autoComplete="off"
                      value={peer.presharedKey}
                      disabled={peer.clearPresharedKey}
                      placeholder={
                        peer.hasPresharedKey
                          ? t('settings.wireguard.presharedKeySet')
                          : t('settings.wireguard.presharedKeyNone')
                      }
                      onChange={(e) => updatePeer(index, { presharedKey: e.target.value })}
                    />
                  </Field>
                  {peer.hasPresharedKey && (
                    <Checkbox
                      checked={peer.clearPresharedKey}
                      onChange={(e) => updatePeer(index, { clearPresharedKey: e.target.checked })}
                    >
                      {t('settings.wireguard.clearPresharedKey')}
                    </Checkbox>
                  )}
                  <Field label={t('settings.wireguard.endpoint')}>
                    <Input
                      value={peer.endpoint}
                      placeholder={t('settings.wireguard.endpointPlaceholder')}
                      onChange={(e) => updatePeer(index, { endpoint: e.target.value })}
                    />
                  </Field>
                  <Field
                    label={t('settings.wireguard.allowedIPs')}
                    hint={t('settings.wireguard.allowedIPsDesc')}
                  >
                    <Input
                      value={peer.allowedIPs}
                      placeholder="10.8.0.0/24"
                      onChange={(e) => updatePeer(index, { allowedIPs: e.target.value })}
                    />
                  </Field>
                  <Field
                    label={t('settings.wireguard.keepalive')}
                    hint={t('settings.wireguard.keepaliveDesc')}
                  >
                    <InputNumber
                      min={0}
                      max={65535}
                      precision={0}
                      value={peer.persistentKeepalive}
                      onChange={(value) => updatePeer(index, { persistentKeepalive: value ?? 0 })}
                    />
                  </Field>
                </Box>
              ))}

              <div className="flex flex-wrap justify-between gap-2">
                <Button
                  icon={<PlusIcon size={14} />}
                  onClick={() => setDraft({ ...draft, peers: [...draft.peers, emptyPeer()] })}
                >
                  {t('settings.wireguard.addPeer')}
                </Button>
                <div className="flex gap-2">
                  {(settings?.publicKey || settings?.configured) && (
                    <Button danger loading={busy === 'remove'} onClick={removeAll}>
                      {t('settings.wireguard.remove')}
                    </Button>
                  )}
                  <Button type="primary" loading={busy === 'save'} onClick={save}>
                    {t('common.save')}
                  </Button>
                </div>
              </div>
            </div>
          )}
        </Section>
      </div>

      <Modal
        open={importOpen}
        title={t('settings.wireguard.importTitle')}
        okText={t('settings.wireguard.importBtn')}
        cancelText={t('common.cancel')}
        okButtonProps={{ disabled: !importText.trim(), loading: busy === 'import' }}
        onOk={doImport}
        onCancel={() => setImportOpen(false)}
      >
        <div className="flex flex-col space-y-3">
          <span className="text-xs text-neutral-500">{t('settings.wireguard.importDesc')}</span>
          <Input.TextArea
            rows={10}
            spellCheck={false}
            className="font-mono"
            value={importText}
            placeholder={
              '[Interface]\nPrivateKey = ...\nAddress = 10.8.0.2/24\n\n[Peer]\nPublicKey = ...'
            }
            onChange={(e) => setImportText(e.target.value)}
          />
          <div>
            <Button size="small" onClick={() => fileInput.current?.click()}>
              {t('settings.wireguard.chooseFile')}
            </Button>
            <input
              ref={fileInput}
              type="file"
              accept=".conf,text/plain"
              className="hidden"
              onChange={(event) => {
                readFile(event.target.files?.[0]);
                event.target.value = '';
              }}
            />
          </div>
        </div>
      </Modal>
    </>
  );
};

type FieldProps = {
  label: string;
  hint?: string;
  children: ReactNode;
};

const Field = ({ label, hint, children }: FieldProps) => (
  <label className="flex flex-col space-y-1">
    <span className="text-sm text-neutral-400">{label}</span>
    {children}
    {hint && <span className="text-xs text-neutral-500">{hint}</span>}
  </label>
);
