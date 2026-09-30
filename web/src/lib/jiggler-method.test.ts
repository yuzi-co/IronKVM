import assert from 'node:assert/strict';
import { test } from 'node:test';

import { jigglerChoiceFrom, jigglerChoices, jigglerMethodOf } from './jiggler-method.ts';

test('an older server with no method reads as the mouse', () => {
  assert.equal(jigglerChoiceFrom(undefined, undefined), 'mouse');
  assert.equal(jigglerChoiceFrom('mouse', ''), 'mouse');
});

test('the key method reads as its key, F15 when none is named', () => {
  assert.equal(jigglerChoiceFrom('key', 'shift'), 'shift');
  assert.equal(jigglerChoiceFrom('key', 'ctrl'), 'ctrl');
  assert.equal(jigglerChoiceFrom('key', ''), 'f15');
  assert.equal(jigglerChoiceFrom('key', 'mouse'), 'mouse');
  assert.equal(jigglerChoiceFrom('key', 'hyper'), 'mouse');
});

test('every choice survives the round trip to the request and back', () => {
  for (const choice of jigglerChoices) {
    const { method, key } = jigglerMethodOf(choice);
    assert.equal(jigglerChoiceFrom(method, key), choice);
  }
  assert.deepEqual(jigglerMethodOf('mouse'), { method: 'mouse', key: '' });
  assert.deepEqual(jigglerMethodOf('f15'), { method: 'key', key: 'f15' });
});
