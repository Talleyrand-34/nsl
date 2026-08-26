// safety.go: the safety floor — pre-push snapshot, audit log, abort on backup
// failure. Lives in package push because every push goes through it.
package push

import (
	"fmt"
	"sync"
	"time"

	"nsl-graph/internal/configparser"
)

// Snapshotter captures the device's current config before apply. The result is
// the path of the backup artifact — null only when no backup was needed
// (DryRun). A non-nil error means the backup failed and the apply must abort.
type Snapshotter interface {
	Snapshot(deviceID string) (string, error)
}

// PushRun is one row of the audit trail. The store is intentionally simple —
// a real CloverDB collection can wrap MemoryRunStore without changing the
// public surface.
type PushRun struct {
	ID            string
	DeviceID      string
	OS            string
	Safety        configparser.SafetyLevel
	Renderer      string
	StartedAt     time.Time
	FinishedAt    time.Time
	BackupPath    string // legacy absolute path; kept for existing callers
	SnapshotID    string // ID of the ConfigSnapshot written by the Store
	BackupRelpath string // relative path to the raw backup file
	ExitStatus    string // "success" | "failed"
	ErrorString   string
}

// RunStore persists PushRun rows. The in-memory implementation is for tests
// and for runs that pre-date the CloverDB collection migration; the production
// store wraps a CloverDB collection.
type RunStore interface {
	Append(r PushRun)
	All() []PushRun
}

// MemoryRunStore is a thread-safe in-memory RunStore for tests.
type MemoryRunStore struct {
	mu   sync.Mutex
	runs []PushRun
}

func NewMemoryRunStore() *MemoryRunStore { return &MemoryRunStore{} }

func (s *MemoryRunStore) Append(r PushRun) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.runs = append(s.runs, r)
}

func (s *MemoryRunStore) All() []PushRun {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]PushRun, len(s.runs))
	copy(out, s.runs)
	return out
}

// Store is the composite store handed to SafetyFloor and to cmd/push.
// Each field may be nil for callers that only need a subset.
type Store struct {
	Runs   RunStore     // persisted PushRun rows
	Snap   SnapshotStore // persisted ConfigSnapshot rows
	Backup BackupStore   // raw device-native backup files
}

// SafetyFloor runs a push under the snapshot-then-apply contract and records
// the result in the RunStore. A backup failure aborts before ApplyFn is called.
//
// Two constructors:
//   - NewSafetyFloor(store, snap) — legacy; uses BackupFn for absolute-path backup;
//     does NOT write snapshots or backup files.
//   - NewSafetyFloorWithStore(store, pushStore) — extended; writes both the
//     ConfigSnapshot and the backup file, and records their IDs in the PushRun.
type SafetyFloor struct {
	store       RunStore
	snapshotter Snapshotter // legacy snapshotter (absolute path); may be nil
	pushStore   *Store     // nil for legacy SafetyFloor
}

func NewSafetyFloor(store RunStore, snap Snapshotter) *SafetyFloor {
	return &SafetyFloor{store: store, snapshotter: snap}
}

// NewSafetyFloorWithStore constructs a SafetyFloor that writes snapshots and
// backup files through pushStore. pushStore.Runs must be the same store passed
// as the first argument.
func NewSafetyFloorWithStore(store RunStore, pushStore *Store) *SafetyFloor {
	return &SafetyFloor{store: store, pushStore: pushStore}
}

// RunParams are the inputs to one push run. ApplyFn is the only required
// callback — backup and render-patch are optional (DryRun has no backup).
type RunParams struct {
	DeviceID string
	OS       string
	Safety   configparser.SafetyLevel
	Renderer string
	BackupFn func() (string, error) // legacy absolute-path backup; used when pushStore is nil
	ApplyFn  func() error

	// SnapshotFn captures raw config bytes and returns (raw, ext, error).
	// Required when pushStore is non-nil; ignored when pushStore is nil.
	SnapshotFn func() (raw []byte, ext string, err error)
}

// Record runs the push once and appends a PushRun row to the store.
func (s *SafetyFloor) Record(p RunParams) {
	started := time.Now()
	run := PushRun{
		ID:        fmt.Sprintf("push-%d", started.UnixNano()),
		DeviceID:  p.DeviceID,
		OS:        p.OS,
		Safety:    p.Safety,
		Renderer:  p.Renderer,
		StartedAt: started,
	}

	// Backup + snapshot path (pushStore takes precedence).
	if p.Safety == configparser.SafetyApply {
		if s.pushStore != nil && p.SnapshotFn != nil {
			raw, ext, err := p.SnapshotFn()
			if err != nil {
				run.FinishedAt = time.Now()
				run.ExitStatus = "failed"
				run.ErrorString = fmt.Sprintf("snapshot: %v", err)
				s.store.Append(run)
				return
			}
			relpath, err := s.pushStore.Backup.Save(p.DeviceID, run.ID, ext, raw)
			if err != nil {
				run.FinishedAt = time.Now()
				run.ExitStatus = "failed"
				run.ErrorString = fmt.Sprintf("backup save: %v", err)
				s.store.Append(run)
				return
			}
			run.BackupRelpath = relpath

			snap := ConfigSnapshot{
				ID:            run.ID,
				DeviceID:      p.DeviceID,
				OS:            p.OS,
				CapturedAt:    started,
				CapturedByRun: run.ID,
				ParserVersion: "v1",
				Sha256Raw:     Sha256(raw),
				Interfaces:    0,
				RawExt:        ext,
				BackupRelpath: relpath,
			}
			id, err := s.pushStore.Snap.Save(snap)
			if err != nil {
				run.FinishedAt = time.Now()
				run.ExitStatus = "failed"
				run.ErrorString = fmt.Sprintf("snapshot save: %v", err)
				s.store.Append(run)
				return
			}
			run.SnapshotID = id

			s.pushStore.Backup.KeepN(p.DeviceID, 50)

		} else if p.BackupFn != nil {
			path, err := p.BackupFn()
			if err != nil {
				run.FinishedAt = time.Now()
				run.ExitStatus = "failed"
				run.ErrorString = fmt.Sprintf("backup: %v", err)
				s.store.Append(run)
				return
			}
			run.BackupPath = path
		}
	}

	if p.ApplyFn != nil {
		if err := p.ApplyFn(); err != nil {
			run.FinishedAt = time.Now()
			run.ExitStatus = "failed"
			run.ErrorString = err.Error()
			s.store.Append(run)
			return
		}
	}

	run.FinishedAt = time.Now()
	run.ExitStatus = "success"
	s.store.Append(run)
}
