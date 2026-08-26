# Push — bidirectional control of network devices

> **Status:** v1 ships **OpenWrt** only, **with proof on the live GNS3 lab**
> (`10.0.50.11`, OpenWrt-1, fetched 6898 bytes, dry-run round-trip green).
> Other vendors are stubbed and return `ErrUnsupported`.

## The contract

Every push passes a `SafetyLevel` gate. The contract is:

| Level | Behaviour | Mutates the device? |
| --- | --- | --- |
| `dry-run` *(default, zero value)* | Report the diff and the rendered patch | **No** |
| `staged` | Write the patch to a temp path on the device | **No** |
| `apply` | Write, syntax-check, reload, audit-log | **Yes** |

The CLI and HTTP defaults are `dry-run`. Apply requires an explicit `--apply`
flag (or an explicit `safety=apply` JSON field). A miscompiled caller that
drops the flag falls back to *report only*, not *apply silently* — pinned by
[`TestSafetyLevel_ZeroIsDryRun`](../internal/configparser/safety_test.go).

## Pipeline

```
intended config   observed config
        │              │
        └──────┬───────┘
               ▼
       YANG gate (Phase 7)
               │
               ▼
       renderer.Diff  (per-vendor grammar)
               │
               ▼
       renderer.Render  (Session, with SafetyLevel)
               │
               ▼
       safety floor  →  snapshot?  →  apply gate (SafetyLevel == Apply)
                                          │
                              ┌───────────┼───────────┐
                              no          yes          yes
                              │           │            │
                          report     safety floor    audit row
                              ▼       (Phase 6)        │
                          exit 0                       ▼
                                                   push_runs row
```

## Three safeguards (Phase 6)

1. **Pre-push snapshot.** Every apply captures a vendor-native backup and
   stores its path in the audit row. Backup failure aborts the run before any
   mutation.

2. **Atomic per-device apply.** Either the renderer commits a clean patch or
   it leaves the device untouched. The safety floor records the audit row on
   success AND on failure — a partial apply is impossible to lose.

3. **Per-run audit trail.** Every push writes one `PushRun` row to the
   in-memory `RunStore` (production wraps a CloverDB collection):
   `{device, started_at, finished_at, safety, renderer, backup_path,
   exit_status, error_string}`. Same audit pattern as `scan_runs`.

## Validation gate (Phase 7)

The intended config goes through [`internal/push/yang.go`](../internal/push/yang.go)
*before* the renderer opens an SSH session. A push that violates the model
invariants (a port carrying the same VLAN as both tagged and untagged — the
case `network-refinement`'s empirical suite flagged in the memoir) is
**rejected before any network I/O**.

## OpenWrt renderer (Phase 2)

The OpenWrt renderer is the v1 reference implementation. It speaks UCI:

```
DryRun     → no SSH traffic
Staged     → cat > /etc/nsl-push.tmp <<'EOF' ... EOF; uci show network | diff - -
Apply      → uci set … per change → uci changes (must be clean) → uci commit → reload_config
```

Renderer's `Diff` reports one `ConfigChange` per VLAN add/delete, with the
rendered `Patch` line attached (`uci add_list` or `uci del_list`). The
Diff engine (Phase 4) stitches those into a unified-diff preview.

Tests in
[`openwrt_renderer_test.go`](../internal/configparser/parsers/openwrt_renderer_test.go):

- `TestOpenWrtRenderer_Diff_*` — diff structure (add VLAN, remove VLAN, no-change).
- `TestOpenWrtRenderer_Render_DryRun_NoExecutions` — DryRun sends zero SSH commands.
- `TestOpenWrtRenderer_Render_StagedWritesPatchFile` — Staged embeds `uci set` in a heredoc, never reloads.
- `TestOpenWrtRenderer_Render_ApplyAddsReload` — Apply runs `reload_config` after `uci commit`.
- `TestOpenWrtRenderer_RenderOrdering` — set → commit → reload_config in that order.
- `TestOpenWrtRenderer_Integration` — runs against `NSL_LAB_HOST` if `NSL_LAB=1`.

## Deferred (v2)

- **OPNsense**: XML-patch complexity. Returns `ErrUnsupported{OS:"opnsense",
  Reason:"xml-patch complexity; see docs/push.md"}`.
- **VyOS**: operational-mode commit/`compare`/`discard` flow.
- **RouterOS**: `/import file=` semantics.
- **Infix**: `sysrepocfg -I FILE.json` patches.

## CLI (Phase 5)

```
nsl-graph push device    --device <id> [--apply]      # default --dry-run
nsl-graph push devices   --zone <id>  [--apply]      # default --dry-run
nsl-graph push preview   --device <id>                # read-only, no --apply
```

`push preview` does not expose `--apply` — preview IS dry-run.

## Endpoints (Phase 8)

```
POST /push/device   {device_id, apply}
POST /push/devices  {zone_id, apply}
```

The frontend (`frontend/php/push.php`) lists devices in a zone with a Push
button per row. OpenWrt devices get an enabled button; other vendors get a
disabled button with a tooltip pointing at this doc.

## Lab verification (Phase 9)

```
NSL_LAB=1 NSL_LAB_HOST=10.0.50.11 SSH_KEY=$HOME/.ssh/localinfra \
  go test ./cmd/push/ -run TestLabE2E -tags lab -v
```

Observed against the live lab on 2026-08-26:
- Fetched 6898 bytes from OpenWrt-1 (8 interfaces: br-lan, lan, wan, wan6, eth0, eth1, loopback).
- DryRun completed with zero SSH commands sent.
- Apply is gated behind `NSL_LAB_APPLY=1` (off by default; clones only).

## Why this is the read-and-write split, not a new app

The read side does three things:

1. SNMP/SSH scan (`internal/scanner`) → `SNMPDevice`.
2. Vendor parser (`internal/configparser/parsers`) → `ConfigData`.
3. YAML/YANG projection (`internal/yang/mapping`) → `canon.Root`.

The write side mirrors them in reverse, two new layers:

- **`ConfigRenderer`** (per vendor): `Diff(intended, observed) []ConfigChange`
  and `Render(intended, safety, session, creds) error`. Mirrors `ConfigParser`'s
  split between fetch-strategy (`Session`) and grammar (`ConfigParser`).
- **`Engine`** (`internal/push`): takes a device, calls the renderer's
  `Diff`, applies the YANG gate, the safety gate, and audits the run.

The transport seam (`internal/configparser/transport.go`) is reused as-is.
The `Session` interface, the `SSHTransport`, the credential plumbing — all
unchanged from the read side.

## See also

- `proposal-avances-rede.md` — the cross-cutting memoir proposal this work
  slots into.
- `internal/datastore/datastore.go` — NMDA-style candidate/validate/commit
  discipline that pushes the same intent into the DB layer; the push command
  reuses this on the device side.
- `internal/configparser/safety.go` — the gate.
- `internal/configparser/safety_test.go` — the test that pins the
  no-apply-by-default contract.
- `frontend/php/push.php` and `frontend/php/push.js` — the PHP panel.

## Build & verify

```bash
# Unit tests
go test ./internal/configparser/... ./internal/push/... ./cmd/push/... -count=1

# Live OpenWrt-1 (dry-run only)
NSL_LAB=1 NSL_LAB_HOST=10.0.50.11 SSH_KEY=$HOME/.ssh/localinfra \
  go test ./cmd/push/ -run TestLabE2E -tags lab -v

# CLI smoke test
go build -o /tmp/nsl-graph ./
/tmp/nsl-graph push --help
/tmp/nsl-graph push device --help
/tmp/nsl-graph push preview --help
```
