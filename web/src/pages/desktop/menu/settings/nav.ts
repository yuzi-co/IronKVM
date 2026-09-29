// The settings sidebar: which group each tab sits in, and which tab the modal
// opens on. Pure apart from the storage it is handed, so the rules can be
// tested without a browser.

export const GROUPS = ['system', 'network', 'access', 'integrations', 'boot', 'browser'] as const;
// 'footer' tabs sit below the groups, apart from them, so groupTabs leaves
// them out.
export type Group = (typeof GROUPS)[number] | 'footer';

export type NavTab = { id: string; group: Group };

// Tabs that were renamed. A remembered id from before the rename still opens
// the same page.
const RENAMED: Record<string, string> = { appearance: 'preferences' };

export function currentId(id: string | null): string | null {
  return id === null ? null : (RENAMED[id] ?? id);
}

// Search words for each tab, beyond its own name: the things a user may look
// for that the tab holds. Most are technical terms that read the same in
// every language, so the list is not translated.
export const KEYWORDS: Record<string, string[]> = {
  device: [
    'hdmi',
    'video',
    'usb',
    'hid',
    'reset hid',
    'virtual',
    'cdrom',
    'disk',
    'audio',
    'serial',
    'console',
    'usb network',
    'ncm',
    'rndis',
    'oled',
    'display',
    'power led',
    'reboot'
  ],
  performance: ['cpu', 'frequency', 'swap', 'zram', 'memory', 'ram'],
  watchdog: ['health', 'ping', 'hang', 'heartbeat'],
  update: ['upgrade', 'firmware', 'version', 'addon', 'offline'],
  network: ['ip', 'dhcp', 'dns', 'wifi', 'wi-fi', 'ethernet', 'hostname', 'mdns', 'gateway'],
  vpn: ['tailscale', 'netbird', 'wireguard'],
  account: ['user', 'password', 'login', 'logout'],
  apiKeys: ['api', 'token', 'key'],
  ssh: ['shell', 'authorized keys', 'host keys'],
  vnc: ['rfb'],
  tls: ['https', 'certificate', 'ssl'],
  ipmi: ['bmc', 'ipmitool'],
  redfish: ['bmc'],
  mcp: ['ai', 'agent', 'llm'],
  netboot: ['pxe', 'ipxe', 'tftp', 'boot'],
  preferences: ['language', 'menu', 'icons', 'title', 'leader key', 'caps lock', 'led'],
  about: ['version', 'hostname', 'information', 'community']
};

export type SearchTab = { id: string; label: string };

// filterTabs returns the tabs whose name, id or search words contain every
// word of the query, in the order given. An empty query matches everything.
export function filterTabs<T extends SearchTab>(tabs: T[], query: string): T[] {
  const words = query.toLocaleLowerCase().split(/\s+/).filter(Boolean);
  if (words.length === 0) return tabs;

  return tabs.filter((tab) => {
    const haystack = [tab.label, tab.id, ...(KEYWORDS[tab.id] ?? [])]
      .join('\n')
      .toLocaleLowerCase();
    return words.every((word) => haystack.includes(word));
  });
}

// groupTabs returns the groups that have at least one tab, in sidebar order,
// each with its tabs in the order they were given.
export function groupTabs<T extends NavTab>(tabs: T[]): { group: Group; tabs: T[] }[] {
  return GROUPS.map((group) => ({ group, tabs: tabs.filter((tab) => tab.group === group) })).filter(
    (entry) => entry.tabs.length > 0
  );
}

// initialTab picks the tab the modal opens on. A pending update wins, since
// the badge that brought the user here is about it. Then the tab they last
// used, if this account can still see it, and otherwise the first one.
export function initialTab(
  ids: string[],
  remembered: string | null,
  updateAvailable: boolean
): string {
  const stored = currentId(remembered);
  if (updateAvailable && ids.includes('update')) return 'update';
  if (stored && ids.includes(stored)) return stored;
  return ids[0] ?? '';
}

type StorageLike = Pick<Storage, 'getItem' | 'setItem'>;

// Storage can be missing or throw (private windows, blocked site data). The
// remembered tab is a convenience, so any failure just means nothing stored.
export function readStored(storage: StorageLike | undefined, key: string): string | null {
  try {
    return storage?.getItem(key) ?? null;
  } catch {
    return null;
  }
}

export function writeStored(storage: StorageLike | undefined, key: string, value: string) {
  try {
    storage?.setItem(key, value);
  } catch {
    // Nothing to do: the next visit opens on the default tab.
  }
}

// browserStorage is localStorage when the page may use it.
export function browserStorage(): StorageLike | undefined {
  try {
    return typeof localStorage === 'undefined' ? undefined : localStorage;
  } catch {
    return undefined;
  }
}

export const LAST_TAB_KEY = 'nano-kvm-settings-tab';
export const VPN_PROVIDER_KEY = 'nano-kvm-settings-vpn';
