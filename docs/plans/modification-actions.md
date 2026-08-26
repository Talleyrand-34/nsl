# Modification actions — what the push pipeline can change today, and what to add

> **Status:** plan, not started. Companion to [`historical-snapshots.md`](historical-snapshots.md).
> Today the push surface is **VLANs only**. This plan widens it to the actions
> that an operator actually needs to push: routes, firewall aliases + rules,
> NAT, DNS forwarders / resolver settings, DHCP scopes, and CARP / HA-IPs.

## What's available right now

A `*ConfigData` carries six populated fields, and the renderers touch only one:

| Model field | Parser support | Renderer support |
|---|---|---|
| `Interfaces[]` | openwrt, vyos, routeros, infix, fortinet, freebsd | openwrt + opnsense (VLAN add/del only) |
| `VLANs[]` | openwrt, freebsd, fortinet | openwrt + opnsense (via `Interfaces[].VLANs[]`) |
| `Routes[]` | openwrt, vyos, routeros, infix, fortinet | **none** |
| `RoutingProtocols[]` | openwrt, vyos, routeros, infix, fortinet | **none** (YANG gate only validates) |
| `FirewallRules[]` | openwrt, fortinet | **none** |
| `SwitchPorts[]` | openwrt | **none** |

Not in the model at all: `DHCP`, `DNS`, `NAT`, `Aliases`, `CARP / VIP`.

Not implemented on any vendor: routes, firewall rules, NAT, DHCP, DNS, HA-IPs.
The push pipeline only mutates VLAN membership.

## Why

VLAN membership is the minimum useful push. Everything else that an operator
wants to roll forward (route a new prefix, allow a new subnet, hand out a new
DHCP range, fail over a VIP to a peer) has to be done by hand-editing the
device or via a separate tool. That's the gap.

## Outcome

After this plan lands:

- **Routes** — static routes added, removed, or updated on OPNsense and OpenWrt.
- **Firewall aliases** — OPNsense `firewall/alias` CRUD wrapped; OpenWrt UCI alias support.
- **Firewall rules** — OPNsense `firewall/filter` and `firewall/{source,dest,nat,one_to_one,npt}` CRUD wrapped; OpenWrt UCI rule CRUD.
- **NAT** — handled as a subset of firewall rules (one_to_one, source NAT, destination NAT, NPTv6).
- **DNS** — OPNsense `dns/dnsmasq` (forwarder) and `dns/unbound` (resolver) settings; OpenWrt UCI `dhcp.@dnsmasq[]`.
- **DHCP** — OPNsense `dhcpv4/{subnet,reservation}` + `dhcpv6`; OpenWrt UCI `dhcp.@dhcp[]` per-interface scopes.
- **HA-IP / CARP** — OPNsense `interfaces/vip_settings` (the "Virtual IPs" page); VyOS `high-availability virtual-address`.

Each of these is one `*Module` wrapper per vendor in `opnsense-api/modules/<area>/<area>.go`, plus a Diff→Render action in the corresponding renderer.

## Architecture

```
                                ConfigData (extended)
                                       |
                                       v
                       +-------------------------------+
                       |  diffRoutes, diffFirewall,    |
                       |  diffNAT, diffDNS, diffDHCP,  |
                       |  diffCARP                     |
                       +---------------+---------------+
                                       |
                                       v  []ConfigChange{Kind, Path, Patch}
                       +-------------------------------+
                       | applyChange dispatches by    |
                       | Kind to the right wrapper:   |
                       |   routes -> RouteModule      |
                       |   alias  -> AliasModule      |
                       |   rule   -> RuleModule       |
                       |   ...                         |
                       +---------------+---------------+
                                       |
                                       v
                   Vendor-specific commit / reload
                   OPNsense: POST /<area>/reconfigure
                   OpenWrt:  uci commit <area>
```

The action surface is built per-action, not per-vendor: a `RouteModule` in
`opnsense-api/modules/routes/routes.go` works the same way as `interfaces.Module`
already does. OpenWrt's SSH-side equivalent is a thin shim that translates
`ConfigChange{Kind: "route-add", Path: "192.168.5.0/24 via 10.0.0.1"}` into
`uci set network.route_foo=...`.

## Model extensions

Add to `internal/configparser/interfaces.go` (or split into per-area files
once the file gets long):

