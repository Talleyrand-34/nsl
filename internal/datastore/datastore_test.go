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

	"nsl-graph/internal/topology"
)

// edge builds a fully-resolved edge with the given confidence. Resolved is the normal
// case; the unresolved case is tested separately.
func edge(from, to, confidence string) topology.ConnectionEdge {
	return topology.ConnectionEdge{
		FromDevicePortID: "dp-" + from,
		ToDevicePortID:   "dp-" + to,
		FromLabel:        from,
		ToLabel:          to,
		Confidence:       confidence,
		RemoteResolved:   true,
	}
}

// The rule this package exists for. The forwarding database proves only that a MAC was
// seen on a port -- not that the device owning it is adjacent. It may be several hops
// away behind an unmanaged switch bridging traffic transparently, in which case the
// "link" is an invention.
func TestWeakEdgeIsRejectedUnlessReviewed(t *testing.T) {
	var c Candidate
	c.Stage(edge("sw1:P1", "sw2:P1", topology.ConfidenceWeak), false)

	violations := Validate(c)

	require.Len(t, violations, 1)
	assert.Equal(t, RuleWeakMustBeReviewed, violations[0].Rule)
	assert.Contains(t, violations[0].Detail, "forwarding database")
}

// A weak link an operator HAS looked at and accepted is legitimate. The rule exists to
// force the decision, not to forbid the outcome.
func TestWeakEdgeIsAcceptedOnceReviewed(t *testing.T) {
	var c Candidate
	c.Stage(edge("sw1:P1", "sw2:P1", topology.ConfidenceWeak), true)

	assert.Empty(t, Validate(c))
}

// Confirmed and candidate edges rest on a direct LLDP/CDP observation, so they need no
// human in the loop.
func TestStrongEdgesNeedNoReview(t *testing.T) {
	for _, confidence := range []string{topology.ConfidenceConfirmed, topology.ConfidenceCandidate} {
		t.Run(confidence, func(t *testing.T) {
			var c Candidate
			c.Stage(edge("sw1:P1", "sw2:P1", confidence), false)

			assert.Empty(t, Validate(c), "a %s edge is direct evidence and needs no review", confidence)
		})
	}
}

// An edge whose far end matches no device port names something we do not model --
// typically an unmanaged switch. Worth reporting; not worth persisting.
func TestUnresolvedEndpointIsRejected(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*topology.ConnectionEdge)
	}{
		{"remote not resolved", func(e *topology.ConnectionEdge) { e.RemoteResolved = false }},
		{"no from port", func(e *topology.ConnectionEdge) { e.FromDevicePortID = "" }},
		{"no to port", func(e *topology.ConnectionEdge) { e.ToDevicePortID = "" }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ed := edge("sw1:P1", "unknown", topology.ConfidenceConfirmed)
			tt.mutate(&ed)

			var c Candidate
			c.Stage(ed, true)

			violations := Validate(c)
			require.Len(t, violations, 1)
			assert.Equal(t, RuleEndpointsResolved, violations[0].Rule)
		})
	}
}

// Validate reports EVERY violation, not just the first. An operator repairing a scan
// wants the whole list, not a game of whack-a-mole.
func TestValidateReportsEveryViolation(t *testing.T) {
	var c Candidate
	c.Stage(edge("a:P1", "b:P1", topology.ConfidenceWeak), false)
	c.Stage(edge("c:P1", "d:P1", topology.ConfidenceWeak), false)

	assert.Len(t, Validate(c), 2)
}

// A single edge can break more than one rule, and both must be reported -- fixing the
// review without fixing the endpoint would otherwise look like progress.
func TestOneEdgeCanBreakSeveralRules(t *testing.T) {
	ed := edge("sw1:P1", "ghost", topology.ConfidenceWeak)
	ed.RemoteResolved = false

	var c Candidate
	c.Stage(ed, false)

	violations := Validate(c)
	require.Len(t, violations, 2)

	rules := []string{violations[0].Rule, violations[1].Rule}
	assert.Contains(t, rules, RuleWeakMustBeReviewed)
	assert.Contains(t, rules, RuleEndpointsResolved)
}

