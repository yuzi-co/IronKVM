// parseUdpPort reads the port field: empty is 0 (off), a number from 1024 to
// 65535 is itself, and anything else is null. The server checks again, and
// also refuses the ports it listens on.
export function parseUdpPort(value: string): number | null {
  const text = value.trim();
  if (text === '') return 0;
  if (!/^\d{1,5}$/.test(text)) return null;

  const port = Number(text);
  return port >= 1024 && port <= 65535 ? port : null;
}
