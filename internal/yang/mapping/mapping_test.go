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
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	e "nsl-graph/internal/repository/entities"
	"nsl-graph/internal/yang/canon"
)

// lab is a minimal but complete specification: two devices in different zones, one
// port each, one connection between them. Tests mutate a copy of it to isolate the
// behaviour under test.
func lab() Source {
	return Source{
		Brands:     []e.Brand{{ID: "b1", Name: "Siemens"}},
		ModelTypes: []e.ModelType{{ID: "mt1", Name: "Switch"}},
		OsTypes:    []e.OsType{{ID: "os1", Name: "openwrt"}},
		Owners:     []e.Owner{{ID: "o1", Name: "IT"}},
		ZoneTypes:  []e.ZoneType{{ID: "zt1", Name: "Security"}},
		Zones: []e.Zone{
			{ID: "z-hq", Name: "HQ", LocationType: "Security", Owner: "IT"},
			{ID: "z-dmz", Name: "DMZ", FatherID: "z-hq", LocationType: "Security", Owner: "IT"},
		},
		Models: []e.ModelDevice{
			{ID: "m1", Model: "XC206", Brand: "Siemens", ModelType: "Switch", OsType: "openwrt"},
		},
		ModelPorts: []e.ModelPort{
			{ID: "mp1", Name: "P1", Model: "XC206", Brand: "Siemens", Positionx: 0, Positiony: 0},
			{ID: "mp2", Name: "P2", Model: "XC206", Brand: "Siemens", Positionx: 1, Positiony: 0},
		},
		Devices: []e.Device{
			{ID: "d1", Label: "sw1", Model: "XC206", Brand: "Siemens", ZoneID: "z-dmz", Owner: "IT", Ips: []string{"10.0.2.20"}},
			{ID: "d2", Label: "sw2", Model: "XC206", Brand: "Siemens", ZoneID: "z-hq", Owner: "IT"},
		},
		DevicePorts: []e.DevicePort{
			{ID: "dp1", DeviceID: "d1", ModelID: "mp1", DevLabel: "sw1", PortName: "P1", MacAddress: "00:1B:1B:00:00:01"},
			{ID: "dp2", DeviceID: "d2", ModelID: "mp1", DevLabel: "sw2", PortName: "P1"},
		},
		Connections: []e.Connection{
			{
				ID: "c1", FromDevice: "sw1", FromModelPort: "P1", ToDevice: "sw2", ToModelPort: "P1",
				DiscoveredVia: []string{"ssh-lldp@sw1:P1", "snmp-lldp@sw2:P1"},
				Confidence:    "confirmed",
				Reviewed:      true,
			},
		},
	}
}

// network returns the single RFC 8345 network from a mapped root.
func network(t *testing.T, root *canon.Root) canon.Network {
	t.Helper()
	require.NotNil(t, root.Networks)
	require.Len(t, root.Networks.Network, 1)
	return root.Networks.Network[0]
}

func hasWarning(warnings []string, substr string) bool {
	for _, w := range warnings {
		if strings.Contains(w, substr) {
			return true
		}
	}
	return false
}

// -----------------------------------------------------------------------------
// Links: the unidirectional rule
// -----------------------------------------------------------------------------