// A scan of a real network always turns up some edges that cannot be committed.
// Refusing the whole scan because of them would make discovery useless -- but so would
// committing them silently. Committable does neither.
func TestCommittablePartitionsRatherThanRefusingEverything(t *testing.T) {
	var c Candidate
	c.Stage(edge("good1:P1", "good2:P1", topology.ConfidenceConfirmed), false)
	c.Stage(edge("bad1:P1", "bad2:P1", topology.ConfidenceWeak), false) // unreviewed weak
	c.Stage(edge("good3:P1", "good4:P1", topology.ConfidenceCandidate), false)

	ok, rejected := Committable(c)

	require.Len(t, ok, 2, "the two sound edges must still be committable")
	assert.Equal(t, "good1:P1", ok[0].Edge.FromLabel)
	assert.Equal(t, "good3:P1", ok[1].Edge.FromLabel)

	require.Len(t, rejected, 1)
	assert.Equal(t, RuleWeakMustBeReviewed, rejected[0].Rule)
}

func TestCommittableAcceptsACleanCandidate(t *testing.T) {
	var c Candidate
	c.Stage(edge("sw1:P1", "sw2:P1", topology.ConfidenceConfirmed), false)

	ok, rejected := Committable(c)

	assert.Len(t, ok, 1)
	assert.Empty(t, rejected)
	assert.NoError(t, ViolationsError(rejected))
}

func TestEmptyCandidateIsValid(t *testing.T) {
	assert.Empty(t, Validate(Candidate{}))
	assert.NoError(t, ViolationsError(nil))
}

// The contradiction that real.db actually contains, nine times over.
//
// A port either tags a VLAN's frames on egress or it does not. Recording both is not a
// duplicate to be de-duplicated -- it is a contradiction, and it means one of the two is
// wrong. Whichever we picked, we would be guessing about how frames leave a switch port.
func TestVlanTaggedAndUntaggedOnOnePortIsRejected(t *testing.T) {
	var c Candidate
	c.StagePort("sw1", "eth0", []VLANMembership{
		{VLAN: "10", Tagged: true},
		{VLAN: "10", Tagged: false},
	})

	violations := Validate(c)

	require.Len(t, violations, 1)
	assert.Equal(t, RuleVlanTaggingIsConsistent, violations[0].Rule)
	assert.Equal(t, "sw1:eth0", violations[0].Target)
	assert.Contains(t, violations[0].Detail, "cannot do both")
}

// The same VLAN listed twice with the SAME tagging is merely redundant, not
// contradictory. It says nothing false, so it is not a violation.
func TestVlanRepeatedWithSameTaggingIsNotAViolation(t *testing.T) {
	var c Candidate
	c.StagePort("sw1", "eth0", []VLANMembership{
		{VLAN: "10", Tagged: true},
		{VLAN: "10", Tagged: true},
	})

	assert.Empty(t, Validate(c))
}

// The same VLAN on DIFFERENT ports may of course be tagged on one and untagged on the
// other -- that is an ordinary trunk-and-access arrangement, not a contradiction.
func TestSameVlanMayDifferAcrossPorts(t *testing.T) {
	var c Candidate
	c.StagePort("sw1", "eth0", []VLANMembership{{VLAN: "10", Tagged: true}})
	c.StagePort("sw1", "eth1", []VLANMembership{{VLAN: "10", Tagged: false}})

	assert.Empty(t, Validate(c))
}

// dot1q-types:vlanid is a uint16 in 1..4094. The domain model holds VLAN ids as strings,
// so every one of these is storable today.
func TestInvalidVlanIdIsRejected(t *testing.T) {
	for _, vlan := range []string{"0", "4095", "4999", "70000", "-1", "eth0", ""} {
		t.Run(vlan, func(t *testing.T) {
			var c Candidate
			c.StagePort("sw1", "eth0", []VLANMembership{{VLAN: vlan}})

			violations := Validate(c)
			require.Len(t, violations, 1)
			assert.Equal(t, RuleVlanIsValid, violations[0].Rule)
		})
	}
}

func TestValidVlanIdsAreAccepted(t *testing.T) {
	var c Candidate
	c.StagePort("sw1", "eth0", []VLANMembership{
		{VLAN: "1", Tagged: false},
		{VLAN: "4094", Tagged: true},
		{VLAN: " 20 ", Tagged: true}, // whitespace is tolerated, not a reason to refuse
	})

	assert.Empty(t, Validate(c))
}

// The error must name the edge and the reason, because it is read by someone deciding
// what to do about a switch.
func TestViolationsErrorNamesTheEdgeAndTheReason(t *testing.T) {
	var c Candidate
	c.Stage(edge("sw1:P1", "sw2:P1", topology.ConfidenceWeak), false)

	err := ViolationsError(Validate(c))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "sw1:P1 <-> sw2:P1")
	assert.Contains(t, err.Error(), "forwarding database")
	assert.Contains(t, err.Error(), RuleWeakMustBeReviewed)
}
