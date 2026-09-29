import { migrateMenuDisabledItems } from './menu-items.ts';

const LANGUAGE_KEY = 'nano-kvm-language';
const VIDEO_MODE_KEY = 'nano-kvm-vide-mode';
const VIDEO_SCALE_KEY = 'nano-kvm-video-scale';
const MOUSE_STYLE_KEY = 'nano-kvm-mouse-style';
const MOUSE_MODE_KEY = 'nano-kvm-mouse-mode';
const MOUSE_SCROLL_DIRECTION_KEY = 'nano-kvm-mouse-scroll-direction';
const MOUSE_SCROLL_INTERVAL_KEY = 'nano-kvm-mouse-scroll-interval';
const SKIP_UPDATE_KEY = 'nano-kvm-check-update';
const KEYBOARD_SYSTEM_KEY = 'nano-kvm-keyboard-system';
const KEYBOARD_LANGUAGE_KEY = 'nano-kvm-keyboard-language';
const PASTE_LAYOUT_KEY = 'nano-kvm-paste-layout';
const PASTE_DELAY_KEY = 'nano-kvm-paste-delay';
const SKIP_MODIFY_PASSWORD_KEY = 'nano-kvm-skip-modify-password';
const MENU_DISABLED_ITEMS_KEY = 'nano-kvm-menu-disabled-items';
const MENU_AUTO_HIDE_KEY = 'nano-kvm-menu-auto-hide';
const KEYBOARD_LED_STATUS_VISIBLE_KEY = 'nano-kvm-keyboard-led-status-visible';
const POWER_CONFIRM_KEY = 'nano-kvm-power-confirm';

type ItemWithExpiry = {
  value: string;
  expiry: number;
};

// Every stored value that has to be decoded goes through here.
//
// A value we cannot read is a value we no longer have. Throwing instead of
// saying so used to take the whole page down: the exception escaped whichever
// component read the entry, and because the entry survived the reload it took
// the page down again on every refresh. A single unreadable setting could shut
// an operator out of the desktop for good, with no way back that did not
// involve the browser's developer tools.
//
// decode is expected to throw on anything it will not vouch for, shape
// included. JSON.parse accepts plenty of well-formed values that are not the
// value we stored, and handing one of those back is the same defect arriving
// one function later.
function readItem<T>(key: string, decode: (raw: string) => T): T | null {
  const raw = localStorage.getItem(key);
  if (raw === null) return null;

  try {
    return decode(raw);
  } catch (error) {
    console.warn(`[nanokvm] discarding unreadable ${key}`, error);
    localStorage.removeItem(key);
    return null;
  }
}

// set the value with expiration time (unit: milliseconds)
function setWithExpiry(key: string, value: string, ttl: number) {
  const now = new Date();

  const item: ItemWithExpiry = {
    value: value,
    expiry: now.getTime() + ttl
  };

  localStorage.setItem(key, JSON.stringify(item));
}

// get the value with expiration time
function getWithExpiry(key: string) {
  const item = readItem(key, (raw) => {
    const parsed = JSON.parse(raw) as ItemWithExpiry;
    if (typeof parsed?.value !== 'string' || typeof parsed?.expiry !== 'number') {
      throw new Error('not an expiring item');
    }
    return parsed;
  });
  if (!item) return null;

  const now = new Date();
  if (now.getTime() > item.expiry) {
    localStorage.removeItem(key);
    return null;
  }

  return item.value;
}

export function getLanguage() {
  return localStorage.getItem(LANGUAGE_KEY);
}

export function setLanguage(language: string) {
  localStorage.setItem(LANGUAGE_KEY, language);
}

export function getVideoMode() {
  return localStorage.getItem(VIDEO_MODE_KEY);
}

export function setVideoMode(mode: string) {
  localStorage.setItem(VIDEO_MODE_KEY, mode);
}

export function getVideoScale(): number | null {
  const scale = localStorage.getItem(VIDEO_SCALE_KEY);
  if (scale && Number(scale)) {
    return Number(scale);
  }
  return null;
}

export function setVideoScale(scale: number): void {
  localStorage.setItem(VIDEO_SCALE_KEY, String(scale));
}

export function getMouseStyle() {
  return localStorage.getItem(MOUSE_STYLE_KEY);
}

export function setMouseStyle(mouse: string) {
  localStorage.setItem(MOUSE_STYLE_KEY, mouse);
}

export function getMouseMode() {
  return localStorage.getItem(MOUSE_MODE_KEY);
}

export function setMouseMode(mouse: string) {
  localStorage.setItem(MOUSE_MODE_KEY, mouse);
}

export function getMouseScrollDirection(): number | null {
  const direction = localStorage.getItem(MOUSE_SCROLL_DIRECTION_KEY);
  if (direction && Number(direction)) {
    return Number(direction);
  }
  return null;
}

export function setMouseScrollDirection(direction: number): void {
  localStorage.setItem(MOUSE_SCROLL_DIRECTION_KEY, String(direction));
}

export function getMouseScrollInterval() {
  const interval = localStorage.getItem(MOUSE_SCROLL_INTERVAL_KEY);
  return interval ? Number(interval) : null;
}

export function setMouseScrollInterval(interval: number): void {
  localStorage.setItem(MOUSE_SCROLL_INTERVAL_KEY, String(interval));
}