```go
// ConfigRoute is already there. Routes differ from RoutingProtocols: static
// routes are flat (this type); dynamic routing is in RoutingProtocols[].

type ConfigAlias struct {
    Name    string   `json:"name"`     // OPNsense alias name, OpenWrt UCI section name
    Type    string   `json:"type"`     // "host", "network", "port", "url", "urltable", ...
    Content []string `json:"content"`  // OPNsense: alias members; OpenWrt: UCI list
    Enabled bool     `json:"enabled"`
    Description string `json:"description,omitempty"`
}

type ConfigFirewallRule struct {
    // already exists. Extend with:
    Action       string `json:"action,omitempty"`        // "pass" | "block" | "reject"
    Quick        bool   `json:"quick,omitempty"`          // OpenWrt
    Interface    string `json:"interface,omitempty"`      // "lan", "wan", ...
    Direction    string `json:"direction,omitempty"`      // "in", "out"
    Source       string `json:"source,omitempty"`         // alias name or CIDR
    Destination  string `json:"destination,omitempty"`
    Protocol     string `json:"protocol,omitempty"`       // "tcp", "udp", "icmp", "any"
    SourcePort   string `json:"source_port,omitempty"`
    DestinationPort string `json:"destination_port,omitempty"`
    Description  string `json:"description,omitempty"`
    Enabled      bool   `json:"enabled"`
    Nat          *ConfigNAT `json:"nat,omitempty"`        // when this rule is also a NAT rule
}

type ConfigNAT struct {
    Type        string `json:"type"`             // "1:1", "source", "dest", "nptv6"
    PublicIP    string `json:"public_ip,omitempty"`
    PrivateIP   string `json:"private_ip,omitempty"`
    PublicPort  string `json:"public_port,omitempty"`
    PrivatePort string `json:"private_port,omitempty"`
}

type ConfigDNSForwarder struct {
    Enabled  bool     `json:"enabled"`
    Listen   []string `json:"listen"`     // interfaces to listen on
    Upstream []string `json:"upstream"`   // upstream resolvers
    Domains  []string `json:"domains,omitempty"` // forward specific domains
}

type ConfigDNSResolver struct { // unbound settings
    Enabled  bool     `json:"enabled"`
    DNSSEC   bool     `json:"dnssec"`
    Listen   []string `json:"listen"`
    Upstream []string `json:"upstream,omitempty"`
}

type ConfigDHCPScope struct {
    Interface  string   `json:"interface"`           // "lan", "opt1", ...
    RangeStart string   `json:"range_start"`
    RangeEnd   string   `json:"range_end"`
    Gateway    string   `json:"gateway,omitempty"`
    DNS        []string `json:"dns,omitempty"`
    Domain     string   `json:"domain,omitempty"`
    LeaseTime  string   `json:"lease_time,omitempty"`
    Enabled    bool     `json:"enabled"`
}

type ConfigDHCPREServation struct {
    MACAddress string `json:"mac"`
    IPAddress  string `json:"ip"`
    Hostname   string `json:"hostname,omitempty"`
    Description string `json:"description,omitempty"`
}

type ConfigVIP struct {            // OPNsense "Virtual IP" / VyOS virtual-address / CARP
    Address    string `json:"address"`        // "10.0.0.10/24"
    Mode       string `json:"mode"`           // "carp", "ipalias", "proxyarp"
    Interface  string `json:"interface"`
    VHID       uint8  `json:"vhid,omitempty"`  // CARP only
    Password   string `json:"password,omitempty"` // CARP only (skipped in push)
    AdvBase    int    `json:"adv_base,omitempty"`
    AdvSkew    int    `json:"adv_skew,omitempty"`
    Description string `json:"description,omitempty"`
}
```

`ConfigData` gets new slices for each, all behind `omitempty` for back-compat
with existing snapshots (see [`historical-snapshots.md`](historical-snapshots.md)
re. `ParserVersion`).

## Wrappers to add in `opnsense-api/modules/`

| Module | Path | Endpoints |
|---|---|---|
| `routes` | `/api/routes/routes/*` | `add_item`, `del_item`, `set`, `get`, `search_item`, `reconfigure` |
| `firewall/alias` | `/api/firewall/alias/*` | `add_item`, `del_item`, `set`, `get`, `search_item`, `reconfigure` |
| `firewall/filter` | `/api/firewall/filter/*` | `add_rule`, `del_rule`, `set_rule`, `get_rule`, `search_rule`, `apply` |
| `firewall/source_nat` | `/api/firewall/source_nat/*` | same shape as filter |
| `firewall/d_nat` | `/api/firewall/d_nat/*` | same shape |
| `firewall/one_to_one` | `/api/firewall/one_to_one/*` | same shape |
| `firewall/npt` | `/api/firewall/npt/*` | same shape |
| `dns/unbound` | `/api/unbound/settings/*` | `get`, `set`, `reconfigure` (note: unbound lives under `unbound`, not `dns`) |
| `dns/dnsmasq` | `/api/dnsmasq/settings/*` | `get`, `set`, `service restart` |
| `dhcpv4/subnet` | `/api/dhcpv4/subnet/*` | `add_item`, `del_item`, `set`, `get`, `search_item`, `reconfigure` |
| `dhcpv4/reservation` | `/api/dhcpv4/reservation/*` | same shape |
| `dhcpv6` | `/api/dhcpv6/*` | same shape |
| `interfaces/vip` | `/api/interfaces/vip_settings/*` | `add_item`, `del_item`, `set`, `get`, `search_item`, `reconfigure` |

