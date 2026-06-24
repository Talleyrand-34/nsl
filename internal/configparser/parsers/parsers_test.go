/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

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
package parsers_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

func createTestDevice(sysDescr, sysName string) s.SNMPDevice {
	return s.SNMPDevice{
		IP:        "192.168.1.1",
		SysName:   sysName,
		SysDescr:  sysDescr,
		Reachable: true,
	}
}

// --- Legacy XML parser ---

func TestOPNsenseLegacyParser_GetDeviceType(t *testing.T) {
	parser := parsers.NewOPNsenseParser()
	assert.Equal(t, "opnsense-xml", parser.GetDeviceType())
}

func TestOPNsenseLegacyParser_SupportsDevice_AlwaysFalse(t *testing.T) {
	parser := parsers.NewOPNsenseParser()
	// Legacy XML parser never auto-detects; always returns false.
	assert.False(t, parser.SupportsDevice(createTestDevice("FreeBSD OPNsense 23.7", "fw01")))
	assert.False(t, parser.SupportsDevice(createTestDevice("freebsd opnsense system", "test")))
	assert.False(t, parser.SupportsDevice(createTestDevice("", "")))
}

// --- FreeBSD ifconfig parser (primary OPNsense parser) ---

func TestFreeBSDParser_GetDeviceType(t *testing.T) {
	parser := parsers.NewFreeBSDParser()
	assert.Equal(t, "opnsense", parser.GetDeviceType())
}

func TestFreeBSDParser_SupportsDevice(t *testing.T) {
	parser := parsers.NewFreeBSDParser()

	// FreeBSD with OPNsense in description
	assert.True(t, parser.SupportsDevice(createTestDevice("FreeBSD OPNsense.localdomain 23.7.9-RELEASE FreeBSD", "OPNsense.localdomain")))
	// FreeBSD (lowercase)
	assert.True(t, parser.SupportsDevice(createTestDevice("freebsd opnsense system", "test")))
	// OPNsense in SysName only
	assert.True(t, parser.SupportsDevice(createTestDevice("FreeBSD system", "OPNsense")))
	// opnsense (lowercase) in SysName
	assert.True(t, parser.SupportsDevice(createTestDevice("FreeBSD system", "opnsense.example.com")))

	// Non-FreeBSD devices should not match
	assert.False(t, parser.SupportsDevice(createTestDevice("Linux Router", "router.example.com")))
	assert.False(t, parser.SupportsDevice(createTestDevice("Cisco IOS Router", "router")))
	assert.False(t, parser.SupportsDevice(createTestDevice("", "")))
}

