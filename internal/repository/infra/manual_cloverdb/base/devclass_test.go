package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- DeviceClass Tests --- //

func TestDevClass_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	devClasses := []string{"Switch", "Router", "Firewall"}
	for _, dc := range devClasses {
		if err := repo.AddDeviceClass(dc); err != nil {
			t.Errorf("failed to add device class %q: %v", dc, err)
		}
	}

	got, _ := repo.GetDeviceClasses()
	for _, want := range devClasses {
		if !devClassSliceContains(got, want) {
			t.Errorf("expected device class %q in list, got %v", want, got)
		}
	}
}

func TestDevClass_AddDuplicate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	devClass := "Switch"
	if err := repo.AddDeviceClass(devClass); err != nil {
		t.Errorf("failed to add device class: %v", err)
	}
	if err := repo.AddDeviceClass(devClass); err == nil {
		t.Errorf("expected error when adding duplicate device class, got nil")
	}
}

func TestDevClass_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Test CREATE operation
	originalName := "Switch"
	if err := repo.AddDeviceClass(originalName); err != nil {
		t.Errorf("failed to add device class %q: %v", originalName, err)
	}

	// Get the device class ID
	devClasses, err := repo.GetDeviceClasses()
	if err != nil {
		t.Fatalf("failed to get device classes: %v", err)
	}

	var devClassId string
	for _, dc := range devClasses {
		if dc.Name == originalName {
			devClassId = dc.ID
			break
		}
	}

	if devClassId == "" {
		t.Fatalf("failed to find device class ID for %q", originalName)
	}

	// Test UPDATE operation
	updatedName := "Managed Switch"
	if err := repo.UpdateDeviceClass(devClassId, updatedName); err != nil {
		t.Errorf("failed to update device class: %v", err)
	}

	// Verify update was successful
	devClassesAfterUpdate, err := repo.GetDeviceClasses()
	if err != nil {
		t.Errorf("failed to get device classes after update: %v", err)
	}

	if !devClassSliceContains(devClassesAfterUpdate, updatedName) {
		t.Errorf("expected updated device class %q in list, got %v", updatedName, devClassesAfterUpdate)
	}

	if devClassSliceContains(devClassesAfterUpdate, originalName) {
		t.Errorf("original device class %q should not exist after update, got %v", originalName, devClassesAfterUpdate)
	}
}

func TestDevClass_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Test CREATE operation
	devClassName := "Router"
	if err := repo.AddDeviceClass(devClassName); err != nil {
		t.Errorf("failed to add device class %q: %v", devClassName, err)
	}

	// Verify device class was created
	devClasses, err := repo.GetDeviceClasses()
	if err != nil {
		t.Errorf("failed to get device classes: %v", err)
	}

	if !devClassSliceContains(devClasses, devClassName) {
		t.Errorf("expected device class %q in list before deletion, got %v", devClassName, devClasses)
	}

	// Test DELETE operation
	if err := repo.DeleteDeviceClass(devClassName); err != nil {
		t.Errorf("failed to delete device class %q: %v", devClassName, err)
	}

	// Verify device class was deleted
	devClassesAfterDelete, err := repo.GetDeviceClasses()
	if err != nil {
		t.Errorf("failed to get device classes after deletion: %v", err)
	}

	if devClassSliceContains(devClassesAfterDelete, devClassName) {
		t.Errorf("device class %q should have been deleted, but got %v", devClassName, devClassesAfterDelete)
	}
}

// Helper function to check if a device class name exists in a slice of e.DevClass
func devClassSliceContains(devClasses []e.DevClass, name string) bool {
	for _, dc := range devClasses {
		if dc.Name == name {
			return true
		}
	}
	return false
}
