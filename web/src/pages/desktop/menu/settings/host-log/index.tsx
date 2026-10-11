import { useCallback, useEffect, useRef, useState } from 'react';
import { Button, Divider, Input, InputNumber, message, Modal, Select, Switch, Tag } from 'antd';
import { DownloadIcon, PauseIcon, PlayIcon, Trash2Icon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/hostlog.ts';
import type {
  HostLogEntry,
  HostLogListenerState,
  HostLogSettings,
  HostLogState
} from '@/api/hostlog.ts';
import { describeFailure } from '@/lib/feedback.ts';
import {
  mergeEntries,
  parseFrame,
  parseList,
  retryDelay,
  severityColor,
  severityName
} from '@/lib/hostlog.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';

import { StaleNote, StatusTag } from '../components/status-tag.tsx';

const statePollMs = 10 * 1000;

// What the server enforces on a save.
const limits = { memoryKb: [64, 4096], diskKb: [256, 65536] } as const;

const stateColors: Record<HostLogListenerState['state'], string> = {
  off: 'default',
  listening: 'green',
  waiting: 'orange',
  error: 'red'
};

const severityOptions = [
  { value: -1, key: 'severityAll' },
  { value: 3, key: 'severityError' },
  { value: 4, key: 'severityWarning' },
  { value: 6, key: 'severityInfo' }
] as const;

function formatClock(value: string) {
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? value : date.toLocaleTimeString();
}

function sameSettings(a: HostLogSettings, b: HostLogSettings) {
  return JSON.stringify(a) === JSON.stringify(b);
}

// The host log: the two listeners and where they hear from, then the lines as
// they arrive, with search.
export const HostLog = () => {
  const { t } = useTranslation();
  const [modal, contextHolder] = Modal.useModal();
  const [settings, setSettings] = useState<HostLogSettings | null>(null);
  const [draft, setDraft] = useState<HostLogSettings | null>(null);
  // The two text boxes are edited as text and parsed on save.
  const [sourcesText, setSourcesText] = useState('');
  const [patternsText, setPatternsText] = useState('');
  const [state, setState] = useState<HostLogState | null>(null);
  const [isStale, setIsStale] = useState(false);
  const [isSaving, setIsSaving] = useState(false);

  const getSettings = useCallback(() => {
    api
      .getHostLogSettings()
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.hostLog.failed')));
          return;
        }
        const s = rsp.data as HostLogSettings;
        setSettings(s);
        setDraft(s);
        setSourcesText(s.sources.join('\n'));
        setPatternsText(s.patterns.join('\n'));
      })
      .catch((err) => message.error(describeFailure(err, t('settings.hostLog.failed'))));
  }, [t]);

  const getState = useCallback(() => {
    api
      .getHostLogState()
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

  const sources = parseList(sourcesText, true);
  const patterns = parseList(patternsText);
  const edited: HostLogSettings | null = draft && { ...draft, sources, patterns };
  const isDirty = !!edited && !!settings && !sameSettings(edited, settings);

  function update(change: Partial<HostLogSettings>) {
    setDraft((current) => (current ? { ...current, ...change } : current));
  }

  function save() {
    if (!edited) return;
    setIsSaving(true);
    api
      .setHostLogSettings(edited)
      .then((rsp) => {
        if (rsp.code !== 0) {
          message.error(describeFailure(rsp, t('settings.hostLog.failed')));
          return;
        }
        const s = rsp.data as HostLogSettings;
        setSettings(s);
        setDraft(s);
        setSourcesText(s.sources.join('\n'));
        setPatternsText(s.patterns.join('\n'));
        message.success(t('settings.hostLog.saved'));
        getState();
      })
      .catch((err) => message.error(describeFailure(err, t('settings.hostLog.failed'))))
      .finally(() => setIsSaving(false));
  }

  function listenerText(s: HostLogListenerState) {
    switch (s.state) {
      case 'listening':
        return t('settings.hostLog.stateListening', { addr: s.addr });
      case 'waiting':
        return t('settings.hostLog.stateWaiting');
      case 'error':
        return s.error ?? t('settings.hostLog.failed');
      default:
        return '';
    }
  }

  function listenerRow(kind: 'syslog' | 'netconsole') {
    if (!draft) return null;
    const listener = draft[kind];
    const live = state?.[kind];
    return (
      <div className="flex flex-col space-y-2 rounded-xl border border-neutral-700/50 bg-neutral-800/40 p-4">
        <div className="flex items-center justify-between">
          <div className="flex flex-col pr-4">
            <span className="text-sm">{t(`settings.hostLog.${kind}`)}</span>
            <span className="text-xs text-neutral-500">{t(`settings.hostLog.${kind}Desc`)}</span>
          </div>
          <Switch
            checked={listener.enabled}
            onChange={(enabled) => update({ [kind]: { ...listener, enabled } })}
          />
        </div>
        <div className="flex items-center justify-between gap-3">
          <span className="text-sm">{t('settings.hostLog.port')}</span>
          <InputNumber
            style={{ width: 110 }}
            min={1}
            max={65535}
            precision={0}
            value={listener.port}
            addonBefore="UDP"
            onChange={(port) => update({ [kind]: { ...listener, port: port ?? 1 } })}
          />
        </div>
        {live && (
          <div className="flex items-center gap-2">
            <Tag color={stateColors[live.state]} className="me-0">
              {t(`settings.hostLog.state.${live.state}`)}
            </Tag>
            <span className="truncate text-xs text-neutral-500">{listenerText(live)}</span>
          </div>
        )}
      </div>
    );
  }

  return (
    <>
      {contextHolder}
      <div className="flex items-center justify-between">
        <span className="text-base">{t('settings.hostLog.title')}</span>
        {draft && (
          <StatusTag
            running={!!settings && (settings.syslog.enabled || settings.netconsole.enabled)}
          />
        )}
      </div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <span className="text-xs text-neutral-500">{t('settings.hostLog.intro')}</span>

        {draft && (
          <div className="flex flex-col space-y-4">
            {listenerRow('syslog')}
            {listenerRow('netconsole')}

            <div className="flex items-center justify-between gap-3">
              <div className="flex flex-col pr-4">
                <span className="text-sm">{t('settings.hostLog.bind')}</span>
                <span className="text-xs text-neutral-500">{t('settings.hostLog.bindDesc')}</span>
              </div>
              <Select
                style={{ width: 200 }}
                value={draft.bind}
                options={[
                  { value: '', label: t('settings.hostLog.bindAll') },
                  { value: 'usb', label: t('settings.hostLog.bindUsb') }
                ]}
                onChange={(bind) => update({ bind })}
              />
            </div>

            <div className="flex flex-col space-y-1">
              <span className="text-sm">{t('settings.hostLog.sources')}</span>
              <span className="text-xs text-neutral-500">{t('settings.hostLog.sourcesDesc')}</span>
              <Input.TextArea
                rows={3}
                value={sourcesText}
                placeholder="172.31.255.2&#10;192.168.1.0/24"
                onChange={(e) => setSourcesText(e.target.value)}
              />
              {state && (
                <span className="text-xs text-neutral-500">
                  {state.sources.length > 0
                    ? t('settings.hostLog.inForce', { list: state.sources.join(', ') })
                    : t('settings.hostLog.inForceAny')}
                </span>
              )}
            </div>

            <div className="flex items-center justify-between gap-3">
              <div className="flex flex-col pr-4">
                <span className="text-sm">{t('settings.hostLog.memory')}</span>
                <span className="text-xs text-neutral-500">{t('settings.hostLog.memoryDesc')}</span>
              </div>
              <InputNumber
                style={{ width: 150 }}
                min={limits.memoryKb[0]}
                max={limits.memoryKb[1]}
                precision={0}
                value={draft.memoryKb}
                addonAfter="KB"
                onChange={(memoryKb) => update({ memoryKb: memoryKb ?? limits.memoryKb[0] })}
              />
            </div>

            <div className="flex items-center justify-between gap-3">
              <div className="flex flex-col pr-4">
                <span className="text-sm">{t('settings.hostLog.disk')}</span>
                <span className="text-xs text-neutral-500">{t('settings.hostLog.diskDesc')}</span>
              </div>
              <Switch checked={draft.disk} onChange={(disk) => update({ disk })} />
            </div>
            {draft.disk && (
              <div className="flex items-center justify-between gap-3">
                <span className="text-sm">{t('settings.hostLog.diskLimit')}</span>
                <InputNumber
                  style={{ width: 150 }}
                  min={limits.diskKb[0]}
                  max={limits.diskKb[1]}
                  precision={0}
                  value={draft.diskKb}
                  addonAfter="KB"
                  onChange={(diskKb) => update({ diskKb: diskKb ?? limits.diskKb[0] })}
                />
              </div>
            )}

            <div className="flex flex-col space-y-1">
              <span className="text-sm">{t('settings.hostLog.patterns')}</span>
              <span className="text-xs text-neutral-500">{t('settings.hostLog.patternsDesc')}</span>
              <Input.TextArea
                rows={4}
                value={patternsText}
                className="font-mono"
                onChange={(e) => setPatternsText(e.target.value)}
              />
            </div>

            <div className="flex justify-end">
              <Button type="primary" disabled={!isDirty} loading={isSaving} onClick={save}>
                {t('settings.hostLog.save')}
              </Button>
            </div>
          </div>
        )}

        {state && (
          <div className={isStale ? 'opacity-60' : ''}>
            {isStale && <StaleNote />}
            <span className="text-xs text-neutral-500">
              {t('settings.hostLog.stats', {
                received: state.received,
                filtered: state.filtered,
                dropped: state.rateLimited,
                count: state.buffered
              })}
            </span>
          </div>
        )}

        <LiveTail
          onClear={() =>
            modal.confirm({
              title: t('settings.hostLog.clearConfirmTitle'),
              content: (
                <span className="text-sm text-neutral-400">
                  {t('settings.hostLog.clearConfirmDesc')}
                </span>
              ),
              okText: t('settings.hostLog.clear'),
              okButtonProps: { danger: true },
              cancelText: t('settings.hostLog.cancel'),
              onOk: () =>
                api
                  .clearHostLog()
                  .then((rsp) => {
                    if (rsp.code !== 0) {
                      message.error(describeFailure(rsp, t('settings.hostLog.failed')));
                      return;
                    }
                    getState();
                    window.dispatchEvent(new Event('hostlog-cleared'));
                  })
                  .catch((err) => message.error(describeFailure(err, t('settings.hostLog.failed'))))
            })
          }
        />
      </div>
    </>
  );
};

type Connection = 'connecting' | 'live' | 'closed';

// The live tail: a socket that sends the newest lines, then each new one, for
// the search and level chosen. A change of either opens a new socket.
const LiveTail = ({ onClear }: { onClear: () => void }) => {
  const { t } = useTranslation();
  const [lines, setLines] = useState<HostLogEntry[]>([]);
  const [query, setQuery] = useState('');
  const [applied, setApplied] = useState('');
  const [severity, setSeverity] = useState(-1);
  const [paused, setPaused] = useState(false);
  const [connection, setConnection] = useState<Connection>('connecting');
  const boxRef = useRef<HTMLDivElement>(null);
  const followRef = useRef(true);

  // The search applies a moment after the last key, so typing does not open a
  // socket per letter.
  useEffect(() => {
    const timer = window.setTimeout(() => setApplied(query), 400);
    return () => window.clearTimeout(timer);
  }, [query]);

  const filter = { q: applied, severity: severity < 0 ? null : severity };

  useEffect(() => {
    if (paused) return;

    let closed = false;
    let socket: WebSocket | null = null;
    let timer: number | undefined;
    let attempt = 0;

    const open = () => {
      setConnection('connecting');
      socket = new WebSocket(
        api.hostLogStreamUrl({ q: applied, severity: severity < 0 ? null : severity })
      );
      socket.onopen = () => {
        attempt = 0;
        setConnection('live');
      };
      socket.onmessage = (event) => {
        const frame = parseFrame(event.data);
        if (!frame) return;
        // The newest lines for this search come first and replace what was
        // shown; every frame after that adds to it.
        setLines((current) =>
          frame.type === 'backlog' ? frame.entries : mergeEntries(current, frame.entries)
        );
      };
      socket.onclose = () => {
        if (closed) return;
        setConnection('closed');
        timer = window.setTimeout(open, retryDelay(attempt++));
      };
      socket.onerror = () => socket?.close();
    };
    open();

    const onCleared = () => setLines([]);
    window.addEventListener('hostlog-cleared', onCleared);

    return () => {
      closed = true;
      window.clearTimeout(timer);
      window.removeEventListener('hostlog-cleared', onCleared);
      socket?.close();
    };
  }, [applied, severity, paused]);

  // Follow the newest line while the view is at the bottom.
  useEffect(() => {
    const box = boxRef.current;
    if (box && followRef.current) box.scrollTop = box.scrollHeight;
  }, [lines]);

  function onScroll() {
    const box = boxRef.current;
    if (box) followRef.current = box.scrollHeight - box.scrollTop - box.clientHeight < 24;
  }

  return (
    <div className="flex flex-col space-y-2">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium">{t('settings.hostLog.live')}</span>
        <Tag
          className="me-0"
          color={paused ? 'default' : connection === 'live' ? 'green' : 'orange'}
        >
          {paused
            ? t('settings.hostLog.paused')
            : connection === 'live'
              ? t('settings.hostLog.connected')
              : t('settings.hostLog.reconnecting')}
        </Tag>
      </div>

      <div className="flex flex-wrap items-center gap-2">
        <Input.Search
          allowClear
          style={{ width: 220 }}
          value={query}
          placeholder={t('settings.hostLog.search')}
          onChange={(e) => setQuery(e.target.value)}
          onSearch={(value) => setApplied(value)}
        />
        <Select
          style={{ width: 170 }}
          value={severity}
          options={severityOptions.map((o) => ({
            value: o.value,
            label: t(`settings.hostLog.${o.key}`)
          }))}
          onChange={setSeverity}
        />
        <Button
          icon={paused ? <PlayIcon size={14} /> : <PauseIcon size={14} />}
          onClick={() => setPaused((p) => !p)}
        >
          {paused ? t('settings.hostLog.resume') : t('settings.hostLog.pause')}
        </Button>
        <Button
          icon={<DownloadIcon size={14} />}
          href={api.hostLogDownloadUrl(filter)}
          download="hostlog.txt"
        >
          {t('settings.hostLog.download')}
        </Button>
        <Button danger icon={<Trash2Icon size={14} />} onClick={onClear}>
          {t('settings.hostLog.clear')}
        </Button>
      </div>

      <div
        ref={boxRef}
        onScroll={onScroll}
        className="h-80 overflow-auto rounded-xl border border-neutral-700/50 bg-neutral-900/60 p-2 font-mono text-xs leading-5"
      >
        {lines.length === 0 ? (
          <span className="text-neutral-500">{t('settings.hostLog.empty')}</span>
        ) : (
          lines.map((line) => (
            <div
              key={line.seq}
              className={`break-all whitespace-pre-wrap ${severityColor(line.severity)}`}
            >
              <span className="text-neutral-500">{formatClock(line.time)}</span>{' '}
              <span className="text-neutral-500">{line.source}</span>{' '}
              {line.severity >= 0 && (
                <span className="text-neutral-500">{severityName(line.severity)}</span>
              )}{' '}
              {line.tag && <span className="text-neutral-400">{line.tag}:</span>} {line.msg}
            </div>
          ))
        )}
      </div>
    </div>
  );
};
