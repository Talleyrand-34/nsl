package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- Device Tests --- //

func TestDevice_AddAndGet(t *testing.T) {
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
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Add device
	if err := repo.AddDevice("SW-01", "Catalyst 9300", "", "", "", false, false); err != nil {
		t.Errorf("failed to add device: %v", err)
	}

	devices, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices: %v", err)
	}

	if !deviceSliceContains(devices, "SW-01") {
		t.Errorf("expected device 'SW-01' in list, got %v", devices)
	}
}

func TestDevice_AddWithZone(t *testing.T) {
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
	if err := repo.AddModel("Catalyst 9300", "Cisco", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}
	if err := repo.AddZoneType("Room"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddOwner("Company A"); err != nil {
		t.Fatalf("failed to add owner: %v", err)
	}
	if err := repo.AddZone("Server Room", "", "", "Company A", "Room"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}

	// Add device with zone
	if err := repo.AddDevice("SW-02", "Catalyst 9300", "", "Server Room", "Company A", false, false); err != nil {
		t.Errorf("failed to add device: %v", err)
	}

	devices, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices: %v", err)
	}

	// Verify device exists and has correct zone
	var sw02 *e.Device
	for i := range devices {
		if devices[i].Label == "SW-02" {
			sw02 = &devices[i]
			break
		}
	}

	if sw02 == nil {
		t.Errorf("expected to find 'SW-02' device")
	} else if sw02.ZoneName != "Server Room" {
		t.Errorf("expected zone 'Server Room', got %q", sw02.ZoneName)
	}
}

func TestDevice_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("ISR 4000", "Cisco", "Router", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Add device
	if err := repo.AddDevice("RTR-01", "ISR 4000", "", "", "", false, false); err != nil {
		t.Errorf("failed to add device: %v", err)
	}

	// Get device ID
	devices, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices: %v", err)
	}

	var deviceId string
	for _, d := range devices {
		if d.Label == "RTR-01" {
			deviceId = d.ID
			break
		}
	}

	if deviceId == "" {
		t.Fatalf("failed to find device ID")
	}

	// Delete device
	if err := repo.DeleteDevice(deviceId); err != nil {
		t.Errorf("failed to delete device: %v", err)
	}

	// Verify deletion
	devicesAfterDelete, err := repo.GetDevices()
	if err != nil {
		t.Errorf("failed to get devices after deletion: %v", err)
	}

	if deviceSliceContains(devicesAfterDelete, "RTR-01") {
		t.Errorf("device 'RTR-01' should have been deleted, but got %v", devicesAfterDelete)
	}
}

func deviceSliceContains(devices []e.Device, deviceName string) bool {
	for _, d := range devices {
		if d.Label == deviceName {
			return true
		}
	}
	return false
}
