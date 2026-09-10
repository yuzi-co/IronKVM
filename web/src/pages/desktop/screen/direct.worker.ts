import Queue from 'yocto-queue';

let canvas: OffscreenCanvas | null = null;
let ctx: OffscreenCanvasRenderingContext2D | null = null;
let rendering: boolean = false;
let flushScheduled: boolean = false;
let decoder: VideoDecoder | null = null;
let streamUrl: string | null = null;
let socket: WebSocket | null = null;
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let reconnectDelayMs = 250;
let stopped = false;
let resyncRequested = false;
let pendingAckTimestamp: number | null = null;
let decodeBackpressured = false;
let reportedFrameWidth = 0;
let reportedFrameHeight = 0;
let configuredCodec: 'avc' | 'hevc' | null = null;

// The board has one hardware encoder, so which of these arrives is the
// operator's global choice, not this viewer's. The stream is read rather than
// asked about: a keyframe says what it is, which keeps the worker right even
// when the codec changes under a live connection.
const avcCodec = 'avc1.42E02A';
// Main profile, level 4.0, the lowest that covers 1080p30.
const hevcCodec = 'hev1.1.6.L120.B0';

const maxQueuedFrames = 1;
const maxReconnectDelayMs = 5_000;
const frameAckMessage = 2;
const streamResyncMessage = 3;
const flowControlWindow = 8;
const decoderHighWatermark = 6;
const decoderLowWatermark = 3;
const frameQueue = new Queue<VideoFrame>();
const frameChannel = new MessageChannel();

type WorkerMessage = {
  type: 'h264' | 'stop';
  canvas?: OffscreenCanvas;
  url?: string;
};

frameChannel.port1.onmessage = () => {
  flushScheduled = false;
  processFrameQueue();
};

self.onmessage = (event: MessageEvent<WorkerMessage>) => {
  const { type, canvas: offscreenCanvas, url } = event.data;

  switch (type) {
    case 'h264':
      if (!offscreenCanvas || !url) {
        return;
      }

      canvas = offscreenCanvas;
      reportedFrameWidth = 0;
      reportedFrameHeight = 0;
      ctx = canvas!.getContext('2d', {
        alpha: false,
        desynchronized: true
      }) as OffscreenCanvasRenderingContext2D;
      streamUrl = url;
      stopped = false;
      connect();
      break;
    case 'stop':
      stopped = true;
      clearReconnectTimer();
      disconnect();
      resetDecoder();
      break;
  }
};

function connect() {
  if (stopped || !streamUrl || socket) {
    return;
  }

  try {
    const url = new URL(streamUrl);
    url.searchParams.set('flow', String(flowControlWindow));
    const nextSocket = new WebSocket(url);
    nextSocket.binaryType = 'arraybuffer';
    socket = nextSocket;

    nextSocket.onopen = () => {
      if (socket !== nextSocket || stopped) {
        return;
      }

      reconnectDelayMs = 250;
      resyncRequested = false;
      pendingAckTimestamp = null;
      decodeBackpressured = false;
    };

    nextSocket.onmessage = (event) => {
      if (socket !== nextSocket || stopped || !(event.data instanceof ArrayBuffer)) {
        return;
      }

      handleWsMessage(event.data);
    };

    nextSocket.onerror = () => {
      if (socket === nextSocket) {
        nextSocket.close();
      }
    };

    nextSocket.onclose = () => {
      if (socket !== nextSocket) {
        return;
      }

      socket = null;
      resyncRequested = false;
      pendingAckTimestamp = null;
      decodeBackpressured = false;
      resetDecoder();
      scheduleReconnect();
    };
  } catch (error) {
    console.error('Failed to create Direct H264 WebSocket:', error);
    scheduleReconnect();
  }
}

function disconnect() {
  const currentSocket = socket;
  socket = null;

  if (currentSocket && currentSocket.readyState !== WebSocket.CLOSED) {
    currentSocket.close();
  }
}

