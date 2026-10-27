import { useCallback, useEffect, useState } from 'react';
import { Alert, Button, Checkbox, Form, Input, message, Select, Switch } from 'antd';
import { useTranslation } from 'react-i18next';

import * as api from '@/api/auth.ts';
import type { OidcSettings, UserRole } from '@/api/auth.ts';
import { describeFailure } from '@/lib/feedback.ts';
import { httpRedirectWarning, scopesFromText, scopesToText } from '@/lib/oidc.ts';

import { CopyRow } from '../components/copy-button.tsx';
import { Section } from '../components/section.tsx';

type Values = Omit<OidcSettings, 'clientSecretSet' | 'redirectUrl' | 'scopes'> & {
  scopes: string;
  clientSecret: string;
  clearClientSecret: boolean;
};

function toValues(s: OidcSettings): Values {
  return {
    enabled: s.enabled,
    providerName: s.providerName,
    issuer: s.issuer,
    clientId: s.clientId,
    clientSecret: '',
    clearClientSecret: false,
    scopes: scopesToText(s.scopes),
    usernameClaim: s.usernameClaim,
    emailDomain: s.emailDomain ?? '',
    autoCreate: s.autoCreate,
    defaultRole: s.defaultRole,
    groupsClaim: s.groupsClaim ?? '',
    adminGroup: s.adminGroup ?? '',
    userGroup: s.userGroup ?? '',
    viewerGroup: s.viewerGroup ?? '',
    hidePasswordLogin: s.hidePasswordLogin,
    redirectUrlOverride: s.redirectUrlOverride ?? ''
  };
}

