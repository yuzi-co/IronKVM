import {
  cloneElement,
  CSSProperties,
  ReactElement,
  PointerEvent as ReactPointerEvent,
  useEffect,
  useRef,
  useState
} from 'react';
import { useAtomValue } from 'jotai';
import { MoveIcon } from 'lucide-react';
import { useTranslation } from 'react-i18next';

import {
  controlRegionModeAtom,
  inputRegionAtom,
  inputRegionSelectingAtom,
  resolutionAtom,
  videoScaleAtom
} from '@/jotai/screen.ts';

import { centred, dragPan, fitContain, isPannable, Pan, panOffset, panTravel } from './pan.ts';

type ViewportSize = {
  width: number;
  height: number;
};

type ScreenViewportProps = {
  children: ReactElement<{ style?: CSSProperties }>;
};

export const ScreenViewport = ({ children }: ScreenViewportProps) => {
  const inputRegion = useAtomValue(inputRegionAtom);
  const mode = useAtomValue(controlRegionModeAtom);
  const selecting = useAtomValue(inputRegionSelectingAtom);
  const videoScale = useAtomValue(videoScaleAtom);
  const resolution = useAtomValue(resolutionAtom);
  const containerRef = useRef<HTMLDivElement>(null);
  const [containerSize, setContainerSize] = useState<ViewportSize>({ width: 0, height: 0 });
  const cropped = mode !== 'off' && !!inputRegion && !selecting;
  const [pan, setPan] = useState<Pan>(centred);

  useEffect(() => {
    const container = containerRef.current;
    if (!container) {
      return;
    }

    const updateSize = () => {
      const rect = container.getBoundingClientRect();
      setContainerSize({ width: rect.width, height: rect.height });
    };
    updateSize();

    const observer = new ResizeObserver(updateSize);
    observer.observe(container);
    return () => observer.disconnect();
  }, []);

  let viewportWidth = containerSize.width;
  let viewportHeight = containerSize.height;
  if (cropped && inputRegion && containerSize.width > 0 && containerSize.height > 0) {
    const scale = Math.min(
      containerSize.width / inputRegion.width,
      containerSize.height / inputRegion.height
    );
    viewportWidth = inputRegion.width * scale;
    viewportHeight = inputRegion.height * scale;
  }

  // A cropped viewport already has the region's shape; otherwise the
  // picture is letterboxed inside it, and the bars are not worth panning to.
  const picture = cropped
    ? { width: viewportWidth, height: viewportHeight }
    : fitContain({ width: viewportWidth, height: viewportHeight }, resolution);
  const travel = panTravel(picture, containerSize, videoScale);
  const offset = panOffset(pan, travel);

  const mediaStyle: CSSProperties =
    cropped && inputRegion
      ? {
          ...children.props.style,
          position: 'absolute',
          left: -(inputRegion.left / inputRegion.width) * viewportWidth,
          top: -(inputRegion.top / inputRegion.height) * viewportHeight,
          width: (inputRegion.frameWidth / inputRegion.width) * viewportWidth,
          height: (inputRegion.frameHeight / inputRegion.height) * viewportHeight,
          maxWidth: 'none',
          maxHeight: 'none',
          objectFit: 'fill',
          transform: 'none'
        }
      : {
          ...children.props.style,
          width: '100%',
          height: '100%',
          objectFit: 'contain',
          transform: 'none'
        };

  return (
    <div
      ref={containerRef}
      className="relative flex h-full min-h-0 w-full min-w-0 items-center justify-center overflow-hidden"
    >
      <div
        id="screen-viewport"
        className="relative overflow-hidden"
        data-cropped={cropped ? 'true' : 'false'}
        style={{
          width: viewportWidth || '100%',
          height: viewportHeight || '100%',
          transform: `translate(${offset.x}px, ${offset.y}px) scale(${videoScale})`,
          transformOrigin: 'center'
        }}
      >
        {cloneElement(children, { style: mediaStyle })}
      </div>
      {isPannable(travel) && (
        <PanHandle
          onDrag={(dx, dy) => setPan((current) => dragPan(current, dx, dy, travel))}
          onCentre={() => setPan(centred)}
        />
      )}
    </div>
  );
};

const TAP_SLOP = 8;
const DOUBLE_TAP_MS = 400;

type PanHandleProps = {
  onDrag: (dx: number, dy: number) => void;
  onCentre: () => void;
};

// PanHandle moves a zoomed picture. It is a control of its own beside the
// picture, rather than a gesture on it, because every gesture on the picture
// already means something to the host: a drag moves the pointer or drags with
// it, two fingers scroll, press-and-hold is the right button, and touch mode
// hands all of them to the host as they are. Dragging the handle drags the
// picture, with a finger or a mouse; a double tap or double click centres it.
const PanHandle = ({ onDrag, onCentre }: PanHandleProps) => {
  const { t } = useTranslation();
  const lastRef = useRef<{ id: number; x: number; y: number; moved: number } | null>(null);
  const lastTapRef = useRef(0);

  function handlePointerDown(e: ReactPointerEvent<HTMLButtonElement>) {
    if (e.pointerType === 'mouse' && e.button !== 0) return;
    e.preventDefault();
    e.stopPropagation();
    try {
      e.currentTarget.setPointerCapture(e.pointerId);
    } catch {
      // Without the capture the drag still works while the pointer stays on
      // the handle.
    }
    lastRef.current = { id: e.pointerId, x: e.clientX, y: e.clientY, moved: 0 };
  }

  function handlePointerMove(e: ReactPointerEvent<HTMLButtonElement>) {
    const last = lastRef.current;
    if (!last || last.id !== e.pointerId) return;
    e.preventDefault();
    e.stopPropagation();
    const dx = e.clientX - last.x;
    const dy = e.clientY - last.y;
    onDrag(dx, dy);
    lastRef.current = {
      id: e.pointerId,
      x: e.clientX,
      y: e.clientY,
      moved: last.moved + Math.hypot(dx, dy)
    };
  }

  // A tap is a press that hardly moved. Two of them close together centre
  // the picture; dblclick is not used, as a phone may not send it.
  function handlePointerUp(e: ReactPointerEvent<HTMLButtonElement>) {
    const last = lastRef.current;
    if (!last || last.id !== e.pointerId) return;
    lastRef.current = null;
    if (e.type !== 'pointerup' || last.moved > TAP_SLOP) return;

    const now = e.timeStamp;
    if (now - lastTapRef.current < DOUBLE_TAP_MS) {
      lastTapRef.current = 0;
      onCentre();
    } else {
      lastTapRef.current = now;
    }
  }

  return (
    <button
      type="button"
      aria-label={t('screen.pan')}
      title={t('screen.pan')}
      className="absolute top-[52px] right-2 z-10 flex size-10 cursor-move touch-none items-center justify-center rounded-full border border-white/15 bg-neutral-800/70 p-0 text-neutral-200 shadow-lg select-none hover:bg-neutral-700/80"
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
      onPointerCancel={handlePointerUp}
      onLostPointerCapture={handlePointerUp}
      onContextMenu={(e) => e.preventDefault()}
    >
      <MoveIcon size={20} />
    </button>
  );
};