func TestFreeBSDParser_ParseConfig(t *testing.T) {
	const fixture = "HOSTNAME:fw01\n" +
		"igc0: flags=8843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST> metric 0 mtu 1500\n" +
		"\toptions=4e507bb<RXCSUM,TXCSUM,VLAN_MTU,VLAN_HWTAGGING,VLAN_HWCSUM,TSO4,LINKSTATE>\n" +
		"\tether aa:bb:cc:dd:ee:ff\n" +
		"\tinet 10.0.0.1/24 broadcast 10.0.0.255\n" +
		"\tstatus: active\n" +
		"\tnd6 options=23<PERFORMNUD,ACCEPT_RTADV,AUTO_LINKLOCAL>\n" +
		"igc0.10: flags=8843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST> metric 0 mtu 1500\n" +
		"\tether aa:bb:cc:dd:ee:ff\n" +
		"\tinet 10.0.10.1/23 broadcast 10.0.11.255\n" +
		"\tvlan: 10 vlanpcp: 0 parent interface: igc0\n" +
		"\tstatus: active\n" +
		"lo0: flags=8049<UP,LOOPBACK,RUNNING,MULTICAST> metric 0 mtu 16384\n" +
		"\tinet 127.0.0.1/8\n" +
		"\tinet6 ::1/128\n" +
		"\tinet6 fe80::1%lo0/64\n" + // link-local — should be skipped
		"\tstatus: active\n" +
		"openvpn: flags=8051<UP,POINTOPOINT,RUNNING,MULTICAST> metric 0 mtu 1500\n" +
		"\tstatus: active\n"

	parser := parsers.NewFreeBSDParser()
	device := createTestDevice("FreeBSD OPNsense.i.t34.dev 14.3-RELEASE-p2", "fw01")

	configData, err := parser.ParseConfig(fixture, device)
	assert.NoError(t, err)
	assert.NotNil(t, configData)

	assert.Equal(t, "fw01", configData.Hostname)
	assert.Equal(t, "opnsense", configData.DeviceType)

	ifaceByName := make(map[string]configparser.ConfigInterface)
	for _, iface := range configData.Interfaces {
		ifaceByName[iface.Name] = iface
	}

	// igc0 — physical, UP, has IP and MAC, also has implicit VLAN 1 untagged
	// (physical interfaces that are parents of VLAN subinterfaces are implicitly in VLAN 1)
	igc0, ok := ifaceByName["igc0"]
	assert.True(t, ok, "igc0 should be present")
	assert.Equal(t, "physical", igc0.Type)
	assert.True(t, igc0.Enabled)
	assert.Equal(t, "aa:bb:cc:dd:ee:ff", igc0.MACAddress)
	assert.Contains(t, igc0.IPAddresses, "10.0.0.1/24")
	assert.Len(t, igc0.VLANs, 1)
	assert.Equal(t, "1", igc0.VLANs[0].ID)
	assert.False(t, igc0.VLANs[0].Tagged) // untagged

	// igc0.10 — VLAN subinterface
	vlanIface, ok := ifaceByName["igc0.10"]
	assert.True(t, ok, "igc0.10 should be present")
	assert.Equal(t, "vlan", vlanIface.Type)
	assert.Equal(t, "igc0", vlanIface.Parent)
	assert.Contains(t, vlanIface.IPAddresses, "10.0.10.1/23")
	assert.Len(t, vlanIface.VLANs, 1)
	assert.Equal(t, "10", vlanIface.VLANs[0].ID)
	assert.True(t, vlanIface.VLANs[0].Tagged)

	// lo0 — loopback, non-scoped IPv6 included, link-local skipped
	lo0, ok := ifaceByName["lo0"]
	assert.True(t, ok, "lo0 should be present")
	assert.Equal(t, "loopback", lo0.Type)
	assert.Contains(t, lo0.IPAddresses, "127.0.0.1/8")
	assert.Contains(t, lo0.IPAddresses, "::1/128")
	for _, ip := range lo0.IPAddresses {
		assert.NotContains(t, ip, "%", "link-local scoped addresses should be excluded")
	}

	// openvpn — tunnel (POINTOPOINT flag)
	vpn, ok := ifaceByName["openvpn"]
	assert.True(t, ok, "openvpn should be present")
	assert.Equal(t, "tunnel", vpn.Type)

	// Top-level VLAN from igc0.10
	assert.Len(t, configData.VLANs, 1)
	assert.Equal(t, "10", configData.VLANs[0].ID)
	assert.Equal(t, "VLAN_10", configData.VLANs[0].Name)
}

func TestFreeBSDParser_ParseConfig_NoHostnamePrefix(t *testing.T) {
	// When fed a raw ifconfig dump (e.g. from --config-source file), no HOSTNAME: line.
	const raw = "em0: flags=8843<UP,BROADCAST,RUNNING,SIMPLEX,MULTICAST> metric 0 mtu 1500\n" +
		"\tether de:ad:be:ef:00:01\n" +
		"\tinet 192.168.1.1/24 broadcast 192.168.1.255\n" +
		"\tstatus: active\n"

	parser := parsers.NewFreeBSDParser()
	device := createTestDevice("FreeBSD 14.3-RELEASE", "fw-from-snmp")

	configData, err := parser.ParseConfig(raw, device)
	assert.NoError(t, err)
	// Falls back to deviceInfo.SysName
	assert.Equal(t, "fw-from-snmp", configData.Hostname)
	assert.Len(t, configData.Interfaces, 1)
	assert.Equal(t, "em0", configData.Interfaces[0].Name)
	assert.Contains(t, configData.Interfaces[0].IPAddresses, "192.168.1.1/24")
}

