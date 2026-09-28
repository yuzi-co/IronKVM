import { http } from '@/lib/http';

export type UserRole = 'admin' | 'user';

export type Account = {
  username: string;
  role: UserRole;
  // Whether a password change also sets the root password (SSH, console).
  systemAccount?: boolean;
};

export type User = Account & {
  enabled: boolean;
  systemAccount?: boolean;
};

export function login(username: string, password: string) {
  const data = {
    username,
    password
  };
  return http.post('/api/auth/login', data);
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
};

export type CreatedAPIKey = Omit<APIKey, 'username'> & {
  key: string;
};

export function getAPIKeys() {
  return http.get('/api/auth/api-keys');
}

export function createAPIKey(name: string) {
  return http.post('/api/auth/api-keys', { name });
}

export function revokeAPIKey(id: string) {
  return http.delete(`/api/auth/api-keys/${encodeURIComponent(id)}`);
}
