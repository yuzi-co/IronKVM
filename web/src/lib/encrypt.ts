import { cbc } from '@noble/ciphers/aes.js';
import { md5 } from '@noble/hashes/legacy.js';

// This key is only used to prevent the data from being transmitted in plaintext.
const SECRET_KEY = 'nanokvm-sipeed-2024';

// The OpenSSL format the server decrypts (server/utils/encrypt.go): the magic
// "Salted__", eight salt bytes, then AES-256-CBC with PKCS#7 padding. The key
// and IV come from EVP_BytesToKey with MD5 and one iteration, as `openssl enc`
// and CryptoJS derive them from a passphrase. This is not WebCrypto because
// the board serves plain http, where crypto.subtle does not exist.
const MAGIC = new TextEncoder().encode('Salted__');
const SALT_BYTES = 8;
const KEY_BYTES = 32;
const IV_BYTES = 16;

function deriveKeyAndIV(passphrase: Uint8Array, salt: Uint8Array) {
  const derived = new Uint8Array(KEY_BYTES + IV_BYTES);
  let previous = new Uint8Array(0);
  for (let filled = 0; filled < KEY_BYTES + IV_BYTES; filled += 16) {
    const input = new Uint8Array(previous.length + passphrase.length + salt.length);
    input.set(previous);
    input.set(passphrase, previous.length);
    input.set(salt, previous.length + passphrase.length);
    previous = md5(input);
    derived.set(previous, filled);
  }
  return { key: derived.slice(0, KEY_BYTES), iv: derived.slice(KEY_BYTES, KEY_BYTES + IV_BYTES) };
}

function toBase64(bytes: Uint8Array) {
  let binary = '';
  for (const byte of bytes) {
    binary += String.fromCharCode(byte);
  }
  return btoa(binary);
}

// Split from encrypt so a test can fix the salt, which is random in use.
export function encryptWithSalt(data: string, salt: Uint8Array) {
  const { key, iv } = deriveKeyAndIV(new TextEncoder().encode(SECRET_KEY), salt);
  const body = cbc(key, iv).encrypt(new TextEncoder().encode(data));

  const out = new Uint8Array(MAGIC.length + salt.length + body.length);
  out.set(MAGIC);
  out.set(salt, MAGIC.length);
  out.set(body, MAGIC.length + salt.length);
  return toBase64(out);
}

export function encrypt(data: string) {
  const salt = crypto.getRandomValues(new Uint8Array(SALT_BYTES));
  return encodeURIComponent(encryptWithSalt(data, salt));
}
