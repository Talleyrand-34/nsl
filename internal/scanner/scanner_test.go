package scanner

import (
	"testing"
)

func TestDeviceDiscoverer_ClassifyDevice(t *testing.T) {
	discoverer := NewDeviceDiscoverer()

	tests := []struct {
		name          string
		device        SNMPDevice
		expectedClass string
	}{
		{
		},
		{
			name:          "Linux is Server",
			device:        SNMPDevice{IP: "192.168.1.100", SysDescr: "Linux server1 5.15.0"},
			expectedClass: "Server",
		},
		{
			name:          "Juniper is Router",
			device:        SNMPDevice{IP: "192.168.1.3", SysDescr: "Juniper Networks, Inc. ex4300"},
			expectedClass: "Router",
		},
		{
			name:          "Unknown defaults to Generic",
			device:        SNMPDevice{IP: "192.168.1.50", SysDescr: ""},
			expectedClass: "Generic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, class := discoverer.ClassifyDevice(tt.device)
			if class != tt.expectedClass {
				t.Errorf("ClassifyDevice() class = %v, want %v", class, tt.expectedClass)
			}
		})
	}
}

func TestDeviceDiscoverer_GenerateDeviceName(t *testing.T) {
	discoverer := NewDeviceDiscoverer()

	tests := []struct {
		name           string
		device         SNMPDevice
		class          string
		expectedPrefix string
	}{
		{
			name:           "Switch should get SW prefix",
			device:         SNMPDevice{IP: "192.168.1.10"},
			class:          "Switch",
			expectedPrefix: "SW-",
		},
		{
			name:           "Router should get RTR prefix",
			device:         SNMPDevice{IP: "192.168.1.1"},
			class:          "Router",
			expectedPrefix: "RTR-",
		},
		{
			name:           "Server should get SRV prefix",
			device:         SNMPDevice{IP: "192.168.1.100"},
			class:          "Server",
			expectedPrefix: "SRV-",
		},
		{
			name:           "Use SysName if available",
			device:         SNMPDevice{IP: "192.168.1.100", SysName: "web-server"},
			class:          "Server",
			expectedPrefix: "web-server",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := discoverer.GenerateDeviceName(tt.device, tt.class)
			if tt.device.SysName != "" {
				if result != tt.expectedPrefix {
					t.Errorf("GenerateDeviceName() = %v, want %v", result, tt.expectedPrefix)
				}
			} else {
				if !startsWith(result, tt.expectedPrefix) {
					t.Errorf("GenerateDeviceName() = %v, should start with %v", result, tt.expectedPrefix)
				}
			}
		})
	}
}

func TestDeviceDiscoverer_SuggestZone(t *testing.T) {
	discoverer := NewDeviceDiscoverer()

	tests := []struct {
		ip   string
		zone string
	}{
		{"192.168.1.1", "LAN"},
		{"10.0.0.1", "Internal"},
		{"172.16.0.1", "Private"},
		{"8.8.8.8", "External"},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			zone := discoverer.SuggestZone(SNMPDevice{IP: tt.ip})
			if zone != tt.zone {
				t.Errorf("SuggestZone(%s) = %v, want %v", tt.ip, zone, tt.zone)
			}
		})
	}
}

func TestScanResult_Structure(t *testing.T) {
	result := ScanResult{
		ID:     "test-1",
		Subnet: "192.168.1.0/24",
	}

	if result.ID == "" {
		t.Error("ScanResult should have an ID")
	}
	if result.Subnet == "" {
		t.Error("ScanResult should have a subnet")
	}
}

func TestDiscoveredDevice_Structure(t *testing.T) {
	device := DiscoveredDevice{
		Device: SNMPDevice{
			IP:      "192.168.1.100",
			SysName: "test-server",
		},
		Brand:         "Linux",
		Model:         "Linux Server",
		ModelType:     "Server",
		SuggestedName: "SRV-100",
		SuggestedZone: "LAN",
	}

	if device.Device.IP == "" {
		t.Error("Device should have an IP")
	}
	if device.SuggestedName == "" {
		t.Error("Device should have a suggested name")
	}
	if device.ModelType == "" {
		t.Error("Device should have a device class")
	}
}

func TestEnumerateSubnet(t *testing.T) {
	ips, err := enumerateSubnet("192.168.1.0/30")
	if err != nil {
		t.Fatalf("enumerateSubnet failed: %v", err)
	}
	// /30 = 4 addresses, minus network and broadcast = 2 usable
	if len(ips) != 2 {
		t.Errorf("Expected 2 IPs for /30, got %d: %v", len(ips), ips)
	}
}

func startsWith(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}
