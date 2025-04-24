package basicops

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"

	e "nsl-graph/internal/repository/entities"
)

// --- Brand Tests --- //

func TestBrand_AddAndGetIndustrialBrands(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	brands := []string{"Fortinet", "Siemens", "Cisco"}
	for _, b := range brands {
		if err := repo.AddBrand(b); err != nil {
			t.Errorf("failed to add brand %q: %v", b, err)
		}
	}

	got := repo.GetBrands()
	for _, want := range brands {
		if !brandSliceContains(got, want) {
			t.Errorf("expected brand %q in list, got %v", want, got)
		}
	}
}

func TestBrand_AddDuplicateIndustrialBrand(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	brand := "Fortinet"
	if err := repo.AddBrand(brand); err != nil {
		t.Errorf("failed to add brand: %v", err)
	}
	if err := repo.AddBrand(brand); err == nil {
		t.Errorf("expected error when adding duplicate brand, got nil")
	}
}

func TestBrand_CaseSensitivity(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	b1 := "fortinet"
	b2 := "Fortinet"
	if err := repo.AddBrand(b1); err != nil {
		t.Errorf("failed to add brand: %v", err)
	}
	if err := repo.AddBrand(b2); err != nil {
		t.Errorf("failed to add brand with different case: %v", err)
	}
	got := repo.GetBrands()
	if !brandSliceContains(got, b1) || !brandSliceContains(got, b2) {
		t.Errorf("expected both %q and %q in list, got %v", b1, b2, got)
	}
}

// Helper function to check if a brand name exists in a slice of e.Brand
func brandSliceContains(brands []e.Brand, name string) bool {
	for _, b := range brands {
		if b.Name == name {
			return true
		}
	}
	return false
}
