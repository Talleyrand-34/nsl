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

	got, _ := repo.GetBrands()
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
	got, _ := repo.GetBrands()
	if !brandSliceContains(got, b1) || !brandSliceContains(got, b2) {
		t.Errorf("expected both %q and %q in list, got %v", b1, b2, got)
	}
}

func TestBrand_CreateOnly(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Test CREATE operation
	brandName := "Cisco Systems"
	if err := repo.AddBrand(brandName); err != nil {
		t.Errorf("failed to add brand %q: %v", brandName, err)
	}

	// Verify brand was created
	brands, err := repo.GetBrands()
	if err != nil {
		t.Errorf("failed to get brands: %v", err)
	}

	if !brandSliceContains(brands, brandName) {
		t.Errorf("expected brand %q in list, got %v", brandName, brands)
	}
}

func TestBrand_CreateAndUpdate(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Test CREATE operation
	originalName := "Juniper"
	if err := repo.AddBrand(originalName); err != nil {
		t.Errorf("failed to add brand %q: %v", originalName, err)
	}

	// Get the brand ID using the existing query method
	brandId, err := repo.query.GetBrandId(context.Background(), originalName)
	if err != nil {
		t.Fatalf("failed to get brand ID: %v", err)
	}

	// Test UPDATE operation
	updatedName := "Juniper Networks"
	if err := repo.UpdateBrand(fmt.Sprintf("%d", brandId), updatedName); err != nil {
		t.Errorf("failed to update brand: %v", err)
	}

	// Verify update was successful
	brands, err := repo.GetBrands()
	if err != nil {
		t.Errorf("failed to get brands after update: %v", err)
	}

	if !brandSliceContains(brands, updatedName) {
		t.Errorf("expected updated brand %q in list, got %v", updatedName, brands)
	}

	if brandSliceContains(brands, originalName) {
		t.Errorf("original brand %q should not exist after update, got %v", originalName, brands)
	}
}

func TestBrand_CreateAndDelete(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Test CREATE operation
	brandName := "Fortinet"
	if err := repo.AddBrand(brandName); err != nil {
		t.Errorf("failed to add brand %q: %v", brandName, err)
	}

	// Verify brand was created
	brands, err := repo.GetBrands()
	if err != nil {
		t.Errorf("failed to get brands: %v", err)
	}

	if !brandSliceContains(brands, brandName) {
		t.Errorf("expected brand %q in list before deletion, got %v", brandName, brands)
	}

	// Test DELETE operation
	if err := repo.DeleteBrand(brandName); err != nil {
		t.Errorf("failed to delete brand %q: %v", brandName, err)
	}

	// Verify brand was deleted
	brandsAfterDelete, err := repo.GetBrands()
	if err != nil {
		t.Errorf("failed to get brands after deletion: %v", err)
	}

	if brandSliceContains(brandsAfterDelete, brandName) {
		t.Errorf("brand %q should have been deleted, but got %v", brandName, brandsAfterDelete)
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
