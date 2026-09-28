import type { FrameContent, MediaSize } from '../screen/geometry.ts';

// A rectangle in page pixels (client coordinates) or in frame pixels, by
// context.
export type Rect = FrameContent;

// Text recognition gets the crop at twice its size when the result still fits
// in this many pixels a side. Screen text is small for tesseract, which reads
// best when a capital letter is about 30 pixels tall, while a larger picture
// costs time and memory in the browser. Past this size the crop is most of a
// 4K screen, and doubling it would make an image of 64 megapixels.
export const maxUpscaledSide = 4096;

// A selection smaller than this, in page pixels, is taken as a click rather
// than a drag.
export const minimumSelectionSize = 4;

// Rounding absorbs float error, so a selection that ends exactly on a pixel
// edge does not take in the next pixel.
const edgeTolerance = 1e-6;

// intersectRects returns the part of a that lies inside b, or null if they do
// not overlap.
export function intersectRects(a: Rect, b: Rect): Rect | null {
  const left = Math.max(a.left, b.left);
  const top = Math.max(a.top, b.top);
  const right = Math.min(a.left + a.width, b.left + b.width);
  const bottom = Math.min(a.top + a.height, b.top + b.height);
  if (right <= left || bottom <= top) {
    return null;
  }

  return { left, top, width: right - left, height: bottom - top };
}

// normalizeSelection turns two corners of a drag into a rectangle, whichever
// way the drag went.
export function normalizeSelection(
  start: { x: number; y: number },
  end: { x: number; y: number }
): Rect {
  return {
    left: Math.min(start.x, end.x),
    top: Math.min(start.y, end.y),
    width: Math.abs(end.x - start.x),
    height: Math.abs(end.y - start.y)
  };
}

// selectionToFrame maps a selection in page pixels onto the pixels of a
// captured frame.
//
// picture is where the page draws the video, with any letterbox bars and any
// scaling already taken out: getRenderedMediaRect gives it for an element that
// uses object-fit: contain. The frame does not have to be the size the video
// element reports. The board sends it at the stream's size, which can differ
// from what the browser decoded, so the mapping goes through the fraction of
// the picture that the selection covers and not through a pixel ratio.
//
// The part of the selection outside the picture, over a letterbox bar, is
// dropped. The result is widened to whole pixels so that it keeps every pixel
// the selection touched, and it is null when nothing is left.
export function selectionToFrame(selection: Rect, picture: Rect, frame: MediaSize): Rect | null {
  if (picture.width <= 0 || picture.height <= 0 || frame.width <= 0 || frame.height <= 0) {
    return null;
  }

  const area = intersectRects(selection, picture);
  if (!area) {
    return null;
  }

  const scaleX = frame.width / picture.width;
  const scaleY = frame.height / picture.height;
  const left = Math.max(0, Math.floor((area.left - picture.left) * scaleX + edgeTolerance));
  const top = Math.max(0, Math.floor((area.top - picture.top) * scaleY + edgeTolerance));
  const right = Math.min(
    frame.width,
    Math.ceil((area.left + area.width - picture.left) * scaleX - edgeTolerance)
  );
  const bottom = Math.min(
    frame.height,
    Math.ceil((area.top + area.height - picture.top) * scaleY - edgeTolerance)
  );
  if (right <= left || bottom <= top) {
    return null;
  }

  return { left, top, width: right - left, height: bottom - top };
}

// upscaleFactor is how much to enlarge a crop before recognition.
export function upscaleFactor(crop: MediaSize): number {
  return crop.width * 2 <= maxUpscaledSide && crop.height * 2 <= maxUpscaledSide ? 2 : 1;
}
