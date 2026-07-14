package application_test

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

	"nsl-graph/internal/datastore"
	"nsl-graph/internal/repository/application"
	"nsl-graph/internal/topology"
)

// twoWiredDevices builds a minimal lab in a temp CloverDB: two devices of the same
// model, one port each. It returns the service and the two device-port IDs.
func twoWiredDevices(t *testing.T) (application.NetServiceInt, string, string) {
	t.Helper()

	repo := setupTestRepository(t)
	service := application.NewNetService(repo)

	require.NoError(t, service.AddBrand("Siemens"))
	require.NoError(t, service.AddModelType("Switch"))
	require.NoError(t, service.AddOsType("openwrt"))
	require.NoError(t, service.AddConnectionType("ethernet"))
	require.NoError(t, service.AddModel("XC206", "Siemens", "Switch", "openwrt"))
	require.NoError(t, service.AddModelPort("P1", "0", "0", "XC206", false, "ethernet", ""))

	modelPorts, err := service.GetModelPorts()
	require.NoError(t, err)
	require.Len(t, modelPorts, 1)
	mpID := modelPorts[0].ID

	require.NoError(t, service.AddDevice("sw1", "XC206", "", "", "", false, false))
	require.NoError(t, service.AddDevice("sw2", "XC206", "", "", "", false, false))

	devices, err := service.GetDevices()
	require.NoError(t, err)
	require.Len(t, devices, 2)

	byLabel := map[string]string{}
	for _, d := range devices {
		byLabel[d.Label] = d.ID
	}

	dp1, err := service.AddDevicePort(byLabel["sw1"], mpID, "", nil)
	require.NoError(t, err)
	dp2, err := service.AddDevicePort(byLabel["sw2"], mpID, "", nil)
	require.NoError(t, err)

	return service, dp1, dp2
}

func stagedEdge(dp1, dp2, confidence string) topology.ConnectionEdge {
	return topology.ConnectionEdge{
		FromDevicePortID: dp1,
		ToDevicePortID:   dp2,
		FromLabel:        "sw1:P1",
		ToLabel:          "sw2:P1",
		Confidence:       confidence,
		Provenance:       []string{"snmp-fdb@sw1:P1"},
		RemoteResolved:   true,
	}
}

// The whole point of the datastore module.
//
// The rule "a weak link may not be committed unreviewed" used to live in
// cmd/scan/connections.go -- in the CLI. The service did not know it, so the HTTP API
// could write a weak link into the same database through a different door. This asserts
// the rule now holds at the service layer, where every caller meets it.
func TestServiceRefusesToCommitAnUnreviewedWeakLink(t *testing.T) {
	service, dp1, dp2 := twoWiredDevices(t)

	var c datastore.Candidate
	c.Stage(stagedEdge(dp1, dp2, topology.ConfidenceWeak), false)

	committed, violations, err := service.ImportConnectionEdgesChecked(c)
	require.NoError(t, err, "a rejected edge is a violation, not a failure of the call")

	assert.Zero(t, committed)
	require.Len(t, violations, 1)
	assert.Equal(t, datastore.RuleWeakMustBeReviewed, violations[0].Rule)

	// And nothing reached the database.
	connections, err := service.GetConnections()
	require.NoError(t, err)
	assert.Empty(t, connections, "a rejected edge must not be persisted")
}

// Reviewed, it commits -- and the evidence survives, which it previously did not.
func TestReviewedWeakLinkCommitsAndKeepsItsEvidence(t *testing.T) {
	service, dp1, dp2 := twoWiredDevices(t)

	var c datastore.Candidate
	c.Stage(stagedEdge(dp1, dp2, topology.ConfidenceWeak), true)

	committed, violations, err := service.ImportConnectionEdgesChecked(c)
	require.NoError(t, err)
	assert.Empty(t, violations)
	assert.Equal(t, 1, committed)

	connections, err := service.GetConnections()
	require.NoError(t, err)
	require.Len(t, connections, 1)

	// The confidence grade used to be computed during the scan and dropped on import,
	// leaving a weak link indistinguishable from a confirmed one once committed.
	assert.Equal(t, topology.ConfidenceWeak, connections[0].Confidence)
	assert.True(t, connections[0].Reviewed)
	assert.Equal(t, []string{"snmp-fdb@sw1:P1"}, connections[0].DiscoveredVia)
}

// A confirmed edge rests on bidirectional LLDP, so it needs no human in the loop.
func TestConfirmedLinkCommitsWithoutReview(t *testing.T) {
	service, dp1, dp2 := twoWiredDevices(t)

	var c datastore.Candidate
	c.Stage(stagedEdge(dp1, dp2, topology.ConfidenceConfirmed), false)

	committed, violations, err := service.ImportConnectionEdgesChecked(c)
	require.NoError(t, err)
	assert.Empty(t, violations)
	assert.Equal(t, 1, committed)

	connections, err := service.GetConnections()
	require.NoError(t, err)
	require.Len(t, connections, 1)
	assert.Equal(t, topology.ConfidenceConfirmed, connections[0].Confidence)
	assert.False(t, connections[0].Reviewed, "nobody reviewed it; it did not need reviewing")
}

// A hand-specified connection carries no confidence -- there is no evidence to grade,
// because it states intent rather than reporting an observation -- and is reviewed by
// definition.
func TestHandSpecifiedConnectionIsReviewedWithNoConfidence(t *testing.T) {
	service, dp1, dp2 := twoWiredDevices(t)

	require.NoError(t, service.AddConnection(dp1, dp2, "ethernet"))

	connections, err := service.GetConnections()
	require.NoError(t, err)
	require.Len(t, connections, 1)

	assert.Empty(t, connections[0].Confidence)
	assert.True(t, connections[0].Reviewed)
}
