import assert from 'node:assert/strict';
import { test } from 'node:test';

import { clipSize, settingsProblem, triggerKey } from './format.ts';

test('a trigger maps to its label key', () => {
  assert.equal(triggerKey('reset'), 'reset');
  assert.equal(triggerKey('signal-lost'), 'signalLost');
  assert.equal(triggerKey('signal-restored'), 'signalRestored');
  assert.equal(triggerKey('something-new'), null);
});

test('a clip size reads in kB or MB', () => {
  assert.equal(clipSize(10), '1 kB');
  assert.equal(clipSize(640_000), '640 kB');
  assert.equal(clipSize(12_345_678), '12.3 MB');
});

test('the settings are held to the server limits', () => {
  const good = { seconds: 60, postSeconds: 10, maxKbps: 0, storageMB: 200 };
  assert.equal(settingsProblem(good), null);
  assert.equal(settingsProblem({ ...good, maxKbps: 1000 }), null);
  assert.equal(settingsProblem({ ...good, postSeconds: 0 }), null);
  assert.equal(settingsProblem({ ...good, seconds: 29 }), 'seconds');
  assert.equal(settingsProblem({ ...good, seconds: 60.5 }), 'seconds');
  assert.equal(settingsProblem({ ...good, postSeconds: 31 }), 'postSeconds');
  assert.equal(settingsProblem({ ...good, maxKbps: 100 }), 'maxKbps');
  assert.equal(settingsProblem({ ...good, storageMB: 10 }), 'storageMB');
});
