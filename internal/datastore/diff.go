// diff.go: compare two canonical trees — what the network should be against what it is.
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
	"fmt"
	"sort"

	"nsl-graph/internal/yang/canon"
)

// Op is what the observed network would have to do to match the intended one.
type Op string

const (
	// OpMissing: intended, but not observed. The network does not match the spec.
	OpMissing Op = "missing"
	// OpUnexpected: observed, but not intended. Something is out there that should not be.
	OpUnexpected Op = "unexpected"
	// OpDiffers: present in both, with different values.
	OpDiffers Op = "differs"
)

// Change is one difference between the intended and observed trees.
//
// It carries two addresses for the same thing, because two different readers need it:
//
//	Path  a YANG instance-identifier -- the address the SCHEMA uses, so a change
//	      reported here names the node a yanglint error would name. Precise, stable,
//	      machine-readable, and full of UUIDs.
//	Where the same location as a human would say it: "sw1/eth0 VLAN 20". This is what
//	      an operator standing at a rack can act on. Nobody has ever found a switch by
//	      its UUID.
type Change struct {
	Op       Op
	Path     string
	Where    string
	Intended string // rendered value; empty for OpUnexpected
	Observed string // rendered value; empty for OpMissing
}

// String renders the change for a human, using Where.
func (c Change) String() string {
	at := c.Where
	if at == "" {
		at = c.Path
	}
	switch c.Op {
	case OpMissing:
		return fmt.Sprintf("- %s: %s — intended, not observed", at, c.Intended)
	case OpUnexpected:
		return fmt.Sprintf("+ %s: %s — observed, not intended", at, c.Observed)
	default:
		return fmt.Sprintf("~ %s: intended %s, observed %s", at, c.Intended, c.Observed)
	}
}

// Diff compares an intended tree against an observed one and reports what does not
// match. The result is sorted by path, so two runs over the same pair of trees produce
// byte-identical output -- a diff you cannot diff is not much use.
//
// It compares what a scan can actually see: which devices exist, which ports they have,
// what VLANs those ports carry, and how they are cabled. Catalogue entries (brands,
// model types) are not compared, because a scan has no opinion about them.
func Diff(intended, observed *canon.Root) []Change {
	var changes []Change

	changes = append(changes, diffNodes(intended, observed)...)
	changes = append(changes, diffLinks(intended, observed)...)

	sort.Slice(changes, func(i, j int) bool {
		if changes[i].Path != changes[j].Path {
			return changes[i].Path < changes[j].Path
		}
		return changes[i].Op < changes[j].Op
	})
	return changes
}

// nodesOf indexes a tree's devices by node-id.
func nodesOf(r *canon.Root) map[string]canon.Node {
	out := map[string]canon.Node{}
	if r == nil || r.Networks == nil {
		return out
	}
	for _, net := range r.Networks.Network {
		for _, n := range net.Nodes {
			out[n.NodeID] = n
		}
	}
	return out
}

// linksOf indexes a tree's links by a direction-independent key.
//
// RFC 8345 links are unidirectional and one connection is exported as two of them, so
// keying by link-id would report every real difference twice and would also treat a
// cable seen from the other end as a different cable. Keying on the unordered pair of
// endpoints is what makes the diff say "this cable" rather than "this arrow".
func linksOf(r *canon.Root) map[string]canon.Link {
	out := map[string]canon.Link{}
	if r == nil || r.Networks == nil {
		return out
	}
	for _, net := range r.Networks.Network {
		for _, l := range net.Links {
			out[endpointKey(l)] = l
		}
	}
	return out
}

// endpointKey identifies a cable regardless of which way the link points.
func endpointKey(l canon.Link) string {
	a := l.Source.SourceNode + "|" + l.Source.SourceTp
	b := l.Destination.DestNode + "|" + l.Destination.DestTp
	if a > b {
		a, b = b, a
	}
	return a + " <-> " + b
}

// nodeName prefers the human-readable device name over the URN, because an operator
// standing at a rack knows "sw1", not "urn:nsl:device:5258c7b6-...".
func nodeName(n canon.Node) string {
	if n.L2 != nil && n.L2.Name != "" {
		return n.L2.Name
	}
	return n.NodeID
}

func nodePath(id string) string {
	return fmt.Sprintf("/networks/network/node[node-id='%s']", id)
}

func tpPath(nodeID, tpID string) string {
	return nodePath(nodeID) + fmt.Sprintf("/termination-point[tp-id='%s']", tpID)
}

func diffNodes(intended, observed *canon.Root) []Change {
	var changes []Change

	in, obs := nodesOf(intended), nodesOf(observed)

	for id, want := range in {
		got, present := obs[id]
		if !present {
			changes = append(changes, Change{
				Op:       OpMissing,
				Path:     nodePath(id),
				Where:    "device " + nodeName(want),
				Intended: nodeName(want),
			})
			continue
		}
		changes = append(changes, diffTerminationPoints(id, want, got)...)
	}

	for id, got := range obs {
		if _, present := in[id]; !present {
			changes = append(changes, Change{
				Op:       OpUnexpected,
				Path:     nodePath(id),
				Where:    "device " + nodeName(got),
				Observed: nodeName(got),
			})
		}
	}

	return changes
}

