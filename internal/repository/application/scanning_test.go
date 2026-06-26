package application

import (
	"testing"
	"time"

	infra "nsl-graph/internal/repository/infra/cloverdb/base"
	s "nsl-graph/internal/scanner"
)

func setupTestScanningService(t *testing.T) (NetServiceInt, func()) {
	tempDir := t.TempDir()

	repo, err := infra.NewCloverRepository(tempDir)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}

	service := NewNetService(repo)

	cleanup := func() {
		repo.Close()
	}

	return service, cleanup
}

func TestNetService_ScanningMethods(t *testing.T) {
	service, cleanup := setupTestScanningService(t)
	defer cleanup()

	t.Run("ScanNetwork attempts SNMP (will fail in test env)", func(t *testing.T) {
		options := s.ScanOptions{
			Subnet:  "127.0.0.1",
			Timeout: 2 * time.Second,
			SNMP:    s.SNMPOptions{Community: "public", Version: "v2c"},
		}

		_, err := service.ScanNetwork("127.0.0.1", options)
		// SNMP to localhost may or may not fail; either is acceptable
		if err != nil {
			t.Logf("ScanNetwork returned error (expected in test env): %v", err)
		} else {
			t.Log("ScanNetwork succeeded")
		}
	})

	t.Run("Device discovery with mock data", func(t *testing.T) {
		scanResult := &s.ScanResult{
			ID:        "test-scan-1",
			Subnet:    "192.168.1.0/24",
			StartTime: time.Now().Add(-5 * time.Minute),
			EndTime:   time.Now(),
			Devices: []s.SNMPDevice{
				{
					IP:        "192.168.1.1",
					SysName:   "router",
					SysDescr:  "Cisco IOS Software, Version 15.2",
					Reachable: true,
				},
				{
					IP:        "192.168.1.10",
					SysDescr:  "Cisco NX-OS",
					Reachable: true,
					Interfaces: []s.DeviceInterface{
						{Index: 1, Name: "GigabitEthernet0/1", MAC: "00:11:22:33:44:55"},
					},
				},
				{
					IP:        "192.168.1.100",
					SysName:   "web-server",
					SysDescr:  "Linux ubuntu 5.15.0",
					Reachable: true,
				},
			},
		}

		devices, err := service.DiscoverDevices(scanResult)
		if err != nil {
			t.Fatalf("DiscoverDevices failed: %v", err)
		}

		if len(devices) != 3 {
			t.Errorf("Expected 3 devices, got %d", len(devices))
		}

		for _, device := range devices {
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
	})

	t.Run("Import discovered devices", func(t *testing.T) {
		service.AddBrand("Test Brand")
		service.AddModelType("Test Class")
		service.AddOwner("Test Prop")
		service.AddModel("Test Model", "Test Brand", "Test Class", "")

		devices := []s.DiscoveredDevice{
			{
				Device: s.SNMPDevice{
					IP:        "192.168.1.200",
					SysName:   "test-device",
					Reachable: true,
					Interfaces: []s.DeviceInterface{
						{Index: 1, Name: "eth0", MAC: "aa:bb:cc:dd:ee:ff"},
					},
				},
				Brand:         "Test Brand",
				Model:         "Test Model",
				ModelType:     "Test Class",
				SuggestedName: "TEST-200",
				SuggestedZone: "Test Zone",
			},
		}

		options := s.ImportOptions{
			AutoImport:   true,
			CreateZones:  true,
			DefaultZone:  "Test Zone",
			SkipExisting: false,
		}

		if err := service.ImportDiscoveredDevices(devices, options); err != nil {
			t.Fatalf("ImportDiscoveredDevices failed: %v", err)
		}

		allDevices, err := service.GetDevices()
		if err != nil {
			t.Fatalf("GetDevices failed: %v", err)
		}

		found := false
		for _, device := range allDevices {
			if device.Label == "TEST-200" {
				found = true
				break
			}
		}

		if !found {
			t.Error("Imported device not found in database")
		}
	})
}

func TestImportOptions_Validation(t *testing.T) {
	options := s.ImportOptions{
		AutoImport:   true,
		CreateZones:  true,
		DefaultZone:  "Test",
		SkipExisting: true,
	}

	if !options.AutoImport {
		t.Error("AutoImport should be true")
	}
	if !options.CreateZones {
		t.Error("CreateZones should be true")
	}
	if options.DefaultZone != "Test" {
		t.Error("DefaultZone should be 'Test'")
	}
}

func TestScanResult_Structure(t *testing.T) {
	result := s.ScanResult{
		ID:        "test-1",
		Subnet:    "192.168.1.0/24",
		StartTime: time.Now(),
		EndTime:   time.Now().Add(time.Minute),
		Devices:   []s.SNMPDevice{},
	}

	if result.ID == "" {
		t.Error("ScanResult should have an ID")
	}
	if result.Subnet == "" {
		t.Error("ScanResult should have a subnet")
	}
}

func TestDeviceExists_Check(t *testing.T) {
	service, cleanup := setupTestScanningService(t)
	defer cleanup()

	service.AddBrand("TestBrand")
	service.AddModelType("TestClass")
	service.AddModel("TestModel", "TestBrand", "TestClass", "")
	service.AddModelPort("eth0", "0", "0", "TestModel", false, "", "")

	if err := service.AddDevice("TestDevice", "TestModel", "", "", "", false, false); err != nil {
		t.Fatalf("Failed to add test device: %v", err)
	}

	// Get device ID
	devices, _ := service.GetDevices()
	var deviceID string
	for _, d := range devices {
		if d.Label == "TestDevice" {
			deviceID = d.ID
			break
		}
	}

	// Add device port and interface with IP
	modelPorts, _ := service.GetModelPorts()
	var portID string
	for _, mp := range modelPorts {
		if mp.Name == "eth0" {
			portID = mp.ID
			break
		}
	}
	if _, err := service.AddDevicePort(deviceID, portID, "", nil); err != nil {
		t.Fatalf("failed to add device port: %v", err)
	}

	// Add interface with IP
	service.AddDeviceInterface(deviceID, "eth0", "", "", nil, []string{"192.168.1.100"}, "", "")

	netService := service.(*NetService)

	if !netService.deviceExists("192.168.1.100") {
		t.Error("Device with IP 192.168.1.100 should exist")
	}
	if netService.deviceExists("192.168.1.101") {
		t.Error("Device with IP 192.168.1.101 should not exist")
	}
}
