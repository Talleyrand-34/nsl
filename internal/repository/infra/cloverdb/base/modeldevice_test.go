package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- ModelDevice Tests --- //

func TestModelDevice_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}

	// Add models
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Errorf("failed to add model: %v", err)
	}

	models, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models: %v", err)
	}

	if !modelSliceContains(models, "EX4300") {
		t.Errorf("expected model 'EX4300' in list, got %v", models)
	}
}

func TestModelDevice_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModelType("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}

	// Add model
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Get model ID
	models, err := repo.GetModels()
	if err != nil {
		t.Fatalf("failed to get models: %v", err)
	}

	var modelId string
	for _, m := range models {
		if m.Model == "EX4300" {
			modelId = m.ID
			break
		}
	}

	if modelId == "" {
		t.Fatalf("failed to find model ID")
	}

	// Get brand and class IDs
	brands, _ := repo.GetBrands()
	var juniperID string
	for _, b := range brands {
		if b.Name == "Juniper" {
			juniperID = b.ID
			break
		}
	}

	modelTypes, _ := repo.GetModelTypes()
	var routerID string
	for _, dc := range modelTypes {
		if dc.Name == "Router" {
			routerID = dc.ID
			break
		}
	}

	// Update model
	if err := repo.UpdateModel(modelId, "EX4300", juniperID, routerID, ""); err != nil {
		t.Errorf("failed to update model: %v", err)
	}

	// Verify update
	modelsAfterUpdate, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models after update: %v", err)
	}

	if !modelSliceContains(modelsAfterUpdate, "EX4300") {
		t.Errorf("expected updated model 'EX4300' in list, got %v", modelsAfterUpdate)
	}
}

func TestModelDevice_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}

	// Add model
	if err := repo.AddModel("MX480", "Juniper", "Router", ""); err != nil {
		t.Errorf("failed to add model: %v", err)
	}

	// Get model ID
	models, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models: %v", err)
	}

	var modelId string
	for _, m := range models {
		if m.Model == "MX480" {
			modelId = m.ID
			break
		}
	}

	if modelId == "" {
		t.Fatalf("failed to find model ID")
	}

	// Delete model
	if err := repo.DeleteModel(modelId); err != nil {
		t.Errorf("failed to delete model: %v", err)
	}

	// Verify deletion
	modelsAfterDelete, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models after deletion: %v", err)
	}

	if modelSliceContains(modelsAfterDelete, "MX480") {
		t.Errorf("model 'MX480' should have been deleted, but got %v", modelsAfterDelete)
	}
}

func modelSliceContains(models []e.ModelDevice, modelName string) bool {
	for _, m := range models {
		if m.Model == modelName {
			return true
		}
	}
	return false
}
