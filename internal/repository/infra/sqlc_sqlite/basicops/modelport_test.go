
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
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestModelPort_AddAndGetModelPorts(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare referenced Brand, DeviceClass, Model
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("F100", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Add model ports
	ports := []struct {
		name      string
		posx      string
		posy      string
		modelName string
	}{
		{"eth0", "10", "20", "F100"},
		{"eth1", "30", "40", "F100"},
	}
	for _, p := range ports {
		if err := repo.AddModelPort(p.name, p.posx, p.posy, p.modelName); err != nil {
			t.Errorf("failed to add model port %q: %v", p.name, err)
		}
	}

	// Retrieve and verify
	got, _ := repo.GetModelPorts()
	for _, want := range ports {
		found := false
		wantX, _ := strconv.Atoi(want.posx)
		wantY, _ := strconv.Atoi(want.posy)
		for _, m := range got {
			if m.Model == want.modelName && m.Positionx == wantX && m.Positiony == wantY {
				found = true
				break
			}
		}
		if !found {
			t.Errorf(
				"expected model port %q/%d/%d for model %q in list, got %+v",
				want.name,
				wantX,
				wantY,
				want.modelName,
				got,
			)
		}
	}
}

func TestModelPort_AddModelPort_InvalidReferences(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare valid model
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("F100", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Invalid model name
	err = repo.AddModelPort("ethX", "10", "20", "NonExistentModel")
	if err == nil {
		t.Errorf("expected error for non-existent model, got nil")
	}

	// Invalid posx
	err = repo.AddModelPort("eth0", "notanumber", "20", "F100")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid posx, got error: %v", err)
	}

	// Invalid posy
	err = repo.AddModelPort("eth0", "10", "notanumber", "F100")
	if err != nil {
		t.Errorf("expected silent fail (nil) for invalid posy, got error: %v", err)
	}
}
