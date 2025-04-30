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

func TestBrand_AddGetAndDeleteIndustrialBrands(t *testing.T) {
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

	// Now test delete
	deleteTarget := "Siemens"
	if err := repo.DeleteBrand(deleteTarget); err != nil {
		t.Errorf("failed to delete brand %q: %v", deleteTarget, err)
	}

	gotAfterDelete, _ := repo.GetBrands()
	if brandSliceContains(gotAfterDelete, deleteTarget) {
		t.Errorf("brand %q should have been deleted, but got %v", deleteTarget, gotAfterDelete)
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
