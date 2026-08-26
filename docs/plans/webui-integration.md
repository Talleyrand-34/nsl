# WebUI integration — bring the push + observability pipeline into the API

> **Status:** plan, rewritten 2026-08-26 after the user asked for a clean
> architecture. **The previous version's "shell out to the cobra binary"
> choice is wrong** — it would have made the API a thin wrapper around
> a CLI process, with subprocess cost, opaque errors, and no type safety.
>
> The CLI is a thin cobra wrapper around the same Go service the API
> uses. One canonical surface, two callers.

## Goal

Make the push + observability pipeline reachable through the existing
HTTP API (`internal/api/<area>/handlers.go`) so the webui can drive
it without subprocess, then rebuild the webui's push surfaces against
the new endpoints. The CLI (`cmd/push/*`) keeps working but becomes a
cobra layer over the same `internal/push.Service`.

## Architecture

```
                +---------------------+
                |   cmd/push/* (cobra)|   6 commands: device, devices, preview,
                |                     |   history, rollback, plus dry-run wrappers.
                +----------+----------+
                           |  calls
                           v
                +---------------------+
                |  internal/push.Service|<-- canonical push primitive.
                |  (NEW)               |    Push / Preview / History / Rollback
                +-----+----------+-----+    LiveConfig / FetchAndSnapshot
                      |          |
                      v          v
        +-------------+--+    +--+-------------+
        | Engine +      |    | SafetyFloor + |   existing pieces, unchanged.
        | Renderers     |    | Store (Runs + |
        |               |    |   Snap + Backup)|
        +---------------+    +----------------+
                      |          |
                      v          v
                +---------------------+
                |   HTTP handlers     |
                |   internal/api/push/ |   4 routes: /push/preview, /push/device,
                |                     |   /push/history, /push/rollback
                +----------+----------+
                           |  HTTP
                           v
                +---------------------+
                |   frontend/php/      |   push.php, push-detail.php,
                |   (PHP/JS)            |   push-history.php, push-diff.php
                +---------------------+
```

**One canonical surface.** `internal/push.Service` is the single entry
point. Cobra wraps it for the CLI; HTTP handlers wrap it for the API.
Both paths go through the same `Engine.Render` / `SafetyFloor.Record`
/ `Store.Snap.List` code. Drift is impossible because there's only one
implementation.

## Why this is better than the previous plan

| Concern | Previous plan (subprocess) | This plan (in-process service) |
|---|---|---|
| Per-request cost | 50-100ms fork+exec; 5-30s push payload | in-process; microsecond dispatcher overhead |
| Type safety | subprocess returns bytes; parse `stdout` | typed return values |
| Error handling | opaque exit code + stderr string | structured `error` with wrap chain |
| Test coverage | E2E-only via shell scripts | unit-testable via fake Store / Engine |
| Argument parsing | duplicated (CLI flags + JSON body) | single Go call signature |
| Cancellation | process kill | `context.CancelFunc` propagates |
| Vault lifecycle | subprocess inherits handler's env | service takes `*secret.Vault` once |
| Concurrency | each push is a process | per-device lock via service mutex |

The subprocess approach was tempting because it "reuses the CLI" —
but the CLI is now the thin wrapper, not the source of truth.

## Outcome

After this plan lands:

- **One Go service** (`internal/push.Service`) exposes `Push`,
  `Preview`, `History`, `Rollback`, `FetchAndSnapshot`, `LiveConfig`.
- **Four HTTP endpoints** (`internal/api/push/handlers.go`) drive
  the service for the webui.
- **Six CLI commands** (`cmd/push/*.go`) keep working — they call the
  same service.
- **Frontend pages** consume the API, not the CLI.

## API surface (canonical)

`internal/push.Service`:

```go
type Service struct {
    repo     repository.PushRepository  // runs + snapshots (CloverDB-backed in prod)
    engine    *Engine                   // renderer registry + per-vendor typed clients
    floor     *SafetyFloor              // backup + snapshot + audit
    vault     *secret.Vault             // decrypt SSH/API credentials
    renderer  string                    // "typed" or "lazy"; informational, in audit row
    perDevLocks *sync.Map               // deviceID → *sync.Mutex
}

type PushRequest struct {
    DeviceID  string
    OS        string
    Safety    configparser.SafetyLevel // DryRun / Staged / Apply
    Intent    *configparser.ConfigData
    Renderer  string                  // optional override
}

type PushResult struct {
    RunID        string
    SnapshotID   string
    BackupRelpath string
    DiffText     string                  // the rendered patch; identical to Preview()
    ExitStatus   string                  // "success" | "failed"
    ErrorString  string
    StartedAt    time.Time
    FinishedAt   time.Time
}

type SnapshotInfo struct {
    ID            string    `json:"id"`
    DeviceID      string    `json:"device_id"`
    OS            string    `json:"os"`
    CapturedAt    time.Time `json:"captured_at"`
    CapturedByRun string    `json:"captured_by_run"`
    BackupRelpath string    `json:"backup_relpath"`
    Interfaces    int       `json:"interfaces_count"`
}

// Service.Push is the canonical entry point: snapshot the device,
// write the snapshot, run diff, run render, commit per-area, audit.
// Safety=DryRun skips the network round-trip entirely. Staged writes
// the patch to a temp file (currently no-op for OPNsense; OpenWrt
// path stores it via the SSH session). Apply mutates.
func (s *Service) Push(ctx context.Context, req PushRequest) (PushResult, error)

// Service.Preview returns the rendered patch text without snapshotting
// or applying. Always read-only.
func (s *Service) Preview(ctx context.Context, deviceID, os string) (string, error)

// Service.History lists the last N snapshots for a device.
func (s *Service) History(ctx context.Context, deviceID string, limit int) ([]SnapshotInfo, error)

// Service.Rollback loads a snapshot and re-renders it via the typed
// renderer. Apply-only; no DryRun (rollback is by definition a write).
func (s *Service) Rollback(ctx context.Context, deviceID, snapshotID string) (PushResult, error)

// Service.LiveConfig fetches the device's current running config and
// returns the parsed *ConfigData. Used by the device-detail page to
// show what's actually on the box right now (not the last snapshot).
func (s *Service) LiveConfig(ctx context.Context, deviceID, os string) (*configparser.ConfigData, error)

// Service.FetchAndSnapshot is the unified fetch+parse+save entry point
// used by both the snapshot-on-push path and the device-detail page.
func (s *Service) FetchAndSnapshot(ctx context.Context, deviceID, os, runID string) (string, error)
```

## HTTP endpoints

`internal/api/push/`:

```
POST /push/preview        — body: {device_id, os}
                            -> {diff_text, sections: [...]} (per-area sections for UI)
GET  /push/history?device=R1&limit=50
                            -> {snapshots: [...]}
POST /push/device          — body: {device_id, os, safety, confirm}
                            -> PushResult JSON
POST /push/rollback         — body: {device_id, snapshot_id, confirm}
                            -> PushResult JSON
GET  /push/live?device=R1   — body: {device_id, os}
                            -> {config: *ConfigData JSON}
GET  /push/status?device=R1 — lightweight poll, last-push-at + last result
POST /push/diff             — body: {device_id, os, sections: ["vlan","route",...]}
                            -> per-section diff text
```

`confirm` is required on Apply/rollback: `"PUSH R1"` / `"ROLLBACK R1"`.
Same string format the CLI prints for confirmation prompts.

## Frontend pages

```
frontend/php/push.php          (rewrite)
   - index: per-zone device list with last-push status
   - per-device "Open" → push-detail.php

frontend/php/push-detail.php    (new)
   - per-device panel: VLANs, routes, NTP, banner, syslog, SNMP, LLDP
     - read from /push/live?device=R1 (LiveConfig)
     - each section has a "Diff" + "Push" button
   - "Push all" at top with safety-level selector + typed-confirm

frontend/php/push-history.php    (new)
   - /push/history?device=R1
   - per-row "Roll back to this" button

frontend/php/push-diff.php       (new)
   - /push/diff?device=R1&sections=vlan,route
   - shows per-section patch text

frontend/php/push.js             (rewrite)
   - thin fetch wrappers, no subprocess anywhere
```

## Files

### New

