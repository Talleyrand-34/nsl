// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

// Package canon is the canonical, YANG-shaped representation of an NSL-Graph
// specification.
//
// These structs mirror the data tree defined by internal/yang/modules -- RFC 8345
// (ietf-network, ietf-network-topology) and RFC 8944 (ietf-l2-topology), augmented
// by nsl-topology and nsl-inventory. Their JSON tags are the RFC 7951 encoding, so
// marshalling one of these with encoding/json produces a document that yanglint
// validates against the modules directly.
//
// This is deliberately NOT the domain model. entities.* remains the domain and
// remains what CloverDB persists; canon is a projection of it onto the standard,
// built on demand by internal/yang/mapping. Keeping the two apart is what confines
// the whole YANG adoption to this directory.
//
// # RFC 7951 encoding rules that shape the tags
//
// A node's name is qualified by its module ("module:name") whenever the module
// differs from that of its parent, and never otherwise. So "networks" carries
// "ietf-network:", the augmented leaves carry "nsl-topology:" or "nsl-inventory:",
// and everything sitting inside its own module's subtree is bare.
//
// An empty (presence) container encodes as {}, which is why the network-type
// markers are *struct{} rather than bool.
package canon

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

// NetworkID is the single RFC 8345 network holding the whole specification.
//
// One network, not one per zone: RFC 8345 requires both endpoints of a link to be
// in the same network ("Must be in the same topology"), so zones-as-networks would
// make every inter-zone link unrepresentable. Zones are an inventory catalogue that
// nodes reference instead. See internal/yang/README.md.
const NetworkID = "urn:nsl:net:specification"

// Root is a complete NSL-Graph specification in RFC 7951 form.
type Root struct {
	Inventory *Inventory `json:"nsl-inventory:inventory,omitempty"`
	Networks  *Networks  `json:"ietf-network:networks,omitempty"`
}

// -----------------------------------------------------------------------------
// nsl-inventory
// -----------------------------------------------------------------------------

// Inventory holds the catalogues every reference in the topology tree resolves
// against. Each such reference is a YANG leafref, so a name absent from here is a
// validation error rather than a dangling string.
type Inventory struct {
	Brands     []Named `json:"brand,omitempty"`
	ModelTypes []Named `json:"model-type,omitempty"`
	OsTypes    []Named `json:"os-type,omitempty"`
	Owners     []Named `json:"owner,omitempty"`
	ZoneTypes  []Named `json:"zone-type,omitempty"`
	Zones      []Zone  `json:"zone,omitempty"`
	Models     []Model `json:"model,omitempty"`
}

// Named is a catalogue entry keyed solely by its name.
type Named struct {
	Name string `json:"name"`
}

// Zone is an administrative, security or location grouping of devices.
type Zone struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Parent   string `json:"parent,omitempty"`
	ZoneType string `json:"zone-type,omitempty"`
	Owner    string `json:"owner,omitempty"`
}

// Model is a catalogue entry for a device model, with its physical port template.
type Model struct {
	ID         string      `json:"id"`
	Name       string      `json:"name"`
	Brand      string      `json:"brand,omitempty"`
	ModelType  string      `json:"model-type,omitempty"`
	OsType     string      `json:"os-type,omitempty"`
	ModelPorts []ModelPort `json:"model-port,omitempty"`
}

// ModelPort is one port in a model's template: the ports every device of that
// model has, and where they sit on the faceplate.
type ModelPort struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	// PositionX and PositionY are not omitempty: 0 is a legitimate faceplate
	// coordinate, and dropping it would silently move the port.
	PositionX                int    `json:"position-x"`
	PositionY                int    `json:"position-y"`
	PortType                 string `json:"port-type,omitempty"`
	Band                     string `json:"band,omitempty"`
	AllowMultipleConnections bool   `json:"allow-multiple-connections,omitempty"`
}

// -----------------------------------------------------------------------------
// ietf-network / ietf-network-topology / ietf-l2-topology
// -----------------------------------------------------------------------------

// Networks is the RFC 8345 top-level container.
type Networks struct {
	Network []Network `json:"network"`
}

// Network is the single topology holding every device and every link.
type Network struct {
	NetworkID    string       `json:"network-id"`
	NetworkTypes NetworkTypes `json:"network-types"`
	Nodes        []Node       `json:"node,omitempty"`
	Links        []Link       `json:"ietf-network-topology:link,omitempty"`
}

