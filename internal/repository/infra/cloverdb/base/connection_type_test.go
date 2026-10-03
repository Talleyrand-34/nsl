// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- ConnectionType Tests --- //

func TestConnectionType_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	connectionTypes := []string{"Ethernet", "Fiber", "Wireless"}
	for _, ct := range connectionTypes {
		if err := repo.AddConnectionType(ct); err != nil {
			t.Errorf("failed to add connection type %q: %v", ct, err)
		}
	}

	got, _ := repo.GetConnectionTypes()
	for _, want := range connectionTypes {
		if !connectionTypeSliceContains(got, want) {
			t.Errorf("expected connection type %q in list, got %v", want, got)
		}
	}
}

func TestConnectionType_AddDuplicate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	connectionType := "Ethernet"
	if err := repo.AddConnectionType(connectionType); err != nil {
		t.Errorf("failed to add connection type: %v", err)
	}
	if err := repo.AddConnectionType(connectionType); err == nil {
		t.Errorf("expected error when adding duplicate connection type, got nil")
	}
}

func TestConnectionType_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	originalName := "Fiber"
	if err := repo.AddConnectionType(originalName); err != nil {
		t.Errorf("failed to add connection type %q: %v", originalName, err)
	}

	connectionTypes, err := repo.GetConnectionTypes()
	if err != nil {
		t.Fatalf("failed to get connection types: %v", err)
	}

	var connectionTypeId string
	for _, ct := range connectionTypes {
		if ct.Name == originalName {
			connectionTypeId = ct.ID
			break
		}
	}

	if connectionTypeId == "" {
		t.Fatalf("failed to find connection type ID for %q", originalName)
	}

	updatedName := "Fiber Optic"
	if err := repo.UpdateConnectionType(connectionTypeId, updatedName); err != nil {
		t.Errorf("failed to update connection type: %v", err)
	}

	connectionTypesAfterUpdate, err := repo.GetConnectionTypes()
	if err != nil {
		t.Errorf("failed to get connection types after update: %v", err)
	}

	if !connectionTypeSliceContains(connectionTypesAfterUpdate, updatedName) {
		t.Errorf("expected updated connection type %q in list, got %v", updatedName, connectionTypesAfterUpdate)
	}

	if connectionTypeSliceContains(connectionTypesAfterUpdate, originalName) {
		t.Errorf("original connection type %q should not exist after update, got %v", originalName, connectionTypesAfterUpdate)
	}
}

func TestConnectionType_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	connectionTypeName := "Wireless"
	if err := repo.AddConnectionType(connectionTypeName); err != nil {
		t.Errorf("failed to add connection type %q: %v", connectionTypeName, err)
	}

	connectionTypes, err := repo.GetConnectionTypes()
	if err != nil {
		t.Errorf("failed to get connection types: %v", err)
	}

	if !connectionTypeSliceContains(connectionTypes, connectionTypeName) {
		t.Errorf("expected connection type %q in list before deletion, got %v", connectionTypeName, connectionTypes)
	}

	if err := repo.DeleteConnectionType(connectionTypeName); err != nil {
		t.Errorf("failed to delete connection type %q: %v", connectionTypeName, err)
	}

	connectionTypesAfterDelete, err := repo.GetConnectionTypes()
	if err != nil {
		t.Errorf("failed to get connection types after deletion: %v", err)
	}

	if connectionTypeSliceContains(connectionTypesAfterDelete, connectionTypeName) {
		t.Errorf("connection type %q should have been deleted, but got %v", connectionTypeName, connectionTypesAfterDelete)
	}
}

func connectionTypeSliceContains(connectionTypes []e.ConnectionType, name string) bool {
	for _, ct := range connectionTypes {
		if ct.Name == name {
			return true
		}
	}
	return false
}