| Path | Purpose |
|---|---|
| `internal/push/service.go` | `Service` struct + `Push`/`Preview`/`History`/`Rollback`/`LiveConfig`/`FetchAndSnapshot` methods |
| `internal/push/service_test.go` | TDD for every Service method using fake `PushRepository` + fake `Engine` + fake `Floor` |
| `internal/api/push/handlers.go` | 6 HTTP handlers + route registration |
| `internal/api/push/handlers_test.go` | httptest-driven coverage for every handler |
| `internal/repository/domain/push.go` | `PushRepository` interface (runs + snapshots) — wraps existing CloverDB code |
| `frontend/php/push-detail.php` | per-device detail page |
| `frontend/php/push-history.php` | snapshot history page |
| `frontend/php/push-diff.php` | diff preview page |
| `frontend/php/action/push*.php` | server-side actions (5 files: preview, device, history, rollback, live) |

### Modified

| Path | Change |
|---|---|
| `cmd/push/push.go` | add `Service` initialization; subcommands call `Service.X` instead of building their own Engine/SafetyFloor |
| `cmd/push/device.go` | rewire to `Service.Push` |
| `cmd/push/devices.go` | rewire to `Service.Preview` / `Service.LiveConfig` |
| `cmd/push/preview.go` | rewire to `Service.Preview` |
| `cmd/push/history.go` | rewire to `Service.History` |
| `cmd/push/rollback.go` | rewire to `Service.Rollback` |
| `internal/api/server.go` | register `push.RegisterRoutes(r, service, pushService)` |
| `frontend/php/push.php` | update supported list, dead-button tooltip, link to detail page |
| `frontend/php/main.php` | add "Push" nav link |
| `internal/push/safety.go` | `SafetyFloor.Record` keeps its API; `Service` wraps it |
| `internal/push/composite_store.go` | rename / add `PushRepository` interface |
| `internal/repository/infra/cloverdb/base/*.go` | implement `PushRepository` against existing collections (or add a thin adapter) |

## `PushRepository` — the storage seam

The push pipeline needs three stores today (runs, snapshots, backup
files). One Go interface keeps them swappable in tests:

```go
// internal/repository/domain/push.go
type PushRepository interface {
    AppendRun(r push.PushRun) error
    ListRuns(deviceID string, limit int) ([]push.PushRun, error)
    LatestRun(deviceID string) (push.PushRun, error)
    SaveSnapshot(s push.ConfigSnapshot) (id string, err error)
    GetSnapshot(id string) (push.ConfigSnapshot, error)
    ListSnapshots(deviceID string, limit int) ([]push.ConfigSnapshot, error)
    LatestSnapshot(deviceID string) (push.ConfigSnapshot, error)
    SaveBackup(deviceID, runID, ext string, raw []byte) (relpath string, err error)
    OpenBackup(deviceID, runID string) (io.ReadCloser, error)
    KeepN(deviceID string, n int) error
}
```

Production impl: a thin adapter over the existing
`internal/repository/infra/cloverdb/base` collection wrappers plus
the filesystem backup store. Test impl: an in-memory map.

## Per-device lock

Concurrent pushes to the same device race: snapshot saved by push A,
diff computed by push B, both apply at once. Service has a
`sync.Map[deviceID]*sync.Mutex`; every Push/Rollback takes the lock for
the duration of the call. Network round-trip is 5-30s; locking for
that long is acceptable for a single-user tool.

## Vault lifecycle

The HTTP server's process owns the vault. The vault auto-locks after
`vaultIdleTimeout` (15 min). Every push request that needs credentials
calls `vault.Unlock()` if locked; if the vault is locked and the request
comes back with `confirm`, the unlock flow must already have run
(handled by a separate `/vault/unlock` endpoint that already exists).

The Service does NOT take a `*secret.Vault` directly. It takes a
`CredentialResolver` interface — a thin abstraction over the vault
that's mockable:

```go
type CredentialResolver interface {
    SSH(deviceID string) (configparser.SSHCredentials, error)
    API(deviceID, os string) (baseURL, key, secret string, err error)
}
```

Production impl reads from `vault.Decrypt(...)`. Test impl returns
fixed creds. This is the only piece of vault wiring that needs new
abstraction.

## Phasing (TDD red-green)

Each phase lands with `go test ./... -race -count=1` green.

1. **`internal/push/Service` skeleton + `PushRepository` interface.**
   Service has the field types but methods panic. `PushRepository`
   declared; in-memory test impl provided. ~120 LOC.
2. **`Service.Preview`** — pure engine call, no vault, no network
   fetch. Tests use a stub Engine + the existing `MemoryRunStore`.
