import { useEffect, useRef } from 'react';
import { notification } from 'antd';
import { useTranslation } from 'react-i18next';

import { client } from '@/lib/websocket.ts';

const NOTIFICATION_KEY = 'input_disconnected';

// Retries before the operator is told. The client retries every three seconds,
// so this is about ten seconds of silence. A server restart reconnects well
// inside that, and nagging about every restart would teach the operator to
// ignore this notice, which is the one notice that must not be ignored.
const ATTEMPTS_BEFORE_WARNING = 4;

// InputDisconnectedWarning tells the operator that the keyboard and the mouse
// are not connected to anything.
//
// Nothing else says it. Keyboard and mouse travel over the websocket at
// /api/ws, and every other part of the page uses ordinary HTTPS requests. When
// that socket cannot open, the UI renders, the menus work, the video plays, and
// no keystroke reaches the managed host. The screen even keeps updating, so the
// board looks alive.
//
// The commonest cause is the certificate. A browser asks the operator about an
// untrusted certificate when it loads the page, and refuses a websocket to the
// same origin WITHOUT asking, reporting nothing the page can catch. So turning
// HTTPS on was enough to remove the input with no message anywhere. The server
// now generates a certificate that names the device's own addresses, which
// removes the mismatch, but a self-signed certificate still has to be trusted
// once and a browser that refuses it fails exactly this way.
export const InputDisconnectedWarning = () => {
  const { t } = useTranslation();
  const [api, contextHolder] = notification.useNotification();
  // A ref, not state. The subscription must be made once: re-running the effect
  // on every open and close would unsubscribe and resubscribe, and onStatus
  // fires on subscribe, so the notice would rebuild itself in a loop.
  const isOpen = useRef(false);

  useEffect(() => {
    return client.onStatus((status) => {
      const stuck = !status.connected && status.attempts >= ATTEMPTS_BEFORE_WARNING;

      if (!stuck) {
        if (isOpen.current) {
          api.destroy(NOTIFICATION_KEY);
          isOpen.current = false;
        }
        return;
      }

      if (isOpen.current) return;
      isOpen.current = true;

      // A socket that has never opened points somewhere different from one that
      // dropped. The first is refused before it starts, which over HTTPS is
      // almost always the certificate. The second usually means the server went
      // away, and it comes back on its own.
      const description = status.everConnected
        ? t('input.disconnectedDropped')
        : window.location.protocol === 'https:'
          ? t('input.disconnectedTls')
          : t('input.disconnectedNever');

      api.error({
        key: NOTIFICATION_KEY,
        message: t('input.disconnected'),
        description,
        placement: 'topRight',
        duration: null,
        onClose: () => {
          isOpen.current = false;
        }
      });
    });
  }, [api, t]);

  return <>{contextHolder}</>;
};
