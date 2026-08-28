package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"errors"
	"net/url"
)

// SecretKey is only used to prevent the data from being transmitted in plaintext.
const SecretKey = "nanokvm-sipeed-2024"

var ErrInvalidCiphertext = errors.New("invalid ciphertext")

// Decrypt reads the salted AES-256-CBC format that CryptoJS and `openssl enc`
// both write: the ASCII magic "Salted__", eight salt bytes, then the cipher
// blocks.
//
// The whole value is checked before any of it is decrypted. The library this
// used before, github.com/mervick/aes-everywhere, checked only the magic and
// the total length, so it panicked on a cipher text that was not a whole number
// of blocks, and again on a padding byte that ran off the front of the
// plaintext. Almost any garbage of the right length gives the second one,
// because the last byte of a wrong decryption is above the block size about
// fifteen times in sixteen. The login route runs this on a request body before
// it checks any credential, so an anonymous caller reached both.
func Decrypt(ciphertext string) (string, error) {
	if ciphertext == "" {
		return "", nil
	}

	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil || len(raw) < 32 || string(raw[:8]) != "Salted__" {
		return "", ErrInvalidCiphertext
	}

	encrypted := raw[16:]
	if len(encrypted)%aes.BlockSize != 0 {
		return "", ErrInvalidCiphertext
	}

	key, iv := deriveKeyAndIV(SecretKey, raw[8:16])
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", ErrInvalidCiphertext
	}

	plaintext := make([]byte, len(encrypted))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, encrypted)

	padding := int(plaintext[len(plaintext)-1])
	if padding < 1 || padding > aes.BlockSize || padding > len(plaintext) {
		return "", ErrInvalidCiphertext
	}
	for _, value := range plaintext[len(plaintext)-padding:] {
		if int(value) != padding {
			return "", ErrInvalidCiphertext
		}
	}

	return string(plaintext[:len(plaintext)-padding]), nil
}

// deriveKeyAndIV is the OpenSSL EVP_BytesToKey derivation with MD5 and one
// iteration, which is what CryptoJS writes. It gives 48 bytes: a 32 byte key
// and a 16 byte IV.
func deriveKeyAndIV(passphrase string, salt []byte) (key []byte, iv []byte) {
	derived := make([]byte, 0, 48)
	previous := []byte(nil)

	for len(derived) < 48 {
		hash := md5.New()
		hash.Write(previous)
		hash.Write([]byte(passphrase))
		hash.Write(salt)
		previous = hash.Sum(nil)
		derived = append(derived, previous...)
	}

	return derived[:32], derived[32:48]
}

// DecodeDecrypt accepts the two representations that reach the server. The web
// UI percent-encodes the base64, and a client that posts a form sends the plain
// base64, because gin has already decoded it.
//
// The raw value is tried first. A second unescape would turn a base64 '+' into
// a space, so a raw cipher text that carries one has to be recognised before
// the percent-decoding path can damage it.
func DecodeDecrypt(data string) (string, error) {
	plaintext, rawErr := Decrypt(data)
	if rawErr == nil {
		return plaintext, nil
	}

	unescaped, err := url.QueryUnescape(data)
	if err != nil || unescaped == data {
		return "", rawErr
	}

	return Decrypt(unescaped)
}