// RFC 8345 links are unidirectional, so one Connection must become TWO links with the
// endpoints swapped. Emitting only one direction leaves the exported topology silently
// asymmetric -- valid against the schema, but wrong, and nothing else would catch it.
func TestConnectionBecomesTwoOppositeLinks(t *testing.T) {
	root, warnings := FromEntities(lab())
	assert.Empty(t, warnings)

	links := network(t, root).Links
	require.Len(t, links, 2, "one connection must produce exactly two links")

	fwd, rev := links[0], links[1]
	assert.Equal(t, "urn:nsl:link:c1:fwd", fwd.LinkID)
	assert.Equal(t, "urn:nsl:link:c1:rev", rev.LinkID)

	// The reverse link is the forward one with source and destination exchanged.
	assert.Equal(t, "urn:nsl:device:d1", fwd.Source.SourceNode)
	assert.Equal(t, "urn:nsl:tp:dp1", fwd.Source.SourceTp)
	assert.Equal(t, "urn:nsl:device:d2", fwd.Destination.DestNode)
	assert.Equal(t, "urn:nsl:tp:dp2", fwd.Destination.DestTp)

	assert.Equal(t, fwd.Destination.DestNode, rev.Source.SourceNode)
	assert.Equal(t, fwd.Destination.DestTp, rev.Source.SourceTp)
	assert.Equal(t, fwd.Source.SourceNode, rev.Destination.DestNode)
	assert.Equal(t, fwd.Source.SourceTp, rev.Destination.DestTp)
}

// A connection names its endpoints by device LABEL and port NAME, not by key. When
// that pair resolves to nothing the connection cannot be emitted -- the schema's
// leafref would not resolve -- so it is dropped and reported.
func TestConnectionWithUnresolvableEndpointIsDropped(t *testing.T) {
	src := lab()
	src.Connections[0].ToModelPort = "P99"

	root, warnings := FromEntities(src)

	assert.Empty(t, network(t, root).Links)
	assert.True(t, hasWarning(warnings, `no port "P99" on device "sw2"`), "got: %v", warnings)
}

// -----------------------------------------------------------------------------
// VLANs: the typed-identifier rules
// -----------------------------------------------------------------------------

// The contradiction found in real.db: the same VLAN on the same port recorded as BOTH
// tagged and untagged. A port either tags a VLAN's frames on egress or it does not.
// This is not a duplicate, and the warning must not call it one.
func TestVlanTaggedAndUntaggedIsReportedAsContradiction(t *testing.T) {
	src := lab()
	src.DevicePorts[0].VlanConfigs = []e.PortVlanConfig{
		{VlanNumber: "10", Tagged: true},
		{VlanNumber: "10", Tagged: false},
	}

	root, warnings := FromEntities(src)

	vlans := network(t, root).Nodes[0].TerminationPoints[0].VlanMembership
	require.Len(t, vlans, 1, "vlan-id is the list key; only one entry may survive")
	assert.Equal(t, uint16(10), vlans[0].VlanID)
	assert.True(t, vlans[0].Tagged, "the first recorded value is kept, deterministically")

	assert.True(t, hasWarning(warnings, "BOTH tagged and untagged"), "got: %v", warnings)
	assert.False(t, hasWarning(warnings, "same tagging"),
		"a contradiction must not be reported as a harmless duplicate")
}

func TestVlanExactDuplicateIsDropped(t *testing.T) {
	src := lab()
	src.DevicePorts[0].VlanConfigs = []e.PortVlanConfig{
		{VlanNumber: "10", Tagged: true},
		{VlanNumber: "10", Tagged: true},
	}

	root, warnings := FromEntities(src)

	assert.Len(t, network(t, root).Nodes[0].TerminationPoints[0].VlanMembership, 1)
	assert.True(t, hasWarning(warnings, "same tagging"), "got: %v", warnings)
}

// dot1q-types:vlanid is a uint16 in 1..4094. entities.PortVlanConfig.VlanNumber is a
// string, so every one of these is storable today and none of them is a VLAN.
func TestVlanRejectedValues(t *testing.T) {
	tests := []struct {
		name string
		vlan string
		want string // substring of the expected warning
	}{
		{"non-numeric", "eth0", "is not a number"},
		{"empty", "", "is not a number"},
		{"negative", "-1", "is not a number"},
		{"zero is reserved", "0", "out of range"},
		{"4095 is reserved", "4095", "out of range"},
		{"above the 12-bit range", "4999", "out of range"},
		{"far above uint16", "70000", "is not a number"}, // ParseUint(_, 10, 16) overflows
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := lab()
			src.DevicePorts[0].VlanConfigs = []e.PortVlanConfig{{VlanNumber: tt.vlan}}

			root, warnings := FromEntities(src)

			assert.Empty(t, network(t, root).Nodes[0].TerminationPoints[0].VlanMembership,
				"an invalid VLAN must not reach the document")
			assert.True(t, hasWarning(warnings, tt.want), "got: %v", warnings)
		})
	}
}

