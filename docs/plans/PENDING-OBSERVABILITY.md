# Pending issues — observability plan, before implementation

> **Status:** written 2026-08-26 on `feat-observability` (just branched from
> `feat-modification-actions`). Captures every question, conflict, or
> unverified assumption in `docs/plans/observability-and-services.md` that
> needs your decision before code lands.
>
> This file is **the contract for implementation**. Resolving these is a
> prerequisite for the first phase.

## Findings

### Critical — wrong endpoints (must fix before any wrapper lands)

These were inherited from the plan's stale API knowledge. Phase 1 of
modification-actions hit the same kind of error (the plan said
`/api/routes/routes/add_item`, the real one is `/api/routes/routes/addroute`).
We're not catching the same trap twice.

**1. OPNsense SNMP module is `netsnmp`, not `snmp`.**
- Plan: `GET /api/snmp/general/get`, `POST /api/snmp/general/set`.
- Real (verified at <https://docs.opnsense.org/development/api/plugins/netsnmp.html>):
  - `GET /api/netsnmp/general/get`
  - `POST /api/netsnmp/general/set`
  - `POST /api/netsnmp/service/reconfigure` (snmpd needs reconfigure after set)
  - `POST /api/netsnmp/service/restart`
  - The `community` field is a single string under `general`, not the rich
    per-community list the model assumes (`ConfigSNMPCommunity`). v3 users
    live under `netsnmp/user/*`.
- **Implication**: `ConfigSNMPCommunity` collapses to a single
  `Community string` field for v2c. v3 users get their own table.
  The plan's model needs rework.

**2. OPNsense Syslog endpoints use plural verbs, not `*_item`.**
- Plan: `GET /api/syslog/destinations/search_item`, `POST /api/syslog/destinations/add_item`, `POST /api/syslog/destinations/del_item/$uuid`, `POST /api/syslog/setting/reconfigure`.
- Real (verified at <https://docs.opnsense.org/development/api/core/syslog.html>):
  - `GET,POST /api/syslog/settings/search_destinations`
  - `POST /api/syslog/settings/add_destination`
  - `POST /api/syslog/settings/del_destination/$uuid`
  - `GET /api/syslog/settings/get` (singular `settings`)
  - `POST /api/syslog/settings/set`
  - `POST /api/syslog/service/reconfigure`
- **Implication**: wrapper path is `syslog/settings`, not `syslog/destinations`. Verbs are `add_destination` / `del_destination` / `search_destinations`.

**3. OPNsense NTP settings endpoint is unverified.**
- Plan: `GET /api/ntp/settings/get`, `POST /api/ntp/settings/set`.
- Status: official docs only list `/api/ntpd/service/{status,meta,gps}`. The
  settings endpoint is plausibly `/api/ntp/settings/get` (it's a plugin module
  pattern in OPNsense) but **not confirmed**. The NTP plugin does expose a
  `settings` controller (per the plan), but the URL is uncertain.
- **Need**: verify against a live OPNsense box, or fall back to direct
  config.xml manipulation via the `core/backup` + manual parsing approach.
  Without verification the wrapper can't be written safely.

### Critical — open design questions (the plan's "Open question" was vague)

**4. SNMP v3 in v1 or v3-only later?**
- The plan said: "default recommendation: include v3 alongside v2c, both
  gated by `ConfigSNMPCommunity.Version`". I now want to confirm before
  modeling it.
- **Question**: ship v2c-only first (faster), or model v3 now (one
  consistent pass)?

**5. Banner: `/etc/banner` only, or also PostLogin message-of-the-day?**
- The model has `LoginBanner` + `PostLogin`. OpenWrt has `/etc/banner`
  (login) but the post-login MOTD lives in `/etc/motd` and is independent.
  OPNsense `system/settings` has a single `banner` field used for pre-login.
- **Question**: ship LoginBanner only first; PostLogin later if needed?

**6. Syslog protocol support — `tls` is in the model, real?**
- Model: `"udp" | "tcp" | "tls"`.
- Real OPNsense destinations carry `transport` (`udp`, `tcp`); TLS syslog is
  supported via additional `certificate` field. The plan's `tls` enum
  value would need a certificate reference — not modelled.
- **Question**: drop `tls` from v1 model; add later with certificate handling?

### Important — read-side gaps the plan didn't address

**7. OpenWrt `OpenWrtParser.Fetch` doesn't pull `system`, `snmpd`, or `system.@log_remote[]`.**
- Phase 1 verified that the OPNsense side reads via REST (works). The
  OpenWrt side reads via `OpenWrtParser.Fetch`, which currently runs
  `uci show network` and `uci show firewall` but NOT `system` or `snmpd`.
- **Implication**: an OpenWrt push of NTP/SNMP/syslog will read *nothing*
  from the device, so the diff will be empty against observed. Every
  intended change will register as an add — but the diff won't know
  what's already there.
- **Need**: extend `OpenWrtParser.Fetch` to run `uci show system` and
  `uci show snmpd`. ~40 LOC. Must land before any OpenWrt observability
  phase starts.

**8. OpenWrt SSH path for syslog + banner can't go through UCI alone.**
- `uci set system.@log_remote[0].server=...` works for syslog on OpenWrt,
  but a config reload needs `/etc/init.d/log restart` (or
  `/etc/init.d/syslog-ng restart` depending on the image). The renderer
  needs to call the right init script per device.
- `/etc/banner` is a plain file, not UCI. The render must `cat > /etc/banner`
  via SSH, not through UCI. That's a different code path from the
  rest of the OpenWrt renderer.
- **Need**: confirm whether the lab's OpenWrt uses `logd` or `syslog-ng`;
  add `banner-write` SSH helper; choose consistent per-vendor init-script
  mapping.

### Important — interactions with phase 1's commit-per-area

**9. NTP and SNMP self-apply on `set`; syslog doesn't.**
- Phase 1's `touchedIface` / `touchedRoutes` model generalises to a
  per-area commit map. For observability:
  - `ntp-set` → no extra commit (the endpoint self-applies)
  - `snmp-community-set` → `netsnmp/service/reconfigure` after set
  - `syslog-target-add` → `syslog/settings/set` then
    `syslog/service/reconfigure`
  - `banner-set` → no extra commit (system/settings/set self-applies)
- The renderer's touched-area map needs an extra entry per new area.
  Ponytail-clean: add a `touchedServices` bool, do per-area apply if
  `touchedServices && syslog` (since only syslog needs an explicit
  reconfigure).

**10. Per-area commit ordering vs audit row.**
- Phase 1 already calls `routes.RouteApply` and `ifaces.OverviewCommit`
  independently. The SafetyFloor audit row records one row per
  Apply, not per area. So a push that touches routes+syslog+NTP gets
  ONE audit row, even though three reconfigure endpoints were hit.
- **Question**: does the operator want per-area audit rows? Or is
  one row per push (current contract) sufficient? The plan doesn't say.

### Nice-to-have — modelling polish

**11. `ConfigNTPServer.Prefer` and `ConfigNTPServer.IBurst`.**
- Both are OpenWrt-only flags. OPNsense has `prefer` and `noselect`
  flags. The model is generic but the mapping needs to be explicit.
- **Question**: one model with both flags, or vendor-specific sub-models?

**12. `ConfigSyslogTarget.Facility` (OpenWrt) vs OPNsense's `facility` field.**
- OpenWrt uses `program=` for filtering by program (in
  `system.@log_remote[0].program`). OPNsense uses `program` for the
  application name too. Same field, same semantics. **Fine to merge.**

**13. `ConfigBanner.PostLogin` and OpenWrt `/etc/motd`.**
- See #5. PostLogin is OpenWrt-specific. OPNsense has nothing analogous.
- **Question**: model it? Or drop for v1?

### Documentation gaps

**14. The plan doesn't say how `fetchObserved` learns the device's NTP/SNMP/syslog config for the OPNsense renderer.**
- Phase 1's renderer added `r.routes.SearchRoute(ctx)` to `fetchObserved`.
  Observability needs analogous `r.ntp.GetSettings()` / `r.snmp.Get()`
  / `r.syslog.SearchDestinations()`. The renderer's `fetchObserved`
  will fan out to 5+ endpoints per push. That's still cheap (all GETs,
  <50ms each on a real box), but it's worth measuring.
- **Implication**: consider adding a `per-area observed fetcher`
  interface so the renderer only fetches what the diff actually needs.
  Ponytail: defer until we see a real perf issue.

**15. The plan's "End-to-end httptest test" file path is `cmd/push/observability_e2e_test.go`.**
- The cmd/push package is the place. The test should mirror `opnsense_e2e_test.go`'s
  shape (lab handler + intent + Render + SafetyFloor). Fine; just calling
  out so we don't reinvent.

## What I need from you before phase 1 (NTP) lands

Three decisions, in priority order:

1. **SNMP module name** — confirm the wrapper path is
   `/api/netsnmp/general/*`, not `/api/snmp/general/*`. (Source:
   <https://docs.opnsense.org/development/api/plugins/netsnmp.html>.)
2. **Syslog endpoint paths** — confirm `/api/syslog/settings/*` with
   `add_destination`/`del_destination`/`search_destinations` verbs.
3. **NTP endpoint** — either confirm `/api/ntp/settings/get`+`set` exists
   on a live box, or pick the fallback approach (parse config.xml).
4. **SNMP v2c-only first** vs **SNMP v3+ v2c in v1**?

Once those four are settled I'll rebase the plan, write the model + the
two OpenWrt-side read helpers (Fetch extension for `system` + `snmpd`),
and start phase 1.