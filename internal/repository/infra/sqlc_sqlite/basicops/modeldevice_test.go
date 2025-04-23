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
	got := repo.GetModels()
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
