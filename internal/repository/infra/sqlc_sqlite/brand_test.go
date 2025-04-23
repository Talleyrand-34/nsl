package sqlc_sqlite

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
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
		if !contains(got, want) {
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

// func TestBrand_AddEmptyAndSpecialChars(t *testing.T) {
// 	repo, err := setupTestRepository(t)
// 	if err != nil {
// 		t.Fatalf("failed to setup repository: %v", err)
// 	}
// 	defer repo.Close()
//
// 	if err := repo.AddBrand(""); err == nil {
// 		t.Errorf("expected error when adding empty brand, got nil")
// 	}
//
// 	special := "Cisco&Co!"
// 	if err := repo.AddBrand(special); err != nil {
// 		t.Errorf("failed to add brand with special chars: %v", err)
// 	}
// 	got := repo.GetBrands()
// 	if !contains(got, special) {
// 		t.Errorf("expected brand %q in list, got %v", special, got)
// 	}
// }

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
	if !contains(got, b1) || !contains(got, b2) {
		t.Errorf("expected both %q and %q in list, got %v", b1, b2, got)
	}
}
