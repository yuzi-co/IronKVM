import assert from 'node:assert/strict';
import { test } from 'node:test';

import { syslogTargetError } from './syslog-target.ts';

// The same cases as the server's TestValidateSyslogTarget.
test('accepts what S01syslogd accepts', () => {
  for (const value of [
    '10.0.0.40',
    '10.0.0.40:5515',
    'logs',
    'logs.example.com',
    'logs.example.com:514',
    'log-host:1',
    'host:65535',
    '[::1]',
    '[::1]:514',
    '[fd00::1]:5515',
    '[FD00::a:1]',
    '[::ffff:10.0.0.1]:514',
    'host.'
  ]) {
    assert.equal(syslogTargetError(value), null, value);
  }
});

test('refuses the rest, saying why', () => {
  const cases: [string, string][] = [
    ['', 'empty'],
    ['-R', 'host'],
    ['-host', 'host'],
    ['.host', 'host'],
    ['host:', 'port'],
    ['host:0', 'port'],
    ['host:65536', 'port'],
    ['host:99999', 'port'],
    ['host:123456', 'port'],
    ['host:5x', 'port'],
    ['host:-1', 'port'],
    [':514', 'host'],
    ['::1', 'ipv6'],
    ['fd00::1:514', 'ipv6'],
    ['[]', 'brackets'],
    ['[]:514', 'brackets'],
    ['[::1]:', 'port'],
    ['[::1]:0', 'port'],
    ['[::1]:70000', 'port'],
    ['[::g]', 'brackets'],
    ['[::1]x', 'ipv6'],
    ['[::1]:514:1', 'port'],
    ['host name', 'host'],
    ['host;reboot', 'host'],
    ['host/24', 'host'],
    ['hóst', 'host'],
    ['[a]b]:5', 'brackets'],
    ['a'.repeat(300), 'long']
  ];
  for (const [value, reason] of cases) {
    assert.equal(syslogTargetError(value), reason, value);
  }
});