func TestOpenWrtParser_GetDeviceType(t *testing.T) {
	parser := parsers.NewOpenWrtParser()
	assert.Equal(t, "openwrt", parser.GetDeviceType())
}

func TestOpenWrtParser_SupportsDevice_Linux_OpenWrt(t *testing.T) {
	parser := parsers.NewOpenWrtParser()

	// Test Linux with OpenWrt in description
	device := createTestDevice("Linux OpenWrt 5.15.134 #0 SMP PREEMPT", "OpenWrt")
	assert.True(t, parser.SupportsDevice(device))

	// Test Linux with openwrt (lowercase)
	device = createTestDevice("Linux openwrt router", "test")
	assert.True(t, parser.SupportsDevice(device))

	// Test Linux with OpenWrt in system name
	device = createTestDevice("Linux system", "OpenWrt")
	assert.True(t, parser.SupportsDevice(device))
}

func TestOpenWrtParser_SupportsDevice_NoMatch(t *testing.T) {
	parser := parsers.NewOpenWrtParser()

	// Test Linux without OpenWrt indicators
	device := createTestDevice("Linux generic system", "router")
	assert.False(t, parser.SupportsDevice(device))

	// Test OpenWrt without Linux (shouldn't match)
	device = createTestDevice("FreeBSD OpenWrt", "OpenWrt")
	assert.False(t, parser.SupportsDevice(device))

	// Test non-Linux system
	device = createTestDevice("FreeBSD system", "router")
	assert.False(t, parser.SupportsDevice(device))
}

func TestFortinetParser_GetDeviceType(t *testing.T) {
	parser := parsers.NewFortinetParser()
	assert.Equal(t, "fortinet", parser.GetDeviceType())
}

func TestFortinetParser_SupportsDevice_Fortinet(t *testing.T) {
	parser := parsers.NewFortinetParser()

	// Test FortiGate in description
	device := createTestDevice("FortiGate-60F v7.4.1,build2463", "FortiGate-60F")
	assert.True(t, parser.SupportsDevice(device))

	// Test fortinet (lowercase) in description
	device = createTestDevice("fortinet security appliance", "FG-device")
	assert.True(t, parser.SupportsDevice(device))

	// Test FortiGate in system name
	device = createTestDevice("Security Device", "FortiGate")
	assert.True(t, parser.SupportsDevice(device))

	// Test FG prefix in system name
	device = createTestDevice("Security Device", "FG-60F")
	assert.True(t, parser.SupportsDevice(device))
}

func TestFortinetParser_SupportsDevice_NoMatch(t *testing.T) {
	parser := parsers.NewFortinetParser()

	// Test non-Fortinet device
	device := createTestDevice("Cisco ASA Security", "ASA-5506")
	assert.False(t, parser.SupportsDevice(device))

	// Test generic device
	device = createTestDevice("Generic Router", "router")
	assert.False(t, parser.SupportsDevice(device))
}

func TestCiscoParser_GetDeviceType(t *testing.T) {
	parser := parsers.NewCiscoParser()
	assert.Equal(t, "cisco", parser.GetDeviceType())
}

func TestCiscoParser_SupportsDevice_Cisco(t *testing.T) {
	parser := parsers.NewCiscoParser()

	// Test Cisco in description
	device := createTestDevice("Cisco IOS Software, C2960X Software", "Switch")
	assert.True(t, parser.SupportsDevice(device))

	// Test cisco (lowercase) in description
	device = createTestDevice("cisco router system", "router")
	assert.True(t, parser.SupportsDevice(device))

	// Test IOS in description
	device = createTestDevice("IOS Software Release 15.1", "router")
	assert.True(t, parser.SupportsDevice(device))
}

func TestCiscoParser_SupportsDevice_NoMatch(t *testing.T) {
	parser := parsers.NewCiscoParser()

	// Test non-Cisco device
	device := createTestDevice("Juniper JUNOS", "juniper-device")
	assert.False(t, parser.SupportsDevice(device))

	// Test generic device
	device = createTestDevice("Linux Router", "router")
	assert.False(t, parser.SupportsDevice(device))
}