Every wrapper follows the existing `interfaces.Module` shape: `func New(c *opnsense.Client) *Module`, typed payloads, returns UUID on add.

## Render side

Each renderer grows new `diff*` and `applyChange` branches keyed by `Kind`:

| Kind | OpenWrt | OPNsense |
|---|---|---|
| `route-add` | `uci set network.<id>=route; uci set network.<id>.target=...; uci set network.<id>.gateway=...; uci commit network` | `POST /api/routes/routes/add_item` then `reconfigure` |
| `route-del` | `uci delete network.<id>; uci commit network` | find UUID by `search_item` then `del_item` |
| `alias-add` | `uci add firewall alias; uci set ...; uci commit firewall` | `POST /api/firewall/alias/add_item` |
| `rule-add` | UCI rule blocks per interface | `POST /api/firewall/filter/add_rule` then `apply` |
| `nat-add` | one of `firewall.redirect` / `firewall.snat` / `firewall.dnat` | `source_nat` / `d_nat` / `one_to_one` |
| `dns-upstream-set` | `uci set dhcp.@dnsmasq[0].server=...; /etc/init.d/dnsmasq restart` | `POST /api/unbound/settings/set` + `reconfigure` |
| `dhcp-scope-add` | UCI `dhcp.@dhcp[<i>]` | `POST /api/dhcpv4/subnet/add_item` + `reconfigure` |
| `dhcp-reservation-add` | UCI `dhcp.@host[<i>]` | `POST /api/dhcpv4/reservation/add_item` |
| `vip-add` | n/a (VyOS: `set high-availability virtual-address <ip> ...` via SSH) | `POST /api/interfaces/vip_settings/add_item` |

## Files

### New

| Path | Purpose |
|---|---|
| `opnsense-api/modules/routes/routes.go` + tests | routes wrapper |
| `opnsense-api/modules/firewall/alias.go` + tests | aliases wrapper |
| `opnsense-api/modules/firewall/filter.go` + tests | rules wrapper |
| `opnsense-api/modules/firewall/{source_nat,d_nat,one_to_one,npt}.go` + tests | NAT wrappers |
| `opnsense-api/modules/unbound/unbound.go` + tests | unbound settings |
| `opnsense-api/modules/dnsmasq/dnsmasq.go` + tests | dnsmasq settings |
| `opnsense-api/modules/dhcpv4/{subnet,reservation}.go` + tests | DHCPv4 wrappers |
| `opnsense-api/modules/dhcpv6/dhcpv6.go` + tests | DHCPv6 wrapper |
| `opnsense-api/modules/interfaces/vip.go` + tests | CARP / VIP wrapper |
| `internal/configparser/types_alias.go` | `ConfigAlias`, `ConfigNAT` |
| `internal/configparser/types_dns.go` | `ConfigDNSForwarder`, `ConfigDNSResolver` |
| `internal/configparser/types_dhcp.go` | `ConfigDHCPScope`, `ConfigDHCPREServation` |
| `internal/configparser/types_vip.go` | `ConfigVIP` |
| `internal/configparser/parsers/diff_routes.go` | shared `diffRoutes(intended, observed)` |
| `internal/configparser/parsers/diff_firewall.go` | shared `diffFirewall(...)` |
| `internal/configparser/parsers/diff_dhcp.go` | shared `diffDHCP(...)` |
| `internal/configparser/parsers/diff_dns.go` | shared `diffDNS(...)` |
| `internal/configparser/parsers/diff_vip.go` | shared `diffVIP(...)` |

### Modified

| Path | Change |
|---|---|
| `internal/configparser/interfaces.go` | `ConfigData` adds the new slices (omitempty); `ConfigFirewallRule` extended |
| `internal/configparser/parsers/openwrt_renderer.go` | new `applyChange` branches; new `parse*` shims for the OpenWrt fetch side |
| `internal/configparser/parsers/openwrt_renderer_test.go` | tests for each new branch |
| `internal/configparser/parsers/opnsense_renderer.go` | new `applyChange` branches; dispatcher table keyed by Kind |
| `internal/configparser/parsers/opnsense_renderer_test.go` | tests for each new branch |
| `internal/configparser/parsers/openwrt.go` | new `parseFirewallAliases`, `parseDHCPScopes`, `parseDNSForwarder`, `parseStaticRoutes` (latter already exists) |
| `internal/configparser/parsers/opnsense_parser.go` | new file: parses OPNsense `config.xml` (or `interfaces_info` + `dhcpv4/*` + `firewall/filter/*`) into the model |
| `docs/push.md` | action surface table |