// NetworkTypes marks what kind of topology this network is. Both markers are
// presence containers, so each encodes as {} when set and vanishes when nil.
type NetworkTypes struct {
	L2Topology  *struct{} `json:"ietf-l2-topology:l2-topology,omitempty"`
	NSLTopology *struct{} `json:"nsl-topology:nsl-topology,omitempty"`
}

// Node is a device: an RFC 8345 node, and an instance of an inventory Model.
type Node struct {
	NodeID string `json:"node-id"`

	// L2 gives us the device's name and management addresses, typed, from RFC 8944
	// rather than remodelled here.
	L2 *L2NodeAttributes `json:"ietf-l2-topology:l2-node-attributes,omitempty"`

	Zone  string `json:"nsl-inventory:zone,omitempty"`
	Model string `json:"nsl-inventory:model,omitempty"`
	Owner string `json:"nsl-inventory:owner,omitempty"`

	Unmanaged   bool   `json:"nsl-topology:unmanaged,omitempty"`
	Invisible   bool   `json:"nsl-topology:invisible,omitempty"`
	ScanProfile string `json:"nsl-topology:scan-profile,omitempty"`

	TerminationPoints []TerminationPoint `json:"ietf-network-topology:termination-point,omitempty"`
}

// L2NodeAttributes are the RFC 8944 device attributes.
type L2NodeAttributes struct {
	Name string `json:"name,omitempty"`
	// ManagementAddress is inet:ip-address in the schema, so a malformed address
	// fails validation rather than being stored as an arbitrary string.
	ManagementAddress []string `json:"management-address,omitempty"`
}

// TerminationPoint is a device port.
type TerminationPoint struct {
	TpID string `json:"tp-id"`

	// L2 gives us the port name and MAC, typed (yang:mac-address), from RFC 8944.
	L2 *L2TPAttributes `json:"ietf-l2-topology:l2-termination-point-attributes,omitempty"`

	ModelPort                string           `json:"nsl-topology:model-port,omitempty"`
	AllowMultipleConnections bool             `json:"nsl-topology:allow-multiple-connections,omitempty"`
	VlanMembership           []VlanMembership `json:"nsl-topology:vlan-membership,omitempty"`
}

// L2TPAttributes are the RFC 8944 port attributes.
type L2TPAttributes struct {
	InterfaceName string `json:"interface-name,omitempty"`
	MacAddress    string `json:"mac-address,omitempty"`
}

// VlanMembership is one VLAN this port belongs to, and whether it is tagged.
//
// VlanID is a uint16 because the schema types it dot1q-types:vlanid (range
// 1..4094). entities.PortVlanConfig.VlanNumber is a string, which is how "4999"
// and "eth0" become storable today.
type VlanMembership struct {
	VlanID uint16 `json:"vlan-id"`
	Tagged bool   `json:"tagged"`
}

// Link is one direction of a connection.
//
// RFC 8345 links are unidirectional, so one entities.Connection becomes TWO of
// these. Forgetting that is the easiest way to make the export lossy.
type Link struct {
	LinkID      string      `json:"link-id"`
	Source      Source      `json:"source"`
	Destination Destination `json:"destination"`

	Confidence    string          `json:"nsl-topology:confidence,omitempty"`
	Reviewed      bool            `json:"nsl-topology:reviewed,omitempty"`
	DiscoveredVia []DiscoveredVia `json:"nsl-topology:discovered-via,omitempty"`
}

// Source is a link's origin endpoint. Both leaves are leafrefs into this same
// network -- the referential integrity Connection.FromDevice (a bare name string)
// does not have.
type Source struct {
	SourceNode string `json:"source-node,omitempty"`
	SourceTp   string `json:"source-tp,omitempty"`
}

// Destination is a link's far endpoint.
type Destination struct {
	DestNode string `json:"dest-node,omitempty"`
	DestTp   string `json:"dest-tp,omitempty"`
}

// DiscoveredVia records one observation of a link: which evidence source saw it,
// and from which endpoint. NSL-Graph stores this today as an opaque string of the
// form "ssh-lldp@opnsense:igc1"; here it is split into its two typed halves.
type DiscoveredVia struct {
	Source     string `json:"source"`
	ObservedOn string `json:"observed-on"`
}
