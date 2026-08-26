# Historical snapshots — versioned push history

> **Status:** plan, not started. Builds on the push pipeline in [`../push.md`](../push.md).
> Goal: every push leaves a durable record (raw backup + parsed snapshot + audit row),
> so rollback is a "render an old snapshot" operation, not a "re-fetch and hope".

## Why

Today the safety floor records `BackupPath` as a string and never writes the file.
The audit row references a path that doesn't exist. Rollback is impossible: there is no
history. The OPNsense renderer is the only one that re-fetches observed state inside
`Render` (via `OverviewList`); OpenWrt and the stubs do not. Even if a backup file
existed, no path is wired to retrieve the parsed config that produced it.

## Outcome

After this plan lands:

- Every Apply records **three** linked artifacts:
  1. Raw device-native bytes on disk (`backups/<device>/<run-id>.<ext>`)
  2. Parsed `*ConfigData` as JSON in a new `config_snapshots` collection
  3. Audit row in `push_runs` with `snapshot_id` and `backup_relpath`
- `nsl-graph push history --device R1` lists the last N snapshots
- `nsl-graph push rollback --device R1 --to <snapshot_id>` re-renders an old snapshot
- Retention: default 50 snapshots per device; older backups are purged

## Architecture

```
   +----------------------------+
   | Fetch(device, os, creds)   |   existing — returns raw bytes
   +-------------+--------------+
                 |
                 v
   +----------------------------+
   | ParseConfig(raw)           |   existing — returns *ConfigData
   +-------------+--------------+
                 |
   +-------------+--------------+
   | BackupStore.Save(raw)      |   NEW — writes <root>/<device>/<run-id>.<ext>
   +-------------+--------------+
                 |  relpath
   +-------------+--------------+
   | SnapshotStore.Save(...)    |   NEW — config_snapshots collection
   +-------------+--------------+
                 |  snapshot_id
   +-------------+--------------+
   | SafetyFloor.Record(...)    |   extended — appends push_runs row with both IDs
   +-------------+--------------+
                 |
                 v
       Engine.Render(SafetyApply)

   push history:    list(run_id, captured_at, sha256_raw, interfaces_count)
   push rollback:   Load(snapshot_id) -> Engine.Render(SafetyApply)
```

## Types

New in `internal/push/`:

```go
// BackupStore persists raw device-native bytes.
type BackupStore interface {
    Save(deviceID, runID, ext string, raw []byte) (relpath string, err error)
    Open(deviceID, runID string) (io.ReadCloser, string, error) // reader, ext, err
    KeepN(deviceID string, n int) error // retention; n <= 0 keeps all
}

// SnapshotStore persists parsed *ConfigData.
type SnapshotStore interface {
    Save(snap ConfigSnapshot) (id string, err error)
    Get(id string) (ConfigSnapshot, error)
    List(deviceID string, limit int) ([]ConfigSnapshot, error)
    Latest(deviceID string) (ConfigSnapshot, error)
}

// ConfigSnapshot is one parsed config at one point in time.
type ConfigSnapshot struct {
    ID            string                  `json:"id"`
    DeviceID      string                  `json:"device_id"`
    OS            string                  `json:"os"`
    CapturedAt    time.Time               `json:"captured_at"`
    CapturedByRun string                  `json:"captured_by_run"`
    ParserVersion string                  `json:"parser_version"`
    Sha256Raw     string                  `json:"sha256_raw"`
    Interfaces    int                     `json:"interfaces_count"`
    RawExt        string                  `json:"raw_ext"`
    BackupRelpath string                  `json:"backup_relpath"`
    Config        *configparser.ConfigData `json:"config"`
}

// Store is the composite handed to SafetyFloor and to cmd/push.
type Store struct {
    Runs   RunStore
    Snap   SnapshotStore
    Backup BackupStore
}
```

`PushRun` gets two new fields (`SnapshotID`, `BackupRelpath`); the existing
`BackupPath` field is renamed to `BackupRelpath` so a relative path from the
configured backup root is stored, not a host-absolute path.

## Fetch orchestrator

```go
// FetchAndSnapshot is the entry point every push uses to materialise
// the historical record. Replaces the implicit "fetch + parse" pair
// currently buried in each renderer's session.
func FetchAndSnapshot(deviceID, os string, fetch func() (raw []byte, ext string, err error),
    parse func(os string, raw []byte) (*configparser.ConfigData, error),
    store *Store, runID string) (snapshotID string, err error)
```

`fetch` and `parse` are injected so this stays in `internal/push/` without
importing `internal/configparser/parsers/*` directly (which would create an
import cycle through `NewOpnsenseRendererForURL`).

## Vendors

`RawExt` is OS-specific, set by the vendor-specific fetch function:

| OS | Native format | Ext | Source |
|---|---|---|---|
| openwrt | UCI files | `conf` | `parsers.OpenWrtFetcher` already returns concatenated UCI blocks |
| opnsense | config.xml | `xml` | new `core.BackupDownload` call (wrapper already exists) |
| vyos | `/config/config.boot` | `boot` | `parsers.VyOSParser.Fetch` already returns it |
| routeros | `/export` | `rsc` | (not implemented; v1 only) |
| infix | sysrepocfg JSON | `json` | (not implemented; v1 only) |

Ponytail: only `openwrt` + `opnsense` are wired this pass. The others get the
type and store plumbing but no FetchAndSnapshot path until their parsers exist.

## Files

### New

