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
	"context"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestModelDevice_AddAndGetModels(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare referenced Brand and DeviceClass
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddDeviceClass("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}

	// Add models
	models := []struct {
		modelName string
		brandName string
		className string
	}{
		{"F100", "Fortinet", "Router"},
		{"F200", "Fortinet", "Router"},
		{"Cisco5000", "Cisco", "Switch"},
	}
	for _, m := range models {
		if err := repo.AddModel(m.modelName, m.brandName, m.className); err != nil {
			t.Errorf("failed to add model %q: %v", m.modelName, err)
		}
	}

	// Retrieve and verify
	got, _ := repo.GetModels()
	for _, want := range models {
		found := false
		for _, m := range got {
			if m.Model == want.modelName && m.Brand == want.brandName && m.Class == want.className {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(
				"expected model %q/%q/%q in list, got %+v",
				want.modelName,
				want.brandName,
				want.className,
				got,
			)
		}
	}
}

func TestModelDevice_AddModel_InvalidReferences(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare one valid brand/class
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}

	// Invalid brand
	err = repo.AddModel("F300", "NonExistentBrand", "Router")
	if err == nil {
		t.Errorf("expected error for non-existent brand, got nil")
	}

	// Invalid class
	err = repo.AddModel("F400", "Fortinet", "NonExistentClass")
	if err == nil {
		t.Errorf("expected error for non-existent device class, got nil")
	}
}

func TestModelDevice_CreateOnly(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Create dependencies: Brand and DeviceClass
	if err := repo.AddBrand("Cisco"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}

	// Test CREATE operation
	modelName := "ISR4431"
	brandName := "Cisco"
	className := "Router"
	
	if err := repo.AddModel(modelName, brandName, className); err != nil {
		t.Errorf("failed to add model %q: %v", modelName, err)
	}

	// Verify model was created
	models, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models: %v", err)
	}

	found := false
	for _, m := range models {
		if m.Model == modelName && m.Brand == brandName && m.Class == className {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected model %q/%q/%q in list, got %+v", modelName, brandName, className, models)
	}
}

func TestModelDevice_CreateAndUpdate(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Create dependencies: Brands and DeviceClasses
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand Juniper: %v", err)
	}
	if err := repo.AddBrand("Juniper Networks"); err != nil {
		t.Fatalf("failed to add brand Juniper Networks: %v", err)
	}
	if err := repo.AddDeviceClass("Switch"); err != nil {
		t.Fatalf("failed to add device class Switch: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class Router: %v", err)
	}

	// Test CREATE operation
	originalModelName := "EX4300"
	brandName := "Juniper"
	className := "Switch"
	
	if err := repo.AddModel(originalModelName, brandName, className); err != nil {
		t.Errorf("failed to add model %q: %v", originalModelName, err)
	}

	// Get the model ID
	models, err := repo.GetModels()
	if err != nil {
		t.Fatalf("failed to get models: %v", err)
	}

	var modelId string
	for _, m := range models {
		if m.Model == originalModelName && m.Brand == brandName && m.Class == className {
			modelId = m.ID
			break
		}
	}
	if modelId == "" {
		t.Fatalf("could not find model ID for %s", originalModelName)
	}

	// Get IDs for new brand and class
	newBrandId, err := repo.query.GetBrandId(context.Background(), "Juniper Networks")
	if err != nil {
		t.Fatalf("failed to get new brand ID: %v", err)
	}
	
	newClassId, err := repo.query.GetClassId(context.Background(), "Router")
	if err != nil {
		t.Fatalf("failed to get new class ID: %v", err)
	}

	// Test UPDATE operation
	updatedModelName := "EX4300-48P"
	if err := repo.UpdateModel(modelId, updatedModelName, fmt.Sprintf("%d", newBrandId), fmt.Sprintf("%d", newClassId)); err != nil {
		t.Errorf("failed to update model: %v", err)
	}

	// Verify update was successful
	modelsAfterUpdate, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models after update: %v", err)
	}

	foundUpdated := false
	foundOriginal := false
	for _, m := range modelsAfterUpdate {
		if m.Model == updatedModelName && m.Brand == "Juniper Networks" && m.Class == "Router" {
			foundUpdated = true
		}
		if m.Model == originalModelName && m.Brand == brandName && m.Class == className {
			foundOriginal = true
		}
	}

	if !foundUpdated {
		t.Errorf("expected updated model %q/Juniper Networks/Router in list, got %+v", updatedModelName, modelsAfterUpdate)
	}
	if foundOriginal {
		t.Errorf("original model %q should not exist after update, got %+v", originalModelName, modelsAfterUpdate)
	}
}

func TestModelDevice_CreateAndDelete(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Create dependencies: Brand and DeviceClass
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Firewall"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}

	// Test CREATE operation
	modelName := "FortiGate-100F"
	brandName := "Fortinet"
	className := "Firewall"
	
	if err := repo.AddModel(modelName, brandName, className); err != nil {
		t.Errorf("failed to add model %q: %v", modelName, err)
	}

	// Verify model was created
	models, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models: %v", err)
	}

	var modelId string
	found := false
	for _, m := range models {
		if m.Model == modelName && m.Brand == brandName && m.Class == className {
			found = true
			modelId = m.ID
			break
		}
	}
	if !found {
		t.Errorf("expected model %q/%q/%q in list before deletion, got %+v", modelName, brandName, className, models)
	}

	// Test DELETE operation
	if err := repo.DeleteModel(modelId); err != nil {
		t.Errorf("failed to delete model with ID %q: %v", modelId, err)
	}

	// Verify model was deleted
	modelsAfterDelete, err := repo.GetModels()
	if err != nil {
		t.Errorf("failed to get models after deletion: %v", err)
	}

	for _, m := range modelsAfterDelete {
		if m.ID == modelId {
			t.Errorf("model with ID %q should have been deleted, but got %+v", modelId, modelsAfterDelete)
		}
	}
}
