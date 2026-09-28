import { useCallback, useEffect, useRef, useState, type PointerEvent } from 'react';
import { Button, Card, theme } from 'antd';
import { useAtomValue, useSetAtom } from 'jotai';
import { useTranslation } from 'react-i18next';

import { keyboardLockAtom } from '@/jotai/keyboard.ts';
import { resolutionAtom } from '@/jotai/screen.ts';

import { getMediaSize, getRenderedMediaRect } from '../screen/geometry.ts';
import { intersectRects, minimumSelectionSize, normalizeSelection, type Rect } from './crop.ts';

export type OcrSelectionResult = {
  // The rectangle the user drew, in page pixels.
  selection: Rect;
  // Where the page drew the whole picture when the drag ended, in page pixels.
  picture: Rect;
};

type Layout = {
  picture: Rect;
  // The part of the picture the user can see: the viewport clips it when the
  // input region crops the view, and the window clips it when zoomed in.
  visible: Rect;
};

// measureLayout finds the picture on the page. The element keeps the frame's
// aspect ratio with object-fit: contain, so the bars around the picture are
// taken out here.
function measureLayout(fallback: { width: number; height: number } | null): Layout | null {
  const screen = document.getElementById('screen');
  if (!screen) {
    return null;
  }

  const mediaSize = getMediaSize(screen, fallback);
  const bounds = screen.getBoundingClientRect();
  if (!mediaSize || bounds.width <= 0 || bounds.height <= 0) {
    return null;
  }

  const picture = getRenderedMediaRect(bounds, mediaSize);
  let visible = intersectRects(picture, {
    left: 0,
    top: 0,
    width: window.innerWidth,
    height: window.innerHeight
  });
  const viewport = document.getElementById('screen-viewport');
  if (visible && viewport) {
    visible = intersectRects(visible, viewport.getBoundingClientRect());
  }

  return visible ? { picture, visible } : null;
}

type OcrSelectionProps = {
  onSelect: (result: OcrSelectionResult) => void;
  onCancel: () => void;
};

// OcrSelection covers the desktop while the user drags a rectangle over the
// text to read. Covering it is also what keeps the mouse off the host: the
// mouse handlers listen on the screen element, which the overlay sits above.
// The keyboard is locked the way the paste dialog locks it.
export const OcrSelection = ({ onSelect, onCancel }: OcrSelectionProps) => {
  const { t } = useTranslation();
  const { token } = theme.useToken();
  const resolution = useAtomValue(resolutionAtom);
  const setKeyboardLock = useSetAtom(keyboardLockAtom);

  const [layout, setLayout] = useState<Layout | null>(null);
  const [selection, setSelection] = useState<Rect | null>(null);
  const startRef = useRef<{ x: number; y: number } | null>(null);

  useEffect(() => {
    setKeyboardLock({ source: 'ocr-selection', locked: true });
    return () => setKeyboardLock({ source: 'ocr-selection', locked: false });
  }, [setKeyboardLock]);

  // In relative mouse mode the pointer may be locked to the screen, and a
  // locked pointer sends its movement to the host whatever lies above it.
  useEffect(() => {
    if (document.pointerLockElement) {
      document.exitPointerLock();
    }
  }, []);

  useEffect(() => {
    function update() {
      setLayout(measureLayout(resolution));
    }

    update();
    window.addEventListener('resize', update);
    return () => window.removeEventListener('resize', update);
  }, [resolution]);

  const cancel = useCallback(() => {
    startRef.current = null;
    setSelection(null);
    onCancel();
  }, [onCancel]);

  useEffect(() => {
    function handleKeyDown(event: KeyboardEvent) {
      if (event.key === 'Escape') {
        event.preventDefault();
        event.stopImmediatePropagation();
        cancel();
      }
    }

    document.addEventListener('keydown', handleKeyDown, true);
    return () => document.removeEventListener('keydown', handleKeyDown, true);
  }, [cancel]);

  function clamp(x: number, y: number, area: Rect) {
    return {
      x: Math.max(area.left, Math.min(area.left + area.width, x)),
      y: Math.max(area.top, Math.min(area.top + area.height, y))
    };
  }

  function handlePointerDown(event: PointerEvent<HTMLDivElement>) {
    if (event.button !== 0) {
      return;
    }

    // The picture moves when the window or the zoom changes, so it is measured
    // again at the start of every drag.
    const current = measureLayout(resolution);
    setLayout(current);
    if (!current) {
      return;
    }

    event.currentTarget.setPointerCapture(event.pointerId);
    const point = clamp(event.clientX, event.clientY, current.visible);
    startRef.current = point;
    setSelection(normalizeSelection(point, point));
  }

  function handlePointerMove(event: PointerEvent<HTMLDivElement>) {
    const start = startRef.current;
    if (!start || !layout) {
      return;
    }

    setSelection(normalizeSelection(start, clamp(event.clientX, event.clientY, layout.visible)));
  }

  function handlePointerUp(event: PointerEvent<HTMLDivElement>) {
    const start = startRef.current;
    startRef.current = null;
    if (event.currentTarget.hasPointerCapture(event.pointerId)) {
      event.currentTarget.releasePointerCapture(event.pointerId);
    }
    if (!start || !layout) {
      return;
    }

    const finished = normalizeSelection(start, clamp(event.clientX, event.clientY, layout.visible));
    if (finished.width < minimumSelectionSize || finished.height < minimumSelectionSize) {
      setSelection(null);
      return;
    }

    setSelection(finished);
    onSelect({ selection: finished, picture: layout.picture });
  }

  return (
    <div
      className="fixed inset-0 z-1100 cursor-crosshair touch-none select-none"
      onContextMenu={(event) => event.preventDefault()}
      onPointerDown={handlePointerDown}
      onPointerMove={handlePointerMove}
      onPointerUp={handlePointerUp}
      onPointerCancel={handlePointerUp}
    >
      {layout && (
        <div
          className="pointer-events-none fixed border border-dashed"
          style={{
            left: layout.visible.left,
            top: layout.visible.top,
            width: layout.visible.width,
            height: layout.visible.height,
            borderColor: token.colorTextSecondary
          }}
        />
      )}
      {selection && (
        // The shadow dims everything around the selection and leaves the
        // text inside it as bright as it was.
        <div
          className="pointer-events-none fixed border"
          style={{
            left: selection.left,
            top: selection.top,
            width: selection.width,
            height: selection.height,
            borderColor: token.colorPrimary,
            boxShadow: '0 0 0 100vmax rgba(0, 0, 0, 0.45)'
          }}
        />
      )}
      <div
        className="fixed top-5 left-1/2 z-1120 max-w-[calc(100%-1rem)] -translate-x-1/2 cursor-default"
        onPointerDown={(event) => event.stopPropagation()}
      >
        <Card
          size="small"
          style={{ boxShadow: token.boxShadowSecondary }}
          styles={{ body: { padding: 8 } }}
        >
          <div className="flex items-center gap-3">
            <span className="text-sm">
              {layout ? t('screen.ocr.hint') : t('screen.ocr.noPicture')}
            </span>
            <Button size="small" onClick={cancel}>
              {t('screen.ocr.cancel')}
            </Button>
          </div>
        </Card>
      </div>
    </div>
  );
};
