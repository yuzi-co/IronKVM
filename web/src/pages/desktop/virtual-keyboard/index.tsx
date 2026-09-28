import { useRef, useState } from 'react';
import { AppleOutlined, WindowsOutlined } from '@ant-design/icons';
import clsx from 'clsx';
import { useAtom } from 'jotai';
import { XIcon } from 'lucide-react';
import Keyboard, { KeyboardButtonTheme } from 'react-simple-keyboard';
import { Drawer } from 'vaul';

import 'react-simple-keyboard/build/css/index.css';
import '@/assets/styles/keyboard.css';

import { Segmented, Select } from 'antd';
import { useTranslation } from 'react-i18next';
import { useMediaQuery } from 'react-responsive';

import { getKeycode, getModifierBit } from '@/lib/keymap.ts';
import * as storage from '@/lib/localstorage.ts';
import { client, MessageEvent } from '@/lib/websocket.ts';
import { isKeyboardOpenAtom } from '@/jotai/keyboard.ts';

import {
  compactDisplay,
  doubleKeys,
  keyboardArrowsOptions,
  keyboardCompactPadOptions,
  keyboardControlPadOptions,
  keyboardOptions,
  modifierKeys,
  specialKeyMap
} from './virtual-keys.ts';

const languages = [
  { value: 'en', label: 'English' },
  { value: 'fr', label: 'French' },
  { value: 'de', label: 'German' },
  { value: 'ru', label: 'Russian' },
  { value: 'ko', label: 'Korean' },
  { value: 'ja', label: 'Japanese' }
];

// The layout names are shown in the UI language. The browser knows every
// language's name in every other, so this needs no strings of our own; the
// English label stays as the fallback for a browser without Intl.DisplayNames.
function localizedLanguages(uiLanguage: string) {
  let names: Intl.DisplayNames | undefined;
  try {
    names = new Intl.DisplayNames([uiLanguage, 'en'], { type: 'language' });
  } catch {
    names = undefined;
  }

  return languages.map((lng) => ({ ...lng, label: names?.of(lng.value) ?? lng.label }));
}

const layoutByLanguage = new Map([
  ['en', 'default'],
  ['ru', 'rus'],
  ['de', 'qwertz'],
  ['fr', 'azerty'],
  ['ko', 'ko'],
  ['ja', 'ja']
]);

// The layout follows from the language and the system. Nothing else sets it, so
// it is derived during render rather than stored and corrected by an effect.
function layoutFor(system: string, language: string) {
  if (language === 'en' && system === 'mac') {
    return 'mac';
  }
  return layoutByLanguage.get(language) ?? 'default';
}

