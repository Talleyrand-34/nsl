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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	s "nsl-graph/internal/scanner"
	"nsl-graph/internal/yang/canon"
)

// specWithIPs is the lab specification, with sw1 given a management address so the
// scan can be resolved against it by IP.
func specWithIPs() Source {
	src := lab()
	src.Devices[0].Ips = []string{"10.0.2.20"}
	return src
}

// scanned builds a scanned device with one physical ethernet port.
func scanned(ip, sysName, port string, vlans ...s.VLANMembership) s.SNMPDevice {
	return s.SNMPDevice{
		IP:        ip,
		SysName:   sysName,
		Reachable: true,
		Interfaces: []s.DeviceInterface{{
			Index:       1,
			IfType:      6, // ethernetCsmacd
			Name:        port,
			MAC:         "00:1B:1B:00:00:01",
			AdminStatus: 1,
			OperStatus:  1,
			VLANs:       vlans,
		}},
	}
}

func observedNodes(t *testing.T, root *canon.Root) []canon.Node {
	t.Helper()
	require.NotNil(t, root.Networks)
	require.Len(t, root.Networks.Network, 1)
	return root.Networks.Network[0].Nodes
}

// The identity problem, and the whole reason FromScan takes the specification.
//
// The spec keys a device by UUID; a scan knows only an IP. Unless the same physical
// device carries the SAME node-id in both trees, a diff between them is meaningless --
// every device would appear once as missing and once as unexpected.
func TestScannedDeviceAdoptsItsSpecificationIdentity(t *testing.T) {
	root, warnings := FromScan([]s.SNMPDevice{scanned("10.0.2.20", "sw1", "P1")}, specWithIPs())

	assert.Empty(t, warnings)

	nodes := observedNodes(t, root)
	require.Len(t, nodes, 1)
	assert.Equal(t, "urn:nsl:device:d1", nodes[0].NodeID,
		"the scanned device must adopt the node-id the specification gave it, not a new one")
}

// A device may have been re-addressed since the specification was written, so the name
// is the fallback.
func TestScannedDeviceResolvesByNameWhenTheIPHasChanged(t *testing.T) {
	root, warnings := FromScan([]s.SNMPDevice{scanned("10.0.2.99", "sw1", "P1")}, specWithIPs())

	assert.Empty(t, warnings)

	nodes := observedNodes(t, root)
	require.Len(t, nodes, 1)
	assert.Equal(t, "urn:nsl:device:d1", nodes[0].NodeID, "resolved by sysName despite the new IP")
}

// Something on the network that nobody wrote down. It must surface as unexpected rather
// than be quietly dropped -- a rogue access point is exactly what an audit is for.
func TestUnknownDeviceIsMintedAnIdentityAndReported(t *testing.T) {
	root, warnings := FromScan([]s.SNMPDevice{scanned("10.0.2.99", "rogue-ap", "P1")}, specWithIPs())

	nodes := observedNodes(t, root)
	require.Len(t, nodes, 1)
	assert.Equal(t, "urn:nsl:device:unknown:10.0.2.99", nodes[0].NodeID)
	assert.Equal(t, "rogue-ap", nodes[0].L2.Name)

	assert.True(t, hasWarning(warnings, "matches nothing in the specification"), "got: %v", warnings)
}

// A port the specification knows adopts its tp-id; one it does not know is minted an id
// so it shows up as unexpected.
func TestPortsAdoptSpecificationIdentitiesWhereTheyExist(t *testing.T) {
	dev := scanned("10.0.2.20", "sw1", "P1")
	dev.Interfaces = append(dev.Interfaces, s.DeviceInterface{
		Index: 2, IfType: 6, Name: "P9", MAC: "00:1B:1B:00:00:09", AdminStatus: 1, OperStatus: 1,
	})

	root, _ := FromScan([]s.SNMPDevice{dev}, specWithIPs())

	tps := observedNodes(t, root)[0].TerminationPoints
	require.Len(t, tps, 2)

	byName := map[string]canon.TerminationPoint{}
	for _, tp := range tps {
		byName[tp.L2.InterfaceName] = tp
	}

	assert.Equal(t, "urn:nsl:tp:dp1", byName["P1"].TpID, "P1 exists in the spec and keeps its id")
	assert.Contains(t, byName["P9"].TpID, "unknown:", "P9 is not in the spec and is minted a new id")
}

