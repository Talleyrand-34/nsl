// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package secret

import (
	"errors"
	"testing"
)

func testKey(t *testing.T) []byte {
	t.Helper()
	k, err := RandomBytes(KeyLen)
	if err != nil {
		t.Fatalf("RandomBytes: %v", err)
	}
	return k
}

func TestEncryptWithKeyRoundTrip(t *testing.T) {
	key := testKey(t)
	plain := "s3cr3t-ssh-password"
	blob, err := EncryptWithKey(plain, key)
	if err != nil {
		t.Fatalf("EncryptWithKey: %v", err)
	}
	if blob == plain {
		t.Fatal("blob must not equal plaintext")
	}
	got, err := DecryptWithKey(blob, key)
	if err != nil {
		t.Fatalf("DecryptWithKey: %v", err)
	}
	if got != plain {
		t.Fatalf("round-trip mismatch: got %q want %q", got, plain)
	}
}

func TestDecryptWithKeyWrongKey(t *testing.T) {
	blob, err := EncryptWithKey("hunter2", testKey(t))
	if err != nil {
		t.Fatalf("EncryptWithKey: %v", err)
	}
	if _, err := DecryptWithKey(blob, testKey(t)); !errors.Is(err, ErrIncorrectPassphrase) {
		t.Fatalf("expected ErrIncorrectPassphrase, got %v", err)
	}
}

func TestEncryptWithKeyIsRandomized(t *testing.T) {
	key := testKey(t)
	a, _ := EncryptWithKey("same", key)
	b, _ := EncryptWithKey("same", key)
	if a == b {
		t.Fatal("two encryptions of the same input must differ (random nonce)")
	}
}

func TestEncryptWithKeyRejectsBadKeyLength(t *testing.T) {
	if _, err := EncryptWithKey("x", []byte("too-short")); err == nil {
		t.Fatal("expected error for a non-32-byte key")
	}
}
