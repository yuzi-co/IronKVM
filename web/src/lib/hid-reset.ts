import * as api from '@/api/hid.ts';
import { client, MessageEvent } from '@/lib/websocket.ts';

export type HidResetResult = { ok: true } | { ok: false; msg?: string };

// resetHid re-enumerates the USB gadget so the target binds its HID interfaces
// again. Keys held down are released first, and the input socket, which goes
// down with the gadget, is closed before and opened again after.
export async function resetHid(): Promise<HidResetResult> {
  client.send(new Uint8Array([MessageEvent.Keyboard, 0, 0, 0, 0, 0, 0, 0, 0]));
  client.close();

  try {
    const rsp = await api.reset();
    return rsp.code === 0 ? { ok: true } : { ok: false, msg: rsp.msg };
  } catch {
    return { ok: false };
  } finally {
    client.connect();
  }
}
