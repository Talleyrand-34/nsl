// Package datastore applies NETCONF's discipline to NSL-Graph's own database, without
// implementing any of NETCONF's protocol.
//
// # What is borrowed, and what is not
//
// Borrowed, from RFC 8342 (NMDA) and RFC 6241 (NETCONF):
//
//   - intended vs observed. The specification says what the network SHOULD be; a scan
//     says what it IS. NSL-Graph has always held both, mixed into the same structs and
//     separated only by convention. NMDA names the distinction and stamps every item
//     with an origin: `intended` (it is there because someone said so) or `learned`
//     (it is there because a scan found it).
//
//   - candidate -> validate -> commit. Changes are staged, checked as a whole, and only
//     then applied. A candidate that breaks a rule is rejected entirely rather than
//     half-applied.
//
// NOT borrowed: the protocol. There is no NETCONF here, no session handling, no
// capability exchange, no XML-RPC. YANG is our schema (RFC 6241's layer 4); the
// transports remain SSH and SNMP. See inv/netconf-yang.md.
//
// # Why the rules live here
//
// The rule that a `weak` link may not be committed unreviewed was previously enforced
// in cmd/scan/connections.go -- that is, in the CLI. The service did not know it and
// the HTTP API did not apply it, so the same database could be corrupted through a
// different door. Moving it here means one rule, enforced once, for every caller.
//
// The same rule also exists as a `must` statement in internal/yang/modules/nsl-topology.yang,
// where it constrains any exported document. Validate is its Go twin: the schema guards
// what leaves the system, this guards what enters it.
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
	"strings"

	"nsl-graph/internal/topology"
)

// Origin is NMDA's per-item provenance (RFC 8342): why is this in the datastore?
type Origin string

const (
	// OriginIntended: it is here because an operator specified it.
	OriginIntended Origin = "intended"
	// OriginLearned: it is here because a scan discovered it.
	OriginLearned Origin = "learned"
)

// StagedEdge is one proposed link in a candidate, together with the operator's verdict
// on it.
//
// Reviewed is the operator's act of acceptance -- the thing that makes a weak link
// committable. In the CLI it is the user ticking the edge in the selection list; over
// the HTTP API it is an explicit flag. Either way it is a decision, not a default.
type StagedEdge struct {
	Edge     topology.ConnectionEdge
	Reviewed bool
}

// Candidate is a staged set of changes: proposed, not yet applied. Nothing here has
// touched the database.
type Candidate struct {
	Edges []StagedEdge
}

// Stage adds an edge to the candidate.
func (c *Candidate) Stage(edge topology.ConnectionEdge, reviewed bool) {
	c.Edges = append(c.Edges, StagedEdge{Edge: edge, Reviewed: reviewed})
}

// Violation is a rule the candidate breaks. It names the rule so the message can be
// traced back to the schema statement that mirrors it.
type Violation struct {
	Rule   string // the rule's name, matching the YANG constraint where one exists
	Target string // which edge, in "from <-> to" form
	Detail string // why it fails, in words an operator can act on
}

func (v Violation) Error() string {
	return fmt.Sprintf("%s: %s (%s)", v.Target, v.Detail, v.Rule)
}

// Rule names. Each corresponds to a constraint in nsl-topology.yang where one exists,
// so a rejection here and a rejection from yanglint say the same thing.
const (
	// RuleWeakMustBeReviewed mirrors the `must` on nsl-topology:confidence.
	RuleWeakMustBeReviewed = "weak-link-must-be-reviewed"
	// RuleEndpointsResolved has no YANG twin: RFC 8345 sets require-instance false on
	// link endpoints, so the schema deliberately permits a link to a node it does not
	// hold. We do not, because our topology is a closed specification rather than a
	// partial view of someone else's.
	RuleEndpointsResolved = "endpoints-must-resolve"
)

// Validate checks a candidate against the commit rules and returns every violation, not
// merely the first: an operator fixing a scan wants the whole list, not a game of
// whack-a-mole.
//
// A candidate with any violation must not be committed. Commit enforces that.
func Validate(c Candidate) []Violation {
	var violations []Violation

	for _, se := range c.Edges {
		target := fmt.Sprintf("%s <-> %s", se.Edge.FromLabel, se.Edge.ToLabel)

		// The forwarding database proves only that a MAC was seen on a port. It does
		// not prove adjacency: the device owning that MAC may be several hops away,
		// behind an unmanaged switch that bridges the traffic transparently. Committing
		// such a link unreviewed would invent topology that does not exist.
		if se.Edge.Confidence == topology.ConfidenceWeak && !se.Reviewed {
			violations = append(violations, Violation{
				Rule:   RuleWeakMustBeReviewed,
				Target: target,
				Detail: "the only evidence is the forwarding database (weak), so the link may " +
					"not exist -- the MAC may be behind an intermediate device. An operator " +
					"must review it before it can be committed",
			})
		}

		// An edge whose far end could not be matched to a device port names something
		// we do not model -- typically an unmanaged switch. It is worth reporting, and
		// it is not worth persisting.
		if !se.Edge.RemoteResolved || se.Edge.FromDevicePortID == "" || se.Edge.ToDevicePortID == "" {
			violations = append(violations, Violation{
				Rule:   RuleEndpointsResolved,
				Target: target,
				Detail: "an endpoint does not resolve to a known device port; the far end is " +
					"probably a device that is not modelled (an unmanaged switch, say)",
			})
		}
	}

	return violations
}

// Committable partitions a candidate into the edges that may be committed and the
// violations that stopped the rest.
//
// This is the shape callers actually want: a scan of a real network always turns up
// some edges that cannot be committed, and refusing the whole scan because of them
// would make discovery useless. What must NOT happen is committing them silently.
func Committable(c Candidate) (ok []StagedEdge, rejected []Violation) {
	blocked := map[string]bool{}
	for _, v := range Validate(c) {
		rejected = append(rejected, v)
		blocked[v.Target] = true
	}

	for _, se := range c.Edges {
		target := fmt.Sprintf("%s <-> %s", se.Edge.FromLabel, se.Edge.ToLabel)
		if !blocked[target] {
			ok = append(ok, se)
		}
	}
	return ok, rejected
}

// ViolationsError renders violations as a single error, or nil if there are none.
func ViolationsError(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}
	lines := make([]string, 0, len(violations))
	for _, v := range violations {
		lines = append(lines, "  "+v.Error())
	}
	return fmt.Errorf("%d change(s) rejected by the commit rules:\n%s",
		len(violations), strings.Join(lines, "\n"))
}
