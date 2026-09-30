import { useEffect, useRef, useState } from 'react';
import { useAuth } from '@/contexts/auth.ts';
import { Button, message, notification } from 'antd';
import { useAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import { getHidStatus } from '@/api/hid.ts';
import { resetHid } from '@/lib/hid-reset.ts';
import { applyMouseMode } from '@/lib/mouse-mode.ts';
import { pollWhileVisible } from '@/lib/visible-poll.ts';
import { mouseModeAtom } from '@/jotai/mouse.ts';

import { HidDeviceStatus, isAbsoluteMouseStalled } from './model.ts';

export { InputDisconnectedWarning } from './input-disconnected.tsx';

const NOTIFICATION_KEY = 'absolute_mouse_stalled';
const POLL_INTERVAL_MS = 10_000;

// AbsoluteMouseWarning tells the operator that absolute mouse reports are going
// nowhere, and offers the mode that works.
//
// Nothing else can say this. The device node is present, the USB gadget is
// bound and configured, and the keyboard keeps working - so every other signal
// reads healthy while the pointer does not move.
export const AbsoluteMouseWarning = () => {
  const { t } = useTranslation();
  const [api, contextHolder] = notification.useNotification();
  const [mouseMode, setMouseMode] = useAtom(mouseModeAtom);
  const { account } = useAuth();
  // Resetting USB is an admin action on the server.
  const isAdmin = account.role === 'admin';
  const isOpen = useRef(false);

  useEffect(() => {
    // Only absolute and touch mode use the endpoint that stalls, so relative
    // mode has nothing to poll for and should cost the device nothing.
    if (mouseMode !== 'absolute' && mouseMode !== 'touch') {
      if (isOpen.current) {
        api.destroy(NOTIFICATION_KEY);
        isOpen.current = false;
      }
      return;
    }

    let cancelled = false;

    function check() {
      getHidStatus()
        .then((rsp) => {
          if (cancelled || rsp.code !== 0) return;

          const devices: HidDeviceStatus[] = rsp.data?.devices ?? [];
          if (isAbsoluteMouseStalled(devices)) {
            open();
          } else if (isOpen.current) {
            api.destroy(NOTIFICATION_KEY);
            isOpen.current = false;
          }
        })
        .catch(() => {
          // A failed status call is not worth telling the operator about. It
          // says nothing about the mouse, and this runs every ten seconds.
        });
    }

    function open() {
      if (isOpen.current) return;
      isOpen.current = true;

      api.warning({
        key: NOTIFICATION_KEY,
        message: t('mouse.absoluteStalled'),
        description: t('mouse.absoluteStalledDesc'),
        placement: 'topRight',
        duration: false,
        // Two remedies, because either can be the right one and the server
        // cannot tell which. Recovering USB is offered first: it keeps the
        // mouse mode the operator chose.
        btn: (
          <div className="flex gap-2">
            {isAdmin && <RecoverUsbButton onDone={close} />}
            <Button type="primary" onClick={switchToRelative}>
              {t('mouse.useRelative')}
            </Button>
          </div>
        ),
        onClose: () => {
          isOpen.current = false;
        }
      });
    }

    function switchToRelative() {
      api.destroy(NOTIFICATION_KEY);
      isOpen.current = false;
      applyMouseMode('relative', setMouseMode);
    }

    function close() {
      api.destroy(NOTIFICATION_KEY);
      isOpen.current = false;
    }

    check();
    const stopPoll = pollWhileVisible(check, POLL_INTERVAL_MS);

    return () => {
      cancelled = true;
      stopPoll();
    };
  }, [mouseMode, api, t, setMouseMode, isAdmin]);

  return <>{contextHolder}</>;
};

// RecoverUsbButton offers the recovery the mouse menu offers. It keeps its own
// state because the notification renders its buttons once: a flag read from a
// ref there never repaints, and the spinner never showed.
const RecoverUsbButton = ({ onDone }: { onDone: () => void }) => {
  const { t } = useTranslation();
  const [isRecovering, setIsRecovering] = useState(false);

  async function recover() {
    if (isRecovering) return;
    setIsRecovering(true);

    const result = await resetHid();
    setIsRecovering(false);
    if (!result.ok) {
      message.error(result.msg || t('mouse.resetHidFailed'));
      return;
    }
    message.success(t('mouse.resetHidDone'));
    onDone();
  }

  return (
    <Button onClick={recover} loading={isRecovering}>
      {t('mouse.resetHid')}
    </Button>
  );
};
