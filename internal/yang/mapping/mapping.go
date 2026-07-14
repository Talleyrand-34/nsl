// Package mapping projects the NSL-Graph domain model (internal/repository/entities)
// onto the canonical YANG tree (internal/yang/canon).
//
// This is the only place the correspondence between the two lives. entities stays
// the domain and stays what CloverDB persists; canon is derived on demand. Keeping
// the mapping in one package is what lets the whole YANG adoption be reverted by
// deleting internal/yang.
//
// # Why it emits warnings instead of failing
//
// The domain model is weakly typed in exactly the places the schema is strict: VLAN
// identifiers are strings, and cross-entity references are names rather than keys.
// A real database therefore contains values the schema rejects -- a VLAN of "4999",
// an owner nobody ever created.
//
// Failing the whole export on the first of those would make the tool useless on the
// data it exists to describe. Instead the mapper drops the offending value, records
// a warning, and emits a document that still validates. The warnings are the point:
// they are a report of everything the hand-rolled schema has been quietly tolerating.
package mapping

/*
  Copyright © 2026 Talleyrand-34 (t34@t34.dev)

  This program is free software: you can redistribute it and/or modify
  it under the terms of the GNU Affero General Public License as published
  by the Free Software Foundation, either version 3 of the License, or
  (at your option) any later version.

  This program is distributed in the hope that it will be useful,
  but WITHOUT ANY WARRANTY; without even the implied warranty of
  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
  GNU Affero General Public License for more details.

  You should have received a copy of the GNU Affero General Public License
  along with this program. If not, see <https://www.gnu.org/licenses/>.
*/

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"nsl-graph/internal/yang/canon"

	e "nsl-graph/internal/repository/entities"
)

// Source is everything the mapper needs, as plain slices. Taking the data rather
// than the service keeps the mapper trivially testable and free of any dependency
// on the repository layer.
type Source struct {
	Brands      []e.Brand
	ModelTypes  []e.ModelType
	OsTypes     []e.OsType
	Owners      []e.Owner
	ZoneTypes   []e.ZoneType
	Zones       []e.Zone
	Models      []e.ModelDevice
	ModelPorts  []e.ModelPort
	Devices     []e.Device
	DevicePorts []e.DevicePort
	Connections []e.Connection
}

// URI prefixes. RFC 8345 keys are inet:uri, so every NSL-Graph identifier is minted
// into a URN rather than emitted as a bare UUID.
const (
	deviceURIPrefix = "urn:nsl:device:"
	tpURIPrefix     = "urn:nsl:tp:"
	linkURIPrefix   = "urn:nsl:link:"
)

// discoverySources are the enum values nsl-topology:discovery-source permits. A
// provenance string naming anything else is dropped with a warning rather than
// emitted, since the schema would reject it.
var discoverySources = map[string]bool{
	"snmp-lldp":  true,
	"snmp-cdp":   true,
	"snmp-fdb":   true,
	"ssh-lldp":   true,
	"ssh-fdb":    true,
	"local-lldp": true,
}

// FromEntities projects a specification onto the canonical YANG tree.
//
// It never fails: values the schema would reject are dropped and reported in the
// returned warnings, so the emitted document always validates.
func FromEntities(src Source) (*canon.Root, []string) {
	m := &mapper{src: src}
	m.index()

	root := &canon.Root{
		Inventory: m.inventory(),
		Networks: &canon.Networks{
			Network: []canon.Network{{
				NetworkID: canon.NetworkID,
				NetworkTypes: canon.NetworkTypes{
					L2Topology:  &struct{}{},
					NSLTopology: &struct{}{},
				},
				Nodes: m.nodes(),
				Links: m.links(),
			}},
		},
	}
	return root, m.warnings
}

