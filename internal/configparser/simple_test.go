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
package configparser_test

import (
	"testing"

	"nsl-graph/internal/configparser"
	"nsl-graph/internal/configparser/parsers"
	s "nsl-graph/internal/scanner"
)

func createSimpleDevice(sysDescr, sysName string) s.SNMPDevice {
	return s.SNMPDevice{
		IP:        "192.168.1.1",
		SysName:   sysName,
		SysDescr:  sysDescr,
		Reachable: true,
	}
}

func TestOsTypeDetection_OPNsense(t *testing.T) {
	registry := configparser.NewConfigParserRegistry()
	parser := parsers.NewFreeBSDParser()

	registry.RegisterParser(parser)

	device := createSimpleDevice("FreeBSD OPNsense.localdomain", "OPNsense")

	// Test auto-detection
	detectedParser, exists := registry.GetParserForDevice(device)
	if !exists {
		t.Fatal("Should have detected OPNsense device")
	}
	if detectedParser.GetOsType() != "opnsense" {
		t.Errorf("Expected opnsense, got %s", detectedParser.GetOsType())
	}

	// Test manual type selection
	manualParser, exists := registry.GetParserForDeviceWithType(device, "opnsense")
	if !exists {
		t.Fatal("Should have found opnsense parser manually")
	}
	if manualParser.GetOsType() != "opnsense" {
		t.Errorf("Expected opnsense, got %s", manualParser.GetOsType())
	}
}

func TestOsTypeDetection_OpenWrt(t *testing.T) {
	registry := configparser.NewConfigParserRegistry()
	parser := parsers.NewOpenWrtParser()

	registry.RegisterParser(parser)

	// Test device with OpenWrt in description
	device := createSimpleDevice("Linux OpenWrt 5.15.134", "OpenWrt")

	detectedParser, exists := registry.GetParserForDevice(device)
	if !exists {
		t.Fatal("Should have detected OpenWrt device")
	}
	if detectedParser.GetOsType() != "openwrt" {
		t.Errorf("Expected openwrt, got %s", detectedParser.GetOsType())
	}

	// Test manual override for problematic device
	problemDevice := createSimpleDevice("Linux AircubeAC 6.6.93", "HeartOfGold")

	// Auto-detection should fail (no 'openwrt' in description)
	_, exists = registry.GetParserForDevice(problemDevice)
	if exists {
		t.Error("Should NOT have detected OpenWrt from this device automatically")
	}

	// Manual type should work
	manualParser, exists := registry.GetParserForDeviceWithType(problemDevice, "openwrt")
	if !exists {
		t.Fatal("Should have found openwrt parser manually")
	}
	if manualParser.GetOsType() != "openwrt" {
		t.Errorf("Expected openwrt, got %s", manualParser.GetOsType())
	}
}

func TestOsTypeDetection_Fortinet(t *testing.T) {
	registry := configparser.NewConfigParserRegistry()
	parser := parsers.NewFortinetParser()

	registry.RegisterParser(parser)

	device := createSimpleDevice("FortiGate-60F v7.4.1", "FortiGate-60F")

	detectedParser, exists := registry.GetParserForDevice(device)
	if !exists {
		t.Fatal("Should have detected Fortinet device")
	}
	if detectedParser.GetOsType() != "fortinet" {
		t.Errorf("Expected fortinet, got %s", detectedParser.GetOsType())
	}
}

func TestManualTypeOverride(t *testing.T) {
	registry := configparser.NewConfigParserRegistry()

	// Register multiple parsers
	opnsenseParser := parsers.NewFreeBSDParser()
	openwrtParser := parsers.NewOpenWrtParser()

	registry.RegisterParser(opnsenseParser)
	registry.RegisterParser(openwrtParser)

	// Device that would auto-detect as OPNsense
	device := createSimpleDevice("FreeBSD OPNsense.localdomain", "OPNsense")

	// Verify auto-detection works
	autoParser, exists := registry.GetParserForDevice(device)
	if !exists || autoParser.GetOsType() != "opnsense" {
		t.Fatal("Auto-detection should find OPNsense")
	}

	// Override with manual type
	manualParser, exists := registry.GetParserForDeviceWithType(device, "openwrt")
	if !exists || manualParser.GetOsType() != "openwrt" {
		t.Fatal("Manual type should override auto-detection")
	}
}

