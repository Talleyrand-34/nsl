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

func TestConnection_AddAndGetConnections(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare all referenced data
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
	if err := repo.AddModelPort("eth1", "15", "25", "F100"); err != nil {
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
	if err := repo.AddDevice("BackupRouter", "F100", "", "HQ", "IT Department"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and modelport IDs
	devices, _ := repo.GetDevices()
	modelPorts, _ := repo.GetModelPorts()
	if len(devices) < 2 || len(modelPorts) < 2 {
		t.Fatalf("expected at least two devices and two model ports")
	}
	deviceID1 := strconv.Itoa(devices[0].ID)
	deviceID2 := strconv.Itoa(devices[1].ID)
	modelPortID1 := strconv.Itoa(modelPorts[0].ID)
	modelPortID2 := strconv.Itoa(modelPorts[1].ID)

	// Add device ports for both devices
	if err := repo.AddDevicePort(deviceID1, modelPortID1); err != nil {
		t.Fatalf("failed to add device port 1: %v", err)
	}
	if err := repo.AddDevicePort(deviceID2, modelPortID2); err != nil {
		t.Fatalf("failed to add device port 2: %v", err)
	}

	// Add connection
	if err := repo.AddConnection(deviceID1, modelPortID1, deviceID2, modelPortID2); err != nil {
		t.Errorf("failed to add connection: %v", err)
	}

	// Retrieve and verify
	connections, _ := repo.GetConnections()
	found := false
	for _, c := range connections {
		if c.FromDevice == devices[0].Name && c.FromModelPort == modelPorts[0].Name &&
			c.ToDevice == devices[1].Name && c.ToModelPort == modelPorts[1].Name {
			found = true
			break
		}
	}

	if !found {
		t.Errorf(
			"expected connection from device %q port %q to device %q port %q, got %+v",
			devices[0].Name, modelPorts[0].Name, devices[1].Name, modelPorts[1].Name, connections,
		)
	}
}

func TestConnection_AddConnection_InvalidIDs(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Invalid device ID
	err = repo.AddConnection("notanumber", "1", "2", "3")
	if err == nil {
		t.Errorf("Expected error invalid number")
	}

	// Invalid model port ID
	err = repo.AddConnection("1", "notanumber", "2", "3")
	if err == nil {
		t.Errorf("Expected error invalid number")
	}

	// Invalid toDevice ID
	err = repo.AddConnection("1", "2", "notanumber", "3")
	if err == nil {
		t.Errorf("Expected error invalid number")
	}

	// Invalid toModelPort ID
	err = repo.AddConnection("1", "2", "3", "notanumber")
	if err == nil {
		t.Errorf("Expected error invalid number")
	}
}

func TestConnection_AddGetAndDeleteConnections(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare all referenced data
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
	if err := repo.AddModelPort("eth1", "15", "25", "F100"); err != nil {
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
	if err := repo.AddDevice("BackupRouter", "F100", "", "HQ", "IT Department"); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and modelport IDs
	devices, _ := repo.GetDevices()
	modelPorts, _ := repo.GetModelPorts()
	if len(devices) < 2 || len(modelPorts) < 2 {
		t.Fatalf("expected at least two devices and two model ports")
	}
	deviceID1 := strconv.Itoa(devices[0].ID)
	deviceID2 := strconv.Itoa(devices[1].ID)
	modelPortID1 := strconv.Itoa(modelPorts[0].ID)
	modelPortID2 := strconv.Itoa(modelPorts[1].ID)

	// Add device ports for both devices
	if err := repo.AddDevicePort(deviceID1, modelPortID1); err != nil {
		t.Fatalf("failed to add device port 1: %v", err)
	}
	if err := repo.AddDevicePort(deviceID2, modelPortID2); err != nil {
		t.Fatalf("failed to add device port 2: %v", err)
	}

	// Add connection
	if err := repo.AddConnection(deviceID1, modelPortID1, deviceID2, modelPortID2); err != nil {
		t.Errorf("failed to add connection: %v", err)
	}

	// Retrieve and verify
	connections, _ := repo.GetConnections()
	var deleteConnID string
	found := false
	for _, c := range connections {
		if c.FromDevice == devices[0].Name && c.FromModelPort == modelPorts[0].Name &&
			c.ToDevice == devices[1].Name && c.ToModelPort == modelPorts[1].Name {
			deleteConnID = strconv.Itoa(c.ID) // Save the connection ID for deletion
			found = true
			break
		}
	}

	if !found {
		t.Fatalf(
			"expected connection from device %q port %q to device %q port %q, got %+v",
			devices[0].Name, modelPorts[0].Name, devices[1].Name, modelPorts[1].Name, connections,
		)
	}

	// --- Test delete by connection ID ---
	if err := repo.DeleteConnection(deleteConnID); err != nil {
		t.Errorf("failed to delete connection with id %q: %v", deleteConnID, err)
	}

	// Check that the connection was deleted
	connectionsAfterDelete, _ := repo.GetConnections()
	for _, c := range connectionsAfterDelete {
		if strconv.Itoa(c.ID) == deleteConnID {
			t.Errorf(
				"connection with id %q should have been deleted, but got %+v",
				deleteConnID,
				connectionsAfterDelete,
			)
		}
	}
}