type mapper struct {
	src      Source
	warnings []string

	// Catalogue membership. A leafref may only point at an entry that exists, so
	// every name is checked against these before it is emitted.
	brands     map[string]bool
	modelTypes map[string]bool
	osTypes    map[string]bool
	owners     map[string]bool
	zoneTypes  map[string]bool
	zoneIDs    map[string]bool

	// modelID resolves a (model name, brand name) pair to a model ID. Both
	// e.ModelDevice and e.ModelPort identify their model by name plus brand rather
	// than by key, so this pair is the only thing that joins them.
	modelID map[modelKey]string

	// portID resolves a (device label, port name) pair to a DevicePort ID.
	// e.Connection names its endpoints that way -- FromDevice is a device *label*
	// and FromModelPort a port *name* -- which is precisely the missing foreign key
	// the schema replaces with a leafref.
	portID map[portKey]string

	portsByDevice map[string][]e.DevicePort
}

type modelKey struct{ model, brand string }
type portKey struct{ device, port string }

func (m *mapper) warnf(format string, args ...any) {
	m.warnings = append(m.warnings, fmt.Sprintf(format, args...))
}

func (m *mapper) index() {
	m.brands = namesOf(m.src.Brands, func(b e.Brand) string { return b.Name })
	m.modelTypes = namesOf(m.src.ModelTypes, func(t e.ModelType) string { return t.Name })
	m.osTypes = namesOf(m.src.OsTypes, func(t e.OsType) string { return t.Name })
	m.owners = namesOf(m.src.Owners, func(o e.Owner) string { return o.Name })
	m.zoneTypes = namesOf(m.src.ZoneTypes, func(t e.ZoneType) string { return t.Name })
	m.zoneIDs = namesOf(m.src.Zones, func(z e.Zone) string { return z.ID })

	m.modelID = make(map[modelKey]string, len(m.src.Models))
	for _, md := range m.src.Models {
		m.modelID[modelKey{md.Model, md.Brand}] = md.ID
	}

	m.portID = make(map[portKey]string, len(m.src.DevicePorts))
	m.portsByDevice = make(map[string][]e.DevicePort)
	for _, dp := range m.src.DevicePorts {
		m.portID[portKey{dp.DevLabel, dp.PortName}] = dp.ID
		m.portsByDevice[dp.DeviceID] = append(m.portsByDevice[dp.DeviceID], dp)
	}
}

func namesOf[T any](xs []T, key func(T) string) map[string]bool {
	out := make(map[string]bool, len(xs))
	for _, x := range xs {
		if k := key(x); k != "" {
			out[k] = true
		}
	}
	return out
}

// ref returns name if it exists in the catalogue, and otherwise "" plus a warning.
// Emitting an unknown name would produce a document the schema rejects: a leafref
// must resolve. Dropping it keeps the export valid and surfaces the dangling
// reference, which the string-typed domain model cannot detect at all.
func (m *mapper) ref(catalogue map[string]bool, name, kind, owner string) string {
	if name == "" {
		return ""
	}
	if !catalogue[name] {
		m.warnf("%s: %s %q does not exist in the catalogue; reference dropped", owner, kind, name)
		return ""
	}
	return name
}

// -----------------------------------------------------------------------------
// Inventory
// -----------------------------------------------------------------------------

