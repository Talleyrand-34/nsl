// snapshot_test.go: TDD for ConfigSnapshot and SnapshotStore.
// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package push

import (
	"fmt"
	"testing"
	"time"
)

func TestConfigSnapshot_JSONRoundTrip(t *testing.T) {
	orig := ConfigSnapshot{
		ID:             "snap-001",
		DeviceID:       "R1",
		OS:             "openwrt",
		CapturedAt:     time.Date(2025, 6, 15, 10, 30, 0, 0, time.UTC),
		CapturedByRun: "push-123",
		ParserVersion:  "v1",
		Sha256Raw:      "abc123",
		Interfaces:     3,
		RawExt:         "conf",
		BackupRelpath:  "R1/push-123.conf",
	}

	data, err := orig.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON: %v", err)
	}

	loaded, err := ParseConfigSnapshot(data)
	if err != nil {
		t.Fatalf("ParseConfigSnapshot: %v", err)
	}

	if loaded.ID != orig.ID {
		t.Errorf("ID: got %q, want %q", loaded.ID, orig.ID)
	}
	if loaded.Sha256Raw != orig.Sha256Raw {
		t.Errorf("Sha256Raw: got %q, want %q", loaded.Sha256Raw, orig.Sha256Raw)
	}
	if loaded.Interfaces != orig.Interfaces {
		t.Errorf("Interfaces: got %d, want %d", loaded.Interfaces, orig.Interfaces)
	}
}

func TestMemorySnapshotStore_SaveGet(t *testing.T) {
	store := NewMemorySnapshotStore()
	snap := makeSnap("snap-1", "R1", "openwrt")

	id, err := store.Save(snap)
	if err != nil {
		t.Fatalf("Save: %v", err)
	}
	if id != "snap-1" {
		t.Errorf("Save id: got %q, want snap-1", id)
	}

	got, err := store.Get("snap-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.ID != snap.ID {
		t.Errorf("Get.ID: got %q, want %q", got.ID, snap.ID)
	}
}

func TestMemorySnapshotStore_List(t *testing.T) {
	store := NewMemorySnapshotStore()
	for i := 1; i <= 4; i++ {
		store.Save(makeSnap(fmt.Sprintf("snap-%d", i), "R1", "openwrt"))
	}

	snaps, err := store.List("R1", 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(snaps) != 4 {
		t.Errorf("List: got %d, want 4", len(snaps))
	}
}

func TestMemorySnapshotStore_List_Limit(t *testing.T) {
	store := NewMemorySnapshotStore()
	for i := 1; i <= 5; i++ {
		store.Save(makeSnap(fmt.Sprintf("snap-%d", i), "R2", "opnsense"))
	}

	snaps, err := store.List("R2", 3)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(snaps) != 3 {
		t.Errorf("List limit: got %d, want 3", len(snaps))
	}
}

func TestMemorySnapshotStore_Latest(t *testing.T) {
	store := NewMemorySnapshotStore()
	store.Save(makeSnap("oldest", "R3", "vyos"))
	store.Save(makeSnap("newest", "R3", "vyos"))

	latest, err := store.Latest("R3")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest.ID != "newest" {
		t.Errorf("Latest.ID: got %q, want newest", latest.ID)
	}
}

func TestMemorySnapshotStore_Get_NotFound(t *testing.T) {
	store := NewMemorySnapshotStore()
	_, err := store.Get("no-such")
	if err == nil {
		t.Error("Get missing: want error, got nil")
	}
}

// makeSnap is a test helper.
func makeSnap(id, deviceID, os string) ConfigSnapshot {
	return ConfigSnapshot{
		ID:             id,
		DeviceID:       deviceID,
		OS:             os,
		CapturedAt:     time.Now(),
		CapturedByRun: "push-test",
		ParserVersion:  "v1",
		Sha256Raw:      "deadbeef",
		Interfaces:     2,
		RawExt:         "conf",
		BackupRelpath:  deviceID + "/backup.conf",
	}
}
