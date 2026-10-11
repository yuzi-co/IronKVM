import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Form, Input, message, Modal, Select } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/auth.ts';
import type { CreatedShareLink, ShareLink } from '@/api/auth.ts';
import { describeFailure } from '@/lib/feedback.ts';
import {
  DEFAULT_SHARE_DURATION,
  SHARE_DURATIONS,
  shareDurationSeconds,
  shareLinkUrl
} from '@/lib/share-link.ts';
import type { ShareDurationKey } from '@/lib/share-link.ts';

import { CopyButton } from '../components/copy-button.tsx';

type CreateValues = {
  label: string;
  duration: ShareDurationKey;
};

// The server cuts a longer label.
const maxLabelLength = 64;

function formatTime(seconds: number) {
  return new Date(seconds * 1000).toLocaleString();
}

// ShareLinks lists the view-only links, makes new ones and revokes them. A
// link shows the screen to whoever opens it, with no control of the host, and
// stops working when it expires or is revoked.
export const ShareLinks = () => {
  const { t } = useTranslation();
  const [messageApi, messageHolder] = message.useMessage();
  const [modal, modalHolder] = Modal.useModal();
  const [createForm] = Form.useForm<CreateValues>();

  const [links, setLinks] = useState<ShareLink[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [isCreating, setIsCreating] = useState(false);
  // The link just made, with its token. It lives only in this state: the
  // device keeps a digest, and the dialog is the one place the URL appears.
  const [created, setCreated] = useState<{ url: string; link: CreatedShareLink } | null>(null);

  // A promise chain rather than async/await, for the same reason as the users
  // list: the effect lint cannot tell that a finally after an await runs later.
  const loadLinks = useCallback(
    () =>
      api
        .getShareLinks()
        .then((rsp) => {
          if (rsp.code !== 0) throw rsp;
          const listed: ShareLink[] = Array.isArray(rsp.data?.links) ? rsp.data.links : [];
          setLinks(listed);
        })
        .catch((err) => {
          messageApi.error(describeFailure(err, t('settings.account.shareLinks.loadFailed')));
        })
        .finally(() => {
          setIsLoading(false);
        }),
    [messageApi, t]
  );

  useEffect(() => {
    loadLinks();
  }, [loadLinks]);

  async function createLink(values: CreateValues) {
    setIsCreating(true);
    try {
      const rsp = await api.createShareLink(
        (values.label ?? '').trim(),
        shareDurationSeconds(values.duration)
      );
      if (rsp.code !== 0) throw rsp;

      const link = rsp.data as CreatedShareLink;
      createForm.resetFields();
      setCreated({ link, url: shareLinkUrl(window.location, link.token) });
      await loadLinks();
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.shareLinks.createFailed')));
    } finally {
      setIsCreating(false);
    }
  }

  function revokeLink(link: ShareLink) {
    modal.confirm({
      title: t('settings.account.shareLinks.revokeConfirm'),
      okText: t('settings.account.shareLinks.revoke'),
      okButtonProps: { danger: true },
      cancelText: t('settings.account.cancelBtn'),
      onOk: async () => {
        try {
          const rsp = await api.revokeShareLink(link.id);
          if (rsp.code !== 0) throw rsp;
          messageApi.success(t('settings.account.shareLinks.revoked'));
        } catch (err) {
          messageApi.error(describeFailure(err, t('settings.account.shareLinks.revokeFailed')));
        }
        await loadLinks();
      }
    });
  }

  return (
    <>
      {messageHolder}
      {modalHolder}
      <div className="mb-1 text-base">{t('settings.account.shareLinks.title')}</div>
      <div className="mb-4 text-sm text-neutral-400">{t('settings.account.shareLinks.desc')}</div>

      <Form<CreateValues>
        form={createForm}
        layout="inline"
        className="mb-4 flex w-full flex-wrap gap-2"
        initialValues={{ label: '', duration: DEFAULT_SHARE_DURATION }}
        onFinish={createLink}
      >
        <Form.Item name="label" className="!me-0 min-w-0 flex-1">
          <Input
            autoComplete="off"
            maxLength={maxLabelLength}
            placeholder={t('settings.account.shareLinks.label')}
          />
        </Form.Item>
        <Form.Item name="duration" className="!me-0">
          <Select<ShareDurationKey>
            className="w-32"
            aria-label={t('settings.account.shareLinks.duration')}
            options={SHARE_DURATIONS.map((duration) => ({
              value: duration.key,
              label: t(`settings.account.shareLinks.durations.${duration.key}`)
            }))}
          />
        </Form.Item>
        <Button type="primary" htmlType="submit" loading={isCreating}>
          {t('settings.account.shareLinks.create')}
        </Button>
      </Form>

      <div className="flex flex-col overflow-hidden rounded-xl border border-neutral-700/50 bg-neutral-800/40">
        {links.length === 0 ? (
          <div className="px-4 py-3.5 text-sm text-neutral-500">
            {isLoading ? '...' : t('settings.account.shareLinks.empty')}
          </div>
        ) : (
          links.map((link, index) => (
            <div
              key={link.id}
              className={`flex items-center justify-between gap-3 px-4 py-3 ${
                index > 0 ? 'border-t border-neutral-800' : ''
              }`}
            >
              <div className="flex min-w-0 flex-col">
                <span className="truncate text-sm font-medium">{link.label || link.id}</span>
                <span className="truncate text-xs text-neutral-500">
                  {t('settings.account.shareLinks.expires', { time: formatTime(link.expiresAt) })}
                  {link.createdBy && ` · ${link.createdBy}`}
                </span>
              </div>
              <Button danger size="small" disabled={isLoading} onClick={() => revokeLink(link)}>
                {t('settings.account.shareLinks.revoke')}
              </Button>
            </div>
          ))
        )}
      </div>

      <Modal
        title={t('settings.account.shareLinks.createdTitle')}
        open={!!created}
        destroyOnHidden
        maskClosable={false}
        onCancel={() => setCreated(null)}
        footer={
          <Button type="primary" onClick={() => setCreated(null)}>
            {t('auth.ok')}
          </Button>
        }
      >
        <div className="flex flex-col space-y-4">
          <Alert type="warning" showIcon message={t('settings.account.shareLinks.createdDesc')} />
          <div className="flex items-center gap-2 rounded-lg border border-neutral-700/50 bg-neutral-800/40 px-3 py-2">
            <span className="min-w-0 flex-1 font-mono text-sm break-all text-neutral-300 select-all">
              {created?.url}
            </span>
            <CopyButton text={created?.url ?? ''} />
          </div>
          {created && (
            <span className="text-xs text-neutral-500">
              {t('settings.account.shareLinks.expires', {
                time: formatTime(created.link.expiresAt)
              })}
            </span>
          )}
        </div>
      </Modal>
    </>
  );
};
