// SPDX-License-Identifier: AGPL-3.0-or-later
package datastore

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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nsl-graph/internal/yang/canon"
)

// tree builds a canonical tree with the given nodes and links.
func tree(nodes []canon.Node, links ...canon.Link) *canon.Root {
	return &canon.Root{
		Networks: &canon.Networks{
			Network: []canon.Network{{
				NetworkID: canon.NetworkID,
				Nodes:     nodes,
				Links:     links,
			}},
		},
	}
}

// node builds a device with one port carrying the given VLANs.
func node(name, port string, vlans ...canon.VlanMembership) canon.Node {
	return canon.Node{
		NodeID: "urn:nsl:device:" + name,
		L2:     &canon.L2NodeAttributes{Name: name},
		TerminationPoints: []canon.TerminationPoint{{
			TpID:           "urn:nsl:tp:" + name + "-" + port,
			L2:             &canon.L2TPAttributes{InterfaceName: port},
			VlanMembership: vlans,
		}},
	}
}

func link(fromNode, fromPort, toNode, toPort string) canon.Link {
	return canon.Link{
		LinkID:      "urn:nsl:link:" + fromNode + "-" + toNode,
		Source:      canon.Source{SourceNode: "urn:nsl:device:" + fromNode, SourceTp: "urn:nsl:tp:" + fromNode + "-" + fromPort},
		Destination: canon.Destination{DestNode: "urn:nsl:device:" + toNode, DestTp: "urn:nsl:tp:" + toNode + "-" + toPort},
	}
}

func opsOf(changes []Change) []Op {
	out := make([]Op, len(changes))
	for i, c := range changes {
		out[i] = c.Op
	}
	return out
}

func TestIdenticalTreesDiffToNothing(t *testing.T) {
	a := tree([]canon.Node{node("sw1", "P1", canon.VlanMembership{VlanID: 10, Tagged: true})})
	b := tree([]canon.Node{node("sw1", "P1", canon.VlanMembership{VlanID: 10, Tagged: true})})

	assert.Empty(t, Diff(a, b), "a network that matches its specification has nothing to report")
}

// A device in the spec that the scan did not find: it is off, unplugged, or gone.
func TestDeviceInSpecButNotObserved(t *testing.T) {
	intended := tree([]canon.Node{node("sw1", "P1"), node("sw2", "P1")})
	observed := tree([]canon.Node{node("sw1", "P1")})

	changes := Diff(intended, observed)

	require.Len(t, changes, 1)
	assert.Equal(t, OpMissing, changes[0].Op)
	assert.Contains(t, changes[0].Path, "urn:nsl:device:sw2")
	assert.Equal(t, "sw2", changes[0].Intended, "report the device NAME, not the URN -- "+
		"an operator at a rack knows sw2, not a UUID")
}

// A device the scan found that nobody specified: rogue hardware, or an undocumented
// addition. Either way, someone should know.
func TestDeviceObservedButNotInSpec(t *testing.T) {
	intended := tree([]canon.Node{node("sw1", "P1")})
	observed := tree([]canon.Node{node("sw1", "P1"), node("rogue", "P1")})

	changes := Diff(intended, observed)

	require.Len(t, changes, 1)
	assert.Equal(t, OpUnexpected, changes[0].Op)
	assert.Equal(t, "rogue", changes[0].Observed)
}

// The comparison that matters most in practice. A VLAN missing from a trunk is the most
// common way a network drifts from its specification, and the hardest to see by eye.
func TestVlanMissingFromPort(t *testing.T) {
	intended := tree([]canon.Node{node("sw1", "P1",
		canon.VlanMembership{VlanID: 10, Tagged: true},
		canon.VlanMembership{VlanID: 20, Tagged: true},
	)})
	observed := tree([]canon.Node{node("sw1", "P1",
		canon.VlanMembership{VlanID: 10, Tagged: true},
	)})

	changes := Diff(intended, observed)

	require.Len(t, changes, 1)
	assert.Equal(t, OpMissing, changes[0].Op)
	assert.Contains(t, changes[0].Path, "vlan-membership[vlan-id='20']")
	assert.Equal(t, "tagged", changes[0].Intended)
}

