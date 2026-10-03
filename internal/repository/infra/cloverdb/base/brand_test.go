// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- Brand Tests --- //

func TestBrand_AddAndGetIndustrialBrands(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	brands := []string{"Fortinet", "Siemens", "Juniper"}
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
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	brand := "Fortinet"
	if err := repo.AddBrand(brand); err != nil {
		t.Errorf("failed to add brand: %v", err)
	}
	if err := repo.AddBrand(brand); err == nil {
		t.Errorf("expected error when adding duplicate brand, got nil")
	}
}

func TestBrand_CaseSensitivity(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

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
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Test CREATE operation
	brandName := "Juniper Networks"
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
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Test CREATE operation
	originalName := "Juniper"
	if err := repo.AddBrand(originalName); err != nil {
		t.Errorf("failed to add brand %q: %v", originalName, err)
	}

	// Get the brand ID
	brands, err := repo.GetBrands()
	if err != nil {
		t.Fatalf("failed to get brands: %v", err)
	}

	var brandId string
	for _, b := range brands {
		if b.Name == originalName {
			println(b.ID)
			brandId = b.ID
			break
		}
	}

	if brandId == "" {
		t.Fatalf("failed to find brand ID for %q", originalName)
	}

	// Test UPDATE operation
	updatedName := "Juniper Networks"
	if err := repo.UpdateBrand(brandId, updatedName); err != nil {
		t.Errorf("failed to update brand: %v", err)
	}

	// Verify update was successful
	brandsAfterUpdate, err := repo.GetBrands()
	if err != nil {
		t.Errorf("failed to get brands after update: %v", err)
	}

	if !brandSliceContains(brandsAfterUpdate, updatedName) {
		t.Errorf("expected updated brand %q in list, got %v", updatedName, brandsAfterUpdate)
	}

	if brandSliceContains(brandsAfterUpdate, originalName) {
		t.Errorf("original brand %q should not exist after update, got %v", originalName, brandsAfterUpdate)
	}
}

func TestBrand_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

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