Ponytail: **don't** try to parse `config.xml` to recover the full OPNsense
config in one pass. Phase 1 reads per-area endpoints (`firewall/filter/*`,
`dhcpv4/*`, `routes/*`, `unbound/settings`) the same way the wrapper layer
does. Each area is small, well-documented, and one API call away. The XML
parser stays out of scope.

## Phasing (TDD, smallest diff each)

Each phase is one logical action: model + parser + wrapper + renderer branch
+ tests. Order by user value:

1. **Routes** — most-requested action. ~250 LOC across model + openwrt +
   opnsense + wrapper + tests.
2. **Firewall aliases** — required for almost every other firewall change.
   ~200 LOC.
3. **Firewall rules + NAT** — biggest win; biggest surface. ~400 LOC.
4. **DHCP scopes + reservations** — ~300 LOC.
5. **DNS forwarder + resolver** — ~250 LOC.
6. **CARP / VIP** — small surface; ~150 LOC. VyOS `high-availability virtual-address` SSH equivalent.
7. **Snapshot-aware rollback** — historical-snapshots + these new actions:
   rollback can render any action, not just VLANs. (`applyChange` already
     dispatches by Kind; the historical part is just plumbing.)

Each phase ends with `go test ./... -race -count=1` green.

## Out of scope (explicit)

- **VRRP / keepalived** — different daemon, different model. Add when needed.
- **Policy-based routing (PBR)** — VyOS `policy route`, OPNsense firewall rule
  with gateway. The rule + route primitives cover most cases; PBR is a refinement.
- **BGP / OSPF / RIP config** — dynamic routing is read-only today (YANG gate
  only validates); pushing protocols needs a transactional protocol daemon
  restart, which is a much bigger commitment. Defer.
- **QoS / traffic shaping / queues** — big modelling problem; separate plan.
- **WireGuard / IPsec** — keyed tunnels; defer.
- **High-availability state synchronisation** (XMLRPC sync on OPNsense) — out
  of scope for a single-node push.
- **Privileged action gating** — every wrapper here requires the same level
  of API privilege. The vault already stores credentials; ACLs on the OPNsense
  side are an operational concern, not a code one.

## Risks

1. **Schema drift between OPNsense versions.** The wrapper layer is per-version
   fragile. Mitigation: wrapper tests pin the JSON shape against a known
   version; `ParserVersion` on snapshots lets us reject older ones.
2. **Transactional semantics.** OPNsense staging needs `apply` or `reconfigure`
   per area; multiple adds followed by a single apply is the common shape. The
   renderer must remember to call `apply` once per area, not per add. Unit-test
   the dispatcher to catch this.
3. **Identifier resolution.** OpenWrt UCI uses string section names; OPNsense
   uses UUIDs. The renderer must look up UUIDs by name before deleting. Pin a
   test that asserts delete-by-path triggers a search first.
4. **CARP passwords.** VIPs with CARP mode carry a cleartext password in the
   model. Mitigation: vault-encrypt before serialise; render decrypts at the
   moment of use. Defer the password field for v1 if not needed.
5. **DNS resolver restart.** Unbound restarts break in-flight queries. Acceptable
   for a push tool; document the side-effect.
6. **DHCP scope overlaps.** Two scopes covering the same range is invalid on
   every NOS. The YANG gate should reject overlapping scopes before render;
   add a test in `internal/yang/`.

## Verification

- Per phase: `go test ./opnsense-api/... ./internal/... -race -count=1` green.
- After phase 3 (rules + NAT): a fresh end-to-end httptest-driven OPNsense test
  in `cmd/push/opnsense_e2e_test.go` that adds an alias, then a rule that
  references it, asserts the call sequence, and confirms the audit row.
- After phase 7: a rollback round-trip via [`historical-snapshots.md`](historical-snapshots.md) where the snapshot was captured before a route was added,
  and the rollback removes it cleanly.

## Open question

**OpenWrt fetch.** Parsing UCI into `*ConfigData` is already mostly done for
interfaces, routes, firewall. But OpenWrt's `dhcp.@dnsmasq[0]` block is fetched
via `uci show dhcp` — do we already pull that? (If not, the plan above grows
by ~80 LOC of UCI parsers.) Want me to scope the OpenWrt fetch deltas before
the first phase starts?