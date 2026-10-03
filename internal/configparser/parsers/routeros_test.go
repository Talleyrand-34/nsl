// SPDX-License-Identifier: AGPL-3.0-or-later
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

// routeros_test.go: the RouterOS parser against real hardware.
//
// The other RouterOS fixtures come from CHR, the virtual edition, which has five
// identical synthetic ethernet ports and nothing else. A CCR1009-7G-1C-1S+ is a
// different proposition: mixed port families (combo1, sfp-sfpplus1), a bridge
// with eight enslaved ports, multi-letter status flags, and a `defconf` comment
// left by the factory configuration. Every one of those is a shape the virtual
// fixtures cannot produce, and each has its own way of going wrong.
package parsers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

// ccr1009 parses the hardware fixture once for the tests below.
func ccr1009(t *testing.T) (*configparser.ConfigData, map[string]configparser.ConfigInterface) {
	t.Helper()
	cd, err := parsers.NewRouterOSParser().ParseConfig(
		fixture(t, "routeros-ccr1009.txt"), s.SNMPDevice{IP: "10.0.50.51"})
	require.NoError(t, err)

	byName := make(map[string]configparser.ConfigInterface, len(cd.Interfaces))
	for _, i := range cd.Interfaces {
		byName[i.Name] = i
	}
	return cd, byName
}

func TestRouterOSParser_HardwarePortFamilies(t *testing.T) {
	cd, byName := ccr1009(t)

	assert.Equal(t, "MikroTik", cd.Hostname)
	require.Len(t, cd.Interfaces, 10, "seven ether, one combo, one SFP+, one bridge")

	// Port names that are not etherN at all. A parser keying off the "ether"
	// prefix — a tempting shortcut on the CHR fixtures, where every port is
	// etherN — loses both of these outright.
	assert.Contains(t, byName, "combo1")
	assert.Contains(t, byName, "sfp-sfpplus1")
	assert.Equal(t, "08:55:31:31:32:99", byName["combo1"].MACAddress)
	assert.Equal(t, "08:55:31:31:32:98", byName["sfp-sfpplus1"].MACAddress)

	// Every physical port carries its own MAC, consecutively assigned.
	for _, name := range []string{"ether1", "ether2", "ether7"} {
		assert.NotEmpty(t, byName[name].MACAddress, "%s must carry a MAC", name)
	}
}

// The management address is the point of the scan: it is how the device was
// reached, and a parser that loses it produces a device nothing can talk to.
func TestRouterOSParser_HardwareAddresses(t *testing.T) {
	_, byName := ccr1009(t)

	assert.Equal(t, []string{"10.0.50.51/24"}, byName["ether1"].IPAddresses,
		"the address the box was scanned on")
	assert.Equal(t, []string{"192.168.88.1/24"}, byName["bridge1"].IPAddresses,
		"the factory default LAN, which lives on the bridge and not on a port")

	// An enslaved port has no address of its own; the bridge holds it. Attributing
	// the bridge's address to its members would invent eight hosts on one subnet.
	for _, name := range []string{"ether2", "ether3", "sfp-sfpplus1", "combo1"} {
		assert.Empty(t, byName[name].IPAddresses, "%s is a bridge member", name)
	}
}

// Bridge membership is what makes the topology readable: without it the eight
// enslaved ports look like eight independent routed interfaces.
func TestRouterOSParser_BridgeMembershipResolvesToParents(t *testing.T) {
	_, byName := ccr1009(t)

	assert.Equal(t, "bridge", byName["bridge1"].Type)
	assert.Empty(t, byName["bridge1"].Parent, "the bridge is nobody's member")

	members := []string{"ether2", "ether3", "ether4", "ether5", "ether6", "ether7", "sfp-sfpplus1", "combo1"}
	for _, name := range members {
		assert.Equal(t, "bridge1", byName[name].Parent, "%s should be enslaved to bridge1", name)
	}

	// ether1 is the routed uplink and deliberately outside the bridge. It is the
	// one port whose membership differs, so it is the one worth asserting.
	assert.Empty(t, byName["ether1"].Parent, "ether1 is routed, not bridged")
}

// RouterOS prints status as flag letters between the index and the first
// key=value, and a port may carry several at once — this fixture has bare "S"
// (slave), "R" (running) and the combined "RS". A tokenizer that assumed a single
// flag letter, or that started reading attributes at a fixed offset, mangles the
// first attribute of every multi-flag record.
func TestRouterOSParser_MultiLetterFlagsDoNotEatAttributes(t *testing.T) {
	_, byName := ccr1009(t)

	// ether2 is "RS" — both running and enslaved.
	assert.Equal(t, "08:55:31:31:32:9B", byName["ether2"].MACAddress,
		"a two-letter flag must not shift attribute parsing")
	// combo1 is " S" — a leading space where a flag letter would be.
	assert.Equal(t, "08:55:31:31:32:99", byName["combo1"].MACAddress)
	// ether1 is "R " — a flag then padding.
	assert.Equal(t, []string{"10.0.50.51/24"}, byName["ether1"].IPAddresses)

	// bridge1 reports `mtu=auto`, which is not a number. Falling back to
	// actual-mtu is what keeps it from being recorded as zero.
	assert.Equal(t, 1500, byName["bridge1"].MTU)
}

// This box has no dynamic routing configured, and must say so rather than
// inventing an instance from an empty menu.
func TestRouterOSParser_HardwareHasNoDynamicRouting(t *testing.T) {
	cd, _ := ccr1009(t)

	assert.Empty(t, cd.RoutingProtocols,
		"no OSPF instance and no BGP connection are configured on this device")
	assert.Empty(t, cd.VLANs, "no VLAN interfaces are configured")

	// Whatever routes exist are static or connected; none was learned.
	for _, r := range cd.Routes {
		assert.NotEqual(t, configparser.RoutingProtoOSPF, r.Protocol)
		assert.NotEqual(t, configparser.RoutingProtoBGP, r.Protocol)
	}
}

func TestRouterOSParser_HardwareValidates(t *testing.T) {
	cd, _ := ccr1009(t)
	assert.Empty(t, parsers.NewRouterOSParser().ValidateConfig(cd))
}
