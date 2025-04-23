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
	devices := repo.GetDevices()
	modelPorts := repo.GetModelPorts()
	if len(devices) == 0 || len(modelPorts) == 0 {
		t.Fatalf("expected at least one device and one model port")
	}
	deviceID := strconv.Itoa(devices[0].Id)
	modelPortID := strconv.Itoa(modelPorts[0].Id)

	// Add device port
	if err := repo.AddDevicePort(deviceID, modelPortID); err != nil {
		t.Errorf("failed to add device port: %v", err)
	}

	// Retrieve and verify
	devPorts := repo.GetDevicePorts()
	found := false
	for _, dp := range devPorts {
		if dp.DeviceId == devices[0].Id && dp.ModelId == modelPorts[0].Id {
			found = true
			break
		}
	}
	if !found {
		t.Errorf(
			"expected device port with device ID %d and model port ID %d in list, got %+v",
			devices[0].Id,
			modelPorts[0].Id,
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
