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

// routing_test.go: the control plane, read back from real routers.
//
// Every fixture under testdata/ was captured from a running node of the lab's
// five-vendor dual ring, which is built three ways over one unchanged physical
// topology and one unchanged addressing plan: static, OSPF, and iBGP over an
// OSPF underlay. That is precisely what makes these tests worth having. The
// interface tables of the three variants are identical, so a parser that reads
// only addresses cannot tell them apart, and every assertion below that would
// still pass on the wrong variant is worthless. What is asserted here is the
// part that differs — router-ids, areas, peers, autonomous systems.
package parsers_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

func fixture(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("testdata", name))
	require.NoError(t, err, "reading fixture %s", name)
	return string(b)
}

// protoOfType returns the single instance of a protocol, failing if there is not
// exactly one — an assertion in its own right, since a parser that emits one BGP
// instance per neighbor is a common and quiet way to get this wrong.
func protoOfType(t *testing.T, cd *configparser.ConfigData, typ string) configparser.ConfigRoutingProtocol {
	t.Helper()
	got := cd.RoutingProtocolsOfType(typ)
	require.Len(t, got, 1, "expected exactly one %s instance", typ)
	return got[0]
}

func areaNetworks(p configparser.ConfigRoutingProtocol, id string) []string {
	for _, a := range p.Areas {
		if a.ID == id {
			return a.Networks
		}
	}
	return nil
}

func neighborAddrs(p configparser.ConfigRoutingProtocol) []string {
	out := make([]string, 0, len(p.Neighbors))
	for _, n := range p.Neighbors {
		out = append(out, n.Address)
	}
	return out
}

// -----------------------------------------------------------------------------
// FRR — shared by OpenWrt, Infix and OPNsense
// -----------------------------------------------------------------------------

// R4 of the OSPF ring: OpenWrt delegating to FRR, area 0 carrying both ring
// spans and the access LAN.
func TestParseFRRConfig_OSPFRingMember(t *testing.T) {
	got := parsers.ParseFRRConfig(fixture(t, "frr-ospf.conf"))

	require.Len(t, got, 1)
	assert.Equal(t, configparser.RoutingProtoOSPF, got[0].Type)
	assert.True(t, got[0].Enabled)
	assert.Equal(t, "4.4.4.4", got[0].RouterID)
	require.Len(t, got[0].Areas, 1)
	assert.Equal(t, "0", got[0].Areas[0].ID)
	assert.ElementsMatch(t, []string{
		"10.1.4.0/24",
		"10.10.34.0/30", "10.10.45.0/30",
		"10.20.34.0/30", "10.20.45.0/30",
	}, got[0].Areas[0].Networks)
}

// R4 of the iBGP ring runs both protocols at once: iBGP over loopbacks for the
// LANs, OSPF underneath to make those loopbacks reachable. Parsing this as one
// protocol, or letting the `exit-address-family` inside the BGP block terminate
// it early and swallow the OSPF stanza, are the two ways to get it wrong.
func TestParseFRRConfig_IBGPOverOSPFUnderlay(t *testing.T) {
	got := parsers.ParseFRRConfig(fixture(t, "frr-bgp-ospf.conf"))

	require.Len(t, got, 2, "iBGP overlay and OSPF underlay are two instances")

	var bgp, ospf configparser.ConfigRoutingProtocol
	for _, p := range got {
		switch p.Type {
		case configparser.RoutingProtoBGP:
			bgp = p
		case configparser.RoutingProtoOSPF:
			ospf = p
		}
	}

	assert.Equal(t, "65000", bgp.LocalAS)
	assert.Equal(t, "4.4.4.4", bgp.RouterID)
	assert.Equal(t, []string{"10.1.4.0/24"}, bgp.Networks)
	assert.ElementsMatch(t,
		[]string{"10.255.0.1", "10.255.0.2", "10.255.0.5"}, neighborAddrs(bgp))
	assert.True(t, bgp.HasIBGP(), "peers share the local AS, so the sessions are internal")
	for _, n := range bgp.Neighbors {
		assert.Equal(t, "65000", n.RemoteAS)
		assert.Equal(t, "lo", n.UpdateSource, "peering is loopback-to-loopback")
		assert.True(t, n.NextHopSelf, "set inside the address-family block")
	}

	assert.Equal(t, "4.4.4.4", ospf.RouterID)
	assert.ElementsMatch(t, []string{
		"10.10.23.0/30", "10.10.45.0/30",
		"10.20.23.0/30", "10.20.45.0/30",
		"10.255.0.4/32",
	}, areaNetworks(ospf, "0"), "the underlay must carry the loopback it advertises")
}

