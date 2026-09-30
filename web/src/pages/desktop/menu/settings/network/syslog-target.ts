// Why a remote syslog target was refused. Each maps to a message under
// settings.network.syslog.errors.
export type SyslogTargetError = 'empty' | 'long' | 'ipv6' | 'brackets' | 'host' | 'port';

// A DNS name is at most 253 characters; brackets and a port add a few more.
const maxLength = 262;

// syslogTargetError mirrors the server's validateSyslogTarget, which mirrors
// valid_target in the image's S01syslogd: a host name, an IPv4 address or a
// bracketed IPv6 address, with an optional port from 1 to 65535. It returns
// null for a target the device will accept.
export function syslogTargetError(target: string): SyslogTargetError | null {
  if (target === '') return 'empty';
  if (target.length > maxLength) return 'long';

  let host: string;
  let port: string | null = null;
  const bracketEnd = target.indexOf(']:');

  if (target.startsWith('[') && bracketEnd !== -1) {
    host = target.slice(0, bracketEnd + 1);
    port = target.slice(bracketEnd + 2);
  } else if (target.startsWith('[') && target.endsWith(']')) {
    host = target;
  } else if (target.split(':').length > 2) {
    return 'ipv6';
  } else if (target.includes(':')) {
    const i = target.lastIndexOf(':');
    host = target.slice(0, i);
    port = target.slice(i + 1);
  } else {
    host = target;
  }

  if (host.startsWith('[') && host.endsWith(']')) {
    if (!/^[0-9A-Fa-f:.]+$/.test(host.slice(1, -1))) return 'brackets';
  } else if (!/^[A-Za-z0-9][A-Za-z0-9.-]*$/.test(host)) {
    return 'host';
  }

  if (port === null) return null;
  if (!/^\d{1,5}$/.test(port)) return 'port';
  const number = Number(port);
  return number >= 1 && number <= 65535 ? null : 'port';
}