| Path | Purpose |
|---|---|
| `internal/push/backup_store.go` | `BackupStore` interface + `FilesystemBackupStore` |
| `internal/push/backup_store_test.go` | TDD |
| `internal/push/snapshot.go` | `ConfigSnapshot` struct + JSON round-trip test |
| `internal/push/snapshot_store.go` | `SnapshotStore` interface + `MemorySnapshotStore` |
| `internal/push/snapshot_store_test.go` | TDD |
| `internal/push/composite_store.go` | `Store`, `FetchAndSnapshot` |
| `internal/push/composite_store_test.go` | TDD |
| `internal/repository/infra/cloverdb/base/config_snapshot.go` | Clover-backed `SnapshotStore` |
| `internal/repository/infra/cloverdb/base/config_snapshot_test.go` | TDD |
| `internal/repository/infra/cloverdb/base/push_run.go` | Clover-backed `RunStore` |
| `internal/repository/infra/cloverdb/base/push_run_test.go` | TDD |
| `internal/repository/infra/cloverdb/base/clover.go` | register new collection names |
| `cmd/push/history.go` | `nsl-graph push history` cobra command |
| `cmd/push/history_test.go` | TDD |
| `cmd/push/rollback.go` | `nsl-graph push rollback` cobra command |
| `cmd/push/rollback_test.go` | TDD |

### Modified

| Path | Change |
|---|---|
| `internal/push/safety.go` | `PushRun` adds `SnapshotID`, `BackupRelpath`; `SafetyFloor.Record` accepts `*Store`; `Record` calls `BackupStore.Save` + `SnapshotStore.Save` before invoking `ApplyFn` |
| `internal/push/safety_test.go` | Tests assert new fields populate; existing tests keep working with `nil` stores (DryRun + Render-only paths) |
| `internal/repository/application/app.go` | Construct `*push.Store` from the existing repository + a configured backup root |
| `cmd/push/push.go` | Register `history` and `rollback` subcommands |
| `cmd/push/opnsense_e2e_test.go` | One new test: end-to-end FetchAndSnapshot round-trip |
| `docs/push.md` | Add a "History" section linking here |

## Phases (TDD red-green, smallest diff each)

1. **`BackupStore` + `FilesystemBackupStore`.** Test: save → open returns same bytes; `KeepN` deletes oldest. ~80 LOC.
2. **`SnapshotStore` + `MemorySnapshotStore`.** Test: save → get → list → latest. ~60 LOC.
3. **`ConfigSnapshot` JSON round-trip.** Test: marshal/unmarshal preserves all fields. ~40 LOC.
4. **Wire into `SafetyFloor.Record`.** Existing e2e tests still pass with nil stores; new test asserts fields populate when stores are non-nil. ~40 LOC.
5. **Clover-backed `RunStore` + `SnapshotStore`.** Test against the existing `setupTestCloverRepository` helper. ~150 LOC.
6. **`FetchAndSnapshot` orchestrator.** Test: mock fetch + parse; assert backup file exists, snapshot row written, sha256 matches. ~80 LOC.
7. **`push history` subcommand.** Test: prints run IDs and timestamps from a fake store. ~80 LOC.
8. **`push rollback` subcommand.** Test: loads snapshot, calls `Engine.Render(SafetyApply)`, audit row tagged `rollback-from-<id>`. ~80 LOC.
9. **Retention policy.** `BackupStore.KeepN(deviceID, n)` default 50; called from `Record` after save. ~40 LOC.
10. **OPNsense FetchAndSnapshot end-to-end.** httptest server returns a config.xml; verify backup saved, snapshot retrievable. ~30 LOC.

## Out of scope

- **Restore-via-raw-bytes over SSH** — every modern NOS uses its own commit semantics; the diff-render path is the canonical rollback.
- **Compression / encryption at rest.** Backups are plain. The DB has its own storage layer; the vault handles secret material.
- **Distributed locking for concurrent pushes.** Single-operator v1; revisit if multi-tenant lands.
- **`fetch_config` from OPNsense via `core.BackupDownload`.** That's the right move for OPNsense but the typed wrapper has a typo'd endpoint name (`/interfaces/overview/list` vs `/interfaces/overview/interfaces_info`); fixing that is a separate cleanup. v1 of this plan uses `/api/interfaces/overview/interfaces_info` for observed and skips the XML backup source for OPNsense — see Risks.
- **VyOS / RouterOS / Infix FetchAndSnapshot paths.** Stubbed until their parsers exist; the type and store plumbing are vendor-agnostic.

## Risks

1. **Snapshot size.** `*ConfigData` serialised can be large. Clover accepts up to 1 MB per document; refuse larger with a typed error. Mitigation: `SnapshotStore.Save` returns `ErrSnapshotTooLarge` when `len(JSON) > 1 MB`.
2. **Parser version drift.** An old snapshot's `Config` may not parse with the current schema. Mitigation: store `ParserVersion`; on `Get`, refuse if `current > stored + 1`.
3. **Disk growth.** Backups accumulate. `KeepN` (default 50) limits per device; `push_runs` rows are **not** auto-purged (they are audit).
4. **OPNsense full-config fetch.** Real OPNsense exposes its full config only as `config.xml` via `/api/core/backup/download`. Parsing that XML into `*ConfigData` is a big job; v1 of this plan keeps OPNsense on `OverviewList`-derived snapshots and notes the full backup is **raw XML on disk only** — re-parsing it on rollback is deferred.

## Verification

Each phase ends with `go test ./internal/push/... ./internal/repository/... ./cmd/push/... -race -count=1` green. Final: `go test ./... -count=1` + manual smoke test that `push history` lists recorded runs and the backup files exist on disk.

## Open question

Default backup root on disk. Proposed: `$XDG_DATA_HOME/nsl-graph/backups` (Linux) / `~/.local/share/nsl-graph/backups` with an env override `NSL_GRAPH_BACKUP_ROOT`. Override accepted, defaults used until told otherwise.