// -----------------------------------------------------------------------------
// VyOS
// -----------------------------------------------------------------------------

func TestVyOSParser_OSPFRingMember(t *testing.T) {
	cd, err := parsers.NewVyOSParser().ParseConfig(
		fixture(t, "vyos-ospf-config.boot"), s.SNMPDevice{IP: "10.255.0.1"})
	require.NoError(t, err)

	assert.Equal(t, "vyos", cd.Hostname)
	assert.Equal(t, []string{configparser.RoutingProtoOSPF}, cd.ControlPlane())

	ospf := protoOfType(t, cd, configparser.RoutingProtoOSPF)
	assert.Equal(t, "1.1.1.1", ospf.RouterID)
	assert.ElementsMatch(t, []string{
		"10.10.51.0/30", "10.20.51.0/30",
		"10.10.12.0/30", "10.20.12.0/30",
		"10.1.1.0/24", "10.255.0.1/32",
	}, areaNetworks(ospf, "0"))
}

// The BGP ring's R1 carries both protocols, and its addressing is byte-for-byte
// the OSPF ring's. ControlPlane() reporting both is the whole point: collapsing
// to a single "primary" protocol would erase the underlay.
func TestVyOSParser_IBGPRingMember(t *testing.T) {
	cd, err := parsers.NewVyOSParser().ParseConfig(
		fixture(t, "vyos-bgp-config.boot"), s.SNMPDevice{IP: "10.255.0.1"})
	require.NoError(t, err)

	assert.Equal(t,
		[]string{configparser.RoutingProtoBGP, configparser.RoutingProtoOSPF},
		cd.ControlPlane())

	bgp := protoOfType(t, cd, configparser.RoutingProtoBGP)
	assert.Equal(t, "65000", bgp.LocalAS, "read from `system-as`")
	assert.Equal(t, "1.1.1.1", bgp.RouterID)
	assert.Equal(t, []string{"10.1.1.0/24"}, bgp.Networks)
	assert.ElementsMatch(t,
		[]string{"10.255.0.2", "10.255.0.4", "10.255.0.5"}, neighborAddrs(bgp))
	assert.True(t, bgp.HasIBGP())
	for _, n := range bgp.Neighbors {
		assert.Equal(t, "lo", n.UpdateSource)
		// nexthop-self sits two levels down, inside address-family ipv4-unicast.
		assert.True(t, n.NextHopSelf, "neighbor %s", n.Address)
	}
}

func TestVyOSParser_InterfacesCarryAddressesAndHardwareIDs(t *testing.T) {
	cd, err := parsers.NewVyOSParser().ParseConfig(
		fixture(t, "vyos-ospf-config.boot"), s.SNMPDevice{IP: "10.255.0.1"})
	require.NoError(t, err)

	byName := map[string]configparser.ConfigInterface{}
	for _, i := range cd.Interfaces {
		byName[i.Name] = i
	}

	eth0 := byName["eth0"]
	assert.Equal(t, []string{"10.10.51.2/30"}, eth0.IPAddresses)
	assert.Equal(t, "0c:46:7e:79:00:00", eth0.MACAddress, "hw-id is the MAC")
	assert.True(t, eth0.Enabled)

	assert.Equal(t, []string{"10.1.1.254/24"}, byName["eth4"].IPAddresses, "the access LAN gateway")
	assert.Equal(t, []string{"10.255.0.1/32"}, byName["lo"].IPAddresses, "the iBGP peering address")
	assert.Empty(t, byName["eth9"].IPAddresses, "an uncabled ring port stays addressless")
}

// -----------------------------------------------------------------------------
// RouterOS
// -----------------------------------------------------------------------------

