import assert from 'node:assert/strict';
import { test } from 'node:test';
import { cbc } from '@noble/ciphers/aes.js';
import { md5 } from '@noble/hashes/legacy.js';

import { encrypt, encryptWithSalt } from './encrypt.ts';

// Made with crypto-js 4.2.0 (the library this code replaced), the passphrase
// "nanokvm-sipeed-2024" and the salt shown: AES.encrypt with the key and IV
// from kdf.OpenSSL.execute, written with format.OpenSSL. Each one also
// decrypted under CryptoJS.AES.decrypt. server/utils/encrypt_test.go reads the
// same format, so a change to these bytes is a change to the login wire format.
const vectors: [string, string, string][] = [
  ['hunter2', '0001020304050607', 'U2FsdGVkX18AAQIDBAUGB3ubnGRcdf8F+3nyu6nwhhI='],
  ['', 'ffeeddccbbaa9988', 'U2FsdGVkX1//7t3Mu6qZiLHH3wAB8hzUhovIpRmm53Q='],
  ['p@ss wörd/+=', '0123456789abcdef', 'U2FsdGVkX18BI0VniavN78oLjpxKxbqz1TGQ34/m9XQ='],
  [
    'exactly16bytes!!',
    'a1b2c3d4e5f60718',
    'U2FsdGVkX1+hssPU5fYHGKZc0cpAyNuCzERB6N5JY+OaPCATTifqr1tFJOBxFlhw'
  ],
  [
    'a-longer-password-that-spans-two-blocks',
    '8899aabbccddeeff',
    'U2FsdGVkX1+Imaq7zN3u/9ET2osCJMuUxbHph6t0fG2xZEYht8ThFWRpcamBjTZ3TSK2ajzk5q96T20A3p4/JA=='
  ]
];

test('matches the bytes crypto-js wrote for a fixed salt', () => {
  for (const [plain, saltHex, want] of vectors) {
    const salt = Uint8Array.from(Buffer.from(saltHex, 'hex'));
    assert.equal(encryptWithSalt(plain, salt), want, `plaintext ${JSON.stringify(plain)}`);
  }
});

test('encrypt uses a fresh salt and url-encodes the result', () => {
  const first = encrypt('hunter2');
  const second = encrypt('hunter2');

  assert.notEqual(first, second);
  assert.match(first, /^U2FsdGVkX1/); // base64 of "Salted__"
  assert.equal(first, encodeURIComponent(decodeURIComponent(first)));

  // Decrypt the way the server does: salted OpenSSL format, MD5 key derivation.
  const raw = Buffer.from(decodeURIComponent(first), 'base64');
  const salt = raw.subarray(8, 16);
  const pass = Buffer.from('nanokvm-sipeed-2024');
  const d1 = md5(Buffer.concat([pass, salt]));
  const d2 = md5(Buffer.concat([d1, pass, salt]));
  const d3 = md5(Buffer.concat([d2, pass, salt]));
  const derived = Buffer.concat([d1, d2, d3]);
  const plain = cbc(derived.subarray(0, 32), derived.subarray(32, 48)).decrypt(raw.subarray(16));
  assert.equal(Buffer.from(plain).toString('utf8'), 'hunter2');
});
