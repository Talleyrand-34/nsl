/*
Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

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

	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

func createTestDevice(sysDescr, sysName string) s.SNMPDevice {
	return s.SNMPDevice{
		IP:       "192.168.1.1",
		SysName:  sysName,
		SysDescr: sysDescr,
		Reachable: true,
	}
}

func TestOPNsenseParser_GetDeviceType(t *testing.T) {
	parser := parsers.NewOPNsenseParser()
	assert.Equal(t, "opnsense", parser.GetDeviceType())
}

func TestOPNsenseParser_SupportsDevice_FreeBSD_OPNsense(t *testing.T) {
	parser := parsers.NewOPNsenseParser()

	// Test FreeBSD with OPNsense in description
	device := createTestDevice("FreeBSD OPNsense.localdomain 23.7.9-RELEASE FreeBSD", "OPNsense.localdomain")
	assert.True(t, parser.SupportsDevice(device))

	// Test with freebsd (lowercase)
	device = createTestDevice("freebsd opnsense system", "test")
	assert.True(t, parser.SupportsDevice(device))

	// Test with OPNsense in system name only
	device = createTestDevice("FreeBSD system", "OPNsense")
	assert.True(t, parser.SupportsDevice(device))

	// Test with opnsense (lowercase) in system name
	device = createTestDevice("FreeBSD system", "opnsense.example.com")
	assert.True(t, parser.SupportsDevice(device))
}

func TestOPNsenseParser_SupportsDevice_NoMatch(t *testing.T) {
	parser := parsers.NewOPNsenseParser()

	// Test device that doesn't match
	device := createTestDevice("Linux Router", "router.example.com")
	assert.False(t, parser.SupportsDevice(device))

	// Test Cisco device
	device = createTestDevice("Cisco IOS Router", "router")
	assert.False(t, parser.SupportsDevice(device))

	// Test empty description
	device = createTestDevice("", "")
	assert.False(t, parser.SupportsDevice(device))
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
	parsers := map[string]interface{}{
		"opnsense": parsers.NewOPNsenseParser(),
		"openwrt":  parsers.NewOpenWrtParser(),
		"fortinet": parsers.NewFortinetParser(),
		"cisco":    parsers.NewCiscoParser(),
	}

	devices := map[string]s.SNMPDevice{
		"opnsense_real":    createTestDevice("FreeBSD OPNsense.i.t34.dev 14.3-RELEASE-p2", "OPNsense.i.t34.dev"),
		"openwrt_real":     createTestDevice("Linux AircubeAC 6.6.93 #0 Mon Jun 23 20:40:36 2025 mips", "HeartOfGold"),
		"fortinet_real":    createTestDevice("FortiGate-60F v7.4.1,build2463,230314", "FortiGate-60F"),
		"cisco_real":       createTestDevice("Cisco IOS Software, C2960X Software", "Switch"),
		"generic_device":   createTestDevice("Generic Network Device", "generic"),
		"linux_generic":    createTestDevice("Linux 5.4.0-generic", "server"),
	}

	// Expected detection matrix
	expectedMatches := map[string]map[string]bool{
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

	for parserName, parser := range parsers {
		for deviceName, device := range devices {
			expected := expectedMatches[parserName][deviceName]

			var actual bool
			switch p := parser.(type) {
			case *parsers.OPNsenseParser:
				actual = p.SupportsDevice(device)
			case *parsers.OpenWrtParser:
				actual = p.SupportsDevice(device)
			case *parsers.FortinetParser:
				actual = p.SupportsDevice(device)
			case *parsers.CiscoParser:
				actual = p.SupportsDevice(device)
			}

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
	opnsenseParser := parsers.NewOPNsenseParser()
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
	parsers := map[string]interface{}{
		"opnsense": parsers.NewOPNsenseParser(),
		"openwrt":  parsers.NewOpenWrtParser(),
		"fortinet": parsers.NewFortinetParser(),
		"cisco":    parsers.NewCiscoParser(),
	}

	// Test empty device data
	emptyDevice := createTestDevice("", "")
	for name, parser := range parsers {
		var supports bool
		switch p := parser.(type) {
		case *parsers.OPNsenseParser:
			supports = p.SupportsDevice(emptyDevice)
		case *parsers.OpenWrtParser:
			supports = p.SupportsDevice(emptyDevice)
		case *parsers.FortinetParser:
			supports = p.SupportsDevice(emptyDevice)
		case *parsers.CiscoParser:
			supports = p.SupportsDevice(emptyDevice)
		}
		assert.False(t, supports, "Parser %s should not support empty device", name)
	}

	// Test case sensitivity
	mixedCaseDevice := createTestDevice("FreeBSD OpNsEnSe System", "OpNsEnSe")
	opnsenseParser := parsers.NewOPNsenseParser()
	assert.True(t, opnsenseParser.SupportsDevice(mixedCaseDevice))
}