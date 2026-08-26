# Recheck — observability plan after phase 1 of modification-actions

> **Status:** written 2026-08-26 after `ec48011` (phase 1 routes) landed on
> `feat-modification-actions`. Re-evaluates `docs/plans/observability-and-services.md`
> against what's now real.

## What phase 1 changed for observability

**Nothing direct.** Phase 1 was routes only — new `opnsense-api/modules/{routes,routing}`, `diff_routes.go`, OPNsense `fetchObserved` extended to pull routes, OpenWrt `routeAddCommands`/`routeDelCommands` wired through `renderChange`. None of this touches NTP/SNMP/syslog/banner.

What it **did** change downstream:

- The OPNsense renderer's `fetchObserved` now hits **two** endpoints per push (interfaces + routes) instead of one. Adding NTP/SNMP/syslog/banner in a future phase will add more endpoints; the per-area commit logic introduced here (`touchedIface`, `touchedRoutes`) generalises cleanly to a third area.
- The diff vocabulary was normalised: `diffVLANs` now emits `vlan-add` instead of `add`. NTP/SNMP/syslog/banner diffs should follow the same convention (`ntp-set`, `snmp-community-set`, etc.) to keep the dispatch table in `applyChange` clean.
- The route diff's per-field Patch format (`network=X`, `gateway=Y`) is the pattern NTP/SNMP/banner diffs should mirror — not one space-separated string.
- `OverviewCommit` now surfaces non-2xx. Same change is needed for the NTP/SNMP/syslog wrappers when those land; their existing `routeSuccess`-style body check won't be enough because those endpoints use different success strings.

## Observability plan items — still open, status of each

| Item | Plan section | Status after phase 1 |
|---|---|---|
| `ConfigNTPConfig` model + parser + wrapper + renderer branch | "Model extensions" | unchanged — not started |
| `ConfigSNMPConfig` model + parser + wrapper + renderer branch | "Model extensions" | unchanged — not started |
| `ConfigSyslogConfig` model + parser + wrapper + renderer branch | "Model extensions" | unchanged — not started |
| `ConfigBanner` model + parser + wrapper + renderer branch | "Model extensions" | unchanged — not started |
| OpenWrt UCI parsers for `system.@timeserver[]`, `snmpd.@config[0]`, `system.@log_remote[]`, `/etc/banner` | "Read-side parsers" | unchanged — not started |
| OPNsense wrappers (`ntp`, `snmp/general`, `syslog/destinations`, `system/settings`) | "Wrappers to add in `opnsense-api/modules/`" | unchanged — not started |
| `diffNTP`/`diffSNMP`/`diffSyslog`/`diffBanner` helpers | "Render side" | unchanged — not started |
| Renderer branches for `ntp-set`, `snmp-community-set`, `syslog-target-add`, `banner-set` | "Render side" | unchanged — not started |
| End-to-end httptest test for the OPNsense path | "Phasing" phase 6 | unchanged — not started |
| Rollback integration test | "Phasing" phase 7 | unchanged — depends on historical-snapshots |

## Read-side backfill items from the original analysis — still open

The previous turn's recommendation: "backfill the read-side gaps as part of phase 1 of modification-actions.md". Phase 1 didn't do that (it was routes-only, and the renderer-side read path for OPNsense routes is now wired via `decodeRoutesInto`). The remaining read-side gaps are unchanged:

| Gap | Effort | Should land where |
|---|---|---|
| Populate `ConfigData.Domain` for openwrt/opnsense/routeros/infix/fortinet | ~80 LOC + tests | observability plan phase 1 (NTP, while we're already touching system-config parsers) |
| Populate `ConfigInterface.{Description,MACAddress,MTU}` for OPNsense | ~80 LOC + tests | modification-actions phase 2 (firewall — same fetch path as aliases) |
| Populate `ConfigData.ConfigVersion` for the missing 4 vendors | ~60 LOC + tests | observability plan phase 1 (alongside Domain) |
| Populate `ConfigData.FirewallRules` for OPNsense (parse `firewall/filter` search response) | ~150 LOC + tests | modification-actions phase 3 (firewall) |

## What changed in the observability plan that should be reflected

**Substantive plan updates** the phase 1 work made evident:

1. **Per-area commit is now the renderer pattern.** The observability plan's "render side" table lists each action with its own UCI/API command, but says nothing about commit. It needs to acknowledge that `ntp-set` triggers `POST /api/ntp/settings/set` (which self-commits), `snmp-community-set` triggers `POST /api/snmp/general/set` (which self-applies), while syslog needs `destinations/add_item` + `setting/reconfigure` — and the renderer must track per-area state the same way it now tracks `touchedIface`/`touchedRoutes`.

2. **The diff Patch format.** The plan shows patch lines like `"uci set dhcp.@dnsmasq[0].server=..."` — i.e. one rendered line per change. The route diff (after phase 1) emits per-field strings. Both styles are valid; the renderer dispatches on `Kind` and reconstructs the payload from `change.New`/`change.Old`/`change.Patch`. The plan should note that **the wrapper layer receives `Kind` + a small payload, not a free-form Patch array**, for type safety.

3. **OpenWrt SSH-side discovery.** The plan assumes OpenWrt `Fetch` already pulls the relevant UCI sections (`system.@timeserver[]`, etc.). It does not — `OpenWrtParser.Fetch` runs `uci show dhcp` and `uci show network` and a few others, but not `uci show system` or `uci show snmpd`. The phase 1 implementation for routes confirms that OPNsense pulls observed via REST, not from the parsed config — so the OpenWrt-side read-back isn't strictly required for the OPNsense e2e, but it is required for OpenWrt parity. **Add a sub-task: extend `OpenWrtParser.Fetch` to include `system`, `snmpd`, and `system` banner file** — likely 80 LOC.

4. **Open question on SNMP v3** from the original plan is **still open**. Phase 1 made it more pressing: the wrapper layer is now small and consistent, so adding v3 alongside v2c is cheap; deferring means a second pass later. **Default recommendation: include v3 (auth: SHA, priv: AES) in the model alongside v2c, both gated by `ConfigSNMPCommunity.Version`.**

## What's left, in priority order

1. **historical-snapshots plan** — the foundation. Without it, audit rows are in-memory only, rollback is impossible. Still not started.
2. **modification-actions phase 2 (firewall aliases)** — the most-requested next action after routes. Aliases are a prerequisite for rules.
3. **modification-actions phase 3 (firewall rules + NAT)** — biggest user-facing win.
4. **modification-actions phase 4 (DHCP)** — ~300 LOC.
5. **modification-actions phase 5 (DNS)** — ~250 LOC.
6. **observability plan phase 1 (NTP + read-side Domain/ConfigVersion backfill)** — smallest scope, gives us a complete NTP push end-to-end on OpenWrt + OPNsense.
7. **observability plan phase 2 (banner)** — single string field; trivial.
8. **observability plan phase 3 (LLDP)** — was dropped from the plan; no work to do.
9. **observability plan phase 4 (syslog)** — ~300 LOC; depends on banner + NTP for sequencing.
10. **observability plan phase 5 (SNMP)** — depends on syslog for sequencing; largest single phase.
11. **End-to-end rollback** — requires historical-snapshots first.

## Recommendation

Continue with **modification-actions phase 2 (firewall aliases)** next. The pattern from phase 1 maps cleanly: typed wrapper, fetch observed, per-area commit, TDD red-green. Aliases unlock the firewall-rules phase that follows. The observability plan items can wait until aliases + rules are shipped.

Tell me which phase to land next; the order above is my recommendation but the call is yours.