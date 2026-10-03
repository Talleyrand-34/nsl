// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- ZoneType Tests --- //

func TestZoneType_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	zoneTypes := []string{"Building", "Floor", "Room"}
	for _, zt := range zoneTypes {
		if err := repo.AddZoneType(zt); err != nil {
			t.Errorf("failed to add zone type %q: %v", zt, err)
		}
	}

	got, _ := repo.GetZonetypes()
	for _, want := range zoneTypes {
		if !zoneTypeSliceContains(got, want) {
			t.Errorf("expected zone type %q in list, got %v", want, got)
		}
	}
}

func TestZoneType_AddDuplicate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	zoneType := "Building"
	if err := repo.AddZoneType(zoneType); err != nil {
		t.Errorf("failed to add zone type: %v", err)
	}
	if err := repo.AddZoneType(zoneType); err == nil {
		t.Errorf("expected error when adding duplicate zone type, got nil")
	}
}

func TestZoneType_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	originalName := "Floor"
	if err := repo.AddZoneType(originalName); err != nil {
		t.Errorf("failed to add zone type %q: %v", originalName, err)
	}

	zoneTypes, err := repo.GetZonetypes()
	if err != nil {
		t.Fatalf("failed to get zone types: %v", err)
	}

	var zoneTypeId string
	for _, zt := range zoneTypes {
		if zt.Name == originalName {
			zoneTypeId = zt.ID
			break
		}
	}

	if zoneTypeId == "" {
		t.Fatalf("failed to find zone type ID for %q", originalName)
	}

	updatedName := "Floor Level"
	if err := repo.UpdateZoneType(zoneTypeId, updatedName); err != nil {
		t.Errorf("failed to update zone type: %v", err)
	}

	zoneTypesAfterUpdate, err := repo.GetZonetypes()
	if err != nil {
		t.Errorf("failed to get zone types after update: %v", err)
	}

	if !zoneTypeSliceContains(zoneTypesAfterUpdate, updatedName) {
		t.Errorf("expected updated zone type %q in list, got %v", updatedName, zoneTypesAfterUpdate)
	}

	if zoneTypeSliceContains(zoneTypesAfterUpdate, originalName) {
		t.Errorf("original zone type %q should not exist after update, got %v", originalName, zoneTypesAfterUpdate)
	}
}

func TestZoneType_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	zoneTypeName := "Room"
	if err := repo.AddZoneType(zoneTypeName); err != nil {
		t.Errorf("failed to add zone type %q: %v", zoneTypeName, err)
	}

	zoneTypes, err := repo.GetZonetypes()
	if err != nil {
		t.Errorf("failed to get zone types: %v", err)
	}

	if !zoneTypeSliceContains(zoneTypes, zoneTypeName) {
		t.Errorf("expected zone type %q in list before deletion, got %v", zoneTypeName, zoneTypes)
	}

	if err := repo.DeleteZoneType(zoneTypeName); err != nil {
		t.Errorf("failed to delete zone type %q: %v", zoneTypeName, err)
	}

	zoneTypesAfterDelete, err := repo.GetZonetypes()
	if err != nil {
		t.Errorf("failed to get zone types after deletion: %v", err)
	}

	if zoneTypeSliceContains(zoneTypesAfterDelete, zoneTypeName) {
		t.Errorf("zone type %q should have been deleted, but got %v", zoneTypeName, zoneTypesAfterDelete)
	}
}

func zoneTypeSliceContains(zoneTypes []e.ZoneType, name string) bool {
	for _, zt := range zoneTypes {
		if zt.Name == name {
			return true
		}
	}
	return false
}