func TestVlanPresentButNotSpecified(t *testing.T) {
	intended := tree([]canon.Node{node("sw1", "P1")})
	observed := tree([]canon.Node{node("sw1", "P1", canon.VlanMembership{VlanID: 66, Tagged: true})})

	changes := Diff(intended, observed)

	require.Len(t, changes, 1)
	assert.Equal(t, OpUnexpected, changes[0].Op)
	assert.Contains(t, changes[0].Path, "vlan-id='66'")
}

// The nastiest case, and the reason tagging is compared rather than just membership.
// The port is in the right VLAN but trunking where it should be an access port. Frames
// flow -- and they flow wrong, which is worse than not flowing at all, because
// everything looks up.
func TestSameVlanDifferentTagging(t *testing.T) {
	intended := tree([]canon.Node{node("sw1", "P1", canon.VlanMembership{VlanID: 10, Tagged: false})})
	observed := tree([]canon.Node{node("sw1", "P1", canon.VlanMembership{VlanID: 10, Tagged: true})})

	changes := Diff(intended, observed)

	require.Len(t, changes, 1)
	assert.Equal(t, OpDiffers, changes[0].Op)
	assert.Equal(t, "untagged", changes[0].Intended)
	assert.Equal(t, "tagged", changes[0].Observed)
}

func TestPortMissingFromDevice(t *testing.T) {
	intended := tree([]canon.Node{{
		NodeID: "urn:nsl:device:sw1",
		L2:     &canon.L2NodeAttributes{Name: "sw1"},
		TerminationPoints: []canon.TerminationPoint{
			{TpID: "urn:nsl:tp:sw1-P1", L2: &canon.L2TPAttributes{InterfaceName: "P1"}},
			{TpID: "urn:nsl:tp:sw1-P2", L2: &canon.L2TPAttributes{InterfaceName: "P2"}},
		},
	}})
	observed := tree([]canon.Node{node("sw1", "P1")})

	changes := Diff(intended, observed)

	require.Len(t, changes, 1)
	assert.Equal(t, OpMissing, changes[0].Op)
	assert.Equal(t, "P2", changes[0].Intended)
}

// A cable in the spec that is not in the network: someone unplugged it.
func TestLinkMissing(t *testing.T) {
	nodes := []canon.Node{node("sw1", "P1"), node("sw2", "P1")}
	intended := tree(nodes, link("sw1", "P1", "sw2", "P1"))
	observed := tree(nodes)

	changes := Diff(intended, observed)

	require.Len(t, changes, 1)
	assert.Equal(t, OpMissing, changes[0].Op)
	assert.Contains(t, changes[0].Path, "link[")
}

// The direction-independence rule. RFC 8345 links are unidirectional and one cable is
// exported as two of them, so a naive diff keyed on link-id would report every real
// difference twice -- and would call a cable seen from the other end a different cable.
func TestLinkDirectionDoesNotMatter(t *testing.T) {
	nodes := []canon.Node{node("sw1", "P1"), node("sw2", "P1")}
	intended := tree(nodes, link("sw1", "P1", "sw2", "P1"))
	observed := tree(nodes, link("sw2", "P1", "sw1", "P1")) // the same cable, reported from the far end

	assert.Empty(t, Diff(intended, observed),
		"the same cable seen from either end is one cable, not two differences")
}

// Both arrows of one connection collapse to a single entry, so a cable that is genuinely
// missing is reported once rather than twice.
func TestBothDirectionsOfOneCableCollapse(t *testing.T) {
	nodes := []canon.Node{node("sw1", "P1"), node("sw2", "P1")}
	intended := tree(nodes,
		link("sw1", "P1", "sw2", "P1"),
		link("sw2", "P1", "sw1", "P1"),
	)
	observed := tree(nodes)

	changes := Diff(intended, observed)

	assert.Len(t, changes, 1, "one cable, one report -- not one per RFC 8345 arrow")
}

