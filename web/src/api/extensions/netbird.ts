import { http } from '@/lib/http.ts';

// Install and update fetch Alpine's package, about 15 MB, and unpack a 43 MB
// binary onto the SD card. The default request timeout is one minute.
const LONG = { timeout: 10 * 60 * 1000 };

// install netbird
export function install() {
  return http.post('/api/extensions/netbird/install', undefined, LONG);
}

// uninstall netbird
export function uninstall() {
  return http.post('/api/extensions/netbird/uninstall');
}

// get netbird status
export function getStatus() {
  return http.get('/api/extensions/netbird/status');
}

// start netbird
export function start() {
  return http.post('/api/extensions/netbird/start');
}

// restart netbird
export function restart() {
  return http.post('/api/extensions/netbird/restart');
}

// stop netbird
export function stop() {
  return http.post('/api/extensions/netbird/stop');
}

// run netbird up
export function up() {
  return http.post('/api/extensions/netbird/up');
}

// run netbird down
export function down() {
  return http.post('/api/extensions/netbird/down');
}

// join with a setup key, or, without one, start an SSO login and get its URL
export function login(setupKey?: string) {
  return http.post('/api/extensions/netbird/login', setupKey ? { setupKey } : undefined, {
    timeout: 3 * 60 * 1000
  });
}

// netbird deregister: removes this peer from the account
export function logout() {
  return http.post('/api/extensions/netbird/logout');
}

// turn start at boot on or off
export function setBoot(enabled: boolean) {
  return http.post('/api/extensions/netbird/boot', { enabled });
}

// get the installed and the latest version
export function getUpdate() {
  return http.get('/api/extensions/netbird/update');
}

// install the latest version
export function update() {
  return http.post('/api/extensions/netbird/update', undefined, LONG);
}
