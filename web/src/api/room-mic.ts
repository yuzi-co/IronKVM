import { http } from '@/lib/http.ts';

// The room microphone's state. Every signed-in viewer may read it.
export function getRoomMic() {
  return http.get('/api/room-mic');
}

// Change the administrator's settings. A field left out keeps its value.
export function setRoomMic(settings: { allowed?: boolean; gain?: number }) {
  return http.post('/api/room-mic', settings);
}
