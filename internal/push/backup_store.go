// backup_store.go: BackupStore — persists raw device-native bytes to disk.
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
	"sort"
)

// BackupStore persists raw device-native bytes.
type BackupStore interface {
	// Save writes raw to <root>/<deviceID>/<runID>.<ext> and returns the
	// relative path from root. The caller provides the OS-specific extension.
	Save(deviceID, runID, ext string, raw []byte) (relpath string, err error)
	// Open returns a reader for the backup, the extension, and an error.
	Open(deviceID, runID string) (io.ReadCloser, string, error)
	// KeepN deletes all but the newest n backups for deviceID.
	// n <= 0 is a no-op.
	KeepN(deviceID string, n int) error
}

// FilesystemBackupStore implements BackupStore on a local directory tree.
type FilesystemBackupStore struct {
	root string
}

// NewFilesystemBackupStore returns a BackupStore that writes to root.
// root is created if it does not exist.
func NewFilesystemBackupStore(root string) *FilesystemBackupStore {
	return &FilesystemBackupStore{root: root}
}

// Save writes raw to <root>/<deviceID>/<runID>.<ext>.
func (f *FilesystemBackupStore) Save(deviceID, runID, ext string, raw []byte) (string, error) {
	dir := filepath.Join(f.root, deviceID)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("mkdirAll: %w", err)
	}
	relpath := filepath.Join(deviceID, fmt.Sprintf("%s.%s", runID, ext))
	path := filepath.Join(f.root, relpath)
	if err := os.WriteFile(path, raw, 0644); err != nil {
		return "", fmt.Errorf("writeFile: %w", err)
	}
	return relpath, nil
}

// Open reads the backup file for deviceID/runID. Returns the reader, the
// extension, and an error if the file does not exist.
func (f *FilesystemBackupStore) Open(deviceID, runID string) (io.ReadCloser, string, error) {
	ents, err := os.ReadDir(filepath.Join(f.root, deviceID))
	if err != nil {
		return nil, "", fmt.Errorf("readDir: %w", err)
	}
	for _, ent := range ents {
		if ent.IsDir() || ent.Name() == "" {
			continue
		}
		name := ent.Name()
		// File is named <runID>.<ext>; match on runID prefix.
		if hasPrefix(name, runID+".") {
			ext := extFromName(name)
			path := filepath.Join(f.root, deviceID, name)
			rc, err := os.Open(path)
			if err != nil {
				return nil, "", err
			}
			return rc, ext, nil
		}
	}
	return nil, "", fmt.Errorf("backup not found for device=%s run=%s", deviceID, runID)
}

// KeepN retains the newest n backup files for deviceID and deletes the rest.
// n <= 0 is a no-op.
func (f *FilesystemBackupStore) KeepN(deviceID string, n int) error {
	if n <= 0 {
		return nil
	}
	dir := filepath.Join(f.root, deviceID)
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil // no dir = nothing to purge
	}

	// Collect regular files sorted by name descending (lexical = newest-first
	// when run IDs contain timestamps or monotonic counters).
	var files []os.DirEntry
	for _, e := range ents {
		if !e.IsDir() {
			files = append(files, e)
		}
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() > files[j].Name()
	})

	if len(files) <= n {
		return nil
	}
	for _, e := range files[n:] {
		os.Remove(filepath.Join(dir, e.Name()))
	}
	return nil
}

// hasPrefix reports whether s starts with the given prefix. It is a
// toy substitute for strings.HasPrefix to keep this file free of the
// strings package dependency.
func hasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// extFromName returns the extension of a filename (the part after the
// last dot). The input must be a non-empty name.
func extFromName(name string) string {
	for i := len(name) - 1; i >= 0; i-- {
		if name[i] == '.' {
			return name[i+1:]
		}
	}
	return ""
}
