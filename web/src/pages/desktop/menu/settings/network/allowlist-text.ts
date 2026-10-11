// The allowed networks field is one text box: entries are separated by new
// lines, commas or spaces. The server checks each entry; this only splits.

export function parseNetworks(text: string): string[] {
  return text
    .split(/[\s,]+/)
    .map((entry) => entry.trim())
    .filter((entry) => entry !== '');
}

export function formatNetworks(networks: string[]): string {
  return networks.join('\n');
}

// sameNetworks says whether the text holds the saved list, in the same order.
export function sameNetworks(text: string, saved: string[]): boolean {
  const entries = parseNetworks(text);
  return entries.length === saved.length && entries.every((entry, i) => entry === saved[i]);
}

// withClient adds the caller's own address to the text, unless it is already
// an entry of its own.
export function withClient(text: string, client: string): string {
  if (!client || parseNetworks(text).includes(client)) return text;
  const trimmed = text.trimEnd();
  return trimmed === '' ? client : `${trimmed}\n${client}`;
}
