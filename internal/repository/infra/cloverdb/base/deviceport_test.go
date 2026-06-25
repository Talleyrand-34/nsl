package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- DevicePort Tests --- //

func TestDevicePort_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and model port IDs
	devices, _ := repo.GetDevices()
	var deviceId string
	for _, d := range devices {
		if d.Name == "SW-01" {
			deviceId = d.ID
			break
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var modelPortId string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortId = mp.ID
			break
		}
	}

	// Add device port
	if _, err := repo.AddDevicePort(deviceId, modelPortId, "aa:bb:cc:dd:ee:ff", []e.PortVlanConfig{}); err != nil {
		t.Errorf("failed to add device port: %v", err)
	}

	// Get and verify
	devicePorts, err := repo.GetDevicePorts()
	if err != nil {
		t.Errorf("failed to get device ports: %v", err)
	}

	if len(devicePorts) == 0 {
		t.Errorf("expected at least one device port, got none")
	}
}

func TestDevicePort_AddWithInvalidModel(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites - two different models
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModel("Catalyst 3850", "Cisco", "Switch"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 3850", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and model port IDs
	devices, _ := repo.GetDevices()
	var deviceId string
	for _, d := range devices {
		if d.Name == "SW-01" {
			deviceId = d.ID
			break
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var modelPortId string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortId = mp.ID
			break
		}
	}

	// Try to add device port with wrong model - should fail
	if _, err := repo.AddDevicePort(deviceId, modelPortId, "", []e.PortVlanConfig{}); err == nil {
		t.Errorf("expected error when adding device port with wrong model, got nil")
	}
}

func TestDevicePort_Delete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "Catalyst 9300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Fatalf("failed to add device: %v", err)
	}

	// Get device and model port IDs
	devices, _ := repo.GetDevices()
	var deviceId string
	for _, d := range devices {
		if d.Name == "SW-01" {
			deviceId = d.ID
			break
		}
	}

	modelPorts, _ := repo.GetModelPorts()
	var modelPortId string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortId = mp.ID
			break
		}
	}

	// Add device port
	if _, err := repo.AddDevicePort(deviceId, modelPortId, "", []e.PortVlanConfig{}); err != nil {
		t.Fatalf("failed to add device port: %v", err)
	}

	// Delete device port
	if err := repo.DeleteDevicePort(deviceId, modelPortId); err != nil {
		t.Errorf("failed to delete device port: %v", err)
	}

	// Verify deletion
	devicePortsAfterDelete, err := repo.GetDevicePorts()
	if err != nil {
		t.Errorf("failed to get device ports after deletion: %v", err)
	}

	if len(devicePortsAfterDelete) > 0 {
		t.Errorf("device port should have been deleted, but got %d ports", len(devicePortsAfterDelete))
	}
}
