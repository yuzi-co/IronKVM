// The toolbar's alert icon and status dots, as pure rules over what the board
// reports, so the thresholds can be tested without a browser.

// What GET /api/vm/health answers (GetHealthRsp in server/proto/vm.go).
export type Health = {
  temperature: number | null;
  storage: { path: string; total: number; available: number } | null;
  vpn: { name: string; title: string; running: boolean }[];
};

export type Severity = 'warning' | 'critical';

export type Alert =
  | { kind: 'temperature'; severity: Severity; celsius: number }
  | { kind: 'storage'; severity: Severity; path: string; available: number; total: number }
  | { kind: 'vpn'; severity: Severity; title: string }
  | { kind: 'stream'; severity: Severity };

// SoC temperature, degrees C. The SG2002 has no cpufreq driver and nothing on
// the board throttles it, so heat is only ever reported, never acted on. The
// limits are deliberately conservative guesses rather than datasheet values:
// 80 means the case or the air around it is holding the heat in, and 90 is
// hot enough that the owner should act now.
export const TEMPERATURE_WARNING_C = 80;
export const TEMPERATURE_CRITICAL_C = 90;

const MiB = 1024 * 1024;
const GiB = 1024 * MiB;

// Space left for images. A warning needs both a small share and a small
// amount: 10% of a 256 GB card is still room for several installers, and a
// 2 GB partition is never above 2 GB free. Below 256 MiB an upload, a
// download or an add-on install fails outright, whatever the size.
export const STORAGE_WARNING_FRACTION = 0.1;
export const STORAGE_WARNING_BYTES = 2 * GiB;
export const STORAGE_CRITICAL_BYTES = 256 * MiB;

export function temperatureAlert(celsius: number | null): Alert | null {
  if (celsius === null || !Number.isFinite(celsius)) return null;
  if (celsius >= TEMPERATURE_CRITICAL_C)
    return { kind: 'temperature', severity: 'critical', celsius };
  if (celsius >= TEMPERATURE_WARNING_C)
    return { kind: 'temperature', severity: 'warning', celsius };
  return null;
}

export function storageAlert(storage: Health['storage']): Alert | null {
  if (!storage || storage.total <= 0) return null;

  const { path, available, total } = storage;
  if (available < STORAGE_CRITICAL_BYTES) {
    return { kind: 'storage', severity: 'critical', path, available, total };
  }
  if (available < total * STORAGE_WARNING_FRACTION && available < STORAGE_WARNING_BYTES) {
    return { kind: 'storage', severity: 'warning', path, available, total };
  }
  return null;
}

// A VPN the owner set to start at boot is expected up; the server lists only
// those. One that is merely installed, or stopped on purpose, is not listed.
export function vpnAlerts(vpn: Health['vpn']): Alert[] {
  return vpn
    .filter((entry) => !entry.running)
    .map((entry) => ({ kind: 'vpn', severity: 'warning', title: entry.title }) as Alert);
}

// The picture as the Screen dot shows it. Unknown draws no dot: nothing has
// said yet whether the stream works.
export type StreamState = 'ok' | 'noSignal' | 'failed' | 'unknown';

// The capture status the board last sent for this viewer's video mode, as
// pages/desktop/capture-status keeps it: null once the stream is fine again.
export type CaptureReport = { result: number; severity: 'error' | 'warning' | '' } | null;

// Result -1 is "no image captured": the host is off, asleep or unplugged,
// which is a fact about the host rather than a fault. The warnings (-4, -5)
// pass on their own within a second or two, so they say nothing either way.
export const CAPTURE_NO_SIGNAL = -1;

export function streamState(report: CaptureReport, pictureShown: boolean): StreamState {
  if (report) {
    if (report.result === CAPTURE_NO_SIGNAL) return 'noSignal';
    if (report.severity === 'error') return 'failed';
    return 'unknown';
  }
  return pictureShown ? 'ok' : 'unknown';
}

export function streamAlert(state: StreamState): Alert | null {
  return state === 'failed' ? { kind: 'stream', severity: 'critical' } : null;
}

// healthAlerts lists what is wrong, worst first. An empty list hides the icon.
export function healthAlerts(health: Health | null, stream: StreamState): Alert[] {
  const alerts: Alert[] = [];
  const add = (alert: Alert | null) => {
    if (alert) alerts.push(alert);
  };

  add(streamAlert(stream));
  if (health) {
    add(temperatureAlert(health.temperature));
    add(storageAlert(health.storage));
    alerts.push(...vpnAlerts(health.vpn ?? []));
  }

  const rank = (severity: Severity) => (severity === 'critical' ? 0 : 1);
  return alerts.sort((a, b) => rank(a.severity) - rank(b.severity));
}

// The worst severity in a list, or null when the list is empty.
export function worstSeverity(alerts: Alert[]): Severity | null {
  if (alerts.length === 0) return null;
  return alerts.some((alert) => alert.severity === 'critical') ? 'critical' : 'warning';
}

// formatBytes names a size the way the alerts and media warnings show it.
export function formatBytes(bytes: number): string {
  if (bytes >= GiB) return `${(bytes / GiB).toFixed(1)} GiB`;
  return `${Math.max(0, Math.round(bytes / MiB))} MiB`;
}
