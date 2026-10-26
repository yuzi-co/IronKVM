// The STUN server field (ironkvm-dist#157). The server checks again; this
// only keeps the Apply button off for a value it would refuse.

// STUN_DISABLED is what the server saves for STUN turned off.
export const STUN_DISABLED = 'disable';

const hostNameRe =
  /^([a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?$/;
const ipv6Re = /^[0-9a-fA-F:.]+$/;

// isStunServer is true for host:port with a port from 1 to 65535, where the
// host is a host name, an IPv4 address or a bracketed IPv6 address.
export function isStunServer(value: string): boolean {
  const text = value.trim();
  if (text.length > 260 || /[\s@/]/.test(text)) return false;

  const match = /^(\[([^\]]+)\]|[^:[\]]+):(\d{1,5})$/.exec(text);
  if (!match) return false;

  const port = Number(match[3]);
  if (port < 1 || port > 65535) return false;

  if (match[2] !== undefined) return match[2].includes(':') && ipv6Re.test(match[2]);
  return hostNameRe.test(match[1]);
}

// stunToSave is what Apply sends: "disable" when the switch is off, else the
// field.
export function stunToSave(enabled: boolean, value: string): string {
  return enabled ? value.trim() : STUN_DISABLED;
}
