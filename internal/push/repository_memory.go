// SPDX-License-Identifier: MIT
// repository_memory.go: in-memory PushRepository for tests.
//
// Concrete tests wire this up so the Service layer exercises the real
// flow without CloverDB. Mirrors the production seam (run + snapshot +
// backup) but holds everything in maps. Thread-safe via a single mutex.
package push

import (
	"bytes"
	"io"
	"sort"
	"sync"
)

// MemoryPushRepository is a goroutine-safe in-memory PushRepository for tests.
type MemoryPushRepository struct {
	mu        sync.Mutex
	runs      map[string][]PushRun            // deviceID -> ordered (newest last)
	snapshots map[string][]ConfigSnapshot    // deviceID -> ordered (newest last)
	backups   map[string]map[string][]byte        // deviceID -> runID -> raw bytes
}

func NewMemoryPushRepository() *MemoryPushRepository {
	return &MemoryPushRepository{
		runs:      map[string][]PushRun{},
		snapshots: map[string][]ConfigSnapshot{},
		backups:   map[string]map[string][]byte{},
	}
}

func (m *MemoryPushRepository) AppendRun(r PushRun) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.runs[r.DeviceID] = append(m.runs[r.DeviceID], r)
	return nil
}

func (m *MemoryPushRepository) ListRuns(deviceID string, limit int) ([]PushRun, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	all := m.runs[deviceID]
	out := append([]PushRun(nil), all...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].StartedAt.After(out[j].StartedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *MemoryPushRepository) LatestRun(deviceID string) (PushRun, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	all := m.runs[deviceID]
	if len(all) == 0 {
		return PushRun{}, &NotFoundError{What: "run", ID: deviceID}
	}
	return all[len(all)-1], nil
}

func (m *MemoryPushRepository) SaveSnapshot(s ConfigSnapshot) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s.ID == "" {
		// Caller left ID empty — assign the run ID, which is the
		// historical convention. Production impl may use a UUID; tests
		// just need determinism.
		s.ID = s.CapturedByRun
	}
	m.snapshots[s.DeviceID] = append(m.snapshots[s.DeviceID], s)
	return s.ID, nil
}

func (m *MemoryPushRepository) GetSnapshot(id string) (ConfigSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, list := range m.snapshots {
		for _, s := range list {
			if s.ID == id {
				return s, nil
			}
		}
	}
	return ConfigSnapshot{}, &NotFoundError{What: "snapshot", ID: id}
}

func (m *MemoryPushRepository) ListSnapshots(deviceID string, limit int) ([]ConfigSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	all := m.snapshots[deviceID]
	out := append([]ConfigSnapshot(nil), all...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].CapturedAt.After(out[j].CapturedAt)
	})
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (m *MemoryPushRepository) LatestSnapshot(deviceID string) (ConfigSnapshot, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	all := m.snapshots[deviceID]
	if len(all) == 0 {
		return ConfigSnapshot{}, &NotFoundError{What: "snapshot", ID: deviceID}
	}
	return all[len(all)-1], nil
}

func (m *MemoryPushRepository) SaveBackup(deviceID, runID, ext string, raw []byte) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.backups[deviceID] == nil {
		m.backups[deviceID] = map[string][]byte{}
	}
	cp := append([]byte(nil), raw...)
	m.backups[deviceID][runID] = cp
	return runID + "." + ext, nil
}

func (m *MemoryPushRepository) OpenBackup(deviceID, runID string) (io.ReadCloser, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	dev, ok := m.backups[deviceID]
	if !ok {
		return nil, &NotFoundError{What: "backup", ID: deviceID + "/" + runID}
	}
	raw, ok := dev[runID]
	if !ok {
		return nil, &NotFoundError{What: "backup", ID: deviceID + "/" + runID}
	}
	return io.NopCloser(bytes.NewReader(raw)), nil
}

func (m *MemoryPushRepository) KeepN(deviceID string, n int) error {
	if n <= 0 {
		return nil
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	runs := m.runs[deviceID]
	if len(runs) > n {
		m.runs[deviceID] = append([]PushRun(nil), runs[len(runs)-n:]...)
	}
	snaps := m.snapshots[deviceID]
	if len(snaps) > n {
		dropped := snaps[:len(snaps)-n]
		m.snapshots[deviceID] = append([]ConfigSnapshot(nil), snaps[len(snaps)-n:]...)
		// Best-effort: drop orphaned backups.
		devBackups := m.backups[deviceID]
		for _, s := range dropped {
			delete(devBackups, s.CapturedByRun)
		}
	}
	return nil
}