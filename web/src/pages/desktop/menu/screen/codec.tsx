import { useEffect, useState } from 'react';
import { Popover } from 'antd';
import { CheckIcon, FilmIcon } from 'lucide-react';

import { updateScreen } from '@/api/vm.ts';

type CodecProps = {
  codec: number;
  setCodec: (codec: number) => void;
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

// Whether this browser can decode HEVC at all. Firefox usually cannot, and
// hardware support varies even where the API exists. There is one encoder on
// the board, so selecting H.265 changes the stream for every viewer: an
// operator whose browser cannot decode it would take the picture away from
// everyone and see no error. Ask before offering the choice.
async function hevcIsDecodable(): Promise<boolean> {
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

export const Codec = ({ codec, setCodec }: CodecProps) => {
  const [hevcSupported, setHevcSupported] = useState(false);

  useEffect(() => {
    let cancelled = false;

    hevcIsDecodable().then((supported) => {
      if (!cancelled) {
        setHevcSupported(supported);
      }
    });

    return () => {
      cancelled = true;
    };
  }, []);

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
            title={disabled ? 'This browser cannot decode H.265' : undefined}
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
        WebRTC mode carries H.264 only.
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