// RouterOS states OSPF membership in interface-templates that name an area by
// its *name*, while the area names its own id separately and points at an
// instance. Resolving template -> area name -> area id -> instance is the whole
// difficulty: matching a template's `area=bb` against an area-id of `0.0.0.0`
// finds nothing and silently drops every network.
func TestRouterOSParser_OSPFTemplatesResolveThroughAreaNames(t *testing.T) {
	cd, err := parsers.NewRouterOSParser().ParseConfig(
		fixture(t, "routeros-ospf.txt"), s.SNMPDevice{IP: "10.255.0.2"})
	require.NoError(t, err)

	assert.Equal(t, []string{configparser.RoutingProtoOSPF}, cd.ControlPlane())

	instances := cd.RoutingProtocolsOfType(configparser.RoutingProtoOSPF)
	require.Len(t, instances, 2, "the stock `default` instance coexists with the configured `os`")

	var os configparser.ConfigRoutingProtocol
	for _, p := range instances {
		if p.Instance == "os" {
			os = p
		}
	}
	require.Equal(t, "os", os.Instance, "the configured instance must be present")
	assert.Equal(t, "2.2.2.2", os.RouterID)
	assert.ElementsMatch(t, []string{
		"10.10.12.0/30", "10.20.12.0/30",
		"10.10.23.0/30", "10.20.23.0/30",
		"10.1.2.0/24",
	}, areaNetworks(os, "0.0.0.0"))
}

// Three /routing/bgp/connection records that share an AS and router-id are one
// BGP speaker with three peers, not three speakers.
func TestRouterOSParser_IBGPPeersGroupIntoOneInstance(t *testing.T) {
	cd, err := parsers.NewRouterOSParser().ParseConfig(
		fixture(t, "routeros-bgp.txt"), s.SNMPDevice{IP: "10.255.0.2"})
	require.NoError(t, err)

	bgp := protoOfType(t, cd, configparser.RoutingProtoBGP)
	assert.Equal(t, "65000", bgp.LocalAS)
	assert.Equal(t, "2.2.2.2", bgp.RouterID)
	assert.ElementsMatch(t,
		[]string{"10.255.0.1", "10.255.0.4", "10.255.0.5"}, neighborAddrs(bgp))
	assert.True(t, bgp.HasIBGP())
}

func TestRouterOSParser_InterfacesAndAddresses(t *testing.T) {
	cd, err := parsers.NewRouterOSParser().ParseConfig(
		fixture(t, "routeros-bgp.txt"), s.SNMPDevice{IP: "10.255.0.2"})
	require.NoError(t, err)

	byName := map[string]configparser.ConfigInterface{}
	for _, i := range cd.Interfaces {
		byName[i.Name] = i
	}
	require.Len(t, cd.Interfaces, 6)

	assert.Equal(t, "0C:E4:44:65:00:00", byName["ether6"].MACAddress)
	assert.Equal(t, []string{"10.10.12.2/30"}, byName["ether6"].IPAddresses)
	assert.Equal(t, []string{"10.1.2.254/24"}, byName["ether10"].IPAddresses)

	lo := byName["lo0"]
	assert.Equal(t, []string{"10.255.0.2/32"}, lo.IPAddresses, "the iBGP peering address")
	// `mtu=auto` is not a number; the parser must fall back to actual-mtu rather
	// than record a zero.
	assert.Equal(t, 1500, lo.MTU)
}

// -----------------------------------------------------------------------------
// Infix
// -----------------------------------------------------------------------------

// Infix binds OSPF per interface rather than per network, which is why
// ConfigOSPFArea carries both fields. An area whose Networks are empty but whose
// Interfaces are populated is correct here, not a parse failure.
func TestInfixParser_OSPFBindsByInterface(t *testing.T) {
	raw := "# sudo sysrepocfg -X -d running -f json -m ietf-interfaces\n" +
		fixture(t, "infix-interfaces.json") +
		"\n# sudo sysrepocfg -X -d running -f json -m ietf-routing\n" +
		fixture(t, "infix-routing.json") + "\n"

	cd, err := parsers.NewInfixParser().ParseConfig(raw, s.SNMPDevice{IP: "10.255.0.3"})
	require.NoError(t, err)

	assert.Equal(t, []string{configparser.RoutingProtoOSPF}, cd.ControlPlane())

	ospf := protoOfType(t, cd, configparser.RoutingProtoOSPF)
	require.Len(t, ospf.Areas, 1)
	assert.Equal(t, "0.0.0.0", ospf.Areas[0].ID)
	assert.Equal(t,
		[]string{"eth0", "eth1", "eth2", "eth3", "eth4"}, ospf.Areas[0].Interfaces)
	assert.Empty(t, ospf.Areas[0].Networks, "Infix names interfaces, not prefixes")
}

