// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package secret

import (
	"testing"
	"time"
)

func newMemVault(idle time.Duration) *Vault {
	var store string
	return NewVault(
		func() (string, error) { return store, nil },
		func(s string) error { store = s; return nil },
		idle,
	)
}

func TestVault_InitUnlockLockRoundTrip(t *testing.T) {
	v := newMemVault(0)
	if v.Initialized() {
		t.Fatal("fresh vault should not be initialized")
	}
	if err := v.Init("master-pass"); err != nil {
		t.Fatalf("init: %v", err)
	}
	if !v.Initialized() || !v.Unlocked() {
		t.Fatal("vault should be initialized and unlocked after Init")
	}
	if err := v.Init("again"); err != ErrVaultExists {
		t.Fatalf("re-init should fail with ErrVaultExists, got %v", err)
	}

	blob, err := v.Encrypt("s3cret")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	v.Lock()
	if v.Unlocked() {
		t.Fatal("vault should be locked")
	}
	if _, err := v.Encrypt("x"); err != ErrVaultLocked {
		t.Fatalf("encrypt while locked should fail, got %v", err)
	}

	if err := v.Unlock("wrong"); err != ErrIncorrectPassphrase {
		t.Fatalf("wrong passphrase should fail, got %v", err)
	}
	if err := v.Unlock("master-pass"); err != nil {
		t.Fatalf("unlock: %v", err)
	}
	got, err := v.Decrypt(blob)
	if err != nil || got != "s3cret" {
		t.Fatalf("decrypt got %q, %v", got, err)
	}
}

func TestVault_IdleAutoLock(t *testing.T) {
	v := newMemVault(20 * time.Millisecond)
	if err := v.Init("m"); err != nil {
		t.Fatalf("init: %v", err)
	}
	time.Sleep(40 * time.Millisecond)
	if v.Unlocked() {
		t.Fatal("vault should auto-lock after idle timeout")
	}
}
