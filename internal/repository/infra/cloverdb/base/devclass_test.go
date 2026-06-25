package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- ModelType Tests --- //

func TestModelType_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	modelTypes := []string{"Switch", "Router", "Firewall"}
	for _, dc := range modelTypes {
		if err := repo.AddModelType(dc); err != nil {
			t.Errorf("failed to add device class %q: %v", dc, err)
		}
	}

	got, _ := repo.GetModelTypes()
	for _, want := range modelTypes {
		if !modelTypeSliceContains(got, want) {
			t.Errorf("expected device class %q in list, got %v", want, got)
		}
	}
}

func TestModelType_AddDuplicate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	modelType := "Switch"
	if err := repo.AddModelType(modelType); err != nil {
		t.Errorf("failed to add device class: %v", err)
	}
	if err := repo.AddModelType(modelType); err == nil {
		t.Errorf("expected error when adding duplicate device class, got nil")
	}
}

func TestModelType_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Test CREATE operation
	originalName := "Switch"
	if err := repo.AddModelType(originalName); err != nil {
		t.Errorf("failed to add device class %q: %v", originalName, err)
	}

	// Get the device class ID
	modelTypes, err := repo.GetModelTypes()
	if err != nil {
		t.Fatalf("failed to get device classes: %v", err)
	}

	var modelTypeId string
	for _, dc := range modelTypes {
		if dc.Name == originalName {
			modelTypeId = dc.ID
			break
		}
	}

	if modelTypeId == "" {
		t.Fatalf("failed to find device class ID for %q", originalName)
	}

	// Test UPDATE operation
	updatedName := "Managed Switch"
	if err := repo.UpdateModelType(modelTypeId, updatedName); err != nil {
		t.Errorf("failed to update device class: %v", err)
	}

	// Verify update was successful
	modelTypesAfterUpdate, err := repo.GetModelTypes()
	if err != nil {
		t.Errorf("failed to get device classes after update: %v", err)
	}

	if !modelTypeSliceContains(modelTypesAfterUpdate, updatedName) {
		t.Errorf("expected updated device class %q in list, got %v", updatedName, modelTypesAfterUpdate)
	}

	if modelTypeSliceContains(modelTypesAfterUpdate, originalName) {
		t.Errorf("original device class %q should not exist after update, got %v", originalName, modelTypesAfterUpdate)
	}
}

func TestModelType_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Test CREATE operation
	modelTypeName := "Router"
	if err := repo.AddModelType(modelTypeName); err != nil {
		t.Errorf("failed to add device class %q: %v", modelTypeName, err)
	}

	// Verify device class was created
	modelTypes, err := repo.GetModelTypes()
	if err != nil {
		t.Errorf("failed to get device classes: %v", err)
	}

	if !modelTypeSliceContains(modelTypes, modelTypeName) {
		t.Errorf("expected device class %q in list before deletion, got %v", modelTypeName, modelTypes)
	}

	// Test DELETE operation
	if err := repo.DeleteModelType(modelTypeName); err != nil {
		t.Errorf("failed to delete device class %q: %v", modelTypeName, err)
	}

	// Verify device class was deleted
	modelTypesAfterDelete, err := repo.GetModelTypes()
	if err != nil {
		t.Errorf("failed to get device classes after deletion: %v", err)
	}

	if modelTypeSliceContains(modelTypesAfterDelete, modelTypeName) {
		t.Errorf("device class %q should have been deleted, but got %v", modelTypeName, modelTypesAfterDelete)
	}
}

// Helper function to check if a device class name exists in a slice of e.ModelType
func modelTypeSliceContains(modelTypes []e.ModelType, name string) bool {
	for _, dc := range modelTypes {
		if dc.Name == name {
			return true
		}
	}
	return false
}
