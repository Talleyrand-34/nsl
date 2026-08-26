# Observability & services — NTP, SNMP, syslog, banner

> **Status:** plan, not started. Companion to
> [`historical-snapshots.md`](historical-snapshots.md) and
> [`modification-actions.md`](modification-actions.md).
> The third plan covers the small per-device services that an operator sets
> once and forgets, but which break audits and observability when wrong:
> NTP, SNMP, syslog forwarding, and the login banner. LLDP-transmit was
> considered and dropped — it is a self-description knob (the box decides
> what to advertise about itself), not a per-device state an operator
> pushes to control behaviour.

## Why

Read-side parsers never populated `NTP`, `SNMP`, `Syslog`, or `Banner` for
most vendors — and the push surface has no diff/render for any of them.
Operators today set these by hand on every box. NSL-Graph discovers a
device but doesn't know what time it thinks it is, whether it sends syslog
anywhere, what its SNMP community is, or what banner greets a login
attempt. The graph is rich on topology and short on operational metadata.

This plan is deliberately smaller than
[`modification-actions.md`](modification-actions.md). Every action here is a
**set/get** pair, not the `add_item`/`del_item`/`reconfigure` multi-stage
shape that routes and firewall rules require. No UUID resolution, no
transactional apply, no rule-graph dependency to reason about.

## Outcome

After this plan lands:

- Every push carries the device's **NTP servers** as part of `*ConfigData`.
- Every push carries the device's **SNMP configuration** (community, traps, listen ACL).
- Every push carries the device's **syslog forwarding targets**.
- Every push carries the device's **login banner**.
- All four are pushable on OpenWrt (SSH/UCI) and OPNsense (REST API). Other
  vendors stay read-only for now.
- Rollback through the historical-snapshots plan restores all four along
  with the rest of the config.


```
                                ConfigData (extended)
                                       |
                                       v
            +---------------------------------------------+
            | diffNTP / diffSNMP / diffSyslog /          |
            | diffBanner                                   |
                                  |
                                  v  []ConfigChange{Kind:"ntp-set", Patch:...}
            +---------------------------------------------+
            | applyChange dispatches Kind to:             |
            |   ntp-set       -> UCI or REST per vendor  |
            |   snmp-community-set ...                     |
            |   syslog-target-add ...                      |
            |   banner-set ...                             |
            +---------------------------------------------+
```

The action surface reuses the same `ConfigChange{Kind, Path, Patch}` shape as
the VLAN renderer. The diff happens against an observed snapshot stored in
the historical DB (see
[`historical-snapshots.md`](historical-snapshots.md)), so a rollback to a
prior snapshot re-applies these too.

## Model extensions

Add to `internal/configparser/interfaces.go` (one file each, to keep the
existing `interfaces.go` from sprawling further):

