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

func TestModelDevice_AddGetAndDeleteModels(t *testing.T) {
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

	// --- Test delete by ID ---
	// Find the ID of the model to delete (e.g., delete "F200")
	var deleteModelID string
	for _, m := range got {
		if m.Model == "F200" && m.Brand == "Fortinet" && m.Class == "Router" {
			deleteModelID = m.ID // assuming m.ID is a string, otherwise use strconv.Itoa(m.ID)
			break
		}
	}
	if deleteModelID == "" {
		t.Fatalf("could not find model ID for F200")
	}
	if err := repo.DeleteModel(deleteModelID); err != nil {
		t.Errorf("failed to delete model with id %q: %v", deleteModelID, err)
	}

	// Check that the model was deleted
	gotAfterDelete, _ := repo.GetModels()
	for _, m := range gotAfterDelete {
		if m.ID == deleteModelID {
			t.Errorf(
				"model with id %q should have been deleted, but got %+v",
				deleteModelID,
				gotAfterDelete,
			)
		}
	}
}
