// snapshot.go: ConfigSnapshot types and SnapshotStore.
package push

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// ConfigSnapshot is one parsed config at one point in time.
type ConfigSnapshot struct {
	ID             string    `json:"id"`
	DeviceID       string    `json:"device_id"`
	OS             string    `json:"os"`
	CapturedAt     time.Time `json:"captured_at"`
	CapturedByRun  string    `json:"captured_by_run"`
	ParserVersion  string    `json:"parser_version"`
	Sha256Raw      string    `json:"sha256_raw"`
	Interfaces     int       `json:"interfaces_count"`
	RawExt         string    `json:"raw_ext"`
	BackupRelpath  string    `json:"backup_relpath"`
	Config         any       `json:"config,omitempty"`
}

// MarshalJSON serialises ConfigSnapshot to JSON.
func (c ConfigSnapshot) MarshalJSON() ([]byte, error) {
	m := map[string]any{
		"id":               c.ID,
		"device_id":        c.DeviceID,
		"os":               c.OS,
		"captured_at":      c.CapturedAt.Format(time.RFC3339Nano),
		"captured_by_run":  c.CapturedByRun,
		"parser_version":   c.ParserVersion,
		"sha256_raw":       c.Sha256Raw,
		"interfaces_count": c.Interfaces,
		"raw_ext":          c.RawExt,
		"backup_relpath":   c.BackupRelpath,
		"config":           c.Config,
	}
	return json.Marshal(m)
}

// ParseConfigSnapshot deserialises a ConfigSnapshot from JSON.
func ParseConfigSnapshot(data []byte) (ConfigSnapshot, error) {
	var raw struct {
		CapturedAt string `json:"captured_at"`
		ConfigSnapshot
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return ConfigSnapshot{}, err
	}
	var snap ConfigSnapshot = raw.ConfigSnapshot
	if raw.CapturedAt != "" {
		t, err := time.Parse(time.RFC3339Nano, raw.CapturedAt)
		if err != nil {
			return ConfigSnapshot{}, err
		}
		snap.CapturedAt = t
	}
	return snap, nil
}

// SnapshotStore persists ConfigSnapshot rows.
type SnapshotStore interface {
	Save(snap ConfigSnapshot) (id string, err error)
	Get(id string) (ConfigSnapshot, error)
	List(deviceID string, limit int) ([]ConfigSnapshot, error)
	Latest(deviceID string) (ConfigSnapshot, error)
}

// MemorySnapshotStore is a thread-safe in-memory SnapshotStore for tests.
type MemorySnapshotStore struct {
	mu    sync.Mutex
	snaps []ConfigSnapshot
	index map[string]int // id -> position in snaps
}

func NewMemorySnapshotStore() *MemorySnapshotStore {
	return &MemorySnapshotStore{index: make(map[string]int)}
}

func (s *MemorySnapshotStore) Save(snap ConfigSnapshot) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.index[snap.ID]; ok {
		return "", fmt.Errorf("snapshot %q already exists", snap.ID)
	}
	s.snaps = append(s.snaps, snap)
	s.index[snap.ID] = len(s.snaps) - 1
	return snap.ID, nil
}

func (s *MemorySnapshotStore) Get(id string) (ConfigSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	pos, ok := s.index[id]
	if !ok {
		return ConfigSnapshot{}, fmt.Errorf("snapshot %q not found", id)
	}
	return s.snaps[pos], nil
}

func (s *MemorySnapshotStore) List(deviceID string, limit int) ([]ConfigSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []ConfigSnapshot
	for _, snap := range s.snaps {
		if snap.DeviceID == deviceID {
			out = append(out, snap)
		}
	}
	// Newest first
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemorySnapshotStore) Latest(deviceID string) (ConfigSnapshot, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var latest ConfigSnapshot
	var latestTime time.Time
	found := false
	for _, snap := range s.snaps {
		if snap.DeviceID == deviceID && (!found || snap.CapturedAt.After(latestTime)) {
			latest = snap
			latestTime = snap.CapturedAt
			found = true
		}
	}
	if !found {
		return ConfigSnapshot{}, fmt.Errorf("no snapshots for device %q", deviceID)
	}
	return latest, nil
}

// Sha256 returns the SHA-256 hex digest of raw bytes.
func Sha256(raw []byte) string {
	h := sha256.Sum256(raw)
	return fmt.Sprintf("%x", h)
}
