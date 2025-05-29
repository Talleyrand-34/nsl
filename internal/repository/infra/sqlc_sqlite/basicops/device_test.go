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
	"context"
	"fmt"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestDevice_AddAndGetDevices(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare referenced data: Proprietary, Brand, DeviceClass, Model, ZoneType, Zone
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "IT Department", "Physical"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddModel("F100", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Get IDs/names for device creation
	zones, _ := repo.GetZones()
	var zoneId string
	for _, z := range zones {
		if z.Name == "HQ" {
			zoneId = strconv.Itoa(z.ID)
			break
		}
	}
	if zoneId == "" {
		t.Fatalf("could not find HQ zone id")
	}

	// Add device by zoneId
	if err := repo.AddDevice("MainRouter", "F100", zoneId, "", "IT Department"); err != nil {
		t.Errorf("failed to add device: %v", err)
	}

	// Add device by zoneName
	if err := repo.AddDevice("BackupRouter", "F100", "", "HQ", "IT Department"); err != nil {
		t.Errorf("failed to add device by zone name: %v", err)
	}

	// Retrieve and verify
	devices, _ := repo.GetDevices()
	expectedDevices := []struct {
		Label       string
		Model       string
		Brand       string
		ZoneName    string
		Proprietary string
	}{
		{"MainRouter", "F100", "Fortinet", "HQ", "IT Department"},
		{"BackupRouter", "F100", "Fortinet", "HQ", "IT Department"},
	}
	for _, want := range expectedDevices {
		found := false
		for _, d := range devices {
			if d.Name == want.Label {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected device %+v in list, got %+v", want, devices)
		}
	}
}

func TestDevice_CreateOnly(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Create dependencies: Proprietary, Brand, DeviceClass, Model, ZoneType, Zone
	if err := repo.AddProprietary("Engineering"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("DataCenter", "", "", "Engineering", "Physical"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddModel("ISR4431", "Cisco", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Test CREATE operation
	deviceLabel := "Core-Router-01"
	modelName := "ISR4431"
	zoneName := "DataCenter"
	proprietaryName := "Engineering"

	if err := repo.AddDevice(deviceLabel, modelName, "", zoneName, proprietaryName); err != nil {
		t.Errorf("failed to add device %q: %v", deviceLabel, err)
	}

	// Verify device was created
	devices, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices: %v", err)
	}

	found := false
	for _, d := range devices {
		if d.Name == deviceLabel && d.Model == modelName && d.ZoneName == zoneName && d.Proprietary == proprietaryName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected device %q with model %q in zone %q, got %+v", deviceLabel, modelName, zoneName, devices)
	}
}

func TestDevice_CreateAndUpdate(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Create dependencies for original device
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary IT Department: %v", err)
	}
	if err := repo.AddProprietary("Network Operations"); err != nil {
		t.Fatalf("failed to add proprietary Network Operations: %v", err)
	}
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand Juniper: %v", err)
	}
	if err := repo.AddBrand("Arista"); err != nil {
		t.Fatalf("failed to add brand Arista: %v", err)
	}
	if err := repo.AddDeviceClass("Switch"); err != nil {
		t.Fatalf("failed to add device class Switch: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class Router: %v", err)
	}
	if err := repo.AddZoneType("Logical"); err != nil {
		t.Fatalf("failed to add zone type Logical: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type Physical: %v", err)
	}
	if err := repo.AddZone("LAN", "", "", "IT Department", "Logical"); err != nil {
		t.Fatalf("failed to add zone LAN: %v", err)
	}
	if err := repo.AddZone("WAN", "", "", "Network Operations", "Physical"); err != nil {
		t.Fatalf("failed to add zone WAN: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch"); err != nil {
		t.Fatalf("failed to add model EX4300: %v", err)
	}
	if err := repo.AddModel("7280R", "Arista", "Router"); err != nil {
		t.Fatalf("failed to add model 7280R: %v", err)
	}

	// Test CREATE operation
	originalDeviceLabel := "Switch-01"
	originalModelName := "EX4300"
	originalZoneName := "LAN"
	originalProprietaryName := "IT Department"

	if err := repo.AddDevice(originalDeviceLabel, originalModelName, "", originalZoneName, originalProprietaryName); err != nil {
		t.Errorf("failed to add device %q: %v", originalDeviceLabel, err)
	}

	// Get the device ID
	devices, err := repo.GetDevices()
	if err != nil {
		t.Fatalf("failed to get devices: %v", err)
	}

	var deviceId string
	for _, d := range devices {
		if d.Name == originalDeviceLabel {
			deviceId = strconv.Itoa(d.ID)
			break
		}
	}
	if deviceId == "" {
		t.Fatalf("could not find device ID for %s", originalDeviceLabel)
	}

	// Get IDs for the update
	newModelId, err := repo.query.GetModelId(context.Background(), "7280R")
	if err != nil {
		t.Fatalf("failed to get new model ID: %v", err)
	}

	newZoneId, err := repo.query.GetZoneId(context.Background(), "WAN")
	if err != nil {
		t.Fatalf("failed to get new zone ID: %v", err)
	}

	newProprietaryId, err := repo.query.GetProprietary(context.Background(), "Network Operations")
	if err != nil {
		t.Fatalf("failed to get new proprietary ID: %v", err)
	}

	// Test UPDATE operation
	updatedDeviceLabel := "Core-Router-01"
	if err := repo.UpdateDevice(deviceId, updatedDeviceLabel, fmt.Sprintf("%d", newModelId), fmt.Sprintf("%d", newZoneId), fmt.Sprintf("%d", newProprietaryId)); err != nil {
		t.Errorf("failed to update device: %v", err)
	}

	// Verify update was successful
	devicesAfterUpdate, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices after update: %v", err)
	}

	foundUpdated := false
	foundOriginal := false
	for _, d := range devicesAfterUpdate {
		if d.Name == updatedDeviceLabel && d.Model == "7280R" && d.ZoneName == "WAN" && d.Proprietary == "Network Operations" {
			foundUpdated = true
		}
		if d.Name == originalDeviceLabel && d.Model == originalModelName {
			foundOriginal = true
		}
	}

	if !foundUpdated {
		t.Errorf("expected updated device %q with model 7280R in WAN zone, got %+v", updatedDeviceLabel, devicesAfterUpdate)
	}
	if foundOriginal {
		t.Errorf("original device %q should not exist after update, got %+v", originalDeviceLabel, devicesAfterUpdate)
	}
}

func TestDevice_CreateAndDelete(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Create dependencies: Proprietary, Brand, DeviceClass, Model, ZoneType, Zone
	if err := repo.AddProprietary("Security Team"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Firewall"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddZoneType("Security"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("DMZ", "", "", "Security Team", "Security"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	if err := repo.AddModel("FortiGate-100F", "Fortinet", "Firewall"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Test CREATE operation
	deviceLabel := "Firewall-DMZ-01"
	modelName := "FortiGate-100F"
	zoneName := "DMZ"
	proprietaryName := "Security Team"

	if err := repo.AddDevice(deviceLabel, modelName, "", zoneName, proprietaryName); err != nil {
		t.Errorf("failed to add device %q: %v", deviceLabel, err)
	}

	// Verify device was created and get its ID
	devices, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices: %v", err)
	}

	var deviceId string
	found := false
	for _, d := range devices {
		if d.Name == deviceLabel && d.Model == modelName && d.ZoneName == zoneName && d.Proprietary == proprietaryName {
			found = true
			deviceId = strconv.Itoa(d.ID)
			break
		}
	}
	if !found {
		t.Errorf("expected device %q with model %q in zone %q before deletion, got %+v", deviceLabel, modelName, zoneName, devices)
	}

	// Test DELETE operation
	if err := repo.DeleteDevice(deviceId); err != nil {
		t.Errorf("failed to delete device with ID %q: %v", deviceId, err)
	}

	// Verify device was deleted
	devicesAfterDelete, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices after deletion: %v", err)
	}

	for _, d := range devicesAfterDelete {
		if strconv.Itoa(d.ID) == deviceId {
			t.Errorf("device with ID %q should have been deleted, but got %+v", deviceId, devicesAfterDelete)
		}
	}
}

// func TestDevice_AddDevice_InvalidReferences(t *testing.T) {
// 	repo, err := setupTestRepository(t)
// 	if err != nil {
// 		t.Fatalf("failed to setup repository: %v", err)
// 	}
// 	defer repo.Close()
//
// 	// Prepare valid model and zone
// 	if err := repo.AddProprietary("IT Department"); err != nil {
// 		t.Fatalf("failed to add proprietary: %v", err)
// 	}
// 	if err := repo.AddBrand("Fortinet"); err != nil {
// 		t.Fatalf("failed to add brand: %v", err)
// 	}
// 	if err := repo.AddDeviceClass("Router"); err != nil {
// 		t.Fatalf("failed to add device class: %v", err)
// 	}
// 	if err := repo.AddZoneType("Physical"); err != nil {
// 		t.Fatalf("failed to add zone type: %v", err)
// 	}
// 	if err := repo.AddZone("HQ", "", "", "IT Department", "Physical"); err != nil {
// 		t.Fatalf("failed to add zone: %v", err)
// 	}
// 	if err := repo.AddModel("F100", "Fortinet", "Router"); err != nil {
// 		t.Fatalf("failed to add model: %v", err)
// 	}
//
// 	// Invalid proprietary
// 	err = repo.AddDevice("BadOwnerDevice", "F100", "", "HQ", "NonExistentOwner")
// 	if err == nil {
// 		t.Errorf("expected error for non-existent proprietary, got nil")
// 	}
//
// 	// Invalid model
// 	err = repo.AddDevice("BadModelDevice", "NonExistentModel", "", "HQ", "IT Department")
// 	if err == nil {
// 		t.Errorf("expected error for non-existent model, got nil")
// 	}
//
// 	// Invalid zone
// 	err = repo.AddDevice("BadZoneDevice", "F100", "", "NonExistentZone", "IT Department")
// 	if err == nil {
// 		t.Errorf("expected error for non-existent zone, got nil")
// 	}
// }