```go
// ConfigNTPServer is one upstream NTP server.
type ConfigNTPServer struct {
    Address     string `json:"address"`            // hostname or IP
    Prefer      bool   `json:"prefer,omitempty"`   // OpenWrt: marks the preferred source
    IBurst      bool   `json:"iburst,omitempty"`
    Enabled     bool   `json:"enabled"`
}

type ConfigNTPConfig struct {
    Enabled    bool              `json:"enabled"`
    Servers    []ConfigNTPServer `json:"servers"`
    LocalClock bool              `json:"local_clock,omitempty"` // OpenWrt: use local clock as fallback
    Timezone   string            `json:"timezone,omitempty"`    // IANA tz, e.g. "Europe/Madrid"
}

// ConfigSNMPCommunity is one SNMP community + access + trap target.
type ConfigSNMPCommunity struct {
    Name    string   `json:"name"`
    Access  string   `json:"access"`           // "ro" | "rw"
    Network []string `json:"network"`          // allowed source networks (CIDR)
    TrapTarget string `json:"trap_target,omitempty"` // hostname/IP for traps
}

type ConfigSNMPConfig struct {
    Enabled     bool                  `json:"enabled"`
    Location    string                `json:"location,omitempty"`   // "Building A, Rack 4"
    Contact     string                `json:"contact,omitempty"`    // "noc@example.com"
    Communities []ConfigSNMPCommunity `json:"communities"`
    ListenInterface string            `json:"listen_interface,omitempty"` // OPNsense binds to a single interface
}

// ConfigSyslogTarget is one remote syslog endpoint.
type ConfigSyslogTarget struct {
    Address     string `json:"address"`             // hostname or IP
    Port        int    `json:"port"`                // default 514
    Protocol    string `json:"protocol,omitempty"`  // "udp" | "tcp" | "tls"
    Facility    string `json:"facility,omitempty"`  // OPNsense: "local0".."local7"
    Program     string `json:"program,omitempty"`   // OpenWrt: filter by program
    LogOnly     string `json:"log_only,omitempty"`  // OPNsense: rfc5424 levels, e.g. "info"
}

type ConfigSyslogConfig struct {
    Enabled bool                `json:"enabled"`
    Targets []ConfigSyslogTarget `json:"targets"`
    PreserveFQDN bool           `json:"preserve_fqdn,omitempty"` // OPNsense
}

// ConfigBanner is the pre-login MOTD / system banner.
type ConfigBanner struct {
    LoginBanner string `json:"login_banner,omitempty"`   // shown before auth
    PostLogin   string `json:"post_login,omitempty"`    // shown after auth (issue / motd)
}

// ConfigLLDPSettings controls what this box advertises over LLDP.
type ConfigLLDPSettings struct {
    Enabled       bool   `json:"enabled"`                  // transmit on/off
    ChassisID     string `json:"chassis_id,omitempty"`     // override the default MAC-derived ID
    SystemName    string `json:"system_name,omitempty"`    // overrides LLDP System-Name TLV
    SystemDesc    string `json:"system_desc,omitempty"`    // overrides LLDP System-Description TLV
    MgmtAddress   string `json:"mgmt_address,omitempty"`   // LLDP Management-Address TLV
    InterfacePattern string `json:"interface_pattern,omitempty"` // OpenWrt: UCI pattern for which ifaces to transmit on
}
```

`ConfigData` gets new fields (all `omitempty` for snapshot back-compat):

```go
type ConfigData struct {
    // ... existing ...
    NTP      *ConfigNTPConfig      `json:"ntp,omitempty"`
    SNMP     *ConfigSNMPConfig     `json:"snmp,omitempty"`
    Syslog   *ConfigSyslogConfig   `json:"syslog,omitempty"`
    Banner   *ConfigBanner         `json:"banner,omitempty"`
    LLDP     *ConfigLLDPSettings   `json:"lldp,omitempty"`
}
```

Ponytail: pointers, not zero-value structs, so the absence of a section is
distinguishable from a configured-but-empty one. Cheap, idiomatic, and the
JSON tag hides the pointer in stored snapshots.

## Vendor paths

| Action | OpenWrt (UCI) | OPNsense (REST) |
|---|---|---|
| NTP | `uci get system.@timeserver[0].server` per server; `system.@system[0].timezone` | `GET /api/ntp/settings/get` → `timeservers`, `enable`, `timezone` |
| SNMP | `uci get snmpd.@config[0]` + `snmpd.@community[0]` per community; `snmpd.@trap[0]` per trap target | `GET /api/snmp/general/get` (community, location, contact); community-only read via `/api/snmp/general/get` |
| Syslog | `uci show system.@log_remote[0]` per target; `system.@log_remote[0].enabled` | `GET /api/syslog/destinations/search_item` + `GET /api/syslog/setting/get` |
| Banner | `uci get system.@system[0].banner_file` + the file contents from `/etc/banner` | `GET /api/system/settings/get` → `banner` |
| LLDP tx | `uci get lldpd.@config[0].tx_interval` + `lldpd.@interface[0].enabled` per interface | `GET /api/lldp/service/get` (enable/disable); `GET /api/lldp/interface/search_item` (per-iface pattern) |

VyOS / RouterOS / Infix / Fortinet: **read-only for now**. Their `Fetch`
already pulls UCI/config.boot/exports; adding the parse rules is a
follow-up pass after the OpenWrt + OPNsense paths land.

## Wrappers to add in `opnsense-api/modules/`

