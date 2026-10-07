import { useAtomValue } from 'jotai';

import { showRoomMicControls, unknownRoomMicStatus } from '@/lib/room-mic.ts';
import { roomMicStatusAtom, roomMicTransportAtom } from '@/jotai/room-mic.ts';

// useRoomMicControls reports whether the audio menu offers the room
// microphone: the kernel has it, an administrator allowed it, and the video
// connection in use carries it.
export function useRoomMicControls(): boolean {
  const status = useAtomValue(roomMicStatusAtom) ?? unknownRoomMicStatus;
  const transport = useAtomValue(roomMicTransportAtom);

  return showRoomMicControls(status, transport);
}
