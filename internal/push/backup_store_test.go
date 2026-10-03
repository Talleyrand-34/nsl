// backup_store_test.go: TDD for BackupStore.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package push

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFilesystemBackupStore_SaveOpen(t *testing.T) {
	root := t.TempDir()
	store := NewFilesystemBackupStore(root)

	raw := []byte("config contents here")
	relpath, err := store.Save("device-A", "run-001", "xml", raw)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}

	// relpath is relative to root
	if !strings.HasPrefix(relpath, "device-A/") || !strings.HasSuffix(relpath, ".xml") {
		t.Errorf("relpath %q not in form device-A/<run-id>.xml", relpath)
	}

	rc, ext, err := store.Open("device-A", "run-001")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer rc.Close()

	if ext != "xml" {
		t.Errorf("ext: got %q, want xml", ext)
	}

	got, err := io.ReadAll(rc)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != string(raw) {
		t.Errorf("content mismatch: got %q, want %q", got, raw)
	}
}

func TestFilesystemBackupStore_KeepN(t *testing.T) {
	root := t.TempDir()
	store := NewFilesystemBackupStore(root)

	// Save 6 backups for one device
	for i := 1; i <= 6; i++ {
		_, err := store.Save("router-B", fmt.Sprintf("run-%03d", i), "conf", rawFromInt(i))
		if err != nil {
			t.Fatalf("Save %d: %v", i, err)
		}
	}

	// KeepN(3) should delete the 3 oldest
	err := store.KeepN("router-B", 3)
	if err != nil {
		t.Fatalf("KeepN: %v", err)
	}

	// Only 3 should remain
	ents, err := os.ReadDir(filepath.Join(root, "router-B"))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	if len(ents) != 3 {
		t.Errorf("after KeepN(3): got %d files, want 3", len(ents))
	}
}

func TestFilesystemBackupStore_KeepN_ZeroOrNegativeKeepsAll(t *testing.T) {
	root := t.TempDir()
	store := NewFilesystemBackupStore(root)

	for i := 1; i <= 5; i++ {
		store.Save("r", fmt.Sprintf("run-%03d", i), "conf", rawFromInt(i))
	}

	store.KeepN("r", 0)  // no-op
	store.KeepN("r", -1) // no-op

	ents, _ := os.ReadDir(filepath.Join(root, "r"))
	if len(ents) != 5 {
		t.Errorf("KeepN(0)/(-1): got %d, want 5", len(ents))
	}
}

func TestFilesystemBackupStore_OpenNotFound(t *testing.T) {
	store := NewFilesystemBackupStore("/nonexistent")
	_, _, err := store.Open("d", "no-such-run")
	if err == nil {
		t.Error("Open for missing path: want error, got nil")
	}
}

// rawFromInt returns a []byte containing the decimal representation of n.
func rawFromInt(n int) []byte {
	return []byte(fmt.Sprintf("run-%03d", n))
}
