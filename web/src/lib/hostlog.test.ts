import assert from 'node:assert/strict';
import { describe, it } from 'node:test';

import type { HostLogEntry } from '../api/hostlog.ts';
import {
  mergeEntries,
  parseFrame,
  parseList,
  plainLine,
  retryDelay,
  severityColor,
  severityName
} from './hostlog.ts';

function entry(seq: number, msg = `line ${seq}`): HostLogEntry {
  return {
    seq,
    time: '2026-10-11T10:00:00Z',
    source: '10.0.0.2',
    proto: 'syslog',
    facility: 1,
    severity: 6,
    msg
  };
}

describe('mergeEntries', () => {
  it('appends only the entries newer than the last', () => {
    const got = mergeEntries([entry(1), entry(2)], [entry(2), entry(3), entry(4)]);
    assert.deepEqual(
      got.map((e) => e.seq),
      [1, 2, 3, 4]
    );
  });

  it('returns the same array when nothing is new', () => {
    const current = [entry(5)];
    assert.equal(mergeEntries(current, [entry(4), entry(5)]), current);
  });

  it('keeps the newest max', () => {
    const got = mergeEntries([entry(1), entry(2), entry(3)], [entry(4), entry(5)], 3);
    assert.deepEqual(
      got.map((e) => e.seq),
      [3, 4, 5]
    );
  });

  it('starts from an empty list', () => {
    assert.equal(mergeEntries([], [entry(7)]).length, 1);
  });
});

describe('parseFrame', () => {
  it('reads a backlog and an entries frame', () => {
    assert.equal(parseFrame(JSON.stringify({ type: 'backlog', entries: [] }))?.type, 'backlog');
    assert.equal(
      parseFrame(JSON.stringify({ type: 'entries', entries: [entry(1)] }))?.entries.length,
      1
    );
  });

  it('refuses anything else', () => {
    assert.equal(parseFrame('not json'), null);
    assert.equal(parseFrame(JSON.stringify({ type: 'other', entries: [] })), null);
    assert.equal(parseFrame(JSON.stringify({ type: 'entries' })), null);
    assert.equal(parseFrame(new ArrayBuffer(1)), null);
  });
});

describe('parseList', () => {
  it('splits phrases on new lines only', () => {
    assert.deepEqual(parseList('Kernel panic\n  disk ([a-z]+) failed \n\nKernel panic'), [
      'Kernel panic',
      'disk ([a-z]+) failed'
    ]);
  });

  it('splits addresses on commas and spaces too', () => {
    assert.deepEqual(parseList('10.0.0.1, 10.1.0.0/16\n2001:db8::1  10.0.0.1', true), [
      '10.0.0.1',
      '10.1.0.0/16',
      '2001:db8::1'
    ]);
  });
});

describe('severity', () => {
  it('names the levels and colours the bad ones', () => {
    assert.equal(severityName(3), 'err');
    assert.equal(severityName(-1), '');
    assert.equal(severityColor(2), 'text-red-400');
    assert.equal(severityColor(4), 'text-amber-400');
    assert.equal(severityColor(6), '');
    assert.equal(severityColor(-1), '');
  });
});

describe('plainLine and retryDelay', () => {
  it('joins the parts that exist', () => {
    assert.equal(
      plainLine({ ...entry(1, 'hello'), host: 'box', tag: 'kernel' }),
      '2026-10-11T10:00:00Z 10.0.0.2 box kernel: hello'
    );
    assert.equal(plainLine(entry(1, 'bare')), '2026-10-11T10:00:00Z 10.0.0.2 bare');
  });

  it('backs off to half a minute', () => {
    assert.equal(retryDelay(0), 1000);
    assert.equal(retryDelay(3), 8000);
    assert.equal(retryDelay(20), 30_000);
  });
});
