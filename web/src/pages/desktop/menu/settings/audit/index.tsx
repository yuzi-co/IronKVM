import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Divider, Input, message, Select, Switch, Table, Tag } from 'antd';
import type { ColumnsType } from 'antd/es/table';
import { DownloadIcon, RefreshCwIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/audit.ts';
import type { AuditEvent, AuditFilter, AuditSettings } from '@/api/audit.ts';
import { describeFailure } from '@/lib/feedback.ts';

import { Section } from '../components/section.tsx';
import { auditFileName, subject, via, who } from './format.ts';

const pageSize = 25;

const outcomeColors: Record<string, string> = { ok: 'green', fail: 'red', denied: 'orange' };

// Audit is the log of who did what on the board: logins, control, power,
// media, settings, macros and what MCP clients and the agent did. The server
// keeps it in a few small files and drops the oldest events when they are
// full; this page reads it, exports it and chooses whether it goes to syslog.
export const Audit = () => {
  const { t } = useTranslation();
  const [messageApi, messageHolder] = message.useMessage();

  const [filter, setFilter] = useState<AuditFilter>({});
  // What the text box holds before it is applied with Enter.
  const [search, setSearch] = useState('');
  const [page, setPage] = useState(1);
  const [events, setEvents] = useState<AuditEvent[]>([]);
  const [total, setTotal] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState('');
  const [settings, setSettings] = useState<AuditSettings | null>(null);
  const [isExporting, setIsExporting] = useState(false);

  const load = useCallback(
    () =>
      api
        .getAudit(filter, (page - 1) * pageSize, pageSize)
        .then((rsp) => {
          if (rsp.code !== 0) throw rsp;
          setEvents(rsp.data?.events ?? []);
          setTotal(rsp.data?.total ?? 0);
          setError('');
        })
        .catch((err) => setError(describeFailure(err, t('settings.audit.loadFailed'))))
        .finally(() => setIsLoading(false)),
    [filter, page, t]
  );

  useEffect(() => {
    void load();
  }, [load]);

  useEffect(() => {
    api
      .getAuditSettings()
      .then((rsp) => {
        if (rsp.code === 0) setSettings(rsp.data as AuditSettings);
      })
      .catch(() => {});
  }, []);

  function changeFilter(next: AuditFilter) {
    setIsLoading(true);
    setFilter(next);
    setPage(1);
  }

  function refresh() {
    setIsLoading(true);
    void load();
  }

  async function exportLog() {
    if (isExporting) return;
    setIsExporting(true);
    try {
      const blob = await api.exportAudit(filter);
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = auditFileName(window.location.hostname, new Date());
      link.click();
      URL.revokeObjectURL(url);
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.audit.exportFailed')));
    } finally {
      setIsExporting(false);
    }
  }

  function setSyslog(syslog: boolean) {
    api
      .setAuditSyslog(syslog)
      .then((rsp) => {
        if (rsp.code !== 0) throw rsp;
        setSettings(rsp.data as AuditSettings);
      })
      .catch((err) => messageApi.error(describeFailure(err, t('settings.audit.saveFailed'))));
  }

  const columns: ColumnsType<AuditEvent> = [
    {
      title: t('settings.audit.time'),
      dataIndex: 'time',
      width: 160,
      render: (value: string) => new Date(value).toLocaleString()
    },
    {
      title: t('settings.audit.who'),
      width: 150,
      render: (_, event) => (
        <div className="flex flex-col">
          <span>{who(event)}</span>
          {(event.via || event.session) && (
            <span className="text-xs text-neutral-500" title={event.session}>
              {via(event)}
            </span>
          )}
        </div>
      )
    },
    {
      title: t('settings.audit.address'),
      dataIndex: 'ip',
      width: 130,
      render: (value?: string) => <span className="font-mono text-xs">{value || '-'}</span>
    },
    {
      title: t('settings.audit.category'),
      dataIndex: 'category',
      width: 100,
      render: (value: string) => t(`settings.audit.categories.${value}`, { defaultValue: value })
    },
    {
      title: t('settings.audit.action'),
      dataIndex: 'action',
      width: 150,
      render: (value: string, event) => (
        <div className="flex flex-col">
          <span>{value}</span>
          {event.tier && (
            <span className="text-xs text-neutral-500">
              {t('settings.audit.tier', { tier: event.tier })}
            </span>
          )}
        </div>
      )
    },
    {
      title: t('settings.audit.target'),
      render: (_, event) => <span className="break-all">{subject(event)}</span>
    },
    {
      title: t('settings.audit.outcome'),
      dataIndex: 'outcome',
      width: 90,
      render: (value: string) => (
        <Tag color={outcomeColors[value] ?? 'default'}>
          {t(`settings.audit.outcomes.${value}`, { defaultValue: value })}
        </Tag>
      )
    }
  ];

  return (
    <>
      {messageHolder}
      <div className="text-base">{t('settings.audit.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <Section
          title={t('settings.audit.log')}
          description={t('settings.audit.description')}
          action={
            <div className="flex shrink-0 gap-2">
              <Button
                icon={<RefreshCwIcon size={14} />}
                loading={isLoading}
                aria-label={t('settings.audit.refresh')}
                onClick={refresh}
              />
              <Button
                icon={<DownloadIcon size={14} />}
                loading={isExporting}
                onClick={() => void exportLog()}
              >
                {t('settings.audit.export')}
              </Button>
            </div>
          }
        >
          <div className="flex flex-wrap gap-2">
            <Select
              allowClear
              className="w-36"
              placeholder={t('settings.audit.category')}
              value={filter.category}
              options={api.AUDIT_CATEGORIES.map((value) => ({
                value,
                label: t(`settings.audit.categories.${value}`)
              }))}
              onChange={(value) => changeFilter({ ...filter, category: value })}
            />
            <Select
              allowClear
              className="w-32"
              placeholder={t('settings.audit.outcome')}
              value={filter.outcome}
              options={api.AUDIT_OUTCOMES.map((value) => ({
                value,
                label: t(`settings.audit.outcomes.${value}`)
              }))}
              onChange={(value) => changeFilter({ ...filter, outcome: value })}
            />
            <Input.Search
              allowClear
              className="w-64"
              placeholder={t('settings.audit.searchPlaceholder')}
              value={search}
              onChange={(e) => {
                setSearch(e.target.value);
                if (!e.target.value && filter.q) changeFilter({ ...filter, q: undefined });
              }}
              onSearch={(value) => changeFilter({ ...filter, q: value.trim() || undefined })}
            />
          </div>

          {error && <Alert type="error" showIcon message={error} />}

          <Table<AuditEvent>
            size="small"
            rowKey={(event) => `${event.time}-${event.actor}-${event.action}-${event.target}`}
            columns={columns}
            dataSource={events}
            loading={isLoading}
            scroll={{ x: 'max-content' }}
            locale={{ emptyText: t('settings.audit.empty') }}
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: false,
              hideOnSinglePage: true,
              onChange: (next) => {
                setIsLoading(true);
                setPage(next);
              }
            }}
          />
        </Section>

        <Section
          title={t('settings.audit.syslog')}
          description={t('settings.audit.syslogDescription')}
          action={
            <Switch
              checked={settings?.syslog ?? false}
              disabled={!settings}
              onChange={setSyslog}
              aria-label={t('settings.audit.syslog')}
            />
          }
        />

        {settings && (
          <span className="text-xs text-neutral-500">
            {t('settings.audit.storage', {
              directory: settings.directory,
              size: Math.round(settings.maxBytes / 1024)
            })}
          </span>
        )}
      </div>
    </>
  );
};
