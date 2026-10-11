import { http } from '@/lib/http.ts';

// What the board knows of the stream's source, for the status bar: the HDMI
// input's mode and the rate the capture loop delivers frames at.
export type StreamInfo = {
  width: number;
  height: number;
  // Millihertz, 0 when the capture library does not say.
  refresh: number;
  // Frames per second over the last few seconds; -1 before any stream ran.
  captureFps: number;
  // H.264 Direct's frames are stamped with their capture time rather than
  // when they left the encoder (libkvm-v4l2 boards). Absent on older servers.
  captureStamps?: boolean;
};

export function getStreamInfo() {
  return http.get('/api/stream/info');
}

// read the board-wide frame detect setting
export function getFrameDetect() {
  return http.get('/api/stream/mjpeg/detect');
}

// enable/disable frame detect
export function updateFrameDetect(enabled: boolean) {
  const data = {
    enabled
  };
  return http.post('/api/stream/mjpeg/detect', data);
}

// pause frame detect for a while (prevent a black screen when opening the page for the first time)
export function stopFrameDetect(duration: number) {
  const data = {
    duration
  };
  return http.post('/api/stream/mjpeg/detect/stop', data);
}

// getScreenshot fetches one full frame of the host screen as a JPEG. When the
// board has no frame to give, it answers with the usual JSON envelope instead,
// so the type of the answer tells the two apart.
export async function getScreenshot(): Promise<Blob> {
  const data = (await http.request({
    method: 'get',
    url: '/api/stream/screenshot',
    responseType: 'blob'
  })) as unknown as Blob;

  if (data.type.startsWith('image/')) {
    return data;
  }

  let message = '';
  try {
    message = (JSON.parse(await data.text()) as { msg?: string }).msg ?? '';
  } catch {
    // An answer that is neither a picture nor an envelope says nothing useful.
  }
  throw new Error(message || 'failed to capture screenshot');
}
