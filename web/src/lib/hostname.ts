// One label of an RFC 1123 host name: letters, digits and hyphens, neither
// starting nor ending with a hyphen, at most 63 characters.
const label = /^[A-Za-z0-9]([A-Za-z0-9-]{0,61}[A-Za-z0-9])?$/;

// isValidHostname mirrors validHostname in server/service/vm/hostname.go: dot
// separated labels and 64 characters at most, the kernel's limit.
export function isValidHostname(name: string): boolean {
  if (name === '' || name.length > 64) return false;
  return name.split('.').every((part) => label.test(part));
}
