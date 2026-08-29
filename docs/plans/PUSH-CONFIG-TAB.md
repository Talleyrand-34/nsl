# Plan: split Dashboard into DB-editing + Push Config tabs

## Why

Today `main.php` is a **CRUD dashboard**. The left column lets the operator pick
`action ∈ {get, add, update, delete}` × `entity ∈ {brand, model, device, port,
vlan, …}` and submit a form that mutates rows in the SQLite store. The right
column renders the network diagram (D2-rendered SVG) of those devices.

That's two responsibilities mashed into one screen:

1. **Edit the topology database** — which devices exist, what their ports /
   VLANs / owners / connections are.
2. **Operate the live devices** — push an intent down to a router/firewall
   with `cmd/push`, see the patch, audit the run.

The push-config work that's been landing in `internal/push/`, `cmd/push/`,
`internal/opnsenseapi/`, and `opnsense-api/` is the second concern. It has no
home yet: nothing on `main.php` runs `engine.NewDefaultEngine()`. The user has
been testing it via a `lab_e2e_test.go` and a stub `cmd/push preview` that
prints "(phase 5 stub)".

This plan **moves the dashboard to be DB-modification only** and **adds a
third tab** — **Push config** — where the device-configuration work lives.

## Scope

### In
- Carve the CRUD form (`action × entity` + the include of `$actionFiles[…]`)
  out of `main.php` and leave it on the Dashboard, trimmed to entities that
  actually mutate the topology (drop dead/scaffold entities if any).
- Add a third tab in `header.php` nav: `push.php` between Dashboard and
  Import devices.
