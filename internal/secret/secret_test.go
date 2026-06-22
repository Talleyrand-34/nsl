package secret

import (
	"errors"
	"testing"
)

func TestEncryptDecryptRoundTrip(t *testing.T) {
	plain := "s3cr3t-ssh-password"
	blob, err := Encrypt(plain, "correct horse")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if blob == plain {
		t.Fatal("blob must not equal plaintext")
	}
	got, err := Decrypt(blob, "correct horse")
	if err != nil {
		t.Fatalf("Decrypt: %v", err)
	}
	if got != plain {
		t.Fatalf("round-trip mismatch: got %q want %q", got, plain)
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	blob, err := Encrypt("hunter2", "right")
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if _, err := Decrypt(blob, "wrong"); !errors.Is(err, ErrIncorrectPassphrase) {
		t.Fatalf("expected ErrIncorrectPassphrase, got %v", err)
	}
}

func TestEncryptIsRandomized(t *testing.T) {
	a, _ := Encrypt("same", "pw")
	b, _ := Encrypt("same", "pw")
	if a == b {
		t.Fatal("two encryptions of the same input must differ (salt/nonce)")
	}
}

func TestEncryptRequiresPassphrase(t *testing.T) {
	if _, err := Encrypt("x", ""); err == nil {
		t.Fatal("expected error for empty passphrase")
	}
}