func diffTerminationPoints(nodeID string, want, got canon.Node) []Change {
	var changes []Change

	dev := nodeName(want)

	index := func(n canon.Node) map[string]canon.TerminationPoint {
		out := map[string]canon.TerminationPoint{}
		for _, tp := range n.TerminationPoints {
			out[tp.TpID] = tp
		}
		return out
	}
	in, obs := index(want), index(got)

	for id, wantTP := range in {
		gotTP, present := obs[id]
		if !present {
			changes = append(changes, Change{
				Op:       OpMissing,
				Path:     tpPath(nodeID, id),
				Where:    dev + "/" + portName(wantTP),
				Intended: portName(wantTP),
			})
			continue
		}
		changes = append(changes, diffVlans(nodeID, id, dev, wantTP, gotTP)...)
	}

	for id, gotTP := range obs {
		if _, present := in[id]; !present {
			changes = append(changes, Change{
				Op:       OpUnexpected,
				Path:     tpPath(nodeID, id),
				Where:    dev + "/" + portName(gotTP),
				Observed: portName(gotTP),
			})
		}
	}

	return changes
}

func portName(tp canon.TerminationPoint) string {
	if tp.L2 != nil && tp.L2.InterfaceName != "" {
		return tp.L2.InterfaceName
	}
	return tp.TpID
}

// diffVlans is the comparison that matters most in practice: a VLAN missing from a
// trunk, or present on a port that should not carry it, is the single most common way a
// network drifts from its specification -- and the hardest to see by eye.
func diffVlans(nodeID, tpID, dev string, want, got canon.TerminationPoint) []Change {
	var changes []Change

	index := func(tp canon.TerminationPoint) map[uint16]canon.VlanMembership {
		out := map[uint16]canon.VlanMembership{}
		for _, v := range tp.VlanMembership {
			out[v.VlanID] = v
		}
		return out
	}
	in, obs := index(want), index(got)

	path := func(vid uint16) string {
		return tpPath(nodeID, tpID) + fmt.Sprintf("/vlan-membership[vlan-id='%d']", vid)
	}
	where := func(vid uint16) string {
		return fmt.Sprintf("%s/%s VLAN %d", dev, portName(want), vid)
	}

	for vid, wantV := range in {
		gotV, present := obs[vid]
		if !present {
			changes = append(changes, Change{
				Op:       OpMissing,
				Path:     path(vid),
				Where:    where(vid),
				Intended: tagging(wantV.Tagged),
			})
			continue
		}
		// Same VLAN, different tagging: the port is in the right VLAN but trunking when
		// it should be an access port, or the reverse. Frames will flow, and they will
		// flow wrong -- which is worse than not flowing at all, because it looks fine.
		if wantV.Tagged != gotV.Tagged {
			changes = append(changes, Change{
				Op:       OpDiffers,
				Path:     path(vid),
				Where:    where(vid),
				Intended: tagging(wantV.Tagged),
				Observed: tagging(gotV.Tagged),
			})
		}
	}

	for vid, gotV := range obs {
		if _, present := in[vid]; !present {
			changes = append(changes, Change{
				Op:       OpUnexpected,
				Path:     path(vid),
				Where:    where(vid),
				Observed: tagging(gotV.Tagged),
			})
		}
	}

	return changes
}

func tagging(tagged bool) string {
	if tagged {
		return "tagged"
	}
	return "untagged"
}

func diffLinks(intended, observed *canon.Root) []Change {
	var changes []Change

	in, obs := linksOf(intended), linksOf(observed)
	names := nodesOf(intended)
	for id, n := range nodesOf(observed) {
		if _, ok := names[id]; !ok {
			names[id] = n
		}
	}

	// ports maps a tp-id to its interface name, across both trees, so a cable can be
	// named the way it is patched: "sw1/eth0 <-> fw1/igc1".
	ports := map[string]string{}
	for _, n := range names {
		for _, tp := range n.TerminationPoints {
			ports[tp.TpID] = portName(tp)
		}
	}

	endpoint := func(nodeID, tpID string) string {
		dev := nodeID
		if n, ok := names[nodeID]; ok {
			dev = nodeName(n)
		}
		if p, ok := ports[tpID]; ok {
			return dev + "/" + p
		}
		return dev
	}

	describe := func(l canon.Link) string {
		return endpoint(l.Source.SourceNode, l.Source.SourceTp) + " <-> " +
			endpoint(l.Destination.DestNode, l.Destination.DestTp)
	}

	for key, wantL := range in {
		if _, present := obs[key]; !present {
			changes = append(changes, Change{
				Op:       OpMissing,
				Path:     "/networks/network/link[" + key + "]",
				Where:    "cable " + describe(wantL),
				Intended: describe(wantL),
			})
		}
	}

	for key, gotL := range obs {
		if _, present := in[key]; !present {
			changes = append(changes, Change{
				Op:       OpUnexpected,
				Path:     "/networks/network/link[" + key + "]",
				Where:    "cable " + describe(gotL),
				Observed: describe(gotL),
			})
		}
	}

	return changes
}
