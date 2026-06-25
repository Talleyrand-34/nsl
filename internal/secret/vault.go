/*
Copyright © 2026 Talleyrand-34 (t34@t34.dev)

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
package secret

import (
	"encoding/base64"
	"errors"
	"strings"
	"sync"
	"time"
)

// Vault is a server-side credential vault. A single master passphrase unlocks a
// random data key (held in memory) that encrypts all stored secrets; once
// unlocked, secrets can be encrypted/decrypted without re-entering the passphrase
// until the vault is locked (explicitly or after an idle timeout). The wrapped
// data key + salt are persisted via the injected load/save callbacks so the vault
// survives restarts (but always starts locked).
//
// Persisted meta format: base64(salt) ‖ "." ‖ EncryptWithKey(dataKey, masterKey).
type Vault struct {
	mu      sync.Mutex
	key     []byte    // decrypted data key; nil when locked
	lastUse time.Time // for idle auto-lock
	idle    time.Duration
	load    func() (string, error)
	save    func(string) error
}

var (
	ErrVaultExists      = errors.New("vault already initialized")
	ErrVaultNotInit     = errors.New("vault not initialized")
	ErrVaultLocked      = errors.New("vault is locked")
	ErrVaultMetaCorrupt = errors.New("vault metadata is corrupt")
)

// NewVault builds a vault that persists its wrapped key via load/save. idle is the
// inactivity timeout after which the vault auto-locks (0 disables idle locking).
func NewVault(load func() (string, error), save func(string) error, idle time.Duration) *Vault {
	return &Vault{load: load, save: save, idle: idle}
}

// Initialized reports whether a master passphrase has been set (meta persisted).
func (v *Vault) Initialized() bool {
	m, err := v.load()
	return err == nil && m != ""
}

// expired reports (locked) if the idle timeout elapsed; caller holds the lock.
func (v *Vault) lockedLocked() bool {
	if v.key == nil {
		return true
	}
	if v.idle > 0 && time.Since(v.lastUse) > v.idle {
		v.key = nil
		return true
	}
	return false
}

// Unlocked reports whether the vault currently holds its data key (and isn't idle-expired).
func (v *Vault) Unlocked() bool {
	v.mu.Lock()
	defer v.mu.Unlock()
	return !v.lockedLocked()
}

// Init sets the master passphrase for a fresh vault, generating a random data key,
// persisting it wrapped, and leaving the vault unlocked. Errors if already set.
func (v *Vault) Init(masterPassphrase string) error {
	if masterPassphrase == "" {
		return errors.New("a master passphrase is required")
	}
	if v.Initialized() {
		return ErrVaultExists
	}
	salt, err := RandomBytes(SaltLen)
	if err != nil {
		return err
	}
	dataKey, err := RandomBytes(KeyLen)
	if err != nil {
		return err
	}
	masterKey, err := DeriveKey(masterPassphrase, salt)
	if err != nil {
		return err
	}
	wrapped, err := EncryptWithKey(string(dataKey), masterKey)
	if err != nil {
		return err
	}
	if err := v.save(base64.StdEncoding.EncodeToString(salt) + "." + wrapped); err != nil {
		return err
	}
	v.mu.Lock()
	v.key = dataKey
	v.lastUse = time.Now()
	v.mu.Unlock()
	return nil
}

// Unlock decrypts the data key with the master passphrase and holds it in memory.
func (v *Vault) Unlock(masterPassphrase string) error {
	meta, err := v.load()
	if err != nil {
		return err
	}
	if meta == "" {
		return ErrVaultNotInit
	}
	parts := strings.SplitN(meta, ".", 2)
	if len(parts) != 2 {
		return ErrVaultMetaCorrupt
	}
	salt, err := base64.StdEncoding.DecodeString(parts[0])
	if err != nil {
		return ErrVaultMetaCorrupt
	}
	masterKey, err := DeriveKey(masterPassphrase, salt)
	if err != nil {
		return err
	}
	dataKey, err := DecryptWithKey(parts[1], masterKey)
	if err != nil {
		return ErrIncorrectPassphrase
	}
	v.mu.Lock()
	v.key = []byte(dataKey)
	v.lastUse = time.Now()
	v.mu.Unlock()
	return nil
}

// Lock clears the in-memory data key.
func (v *Vault) Lock() {
	v.mu.Lock()
	v.key = nil
	v.mu.Unlock()
}

// Encrypt encrypts a secret with the vault data key. Requires an unlocked vault.
func (v *Vault) Encrypt(plaintext string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.lockedLocked() {
		return "", ErrVaultLocked
	}
	v.lastUse = time.Now()
	return EncryptWithKey(plaintext, v.key)
}

// Decrypt decrypts a secret with the vault data key. Requires an unlocked vault.
func (v *Vault) Decrypt(blob string) (string, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.lockedLocked() {
		return "", ErrVaultLocked
	}
	v.lastUse = time.Now()
	return DecryptWithKey(blob, v.key)
}
