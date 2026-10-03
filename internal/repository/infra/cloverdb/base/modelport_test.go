// SPDX-License-Identifier: AGPL-3.0-or-later
// Copyright (C) 2025 Talleyrand-34 (t34@t34.dev)
//
// This file is part of NSL-Graph, released under the AGPL-3.0.

package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- ModelPort Tests --- //

func TestModelPort_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Add model port
	if err := repo.AddModelPort("GigabitEthernet1/0/1", "0", "0", "EX4300", false, "", ""); err != nil {
		t.Errorf("failed to add model port: %v", err)
	}

	modelPorts, err := repo.GetModelPorts()
	if err != nil {
		t.Errorf("failed to get model ports: %v", err)
	}

	if !modelPortSliceContains(modelPorts, "GigabitEthernet1/0/1") {
		t.Errorf("expected model port 'GigabitEthernet1/0/1' in list, got %v", modelPorts)
	}
}

func TestModelPort_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Add model port
	if err := repo.AddModelPort("Gi1/0/1", "0", "0", "EX4300", false, "", ""); err != nil {
		t.Fatalf("failed to add model port: %v", err)
	}

	// Get model port ID
	modelPorts, err := repo.GetModelPorts()
	if err != nil {
		t.Fatalf("failed to get model ports: %v", err)
	}

	var modelPortId string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/1" {
			modelPortId = mp.ID
			break
		}
	}

	if modelPortId == "" {
		t.Fatalf("failed to find model port ID")
	}

	// Get model ID
	models, _ := repo.GetModels()
	var modelId string
	for _, m := range models {
		if m.Model == "EX4300" {
			modelId = m.ID
			break
		}
	}

	// Update model port
	if err := repo.UpdateModelPort(modelPortId, "Gi1/0/2", "1", "0", modelId, false); err != nil {
		t.Errorf("failed to update model port: %v", err)
	}

	// Verify update
	modelPortsAfterUpdate, err := repo.GetModelPorts()
	if err != nil {
		t.Errorf("failed to get model ports after update: %v", err)
	}

	if !modelPortSliceContains(modelPortsAfterUpdate, "Gi1/0/2") {
		t.Errorf("expected updated model port 'Gi1/0/2' in list, got %v", modelPortsAfterUpdate)
	}
}

func TestModelPort_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddBrand("Juniper"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddModelType("Switch"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("EX4300", "Juniper", "Switch", ""); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Add model port
	if err := repo.AddModelPort("Gi1/0/3", "0", "1", "EX4300", false, "", ""); err != nil {
		t.Errorf("failed to add model port: %v", err)
	}

	// Get model port ID
	modelPorts, err := repo.GetModelPorts()
	if err != nil {
		t.Errorf("failed to get model ports: %v", err)
	}

	var modelPortId string
	for _, mp := range modelPorts {
		if mp.Name == "Gi1/0/3" {
			modelPortId = mp.ID
			break
		}
	}

	if modelPortId == "" {
		t.Fatalf("failed to find model port ID")
	}

	// Delete model port
	if err := repo.DeleteModelPort(modelPortId); err != nil {
		t.Errorf("failed to delete model port: %v", err)
	}

	// Verify deletion
	modelPortsAfterDelete, err := repo.GetModelPorts()
	if err != nil {
		t.Errorf("failed to get model ports after deletion: %v", err)
	}

	if modelPortSliceContains(modelPortsAfterDelete, "Gi1/0/3") {
		t.Errorf("model port 'Gi1/0/3' should have been deleted, but got %v", modelPortsAfterDelete)
	}
}

func modelPortSliceContains(modelPorts []e.ModelPort, portName string) bool {
	for _, mp := range modelPorts {
		if mp.Name == portName {
			return true
		}
	}
	return false
}
