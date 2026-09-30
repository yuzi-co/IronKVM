import { http } from '@/lib/http.ts';

// Install and update download a release, which can take minutes on the
// board's link. The default request timeout is one minute.
const LONG = { timeout: 10 * 60 * 1000 };

// install tailscale
export function install() {
  return http.post('/api/extensions/tailscale/install', undefined, LONG);
}

// uninstall tailscale
export function uninstall() {
  return http.post('/api/extensions/tailscale/uninstall');
}

// get tailscale status
export function getStatus() {
  return http.get('/api/extensions/tailscale/status');
}

// start the daemon if it is not running, then tailscale up
export function connect() {
  return http.post('/api/extensions/tailscale/connect', undefined, { timeout: 3 * 60 * 1000 });
}

// tailscale down, then stop the daemon
export function disconnect() {
  return http.post('/api/extensions/tailscale/disconnect');
}

// restart tailscale
export function restart() {
  return http.post('/api/extensions/tailscale/restart');
}

// login tailscale
export function login() {
  return http.post('/api/extensions/tailscale/login');
}

// logout tailscale
export function logout() {
  return http.post('/api/extensions/tailscale/logout');
}

// turn start at boot on or off
export function setBoot(enabled: boolean) {
  return http.post('/api/extensions/tailscale/boot', { enabled });
}

// get the installed and the latest version
export function getUpdate() {
  return http.get('/api/extensions/tailscale/update');
}

// install the latest version
export function update() {
  return http.post('/api/extensions/tailscale/update', undefined, LONG);
}
