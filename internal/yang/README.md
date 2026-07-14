# `internal/yang` — the schema baseline

NSL-Graph's data model, expressed in **YANG 1.1** ([RFC 7950][rfc7950]).

This directory is self-contained: the standard modules we build on, the two modules we
author, and the fixtures that prove they work. Deleting it reverts the entire adoption.

Background and rationale: [`inv/yang.md`](../../inv/yang.md). On why this involves no
NETCONF: [`inv/netconf-yang.md`](../../inv/netconf-yang.md).

```
internal/yang/
  modules/     the .yang files — vendored standards + nsl-*
  canon/       the canonical model: Go structs mirroring the YANG tree, RFC 7951 tags
  mapping/     entities ⇄ canon — the only place the correspondence lives
  testdata/    instance fixtures, positive and negative
```

`entities` remains the domain model and remains what CloverDB persists. `canon` is a
*projection* of it onto the standard, built on demand. Keeping the two apart is what
confines the whole adoption to this directory — deleting `internal/yang/` reverts it.

## What this is for

Three things, in order of importance:

1. **A formally specified model.** The entity model stops being an assertion and becomes a
   citation. NSL-Graph's `Device` / `DevicePort` / `Connection` *is* RFC 8345's
   `node` / `termination-point` / `link` — and now that is checkable rather than claimed.
2. **A baseline for multivendor integration.** One canonical model that every vendor adapter
   reads into and renders out of.
3. **A standard export format.** RFC 7951 JSON that any YANG-aware tool can consume.

**It is not a protocol.** There is no NETCONF here and none is planned. YANG is the schema
(layer 4 of RFC 6241); the transports remain SSH and SNMP, exactly as before.

## The modules we author

### `nsl-topology.yang`

Augments RFC 8345 with the things the standards genuinely lack — which is to say, with
NSL-Graph's actual contribution:

- **`confidence`** (`confirmed` / `candidate` / `weak`) — the evidence ladder. RFC 8345 models
  a topology as *fact*; NMDA's `origin` distinguishes `learned` from `intended` but says
  nothing about *how strongly* something was learned. This does.
- **`discovered-via`** — provenance: which source observed the link, and from which endpoint.
- **`reviewed`** — whether an operator has accepted it into the specification.
- **`vlan-membership`** on a termination point — the topology-tree projection of the 802.1Q
  member set.

The load-bearing part is this `must`, which lifts NSL-Graph's import rule out of Go and into
the schema:

```yang
leaf confidence {
  type confidence;
  must ". != 'weak' or ../reviewed = 'true'" {
    error-message "A link whose only evidence is the forwarding database (weak) must be
                   reviewed by an operator before it can be committed: ...";
  }
}
```

`internal/topology` enforces exactly this in Go today. In the schema it is enforced once, for
the CLI, the HTTP API and the web frontend alike, and it is legible to someone who has never
read the Go.

### `nsl-inventory.yang`

The catalogues and the physical layer — brands, model types, OS types, owners, zone types,
**zones**, and device **models** with their **port templates** (including faceplate
`position-x` / `position-y`).

No IETF counterpart exists, and that is not an oversight: RFC 8345 models a topology a device
already knows about. It has no notion of a catalogue of models you might buy, or of where a
port sits on a faceplate — because the IETF is not drawing pictures of switches, and NSL-Graph
is.

It also carries the **model vs. device** distinction that the rest of the codebase rests on: a
*model* is a catalogue entry with a port template; a *device* (an RFC 8345 node) is an instance
of one, and its termination points are instances of the model's ports.

#### Why a zone is not an RFC 8345 `network`

The tempting mapping — one zone, one `network` — is wrong, and it is worth recording why,
because it looks right until you try to use it.

RFC 8345 requires both endpoints of a link to sit in the same network: `source-node` is a
leafref to `../../../nw:node/node-id`, and the RFC says so in words — *"Must be in the same
topology."* Zones-as-networks therefore makes every **inter-zone** link unrepresentable — and
the firewall between the DMZ and the LAN is precisely the link that matters most.

The mismatch is conceptual rather than incidental. An RFC 8345 `network` is a topology
**layer** — an L3 topology riding on an optical one, related by `supporting-network`. A zone
is an administrative grouping *within* one layer. They are different axes.

So: the L2 topology is a **single** `network` (`urn:nsl:net:specification`), and zone
membership is a leafref on the node. Zone nesting lives in the zone list's own `parent`
leafref, not in `supporting-network`.

