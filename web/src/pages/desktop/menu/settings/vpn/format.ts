export const vpnTitles: Record<string, string> = {
  tailscale: 'Tailscale',
  netbird: 'NetBird'
};

export function formatBytes(bytes: number): string {
  if (!bytes || bytes <= 0) return '-';
  const mib = bytes / 1024 / 1024;
  return mib >= 1024 ? `${(mib / 1024).toFixed(1)} GiB` : `${mib.toFixed(1)} MiB`;
}

export function formatUptime(sec: number): string {
  if (!sec || sec <= 0) return '-';
  const days = Math.floor(sec / 86400);
  const hours = Math.floor((sec % 86400) / 3600);
  const minutes = Math.floor((sec % 3600) / 60);
  if (days > 0) return `${days}d ${hours}h`;
  if (hours > 0) return `${hours}h ${minutes}m`;
  return `${minutes}m`;
}