function scheduleReconnect() {
  if (stopped || reconnectTimer !== null || !streamUrl) {
    return;
  }

  const delay = reconnectDelayMs;
  reconnectDelayMs = Math.min(reconnectDelayMs * 2, maxReconnectDelayMs);
  reconnectTimer = setTimeout(() => {
    reconnectTimer = null;
    connect();
  }, delay);
}

function clearReconnectTimer() {
  if (reconnectTimer !== null) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
}

// Read the codec out of an Annex-B keyframe. H.264 opens with an SPS, whose
// nal_unit_type is 7 in the low five bits. HEVC opens with a VPS, type 32 in
// bits 1 to 6 of the same byte. Nothing else needs to agree for this to work,
// which is why the codec travels in no header and needs no API call.
function detectCodec(data: Uint8Array): 'avc' | 'hevc' | null {
  for (let i = 0; i + 3 < data.length; ) {
    let prefix = 0;
    if (i + 4 < data.length && data[i] === 0 && data[i + 1] === 0 && data[i + 2] === 0 && data[i + 3] === 1) {
      prefix = 4;
    } else if (data[i] === 0 && data[i + 1] === 0 && data[i + 2] === 1) {
      prefix = 3;
    }
    if (prefix === 0) {
      i++;
      continue;
    }

    const nal = i + prefix;
    if (nal >= data.length) {
      break;
    }
    if ((data[nal] & 0x1f) === 7) {
      return 'avc';
    }
    const hevcType = (data[nal] >> 1) & 0x3f;
    if (hevcType === 32 || hevcType === 33) {
      return 'hevc';
    }
    i = nal + 1;
  }

  return null;
}

function handleWsMessage(message: ArrayBuffer) {
  try {
    if (message.byteLength < 9) {
      return;
    }

    const view = new DataView(message);
    const isKeyFrame = view.getUint8(0) === 1;
    const timestamp = Number(view.getBigUint64(1, true));
    const data = new Uint8Array(message, 9);

    // A keyframe carries the parameter sets, so it is the only frame that can
    // say which codec this is, and the only frame a fresh decoder can start
    // on. The operator can change the codec under a live connection: the
    // encoder rebuilds its channel and the next keyframe arrives in the new
    // codec, so a decoder configured for the old one is torn down here rather
    // than left to fail on data it cannot read.
    if (isKeyFrame) {
      const seen = detectCodec(data);
      if (seen) {
        if (seen !== configuredCodec && decoder) {
          resetDecoder();
        }
        configuredCodec = seen;
      }
    }

    if (!decoder) {
      if (!isKeyFrame) {
        requestStreamResync();
        return;
      }

      resyncRequested = false;
      const initial = createDecoder();
      if (!initial) {
        requestStreamResync();
        return;
      }

      decoder = initial;
      resyncRequested = false;
    }

    if (decoder?.state === 'configured') {
      decode(decoder, isKeyFrame, timestamp, data);
    }
  } catch (error) {
    console.error('Error processing WebSocket message in worker:', error);
  }
}

function createDecoder(): VideoDecoder | null {
  if (!self.VideoDecoder) {
    console.log('Error: WebCodecs API not supported in this worker.');
    return null;
  }

  let instance: VideoDecoder | null = null;
  const init = {
    output: (frame: VideoFrame) => {
      handleDecodedFrame(instance, frame);
    },
    error: () => {
      if (decoder === instance) {
        requestStreamResync();
        resetDecoder();
      }
    }
  };

  try {
    instance = new VideoDecoder(init);
    const configuredDecoder = instance;
    instance.ondequeue = () => {
      releaseDecodeBackpressure(configuredDecoder);
    };
    instance.configure({
      codec: configuredCodec === 'hevc' ? hevcCodec : avcCodec,
      hardwareAcceleration: 'prefer-hardware',
      optimizeForLatency: true
    });
    return instance;
  } catch (err) {
    if (instance && instance.state !== 'closed') {
      instance.close();
    }
    console.log(err);
    return null;
  }
}

