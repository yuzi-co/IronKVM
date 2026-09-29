import assert from 'node:assert/strict';
import { test } from 'node:test';

import { keyLabel, sshCommand } from './command.ts';

test('leaves the default port out', () => {
  assert.equal(sshCommand('10.0.0.5', 22), 'ssh root@10.0.0.5');
  assert.equal(sshCommand('kvm.local', 0), 'ssh root@kvm.local');
});

test('names any other port', () => {
  assert.equal(sshCommand('10.0.0.5', 2222), 'ssh -p 2222 root@10.0.0.5');
});

test('takes an IPv6 address as it is', () => {
  assert.equal(sshCommand('fd7a:115c::1', 22), 'ssh root@fd7a:115c::1');
});

test('labels key types the way ssh-keygen does', () => {
  assert.equal(keyLabel('ssh-ed25519'), 'ED25519');
  assert.equal(keyLabel('ssh-rsa'), 'RSA');
  assert.equal(keyLabel('ecdsa-sha2-nistp256'), 'ECDSA');
  assert.equal(keyLabel('sk-ssh-ed25519@openssh.com'), 'ED25519-SK');
  assert.equal(keyLabel('sk-ecdsa-sha2-nistp256@openssh.com'), 'ECDSA-SK');
  assert.equal(keyLabel('ssh-ed25519-cert-v01@openssh.com'), 'ED25519-CERT');
});