| Module | Path | Endpoints |
|---|---|---|
| `ntp` | `/api/ntp/settings/*` | `get`, `set` |
| `snmp/general` | `/api/snmp/general/*` | `get`, `set` |
| `syslog/destinations` | `/api/syslog/destinations/*` | `search_item`, `get`, `add_item`, `del_item`, `set`, `reconfigure` |
| `syslog/setting` | `/api/syslog/setting/*` | `get`, `set` |
| `system/settings` | `/api/system/settings/*` | `get`, `set` |
| `lldp/service` | `/api/lldp/service/*` | `get`, `set`, `reconfigure` |
| `lldp/interface` | `/api/lldp/interface/*` | `search_item`, `get`, `set`, `reconfigure` |

Each wrapper follows the existing `interfaces.Module` shape. None require
`reconfigure`-style two-stage apply for v1 — every endpoint above either
self-applies on `set` or doesn't have a separate apply step.

## Render side

Five new `applyChange` branches keyed by `Kind`:

| Kind | OpenWrt | OPNsense |
|---|---|---|
| `ntp-set` | `uci set system.@timeserver[0].server='<host>'; uci commit system; /etc/init.d/sysntpd reload` | `POST /api/ntp/settings/set` |
| `snmp-community-set` | `uci set snmpd.@community[0].rocommunity='<name>'; uci commit snmpd; /etc/init.d/snmpd reload` | `POST /api/snmp/general/set` |
| `syslog-target-add` | `uci add system log_remote; uci set ...; uci commit system; /etc/init.d/log restart` | `POST /api/syslog/destinations/add_item` then `POST /syslog/setting/reconfigure` |
| `banner-set` | `cat > /etc/banner <<EOF\n<text>\nEOF` (Staged) ; commit reload (Apply) | `POST /api/system/settings/set` with `banner` field |
| `lldp-transmit-set` | `uci set lldpd.@config[0].enabled='1'; uci commit lldpd; /etc/init.d/lldpd reload` | `POST /api/lldp/service/set` then `reconfigure` |

The OpenWrt path is **read via SSH** (`Fetch`) and **written via UCI** over
SSH (`Apply`). The OPNsense path is **REST throughout**.

## Files

### New

| Path | Purpose |
|---|---|
| `internal/configparser/types_ntp.go` | `ConfigNTPServer`, `ConfigNTPConfig` |
| `internal/configparser/types_snmp.go` | `ConfigSNMPCommunity`, `ConfigSNMPConfig` |
| `internal/configparser/types_syslog.go` | `ConfigSyslogTarget`, `ConfigSyslogConfig` |
| `internal/configparser/types_banner.go` | `ConfigBanner` |
| `internal/configparser/types_lldp.go` | `ConfigLLDPSettings` |
| `internal/configparser/parsers/diff_ntp.go` | shared `diffNTP(intended, observed)` |
| `internal/configparser/parsers/diff_snmp.go` | shared `diffSNMP(...)` |
| `internal/configparser/parsers/diff_syslog.go` | shared `diffSyslog(...)` |
| `internal/configparser/parsers/diff_banner.go` | shared `diffBanner(...)` |
| `internal/configparser/parsers/diff_lldp.go` | shared `diffLLDP(...)` |
| `opnsense-api/modules/ntp/ntp.go` + tests | NTP wrapper |
| `opnsense-api/modules/snmp/general.go` + tests | SNMP wrapper |
| `opnsense-api/modules/syslog/destinations.go` + tests | destinations wrapper |
| `opnsense-api/modules/syslog/setting.go` + tests | global setting wrapper |
| `opnsense-api/modules/system/settings.go` + tests | banner wrapper |
| `opnsense-api/modules/lldp/service.go` + tests | lldp service wrapper |
| `opnsense-api/modules/lldp/interface.go` + tests | lldp per-interface wrapper |
| `cmd/push/observability_e2e_test.go` | end-to-end httptest test for the OPNsense path |

### Modified