func TestVlansAreSortedAndTaggingPreserved(t *testing.T) {
	src := lab()
	src.DevicePorts[0].VlanConfigs = []e.PortVlanConfig{
		{VlanNumber: "20", Tagged: true},
		{VlanNumber: "10", Tagged: false},
		{VlanNumber: "4094", Tagged: true},
	}

	root, _ := FromEntities(src)

	assert.Equal(t, []canon.VlanMembership{
		{VlanID: 10, Tagged: false},
		{VlanID: 20, Tagged: true},
		{VlanID: 4094, Tagged: true},
	}, network(t, root).Nodes[0].TerminationPoints[0].VlanMembership)
}

// -----------------------------------------------------------------------------
// Leafrefs: dangling references must not be emitted
// -----------------------------------------------------------------------------

// Every cross-entity reference in the schema is a leafref, which MUST resolve. The
// domain model holds these as bare name strings with no integrity, so a real database
// contains names that point at nothing. Emitting one would produce a document the
// schema rejects, so it is dropped -- and reported, because the domain model cannot
// detect it at all.
func TestDanglingReferencesAreDroppedAndReported(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Source)
		want   string
		check  func(*testing.T, *canon.Root)
	}{
		{
			name:   "unknown device owner",
			mutate: func(s *Source) { s.Devices[0].Owner = "ghost" },
			want:   `owner "ghost" does not exist`,
			check: func(t *testing.T, r *canon.Root) {
				assert.Empty(t, r.Networks.Network[0].Nodes[0].Owner)
			},
		},
		{
			name:   "unknown zone on device",
			mutate: func(s *Source) { s.Devices[0].ZoneID = "z-ghost" },
			want:   `zone "z-ghost" does not exist`,
			check: func(t *testing.T, r *canon.Root) {
				assert.Empty(t, r.Networks.Network[0].Nodes[0].Zone)
			},
		},
		{
			name:   "unknown model on device",
			mutate: func(s *Source) { s.Devices[0].Model = "Ghost9000" },
			want:   `no model "Siemens"/"Ghost9000" exists`,
			check: func(t *testing.T, r *canon.Root) {
				assert.Empty(t, r.Networks.Network[0].Nodes[0].Model)
			},
		},
		{
			name:   "unknown brand on model",
			mutate: func(s *Source) { s.Models[0].Brand = "Ghost" },
			want:   `brand "Ghost" does not exist`,
			check: func(t *testing.T, r *canon.Root) {
				assert.Empty(t, r.Inventory.Models[0].Brand)
			},
		},
		{
			name:   "unknown os-type on model",
			mutate: func(s *Source) { s.Models[0].OsType = "plan9" },
			want:   `os-type "plan9" does not exist`,
			check: func(t *testing.T, r *canon.Root) {
				assert.Empty(t, r.Inventory.Models[0].OsType)
			},
		},
		{
			name:   "unknown zone-type on zone",
			mutate: func(s *Source) { s.Zones[0].LocationType = "Nowhere" },
			want:   `zone-type "Nowhere" does not exist`,
			check: func(t *testing.T, r *canon.Root) {
				assert.Empty(t, r.Inventory.Zones[0].ZoneType)
			},
		},
		{
			name:   "unknown parent zone",
			mutate: func(s *Source) { s.Zones[1].FatherID = "z-ghost" },
			want:   `parent zone "z-ghost" does not exist`,
			check: func(t *testing.T, r *canon.Root) {
				assert.Empty(t, r.Inventory.Zones[1].Parent)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := lab()
			tt.mutate(&src)

			root, warnings := FromEntities(src)

			assert.True(t, hasWarning(warnings, tt.want), "got: %v", warnings)
			tt.check(t, root)

			// The rest of the document must survive: one bad reference is not a
			// reason to lose the network.
			assert.Len(t, root.Networks.Network[0].Nodes, 2)
		})
	}
}

