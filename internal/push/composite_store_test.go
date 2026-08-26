// composite_store_test.go: TDD for FetchAndSnapshot.
package push

import (
	"errors"
	"testing"

	"nsl-graph/internal/configparser"
)

func TestFetchAndSnapshot_Success(t *testing.T) {
	runStore := NewMemoryRunStore()
	snapStore := NewMemorySnapshotStore()
	backupStore := NewFilesystemBackupStore(t.TempDir())
	store := &Store{
		Runs:   runStore,
		Snap:   snapStore,
		Backup: backupStore,
	}

	calls := 0
	snapshotID, err := FetchAndSnapshot(
		"R1", "openwrt",
		func() ([]byte, string, error) {
			calls++
			return []byte("uci config contents"), "conf", nil
		},
		func(os string, raw []byte) (*configparser.ConfigData, error) {
			return &configparser.ConfigData{OsType: os, Interfaces: []configparser.ConfigInterface{{Name: "br0"}}}, nil
		},
		store, "push-run-001",
	)
	if err != nil {
		t.Fatalf("FetchAndSnapshot: %v", err)
	}
	if snapshotID == "" {
		t.Error("snapshotID must be non-empty")
	}
	if calls != 1 {
		t.Errorf("fetch called: got %d, want 1", calls)
	}
}

func TestFetchAndSnapshot_FetchError(t *testing.T) {
	store := &Store{
		Runs:   NewMemoryRunStore(),
		Snap:   NewMemorySnapshotStore(),
		Backup: NewFilesystemBackupStore(t.TempDir()),
	}

	_, err := FetchAndSnapshot(
		"R1", "openwrt",
		func() ([]byte, string, error) {
			return nil, "", errors.New("device unreachable")
		},
		nil,
		store, "push-run-002",
	)
	if err == nil {
		t.Error("FetchAndSnapshot must propagate fetch errors")
	}
}

func TestFetchAndSnapshot_EmptyBackupRoot(t *testing.T) {
	store := &Store{
		Runs:   NewMemoryRunStore(),
		Snap:   NewMemorySnapshotStore(),
		Backup: NewFilesystemBackupStore(""),
	}

	_, err := FetchAndSnapshot(
		"R1", "openwrt",
		func() ([]byte, string, error) {
			return []byte("data"), "conf", nil
		},
		nil,
		store, "push-run-003",
	)
	if err != nil {
		t.Errorf("FetchAndSnapshot with empty root must not fail: %v", err)
	}
}
