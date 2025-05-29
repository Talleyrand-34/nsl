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
package basicops

import (
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestDevicePort_AddAndGetDevicePorts(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare all referenced data: Brand, DeviceClass, Model, ModelPort, Proprietary, ZoneType, Zone, Device
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("F100", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("eth0", "10", "20", "F100"); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "IT Department", "Physical"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddDevice("MainRouter", "F100", "", "HQ", "IT Department"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and modelport IDs
	devices, _ := repo.GetDevices()
	modelPorts, _ := repo.GetModelPorts()
	if len(devices) == 0 || len(modelPorts) == 0 {
		t.Fatalf("expected at least one device and one model port")
	}
	deviceID := strconv.Itoa(devices[0].ID)
	modelPortID := strconv.Itoa(modelPorts[0].ID)

	// Add device port with empty mac_address
	if err := repo.AddDevicePort(deviceID, modelPortID, ""); err != nil {
		t.Errorf("failed to add device port: %v", err)
	}

	// Retrieve and verify
	devPorts, _ := repo.GetDevicePorts()
	found := false
	for _, dp := range devPorts {
		if dp.DeviceID == devices[0].ID && dp.ModelID == modelPorts[0].ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf(
			"expected device port with device ID %d and model port ID %d in list, got %+v",
			devices[0].ID,
			modelPorts[0].ID,
			devPorts,
		)
	}
}

func TestDevicePort_AddDevicePort_InvalidIDs(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Invalid device ID
	err = repo.AddDevicePort("notanumber", "1", "")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid device ID, got error: %v", err)
	}

	// Invalid model port ID
	err = repo.AddDevicePort("1", "notanumber", "")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid model port ID, got error: %v", err)
	}
}

func TestDevicePort_AddGetAndDeleteDevicePorts(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare all referenced data: Brand, DeviceClass, Model, ModelPort, Proprietary, ZoneType, Zone, Device
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("F100", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("eth0", "10", "20", "F100"); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "IT Department", "Physical"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddDevice("MainRouter", "F100", "", "HQ", "IT Department"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and modelport IDs
	devices, _ := repo.GetDevices()
	modelPorts, _ := repo.GetModelPorts()
	if len(devices) == 0 || len(modelPorts) == 0 {
		t.Fatalf("expected at least one device and one model port")
	}
	deviceID := strconv.Itoa(devices[0].ID)
	modelPortID := strconv.Itoa(modelPorts[0].ID)

	// Add device port with empty mac_address
	if err := repo.AddDevicePort(deviceID, modelPortID, ""); err != nil {
		t.Errorf("failed to add device port: %v", err)
	}

	// Retrieve and verify
	devPorts, _ := repo.GetDevicePorts()
	found := false
	for _, dp := range devPorts {
		if dp.DeviceID == devices[0].ID && dp.ModelID == modelPorts[0].ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf(
			"expected device port with device ID %d and model port ID %d in list, got %+v",
			devices[0].ID,
			modelPorts[0].ID,
			devPorts,
		)
	}

	// --- Test delete by deviceID and modelPortID ---
	if err := repo.DeleteDevicePort(deviceID, modelPortID); err != nil {
		t.Errorf(
			"failed to delete device port with deviceID=%s and modelPortID=%s: %v",
			deviceID,
			modelPortID,
			err,
		)
	}

	// Verify it is no longer present
	devPortsAfterDelete, _ := repo.GetDevicePorts()
	for _, dp := range devPortsAfterDelete {
		if dp.DeviceID == devices[0].ID && dp.ModelID == modelPorts[0].ID {
			t.Errorf(
				"device port with device ID %d and model port ID %d should have been deleted, but got %+v",
				devices[0].ID,
				modelPorts[0].ID,
				devPortsAfterDelete,
			)
		}
	}
}

func TestDevicePort_AddWithMacAddress(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare all referenced data
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("C2960", "Cisco", "Switch"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("GigabitEthernet0/1", "0", "1", "C2960"); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddProprietary("Network Team"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Network"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("DataCenter", "", "", "Network Team", "Network"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddDevice("CoreSwitch", "C2960", "", "DataCenter", "Network Team"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and modelport IDs
	devices, _ := repo.GetDevices()
	modelPorts, _ := repo.GetModelPorts()
	if len(devices) == 0 || len(modelPorts) == 0 {
		t.Fatalf("expected at least one device and one model port")
	}
	deviceID := strconv.Itoa(devices[0].ID)
	modelPortID := strconv.Itoa(modelPorts[0].ID)

	// Test with valid MAC address
	expectedMac := "AA:BB:CC:DD:EE:FF"
	if err := repo.AddDevicePort(deviceID, modelPortID, expectedMac); err != nil {
		t.Errorf("failed to add device port with MAC address: %v", err)
	}

	// Retrieve and verify MAC address is stored correctly
	devPorts, _ := repo.GetDevicePorts()
	found := false
	for _, dp := range devPorts {
		if dp.DeviceID == devices[0].ID && dp.ModelID == modelPorts[0].ID {
			if dp.MacAddress != expectedMac {
				t.Errorf("expected MAC address %s, got %s", expectedMac, dp.MacAddress)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("device port with MAC address not found")
	}
}

func TestDevicePort_AddWithEmptyMacAddress(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare all referenced data
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Firewall"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("SRX300", "Juniper", "Firewall"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("ge-0/0/0", "0", "0", "SRX300"); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddProprietary("Security Team"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("DMZ"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("Perimeter", "", "", "Security Team", "DMZ"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddDevice("EdgeFirewall", "SRX300", "", "Perimeter", "Security Team"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and modelport IDs
	devices, _ := repo.GetDevices()
	modelPorts, _ := repo.GetModelPorts()
	if len(devices) == 0 || len(modelPorts) == 0 {
		t.Fatalf("expected at least one device and one model port")
	}
	deviceID := strconv.Itoa(devices[0].ID)
	modelPortID := strconv.Itoa(modelPorts[0].ID)

	// Test with empty MAC address
	if err := repo.AddDevicePort(deviceID, modelPortID, ""); err != nil {
		t.Errorf("failed to add device port with empty MAC address: %v", err)
	}

	// Retrieve and verify MAC address is empty
	devPorts, _ := repo.GetDevicePorts()
	found := false
	for _, dp := range devPorts {
		if dp.DeviceID == devices[0].ID && dp.ModelID == modelPorts[0].ID {
			if dp.MacAddress != "" {
				t.Errorf("expected empty MAC address, got %s", dp.MacAddress)
			}
			found = true
			break
		}
	}
	if !found {
		t.Errorf("device port with empty MAC address not found")
	}
}

func TestDevicePort_MultipleMacAddresses(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare all referenced data
	if err := repo.AddBrand("HP"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Server"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("ProLiant", "HP", "Server"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("eth0", "0", "0", "ProLiant"); err != nil {
		t.Fatalf("failed to add model port eth0: %v", err)
	}
	if err := repo.AddModelPort("eth1", "1", "0", "ProLiant"); err != nil {
		t.Fatalf("failed to add model port eth1: %v", err)
	}
	if err := repo.AddProprietary("IT Operations"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Production"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("ServerRoom", "", "", "IT Operations", "Production"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddDevice("WebServer01", "ProLiant", "", "ServerRoom", "IT Operations"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and modelport IDs
	devices, _ := repo.GetDevices()
	modelPorts, _ := repo.GetModelPorts()
	if len(devices) == 0 || len(modelPorts) < 2 {
		t.Fatalf("expected at least one device and two model ports")
	}
	deviceID := strconv.Itoa(devices[0].ID)

	// Add device ports with different MAC addresses
	mac1 := "11:22:33:44:55:66"
	mac2 := "AA:BB:CC:DD:EE:FF"
	
	if err := repo.AddDevicePort(deviceID, strconv.Itoa(modelPorts[0].ID), mac1); err != nil {
		t.Errorf("failed to add first device port: %v", err)
	}
	if err := repo.AddDevicePort(deviceID, strconv.Itoa(modelPorts[1].ID), mac2); err != nil {
		t.Errorf("failed to add second device port: %v", err)
	}

	// Retrieve and verify both MAC addresses
	devPorts, _ := repo.GetDevicePorts()
	foundPorts := 0
	for _, dp := range devPorts {
		if dp.DeviceID == devices[0].ID {
			if dp.ModelID == modelPorts[0].ID && dp.MacAddress != mac1 {
				t.Errorf("expected MAC address %s for first port, got %s", mac1, dp.MacAddress)
			}
			if dp.ModelID == modelPorts[1].ID && dp.MacAddress != mac2 {
				t.Errorf("expected MAC address %s for second port, got %s", mac2, dp.MacAddress)
			}
			foundPorts++
		}
	}
	if foundPorts != 2 {
		t.Errorf("expected 2 device ports, found %d", foundPorts)
	}
}
