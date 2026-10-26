import assert from 'node:assert/strict';
import { test } from 'node:test';

import { centred, clampPan, dragPan, fitContain, isPannable, panOffset, panTravel } from './pan.ts';

const phone = { width: 360, height: 203 };

test('at 100 % the picture has no travel and cannot be panned', () => {
  const travel = panTravel(phone, phone, 1);
  assert.deepEqual(travel, { x: 0, y: 0 });
  assert.equal(isPannable(travel), false);
});

test('at 200 % the picture travels half its enlarged overflow each way', () => {
  const travel = panTravel(phone, phone, 2);
  assert.deepEqual(travel, { x: 180, y: 101.5 });
  assert.equal(isPannable(travel), true);
});

test('a letterboxed viewport only travels on the axis that overflows', () => {
  // A 16:9 picture in a tall container: 360 x 203 inside 360 x 600.
  const travel = panTravel(phone, { width: 360, height: 600 }, 1.5);
  assert.deepEqual(travel, { x: 90, y: 0 });
});

test('a scale that is not a number gives no travel', () => {
  assert.deepEqual(panTravel(phone, phone, Number.NaN), { x: 0, y: 0 });
  assert.deepEqual(panTravel(phone, phone, 0), { x: 0, y: 0 });
});

test('the picture follows the finger and stops at its edge', () => {
  const travel = { x: 180, y: 100 };
  const moved = dragPan(centred, 90, -50, travel);
  assert.deepEqual(moved, { x: 0.5, y: -0.5 });
  assert.deepEqual(panOffset(moved, travel), { x: 90, y: -50 });

  const past = dragPan(moved, 1000, -1000, travel);
  assert.deepEqual(past, { x: 1, y: -1 });
  assert.deepEqual(panOffset(past, travel), { x: 180, y: -100 });
});

test('an axis without travel stays centred however far the drag goes', () => {
  assert.deepEqual(dragPan({ x: 0.4, y: 0.4 }, 50, 50, { x: 90, y: 0 }), {
    x: 0.4 + 50 / 90,
    y: 0
  });
});

test('a pan kept across a zoom out lands inside the smaller travel', () => {
  const pan = { x: 1, y: -1 };
  assert.deepEqual(panOffset(pan, panTravel(phone, phone, 1.5)), { x: 90, y: -50.75 });
  assert.deepEqual(panOffset(pan, panTravel(phone, phone, 1)), { x: 0, y: 0 });
});

test('a pan out of range or not a number is pulled back', () => {
  assert.deepEqual(clampPan({ x: 3, y: -Infinity }), { x: 1, y: 0 });
  assert.deepEqual(clampPan({ x: Number.NaN, y: -0.25 }), { x: 0, y: -0.25 });
});

test('a 16:9 picture on a portrait phone is letterboxed top and bottom', () => {
  const picture = fitContain({ width: 360, height: 700 }, { width: 1920, height: 1080 });
  assert.deepEqual(picture, { width: 360, height: 202.5 });
  // At 200 % it is 405 px tall in a 700 px container: no vertical travel.
  assert.deepEqual(panTravel(picture, { width: 360, height: 700 }, 2), { x: 180, y: 0 });
});

test('without a media size the picture is taken to fill the viewport', () => {
  const box = { width: 360, height: 700 };
  assert.deepEqual(fitContain(box, null), box);
  assert.deepEqual(fitContain(box, { width: 0, height: 0 }), box);
});
