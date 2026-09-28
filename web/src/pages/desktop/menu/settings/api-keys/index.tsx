import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Alert, Button, Divider, Form, Input, message, Modal } from 'antd';
import { CheckIcon, CopyIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/auth.ts';
import type { APIKey, CreatedAPIKey } from '@/api/auth.ts';
import { writeClipboardText } from '@/lib/clipboard.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { getBaseUrl, getHostname, getPort } from '@/lib/service.ts';

import { CopyBlock, CopyRow } from '../components/copy-button.tsx';

type CreateValues = {
  name: string;
};

// The server refuses a longer name.
const maxNameLength = 64;

function formatCreatedAt(createdAt: number) {
  return createdAt ? new Date(createdAt * 1000).toLocaleString() : '-';
}

// scrapeConfigFor is a Prometheus job for this board, as server/router/metrics.go
// describes it: the key goes in the authorization block, which Prometheus sends
// as a Bearer token. Over https the board's own certificate is self-signed, so
// the job skips verification until a trusted one is installed.
function scrapeConfigFor(https: boolean, target: string) {
  const lines = [
    'scrape_configs:',
    '  - job_name: ironkvm',
    `    scheme: ${https ? 'https' : 'http'}`,
    '    metrics_path: /api/metrics',
    '    authorization:',
    '      credentials: <api key>'
  ];
  if (https) lines.push('    tls_config:', '      insecure_skip_verify: true');
  lines.push('    static_configs:', `      - targets: ['${target}']`);
  return lines.join('\n');
}

export const APIKeys = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  // The server lists every account's keys to an administrator, so the owner
  // is only worth a column there.
  const isAdmin = account.role === 'admin';

  const [messageApi, messageHolder] = message.useMessage();

  const metricsUrl = `${getBaseUrl('http')}/api/metrics`;
  const scrapeConfig = scrapeConfigFor(
    window.location.protocol === 'https:',
    `${getHostname()}:${getPort()}`
  );
  const [modal, modalHolder] = Modal.useModal();
  const [createForm] = Form.useForm<CreateValues>();

  const [keys, setKeys] = useState<APIKey[]>([]);
  // The list loads on mount, so it starts in its loading state and loadKeys
  // only clears the flag.
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  // The secret of a key just created. It lives only in this state: the server
  // does not keep it, and it is dropped when the dialog closes.
  const [createdKey, setCreatedKey] = useState<CreatedAPIKey | null>(null);
  const [isCopied, setIsCopied] = useState(false);

  // A promise chain rather than async/await, for the same reason as the users
  // list: the effect lint cannot tell that a finally after an await runs later.
  const loadKeys = useCallback(
    () =>
      api
        .getAPIKeys()
        .then((rsp) => {
          if (rsp.code !== 0) throw rsp;

          const listed: APIKey[] = Array.isArray(rsp.data?.keys) ? rsp.data.keys : [];
          setKeys([...listed].sort((a, b) => b.createdAt - a.createdAt));
        })
        .catch((err) => {
          messageApi.error(describeFailure(err, t('settings.apiKeys.loadFailed')));
        })
        .finally(() => {
          setIsLoading(false);
        }),
    [messageApi, t]
  );

  useEffect(() => {
    loadKeys();
  }, [loadKeys]);

  async function createKey(values: CreateValues) {
    setIsCreating(true);
    try {
      const rsp = await api.createAPIKey(values.name.trim());
      if (rsp.code !== 0) throw rsp;

      createForm.resetFields();
      setIsCopied(false);
      setCreatedKey(rsp.data as CreatedAPIKey);
      await loadKeys();
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.apiKeys.createFailed')));
    } finally {
      setIsCreating(false);
    }
  }

  function revokeKey(key: APIKey) {
    modal.confirm({
      title: t('settings.apiKeys.revokeConfirmTitle'),
      content: (
        <span className="text-sm text-neutral-400">
          {t('settings.apiKeys.revokeConfirmDesc', {
            name: key.name || t('settings.apiKeys.unnamed')
          })}
        </span>
      ),
      okText: t('settings.apiKeys.revoke'),
      okButtonProps: { danger: true },
      cancelText: t('settings.apiKeys.cancelBtn'),
      onOk: async () => {
        try {
          const rsp = await api.revokeAPIKey(key.id);
          if (rsp.code !== 0) throw rsp;
          messageApi.success(t('settings.apiKeys.revoked'));
        } catch (err) {
          messageApi.error(describeFailure(err, t('settings.apiKeys.revokeFailed')));
        }
        await loadKeys();
      }
    });
  }

  async function copyCreatedKey() {
    if (!createdKey?.key) return;
    try {
      await writeClipboardText(createdKey.key);
      setIsCopied(true);
    } catch {
      messageApi.error(t('settings.apiKeys.copyFailed'));
    }
  }

  function closeCreatedKey() {
    setCreatedKey(null);
    setIsCopied(false);
  }

  return (
    <>
      {messageHolder}
      {modalHolder}
      <div className="text-base">{t('settings.apiKeys.title')}</div>
      <Divider className="opacity-50" />

      <div className="flex flex-col space-y-6">
        <div className="flex flex-col space-y-1">
          <span className="text-sm text-neutral-400">{t('settings.apiKeys.description')}</span>
          <span className="text-xs text-neutral-500">{t('settings.apiKeys.mcpNote')}</span>
        </div>

        <Form<CreateValues>
          form={createForm}
          layout="inline"
          className="flex w-full flex-nowrap gap-2"
          onFinish={createKey}
        >
          <Form.Item
            name="name"
            className="!me-0 min-w-0 flex-1"
            rules={[
              { required: true, whitespace: true, message: t('settings.apiKeys.nameRequired') },
              { max: maxNameLength, message: t('settings.apiKeys.nameTooLong') }
            ]}
          >
            <Input
              autoComplete="off"
              maxLength={maxNameLength}
              placeholder={t('settings.apiKeys.namePlaceholder')}
            />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={isCreating}>
            {t('settings.apiKeys.create')}
          </Button>
        </Form>

        <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
          {keys.length === 0 ? (
            <div className="px-4 py-3.5 text-sm text-neutral-500">
              {isLoading ? '...' : t('settings.apiKeys.empty')}
            </div>
          ) : (
            keys.map((key, index) => (
              <div
                key={key.id}
                className={`flex items-center justify-between gap-3 px-4 py-3 ${
                  index > 0 ? 'border-t border-neutral-800' : ''
                }`}
              >
                <div className="flex min-w-0 flex-col">
                  <span className="truncate text-sm font-medium">
                    {key.name || t('settings.apiKeys.unnamed')}
                  </span>
                  <span className="truncate text-xs text-neutral-500">
                    {t('settings.apiKeys.created')} {formatCreatedAt(key.createdAt)}
                    {isAdmin && key.username && (
                      <>
                        {' · '}
                        {t('settings.apiKeys.owner')} {key.username}
                      </>
                    )}
                  </span>
                </div>
                <Button danger size="small" disabled={isLoading} onClick={() => revokeKey(key)}>
                  {t('settings.apiKeys.revoke')}
                </Button>
              </div>
            ))
          )}
        </div>

        <div className="flex flex-col space-y-3">
          <span className="text-sm font-medium">{t('settings.apiKeys.monitoring')}</span>
          <span className="text-xs text-neutral-500">{t('settings.apiKeys.monitoringDesc')}</span>
          <div className="rounded-xl border border-neutral-700/50 bg-neutral-800/40 px-4 py-3 text-sm">
            <CopyRow label={t('settings.apiKeys.metricsUrl')} value={metricsUrl} />
          </div>
          <CopyBlock title={t('settings.apiKeys.scrapeConfig')} text={scrapeConfig} />
        </div>
      </div>

      <Modal
        title={t('settings.apiKeys.newKeyTitle')}
        open={!!createdKey}
        destroyOnHidden
        maskClosable={false}
        onCancel={closeCreatedKey}
        footer={
          <Button type="primary" onClick={closeCreatedKey}>
            {t('settings.apiKeys.done')}
          </Button>
        }
      >
        <div className="flex flex-col space-y-4">
          <Alert type="warning" showIcon message={t('settings.apiKeys.newKeyWarning')} />
          <div className="flex items-center gap-2 rounded-lg border border-neutral-700/50 bg-neutral-800/40 px-3 py-2">
            <span className="min-w-0 flex-1 font-mono text-sm break-all text-neutral-300 select-all">
              {createdKey?.key}
            </span>
            <Button
              size="small"
              icon={
                isCopied ? (
                  <CheckIcon size={15} className="text-green-500" />
                ) : (
                  <CopyIcon size={15} />
                )
              }
              onClick={copyCreatedKey}
            >
              {isCopied ? t('settings.apiKeys.copied') : t('settings.apiKeys.copy')}
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
};
