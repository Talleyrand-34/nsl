package basicops

import (
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
	zones := repo.GetZones()
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
	devices := repo.GetDevices()
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
