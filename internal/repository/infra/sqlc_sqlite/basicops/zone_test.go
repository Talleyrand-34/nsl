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

func TestZone_AddAndGetZones(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare referenced data
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}

	// Add root zone (no father)
	if err := repo.AddZone(
		"Main Building", // name
		"",              // fatherid
		"",              // father
		"IT Department", // proprietary
		"Physical",      // zonename
	); err != nil {
		t.Fatalf("failed to add root zone: %v", err)
	}

	// Add child zone by father name
	if err := repo.AddZone(
		"Server Room",
		"",              // fatherid
		"Main Building", // father
		"IT Department",
		"Physical",
	); err != nil {
		t.Fatalf("failed to add child zone: %v", err)
	}

	// Add another proprietary and zone type for a different zone
	if err := repo.AddProprietary("Facility Management"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("VPN"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	// Add zone with new proprietary and zone type, by fatherid
	zones, _ := repo.GetZones()
	var mainBuildingId string
	for _, z := range zones {
		if z.Name == "Main Building" {
			mainBuildingId = strconv.Itoa(z.ID)
			break
		}
	}
	if mainBuildingId == "" {
		t.Fatalf("could not find Main Building zone id")
	}
	if err := repo.AddZone(
		"Remote Office",
		mainBuildingId, // fatherid
		"",             // father
		"Facility Management",
		"VPN",
	); err != nil {
		t.Fatalf("failed to add zone with fatherid: %v", err)
	}

	// Check all zones
	got, _ := repo.GetZones()
	names := []string{"Main Building", "Server Room", "Remote Office"}
	for _, want := range names {
		found := false
		for _, z := range got {
			if z.Name == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected zone %q in list, got %+v", want, got)
		}
	}
}

func TestZone_AddGetAndDeleteZones(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare referenced data
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}

	// Add root zone (no father)
	if err := repo.AddZone(
		"Main Building", // name
		"",              // fatherid
		"",              // father
		"IT Department", // proprietary
		"Physical",      // zonename
	); err != nil {
		t.Fatalf("failed to add root zone: %v", err)
	}

	// Add child zone by father name
	if err := repo.AddZone(
		"Server Room",
		"",              // fatherid
		"Main Building", // father
		"IT Department",
		"Physical",
	); err != nil {
		t.Fatalf("failed to add child zone: %v", err)
	}

	// Add another proprietary and zone type for a different zone
	if err := repo.AddProprietary("Facility Management"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("VPN"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	// Add zone with new proprietary and zone type, by fatherid
	zones, _ := repo.GetZones()
	var mainBuildingId string
	for _, z := range zones {
		if z.Name == "Main Building" {
			mainBuildingId = strconv.Itoa(z.ID)
			break
		}
	}
	if mainBuildingId == "" {
		t.Fatalf("could not find Main Building zone id")
	}
	if err := repo.AddZone(
		"Remote Office",
		mainBuildingId, // fatherid
		"",             // father
		"Facility Management",
		"VPN",
	); err != nil {
		t.Fatalf("failed to add zone with fatherid: %v", err)
	}

	// Check all zones
	got, _ := repo.GetZones()
	names := []string{"Main Building", "Server Room", "Remote Office"}
	for _, want := range names {
		found := false
		for _, z := range got {
			if z.Name == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected zone %q in list, got %+v", want, got)
		}
	}

	// --- Test delete by ID ---
	// Find the ID of the zone to delete
	var deleteZoneID string
	for _, z := range got {
		if z.Name == "Remote Office" {
			deleteZoneID = strconv.Itoa(z.ID)
			break
		}
	}
	if deleteZoneID == "" {
		t.Fatalf("could not find Remote Office zone id")
	}
	if err := repo.DeleteZone(deleteZoneID); err != nil {
		t.Errorf("failed to delete zone with id %q: %v", deleteZoneID, err)
	}

	// Check that the zone was deleted
	gotAfterDelete, _ := repo.GetZones()
	for _, z := range gotAfterDelete {
		if z.Name == "Remote Office" {
			t.Errorf(
				"zone %q should have been deleted, but got %+v",
				"Remote Office",
				gotAfterDelete,
			)
		}
	}
}

// func TestZone_AddZone_InvalidReferences(t *testing.T) {
// 	repo, err := setupTestRepository(t)
// 	if err != nil {
// 		t.Fatalf("failed to setup repository: %v", err)
// 	}
// 	defer repo.Close()
//
// 	// Try to add zone with non-existent proprietary
// 	err = repo.AddZone("Invalid Owner Zone", "", "", "NonExistentOwner", "Physical")
// 	if err == nil {
// 		t.Errorf("expected error when adding zone with invalid proprietary, got nil")
// 	}
//
// 	// Try to add zone with non-existent zone type
// 	if err := repo.AddProprietary("IT Department"); err != nil {
// 		t.Fatalf("failed to add proprietary: %v", err)
// 	}
// 	err = repo.AddZone("Invalid Zone Type", "", "", "IT Department", "NonExistentZoneType")
// 	if err == nil {
// 		t.Errorf("expected error when adding zone with invalid zone type, got nil")
// 	}
//
// 	// Try to add zone with invalid father name
// 	if err := repo.AddZoneType("Physical"); err != nil {
// 		t.Fatalf("failed to add zone type: %v", err)
// 	}
// 	err = repo.AddZone("Invalid Father", "", "NonExistentFather", "IT Department", "Physical")
// 	if err == nil {
// 		t.Errorf("expected error when adding zone with invalid father, got nil")
// 	}
//
// 	// Try to add zone with invalid fatherid
// 	err = repo.AddZone("Invalid FatherID", "9999", "", "IT Department", "Physical")
// 	if err == nil {
// 		t.Errorf("expected error when adding zone with invalid fatherid, got nil")
// 	}
// }
