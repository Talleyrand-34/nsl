/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

// Package secret provides the cryptographic primitives behind the credential
// Vault (see vault.go): scrypt key derivation from the master passphrase
// (DeriveKey) and AES-256-GCM encryption under an already-derived/random key
// (EncryptWithKey / DecryptWithKey). Secrets at rest (e.g. scan-profile SSH
// passwords) are encrypted under the vault's random data key, which is itself
// wrapped by the master passphrase — so the per-secret scrypt cost is paid once,
// at unlock, not on every access. A wrong key fails authentication rather than
// returning garbage.
package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	saltLen = 16
	keyLen  = 32 // AES-256

	// scrypt cost parameters (interactive-grade).
	scryptN = 32768
	scryptR = 8
	scryptP = 1
)

// ErrIncorrectPassphrase is returned when a blob cannot be authenticated with
// the supplied passphrase (wrong passphrase or tampered data).
var ErrIncorrectPassphrase = errors.New("incorrect passphrase")

func deriveKey(passphrase string, salt []byte) ([]byte, error) {
	return scrypt.Key([]byte(passphrase), salt, scryptN, scryptR, scryptP, keyLen)
}

// DeriveKey derives a 32-byte AES key from a passphrase and salt (scrypt). Used
// by the vault to wrap/unwrap its data key with the master passphrase.
func DeriveKey(passphrase string, salt []byte) ([]byte, error) {
	return deriveKey(passphrase, salt)
}

// RandomBytes returns n cryptographically-random bytes.
func RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return nil, err
	}
	return b, nil
}

// SaltLen / KeyLen expose the sizes used for vault key material.
const (
	SaltLen = saltLen
	KeyLen  = keyLen
)

// EncryptWithKey encrypts plaintext with a raw 32-byte AES key (no scrypt) and
// returns base64(nonce ‖ ciphertext). Use when the key is already derived/random
// (e.g. the vault data key), so the cost isn't paid per secret.
func EncryptWithKey(plaintext string, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("generating nonce: %w", err)
	}
	ct := gcm.Seal(nil, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ct...)), nil
}

// DecryptWithKey reverses EncryptWithKey. Returns ErrIncorrectPassphrase when the
// key is wrong or the blob was tampered with.
func DecryptWithKey(blob string, key []byte) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(blob)
	if err != nil {
		return "", fmt.Errorf("invalid secret blob: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ns := gcm.NonceSize()
	if len(raw) < ns {
		return "", errors.New("invalid secret blob: too short")
	}
	pt, err := gcm.Open(nil, raw[:ns], raw[ns:], nil)
	if err != nil {
		return "", ErrIncorrectPassphrase
	}
	return string(pt), nil
}