// A model-port whose (model, brand) pair matches no model cannot be placed in the
// tree at all -- model-port is nested under model.
func TestOrphanModelPortIsDropped(t *testing.T) {
	src := lab()
	src.ModelPorts = append(src.ModelPorts, e.ModelPort{
		ID: "mp-orphan", Name: "P9", Model: "Ghost9000", Brand: "Siemens",
	})

	root, warnings := FromEntities(src)

	require.Len(t, root.Inventory.Models, 1)
	assert.Len(t, root.Inventory.Models[0].ModelPorts, 2, "the orphan must not be attached")
	assert.True(t, hasWarning(warnings, `no model "Siemens"/"Ghost9000" exists`), "got: %v", warnings)
}

// -----------------------------------------------------------------------------
// Provenance
// -----------------------------------------------------------------------------

func TestProvenanceIsSplitIntoTypedHalves(t *testing.T) {
	root, warnings := FromEntities(lab())
	assert.Empty(t, warnings)

	via := network(t, root).Links[0].DiscoveredVia
	assert.Equal(t, []canon.DiscoveredVia{
		{Source: "snmp-lldp", ObservedOn: "sw2:P1"},
		{Source: "ssh-lldp", ObservedOn: "sw1:P1"},
	}, via, "sorted, and split on the '@'")
}

func TestProvenanceRejectsMalformedAndUnknownSources(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"no @ separator", "ssh-lldp", `not of the form`},
		{"source not in the enum", "carrier-pigeon@sw1:P1", `unknown source "carrier-pigeon"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src := lab()
			src.Connections[0].DiscoveredVia = []string{tt.raw}

			root, warnings := FromEntities(src)

			assert.Empty(t, network(t, root).Links[0].DiscoveredVia,
				"a value the discovery-source enum would reject must not be emitted")
			assert.True(t, hasWarning(warnings, tt.want), "got: %v", warnings)
		})
	}
}

// -----------------------------------------------------------------------------
// Ports: type normalisation and the band gate
// -----------------------------------------------------------------------------

// The schema gates `band` on `when "../port-type = 'wifi'"`, so a band on a wired port
// is a validation error rather than a field everyone agrees to ignore.
func TestBandIsOnlyEmittedForWifiPorts(t *testing.T) {
	src := lab()
	src.ModelPorts = []e.ModelPort{
		{ID: "mp1", Name: "P1", Model: "XC206", Brand: "Siemens", PortType: "", Band: "5GHz"},
		{ID: "mp2", Name: "radio0", Model: "XC206", Brand: "Siemens", PortType: "wifi", Band: "5GHz"},
	}

	root, _ := FromEntities(src)
	ports := root.Inventory.Models[0].ModelPorts
	require.Len(t, ports, 2)

	assert.Equal(t, "ethernet", ports[0].PortType, `the domain's "" means wired; the schema names it`)
	assert.Empty(t, ports[0].Band, "a band on a wired port would fail the schema's `when`")

	assert.Equal(t, "wifi", ports[1].PortType)
	assert.Equal(t, "5GHz", ports[1].Band)
}

// -----------------------------------------------------------------------------
// Encoding
// -----------------------------------------------------------------------------

// position-x/position-y of 0 must be EMITTED. Zero is a legitimate faceplate
// coordinate -- it is the top-left port -- so an `omitempty` on these tags would
// silently move every port at the origin. Only the marshalled JSON can see this; the
// struct looks identical either way.
func TestZeroPositionsSurviveMarshalling(t *testing.T) {
	root, _ := FromEntities(lab())

	out, err := json.Marshal(root)
	require.NoError(t, err)

	var doc map[string]any
	require.NoError(t, json.Unmarshal(out, &doc))

	inv := doc["nsl-inventory:inventory"].(map[string]any)
	model := inv["model"].([]any)[0].(map[string]any)
	port := model["model-port"].([]any)[0].(map[string]any)

	require.Contains(t, port, "position-x", "position-x=0 was dropped by omitempty")
	require.Contains(t, port, "position-y", "position-y=0 was dropped by omitempty")
	assert.Equal(t, float64(0), port["position-x"])
	assert.Equal(t, float64(0), port["position-y"])
}

