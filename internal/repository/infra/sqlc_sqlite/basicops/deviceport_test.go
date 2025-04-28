
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

	// Add device port
	if err := repo.AddDevicePort(deviceID, modelPortID); err != nil {
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
	err = repo.AddDevicePort("notanumber", "1")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid device ID, got error: %v", err)
	}

	// Invalid model port ID
	err = repo.AddDevicePort("1", "notanumber")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid model port ID, got error: %v", err)
	}
}