func (m *mapper) inventory() *canon.Inventory {
	inv := &canon.Inventory{}

	for _, b := range m.src.Brands {
		inv.Brands = append(inv.Brands, canon.Named{Name: b.Name})
	}
	for _, t := range m.src.ModelTypes {
		inv.ModelTypes = append(inv.ModelTypes, canon.Named{Name: t.Name})
	}
	for _, t := range m.src.OsTypes {
		inv.OsTypes = append(inv.OsTypes, canon.Named{Name: t.Name})
	}
	for _, o := range m.src.Owners {
		inv.Owners = append(inv.Owners, canon.Named{Name: o.Name})
	}
	for _, t := range m.src.ZoneTypes {
		inv.ZoneTypes = append(inv.ZoneTypes, canon.Named{Name: t.Name})
	}

	for _, z := range m.src.Zones {
		where := fmt.Sprintf("zone %q", z.Name)
		cz := canon.Zone{
			ID:       z.ID,
			Name:     z.Name,
			ZoneType: m.ref(m.zoneTypes, z.LocationType, "zone-type", where),
			Owner:    m.ref(m.owners, z.Owner, "owner", where),
		}
		if z.FatherID != "" {
			if !m.zoneIDs[z.FatherID] {
				m.warnf("%s: parent zone %q does not exist; reference dropped", where, z.FatherID)
			} else {
				cz.Parent = z.FatherID
			}
		}
		inv.Zones = append(inv.Zones, cz)
	}

	// ModelPorts are a flat list identifying their model by (name, brand), so they
	// are grouped back under their model here.
	portsByModel := make(map[string][]canon.ModelPort)
	for _, mp := range m.src.ModelPorts {
		id, ok := m.modelID[modelKey{mp.Model, mp.Brand}]
		if !ok {
			m.warnf("model-port %q: no model %q/%q exists; port dropped", mp.Name, mp.Brand, mp.Model)
			continue
		}
		portsByModel[id] = append(portsByModel[id], canon.ModelPort{
			ID:                       mp.ID,
			Name:                     mp.Name,
			PositionX:                mp.Positionx,
			PositionY:                mp.Positiony,
			PortType:                 normalisePortType(mp.PortType),
			Band:                     bandFor(mp.PortType, mp.Band),
			AllowMultipleConnections: mp.AllowMultipleConnections,
		})
	}

	for _, md := range m.src.Models {
		where := fmt.Sprintf("model %q", md.Model)
		inv.Models = append(inv.Models, canon.Model{
			ID:         md.ID,
			Name:       md.Model,
			Brand:      m.ref(m.brands, md.Brand, "brand", where),
			ModelType:  m.ref(m.modelTypes, md.ModelType, "model-type", where),
			OsType:     m.ref(m.osTypes, md.OsType, "os-type", where),
			ModelPorts: portsByModel[md.ID],
		})
	}

	return inv
}

// normalisePortType maps the domain's free-form port type onto the schema's
// enumeration. The domain uses "" to mean wired; the schema names it.
func normalisePortType(t string) string {
	switch t {
	case "", "ethernet":
		return "ethernet"
	case "wifi":
		return "wifi"
	default:
		return "ethernet"
	}
}

// bandFor returns the radio band only for WiFi ports. The schema gates `band` on
// `when "../port-type = 'wifi'"`, so a band on a wired port is a validation error
// rather than a field everyone agrees to ignore.
func bandFor(portType, band string) string {
	if normalisePortType(portType) != "wifi" {
		return ""
	}
	return band
}

// -----------------------------------------------------------------------------
// Nodes and termination points
// -----------------------------------------------------------------------------

func (m *mapper) nodes() []canon.Node {
	nodes := make([]canon.Node, 0, len(m.src.Devices))

	for _, d := range m.src.Devices {
		where := fmt.Sprintf("device %q", d.Label)

		n := canon.Node{
			NodeID:      deviceURIPrefix + d.ID,
			Owner:       m.ref(m.owners, d.Owner, "owner", where),
			Unmanaged:   d.IsUnmanaged,
			Invisible:   d.IsInvisible,
			ScanProfile: d.Profile,
		}

		if d.Label != "" || len(d.Ips) > 0 {
			n.L2 = &canon.L2NodeAttributes{
				Name:              d.Label,
				ManagementAddress: d.Ips,
			}
		}

		if d.ZoneID != "" {
			if !m.zoneIDs[d.ZoneID] {
				m.warnf("%s: zone %q does not exist; reference dropped", where, d.ZoneID)
			} else {
				n.Zone = d.ZoneID
			}
		}

		if id, ok := m.modelID[modelKey{d.Model, d.Brand}]; ok {
			n.Model = id
		} else if d.Model != "" {
			m.warnf("%s: no model %q/%q exists; reference dropped", where, d.Brand, d.Model)
		}

		n.TerminationPoints = m.terminationPoints(d)
		nodes = append(nodes, n)
	}

	return nodes
}