// The presence containers that mark the network type encode as {} in RFC 7951, not as
// booleans -- which is why they are *struct{} rather than bool.
func TestNetworkTypeMarkersEncodeAsEmptyContainers(t *testing.T) {
	root, _ := FromEntities(lab())

	out, err := json.Marshal(root)
	require.NoError(t, err)

	assert.Contains(t, string(out), `"ietf-l2-topology:l2-topology":{}`)
	assert.Contains(t, string(out), `"nsl-topology:nsl-topology":{}`)
}

func TestEmptySourceProducesAValidEmptyDocument(t *testing.T) {
	root, warnings := FromEntities(Source{})

	assert.Empty(t, warnings)
	require.NotNil(t, root.Networks)

	net := network(t, root)
	assert.Equal(t, canon.NetworkID, net.NetworkID)
	assert.Empty(t, net.Nodes)
	assert.Empty(t, net.Links)

	_, err := json.Marshal(root)
	assert.NoError(t, err)
}

// The confidence grade and the review state now survive the round trip. They used to
// be computed during a scan and DISCARDED at import, so a weak link and a confirmed one
// were indistinguishable once in the database -- which made the confidence ladder, the
// whole point of multi-source discovery, a per-scan curiosity.
//
// Both links of a connection carry the same evidence: it is one cable, and the fact
// that RFC 8345 makes us describe it as two arrows does not make it two observations.
func TestConfidenceAndReviewSurviveTheMapping(t *testing.T) {
	src := lab()
	src.Connections[0].Confidence = "weak"
	src.Connections[0].Reviewed = false

	root, warnings := FromEntities(src)
	assert.Empty(t, warnings)

	links := network(t, root).Links
	require.Len(t, links, 2)
	for _, l := range links {
		assert.Equal(t, "weak", l.Confidence)
		assert.False(t, l.Reviewed)
	}
}

// A link a human specified by hand carries no confidence: there is no evidence to
// grade, because it states intent rather than reporting an observation. It is reviewed
// by definition -- someone said so.
func TestHandSpecifiedLinkHasNoConfidence(t *testing.T) {
	src := lab()
	src.Connections[0].Confidence = ""
	src.Connections[0].Reviewed = true
	src.Connections[0].DiscoveredVia = nil

	root, warnings := FromEntities(src)
	assert.Empty(t, warnings)

	for _, l := range network(t, root).Links {
		assert.Empty(t, l.Confidence)
		assert.True(t, l.Reviewed)
		assert.Empty(t, l.DiscoveredVia)
	}
}

// nsl-topology:confidence is an enumeration. A grade outside it would fail validation,
// so it is dropped and reported rather than smuggled into the document.
func TestUnknownConfidenceIsDropped(t *testing.T) {
	src := lab()
	src.Connections[0].Confidence = "pretty-sure"

	root, warnings := FromEntities(src)

	assert.Empty(t, network(t, root).Links[0].Confidence)
	assert.True(t, hasWarning(warnings, `confidence "pretty-sure" is not one of`), "got: %v", warnings)
}

// MAC addresses are typed yang:mac-address in the schema, whose canonical form is
// lower-case. The domain model stores whatever the device reported.
func TestMacAddressesAreLowercased(t *testing.T) {
	root, _ := FromEntities(lab())

	tp := network(t, root).Nodes[0].TerminationPoints[0]
	require.NotNil(t, tp.L2)
	assert.Equal(t, "00:1b:1b:00:00:01", tp.L2.MacAddress)
}
