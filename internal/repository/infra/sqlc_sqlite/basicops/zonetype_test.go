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

// --- Zonetypes test --- //
func TestZoneType_AddAndGetTypicalZoneTypes(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	zoneTypes := []string{"Physical", "VPN", "DMZ", "Guest"}
	for _, z := range zoneTypes {
		if err := repo.AddZoneType(z); err != nil {
			t.Errorf("failed to add zone type %q: %v", z, err)
		}
	}

	got, _ := repo.GetZonetypes()
	for _, want := range zoneTypes {
		if !zoneTypeSliceContains(got, want) {
			t.Errorf("expected zone type %q in list, got %v", want, got)
		}
	}
}

func TestZoneType_AddDuplicateZoneType(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	zone := "VPN"
	if err := repo.AddZoneType(zone); err != nil {
		t.Errorf("failed to add zone type: %v", err)
	}
	if err := repo.AddZoneType(zone); err == nil {
		t.Errorf("expected error when adding duplicate zone type, got nil")
	}
}

// func TestZoneType_AddEmptyAndSpecialChars(t *testing.T) {
// 	repo, err := setupTestRepository(t)
// 	if err != nil {
// 		t.Fatalf("failed to setup repository: %v", err)
// 	}
// 	defer repo.Close()
//
// 	if err := repo.AddZoneType(""); err == nil {
// 		t.Errorf("expected error when adding empty zone type, got nil")
// 	}
//
// 	special := "VPN#1"
// 	if err := repo.AddZoneType(special); err != nil {
// 		t.Errorf("failed to add zone type with special chars: %v", err)
// 	}
// 	got := repo.GetZonetypes()
// 	if !contains(got, special) {
// 		t.Errorf("expected zone type %q in list, got %v", special, got)
// 	}
// }

func TestZoneType_CaseSensitivity(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	z1 := "vpn"
	z2 := "VPN"
	if err := repo.AddZoneType(z1); err != nil {
		t.Errorf("failed to add zone type: %v", err)
	}
	if err := repo.AddZoneType(z2); err != nil {
		t.Errorf("failed to add zone type with different case: %v", err)
	}
	got, _ := repo.GetZonetypes()
	if !zoneTypeSliceContains(got, z1) || !zoneTypeSliceContains(got, z2) {
		t.Errorf("expected both %q and %q in list, got %v", z1, z2, got)
	}
}

func TestZoneType_AddGetAndDeleteTypicalZoneTypes(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	zoneTypes := []string{"Physical", "VPN", "DMZ", "Guest"}
	for _, z := range zoneTypes {
		if err := repo.AddZoneType(z); err != nil {
			t.Errorf("failed to add zone type %q: %v", z, err)
		}
	}

	got, _ := repo.GetZonetypes()
	for _, want := range zoneTypes {
		if !zoneTypeSliceContains(got, want) {
			t.Errorf("expected zone type %q in list, got %v", want, got)
		}
	}

	// Now test delete
	deleteTarget := "DMZ"
	if err := repo.DeleteZoneType(deleteTarget); err != nil {
		t.Errorf("failed to delete zone type %q: %v", deleteTarget, err)
	}

	gotAfterDelete, _ := repo.GetZonetypes()
	if zoneTypeSliceContains(gotAfterDelete, deleteTarget) {
		t.Errorf("zone type %q should have been deleted, but got %v", deleteTarget, gotAfterDelete)
	}
}

// Helper function to check if a brand name exists in a slice of e.Brand
func zoneTypeSliceContains(zoneTypes []e.ZoneType, name string) bool {
	for _, b := range zoneTypes {
		if b.Name == name {
			return true
		}
	}
	return false
}
