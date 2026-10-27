import { useCallback, useEffect, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Button, Form, Input, message, Modal, Popconfirm, Select, Switch, Tag } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/auth.ts';
import { User, UserRole } from '@/api/auth.ts';
import { encrypt } from '@/lib/encrypt.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { useServerFeature } from '@/hooks/useServerFeature.ts';
import { parseRole } from '@/lib/roles.ts';

type CreateValues = {
  username: string;
  password: string;
  role: UserRole;
};

type PasswordValues = {
  password: string;
};

export const Users = () => {
  const { t } = useTranslation();
  const { account } = useAuth();
  const [messageApi, contextHolder] = message.useMessage();
  const [createForm] = Form.useForm<CreateValues>();
  const [passwordForm] = Form.useForm<PasswordValues>();

  // The viewer role is offered only by a server that has it, and an account
  // that already is one keeps its entry in the list either way.
  const hasViewerRole = useServerFeature('role-viewer') === true;
  function roleOptions(current?: UserRole) {
    const options: { value: UserRole; label: string }[] = [
      { value: 'admin', label: t('settings.account.roles.admin') },
      { value: 'user', label: t('settings.account.roles.user') }
    ];
    if (hasViewerRole || current === 'viewer') {
      options.push({ value: 'viewer', label: t('settings.account.roles.viewer') });
    }
    return options;
  }

  const [users, setUsers] = useState<User[]>([]);
  // The list loads on mount, so the table starts in its loading state. Every
  // other caller of loadUsers sets the flag before its own request, which is
  // why loadUsers only clears it.
  const [isLoading, setIsLoading] = useState(true);
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [passwordUser, setPasswordUser] = useState<string | null>(null);

  // A promise chain rather than async/await: the lint that forbids a
  // synchronous setState in an effect cannot see that a finally block after an
  // await runs later, and reads this as if it cleared the flag immediately.
  const loadUsers = useCallback(
    () =>
      api
        .getUsers()
        .then((rsp) => {
          if (rsp.code !== 0) {
            throw rsp;
          }

          const data = Array.isArray(rsp.data) ? rsp.data : rsp.data?.users;
          setUsers(
            (Array.isArray(data) ? data : []).map((user: User) => ({
              username: user.username,
              role: parseRole(user.role),
              enabled: user.enabled !== false,
              systemAccount: user.systemAccount === true,
              totpEnabled: user.totpEnabled === true,
              oidcLinked: user.oidcLinked === true
            }))
          );
        })
        .catch((err) => {
          messageApi.error(describeFailure(err, t('settings.account.users.loadFailed')));
        })
        .finally(() => {
          setIsLoading(false);
        }),
    [messageApi, t]
  );

  useEffect(() => {
    loadUsers();
  }, [loadUsers]);

  async function createUser(values: CreateValues) {
    setIsLoading(true);
    try {
      const rsp = await api.createUser(values.username, encrypt(values.password), values.role);
      if (rsp.code !== 0) throw rsp;

      setIsCreateOpen(false);
      createForm.resetFields();
      messageApi.success(t('settings.account.users.created'));
      await loadUsers();
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.users.saveFailed')));
      setIsLoading(false);
    }
  }

  async function updateUser(user: User, changes: Partial<Pick<User, 'role' | 'enabled'>>) {
    setIsLoading(true);
    try {
      const rsp = await api.updateUser(user.username, changes);
      if (rsp.code !== 0) throw rsp;
      await loadUsers();
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.users.saveFailed')));
      setIsLoading(false);
    }
  }

  async function deleteUser(username: string) {
    setIsLoading(true);
    try {
      const rsp = await api.deleteUser(username);
      if (rsp.code !== 0) throw rsp;
      messageApi.success(t('settings.account.users.deleted'));
      await loadUsers();
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.users.deleteFailed')));
      setIsLoading(false);
    }
  }

  async function resetTotp(username: string) {
    setIsLoading(true);
    try {
      const rsp = await api.resetUserTotp(username);
      if (rsp.code !== 0) throw rsp;
      messageApi.success(t('totp.users.resetDone'));
      await loadUsers();
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.users.saveFailed')));
      setIsLoading(false);
    }
  }

  async function unlinkOidc(username: string) {
    setIsLoading(true);
    try {
      const rsp = await api.unlinkUserOidc(username);
      if (rsp.code !== 0) throw rsp;
      messageApi.success(t('settings.account.users.ssoUnlinked'));
      await loadUsers();
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.users.saveFailed')));
      setIsLoading(false);
    }
  }

  async function resetPassword(values: PasswordValues) {
    if (!passwordUser) return;
    setIsLoading(true);
    try {
      const rsp = await api.resetUserPassword(passwordUser, encrypt(values.password));
      if (rsp.code !== 0) throw rsp;
      setPasswordUser(null);
      passwordForm.resetFields();
      messageApi.success(t('settings.account.users.passwordUpdated'));
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.users.saveFailed')));
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <>
      {contextHolder}
      <div className="mb-4 flex items-center justify-between">
        <div className="text-base">{t('settings.account.users.title')}</div>
        <Button type="primary" onClick={() => setIsCreateOpen(true)}>
          {t('settings.account.users.create')}
        </Button>
      </div>

      <div className="flex flex-col space-y-2 opacity-100">
        {users.map((user) => {
          const isSelf = user.username === account.username;
          const isProtected = isSelf || user.systemAccount;
          return (
            <div
              key={user.username}
              className="flex flex-wrap items-center gap-2 rounded-md bg-neutral-800/60 px-3 py-2"
            >
              <span className="min-w-0 flex-1 truncate">
                {user.username}
                {user.systemAccount && (
                  <span className="ml-2 text-xs text-neutral-500">
                    {t('settings.account.users.deviceOwner')}
                  </span>
                )}
                {user.totpEnabled && (
                  <Tag className="ml-2" color="green">
                    {t('totp.users.tag')}
                  </Tag>
                )}
                {user.oidcLinked && (
                  <Tag className="ml-2" color="blue">
                    {t('settings.account.users.ssoTag')}
                  </Tag>
                )}
              </span>
              <Select<UserRole>
                className="w-28"
                value={user.role}
                disabled={isLoading || isProtected}
                options={roleOptions(user.role)}
                onChange={(role) => updateUser(user, { role })}
              />
              <Switch
                checked={user.enabled}
                disabled={isLoading || isProtected}
                checkedChildren={t('settings.account.users.enabled')}
                unCheckedChildren={t('settings.account.users.disabled')}
                onChange={(enabled) => updateUser(user, { enabled })}
              />
              <Button
                size="small"
                disabled={isLoading || isProtected}
                onClick={() => setPasswordUser(user.username)}
              >
                {t('settings.account.users.resetPassword')}
              </Button>
              {user.totpEnabled && !isSelf && (
                <Popconfirm
                  title={t('totp.users.resetConfirm', { username: user.username })}
                  okText={t('settings.account.okBtn')}
                  cancelText={t('settings.account.cancelBtn')}
                  onConfirm={() => resetTotp(user.username)}
                >
                  <Button size="small" disabled={isLoading}>
                    {t('totp.users.reset')}
                  </Button>
                </Popconfirm>
              )}
              {user.oidcLinked && (
                <Popconfirm
                  title={t('settings.account.users.ssoUnlinkConfirm', { username: user.username })}
                  okText={t('settings.account.okBtn')}
                  cancelText={t('settings.account.cancelBtn')}
                  onConfirm={() => unlinkOidc(user.username)}
                >
                  <Button size="small" disabled={isLoading}>
                    {t('settings.account.users.ssoUnlink')}
                  </Button>
                </Popconfirm>
              )}
              <Popconfirm
                title={t('settings.account.users.deleteConfirm')}
                okText={t('settings.account.okBtn')}
                cancelText={t('settings.account.cancelBtn')}
                disabled={isProtected}
                onConfirm={() => deleteUser(user.username)}
              >
                <Button danger size="small" disabled={isLoading || isProtected}>
                  {t('settings.account.users.delete')}
                </Button>
              </Popconfirm>
            </div>
          );
        })}
      </div>

      <Modal
        title={t('settings.account.users.create')}
        open={isCreateOpen}
        footer={null}
        destroyOnHidden
        onCancel={() => setIsCreateOpen(false)}
      >
        <Form<CreateValues>
          form={createForm}
          layout="vertical"
          initialValues={{ role: 'user' }}
          onFinish={createUser}
        >
          <Form.Item
            name="username"
            label={t('auth.placeholderUsername')}
            rules={[
              { required: true, message: t('auth.noEmptyUsername') },
              { pattern: /^[A-Za-z0-9][A-Za-z0-9_.-]{0,31}$/, message: t('auth.illegalUsername') }
            ]}
          >
            <Input autoComplete="off" />
          </Form.Item>
          <Form.Item
            name="password"
            label={t('auth.placeholderPassword')}
            rules={[
              { required: true, message: t('auth.noEmptyPassword') },
              { min: 8, max: 72, message: t('auth.passwordLength') }
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Form.Item name="role" label={t('settings.account.role')}>
            <Select options={roleOptions()} />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={isLoading} className="w-full">
            {t('settings.account.users.create')}
          </Button>
        </Form>
      </Modal>

      <Modal
        title={t('settings.account.users.resetPassword')}
        open={!!passwordUser}
        footer={null}
        destroyOnHidden
        onCancel={() => setPasswordUser(null)}
      >
        <Form<PasswordValues> form={passwordForm} layout="vertical" onFinish={resetPassword}>
          <Form.Item
            name="password"
            label={t('auth.placeholderPassword')}
            rules={[
              { required: true, message: t('auth.noEmptyPassword') },
              { min: 8, max: 72, message: t('auth.passwordLength') }
            ]}
          >
            <Input.Password autoComplete="new-password" />
          </Form.Item>
          <Button type="primary" htmlType="submit" loading={isLoading} className="w-full">
            {t('auth.ok')}
          </Button>
        </Form>
      </Modal>
    </>
  );
};
