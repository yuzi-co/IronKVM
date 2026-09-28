import { http } from '@/lib/http.ts';

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