3. **`Service.LiveConfig`** — fetch observed from device via vault
   credentials. Tests use a fake `CredentialResolver` and the existing
   httptest pattern. Real-device path passes through `Engine` → renderer
   `fetchObserved`.
4. **`Service.FetchAndSnapshot`** — wraps LiveConfig + writes
   snapshot + writes backup. Tests use a fake `PushRepository`.
5. **`Service.History`** — wraps `repo.ListSnapshots`.
6. **`Service.Push` (DryRun path)** — wraps Service.Preview +
   Service.FetchAndSnapshot, but skips render.
7. **`Service.Push` (Apply path)** — calls SafetyFloor.Record with
   proper SnapshotFn. Per-device lock. Apply audit row.
8. **`Service.Rollback`** — load snapshot, call renderer.Render.
9. **HTTP handlers** (`/push/preview`, `/push/history`, `/push/device`,
   `/push/rollback`, `/push/live`, `/push/diff`, `/push/status`).
   Handlers are thin: parse JSON body, call Service, render JSON.
10. **`cmd/push/*` rewiring** — every subcommand now calls Service.
    Existing `lab_e2e_test.go` + `opnsense_e2e_test.go` +
    `observability_e2e_test.go` keep passing; they test through Service
    transparently.
11. **Frontend `push.php` rewrite** — index page, supported-list fix,
    device status indicator. ~80 LOC.
12. **Frontend `push-detail.php`** — per-device sections, each with
    Diff + Push. ~250 LOC.
13. **Frontend `push-history.php`** — snapshot list with rollback
    buttons. ~80 LOC.
14. **Frontend `push-diff.php`** — preview-only page. ~60 LOC.
15. **End-to-end manual smoke** — full round-trip from index →
    detail → diff → apply → history → rollback. Browser-side, no
    automated tests (no PHP test runner wired).

## Verification

- Each phase: `go test ./... -race -count=1` green.
- Phase 7 (Push Apply): a full round-trip against an httptest-driven
  OPNsense lab; verify the audit row, the snapshot row, and the
  per-area commit sequence.
- Phase 10 (CLI rewire): `cmd/push/lab_e2e_test.go`,
  `opnsense_e2e_test.go`, `observability_e2e_test.go`, `push_test.go`
  all keep passing. **If they break, the rewire is wrong.**
- Phase 15 (manual): the index page lists devices with last-push
  status; clicking a device shows the device detail with all sections;
  pushing adds a row to history; rollback reverts.

## Out of scope

- **PHP unit tests.** Manual smoke per phase.
- **Multi-device atomic push.** Per-device lock only.
- **Websocket streaming.** Polling-only status.
- **Auth on push endpoints.** The vault gate is sufficient for v1.
- **Audit row drill-down UI.** `/push/history` lists rows; the
  `ErrorString` field carries enough context for v1.
- **Per-section push without lockstep.** A per-device mutex covers
  the device; sections don't need their own lock.

## Risks

1. **Credential resolution timeout.** Vault decrypt is fast (<1ms)
   but a missing profile blocks every push for that device. Service
   returns a clean `ErrCredentialsUnavailable` and the HTTP handler
   surfaces it as 503 with a "set up a profile first" hint.
2. **Live config fetch on dead device.** `Service.LiveConfig` blocks
   for the device's SSH/API timeout (30s default). The HTTP handler
   has its own 60s deadline. Polling the device-status endpoint
   during the fetch is too much; just let the request time out cleanly.
3. **In-process state vs request lifecycle.** The Service is a long-
   lived singleton; its `perDevLocks` map grows under load. Use
   `sync.Map` with delete-on-release so unlocked devices don't leak.
4. **JSON marshalling of `*ConfigData`.** The model has pointer
   fields (`*ConfigNTPConfig` etc.). The existing `MarshalJSON` for
   `ConfigSnapshot` is hand-rolled. Webui deserialisation tests are
   needed to catch new field omissions.
5. **Push cancellation.** A handler-side `context.CancelFunc` lets the
   operator abort a hung push. The Service must honour the context
   in every blocking call (renderer.Render, snapshot save, vault
   decrypt — last is non-blocking; first two must).

## Recommendation

Start with **phase 1** (`Service` skeleton + `PushRepository`).
Smallest diff, biggest architecture benefit. Everything after this
phase plugs into a typed surface.

Tell me to start phase 1 (or pick a different phase) and I'll proceed.