func TestAllParsers_DetectionMatrix(t *testing.T) {
	// Test matrix of devices vs parsers to ensure proper detection
	allParsers := map[string]configparser.ConfigParser{
		"opnsense":     parsers.NewFreeBSDParser(),
		"opnsense-xml": parsers.NewOPNsenseParser(),
		"openwrt":      parsers.NewOpenWrtParser(),
		"fortinet":     parsers.NewFortinetParser(),
		"cisco":        parsers.NewCiscoParser(),
	}

	devices := map[string]s.SNMPDevice{
		"opnsense_real":  createTestDevice("FreeBSD OPNsense.i.t34.dev 14.3-RELEASE-p2", "OPNsense.i.t34.dev"),
		"openwrt_real":   createTestDevice("Linux AircubeAC 6.6.93 #0 Mon Jun 23 20:40:36 2025 mips", "HeartOfGold"),
		"fortinet_real":  createTestDevice("FortiGate-60F v7.4.1,build2463,230314", "FortiGate-60F"),
		"cisco_real":     createTestDevice("Cisco IOS Software, C2960X Software", "Switch"),
		"generic_device": createTestDevice("Generic Network Device", "generic"),
		"linux_generic":  createTestDevice("Linux 5.4.0-generic", "server"),
	}

	// Expected detection matrix
	expectedMatches := map[string]map[string]bool{
		"opnsense-xml": {
			"opnsense_real":  false,
			"openwrt_real":   false,
			"fortinet_real":  false,
			"cisco_real":     false,
			"generic_device": false,
			"linux_generic":  false,
		},
		"opnsense": {
			"opnsense_real":  true,
			"openwrt_real":   false,
			"fortinet_real":  false,
			"cisco_real":     false,
			"generic_device": false,
			"linux_generic":  false,
		},
		"openwrt": {
			"opnsense_real":  false,
			"openwrt_real":   false, // Actual device lacks 'openwrt' in description
			"fortinet_real":  false,
			"cisco_real":     false,
			"generic_device": false,
			"linux_generic":  false, // Linux but no OpenWrt indicators
		},
		"fortinet": {
			"opnsense_real":  false,
			"openwrt_real":   false,
			"fortinet_real":  true,
			"cisco_real":     false,
			"generic_device": false,
			"linux_generic":  false,
		},
		"cisco": {
			"opnsense_real":  false,
			"openwrt_real":   false,
			"fortinet_real":  false,
			"cisco_real":     true,
			"generic_device": false,
			"linux_generic":  false,
		},
	}

	for parserName, p := range allParsers {
		for deviceName, device := range devices {
			expected := expectedMatches[parserName][deviceName]
			actual := p.SupportsDevice(device)
			assert.Equal(t, expected, actual,
				"Parser %s should %s device %s (SysDescr: %s)",
				parserName,
				map[bool]string{true: "detect", false: "NOT detect"}[expected],
				deviceName,
				device.SysDescr,
			)
		}
	}
}

func TestParserDetection_RealWorldCases(t *testing.T) {
	// Test with actual SNMP data from the network
	opnsenseParser := parsers.NewFreeBSDParser()
	openwrtParser := parsers.NewOpenWrtParser()
	fortinetParser := parsers.NewFortinetParser()
	ciscoParser := parsers.NewCiscoParser()

	// Real OPNsense device (10.0.0.1)
	opnsenseDevice := createTestDevice(
		"FreeBSD OPNsense.i.t34.dev 14.3-RELEASE-p2 FreeBSD 14.3-RELEASE-p2 stable/25.7-n271676-ab2281de1853 SMP amd64",
		"OPNsense.i.t34.dev",
	)
	assert.True(t, opnsenseParser.SupportsDevice(opnsenseDevice))
	assert.False(t, openwrtParser.SupportsDevice(opnsenseDevice))
	assert.False(t, fortinetParser.SupportsDevice(opnsenseDevice))
	assert.False(t, ciscoParser.SupportsDevice(opnsenseDevice))

	// Real OpenWrt device (10.0.0.245) - currently lacks OpenWrt in description
	openwrtDevice := createTestDevice(
		"Linux AircubeAC 6.6.93 #0 Mon Jun 23 20:40:36 2025 mips",
		"HeartOfGold",
	)
	assert.False(t, opnsenseParser.SupportsDevice(openwrtDevice))
	assert.False(t, openwrtParser.SupportsDevice(openwrtDevice)) // Current limitation
	assert.False(t, fortinetParser.SupportsDevice(openwrtDevice))
	assert.False(t, ciscoParser.SupportsDevice(openwrtDevice))

	// TP-Link switch (10.0.0.4)
	tplinkDevice := createTestDevice(
		"JetStream 24-Port Gigabit Smart Switch with 4 SFP Slots",
		"T1600G-28TS",
	)
	assert.False(t, opnsenseParser.SupportsDevice(tplinkDevice))
	assert.False(t, openwrtParser.SupportsDevice(tplinkDevice))
	assert.False(t, fortinetParser.SupportsDevice(tplinkDevice))
	assert.False(t, ciscoParser.SupportsDevice(tplinkDevice))
}