func TestInfixParser_DecodesIETFInterfaceAddresses(t *testing.T) {
	raw := "# sudo sysrepocfg -X -d running -f json -m ietf-interfaces\n" +
		fixture(t, "infix-interfaces.json") + "\n"

	cd, err := parsers.NewInfixParser().ParseConfig(raw, s.SNMPDevice{IP: "10.255.0.3"})
	require.NoError(t, err)

	byName := map[string]configparser.ConfigInterface{}
	for _, i := range cd.Interfaces {
		byName[i.Name] = i
	}

	// ip + prefix-length are separate YANG leaves and must be rejoined as CIDR.
	assert.Equal(t, []string{"10.10.23.2/30"}, byName["eth0"].IPAddresses)
	assert.Equal(t, []string{"10.20.23.2/30"}, byName["eth1"].IPAddresses)
	assert.Equal(t, []string{"10.10.34.1/30"}, byName["eth2"].IPAddresses)
	assert.Equal(t, []string{"10.20.34.1/30"}, byName["eth3"].IPAddresses)
	assert.Equal(t, []string{"10.1.3.254/24"}, byName["eth4"].IPAddresses, "the access LAN gateway")
	assert.Empty(t, byName["eth5"].IPAddresses)
}

// -----------------------------------------------------------------------------
// Cross-vendor
// -----------------------------------------------------------------------------

// The ring is deliberately heterogeneous, so the same area 0 is described three
// different ways: VyOS and FRR list prefixes, RouterOS lists prefixes reached
// through a named template, and Infix lists interfaces. The normalised model has
// to make those comparable — that is what it is for.
func TestControlPlaneIsComparableAcrossVendors(t *testing.T) {
	vyos, err := parsers.NewVyOSParser().ParseConfig(
		fixture(t, "vyos-ospf-config.boot"), s.SNMPDevice{IP: "10.255.0.1"})
	require.NoError(t, err)
	routeros, err := parsers.NewRouterOSParser().ParseConfig(
		fixture(t, "routeros-ospf.txt"), s.SNMPDevice{IP: "10.255.0.2"})
	require.NoError(t, err)
	infix, err := parsers.NewInfixParser().ParseConfig(
		"# sudo sysrepocfg -X -d running -f json -m ietf-routing\n"+
			fixture(t, "infix-routing.json")+"\n", s.SNMPDevice{IP: "10.255.0.3"})
	require.NoError(t, err)

	for _, cd := range []*configparser.ConfigData{vyos, routeros, infix} {
		assert.Equal(t, []string{configparser.RoutingProtoOSPF}, cd.ControlPlane())
	}

	// Every member of the OSPF ring shares one area and one router-id per node,
	// and each router-id is the node's index repeated — R1 1.1.1.1, R2 2.2.2.2,
	// R3 taken from its loopback. Assert the two that state one.
	assert.Equal(t, "1.1.1.1",
		protoOfType(t, vyos, configparser.RoutingProtoOSPF).RouterID)

	var os configparser.ConfigRoutingProtocol
	for _, p := range routeros.RoutingProtocolsOfType(configparser.RoutingProtoOSPF) {
		if p.Instance == "os" {
			os = p
		}
	}
	assert.Equal(t, "2.2.2.2", os.RouterID)
}

// A statically-routed ring member has no dynamic protocol at all, and must
// report that positively rather than looking like a parse failure.
func TestControlPlane_StaticWhenNoProtocolConfigured(t *testing.T) {
	cd := &configparser.ConfigData{
		Routes: []configparser.ConfigRoute{
			{Network: "10.1.1.0/24", Gateway: "10.10.12.1", Protocol: configparser.RoutingProtoStatic},
		},
	}
	assert.Equal(t, []string{configparser.RoutingProtoStatic}, cd.ControlPlane())

	assert.Nil(t, (&configparser.ConfigData{}).ControlPlane(),
		"a device with neither routes nor protocols has no control plane to report")
}

// An operator with a MikroTik in front of them types "mikrotik", not the
// OS name the parser registers under. Rejecting that as an unknown device type
// is a pointless obstacle, so the registry resolves the common vendor names.
func TestRegistryResolvesVendorAliases(t *testing.T) {
	for alias, canonical := range map[string]string{
		"mikrotik":  "routeros",
		"MikroTik":  "routeros",
		"fortigate": "fortinet",
		"pfsense":   "opnsense",
		"vyatta":    "vyos",
		"routeros":  "routeros", // a canonical name still resolves to itself
	} {
		parser, ok := configparser.DefaultRegistry.GetParser(alias)
		require.True(t, ok, "alias %q should resolve", alias)
		assert.Equal(t, canonical, parser.GetOsType(), "alias %q", alias)
	}

	_, ok := configparser.DefaultRegistry.GetParser("nonesuch")
	assert.False(t, ok, "an unknown name must stay unknown, not silently match")
}