export function getSkipUpdate() {
  const skip = getWithExpiry(SKIP_UPDATE_KEY);
  return skip === 'true';
}

export function setSkipUpdate(skip: boolean) {
  const expiry = 3 * 24 * 60 * 60 * 1000; // 3 days
  setWithExpiry(SKIP_UPDATE_KEY, String(skip), expiry);
}

export function setKeyboardSystem(system: string) {
  localStorage.setItem(KEYBOARD_SYSTEM_KEY, system);
}

export function getKeyboardSystem() {
  return localStorage.getItem(KEYBOARD_SYSTEM_KEY);
}

export function setKeyboardLanguage(language: string) {
  localStorage.setItem(KEYBOARD_LANGUAGE_KEY, language);
}

export function getKeyboardLanguage() {
  return localStorage.getItem(KEYBOARD_LANGUAGE_KEY);
}

export function setPasteLayout(layout: string) {
  localStorage.setItem(PASTE_LAYOUT_KEY, layout);
}

export function getPasteLayout() {
  return localStorage.getItem(PASTE_LAYOUT_KEY);
}

export function setPasteDelay(delay: number) {
  localStorage.setItem(PASTE_DELAY_KEY, String(delay));
}

export function getPasteDelay() {
  const delay = Number(localStorage.getItem(PASTE_DELAY_KEY));
  return Number.isFinite(delay) && delay > 0 ? delay : null;
}

export function setSkipModifyPassword(skip: boolean) {
  const expiry = 3 * 24 * 60 * 60 * 1000; // 3 days
  setWithExpiry(SKIP_MODIFY_PASSWORD_KEY, String(skip), expiry);
}

export function getSkipModifyPassword() {
  const skip = getWithExpiry(SKIP_MODIFY_PASSWORD_KEY);
  return skip === 'true';
}

export function setMenuDisabledItems(items: string[]) {
  const value = JSON.stringify(items);
  localStorage.setItem(MENU_DISABLED_ITEMS_KEY, value);
}

export function getMenuDisabledItems(): string[] {
  const items = readItem(MENU_DISABLED_ITEMS_KEY, (raw) => {
    const parsed = JSON.parse(raw) as unknown;
    if (!Array.isArray(parsed) || parsed.some((item) => typeof item !== 'string')) {
      throw new Error('not a list of menu items');
    }
    return parsed as string[];
  });
  if (!items) return [];

  // A list saved before the menu bar was regrouped names buttons that no
  // longer exist. It is rewritten once, so the switches read it as saved.
  const migrated = migrateMenuDisabledItems(items);
  if (migrated !== items) setMenuDisabledItems(migrated);

  return migrated;
}

export function getMenuDisplayMode(): string {
  const value = localStorage.getItem(MENU_AUTO_HIDE_KEY);
  return value || 'auto';
}

export function setMenuDisplayMode(mode: string) {
  localStorage.setItem(MENU_AUTO_HIDE_KEY, mode);
}

export function getKeyboardLedStatusVisible(): boolean {
  const value = localStorage.getItem(KEYBOARD_LED_STATUS_VISIBLE_KEY);
  return value !== 'false';
}

export function setKeyboardLedStatusVisible(visible: boolean) {
  localStorage.setItem(KEYBOARD_LED_STATUS_VISIBLE_KEY, String(visible));
}

// Power buttons ask before they act unless the operator turned that off. No
// stored value means the operator never chose, and a stray click must not
// switch a host off, so the default is to ask.
export function getPowerConfirm() {
  const enabled = localStorage.getItem(POWER_CONFIRM_KEY);
  return enabled !== 'false';
}

export function setPowerConfirm(enabled: boolean) {
  localStorage.setItem(POWER_CONFIRM_KEY, String(enabled));
}

// The serial terminal's last settings, so the dialog opens with the port and
// line settings that worked last time. Storage that is blocked or holds
// something else just means the defaults.
const SERIAL_SETTINGS_KEY = 'nano-kvm-serial-settings';

export type SerialSettings = {
  port: string;
  baudrate: number;
  parity: string;
  flowControl: string;
  dataBits: number;
  stopBits: number;
};

export function getSerialSettings(): Partial<SerialSettings> | null {
  try {
    return readItem(SERIAL_SETTINGS_KEY, (raw) => {
      const parsed = JSON.parse(raw);
      if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
        throw new Error('not a settings object');
      }
      return parsed as Partial<SerialSettings>;
    });
  } catch {
    return null;
  }
}

export function setSerialSettings(settings: SerialSettings) {
  try {
    localStorage.setItem(SERIAL_SETTINGS_KEY, JSON.stringify(settings));
  } catch {
    // Not remembered; the dialog still works.
  }
}

// The language the text recognition read last, so the next read uses it.
const OCR_LANGUAGE_KEY = 'nano-kvm-ocr-language';

export function getOcrLanguage(): string | null {
  try {
    return localStorage.getItem(OCR_LANGUAGE_KEY);
  } catch {
    return null;
  }
}

export function setOcrLanguage(language: string) {
  try {
    localStorage.setItem(OCR_LANGUAGE_KEY, language);
  } catch {
    // Not remembered; the next read starts from the default.
  }
}
