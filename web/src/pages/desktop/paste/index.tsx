import { useEffect } from 'react';
import { useAtom } from 'jotai';

import { getPasteStatus } from '@/api/hid.ts';
import { pasteStatusAtom } from '@/jotai/paste.ts';

import { PasteDialog } from './dialog.tsx';
import { PasteProgress } from './progress.tsx';

// How often the progress of a paste is fetched while it types.
const pollIntervalMs = 500;

// Paste holds the paste dialog and the progress of a paste typing on the host.
// Both live here rather than in the keyboard menu, because the shortcut opens
// the dialog and starts pastes with the menu closed.
export const Paste = () => {
  const [status, setStatus] = useAtom(pasteStatusAtom);
  const isTyping = status?.status === 'typing';

  // A paste started before a reload, or in another tab, is still typing on the
  // host. Picking it up gives this page its progress and its cancel.
  useEffect(() => {
    getPasteStatus()
      .then((rsp) => {
        if (rsp.code === 0 && rsp.data?.status === 'typing') {
          setStatus(rsp.data);
        }
      })
      .catch(() => {});
  }, [setStatus]);

  useEffect(() => {
    if (!isTyping) return;

    const timer = setInterval(() => {
      getPasteStatus()
        .then((rsp) => {
          if (rsp.code === 0 && rsp.data) {
            setStatus(rsp.data);
          }
        })
        .catch(() => {});
    }, pollIntervalMs);

    return () => clearInterval(timer);
  }, [isTyping, setStatus]);

  return (
    <>
      <PasteDialog />
      <PasteProgress />
    </>
  );
};
