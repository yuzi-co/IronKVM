import { http } from '@/lib/http';

import type { UserRole } from '@/lib/roles.ts';

export type { UserRole };

export type Account = {
  username: string;
  role: UserRole;
  // Whether a password change also sets the root password (SSH, console).
  systemAccount?: boolean;
  // Whether the web login asks this account for a TOTP code.
  totpEnabled?: boolean;
};

export type User = Account & {
  enabled: boolean;
  systemAccount?: boolean;
  // Whether a single sign-on has linked the account to an account at the
  // identity provider. Later sign-ins must come from that provider account.
  oidcLinked?: boolean;
};

export function login(username: string, password: string) {
  const data = {
    username,
    password
  };
  return http.post('/api/auth/login', data);
}

// The login reply when the password was right and the account has TOTP on.
// The challenge names the half finished login for loginTotp.
export const LOGIN_TOTP_REQUIRED = -6;
// The challenge expired or was used up: start again from the password.
export const LOGIN_TOTP_EXPIRED = -7;

export function loginTotp(challenge: string, code: string) {
  return http.post('/api/auth/login/totp', { challenge, code });
}

export type TotpState = {
  enabled: boolean;
  // Unused recovery codes left.
  recoveryCodes: number;
};

export type TotpSetup = {
  secret: string;
  uri: string;
};

export function getTotp() {
  return http.get('/api/auth/totp');
}

// The password is encrypted the way the login form sends it.
export function setupTotp(password: string) {
  return http.post('/api/auth/totp/setup', { password });
}

export function enableTotp(code: string) {
  return http.post('/api/auth/totp/enable', { code });
}

export function disableTotp(password: string) {
  return http.post('/api/auth/totp/disable', { password });
}

export function regenerateRecoveryCodes(password: string) {
  return http.post('/api/auth/totp/recovery-codes', { password });
}

export function resetUserTotp(username: string) {
  return http.delete(`/api/auth/users/${encodeURIComponent(username)}/totp`);
}

// Forget which identity provider account an account is linked to, for a user
// who was recreated at the provider. The next single sign-on links it again.
export function unlinkUserOidc(username: string) {
  return http.delete(`/api/auth/users/${encodeURIComponent(username)}/oidc`);
}

// Login through an OpenID Connect provider. The login page sends the browser
// to OIDC_LOGIN_URL, which redirects to the provider; the provider sends it
// back to the server's callback, which redirects to / or to the login page
// with ?oidc=<reason>.
export const OIDC_LOGIN_URL = '/api/auth/oidc/login';

export type OidcInfo = {
  enabled: boolean;
  providerName?: string;
  hidePasswordLogin?: boolean;
};

export function getOidcInfo() {
  return http.get('/api/auth/oidc');
}

export type OidcSettings = {
  enabled: boolean;
  providerName: string;
  issuer: string;
  clientId: string;
  // The secret itself never comes back, only whether one is set.
  clientSecretSet: boolean;
  scopes: string[];
  usernameClaim: 'preferred_username' | 'email';
  emailDomain: string;
  autoCreate: boolean;
  defaultRole: UserRole;
  groupsClaim: string;
  adminGroup: string;
  userGroup: string;
  viewerGroup: string;
  hidePasswordLogin: boolean;
  // The callback address to register at the provider: the override when one is
  // set, else the board's as this page reached it.
  redirectUrl: string;
  // Replaces the derived redirect URL, for a reverse proxy; empty when unset.
  redirectUrlOverride: string;
};

// An empty clientSecret keeps the one on the board; clearClientSecret removes it.
export type OidcSettingsUpdate = Omit<OidcSettings, 'clientSecretSet' | 'redirectUrl'> & {
  clientSecret?: string;
  clearClientSecret?: boolean;
};

export function getOidcSettings() {
  return http.get('/api/auth/oidc/settings');
}

export function setOidcSettings(settings: OidcSettingsUpdate) {
  return http.post('/api/auth/oidc/settings', settings);
}

export function logout() {
  return http.post('/api/auth/logout');
}

export function getAccount() {
  return http.get('/api/auth/account');
}

export function changePassword(currentPassword: string, password: string) {
  return http.post('/api/auth/password', { currentPassword, password });
}

export function isPasswordUpdated() {
  return http.get('/api/auth/password');
}

export function getUsers() {
  return http.get('/api/auth/users');
}

export function createUser(username: string, password: string, role: UserRole) {
  return http.post('/api/auth/users', { username, password, role });
}

export function updateUser(username: string, data: Partial<Pick<User, 'role' | 'enabled'>>) {
  return http.request({
    method: 'put',
    url: `/api/auth/users/${encodeURIComponent(username)}`,
    data
  });
}

export function deleteUser(username: string) {
  return http.delete(`/api/auth/users/${encodeURIComponent(username)}`);
}

export function resetUserPassword(username: string, password: string) {
  return http.post(`/api/auth/users/${encodeURIComponent(username)}/password`, { password });
}

// APIKey describes an issued key. The secret is not part of it: the device
// keeps only a digest, and hands the secret back once, when the key is created.
export type APIKey = {
  id: string;
  name: string;
  createdAt: number;
  username: string;
  // full acts as the owner; view can only watch the video. A server without
  // the api-key-scope feature sends none, and every key there is full.
  scope?: APIKeyScope;
};

export type APIKeyScope = 'full' | 'view';

export type CreatedAPIKey = Omit<APIKey, 'username'> & {
  key: string;
};

export function getAPIKeys() {
  return http.get('/api/auth/api-keys');
}

export function createAPIKey(name: string, scope?: APIKeyScope) {
  return http.post('/api/auth/api-keys', scope ? { name, scope } : { name });
}

export function revokeAPIKey(id: string) {
  return http.delete(`/api/auth/api-keys/${encodeURIComponent(id)}`);
}

// ShareLink is an expiring view-only link. The token is not part of it: the
// device keeps only a digest and hands the token back once, when the link is
// created.
export type ShareLink = {
  id: string;
  label: string;
  createdBy: string;
  createdAt: number;
  expiresAt: number;
};

export type CreatedShareLink = ShareLink & {
  token: string;
};

export function getShareLinks() {
  return http.get('/api/auth/share-links');
}

export function createShareLink(label: string, ttlSeconds: number) {
  return http.post('/api/auth/share-links', { label, ttlSeconds });
}

export function revokeShareLink(id: string) {
  return http.delete(`/api/auth/share-links/${encodeURIComponent(id)}`);
}

// redeemShareLink opens a view-only session for whoever holds a link's token.
export function redeemShareLink(token: string) {
  return http.post('/api/auth/share', { token });
}

// LoginSession is one sign-in, until it expires or is revoked. Times are Unix
// seconds. The id is a digest; the secret the cookie carries is never sent.
export type LoginSession = {
  id: string;
  username: string;
  createdAt: number;
  lastUsedAt: number;
  expiresAt: number;
  ip: string;
  client: string;
  current: boolean;
};

export function getSessions() {
  return http.get('/api/auth/sessions');
}

// getAllSessions lists every account's sessions, for an administrator.
export function getAllSessions() {
  return http.get('/api/auth/sessions/all');
}

export function revokeSession(id: string) {
  return http.delete(`/api/auth/sessions/${encodeURIComponent(id)}`);
}

export function revokeOtherSessions() {
  return http.post('/api/auth/sessions/revoke-others');
}
