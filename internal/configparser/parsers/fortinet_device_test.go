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

// fortinet_device_test.go: the FortiGate parser against a real FortiOS box.
//
// The Fortinet parser predated any FortiGate to test it on, and the three things
// it got wrong were all things no invented fixture would have contained: a
// terminal prompt glued to the first line of every reply, a pager marker spliced
// into the middle of a config block, and — the one that mattered — an interface
// whose address the configuration simply does not state.
//
// The fixtures here are captured verbatim from a FortiGate-30D running FortiOS
// 6.0.12, prompt echo and all.
package parsers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

func fgt30d(t *testing.T) (*configparser.ConfigData, map[string]configparser.ConfigInterface) {
	t.Helper()
	cd, err := parsers.NewFortinetParser().ParseConfig(
		fixture(t, "fortios-6.0-fgt30d.txt"), s.SNMPDevice{IP: "10.0.50.50"})
	require.NoError(t, err)

	byName := make(map[string]configparser.ConfigInterface, len(cd.Interfaces))
	for _, i := range cd.Interfaces {
		byName[i.Name] = i
	}
	return cd, byName
}

// The bug this fixture exists for.
//
// `show system interface` says only `set mode dhcp` for the WAN port — the
// address is nowhere in the configuration, because the device was not configured
// with one. The old parser therefore recorded a firewall with no reachable
// address while it was answering, at that moment, on 10.0.50.50. A branch
// firewall with a DHCP or PPPoE WAN is the normal case, not an edge case, so
// this was most FortiGates.
func TestFortinetParser_RecoversTheDHCPAddressFromRuntimeState(t *testing.T) {
	_, byName := fgt30d(t)

	assert.Equal(t, []string{"10.0.50.50/24"}, byName["wan"].IPAddresses,
		"the address the device was scanned on must survive the scan")

	// The netmask arrives dotted ("ip: 10.0.50.50 255.255.255.0"), not as a
	// prefix length, so this also pins the conversion.
	assert.Equal(t, []string{"192.168.1.99/24"}, byName["lan"].IPAddresses,
		"a statically configured address still comes from the config")

	// modem is PPPoE and down, with the placeholder 0.0.0.0 0.0.0.0 in runtime
	// state. A parser that took that literally would put every idle WAN port of
	// every FortiGate on the same address.
	assert.Empty(t, byName["modem"].IPAddresses)
}

// FortiOS emits `config router ospf` and `config router bgp` blocks containing
// nothing but empty `redistribute` placeholders even when neither protocol is
// configured — as on this device, which runs neither. Treating the block's
// presence as evidence would report OSPF and BGP on every FortiGate ever
// scanned, which is worse than reporting nothing: it is confidently wrong.
func TestFortinetParser_EmptyRouterBlocksAreNotAControlPlane(t *testing.T) {
	cd, _ := fgt30d(t)

	assert.Empty(t, cd.RoutingProtocols,
		"empty redistribute skeletons are not a configured protocol")
	assert.Equal(t, []string{configparser.RoutingProtoStatic}, cd.ControlPlane(),
		"one static default route is the whole control plane here")
}

func TestFortinetParser_StaticDefaultRoute(t *testing.T) {
	cd, _ := fgt30d(t)

	require.Len(t, cd.Routes, 1)
	r := cd.Routes[0]
	// The route states no destination at all; on FortiOS that means default.
	assert.Equal(t, "0.0.0.0/0", r.Network, "a route with no `set dst` is the default route")
	assert.Equal(t, "10.0.50.1", r.Gateway)
	assert.Equal(t, "wan", r.Interface)
	assert.Equal(t, configparser.RoutingProtoStatic, r.Protocol)
}

func TestFortinetParser_InterfacesTypesAndIdentity(t *testing.T) {
	cd, byName := fgt30d(t)

	assert.Equal(t, "FGT30D3X15012871", cd.Hostname)
	require.Len(t, cd.Interfaces, 4)

	assert.Equal(t, "physical", byName["wan"].Type)
	assert.Equal(t, "physical", byName["modem"].Type)
	assert.Equal(t, "tunnel", byName["ssl.root"].Type)
	// `hard-switch` is a bridge over the built-in switch ports, not a port.
	assert.Equal(t, "bridge", byName["lan"].Type)

	// `set alias "SSL VPN interface"` is quoted and contains spaces; a splitter
	// that broke on whitespace would keep only "SSL".
	assert.Equal(t, "SSL VPN interface", byName["ssl.root"].Description)

	// MACs come from a separate `diagnose` command per interface.
	assert.Equal(t, "90:6c:ac:0a:c1:79", byName["wan"].MACAddress)

	// A port that is down operationally is reported disabled.
	assert.False(t, byName["modem"].Enabled, "modem is down")
	assert.True(t, byName["wan"].Enabled)
}

// FortiOS splices `--More--` into the middle of a line when the console pager is
// at its default setting, e.g. "--More--                  set snmp-index 4".
// No content is lost, so the parser strips the marker and keeps the line —
// which is the right fix, because the alternative is reconfiguring somebody's
// firewall in order to read it.
func TestFortinetParser_SurvivesPagerMarkers(t *testing.T) {
	cd, err := parsers.NewFortinetParser().ParseConfig(
		fixture(t, "fortios-paginated-interface.txt"), s.SNMPDevice{IP: "10.0.50.50"})
	require.NoError(t, err)

	require.Len(t, cd.Interfaces, 4, "pagination must not lose or split an interface")

	byName := make(map[string]configparser.ConfigInterface, len(cd.Interfaces))
	for _, i := range cd.Interfaces {
		byName[i.Name] = i
	}
	// `lan` is the entry the pager marker lands inside.
	assert.Contains(t, byName, "lan")
	assert.Equal(t, []string{"192.168.1.99/24"}, byName["lan"].IPAddresses,
		"the interrupted entry keeps its address")
	assert.Equal(t, "bridge", byName["lan"].Type)
}

// Every reply carries the device prompt on its first line
// ("FGT30D3X15012871 # config system interface") and a bare prompt at the end.
// Left in place, the prompt turns the opening line of each block into an
// unparseable one, which on FortiOS is the line that names the section.
func TestFortinetParser_StripsPromptEcho(t *testing.T) {
	cd, byName := fgt30d(t)

	assert.NotContains(t, cd.Hostname, "#")
	for name := range byName {
		assert.NotContains(t, name, "#", "interface name %q kept prompt text", name)
		assert.NotContains(t, name, " ", "interface name %q kept prompt text", name)
	}
}

func TestFortinetParser_RealDeviceValidates(t *testing.T) {
	cd, _ := fgt30d(t)
	assert.Empty(t, parsers.NewFortinetParser().ValidateConfig(cd))
}