func TestParserDetection_EdgeCases(t *testing.T) {
	allParsers := map[string]configparser.ConfigParser{
		"opnsense": parsers.NewFreeBSDParser(),
		"openwrt":  parsers.NewOpenWrtParser(),
		"fortinet": parsers.NewFortinetParser(),
		"cisco":    parsers.NewCiscoParser(),
	}

	// Test empty device data
	emptyDevice := createTestDevice("", "")
	for name, p := range allParsers {
		assert.False(t, p.SupportsDevice(emptyDevice), "Parser %s should not support empty device", name)
	}

	// Test case sensitivity
	mixedCaseDevice := createTestDevice("FreeBSD OpNsEnSe System", "OpNsEnSe")
	assert.True(t, parsers.NewFreeBSDParser().SupportsDevice(mixedCaseDevice))
}

func TestFortinetParser_ParseConfig(t *testing.T) {
	// Fixture mirrors real device output: show system global + show system interface
	// show system interface produces clean output with no nested config...end blocks
	const minimalConfig = `
config system global
    set hostname "fw-branch"
end
config system interface
    edit "wan"
        set ip 10.0.0.247 255.255.255.0
        set allowaccess ping https ssh http fgfm
        set type physical
        set role wan
        set snmp-index 1
    next
    edit "ssl.root"
        set type tunnel
        set alias "SSL VPN interface"
        set snmp-index 3
    next
    edit "vlan4"
        set ip 10.0.4.247 255.255.255.0
        set allowaccess ping
        set role lan
        set snmp-index 5
        set interface "wan"
        set vlanid 4
    next
end
`
	parser := parsers.NewFortinetParser()
	device := createTestDevice("FortiGate-60F v7.4.1,build2463", "fw-branch")

	configData, err := parser.ParseConfig(minimalConfig, device)
	assert.NoError(t, err)
	assert.NotNil(t, configData)

	assert.Equal(t, "fw-branch", configData.Hostname)

	ifaceByName := make(map[string]configparser.ConfigInterface)
	for _, iface := range configData.Interfaces {
		ifaceByName[iface.Name] = iface
	}

	// wan: physical interface
	wan, ok := ifaceByName["wan"]
	assert.True(t, ok, "wan interface should be present")
	assert.Equal(t, "physical", wan.Type)
	assert.Contains(t, wan.IPAddresses, "10.0.0.247/24")

	// vlan4: VLAN interface with correct VLAN ID and parent
	vlan4, ok := ifaceByName["vlan4"]
	assert.True(t, ok, "vlan4 interface should be present")
	assert.Equal(t, "vlan", vlan4.Type)
	assert.Equal(t, "wan", vlan4.Parent)
	assert.Contains(t, vlan4.IPAddresses, "10.0.4.247/24")
	assert.Len(t, vlan4.VLANs, 1)
	assert.Equal(t, "4", vlan4.VLANs[0].ID)
}

