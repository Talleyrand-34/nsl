// SPDX-License-Identifier: AGPL-3.0-or-later
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
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nsl-graph/internal/repository/application"
	e "nsl-graph/internal/repository/entities"
	"nsl-graph/internal/yang/canon"
)

// TestExportYANG covers the wiring that internal/yang/mapping's pure tests
// deliberately do not: that ExportYANG composes the right service getters, and that a
// specification built through the ordinary Add* API comes back out as a well-formed
// RFC 7951 document.
//
// It reuses setupTestRepository (application_test.go) -- a CloverDB in t.TempDir().
func TestExportYANG(t *testing.T) {
	repo := setupTestRepository(t)
	service := application.NewNetService(repo)

	// Catalogues.
	require.NoError(t, service.AddBrand("Siemens"))
	require.NoError(t, service.AddModelType("Switch"))
	// EnsureOsType, not AddOsType: NewNetService seeds the catalogue from the
	// registered config parsers, so whether "openwrt" already exists depends on
	// which packages this test binary links.
	require.NoError(t, service.EnsureOsType("openwrt"))
	require.NoError(t, service.AddOwner("IT"))
	require.NoError(t, service.AddZoneType("Security"))
	require.NoError(t, service.AddConnectionType("ethernet"))
	require.NoError(t, service.AddZone("DMZ", "", "", "IT", "Security"))

	zones, err := service.GetZones()
	require.NoError(t, err)
	require.Len(t, zones, 1)
	zoneID := zones[0].ID

	// A model and its port template.
	require.NoError(t, service.AddModel("XC206", "Siemens", "Switch", "openwrt"))
	require.NoError(t, service.AddModelPort("P1", "0", "0", "XC206", false, "ethernet", ""))
	require.NoError(t, service.AddModelPort("P2", "1", "0", "XC206", false, "ethernet", ""))

	// CloverDB is a document store and does not promise insertion order, so index by
	// name rather than assuming modelPorts[0] is P1.
	modelPorts, err := service.GetModelPorts()
	require.NoError(t, err)
	require.Len(t, modelPorts, 2)

	modelPortID := map[string]string{}
	for _, mp := range modelPorts {
		modelPortID[mp.Name] = mp.ID
	}
	require.Contains(t, modelPortID, "P1")

	// Two devices instantiated from it.
	require.NoError(t, service.AddDevice("sw1", "XC206", zoneID, "DMZ", "IT", false, false))
	require.NoError(t, service.AddDevice("sw2", "XC206", zoneID, "DMZ", "IT", false, false))

	devices, err := service.GetDevices()
	require.NoError(t, err)
	require.Len(t, devices, 2)

	byLabel := map[string]e.Device{}
	for _, d := range devices {
		byLabel[d.Label] = d
	}

	// One port on each, sw1's carrying VLANs.
	dp1, err := service.AddDevicePort(byLabel["sw1"].ID, modelPortID["P1"], "00:1B:1B:00:00:01",
		[]e.PortVlanConfig{{VlanNumber: "10", Tagged: false}, {VlanNumber: "20", Tagged: true}})
	require.NoError(t, err)
	dp2, err := service.AddDevicePort(byLabel["sw2"].ID, modelPortID["P1"], "00:1B:1B:00:00:02", nil)
	require.NoError(t, err)

	require.NoError(t, service.AddConnection(dp1, dp2, "ethernet", "ssh-lldp@sw1:P1"))

	// --- the thing under test -------------------------------------------------

	out, warnings, err := service.ExportYANG()
	require.NoError(t, err)
	assert.Empty(t, warnings, "a specification built through the normal API must export cleanly")

	var root canon.Root
	require.NoError(t, json.Unmarshal(out, &root), "the export must be well-formed JSON")

	// Inventory came through.
	require.NotNil(t, root.Inventory)
	assert.Len(t, root.Inventory.Brands, 1)
	assert.Len(t, root.Inventory.Zones, 1)
	require.Len(t, root.Inventory.Models, 1)
	assert.Equal(t, "XC206", root.Inventory.Models[0].Name)
	assert.Len(t, root.Inventory.Models[0].ModelPorts, 2, "the port template must hang off the model")

	// Topology: one network, devices as nodes, ports as termination points.
	require.NotNil(t, root.Networks)
	require.Len(t, root.Networks.Network, 1)
	net := root.Networks.Network[0]
	assert.Equal(t, canon.NetworkID, net.NetworkID)
	require.Len(t, net.Nodes, 2)

	nodesByName := map[string]canon.Node{}
	for _, n := range net.Nodes {
		require.NotNil(t, n.L2)
		nodesByName[n.L2.Name] = n
	}
	require.Contains(t, nodesByName, "sw1")

	sw1 := nodesByName["sw1"]
	assert.Equal(t, zoneID, sw1.Zone)
	assert.Equal(t, "IT", sw1.Owner)
	require.Len(t, sw1.TerminationPoints, 1)

	tp := sw1.TerminationPoints[0]
	require.NotNil(t, tp.L2)
	assert.Equal(t, "P1", tp.L2.InterfaceName)
	assert.Equal(t, "00:1b:1b:00:00:01", tp.L2.MacAddress)
	assert.Equal(t, []canon.VlanMembership{
		{VlanID: 10, Tagged: false},
		{VlanID: 20, Tagged: true},
	}, tp.VlanMembership)

	// One connection, two unidirectional RFC 8345 links.
	require.Len(t, net.Links, 2, "a connection must export as two links, one per direction")
	assert.Equal(t, net.Links[0].Source.SourceNode, net.Links[1].Destination.DestNode)
	assert.Equal(t, net.Links[0].Destination.DestNode, net.Links[1].Source.SourceNode)

	assert.Equal(t, []canon.DiscoveredVia{{Source: "ssh-lldp", ObservedOn: "sw1:P1"}},
		net.Links[0].DiscoveredVia)
}

// An empty database must still export a well-formed document rather than a null tree.
func TestExportYANG_EmptyDatabase(t *testing.T) {
	repo := setupTestRepository(t)
	service := application.NewNetService(repo)

	out, warnings, err := service.ExportYANG()
	require.NoError(t, err)
	assert.Empty(t, warnings)

	var root canon.Root
	require.NoError(t, json.Unmarshal(out, &root))
	require.NotNil(t, root.Networks)
	require.Len(t, root.Networks.Network, 1)
	assert.Empty(t, root.Networks.Network[0].Nodes)
}