| Path | Change |
|---|---|
| `internal/configparser/interfaces.go` | `ConfigData` adds the 5 new `*Config` pointers |
| `internal/configparser/parsers/openwrt.go` | new `parseNTPConfig`, `parseSNMPConfig`, `parseSyslogConfig`, `parseBanner`, `parseLLDPConfig` shims — invoked from the existing `ParseConfig` |
| `internal/configparser/parsers/openwrt_renderer.go` | new `applyChange` branches; `renderUCICommands` adds per-area entries |
| `internal/configparser/parsers/openwrt_renderer_test.go` | tests for each branch |
| `internal/configparser/parsers/opnsense_renderer.go` | new `applyChange` branches + dispatch in the existing `applyChange` |
| `internal/configparser/parsers/opnsense_renderer_test.go` | tests for each branch |
| `internal/repository/infra/cloverdb/base/config_snapshot.go` | no model change required (pointers serialise cleanly); `ParserVersion` bumped to 2 |
| `docs/push.md` | new section listing the action surface including these |

### Out of scope (separate files, future plan)

- **`opnsense-api/modules/radius/`** — auth sources. Sensitive (lockout
  risk); not in this plan.
- **`opnsense-api/modules/ipsec/`** — IPsec tunnels. Separate plan.
- **`opnsense-api/modules/wireguard/`** — WireGuard. Separate plan.
- **LLDP neighbour table read** — that's a scan-side feature, not a push
  concern. Already partially handled by `internal/scanner`.
- **Read-side parsing of NTP / SNMP / syslog / banner / LLDP for
  vyos / routeros / infix / fortinet** — read-only parity is a follow-up
  pass. Only OpenWrt + OPNsense are pushable in this plan.

## Phasing (TDD, smallest diff each)

Each phase is one logical action: model + parser shim + wrapper + renderer
branch + tests. Order by complexity, lowest first:

1. **NTP** — simplest shape (single config block, no list-of-lists). ~200 LOC.
2. **Banner** — single string field per vendor. ~150 LOC.
3. **LLDP** — small but has per-interface pattern on OpenWrt. ~250 LOC.
4. **Syslog** — list-of-targets + global setting. ~300 LOC.
5. **SNMP** — community list + trap target + global contact/location. ~350 LOC.
6. **End-to-end test** — extend `cmd/push/opnsense_e2e_test.go` (or new
   file) to push NTP + banner + SNMP in one go. ~100 LOC.
7. **Rollback integration** — verify a snapshot captured before these
   existed restores cleanly when the device has them configured. Tests
   require the historical-snapshots plan to land first.

Each phase ends with `go test ./opnsense-api/... ./internal/... ./cmd/push/... -race -count=1` green.

## Risks

1. **SNMP community exposure.** The model carries community strings in
   cleartext. Mitigation: vault-encrypt in transit; do not log; redact from
   audit rows. Defer to a follow-up if the vault API isn't ready.
2. **Banner content size.** A multi-KB MOTD is a normal operator thing. Keep
   the model field unconstrained, but cap at 4 KB to bound snapshot size.
3. **Syslog restart.** `/etc/init.d/log restart` may briefly drop local
   log messages on OpenWrt. Acceptable for a push tool; document.
4. **SNMP apply semantics.** OPNsense's `/api/snmp/general/set` activates
   the daemon on set; there's no separate apply. The test must observe that
   the daemon port is listening, not just that the API returned 200.
5. **LLDP transmit-cycle restart.** Briefly disrupts neighbour discovery.
   Acceptable.
6. **Timezone.** OpenWrt stores the timezone as a single string; OPNsense
   splits between GUI locale and CLI POSIX TZ. Diff is by POSIX TZ string.
   Document this in the test that pins the round-trip.

## Verification

- Per phase: `go test ./opnsense-api/... ./internal/... ./cmd/push/... -race -count=1` green.
- After phase 5: a complete OPNsense push that configures all five areas in
  one Render call, with the lab handler asserting each `POST` lands on the
  right endpoint and the audit row records `success`.
- After phase 7: rollback test that uses a snapshot from before NTP/SNMP
  existed, applies it, and verifies the device ends up with the older
  config (only achievable if historical-snapshots is already merged).

## Open question

**SNMP v3.** This plan models SNMP v2c only (community strings). v3 (auth +
priv) is a much bigger surface and a different keying model. Should v3
support land in this plan or as a follow-up? Most small/medium sites are
still on v2c; v3 is "if you have to ask, you already know you need it."

**Tell me "v2c only for v1" or "include v3"** — and I'll start phase 1
(NTP) once the historical-snapshots and modification-actions plans are
underway.