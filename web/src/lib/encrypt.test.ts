import assert from 'node:assert/strict';
import { test } from 'node:test';

import CryptoJS from 'crypto-js';

import { encrypt } from './encrypt.ts';

// The server decrypts what encrypt sends with the same passphrase, in the
// OpenSSL format the full crypto-js build produces. Importing only its AES
// module must not change that format.
test('produces what the full crypto-js build decrypts', () => {
  const plain = 'p@ss wörd/+=';
  const sent = decodeURIComponent(encrypt(plain));

  assert.match(sent, /^U2FsdGVkX1/); // base64 of "Salted__"
  const back = CryptoJS.AES.decrypt(sent, 'nanokvm-sipeed-2024').toString(CryptoJS.enc.Utf8);
  assert.equal(back, plain);
});
