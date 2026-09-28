// Run with `pnpm test`. Node strips the types itself, so these tests need no
// build step and no test framework beyond node:test.
//
// The viewer never rotates the picture, so there is no rotation to map.

import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import { getRenderedMediaRect } from '../screen/geometry.ts';
import {
  intersectRects,
  maxUpscaledSide,
  normalizeSelection,
  selectionToFrame,
  upscaleFactor
} from './crop.ts';

describe('selectionToFrame', () => {
  it('maps the whole picture to the whole frame', () => {
    const picture = { left: 100, top: 50, width: 960, height: 540 };
    const frame = { width: 1920, height: 1080 };

    assert.deepEqual(selectionToFrame(picture, picture, frame), {
      left: 0,
      top: 0,
      width: 1920,
      height: 1080
    });
  });

  it('scales by the frame size, not the size shown on the page', () => {
    const picture = { left: 100, top: 50, width: 960, height: 540 };
    const selection = { left: 110, top: 60, width: 100, height: 50 };

    assert.deepEqual(selectionToFrame(selection, picture, { width: 1920, height: 1080 }), {
      left: 20,
      top: 20,
      width: 200,
      height: 100
    });
    assert.deepEqual(selectionToFrame(selection, picture, { width: 1280, height: 720 }), {
      left: 13,
      top: 13,
      width: 134,
      height: 67
    });
  });

  it('drops the part of a selection over a pillarbox bar', () => {
    // A 4:3 frame in a wide element: bars on the left and the right.
    const element = { left: 0, top: 0, width: 1000, height: 500 };
    const frame = { width: 1024, height: 768 };
    const picture = getRenderedMediaRect(element, frame);
    assert.equal(picture.height, 500);
    assert.ok(Math.abs(picture.left - 500 / 3) < 1e-9);

    // From inside the left bar to the middle of the picture.
    const selection = { left: 50, top: 0, width: 450, height: 250 };
    assert.deepEqual(selectionToFrame(selection, picture, frame), {
      left: 0,
      top: 0,
      width: 512,
      height: 384
    });
  });

  it('drops the part of a selection over a letterbox bar', () => {
    // A 16:9 frame in a square element: bars above and below.
    const element = { left: 20, top: 30, width: 800, height: 800 };
    const frame = { width: 1920, height: 1080 };
    const picture = getRenderedMediaRect(element, frame);
    assert.deepEqual(picture, { left: 20, top: 205, width: 800, height: 450 });

    // From the middle of the picture down into the bottom bar.
    const selection = { left: 420, top: 430, width: 400, height: 400 };
    assert.deepEqual(selectionToFrame(selection, picture, frame), {
      left: 960,
      top: 540,
      width: 960,
      height: 540
    });
  });

  it('finds nothing in a selection that covers only a bar', () => {
    const picture = getRenderedMediaRect(
      { left: 0, top: 0, width: 800, height: 800 },
      { width: 1920, height: 1080 }
    );

    assert.equal(
      selectionToFrame({ left: 0, top: 0, width: 800, height: 150 }, picture, {
        width: 1920,
        height: 1080
      }),
      null
    );
  });

  it('follows a picture the viewer has zoomed out', () => {
    // The zoom is a CSS transform, so the element's bounding box is already
    // the size on the screen: 0.75 of a 1280x720 box, centred in it.
    const element = { left: 160, top: 90, width: 960, height: 540 };
    const frame = { width: 1920, height: 1080 };
    const picture = getRenderedMediaRect(element, frame);

    assert.deepEqual(
      selectionToFrame({ left: 160, top: 90, width: 480, height: 270 }, picture, frame),
      { left: 0, top: 0, width: 960, height: 540 }
    );
  });

  it('keeps every pixel a selection touches', () => {
    // Three frame pixels for each page pixel, and a selection that starts and
    // ends part of the way into a page pixel.
    const picture = { left: 0, top: 0, width: 100, height: 100 };
    const frame = { width: 300, height: 300 };

    assert.deepEqual(
      selectionToFrame({ left: 10.5, top: 20.2, width: 5.2, height: 1 }, picture, frame),
      { left: 31, top: 60, width: 17, height: 4 }
    );
  });

  it('does not take in the next pixel at an exact edge', () => {
    // 1/3 has no exact binary form, so the edge lands a hair either side of
    // a whole frame pixel.
    const picture = { left: 0, top: 0, width: 300, height: 300 };
    const frame = { width: 100, height: 100 };

    assert.deepEqual(
      selectionToFrame({ left: 30, top: 30, width: 30, height: 30 }, picture, frame),
      { left: 10, top: 10, width: 10, height: 10 }
    );
  });

  it('refuses a picture or a frame with no size', () => {
    const selection = { left: 0, top: 0, width: 10, height: 10 };

    assert.equal(
      selectionToFrame(
        selection,
        { left: 0, top: 0, width: 0, height: 0 },
        {
          width: 1920,
          height: 1080
        }
      ),
      null
    );
    assert.equal(
      selectionToFrame(
        selection,
        { left: 0, top: 0, width: 100, height: 100 },
        {
          width: 0,
          height: 0
        }
      ),
      null
    );
  });

  it('maps a picture cropped to the input region', () => {
    // With the input region on, the element is larger than the viewport that
    // clips it. Only the visible part can be selected, and it still maps
    // through the whole picture.
    const frame = { width: 1920, height: 1080 };
    const picture = { left: -240, top: 0, width: 1920, height: 1080 };
    const viewport = { left: 0, top: 0, width: 1440, height: 1080 };
    const visible = intersectRects(picture, viewport);
    assert.deepEqual(visible, viewport);

    const selection = intersectRects({ left: -100, top: 0, width: 200, height: 100 }, visible!);
    assert.deepEqual(selectionToFrame(selection!, picture, frame), {
      left: 240,
      top: 0,
      width: 100,
      height: 100
    });
  });
});

describe('intersectRects', () => {
  it('is null for rectangles that only share an edge', () => {
    assert.equal(
      intersectRects(
        { left: 0, top: 0, width: 10, height: 10 },
        { left: 10, top: 0, width: 10, height: 10 }
      ),
      null
    );
  });
});

describe('normalizeSelection', () => {
  it('gives the same rectangle whichever way the drag went', () => {
    const expected = { left: 10, top: 20, width: 30, height: 40 };

    assert.deepEqual(normalizeSelection({ x: 10, y: 20 }, { x: 40, y: 60 }), expected);
    assert.deepEqual(normalizeSelection({ x: 40, y: 60 }, { x: 10, y: 20 }), expected);
    assert.deepEqual(normalizeSelection({ x: 40, y: 20 }, { x: 10, y: 60 }), expected);
  });
});

describe('upscaleFactor', () => {
  it('doubles a crop while the result fits', () => {
    assert.equal(upscaleFactor({ width: 400, height: 80 }), 2);
    assert.equal(upscaleFactor({ width: maxUpscaledSide / 2, height: 100 }), 2);
  });

  it('leaves a large crop alone', () => {
    assert.equal(upscaleFactor({ width: maxUpscaledSide / 2 + 1, height: 100 }), 1);
    assert.equal(upscaleFactor({ width: 100, height: maxUpscaledSide / 2 + 1 }), 1);
  });
});
