// isValidIPv4 accepts a dotted quad with no leading zeros, which some tools
// would read as octal.
export function isValidIPv4(value: string) {
  const parts = value.split('.');
  if (parts.length !== 4) return false;

  return parts.every((part) => {
    if (!/^\d+$/.test(part)) return false;
    if (part.length > 1 && part.startsWith('0')) return false;

    const number = Number(part);
    return number >= 0 && number <= 255;
  });
}

// isValidIPv6 accepts a plain IPv6 address: no zone id, no brackets, no
// prefix. The URL parser knows every legal form (compression, an embedded
// IPv4 tail), so it does the parsing and the checks here are only that it was
// the whole input.
export function isValidIPv6(value: string) {
  if (value === '' || value !== value.trim() || !value.includes(':')) return false;
  if (!/^[0-9a-fA-F:.]+$/.test(value)) return false;

  try {
    const url = new URL(`http://[${value}]/`);
    return url.hostname.startsWith('[');
  } catch {
    return false;
  }
}
