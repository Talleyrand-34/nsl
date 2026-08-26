// config_snapshot_test.go: TDD for Clover-backed SnapshotStore persistence.
package basicops

import (
	"testing"
	"time"

	p "nsl-graph/internal/push"
)

func TestCloverSnapshot_SaveAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setupTestCloverRepository: %v", err)
	}
	defer cleanup()

	snap := p.ConfigSnapshot{
		ID:             "snap-001",
		DeviceID:       "R1",
		OS:             "openwrt",
		CapturedAt:     time.Now(),
		CapturedByRun:  "push-001",
		ParserVersion:  "v1",
		Sha256Raw:      "deadbeef",
		Interfaces:     3,
		RawExt:         "conf",
		BackupRelpath: "R1/push-001.conf",
	}
	id, err := repo.SaveSnapshot(snap)
	if err != nil {
		t.Fatalf("SaveSnapshot: %v", err)
	}
	if id != "snap-001" {
		t.Errorf("SaveSnapshot returned id: got %q, want snap-001", id)
	}

	got, err := repo.GetSnapshot("snap-001")
	if err != nil {
		t.Fatalf("GetSnapshot: %v", err)
	}
	if got.ID != snap.ID {
		t.Errorf("ID: got %q, want %q", got.ID, snap.ID)
	}
	if got.Sha256Raw != snap.Sha256Raw {
		t.Errorf("Sha256Raw: got %q, want %q", got.Sha256Raw, snap.Sha256Raw)
	}
	if got.Interfaces != snap.Interfaces {
		t.Errorf("Interfaces: got %d, want %d", got.Interfaces, snap.Interfaces)
	}
}

func TestCloverSnapshot_List(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setupTestCloverRepository: %v", err)
	}
	defer cleanup()

	for i := 1; i <= 5; i++ {
		repo.SaveSnapshot(p.ConfigSnapshot{
			ID:         "snap-" + string(rune('0'+i)),
			DeviceID:   "R2",
			OS:         "opnsense",
			CapturedAt: time.Now().Add(time.Duration(i) * time.Hour),
			Sha256Raw:  "abc123",
		})
	}

	snaps, err := repo.ListSnapshots("R2", 10)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snaps) != 5 {
		t.Errorf("ListSnapshots: got %d, want 5", len(snaps))
	}

	// Newest first
	if !snaps[0].CapturedAt.After(snaps[2].CapturedAt) {
		t.Error("snapshots should be sorted newest-first")
	}
}

func TestCloverSnapshot_List_Limit(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setupTestCloverRepository: %v", err)
	}
	defer cleanup()

	for i := 1; i <= 5; i++ {
		repo.SaveSnapshot(p.ConfigSnapshot{
			ID:         "snap-" + string(rune('0'+i)),
			DeviceID:   "R3",
			OS:         "vyos",
			CapturedAt: time.Now().Add(time.Duration(i) * time.Hour),
		})
	}

	snaps, err := repo.ListSnapshots("R3", 3)
	if err != nil {
		t.Fatalf("ListSnapshots: %v", err)
	}
	if len(snaps) != 3 {
		t.Errorf("ListSnapshots limit: got %d, want 3", len(snaps))
	}
}

func TestCloverSnapshot_Latest(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setupTestCloverRepository: %v", err)
	}
	defer cleanup()

	repo.SaveSnapshot(p.ConfigSnapshot{ID: "oldest", DeviceID: "R4", CapturedAt: time.Now().Add(-time.Hour)})
	repo.SaveSnapshot(p.ConfigSnapshot{ID: "newest", DeviceID: "R4", CapturedAt: time.Now()})

	latest, err := repo.LatestSnapshot("R4")
	if err != nil {
		t.Fatalf("LatestSnapshot: %v", err)
	}
	if latest.ID != "newest" {
		t.Errorf("LatestSnapshot.ID: got %q, want newest", latest.ID)
	}
}

func TestCloverSnapshot_Get_NotFound(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("setupTestCloverRepository: %v", err)
	}
	defer cleanup()

	_, err = repo.GetSnapshot("no-such")
	if err == nil {
		t.Error("GetSnapshot for missing id: want error, got nil")
	}
}
