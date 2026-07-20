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
// The fixtures here are captured verbatim from two real boxes, prompt echo and
// all: a FortiGate-30D on FortiOS 6.0.12, and a FortiGate-60C on 5.2.15. Two
// devices rather than one on purpose — a parser built against a single box
// cannot tell which of its behaviours are FortiOS and which are that box.
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

// -----------------------------------------------------------------------------
// A second, independent FortiGate — FortiOS 5.2 on different hardware
// -----------------------------------------------------------------------------
//
// The parser was rebuilt against exactly one device, which is one device's worth
// of evidence: every shape it handles could have been an accident of that box's
// model and firmware. This fixture is a FortiGate-60C on FortiOS 5.2.15 — a major
// version older, different hardware, different port names — and it is here to
// catch the assumptions the 30D let pass.
//
// It differs in ways that matter: 5.2 stamps `set vdom "root"` on every interface,
// indents the router blocks twice as deep, and reports the built-in switch port as
// `type physical` where the 30D called its equivalent `hard-switch`.

func fgt60c(t *testing.T) (*configparser.ConfigData, map[string]configparser.ConfigInterface) {
	t.Helper()
	cd, err := parsers.NewFortinetParser().ParseConfig(
		fixture(t, "fortios-5.2-fgt60c.txt"), s.SNMPDevice{IP: "10.0.50.105"})
	require.NoError(t, err)

	byName := make(map[string]configparser.ConfigInterface, len(cd.Interfaces))
	for _, i := range cd.Interfaces {
		byName[i.Name] = i
	}
	return cd, byName
}

func TestFortinetParser_FortiOS52_InterfacesAndIdentity(t *testing.T) {
	cd, byName := fgt60c(t)

	assert.Equal(t, "FGT60C3G13029672", cd.Hostname)
	require.Len(t, cd.Interfaces, 6, "dmz, wan2, wan1, modem, ssl.root, internal")

	for _, n := range []string{"dmz", "wan2", "wan1", "modem", "ssl.root", "internal"} {
		assert.Contains(t, byName, n)
	}

	// 5.2 stamps `set vdom "root"` on every interface. It is an unmodelled field,
	// and an unmodelled field must be ignored rather than swallow the ones after it.
	assert.Equal(t, "physical", byName["wan1"].Type)
	assert.Equal(t, "tunnel", byName["ssl.root"].Type)
	assert.Equal(t, "SSL VPN interface", byName["ssl.root"].Description)

	assert.Equal(t, "08:5b:0e:3f:49:79", byName["wan1"].MACAddress)
	assert.Equal(t, "08:5b:0e:3f:49:78", byName["internal"].MACAddress)
}

// The headline bug, re-checked on a device that has never been used to develop
// the fix: wan1 is `set mode dhcp` and states no address in the configuration,
// yet the box answers on 10.0.50.105 — which is how it was scanned.
func TestFortinetParser_FortiOS52_RecoversTheDHCPAddress(t *testing.T) {
	_, byName := fgt60c(t)

	assert.Equal(t, []string{"10.0.50.105/24"}, byName["wan1"].IPAddresses,
		"the DHCP address must come from runtime state, as on 6.0")
	assert.Equal(t, []string{"192.168.1.99/24"}, byName["internal"].IPAddresses,
		"a statically configured address still comes from the config")
	assert.Equal(t, []string{"10.10.10.1/24"}, byName["dmz"].IPAddresses)

	// wan2 and modem are down with no address; the 0.0.0.0 placeholder in runtime
	// state must not be taken literally.
	assert.Empty(t, byName["wan2"].IPAddresses)
	assert.Empty(t, byName["modem"].IPAddresses)
	assert.False(t, byName["wan2"].Enabled)
	assert.True(t, byName["wan1"].Enabled)
}

// 5.2 emits the same empty `config router ospf` placeholder as 6.0, but indented
// twice as deep. A parser keying on indentation rather than block structure would
// read one version correctly and the other as a configured protocol.
func TestFortinetParser_FortiOS52_EmptySkeletonAcrossVersions(t *testing.T) {
	cd, _ := fgt60c(t)

	assert.Empty(t, cd.RoutingProtocols,
		"neither OSPF nor BGP is configured on this device, at either indentation")
}

// This box's default route is installed by the DHCP client, not configured: the
// RIB holds `S* 0.0.0.0/0 via 10.0.50.1, wan1` while `show router static` is
// empty. Reporting no routes is therefore correct — the configuration genuinely
// contains none — and the test pins that rather than the more tempting assertion
// that a reachable device must have a default route.
func TestFortinetParser_FortiOS52_NoConfiguredRoutes(t *testing.T) {
	cd, _ := fgt60c(t)

	assert.Empty(t, cd.Routes, "show router static is empty; the default route is DHCP-installed")
	assert.Nil(t, cd.ControlPlane(),
		"no configured routes and no dynamic protocol is an honest empty answer")
}

func TestFortinetParser_FortiOS52_Validates(t *testing.T) {
	cd, _ := fgt60c(t)
	assert.Empty(t, parsers.NewFortinetParser().ValidateConfig(cd))
}