// Output must be stable: a diff you cannot diff is not much use, and an operator
// re-running an audit should not see the same findings shuffled.
func TestDiffIsDeterministic(t *testing.T) {
	intended := tree([]canon.Node{
		node("sw3", "P1", canon.VlanMembership{VlanID: 30, Tagged: true}),
		node("sw1", "P1", canon.VlanMembership{VlanID: 10, Tagged: true}),
		node("sw2", "P1", canon.VlanMembership{VlanID: 20, Tagged: true}),
	})
	observed := tree([]canon.Node{node("sw1", "P1")})

	first := Diff(intended, observed)
	for i := 0; i < 20; i++ {
		assert.Equal(t, first, Diff(intended, observed),
			"map iteration order must not leak into the output")
	}
	assert.NotEmpty(t, first)
}

func TestEmptyTreesAreHandled(t *testing.T) {
	assert.Empty(t, Diff(nil, nil))
	assert.Empty(t, Diff(&canon.Root{}, &canon.Root{}))

	// An empty spec against a populated network: everything out there is unexpected.
	observed := tree([]canon.Node{node("sw1", "P1")})
	changes := Diff(&canon.Root{}, observed)
	require.Len(t, changes, 1)
	assert.Equal(t, []Op{OpUnexpected}, opsOf(changes))
}

// The rendered form is what an operator actually reads.
func TestChangeRendersReadably(t *testing.T) {
	assert.Contains(t,
		Change{Op: OpMissing, Path: "/x", Intended: "sw2"}.String(),
		"intended, not observed")
	assert.Contains(t,
		Change{Op: OpUnexpected, Path: "/x", Observed: "rogue"}.String(),
		"observed, not intended")
	assert.Contains(t,
		Change{Op: OpDiffers, Path: "/x", Intended: "untagged", Observed: "tagged"}.String(),
		"intended untagged, observed tagged")
}

// Every change carries BOTH addresses: the instance-identifier for the schema and the
// machine, and a human location for the operator. Nobody has ever found a switch by its
// UUID, and the path is nothing but UUIDs.
func TestEveryChangeCarriesAHumanLocation(t *testing.T) {
	intended := tree([]canon.Node{
		node("sw1", "eth0",
			canon.VlanMembership{VlanID: 10, Tagged: false},
			canon.VlanMembership{VlanID: 20, Tagged: true},
		),
		node("sw2", "eth0"),
	}, link("sw1", "eth0", "sw2", "eth0"))

	observed := tree([]canon.Node{
		node("sw1", "eth0", canon.VlanMembership{VlanID: 10, Tagged: true}),
		node("rogue", "eth0"),
	})

	changes := Diff(intended, observed)
	require.NotEmpty(t, changes)

	for _, c := range changes {
		assert.NotEmpty(t, c.Where, "change at %s has no human location", c.Path)
		assert.NotContains(t, c.Where, "urn:nsl:", "Where must not leak URNs: %q", c.Where)
		assert.NotEmpty(t, c.Path, "change at %q has no machine path", c.Where)
	}

	// And they say what an operator would say.
	var located []string
	for _, c := range changes {
		located = append(located, c.Where)
	}
	assert.Contains(t, located, "sw1/eth0 VLAN 10", "the tagging change is on sw1's eth0, VLAN 10")
	assert.Contains(t, located, "sw1/eth0 VLAN 20", "VLAN 20 is missing from sw1's eth0")
	assert.Contains(t, located, "device sw2", "sw2 was not observed")
	assert.Contains(t, located, "device rogue", "rogue was not specified")
	assert.Contains(t, located, "cable sw1/eth0 <-> sw2/eth0", "the cable names both ends by port")
}