func (m *mapper) terminationPoints(d e.Device) []canon.TerminationPoint {
	ports := m.portsByDevice[d.ID]
	if len(ports) == 0 {
		return nil
	}

	tps := make([]canon.TerminationPoint, 0, len(ports))
	for _, dp := range ports {
		tp := canon.TerminationPoint{
			TpID: tpURIPrefix + dp.ID,
			// DevicePort.ModelID holds the MODEL PORT id, despite the name.
			ModelPort:      dp.ModelID,
			VlanMembership: m.vlans(dp, d.Label),
		}
		if dp.PortName != "" || dp.MacAddress != "" {
			tp.L2 = &canon.L2TPAttributes{
				InterfaceName: dp.PortName,
				MacAddress:    strings.ToLower(dp.MacAddress),
			}
		}
		tps = append(tps, tp)
	}
	return tps
}

// vlans converts a port's VLAN configs, discarding any the schema would reject.
//
// This is where the string-typed VLAN identifier meets dot1q-types:vlanid (uint16,
// 1..4094). Anything that is not a decimal number in range never existed as a VLAN;
// it is a typo that the domain model had no way to refuse.
func (m *mapper) vlans(dp e.DevicePort, devLabel string) []canon.VlanMembership {
	if len(dp.VlanConfigs) == 0 {
		return nil
	}

	// vlan-id is the list key, so at most one entry per VLAN may be emitted.
	seen := make(map[uint16]bool, len(dp.VlanConfigs))
	out := make([]canon.VlanMembership, 0, len(dp.VlanConfigs))

	for _, vc := range dp.VlanConfigs {
		where := fmt.Sprintf("device %q port %q", devLabel, dp.PortName)

		n, err := strconv.ParseUint(strings.TrimSpace(vc.VlanNumber), 10, 16)
		if err != nil {
			m.warnf("%s: VLAN %q is not a number; dropped (802.1Q VLAN IDs are 1..4094)", where, vc.VlanNumber)
			continue
		}
		if n < 1 || n > 4094 {
			m.warnf("%s: VLAN %d is out of range; dropped (802.1Q VLAN IDs are 1..4094)", where, n)
			continue
		}
		vid := uint16(n)

		if tagged, dup := seen[vid]; dup {
			if tagged != vc.Tagged {
				// A port either tags a VLAN's frames on egress or it does not.
				// Recording both is not a duplicate; it is a contradiction, and it
				// means one of the two is wrong. Keep the first deterministically
				// and say so loudly -- the alternative is to pick silently, which
				// is what the string-typed domain model does today.
				m.warnf("%s: VLAN %d is recorded as BOTH tagged and untagged -- a port cannot do both. "+
					"Keeping tagged=%t (the first recorded); the other is dropped. This is a data bug, "+
					"most likely from merging the 802.1Q egress-port and untagged-port sets on import.",
					where, vid, tagged)
			} else {
				m.warnf("%s: VLAN %d appears twice with the same tagging; duplicate dropped", where, vid)
			}
			continue
		}
		seen[vid] = vc.Tagged

		out = append(out, canon.VlanMembership{VlanID: vid, Tagged: vc.Tagged})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].VlanID < out[j].VlanID })
	return out
}

// -----------------------------------------------------------------------------
// Links
// -----------------------------------------------------------------------------

