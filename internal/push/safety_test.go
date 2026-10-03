// safety_test.go: TDD for the push safety floor.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package push

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"nsl-graph/internal/configparser"
)

func TestSafetyFloor_RunRecordedOnFailure(t *testing.T) {
	store := newMemoryRunStore()
	rec := NewSafetyFloor(store, &fakeSnapshotter{path: "/tmp/backup.tar.gz"})
	rec.Record(RunParams{
		DeviceID: "R1",
		OS:       "openwrt",
		Safety:   configparser.SafetyApply,
		Renderer: "OpenWrtRenderer",
		BackupFn: func() (string, error) { return "/tmp/backup.tar.gz", nil },
		ApplyFn:  func() error { return errors.New("apply failed") },
	})

	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run recorded, got %d", len(runs))
	}
	if runs[0].ExitStatus != "failed" {
		t.Errorf("failed apply must record exit_status=failed; got %q", runs[0].ExitStatus)
	}
	if runs[0].BackupPath != "/tmp/backup.tar.gz" {
		t.Errorf("BackupPath must be present even on failure; got %q", runs[0].BackupPath)
	}
}

func TestSafetyFloor_RunRecordedOnSuccess(t *testing.T) {
	store := newMemoryRunStore()
	rec := NewSafetyFloor(store, &fakeSnapshotter{path: "/tmp/backup.tar.gz"})
	rec.Record(RunParams{
		DeviceID: "R1",
		OS:       "openwrt",
		Safety:   configparser.SafetyApply,
		Renderer: "OpenWrtRenderer",
		BackupFn: func() (string, error) { return "/tmp/backup.tar.gz", nil },
		ApplyFn:  func() error { return nil },
	})

	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].ExitStatus != "success" {
		t.Errorf("apply with nil error must record exit_status=success; got %q", runs[0].ExitStatus)
	}
}

func TestSafetyFloor_DryRunStillRecords(t *testing.T) {
	store := newMemoryRunStore()
	rec := NewSafetyFloor(store, &fakeSnapshotter{})
	rec.Record(RunParams{
		DeviceID: "R1",
		OS:       "openwrt",
		Safety:   configparser.SafetyDryRun,
		Renderer: "OpenWrtRenderer",
		ApplyFn:  func() error { return nil },
	})

	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("DryRun must still record a run; got %d", len(runs))
	}
	if runs[0].Safety != configparser.SafetyDryRun {
		t.Errorf("Safety field must reflect the actual level; got %v", runs[0].Safety)
	}
}

func TestSafetyFloor_BackupFailureAborts(t *testing.T) {
	store := newMemoryRunStore()
	rec := NewSafetyFloor(store, &fakeSnapshotter{})
	called := false
	rec.Record(RunParams{
		DeviceID: "R1",
		OS:       "openwrt",
		Safety:   configparser.SafetyApply,
		Renderer: "OpenWrtRenderer",
		BackupFn: func() (string, error) { return "", errors.New("disk full") },
		ApplyFn:  func() error { called = true; return nil },
	})

	if called {
		t.Error("ApplyFn must NOT run when backup fails")
	}
	runs := store.All()
	if len(runs) != 1 {
		t.Fatalf("backup-failure must still record a run; got %d", len(runs))
	}
	if runs[0].ExitStatus != "failed" {
		t.Errorf("backup failure must be recorded as failed; got %q", runs[0].ExitStatus)
	}
}

func TestSafetyFloor_StoreWired_PopulatesSnapshotIDAndBackupRelpath(t *testing.T) {
	runStore := newMemoryRunStore()
	snapStore := &fakeSnapshotStore{snaps: make(map[string]ConfigSnapshot)}
	backupStore := &fakeBackupStore{dir: t.TempDir()}
	sfloor := NewSafetyFloorWithStore(runStore, &Store{
		Runs:   runStore,
		Snap:   snapStore,
		Backup: backupStore,
	})

	sfloor.Record(RunParams{
		DeviceID: "R1",
		OS:       "openwrt",
		Safety:   configparser.SafetyApply,
		Renderer: "OpenWrtRenderer",
		SnapshotFn: func() ([]byte, string, error) {
			return []byte("config contents"), "conf", nil
		},
		ApplyFn: func() error { return nil },
	})

	runs := runStore.All()
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if runs[0].SnapshotID == "" {
		t.Error("SnapshotID must be populated when Store is wired")
	}
	if runs[0].BackupRelpath == "" {
		t.Error("BackupRelpath must be populated when Store is wired")
	}
}

func TestSafetyFloor_NilStore_DoesNotCrash(t *testing.T) {
	store := newMemoryRunStore()
	sfloor := NewSafetyFloor(store, &fakeSnapshotter{path: "/tmp/backup.gz"})
	sfloor.Record(RunParams{
		DeviceID: "R1",
		OS:       "openwrt",
		Safety:   configparser.SafetyApply,
		Renderer: "OpenWrtRenderer",
		BackupFn: func() (string, error) { return "/tmp/backup.gz", nil },
		ApplyFn:  func() error { return nil },
	})
	runs := store.All()
	if len(runs) != 1 || runs[0].ExitStatus != "success" {
		t.Errorf("nil-store path: got %v", runs)
	}
}

// ---------------------------------------------------------------------------
// Test doubles
// ---------------------------------------------------------------------------

type fakeSnapshotter struct {
	path string
}

func (f *fakeSnapshotter) Snapshot(deviceID string) (string, error) {
	if f.path == "" {
		return "", errors.New("no snapshot configured")
	}
	return f.path, nil
}

// Alias so the new exported name is available in this file too.
var newMemoryRunStore = NewMemoryRunStore

// ---------------------------------------------------------------------------
// Store test doubles
// ---------------------------------------------------------------------------

type fakeSnapshotStore struct {
	snaps     map[string]ConfigSnapshot
	nextID    int
	saveCalls int
}

func (f *fakeSnapshotStore) Save(snap ConfigSnapshot) (string, error) {
	f.saveCalls++
	id := fmt.Sprintf("snap-%d", f.nextID)
	f.nextID++
	snap.ID = id
	f.snaps[id] = snap
	return id, nil
}

func (f *fakeSnapshotStore) Get(id string) (ConfigSnapshot, error) {
	if s, ok := f.snaps[id]; ok {
		return s, nil
	}
	return ConfigSnapshot{}, fmt.Errorf("not found")
}

func (f *fakeSnapshotStore) List(deviceID string, limit int) ([]ConfigSnapshot, error) {
	return nil, nil
}

func (f *fakeSnapshotStore) Latest(deviceID string) (ConfigSnapshot, error) {
	return ConfigSnapshot{}, fmt.Errorf("not found")
}

type fakeBackupStore struct {
	dir string
}

func (f *fakeBackupStore) Save(deviceID, runID, ext string, raw []byte) (string, error) {
	rel := fmt.Sprintf("%s/%s.%s", deviceID, runID, ext)
	full := filepath.Join(f.dir, rel)
	os.MkdirAll(filepath.Dir(full), 0755)
	os.WriteFile(full, raw, 0644)
	return rel, nil
}

func (f *fakeBackupStore) Open(deviceID, runID string) (io.ReadCloser, string, error) {
	return nil, "", fmt.Errorf("not implemented")
}

func (f *fakeBackupStore) KeepN(deviceID string, n int) error { return nil }