func TestInvalidManualType(t *testing.T) {
	registry := configparser.NewConfigParserRegistry()
	opnsenseParser := parsers.NewFreeBSDParser()
	registry.RegisterParser(opnsenseParser)

	device := createSimpleDevice("FreeBSD OPNsense.localdomain", "OPNsense")

	// Try invalid type - should fallback to auto-detection
	fallbackParser, exists := registry.GetParserForDeviceWithType(device, "invalidtype")
	if !exists || fallbackParser.GetOsType() != "opnsense" {
		t.Fatal("Invalid type should fallback to auto-detection")
	}
}

func TestNoParserAvailable(t *testing.T) {
	registry := configparser.NewConfigParserRegistry()

	device := createSimpleDevice("Unknown Device", "unknown")

	// No parsers registered - should not find any
	_, exists := registry.GetParserForDevice(device)
	if exists {
		t.Error("Should not find parser for unknown device")
	}

	_, exists = registry.GetParserForDeviceWithType(device, "opnsense")
	if exists {
		t.Error("Should not find unregistered parser type")
	}
}

func TestRealWorldScenarios(t *testing.T) {
	registry := configparser.NewConfigParserRegistry()

	// Register real parsers
	registry.RegisterParser(parsers.NewFreeBSDParser())
	registry.RegisterParser(parsers.NewOpenWrtParser())
	registry.RegisterParser(parsers.NewFortinetParser())

	testCases := []struct {
		name         string
		device       s.SNMPDevice
		expectedType string
		shouldDetect bool
	}{
		{
			name:         "Real OPNsense",
			device:       createSimpleDevice("FreeBSD OPNsense.i.t34.dev 14.3-RELEASE-p2", "OPNsense.i.t34.dev"),
			expectedType: "opnsense",
			shouldDetect: true,
		},
		{
			name:         "Real OpenWrt (problematic)",
			device:       createSimpleDevice("Linux AircubeAC 6.6.93 #0 Mon Jun 23 20:40:36 2025 mips", "HeartOfGold"),
			expectedType: "",
			shouldDetect: false, // Current limitation - no 'openwrt' in description
		},
		{
			name:         "TP-Link Switch",
			device:       createSimpleDevice("JetStream 24-Port Gigabit Smart Switch", "T1600G-28TS"),
			expectedType: "",
			shouldDetect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, exists := registry.GetParserForDevice(tc.device)

			if tc.shouldDetect {
				if !exists {
					t.Errorf("Should have detected OS type for %s", tc.name)
					return
				}
				if parser.GetOsType() != tc.expectedType {
					t.Errorf("Expected %s, got %s", tc.expectedType, parser.GetOsType())
				}
			} else {
				if exists {
					t.Errorf("Should NOT have detected OS type for %s", tc.name)
				}
			}

			// Test manual override for problematic device
			if tc.name == "Real OpenWrt (problematic)" {
				manualParser, exists := registry.GetParserForDeviceWithType(tc.device, "openwrt")
				if !exists || manualParser.GetOsType() != "openwrt" {
					t.Error("Manual type should work for problematic OpenWrt device")
				}
			}
		})
	}
}

func TestConfigParserOptions_OsType(t *testing.T) {
	options := configparser.ConfigParserOptions{
		Source: configparser.ConfigSourceSSH,
		OsType: "opnsense",
		SSHCredentials: &configparser.SSHCredentials{
			Username: "admin",
			Password: "password",
		},
		DiscrepancyAction: configparser.DiscrepancyActionPreferSNMP,
		MergeWithSNMP:     true,
	}

	if options.OsType != "opnsense" {
		t.Errorf("Expected opnsense, got %s", options.OsType)
	}
	if options.Source != configparser.ConfigSourceSSH {
		t.Errorf("Expected SSH source, got %v", options.Source)
	}
}