// links converts each connection into TWO unidirectional RFC 8345 links.
//
// RFC 8345 links are directed; an NSL-Graph connection is not. Emitting only one
// direction would make the exported topology asymmetric and silently unusable to
// any consumer that walks it from the far end.
func (m *mapper) links() []canon.Link {
	links := make([]canon.Link, 0, len(m.src.Connections)*2)

	for _, c := range m.src.Connections {
		where := fmt.Sprintf("connection %s", c.ID)

		fromTP, okFrom := m.portID[portKey{c.FromDevice, c.FromModelPort}]
		if !okFrom {
			m.warnf("%s: no port %q on device %q; connection dropped", where, c.FromModelPort, c.FromDevice)
			continue
		}
		toTP, okTo := m.portID[portKey{c.ToDevice, c.ToModelPort}]
		if !okTo {
			m.warnf("%s: no port %q on device %q; connection dropped", where, c.ToModelPort, c.ToDevice)
			continue
		}

		fromDev, okFD := m.deviceIDByLabel(c.FromDevice)
		toDev, okTD := m.deviceIDByLabel(c.ToDevice)
		if !okFD || !okTD {
			m.warnf("%s: endpoint device not found; connection dropped", where)
			continue
		}

		via := m.provenance(c.DiscoveredVia, where)
		confidence := m.confidence(c.Confidence, where)

		fwd := canon.Link{
			LinkID:        linkURIPrefix + c.ID + ":fwd",
			Source:        canon.Source{SourceNode: deviceURIPrefix + fromDev, SourceTp: tpURIPrefix + fromTP},
			Destination:   canon.Destination{DestNode: deviceURIPrefix + toDev, DestTp: tpURIPrefix + toTP},
			Confidence:    confidence,
			Reviewed:      c.Reviewed,
			DiscoveredVia: via,
		}
		rev := canon.Link{
			LinkID:        linkURIPrefix + c.ID + ":rev",
			Source:        canon.Source{SourceNode: deviceURIPrefix + toDev, SourceTp: tpURIPrefix + toTP},
			Destination:   canon.Destination{DestNode: deviceURIPrefix + fromDev, DestTp: tpURIPrefix + fromTP},
			Confidence:    confidence,
			Reviewed:      c.Reviewed,
			DiscoveredVia: via,
		}

		links = append(links, fwd, rev)
	}

	return links
}

// confidences are the values nsl-topology:confidence permits. A grade outside them
// would fail validation, so it is dropped and reported rather than emitted.
var confidences = map[string]bool{
	"confirmed": true,
	"candidate": true,
	"weak":      true,
}

// confidence validates a persisted grade. Empty is legitimate and common: a link a
// human specified by hand has no evidence to grade, because it states intent rather
// than reporting an observation.
func (m *mapper) confidence(c, where string) string {
	if c == "" {
		return ""
	}
	if !confidences[c] {
		m.warnf("%s: confidence %q is not one of confirmed/candidate/weak; dropped", where, c)
		return ""
	}
	return c
}

func (m *mapper) deviceIDByLabel(label string) (string, bool) {
	for _, d := range m.src.Devices {
		if d.Label == label {
			return d.ID, true
		}
	}
	return "", false
}

// provenance splits NSL-Graph's opaque provenance strings into their typed halves.
//
// The stored form is "<source>@<device>:<port>", e.g. "ssh-lldp@opnsense:igc1".
// The source must be one of the discovery-source enum values; anything else is a
// string the schema will not accept, and is reported rather than smuggled through.
func (m *mapper) provenance(raw []string, where string) []canon.DiscoveredVia {
	if len(raw) == 0 {
		return nil
	}

	seen := make(map[canon.DiscoveredVia]bool, len(raw))
	out := make([]canon.DiscoveredVia, 0, len(raw))

	for _, s := range raw {
		source, observedOn, found := strings.Cut(s, "@")
		if !found {
			m.warnf("%s: provenance %q is not of the form \"source@device:port\"; dropped", where, s)
			continue
		}
		if !discoverySources[source] {
			m.warnf("%s: provenance %q names an unknown source %q; dropped", where, s, source)
			continue
		}

		// source and observed-on together are the list key, so duplicates would be
		// an invalid instance.
		dv := canon.DiscoveredVia{Source: source, ObservedOn: observedOn}
		if seen[dv] {
			continue
		}
		seen[dv] = true
		out = append(out, dv)
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Source != out[j].Source {
			return out[i].Source < out[j].Source
		}
		return out[i].ObservedOn < out[j].ObservedOn
	})
	return out
}