// Oidc is the administrator's form for login through an OpenID Connect
// provider. The client secret is write-only: the board says whether one is set
// and never sends it back.
export const Oidc = () => {
  const { t } = useTranslation();
  const [messageApi, contextHolder] = message.useMessage();
  const [form] = Form.useForm<Values>();

  const [settings, setSettings] = useState<OidcSettings | null>(null);
  const [isSaving, setIsSaving] = useState(false);
  const usernameClaim = Form.useWatch('usernameClaim', form);
  const clearSecret = Form.useWatch('clearClientSecret', form);

  const apply = useCallback(
    (s: OidcSettings) => {
      setSettings(s);
      form.setFieldsValue(toValues(s));
    },
    [form]
  );

  useEffect(() => {
    api
      .getOidcSettings()
      .then((rsp) => {
        if (rsp.code !== 0) throw rsp;
        apply(rsp.data);
      })
      .catch((err) => {
        messageApi.error(describeFailure(err, t('settings.account.oidc.loadFailed')));
      });
  }, [apply, messageApi, t]);

  async function save(values: Values) {
    setIsSaving(true);
    try {
      const rsp = await api.setOidcSettings({
        ...values,
        scopes: scopesFromText(values.scopes),
        clientSecret: values.clearClientSecret ? '' : values.clientSecret.trim(),
        redirectUrlOverride: values.redirectUrlOverride.trim()
      });
      if (rsp.code !== 0) throw rsp;
      apply(rsp.data);
      messageApi.success(t('settings.account.oidc.saved'));
    } catch (err) {
      messageApi.error(describeFailure(err, t('settings.account.oidc.saveFailed')));
    } finally {
      setIsSaving(false);
    }
  }

  const roleOptions: { value: UserRole; label: string }[] = [
    { value: 'viewer', label: t('settings.account.roles.viewer') },
    { value: 'user', label: t('settings.account.roles.user') },
    { value: 'admin', label: t('settings.account.roles.admin') }
  ];

  return (
    <Section title={t('settings.account.oidc.title')} description={t('settings.account.oidc.desc')}>
      {contextHolder}
      {settings && (
        <>
          <CopyRow label={t('settings.account.oidc.redirectUrl')} value={settings.redirectUrl} />
          <span className="text-xs text-neutral-500">
            {t('settings.account.oidc.redirectUrlDesc')}
          </span>
          {httpRedirectWarning(settings.redirectUrl) && (
            <Alert type="warning" showIcon message={t('settings.account.oidc.httpWarning')} />
          )}

          <Form<Values> form={form} layout="vertical" onFinish={save}>
            <Form.Item
              name="enabled"
              label={t('settings.account.oidc.enabled')}
              valuePropName="checked"
            >
              <Switch />
            </Form.Item>
            <Form.Item name="providerName" label={t('settings.account.oidc.providerName')}>
              <Input maxLength={64} placeholder="Authentik" />
            </Form.Item>
            <Form.Item name="issuer" label={t('settings.account.oidc.issuer')}>
              <Input placeholder="https://auth.example.com/application/o/ironkvm/" />
            </Form.Item>
            <Form.Item name="clientId" label={t('settings.account.oidc.clientId')}>
              <Input />
            </Form.Item>
            <Form.Item
              name="clientSecret"
              label={t('settings.account.oidc.clientSecret')}
              extra={settings.clientSecretSet ? t('settings.account.oidc.clientSecretSet') : undefined}
            >
              <Input.Password autoComplete="new-password" disabled={!!clearSecret} />
            </Form.Item>
            {settings.clientSecretSet && (
              <Form.Item name="clearClientSecret" valuePropName="checked">
                <Checkbox>{t('settings.account.oidc.clearSecret')}</Checkbox>
              </Form.Item>
            )}
            <Form.Item name="scopes" label={t('settings.account.oidc.scopes')}>
              <Input placeholder="profile email groups" />
            </Form.Item>
            <Form.Item
              name="redirectUrlOverride"
              label={t('settings.account.oidc.redirectUrlOverride')}
              extra={t('settings.account.oidc.redirectUrlOverrideDesc')}
            >
              <Input placeholder="https://kvm.example.com/api/auth/oidc/callback" />
            </Form.Item>

            <div className="pb-3 text-xs text-neutral-500">
              {t('settings.account.oidc.mappingDesc')}
            </div>
            <Form.Item name="usernameClaim" label={t('settings.account.oidc.usernameClaim')}>
              <Select
                options={[
                  { value: 'preferred_username', label: 'preferred_username' },
                  { value: 'email', label: 'email' }
                ]}
              />
            </Form.Item>
            {usernameClaim === 'email' && (
              <Form.Item
                name="emailDomain"
                label={t('settings.account.oidc.emailDomain')}
                extra={t('settings.account.oidc.emailDomainDesc')}
              >
                <Input placeholder="example.com" />
              </Form.Item>
            )}
            <Form.Item name="autoCreate" valuePropName="checked">
              <Checkbox>{t('settings.account.oidc.autoCreate')}</Checkbox>
            </Form.Item>
            <Form.Item name="defaultRole" label={t('settings.account.oidc.defaultRole')}>
              <Select options={roleOptions} />
            </Form.Item>

            <div className="pb-3 text-xs text-neutral-500">
              {t('settings.account.oidc.groupsDesc')}
            </div>
            <Form.Item name="groupsClaim" label={t('settings.account.oidc.groupsClaim')}>
              <Input placeholder="groups" />
            </Form.Item>
            <Form.Item name="adminGroup" label={t('settings.account.oidc.adminGroup')}>
              <Input />
            </Form.Item>
            <Form.Item name="userGroup" label={t('settings.account.oidc.userGroup')}>
              <Input />
            </Form.Item>
            <Form.Item name="viewerGroup" label={t('settings.account.oidc.viewerGroup')}>
              <Input />
            </Form.Item>

            <Form.Item
              name="hidePasswordLogin"
              valuePropName="checked"
              extra={t('settings.account.oidc.hidePasswordDesc')}
            >
              <Checkbox>{t('settings.account.oidc.hidePassword')}</Checkbox>
            </Form.Item>

            <Button type="primary" htmlType="submit" loading={isSaving}>
              {t('settings.account.oidc.save')}
            </Button>
          </Form>
        </>
      )}
    </Section>
  );
};
