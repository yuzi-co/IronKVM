import assert from 'node:assert/strict';
import { test } from 'node:test';

import {
  formatBytes,
  healthAlerts,
  STORAGE_CRITICAL_BYTES,
  storageAlert,
  streamState,
  temperatureAlert,
  vpnAlerts,
  worstSeverity,
  type Health
} from './health.ts';

const GiB = 1024 ** 3;
const MiB = 1024 ** 2;

test('temperature below 80 is fine, 80 warns, 90 is critical', () => {
  assert.equal(temperatureAlert(null), null);
  assert.equal(temperatureAlert(79.9), null);
  assert.equal(temperatureAlert(80)?.severity, 'warning');
  assert.equal(temperatureAlert(90)?.severity, 'critical');
});

test('storage warns only when both the share and the amount are small', () => {
  const store = (available: number, total: number) => ({ path: '/data', available, total });

  assert.equal(storageAlert(null), null);
  // 5% of a large card is still plenty of room.
  assert.equal(storageAlert(store(12 * GiB, 240 * GiB)), null);
  // A small partition with most of it free is fine.
  assert.equal(storageAlert(store(1.5 * GiB, 2 * GiB)), null);
  assert.equal(storageAlert(store(1 * GiB, 16 * GiB))?.severity, 'warning');
  assert.equal(storageAlert(store(STORAGE_CRITICAL_BYTES - 1, 240 * GiB))?.severity, 'critical');
  assert.equal(storageAlert(store(100 * MiB, 200 * MiB))?.severity, 'critical');
});

test('a VPN that starts at boot but is not running warns', () => {
  const alerts = vpnAlerts([
    { name: 'tailscale', title: 'Tailscale', running: false },
    { name: 'netbird', title: 'NetBird', running: true }
  ]);
  assert.deepEqual(alerts, [{ kind: 'vpn', severity: 'warning', title: 'Tailscale' }]);
});

test('stream state: no signal is not a failure, warnings are unknown', () => {
  assert.equal(streamState(null, false), 'unknown');
  assert.equal(streamState(null, true), 'ok');
  assert.equal(streamState({ result: -1, severity: 'error' }, true), 'noSignal');
  assert.equal(streamState({ result: -2, severity: 'error' }, true), 'failed');
  assert.equal(streamState({ result: -5, severity: 'warning' }, false), 'unknown');
});

test('alerts come worst first, and nothing wrong means no icon', () => {
  const health: Health = {
    temperature: 82,
    storage: { path: '/data', available: 10 * MiB, total: 8 * GiB },
    vpn: [{ name: 'netbird', title: 'NetBird', running: false }]
  };

  const alerts = healthAlerts(health, 'ok');
  assert.deepEqual(
    alerts.map((alert) => alert.kind),
    ['storage', 'temperature', 'vpn']
  );
  assert.equal(worstSeverity(alerts), 'critical');

  assert.deepEqual(healthAlerts(null, 'noSignal'), []);
  assert.deepEqual(
    healthAlerts(null, 'failed').map((alert) => alert.kind),
    ['stream']
  );
  assert.equal(worstSeverity([]), null);
});

test('sizes read in MiB below a GiB and in GiB above', () => {
  assert.equal(formatBytes(300 * MiB), '300 MiB');
  assert.equal(formatBytes(1.5 * GiB), '1.5 GiB');
});
