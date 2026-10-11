// The device stamps each video frame with when its capture began, in the RTP
// header extension abs-capture-time (ironkvm-dist trial 78), as PiKVM does to
// measure its latency. Chrome supports the extension but leaves it out of its
// offers unless the page asks for it: getHeaderExtensionsToNegotiate() lists
// it as "stopped". The server only sends what the offer negotiated, so this
// turns it on for the video transceiver before the offer is made.
//
// Once negotiated, RTCRtpReceiver.getSynchronizationSources()[0] carries, per
// frame:
//   captureTimestamp         the capture time on the device's NTP clock (ms)
//   senderCaptureTimeOffset  0 (the device captures and sends)
//   timestamp                when the frame reached the track, Unix epoch ms
// so the capture-to-display latency is
//   timestamp - (captureTimestamp + senderCaptureTimeOffset - 2208988800000)
// which needs the browser's and the device's clocks in step (NTP on both).

export const ABS_CAPTURE_TIME_URI = 'http://www.webrtc.org/experiments/rtp-hdrext/abs-capture-time';

/** NTP's epoch (1900) to Unix's (1970), in milliseconds. */
export const NTP_EPOCH_OFFSET_MS = 2208988800000;

type HeaderExtension = { uri: string; direction: string };

type ExtensionControl = {
  getHeaderExtensionsToNegotiate?: () => HeaderExtension[];
  setHeaderExtensionsToNegotiate?: (extensions: HeaderExtension[]) => void;
};

/**
 * Asks the browser to offer abs-capture-time on a receive-only transceiver.
 * Answers whether it is now offered. Browsers without the header extension
 * control API (Firefox, Safari) are left alone.
 */
export function offerAbsCaptureTime(target: object): boolean {
  // Not in every TypeScript DOM library yet, hence the cast.
  const transceiver = target as ExtensionControl;
  if (!transceiver.getHeaderExtensionsToNegotiate || !transceiver.setHeaderExtensionsToNegotiate) {
    return false;
  }

  try {
    const extensions = transceiver.getHeaderExtensionsToNegotiate();
    const ext = extensions.find((e) => e.uri === ABS_CAPTURE_TIME_URI);
    if (!ext) {
      return false;
    }
    if (ext.direction === 'stopped') {
      ext.direction = 'recvonly';
      transceiver.setHeaderExtensionsToNegotiate(extensions);
    }
    return true;
  } catch (err) {
    console.warn('abs-capture-time could not be offered:', err);
    return false;
  }
}

/**
 * The latency from capture to the frame reaching the track, in ms, from a
 * synchronization source, or undefined without a capture time.
 */
export function captureLatencyMs(source: {
  timestamp: number;
  captureTimestamp?: number;
  senderCaptureTimeOffset?: number;
}): number | undefined {
  if (source.captureTimestamp === undefined) {
    return undefined;
  }
  return (
    source.timestamp -
    (source.captureTimestamp + (source.senderCaptureTimeOffset ?? 0) - NTP_EPOCH_OFFSET_MS)
  );
}
