import { useState } from 'react';
import { message } from 'antd';
import { useTranslation } from 'react-i18next';

import { KeyboardReport } from '@/lib/keyboard.ts';
import { client, MessageEvent } from '@/lib/websocket.ts';
import { Kbd, KbdGroup } from '@/components/ui/kbd.tsx';

import type { Shortcut as ShortcutInterface } from './types.ts';

type ShortcutProps = {
  shortcut: ShortcutInterface;
};

export const Shortcut = ({ shortcut }: ShortcutProps) => {
  const { t } = useTranslation();
  const [isLoading, setIsLoading] = useState(false);

  // sendShortcut reports whether every report went out. With the input
  // connection down nothing reaches the host, and the click must say so.
  async function sendShortcut() {
    const keyboard = new KeyboardReport();

    let sent = true;
    shortcut.keys.forEach((key) => {
      const report = keyboard.keyDown(key.code);
      sent = send(report) && sent;
    });

    // The release goes out even after a failure, so no key is left held.
    const report = keyboard.reset();
    sent = send(report) && sent;
    return sent;
  }

  function send(report: Uint8Array) {
    const data = new Uint8Array([MessageEvent.Keyboard, ...report]);
    return client.send(data);
  }

  async function handleClick(): Promise<void> {
    if (isLoading) return;
    setIsLoading(true);

    try {
      if (!(await sendShortcut())) {
        message.error(t('keyboard.shortcut.sendFailed'));
      }
    } catch (err) {
      console.log(err);
      message.error(t('keyboard.shortcut.sendFailed'));
    } finally {
      setIsLoading(false);
    }
  }

  return (
    <button
      type="button"
      className="flex h-[32px] w-full cursor-pointer items-center space-x-1 rounded p-0 px-3 text-left hover:bg-neutral-700/30"
      onClick={handleClick}
    >
      {shortcut.keys.map((key, index) => (
        <KbdGroup key={index}>
          <Kbd>{key.label}</Kbd>
        </KbdGroup>
      ))}
    </button>
  );
};
