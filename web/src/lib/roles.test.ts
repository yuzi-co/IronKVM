import assert from 'node:assert/strict';
import { test } from 'node:test';

import { parseRole } from './roles.ts';

test('reads the three roles', () => {
  assert.equal(parseRole('admin'), 'admin');
  assert.equal(parseRole('user'), 'user');
  assert.equal(parseRole('viewer'), 'viewer');
});

test('anything else is a user, never an administrator or a viewer', () => {
  for (const value of [undefined, null, '', 'root', 'Admin', 'VIEWER', 7, {}]) {
    assert.equal(parseRole(value), 'user', String(value));
  }
});