func TestOPNsenseParser_ParseConfig(t *testing.T) {
	const minimalConfig = `<?xml version="1.0"?>
<opnsense>
  <version>23.7</version>
  <system>
    <hostname>fw01</hostname>
    <domain>example.com</domain>
  </system>
  <interfaces>
    <wan>
      <if>em0</if>
      <descr>WAN</descr>
      <enable>1</enable>
      <ipaddr>203.0.113.1</ipaddr>
      <subnet>24</subnet>
    </wan>
    <lan>
      <if>em1</if>
      <descr>LAN</descr>
      <enable>1</enable>
      <ipaddr>192.168.1.1</ipaddr>
      <subnet>24</subnet>
    </lan>
  </interfaces>
  <vlans>
    <vlan>
      <if>vtnet0</if>
      <tag>10</tag>
      <descr>Management VLAN</descr>
    </vlan>
  </vlans>
</opnsense>`

	parser := parsers.NewOPNsenseParser()
	device := createTestDevice("FreeBSD OPNsense 23.7", "fw01")

	configData, err := parser.ParseConfig(minimalConfig, device)
	assert.NoError(t, err)
	assert.NotNil(t, configData)

	assert.Equal(t, "fw01", configData.Hostname)
	assert.Equal(t, "example.com", configData.Domain)

	// 2 named interfaces (wan, lan) + 1 VLAN interface entry (vtnet0.10)
	assert.Len(t, configData.Interfaces, 3)
	ifaceNames := make(map[string]bool)
	for _, iface := range configData.Interfaces {
		ifaceNames[iface.Name] = true
	}
	assert.True(t, ifaceNames["wan"])
	assert.True(t, ifaceNames["lan"])
	assert.True(t, ifaceNames["vtnet0.10"])

	assert.Len(t, configData.VLANs, 1)
	assert.Equal(t, "10", configData.VLANs[0].ID)
}

func TestOpenWrtParser_ParseBoardJSON(t *testing.T) {
	const boardJSONFixture = `# uci show
network.loopback=interface
network.loopback.device='lo'
network.loopback.proto='static'
# cat /etc/board.json
{
	"model": {
		"id": "ubnt,aircube-ac",
		"name": "Ubiquiti airCube AC"
	},
	"switch": {
		"switch0": {
			"enable": true,
			"reset": true,
			"ports": [
				{
					"num": 0,
					"device": "eth0",
					"need_tag": false,
					"want_untag": false
				},
				{
					"num": 2,
					"role": "lan",
					"index": 1
				},
				{
					"num": 3,
					"role": "lan",
					"index": 2
				},
				{
					"num": 5,
					"role": "lan",
					"index": 3
				},
				{
					"num": 4,
					"role": "wan"
				}
			],
			"roles": [
				{
					"role": "lan",
					"ports": "2 3 5 0t",
					"device": "eth0.1"
				},
				{
					"role": "wan",
					"ports": "4 0t",
					"device": "eth0.2"
				}
			]
		}
	},
	"network": {
		"lan": {
			"device": "eth0.1",
			"protocol": "static"
		},
		"wan": {
			"device": "eth0.2",
			"protocol": "dhcp"
		}
	}
}`

	parser := parsers.NewOpenWrtParser()
	device := createTestDevice("Linux AircubeAC 6.6.93", "HeartOfGold")

	configData, err := parser.ParseConfig(boardJSONFixture, device)
	assert.NoError(t, err)
	assert.NotNil(t, configData)

	// This airCube board.switch exposes only internal swconfig fabric ports (no
	// per-port "device": they share the eth0 CPU uplink via eth0.1/eth0.2). Such
	// ports aren't individually addressable by LLDP/FDB, so they must NOT be
	// emitted as switch ports — otherwise scan-host would invent phantom lan1..lan3
	// ports the connection scan can never match. The real uplink (eth0) comes from
	// the UCI interface parse instead.
	portByName := make(map[string]configparser.SwitchPortInfo)
	for _, port := range configData.SwitchPorts {
		portByName[port.PortName] = port
	}
	assert.NotContains(t, portByName, "lan1")
	assert.NotContains(t, portByName, "lan2")
	assert.NotContains(t, portByName, "lan3")
	assert.NotContains(t, portByName, "wan0")

	for _, port := range configData.SwitchPorts {
		assert.NotEqual(t, "eth0", port.PortName, "CPU port eth0 should not be included")
		assert.NotEqual(t, 0, port.PortNumber, "CPU port should not be included")
		assert.NotContains(t, port.PortName, ".", "VLAN subinterfaces are not physical switch ports")
	}
}