- Build `push.php` as a two-column page that is **the write-side mirror of
  `import.php`**:
  - Left column: operations over a selected device (push intent, preview,
    dry-run, apply, rollback, audit history).
  - Right column: a **diagram of the devices the user selected**, so the
    operator can see the physical layout under the device they're about to
    reconfigure (which neighbors, which VLANs, which port it's on).
- Re-route the existing `cmd/push` cobra commands so their HTTP entry points
  sit under `/api/v1/push/*` instead of being a free-floating CLI. That gives
  the new tab something to call.

### Out (deferred)
- **Real push transport**. The renderer engine, the OPNsense API client, the
  push command shell, and the audit log are already in place from the
  `feat-push-config` branch. Wiring those to the new tab UI is in scope.
  **Re-implementing** the push command or the OPNsense API client is NOT in
  scope — those are already done.
- **Replacing the existing device-actions on main.php** (e.g. "Add Device").
  Those stay on the Dashboard as DB rows. The Push tab does NOT edit the
  device record; it operates the *configuration* of an already-existing
  device.
- **Live OPNsense**. Live testing against a real `OPNsense-gns3` clone is the
  next plan's job once this UI lands. UI plumbing can be tested with
  httptest.Server stubs in Go.

## Why two tabs, not one merged page

The two columns of `main.php` are already fighting each other:

- The CRUD form posts a *PHP form* — server-side render, full page reload.
- The diagram re-renders on every change — JS-driven, partial reload via the
  D2 image cache.

The Push tab needs **JS-driven partial updates** for the operation panel (the
operator runs a preview, sees the patch, hits Apply, sees the audit row —
without losing the diagram). Mashing JS interactivity into the PHP-form
dashboard would be a worse surface for both.

So the rule: **Dashboard = PHP form over DB. Push = JS app over the engine.**

## Tab layout

```
┌─────────────────────────────────────────────────────────────────────────┐
│  [Dashboard]  [Push config]  [Import devices]                  [theme]  │  ← appbar (header.php)
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────── operations ────────┐    ┌────── topology preview ──────────┐  │
│  │                            │    │                                  │  │
│  │  Device:   [select v]      │    │   ┌────┐                        │  │
│  │  Profile:  [select v]      │    │   │ R1 │──┐                     │  │
│  │  ──────────────────────    │    │   └────┘  │   ┌────┐            │  │
│  │  Intent editor             │    │            └───│ R2 │            │  │
│  │  ┌──────────────────────┐  │    │                └────┘            │  │
│  │  │ host: 10.0.2.1       │  │    │                                  │  │
│  │  │ hostname: gw-fic     │  │    │  Selected: R1                    │  │
│  │  │ interfaces:          │  │    │  ────────────────                │  │
│  │  │   - lan0: 10.0.2.1/24│  │    │  R1 → R2 (VLAN 10, port eth1)   │  │
│  │  │   - wan0: dhcp       │  │    │  R1 → SW1 (VLAN 10, port eth0)  │  │
│  │  │ vlans: [10,20]       │  │    │                                  │  │
│  │  └──────────────────────┘  │    │  (drawn from connections +        │  │
│  │                            │    │   ports, scoped to selection)    │  │
│  │  [Preview]  [Dry-run]      │    │                                  │  │
│  │  [Apply]    [Rollback]     │    │                                  │  │
│  │                            │    │                                  │  │
│  │  Audit log                 │    │                                  │  │
│  │  ┌──────────────────────┐  │    │                                  │  │
│  │  │ 12:01  preview R1 OK │  │    │                                  │  │
│  │  │ 12:02  apply R1 ✓    │  │    │                                  │  │
│  │  └──────────────────────┘  │    │                                  │  │
│  └────────────────────────────┘    └──────────────────────────────────┘  │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

Left column is the **engine panel**. Right column is a **scoped topology
preview**: not the full network, just the subgraph of the selected device
and its neighbors.

## What the Push tab does

The Push tab is a thin JS front-end over the engine that already exists. Five
verbs:

| Verb        | Engine call              | What it does                                  |
|-------------|--------------------------|-----------------------------------------------|
| Preview     | `engine.Preview(ctx, devID, os)`  | Render the patch (diff of intent vs observed) without applying. Show the text in a `<pre>`. |
| Dry-run     | `engine.Render(safety=SafetyDryRun, …)` | Same render, but mark "would push but don't". Default for the Preview button. |
| Apply       | `engine.Render(safety=SafetyApply, …)` + `iface.OverviewCommit` | SSH to the device, write the patch, reload. Audited. |
| Rollback    | `cmd/push rollback --device <id>` | Restore the prior snapshot from the audit log. |
| History     | `SELECT … FROM push_runs WHERE device_id=?` | Show past runs in the audit log below the buttons. |

### Files

```
frontend/php/push.php                              ← new, the tab page
frontend/php/action/push_actions.php               ← new, PHP dispatcher
frontend/php/action/push_engine.php                ← new, device/profile pickers
frontend/php/action/push_preview_panel.php         ← new, left-column panel
frontend/php/action/push_topology_panel.php        ← new, right-column panel
frontend/php/push.js                               ← new, JS client to /api/v1/push/*
frontend/php/action/_push_helpers.php              ← new, small helpers (dev_id, json call)
internal/api/push/handlers.go                      ← new, the HTTP surface
internal/api/push/preview.go                       ← new, Preview handler
internal/api/push/apply.go                         ← new, Apply handler
internal/api/push/rollback.go                      ← new, Rollback handler
internal/api/push/audit.go                         ← new, history handler
internal/repository/application/push.go            ← new, push_runs read/write
internal/repository/entities/pushrun.go            ← new, entity
internal/repository/infra/cloverdb/base/pushruns.go ← new, table + CRUD
internal/push/engine.go                            ← existing, now reachable via HTTP
cmd/push/preview.go                                ← existing, refactored to call engine
cmd/push/apply.go                                  ← existing, refactored to call engine
```

### Routes

| HTTP               | What it does                          |
|--------------------|---------------------------------------|
| `GET /push.php`    | The tab page (two columns).           |
| `POST /api/v1/push/preview`  | Body: `{device_id, os}`. Returns rendered patch text. |
| `POST /api/v1/push/apply`    | Body: `{device_id, os}`. Returns `push_run_id`. |
| `POST /api/v1/push/rollback` | Body: `{push_run_id}`. Returns status. |
| `GET /api/v1/push/history?device_id=…` | Returns last 20 rows. |
| `GET /api/v1/push/topology?device_id=…` | Returns the device + 1-hop neighbors + their connections, as JSON for the right-column renderer. |

### Data model

```go
// internal/repository/entities/pushrun.go
type PushRun struct {
    ID          string    `json:"id"`           // ULID
    DeviceID    string    `json:"device_id"`    // → devices.id
    OS          string    `json:"os"`           // "openwrt" | "opnsense" | …
    Mode        string    `json:"mode"`         // "preview" | "dry-run" | "apply" | "rollback"
    SafetyLevel string    `json:"safety_level"` // SafetyDryRun | SafetyStaged | SafetyApply
    PatchText   string    `json:"patch_text"`   // empty for rollback
    Result      string    `json:"result"`       // "ok" | "error: …"
    StartedAt   time.Time `json:"started_at"`
    FinishedAt  time.Time `json:"finished_at,omitempty"`
    SnapshotID  string    `json:"snapshot_id,omitempty"` // populated on Apply; consumed by Rollback
}
```

A new `push_runs` collection in cloverdb; one row per run. The Rollback
handler reads the latest `apply` row for a device, fetches its snapshot, and
runs `Render(safety=SafetyRollback)` against the renderer (or, for the lab
path, calls `cmd/push rollback`).

## Why I want to scope the topology preview

The full network diagram on the Dashboard is fine for "is this VLAN right?" —
you're inspecting a property of the database. On the Push tab you're about
to **mutate a device's live config**. What you actually want to see is:

- The neighbors that share a subnet / VLAN with the device (a bad VLAN push
  will blackhole them — you want them in front of you).
- The port the device is connected on, on each neighbor (so you know which
  side reloaded).
- The OS of each neighbor (because if you push VLAN 10 and the neighbor is
  OpenWRT 19.07 which doesn't support VLAN 10 on its switch chip, you need
  to know that before Apply).

So the right column is `device + 1-hop neighbors + their ports + their OS`.
That's a single SQL query against the existing connections / deviceports /
devices tables. The data is already there; this is a presentation change.

The render is **D2 inline** — no SVG cache, no full-network redraw. A small
D2 template per page that says `{device} -> {neighbor}: {port}-{port}`. D2
compiles to SVG in <100 ms for ≤50 nodes. The JS calls a `POST /api/v1/d2`
that runs the same `d2` binary the Dashboard already calls; the only
difference is the source string is generated server-side from the scope.

## Plan for the implementation (incremental, 7 phases)

Each phase is a single commit and lands in a way that keeps `go test ./...`
green and the browser usable. The user sees working software at every phase
boundary, not at the end.

### Phase 1: extract dashboard CRUD
**Goal.** Dashboard is unambiguously "edit the DB". Drop nothing yet — just
verify it still works after the carve-out.

- [ ] Read `main.php` end-to-end. Identify the entities in the
  `action × entity` grid that are actually wired (`get/add/update/delete`
  handlers in `$actionFiles`).
- [ ] Confirm `deviceport` and `modelport` have live handlers. If not,
  leave them (the existing scaffolding handles `else echo 'Select an
  action'`).
- [ ] No code change. **Verify**: open `main.php`, confirm CRUD still works
  on at least one entity (brand). Move on.

This phase is **a no-op verification step**. The point is to confirm the
baseline before splitting.

### Phase 2: add the third nav slot
**Goal.** Browser shows three tabs; clicking Push config goes to a stub page.

- [ ] Edit `frontend/php/header.php` `$navItems` to add
  `'push.php' => 'Push config'` between Dashboard and Import.
- [ ] Create `frontend/php/push.php` as a two-column skeleton:
  ```html
  <div class="grid">
      <div class="actions"><h3>operations (stub)</h3></div>
      <div class="results"><h3>topology (stub)</h3></div>
  </div>
  ```
- [ ] Use the same `<section class="actions">` and `<section class="results">`
  classes as `main.php` so it slots into the existing CSS.

Acceptance: `http://host/push.php` renders two columns; tab is highlighted in
the nav.

### Phase 3: push_runs table + entity
**Goal.** Backend has the audit-row storage ready before the UI needs it.

- [ ] Add `PushRun` to `internal/repository/entities/`.
- [ ] Add `internal/repository/infra/cloverdb/base/pushruns.go` with
  `AddPushRun` / `GetPushRun` / `ListPushRuns(deviceID, limit)`. Mirror the
  shape of `scanprofile_devices.go`.
- [ ] Add a service method `internal/repository/application/push.go`
  `RecordPushRun`, `GetPushRuns(deviceID)`. Use the same DI pattern as
  `scanrun.go`.
- [ ] TDD: write `pushruns_test.go` with `TestPushRun_AddGet`,
  `TestPushRun_List`, `TestPushRun_Result`. Run green.

Acceptance: `go test ./internal/repository/...` is green.

### Phase 4: HTTP push surface
**Goal.** The four `/api/v1/push/*` routes exist and are testable.

- [ ] Add `internal/api/push/handlers.go` with a `Register(mux)` function.
- [ ] `preview.go`: parse `{device_id, os}`, look up the device's intent
  (from `intent` field of the `devices` row, or — if absent — a synthesized
  default), look up observed state via the renderer's `Fetch()`, call
  `engine.Preview(ctx, devID, os)`, return the text.
  - **TDD first.** Write a test using `httptest.Server` + the existing
    `fakeRenderer` pattern from `internal/push/diff_engine_test.go`.
- [ ] `apply.go`: same but with `SafetyApply` and a real `*opnsense.Client`
  constructed from env vars (this is the lab-path wire-up the HANDOFF
  flagged).
- [ ] `rollback.go`: looks up the latest `push_run` for the device, reads
  its `snapshot_id`, calls `cmd/push rollback --run <id>` (or
  `engine.Render(safety=SafetyRollback, …)` if no snapshot yet — that's the
  error path).
- [ ] `audit.go`: `GET /api/v1/push/history?device_id=…`.
- [ ] `topology.go`: `GET /api/v1/push/topology?device_id=…` returns
  `{center: device, neighbors: [{device, port, neighbor, neighbor_port,
  vlan, os}, …]}`.
- [ ] Wire `Register(mux)` into `main.go`'s route table.

Acceptance: `go test ./...` green; `curl -X POST /api/v1/push/preview` with
a fake device returns 200 + text.

### Phase 5: left-column panel (operations)
**Goal.** The user can pick a device, see its current intent, click
Preview, and see the rendered patch.

- [ ] `push_preview_panel.php`: device dropdown populated by `GET /api/v1/devices`,
  OS auto-selected from the device's `os` field.
- [ ] `<textarea>` shows the intent JSON (read-only for v1).
- [ ] Preview button POSTs to `/api/v1/push/preview`, displays the returned
  text in a `<pre>` below.
- [ ] Apply, Dry-run, Rollback buttons next to it; same handler pattern.
- [ ] Audit log section at the bottom: `<div id="push-history">` populated
  by `GET /api/v1/push/history?device_id=…`. Auto-refresh every 5s.
- [ ] `push.js`: ~150 LOC. Standard fetch + DOM update, no framework.

Acceptance: pick a device, click Preview, see patch text appear. Click
Apply (against httptest stub), see audit row appended.

### Phase 6: right-column panel (scoped topology)
**Goal.** The user sees the device and its neighbors.

- [ ] `push_topology_panel.php`: empty `<div id="push-topology">`.
- [ ] `push.js`: when the device picker changes, `fetch('/api/v1/push/topology?device_id=' + id)`,
  render an inline D2 source string from the response, POST it to
  `/api/v1/d2`, embed the returned SVG.
- [ ] Below the diagram, a small `<table>` listing the neighbors with their
  OS / port / VLAN. Highlight neighbors whose OS would not accept the
  intended VLAN (a known unsupported list lives in
  `internal/push/unsupported.go` — but that's a follow-up; for now just
  show the table).

Acceptance: pick R1 → diagram shows R1 + its neighbors + edges; pick R3 →
diagram updates without page reload.

### Phase 7: end-to-end smoke
**Goal.** No regression; the feature works end-to-end against httptest.

- [ ] Add `cmd/push/e2e_smoke_test.go` that:
  1. Spins up the API with `httptest.Server`.
  2. Adds a fake device to the store.
  3. POSTs to `/api/v1/push/preview`, asserts non-empty patch text.
  4. POSTs to `/api/v1/push/apply`, asserts a `push_run_id` returns.
  5. GETs `/api/v1/push/history`, asserts the run is listed.
- [ ] Run `go test ./...`. Must be green.
- [ ] Boot the PHP server, open `push.php`, click through the four buttons
  against a stub device. Screenshot the audit row in the chat log.

Acceptance: `go test ./...` green; manual click-through works.

## What this plan is NOT doing (the user's instinct to be cautious)

- It does NOT re-implement the push engine, the OPNsense client, the audit
  store, or the cmd/push cobra commands. All of those exist; the plan
  plumbs UI → HTTP → existing engine.
- It does NOT replace any Dashboard functionality. CRUD on devices, ports,
  VLANs, etc. stays on `main.php`. The Dashboard is being **clarified**, not
  trimmed.
- It does NOT add a new framework (no React, no Svelte). `push.js` is plain
  ES6 fetch + DOM. The Dashboard already uses inline ES6; consistency wins.

## Risks

- **`push.php` styling drift from `main.php`.** Mitigation: reuse the same
  `<div class="grid">` / `<section class="actions">` /
  `<section class="results">` classes. If a class is missing, copy it from
  `main.php` rather than inventing.
- **Engine wiring surprises.** `NewDefaultEngine()` returns an engine whose
  `opnsense` entry is currently a stub. The HANDOFF explicitly flagged
  wiring the typed `*opnsense.Client` into the default engine when
  `OPNSENSE_*` env vars are set. That's a Phase 4 dependency and is the
  riskiest single change. Mitigation: env-driven upgrade is opt-in
  (default still returns ErrUnsupported), so the unit tests for the engine
  pass with the env unset.
- **D2 re-render latency.** D2 compile is the slowest link in the topology
  preview. Mitigation: cache the SVG keyed by `device_id` for 30 s
  (server-side `If-Modified-Since` header on the rendered SVG). Out of scope
  for this plan; flag as a follow-up.

## Acceptance for the plan itself

The plan is accepted when the user agrees that:

1. The Dashboard → DB-only carve-out is a no-op for them (they confirm they
   still see CRUD on `main.php`).
2. The third tab name is **"Push config"** (or whatever they prefer; this
   doc uses "Push config" as the default).
3. The two-column layout (operations left, topology right) is the right
   shape.
4. The 7-phase incremental order is acceptable; they can ship or stop after
   any phase.
5. The Push tab does NOT modify the device row in the DB. It only operates
   the device's live config. (If they want device-editing on the Push tab,
   that's a separate plan.)

## File-by-file diff budget (rough)

```
+ frontend/php/push.php                              ~30 lines
+ frontend/php/action/push_actions.php               ~20
+ frontend/php/action/push_engine.php                ~40
+ frontend/php/action/push_preview_panel.php         ~120
+ frontend/php/action/push_topology_panel.php        ~40
+ frontend/php/push.js                               ~150
+ internal/api/push/handlers.go                      ~50
+ internal/api/push/preview.go                       ~80
+ internal/api/push/apply.go                         ~100
+ internal/api/push/rollback.go                      ~70
+ internal/api/push/audit.go                        ~40
+ internal/api/push/topology.go                      ~80
+ internal/repository/entities/pushrun.go            ~25
+ internal/repository/infra/cloverdb/base/pushruns.go ~120
+ internal/repository/application/push.go            ~50
+ test files                                         ~300
+ 1 line in frontend/php/header.php
~ 0 lines changed in main.php
```

≈ **+1,300 LOC, 0 deletions.** The Dashboard is untouched.

## Next step

Phase 1 (no-op verify) + Phase 2 (third nav slot + skeleton page) is one
small commit. Land it and stop for user feedback before Phase 3.
