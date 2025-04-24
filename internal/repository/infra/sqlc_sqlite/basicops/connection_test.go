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
	devices := repo.GetDevices()
	modelPorts := repo.GetModelPorts()
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
	connections := repo.GetConnections()
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
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid fromDevice ID, got error: %v", err)
	}

	// Invalid model port ID
	err = repo.AddConnection("1", "notanumber", "2", "3")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid fromModelPort ID, got error: %v", err)
	}

	// Invalid toDevice ID
	err = repo.AddConnection("1", "2", "notanumber", "3")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid toDevice ID, got error: %v", err)
	}

	// Invalid toModelPort ID
	err = repo.AddConnection("1", "2", "3", "notanumber")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid toModelPort ID, got error: %v", err)
	}
}
