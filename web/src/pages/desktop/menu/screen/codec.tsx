import { useEffect, useState } from 'react';
import { Popover } from 'antd';
import { CheckIcon, FilmIcon } from 'lucide-react';

import { updateScreen } from '@/api/vm.ts';

type CodecProps = {
  codec: number;
  setCodec: (codec: number) => void;
  // Which delivery path is in use. The two carry H.265 through different
  // browser machinery and support for them is not the same, so the question
  // "can this browser decode H.265" has two different answers.
  videoMode: string;
};

// libkvm's public numbering, which is not the same as mmf's and runs the other
// way round. The server converts; nothing in the browser should.
export const CODEC_H264 = 1;
export const CODEC_H265 = 2;

// Kept in step with direct.worker.ts. The worker reads the codec out of the
// bitstream rather than being told, so these two only have to agree on what
// they ask the browser to decode.
const hevcCodecString = 'hev1.1.6.L120.B0';

const codecList = [
  { key: CODEC_H264, label: 'H.264' },
  { key: CODEC_H265, label: 'H.265' }
];

// The direct path decodes with WebCodecs. Firefox usually cannot decode HEVC
// there, and hardware support varies even where the API exists.
async function hevcIsDecodableByWebCodecs(): Promise<boolean> {
  if (typeof VideoDecoder === 'undefined' || !VideoDecoder.isConfigSupported) {
    return false;
  }

  try {
    const support = await VideoDecoder.isConfigSupported({ codec: hevcCodecString });
    return support.supported === true;
  } catch {
    return false;
  }
}

// WebRTC decodes through the peer connection instead, and its HEVC support is
// narrower than WebCodecs': a browser that decodes HEVC on the direct path may
// still not receive it over WebRTC. Ask the receiver rather than assuming the
// two agree.
function hevcIsReceivableByWebRTC(): boolean {
  if (typeof RTCRtpReceiver === 'undefined' || !RTCRtpReceiver.getCapabilities) {
    return false;
  }

  try {
    const capabilities = RTCRtpReceiver.getCapabilities('video');
    return (
      capabilities?.codecs?.some((entry) => entry.mimeType.toLowerCase() === 'video/h265') === true
    );
  } catch {
    return false;
  }
}

// There is one encoder on the board, so selecting H.265 changes the stream for
// every viewer. An operator whose browser cannot decode it would take the
// picture away from everyone and see no error, so ask before offering it.
async function hevcIsUsable(videoMode: string): Promise<boolean> {
  if (videoMode === 'h264') {
    return hevcIsReceivableByWebRTC();
  }

  return hevcIsDecodableByWebCodecs();
}

export const Codec = ({ codec, setCodec, videoMode }: CodecProps) => {
  const [hevcSupported, setHevcSupported] = useState(false);

  useEffect(() => {
    let cancelled = false;

    hevcIsUsable(videoMode).then((supported) => {
      if (!cancelled) {
        setHevcSupported(supported);
      }
    });

    return () => {
      cancelled = true;
    };
  }, [videoMode]);

  async function update(value: number) {
    if (value === codec) return;
    if (value === CODEC_H265 && !hevcSupported) return;

    const rsp = await updateScreen('codec', value);
    if (rsp.code !== 0) {
      return;
    }

    setCodec(value);
  }

  const content = (
    <>
      {codecList.map((item) => {
        const disabled = item.key === CODEC_H265 && !hevcSupported;

        return (
          <div
            key={item.key}
            className={
              disabled
                ? 'flex select-none items-center rounded py-1 pl-1 pr-6 opacity-40'
                : 'flex cursor-pointer select-none items-center rounded py-1 pl-1 pr-6 hover:bg-neutral-700/70'
            }
            onClick={() => update(item.key)}
            title={
              disabled
                ? videoMode === 'h264'
                  ? 'This browser cannot receive H.265 over WebRTC'
                  : 'This browser cannot decode H.265'
                : undefined
            }
          >
            <div className="flex h-[14px] w-[20px] items-end text-blue-500">
              {item.key === codec && <CheckIcon size={14} />}
            </div>
            <span>{item.label}</span>
          </div>
        );
      })}
      <div className="max-w-[220px] px-1 pt-2 text-xs text-neutral-400">
        The board has one encoder, so this changes the stream for every viewer.
        Reconnect to apply it to a running WebRTC session.
      </div>
    </>
  );

  return (
    <Popover content={content} placement="rightTop" arrow={false} align={{ offset: [14, 0] }}>
      <div className="flex h-[30px] cursor-pointer items-center space-x-2 rounded px-3 text-neutral-300 hover:bg-neutral-700/70">
        <FilmIcon size={18} />
        <span className="select-none text-sm">Codec</span>
      </div>
    </Popover>
  );
};
