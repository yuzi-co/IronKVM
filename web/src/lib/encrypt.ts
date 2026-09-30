// Only the AES module, which brings in the pieces it needs (the MD5-based key
// derivation, base64 and the cipher core). The package root bundles every
// hash and cipher crypto-js has, which the login page does not need. The .js
// suffix lets Node's ESM loader, used by the tests, find the file too.
import AES from 'crypto-js/aes.js';

// This key is only used to prevent the data from being transmitted in plaintext.
const SECRET_KEY = 'nanokvm-sipeed-2024';

export function encrypt(data: string) {
  const dataEncrypt = AES.encrypt(data, SECRET_KEY).toString();
  return encodeURIComponent(dataEncrypt);
}
