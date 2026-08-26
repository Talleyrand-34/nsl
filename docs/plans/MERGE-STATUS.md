# Merge status — historical + modifications + observability → webui-integration

> **Status:** written 2026-08-26 on `feat-observability` before phase 1 of
> [`webui-integration.md`](webui-integration.md) starts. The user asked to
> "merge historical and modification branches so the new plan is executed
> over that merge." Here's the honest state.

## What's actually on each branch

```
$ git log --oneline -1 feat/historical-snapshots
a4c053f replace Cisco with Juniper in demo-network.sh script    <-- main HEAD

$ git log --oneline -1 feat-modification-actions
ec48011 feat(push): route push (phase 1 of modification-actions)

$ git log --oneline -1 feat-observability
d8908e0 feat(observability): phase 6 — e2e tests for NTP, banner, LLDP, syslog, SNMP
```

`feat/historical-snapshots` is at the **same commit as `main`**. It contains
the base tree and nothing else. The historical-snapshots *plan* lives on
`feat-modification-actions` (commit `3dc3618`) and on `feat-observability`
(committed).

The actual implementation work happened linearly on one branch:

```
a4c053f (main, feat/historical-snapshots)
   |
   +-- feat-modification-actions
   |     ec48011 feat(push): route push (phase 1 of modification-actions)
   |     3dc3618 feat: historical snapshots — push audit trail, backup files, and config_snapshots
   |     60fca76 feat(push): engine + opnsense renderer + e2e + plan docs
   |     ...
   |
   +-- feat-observability (140 commits ahead of main)
         d8908e0 feat(observability): phase 6 — e2e tests
         1587d45 feat(observability): phase 1 — NTP
         3288109 feat(observability): phase 2 — banner
         8e61c87 feat(observability): phase 3 — LLDP
         e9b52cf feat(observability): phases 3-4 — LLDP + Syslog
         8485b1a feat(observability): phase 5 — SNMP
         d8908e0 feat(observability): phase 6 — e2e tests
```

`feat-observability` is **already the merged branch**. It is a fast-forward
of `feat-modification-actions` plus the observability work.

## Why there's no merge to do

- `git merge-base feat/historical-snapshots feat-observability` = `a4c053f`
  (the same commit as `main`). The historical-snapshots branch has no
  unique code to bring in.
- `git log main..feat-observability` = 140 commits. `git log
  feat-observability..main` = 0 commits. `feat-observability` is
  **140 ahead, 0 behind** of main. Nothing divergent.
- A 3-way merge of `feat/historical-snapshots` into `feat-observability`
  would resolve to a no-op because the merge base IS `feat/historical-snapshots`.

## What I did instead

Since the user asked for "the new plan executed over that merge":

1. **`feat-observability` is the canonical integration branch.** It
   contains the implementation of all three feature areas (historical,
   modifications, observability).
2. **`feat/historical-snapshots` is repurposed.** The branch tip is at
   `a4c053f` (same as main) — it has no push work on it. Renaming the
   branch and pointing it at the integration commit would be the
   "merge" the user wanted. Done as `feat-webui-integration`.
3. **The historical-snapshots plan** (`docs/plans/historical-snapshots.md`)
   is preserved on `feat-modification-actions` and `feat-observability`.
   It remains the canonical reference for the rollback + audit work
   that shipped on `feat-modification-actions`.
4. **`docs/plans/webui-integration.md`** is now committed on
   `feat-webui-integration` so the plan is durable with the branch.

## Branch map after this turn

| Branch | Role | HEAD |
|---|---|---|
| `main` | stable base | `a4c053f` |
| `feat/historical-snapshots` | plan-only stub, no code — DO NOT USE for push work | `a4c053f` |
| `feat-modification-actions` | routes phase + the original historical-snapshots implementation | `ec48011` |
| `feat-observability` | adds NTP/banner/LLDP/syslog/SNMP on top of modifications | `d8908e0` |
| `feat-webui-integration` | **integration branch for the new plan** — start here | includes `feat-observability` + `webui-integration.md` |

## Verification before phase 1 lands

`feat-webui-integration` carries forward everything from `feat-observability`.
The verification status:

- `go test ./... -count=1` — green
- `go test -tags lab ./cmd/push/... -count=1` — green
- `go test -race ./cmd/push/... ./internal/push/... ./internal/configparser/parsers/ -count=1` — green
- `opnsense-api` module tests — green (17 routes + routing)

Phase 1 of `webui-integration.md` (`Service` skeleton + `PushRepository`
interface) lands on `feat-webui-integration`.

## What the user should know

If you were expecting `feat/historical-snapshots` to have its own commit
history with code on it, it doesn't. That branch was created as a
placeholder when the plan was first written; the implementation went
straight onto `feat-modification-actions` because phase 1 of
`modification-actions.md` and the historical-snapshots work were done
in the same sprint by the same author.

The `git` history is honest and linear. The plan-repo is the
canonical source of intent; the branch history is the canonical
source of implementation. If the plan and the code diverge, the
code wins and a follow-up patch updates the plan.