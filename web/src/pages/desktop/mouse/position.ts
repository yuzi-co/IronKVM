import type { InputRegion, Resolution } from '@/types';

import {
  fullFrameContent,
  getConfiguredFrameContent,
  getMediaSize,
  getRenderedMediaRect
} from '../screen/geometry.ts';

// getScreenPosition maps a point in the browser window to the host screen, as
// 0 to 1 on each axis. It follows the cropped viewport, the letterboxing of
// the video and the configured input region. A point outside the host screen
// is null, or with clamp set is pulled to the nearest edge, which is what a
// finger dragged past the edge needs.
export function getScreenPosition(
  target: HTMLElement,
  clientX: number,
  clientY: number,
  resolution: Resolution | null | undefined,
  inputRegion: InputRegion | null | undefined,
  clamp = false
): { x: number; y: number } | null {
  const area = getScreenArea(target, resolution, inputRegion);
  if (!area) {
    return null;
  }

  const x = (clientX - area.left) / area.width;
  const y = (clientY - area.top) / area.height;

  if (x < 0 || x > 1 || y < 0 || y > 1) {
    if (!clamp) {
      return null;
    }
    return { x: Math.max(0, Math.min(1, x)), y: Math.max(0, Math.min(1, y)) };
  }

  return { x, y };
}

// getScreenArea is the rectangle, in window coordinates, that shows the host
// screen.
function getScreenArea(
  target: HTMLElement,
  resolution: Resolution | null | undefined,
  inputRegion: InputRegion | null | undefined
): { left: number; top: number; width: number; height: number } | null {
  const viewport = target.parentElement;
  if (viewport?.id === 'screen-viewport' && viewport.dataset.cropped === 'true') {
    const viewportRect = viewport.getBoundingClientRect();
    if (viewportRect.width <= 0 || viewportRect.height <= 0) {
      return null;
    }
    return viewportRect;
  }

  const rect = target.getBoundingClientRect();
  if (rect.width <= 0 || rect.height <= 0) {
    return null;
  }

  const mediaSize =
    getMediaSize(target) ||
    (resolution && resolution.width > 0 && resolution.height > 0 ? resolution : null);
  if (!mediaSize) {
    return rect;
  }

  const renderedMediaRect = getRenderedMediaRect(rect, mediaSize);
  const frameContent =
    (inputRegion && getConfiguredFrameContent(inputRegion, mediaSize)) ||
    fullFrameContent(mediaSize);
  const frameScaleX = renderedMediaRect.width / mediaSize.width;
  const frameScaleY = renderedMediaRect.height / mediaSize.height;

  return {
    left: renderedMediaRect.left + frameContent.left * frameScaleX,
    top: renderedMediaRect.top + frameContent.top * frameScaleY,
    width: frameContent.width * frameScaleX,
    height: frameContent.height * frameScaleY
  };
}
