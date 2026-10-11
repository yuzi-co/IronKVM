// The checks the page makes on a login server before it asks the device. The
// server checks again (service/extensions/tailscale/loginserver.go) and its
// answer is the one that counts; this only saves a round trip for a typo.

// isLoginServer reports whether value can be sent: empty, for Tailscale's own
// server, or an https URL with a host and no credentials, query or fragment.
export function isLoginServer(value: string): boolean {
  const s = value.trim();
  if (s === '') return true;
  if (/\s/.test(s)) return false;

  let u: URL;
  try {
    u = new URL(s);
  } catch {
    return false;
  }
  return (
    u.protocol === 'https:' &&
    u.hostname !== '' &&
    u.username === '' &&
    u.password === '' &&
    u.search === '' &&
    u.hash === '' &&
    !s.includes('?') &&
    !s.includes('#')
  );
}

// sameLoginServer compares what the field holds with what the device uses,
// ignoring a trailing slash and the spaces around it.
export function sameLoginServer(a: string, b: string): boolean {
  const norm = (s: string) => s.trim().replace(/\/+$/, '');
  return norm(a) === norm(b);
}
