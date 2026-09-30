import { useEffect, useLayoutEffect, useRef } from 'react';
import { useAtomValue } from 'jotai';

import { client, MessageEvent } from '@/lib/websocket.ts';
import { inputRegionAtom, resolutionAtom } from '@/jotai/screen.ts';

import { getScreenPosition } from './position.ts';

// The gadget's touch screen takes two contacts, with IDs the host uses to
// tell the fingers apart.
const MAX_CONTACTS = 2;

// A touch device keeps reporting while a finger is down, and Linux's
// hid-multitouch lifts a contact that goes quiet for about 100 ms. A finger
// held still sends no pointer events, so the frame is repeated at this pace.
const KEEP_ALIVE_MS = 50;

// X and Y run from 0 to this on the host, like the absolute mouse.
const COORD_MAX = 0x7fff;

type Contact = {
  id: number;
  x: number;
  y: number;
  down: boolean;
};

function disableEvent(event: Event) {
  event.preventDefault();
  event.stopPropagation();
}

// Touch sends the fingers on the screen to the host's touch screen as they
// are, through the Pointer Events API. The host reads taps, drags,
// two-finger scrolling and press-and-hold itself, the way it does for a
// touch screen of its own, so nothing here interprets a gesture. A mouse
// works too: the left button is a finger.
export const Touch = () => {
  const resolution = useAtomValue(resolutionAtom);
  const inputRegion = useAtomValue(inputRegionAtom);
  // Read through a ref so a region change does not lift the fingers on the
  // host by re-registering the handlers mid-gesture.
  const inputRegionRef = useRef(inputRegion);
  useLayoutEffect(() => {
    inputRegionRef.current = inputRegion;
  }, [inputRegion]);

  useEffect(() => {
    const screen = document.getElementById('screen');
    if (!screen) return;
    const target = screen;

    const contacts = new Map<number, Contact>();
    let moveFrame: number | null = null;
    let keepAlive: ReturnType<typeof setInterval> | null = null;

    // Without this the browser pans, zooms or scrolls the page with the same
    // fingers, and cancels the pointers when it does.
    const previousTouchAction = target.style.touchAction;
    target.style.touchAction = 'none';

    target.addEventListener('pointerdown', handlePointerDown);
    target.addEventListener('pointermove', handlePointerMove);
    target.addEventListener('pointerup', handlePointerUp);
    target.addEventListener('pointercancel', handlePointerUp);
    target.addEventListener('lostpointercapture', handlePointerUp);
    target.addEventListener('contextmenu', disableEvent);
    target.addEventListener('click', disableEvent);

    function getPosition(e: PointerEvent, clamp: boolean) {
      const position = getScreenPosition(
        target,
        e.clientX,
        e.clientY,
        resolution,
        inputRegionRef.current,
        clamp
      );
      if (!position) return null;

      return {
        x: Math.round(position.x * COORD_MAX),
        y: Math.round(position.y * COORD_MAX)
      };
    }

    function freeContactId() {
      for (let id = 0; id < MAX_CONTACTS; id++) {
        if (![...contacts.values()].some((contact) => contact.id === id)) return id;
      }
      return null;
    }

    function handlePointerDown(e: PointerEvent) {
      disableEvent(e);
      if (e.pointerType === 'mouse' && e.button !== 0) return;
      if (contacts.has(e.pointerId)) return;

      // A third finger has no contact to be. It is left out rather than
      // taking the place of one that is down.
      const id = freeContactId();
      if (id === null) return;

      const position = getPosition(e, false);
      if (!position) return;

      try {
        target.setPointerCapture(e.pointerId);
      } catch {
        // A pointer that cannot be captured still works while it stays on
        // the screen.
      }

      contacts.set(e.pointerId, { id, ...position, down: true });
      sendFrame();
    }

    function handlePointerMove(e: PointerEvent) {
      const contact = contacts.get(e.pointerId);
      if (!contact) return;
      disableEvent(e);

      // A finger dragged past the edge of the host screen stays on its edge.
      const position = getPosition(e, true);
      if (!position) return;

      contact.x = position.x;
      contact.y = position.y;
      queueFrame();
    }

    // Up, cancel and a lost capture all lift the finger. The frame that
    // lifts it still names it, with the tip up, and then it is gone.
    function handlePointerUp(e: PointerEvent) {
      const contact = contacts.get(e.pointerId);
      if (!contact) return;
      disableEvent(e);

      const position = getPosition(e, true);
      if (position) {
        contact.x = position.x;
        contact.y = position.y;
      }
      contact.down = false;
      sendFrame();
    }

    function queueFrame() {
      if (moveFrame !== null) return;

      moveFrame = requestAnimationFrame(() => {
        moveFrame = null;
        sendFrame();
      });
    }

    // sendFrame sends every contact at once, then forgets the lifted ones.
    function sendFrame() {
      if (moveFrame !== null) {
        cancelAnimationFrame(moveFrame);
        moveFrame = null;
      }
      if (contacts.size === 0) return;

      const data = [MessageEvent.Touch, contacts.size];
      for (const contact of contacts.values()) {
        data.push(
          contact.down ? 1 : 0,
          contact.id,
          contact.x & 0xff,
          contact.x >> 8,
          contact.y & 0xff,
          contact.y >> 8
        );
      }
      client.send(new Uint8Array(data));

      for (const [pointerId, contact] of contacts) {
        if (!contact.down) contacts.delete(pointerId);
      }
      updateKeepAlive();
    }

    function updateKeepAlive() {
      if (contacts.size > 0 && keepAlive === null) {
        keepAlive = setInterval(sendFrame, KEEP_ALIVE_MS);
      } else if (contacts.size === 0 && keepAlive !== null) {
        clearInterval(keepAlive);
        keepAlive = null;
      }
    }

    return () => {
      // Leaving touch mode lifts every finger, or the host keeps one down.
      for (const contact of contacts.values()) {
        contact.down = false;
      }
      sendFrame();
      if (keepAlive !== null) {
        clearInterval(keepAlive);
      }

      target.style.touchAction = previousTouchAction;
      target.removeEventListener('pointerdown', handlePointerDown);
      target.removeEventListener('pointermove', handlePointerMove);
      target.removeEventListener('pointerup', handlePointerUp);
      target.removeEventListener('pointercancel', handlePointerUp);
      target.removeEventListener('lostpointercapture', handlePointerUp);
      target.removeEventListener('contextmenu', disableEvent);
      target.removeEventListener('click', disableEvent);
    };
  }, [resolution]);

  return <></>;
};