function handleDecodedFrame(source: VideoDecoder | null, frame: VideoFrame) {
  if (!source) {
    frame.close();
    return;
  }

  if (source !== decoder) {
    frame.close();
    return;
  }

  frameQueue.enqueue(frame);
  while (frameQueue.size > maxQueuedFrames) {
    const droppedFrame = frameQueue.dequeue();
    if (droppedFrame) {
      droppedFrame.close();
    }
  }

  if (!rendering) {
    rendering = true;
    scheduleFrameQueue();
  }
}

function decode(target: VideoDecoder, isKeyFrame: boolean, timestamp: number, data: Uint8Array) {
  const chunk = new EncodedVideoChunk({
    type: isKeyFrame ? 'key' : 'delta',
    timestamp: timestamp,
    data: data
  });

  try {
    target.decode(chunk);
    pendingAckTimestamp = timestamp;
    if (target.decodeQueueSize >= decoderHighWatermark) {
      decodeBackpressured = true;
    }
    releaseDecodeBackpressure(target);
  } catch (err: any) {
    if (err.name === 'TypeError' || err.message.includes('configured')) {
      requestStreamResync();
      resetDecoder();
    }
  }
}

function processFrameQueue() {
  const frame = frameQueue.dequeue();
  if (frame) {
    try {
      renderFrame(frame);
    } catch (error) {
      console.error('Failed to render Direct H264 frame:', error);
      requestStreamResync();
      resetDecoder();
      return;
    }
  }

  if (frameQueue.size > 0) {
    scheduleFrameQueue();
  } else {
    rendering = false;
  }
}

function scheduleFrameQueue() {
  if (flushScheduled) {
    return;
  }

  flushScheduled = true;
  frameChannel.port2.postMessage(null);
}

function renderFrame(frame: VideoFrame) {
  if (!canvas || !ctx) {
    frame.close();
    return;
  }

  try {
    if (canvas.width !== frame.displayWidth || canvas.height !== frame.displayHeight) {
      canvas.width = frame.displayWidth;
      canvas.height = frame.displayHeight;
    }
    ctx.drawImage(frame, 0, 0, canvas.width, canvas.height);

    if (reportedFrameWidth !== frame.displayWidth || reportedFrameHeight !== frame.displayHeight) {
      reportedFrameWidth = frame.displayWidth;
      reportedFrameHeight = frame.displayHeight;
      self.postMessage({
        type: 'frame-size',
        width: reportedFrameWidth,
        height: reportedFrameHeight
      });
    }
  } finally {
    frame.close();
  }
}

function resetDecoder() {
  if (decoder && decoder.state !== 'closed') {
    try {
      decoder.close();
    } catch (err) {
      console.log(err);
    }
  }

  decoder = null;
  pendingAckTimestamp = null;
  decodeBackpressured = false;
  rendering = false;
  flushScheduled = false;

  Array.from(frameQueue.drain()).forEach((frame) => frame.close());
}

function releaseDecodeBackpressure(source: VideoDecoder) {
  if (source !== decoder || pendingAckTimestamp === null) {
    return;
  }

  if (decodeBackpressured && source.decodeQueueSize > decoderLowWatermark) {
    return;
  }

  decodeBackpressured = false;
  acknowledgeFrame(pendingAckTimestamp);
  pendingAckTimestamp = null;
}

function acknowledgeFrame(timestamp: number) {
  const currentSocket = socket;
  if (!currentSocket || currentSocket.readyState !== WebSocket.OPEN) {
    return;
  }

  const message = new ArrayBuffer(9);
  const view = new DataView(message);
  view.setUint8(0, frameAckMessage);
  view.setBigUint64(1, BigInt(timestamp), true);
  try {
    currentSocket.send(message);
  } catch {
    currentSocket.close();
  }
}

function requestStreamResync() {
  if (resyncRequested) {
    return;
  }

  const currentSocket = socket;
  if (!currentSocket || currentSocket.readyState !== WebSocket.OPEN) {
    return;
  }

  try {
    currentSocket.send(new Uint8Array([streamResyncMessage]));
    resyncRequested = true;
  } catch {
    currentSocket.close();
  }
}