// The observed tree applies the same typing rules as the specification side. A device
// that reports VLAN 4999 is reporting nonsense; letting it through would show it as an
// "unexpected VLAN" and send an operator hunting for a VLAN that cannot exist.
func TestObservedVlansAreTypedTheSameWayAsIntendedOnes(t *testing.T) {
	dev := scanned("10.0.2.20", "sw1", "P1",
		s.VLANMembership{VLANNumber: "10", Tagged: true},
		s.VLANMembership{VLANNumber: "4999", Tagged: true},
		s.VLANMembership{VLANNumber: "eth0", Tagged: false},
	)

	root, warnings := FromScan([]s.SNMPDevice{dev}, specWithIPs())

	vlans := observedNodes(t, root)[0].TerminationPoints[0].VlanMembership
	assert.Equal(t, []canon.VlanMembership{{VlanID: 10, Tagged: true}}, vlans)

	assert.True(t, hasWarning(warnings, "out of range"), "got: %v", warnings)
	assert.True(t, hasWarning(warnings, "is not a number"), "got: %v", warnings)
}

// A bridge or VLAN sub-interface is a logical construct that the specification models as
// a DeviceInterface, not as a port. Comparing it against the port template would report
// differences that are not differences.
func TestLogicalInterfacesAreNotTreatedAsPorts(t *testing.T) {
	dev := scanned("10.0.2.20", "sw1", "P1")
	dev.Interfaces = append(dev.Interfaces,
		s.DeviceInterface{Index: 2, IfType: 53, Name: "br-lan", IsBridge: true, AdminStatus: 1, OperStatus: 1},
		s.DeviceInterface{Index: 3, IfType: 53, Name: "eth0.10", Parent: "eth0", AdminStatus: 1, OperStatus: 1},
	)

	root, _ := FromScan([]s.SNMPDevice{dev}, specWithIPs())

	tps := observedNodes(t, root)[0].TerminationPoints
	require.Len(t, tps, 1, "only the physical port is a termination point")
	assert.Equal(t, "P1", tps[0].L2.InterfaceName)
}

// A device scan sees interfaces, not cables. Adjacency comes from correlating LLDP/CDP
// and forwarding tables across several hosts, which is what `scan connections` does.
// Emitting no links means the diff reports no link changes -- which is honest. Emitting
// an empty link set as if it were authoritative would report every cable in the
// specification as missing.
func TestScanEmitsNoLinks(t *testing.T) {
	root, _ := FromScan([]s.SNMPDevice{scanned("10.0.2.20", "sw1", "P1")}, specWithIPs())

	assert.Empty(t, root.Networks.Network[0].Links,
		"a device scan cannot see cabling and must not pretend otherwise")
}

// Loopback and unspecified addresses identify nothing and must not be used to resolve a
// device -- they would match every device that has one.
func TestUselessAddressesAreIgnoredForResolution(t *testing.T) {
	for _, ip := range []string{"127.0.0.1", "0.0.0.0", ""} {
		t.Run(ip, func(t *testing.T) {
			assert.Empty(t, usableAddr(ip))
		})
	}
	assert.Equal(t, "10.0.2.20", usableAddr("10.0.2.20/24"), "a CIDR suffix is stripped")
}

func TestEmptyScanProducesAnEmptyObservedTree(t *testing.T) {
	root, warnings := FromScan(nil, specWithIPs())

	assert.Empty(t, warnings)
	assert.Empty(t, observedNodes(t, root))
	assert.Equal(t, canon.NetworkID, root.Networks.Network[0].NetworkID)
}
