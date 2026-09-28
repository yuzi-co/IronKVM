// The settings sidebar: which group each tab sits in, and which tab the modal
// opens on. Pure apart from the storage it is handed, so the rules can be
// tested without a browser.

export const GROUPS = ['general', 'device', 'network', 'remote', 'boot'] as const;
export type Group = (typeof GROUPS)[number];

export type NavTab = { id: string; group: Group };

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
export function initialTab(ids: string[], stored: string | null, updateAvailable: boolean): string {
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