## The standards we build on

Vendored, not fetched at build time, so the build is reproducible and offline.

| Module | Source | Gives us |
|---|---|---|
| `ietf-network`, `ietf-network-topology` | [RFC 8345][rfc8345] | `network` / `node` / `termination-point` / `link`; **leafrefs** between them |
| `ietf-l2-topology` | [RFC 8944][rfc8944] | `mac-address`, `management-address`, `interface-name` — typed, on the topology tree |
| `ietf-interfaces`, `ietf-ip` | [RFC 8343][rfc8343], [RFC 8344][rfc8344] | the interface / IP model |
| `ieee802-dot1q-bridge`, `ieee802-dot1q-types` | IEEE 802.1Q | `vlanid` (uint16, 1..4094); `egress-ports` / `untagged-ports` |
| `ieee802-types`, `iana-if-type`, `ietf-inet-types`, `ietf-yang-types` | — | base typedefs required by the above |

Provenance: all from [`github.com/YangModels/yang`](https://github.com/YangModels/yang)
(`standard/ietf/RFC`, `standard/ieee/published/802.1`, `standard/ieee/published/802`,
`standard/iana`), retrieved 2026-07-14. The `@<revision>` suffix has been stripped from the
filenames; the authoritative revision is the `revision` statement inside each file. **Do not
edit them** — they are inputs, not source.

## What the schema catches that the Go does not

Not hypothetical. Each of these is exercised by a fixture in `testdata/` and enforced by
`make yang-validate`:

| Today in Go | In the schema |
|---|---|
| `Vlan.VlanID` is a **`string`** — `"4999"` and `"eth0"` are storable | `dot1q-types:vlanid` — `uint16`, range `1..4094`. Rejected. |
| `Connection.FromDevice` is a device **name**, a bare string — a rename orphans it silently | a **`leafref`**. A dangling reference is a validation error, not a support ticket. |
| The weak-link import rule lives in one Go function | a **`must`**, enforced for every interface at once |
| `DevicePort.MacAddress`, `Device.Ips` are strings | `yang:mac-address`, `inet:ip-address` |

## Exporting

```bash
nsl-graph export yang -s test-dbs/real.db > topology.json
```

Emits the whole specification as RFC 7951 JSON. Devices are nodes, device ports are
termination points, and **each connection becomes two links** — RFC 8345 links are
unidirectional, and forgetting that is the easiest way to make the export silently
asymmetric.

Values the schema will not accept are **dropped and reported on stderr**, so the
emitted document always validates. Failing the whole export on the first bad value
would make the tool useless on exactly the data it exists to describe. `--strict`
inverts that and fails instead.

**The warnings are the interesting output.** They are a report of what the
hand-rolled schema has been quietly tolerating. Running it against `real.db` today
turns up nine of these:

```
warning: device "10.0.2.242" port "eth0": VLAN 6 is recorded as BOTH tagged and
untagged -- a port cannot do both. Keeping tagged=true (the first recorded); the
other is dropped. This is a data bug, most likely from merging the 802.1Q
egress-port and untagged-port sets on import.
```

A port either tags a VLAN's frames on egress or it does not. Recording both is not a
duplicate, it is a contradiction — and nothing in the current model can refuse it.

## Verifying

```bash
make yang-validate         # modules + fixtures
make yang-validate-export  # export a real DB and validate what actually comes out
```

The second is the claim that matters. Not *"the modules parse"* but *"what this tool
emits conforms to RFC 8345"* — which, as of now, it does.

Needs `yanglint` (Debian/Ubuntu: `sudo apt install libyang3-tools`). It does two things:

1. **Schema check** — every module parses and every augment resolves.
2. **Instance check** — `testdata/sample-topology.json` must **pass**, and
   `testdata/invalid-weak-unreviewed.json` must **fail**. The negative fixture is the one that
   matters: if it ever starts passing, the `must` has stopped doing its job and the guarantee
   is gone.

`pyang` (`pipx install pyang`) is a useful second opinion on the modules, but it does not
validate instance data — for that, `yanglint` is required.

[rfc7950]: https://www.rfc-editor.org/rfc/rfc7950
[rfc8343]: https://www.rfc-editor.org/rfc/rfc8343
[rfc8344]: https://www.rfc-editor.org/rfc/rfc8344
[rfc8345]: https://www.rfc-editor.org/rfc/rfc8345
[rfc8944]: https://www.rfc-editor.org/rfc/rfc8944
