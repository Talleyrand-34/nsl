package sqlc_sqlite

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Helper to check if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func setupTestRepository(t *testing.T) (SQLiteRepository, error) {
	// Use in-memory DB for tests, or a temp file
	db, err := sql.Open("sqlite3", ":memory:")
	// db, err := sql.Open("sqlite3", "/tmp/test.db")
	// db, err := sql.Open("sqlite3", "file:memdb1?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	repo, err := NewSQLiteRepositoryFromDB(db)
	// repo, err := NewSQLiteRepository("/tmp/test.db")
	if err != nil {
		return SQLiteRepository{}, err
	}

	return repo, nil
}

// --- Cross-table isolation ---

func TestBrandAndDeviceClass_Isolation(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	brand := "Fortinet"
	class := "Router"

	if err := repo.AddBrand(brand); err != nil {
		t.Errorf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass(class); err != nil {
		t.Errorf("failed to add device class: %v", err)
	}

	brands := repo.GetBrands()
	classes := repo.GetDeviceClasses()

	if contains(brands, class) {
		t.Errorf("device class name %q appeared in brands list: %v", class, brands)
	}
	if contains(classes, brand) {
		t.Errorf("brand name %q appeared in device classes list: %v", brand, classes)
	}
}

// --- Generic --- //

// Test adding and retrieving device classes
func TestSQLiteRepository_AddAndGetDeviceClasses(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	deviceClassName := "TestDeviceClass"

	// Add a device class
	if err := repo.AddDeviceClass(deviceClassName); err != nil {
		t.Errorf("failed to add device class: %v", err)
	}

	// Retrieve device classes
	deviceClasses := repo.GetDeviceClasses()
	found := false
	for _, dc := range deviceClasses {
		if dc == deviceClassName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf(
			"device class %q not found in device classes list: %v",
			deviceClassName,
			deviceClasses,
		)
	}
}

// Test duplicate AddDeviceClass
func TestSQLiteRepository_AddDeviceClass_Duplicate(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	deviceClassName := "TestDeviceClass"

	// Add once
	if err := repo.AddDeviceClass(deviceClassName); err != nil {
		t.Errorf("failed to add device class: %v", err)
	}
	// Add again, should error
	err = repo.AddDeviceClass(deviceClassName)
	if err == nil {
		t.Errorf("expected error when adding duplicate device class, got nil")
	}
}
