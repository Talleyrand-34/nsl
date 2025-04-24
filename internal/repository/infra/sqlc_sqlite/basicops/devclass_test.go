package basicops

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

	e "nsl-graph/internal/repository/entities"
)

// --- DeviceClass Tests ---

func TestDeviceClass_AddAndGetTypicalClasses(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	deviceClasses := []string{"Router", "Switch", "Endpoint", "AP"}
	for _, c := range deviceClasses {
		if err := repo.AddDeviceClass(c); err != nil {
			t.Errorf("failed to add device class %q: %v", c, err)
		}
	}

	got := repo.GetDeviceClasses()
	for _, want := range deviceClasses {
		if !devClassSliceContains(got, want) {
			t.Errorf("expected device class %q in list, got %v", want, got)
		}
	}
}

func TestDeviceClass_AddDuplicateClass(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	class := "Router"
	if err := repo.AddDeviceClass(class); err != nil {
		t.Errorf("failed to add device class: %v", err)
	}
	if err := repo.AddDeviceClass(class); err == nil {
		t.Errorf("expected error when adding duplicate device class, got nil")
	}
}

func TestDeviceClass_CaseSensitivity(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	c1 := "ap"
	c2 := "AP"
	if err := repo.AddDeviceClass(c1); err != nil {
		t.Errorf("failed to add device class: %v", err)
	}
	if err := repo.AddDeviceClass(c2); err != nil {
		t.Errorf("failed to add device class with different case: %v", err)
	}
	got := repo.GetDeviceClasses()
	if !devClassSliceContains(got, c1) || !devClassSliceContains(got, c2) {
		t.Errorf("expected both %q and %q in list, got %v", c1, c2, got)
	}
}

// Helper function to check if a brand name exists in a slice of e.Brand
func devClassSliceContains(brands []e.DevClass, name string) bool {
	for _, b := range brands {
		if b.Name == name {
			return true
		}
	}
	return false
}