export const VirtualKeyboard = () => {
  const { t, i18n } = useTranslation();
  const isBigScreen = useMediaQuery({ minWidth: 850 });

  const [isKeyboardOpen, setIsKeyboardOpen] = useAtom(isKeyboardOpenAtom);

  // Read on the first render. Set from an effect, the stored language never
  // reached the Select below, because its `defaultValue` only applies on mount.
  const [keyboardSystem, setKeyboardSystem] = useState(() => {
    const stored = storage.getKeyboardSystem();
    return stored && ['win', 'mac'].includes(stored) ? stored : 'win';
  });
  const [keyboardLanguage, setKeyboardLanguage] = useState(() => {
    const stored = storage.getKeyboardLanguage();
    return stored && languages.some((lng) => lng.value === stored) ? stored : 'en';
  });
  const keyboardLayout = layoutFor(keyboardSystem, keyboardLanguage);
  const [activeModifierKeys, setActiveModifierKeys] = useState<string[]>([]);

  const keyboardRef = useRef<any>(null);

  const systems = [
    { value: 'win', icon: <WindowsOutlined /> },
    { value: 'mac', icon: <AppleOutlined /> }
  ];

  // Press key
  function onKeyPress(key: string) {
    if (modifierKeys.includes(key)) {
      if (activeModifierKeys.includes(key)) {
        sendModifierKeyDown();
        sendModifierKeyUp();
      } else {
        setActiveModifierKeys([...activeModifierKeys, key]);
      }
      return;
    }

    sendKeydown(key);
  }

  // Release key
  function onKeyReleased(key: string) {
    if (modifierKeys.includes(key)) {
      return;
    }

    sendKeyup();
  }

  // Send all keys
  function sendKeydown(key: string) {
    const code = getKeyboardCode(key);
    if (!code) {
      console.log('unknown code: ', key);
      return;
    }

    const modifier = sendModifierKeyDown();

    send(modifier, code);
  }

  function getKeyboardCode(key: string) {
    // AZERTY: swap A↔Q and Z↔W on French physical positions
    if (keyboardLanguage === 'fr' && key.endsWith('_azerty')) {
      const base = key.replace('_azerty', '');
      if (base === 'KeyA') return getKeycode('KeyQ');
      if (base === 'KeyQ') return getKeycode('KeyA');
      if (base === 'KeyZ') return getKeycode('KeyW');
      if (base === 'KeyW') return getKeycode('KeyZ');
      // all other labels use their own code
      return getKeycode(base);
    }

    if (keyboardLanguage === 'de' && key.endsWith('_qwertz')) {
      const base = key.replace('_qwertz', '');
      // Tausch
      if (base === 'KeyZ') return getKeycode('KeyY');
      if (base === 'KeyY') return getKeycode('KeyZ');
      // all other labels use their own code
      return getKeycode(base);
    }

    if (keyboardLanguage === 'ko' && key.endsWith('_ko')) {
      const base = key.replace('_ko', '');
      return getKeycode(base);
    }

    if (keyboardLanguage === 'ja' && key.endsWith('_ja')) {
      const base = key.replace('_ja', '');

      // The Backquote position on a JIS keyboard is the Zenkaku/Hankaku key,
      // which toggles the IME. Sending the plain Grave usage (0x35) toggles it
      // only on a host whose active layout really is JIS 106/109; Lang5 (0x94)
      // is the layout-independent USB HID usage for the same toggle, so it
      // works wherever the host has a Japanese IME at all.
      if (base === 'Backquote') return getKeycode('Lang5');

      return getKeycode(base);
    }

    const specialKey = specialKeyMap.get(key);
    if (specialKey) {
      return getKeycode(specialKey);
    }

    return getKeycode(key);
  }

  // Release all keys
  function sendKeyup() {
    sendModifierKeyUp();
    send(0, 0);
  }

  // Send modifier keys
  function sendModifierKeyDown() {
    let modifier = 0;

    activeModifierKeys.forEach((modifierKey) => {
      const key = specialKeyMap.get(modifierKey)!;

      modifier |= getModifierBit(key)!;
      const code = getKeycode(key)!;

      send(modifier, code);
    });

    return modifier;
  }

  // Release modifier keys
  function sendModifierKeyUp() {
    if (activeModifierKeys.length === 0) return;

    activeModifierKeys.forEach(() => {
      send(0, 0);
    });

    setActiveModifierKeys([]);
  }

  function send(modifier: number, code: number) {
    const data = new Uint8Array([MessageEvent.Keyboard, modifier, 0, code, 0, 0, 0, 0, 0]);
    client.send(data);
  }

  function selectSystem(system: string) {
    setKeyboardSystem(system);
    storage.setKeyboardSystem(system);
  }

  function selectLanguage(language: string) {
    setKeyboardLanguage(language);
    storage.setKeyboardLanguage(language);
  }

  function getButtonTheme(): KeyboardButtonTheme[] {
    const theme = [{ class: 'hg-double', buttons: doubleKeys.join(' ') }];

    if (activeModifierKeys.length > 0) {
      const buttons = activeModifierKeys.join(' ');
      theme.push({ class: 'hg-highlight', buttons });
    }

    return theme;
  }

  return (
    <Drawer.Root open={isKeyboardOpen} onOpenChange={setIsKeyboardOpen} modal={false}>
      <Drawer.Portal>
        <Drawer.Content
          className={clsx(
            'fixed right-0 bottom-0 left-0 z-999 mx-auto overflow-hidden rounded bg-neutral-900 outline-hidden',
            isBigScreen ? 'w-[820px]' : 'w-full max-w-[650px]'
          )}
        >
          {/* header */}
          <div className="flex items-center justify-between px-3 py-1">
            <div className="flex items-center space-x-5">
              <Select
                size="small"
                style={{ minWidth: 90 }}
                defaultValue={keyboardLanguage}
                options={localizedLanguages(i18n.language)}
                onChange={selectLanguage}
              />

              {keyboardLanguage === 'en' && (
                <Segmented
                  size="small"
                  options={systems}
                  value={keyboardSystem}
                  onChange={selectSystem}
                />
              )}
            </div>

            <div className="flex items-center justify-end">
              <button
                type="button"
                aria-label={t('keyboard.close')}
                className="flex h-[24px] w-[24px] cursor-pointer items-center justify-center rounded p-0 text-neutral-400 hover:bg-neutral-700 hover:text-white"
                onClick={() => setIsKeyboardOpen(false)}
              >
                <XIcon size={18} />
              </button>
            </div>
          </div>

          <div className="h-px shrink-0 bg-neutral-700" />

          <div
            data-vaul-no-drag
            className={clsx('keyboardContainer w-full', !isBigScreen && 'keyboard-compact')}
          >
            {/* main keyboard */}
            <Keyboard
              buttonTheme={getButtonTheme()}
              keyboardRef={(r) => (keyboardRef.current = r)}
              onKeyPress={onKeyPress}
              onKeyReleased={onKeyReleased}
              layoutName={keyboardLayout}
              {...keyboardOptions}
              display={
                isBigScreen
                  ? keyboardOptions.display
                  : { ...keyboardOptions.display, ...compactDisplay }
              }
            />

            {/* control keyboard */}
            {isBigScreen ? (
              <div className="controlArrows">
                <Keyboard
                  onKeyPress={onKeyPress}
                  onKeyReleased={onKeyReleased}
                  {...keyboardControlPadOptions}
                />

                <Keyboard
                  onKeyPress={onKeyPress}
                  onKeyReleased={onKeyReleased}
                  {...keyboardArrowsOptions}
                />
              </div>
            ) : (
              <Keyboard
                onKeyPress={onKeyPress}
                onKeyReleased={onKeyReleased}
                {...keyboardCompactPadOptions}
              />
            )}
          </div>
        </Drawer.Content>
        <Drawer.Overlay />
      </Drawer.Portal>
    </Drawer.Root>
  );
};
