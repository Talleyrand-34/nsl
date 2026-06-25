package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- Zone Tests --- //

func TestZone_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddZoneType("Building"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddOwner("Company A"); err != nil {
		t.Fatalf("failed to add owner: %v", err)
	}

	// Add zone
	if err := repo.AddZone("Building A", "", "", "Company A", "Building"); err != nil {
		t.Errorf("failed to add zone: %v", err)
	}

	zones, err := repo.GetZones()
	if err != nil {
		t.Errorf("failed to get zones: %v", err)
	}

	if !zoneSliceContains(zones, "Building A") {
		t.Errorf("expected zone 'Building A' in list, got %v", zones)
	}
}

func TestZone_AddWithFather(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddZoneType("Building"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZoneType("Floor"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddOwner("Company A"); err != nil {
		t.Fatalf("failed to add owner: %v", err)
	}

	// Add parent zone
	if err := repo.AddZone("Building A", "", "", "Company A", "Building"); err != nil {
		t.Errorf("failed to add zone: %v", err)
	}

	// Add child zone
	if err := repo.AddZone("Floor 1", "", "Building A", "Company A", "Floor"); err != nil {
		t.Errorf("failed to add child zone: %v", err)
	}

	zones, err := repo.GetZones()
	if err != nil {
		t.Errorf("failed to get zones: %v", err)
	}

	// Verify child zone exists and has correct father
	var floor1 *e.Zone
	for i := range zones {
		if zones[i].Name == "Floor 1" {
			floor1 = &zones[i]
			break
		}
	}

	if floor1 == nil {
		t.Errorf("expected to find 'Floor 1' zone")
	} else if floor1.Father != "Building A" {
		t.Errorf("expected father 'Building A', got %q", floor1.Father)
	}
}

func TestZone_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	// Setup prerequisites
	if err := repo.AddZoneType("Building"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddOwner("Company A"); err != nil {
		t.Fatalf("failed to add owner: %v", err)
	}

	// Add zone
	if err := repo.AddZone("Building B", "", "", "Company A", "Building"); err != nil {
		t.Errorf("failed to add zone: %v", err)
	}

	// Get zone ID
	zones, err := repo.GetZones()
	if err != nil {
		t.Errorf("failed to get zones: %v", err)
	}

	var zoneId string
	for _, z := range zones {
		if z.Name == "Building B" {
			zoneId = z.ID
			break
		}
	}

	if zoneId == "" {
		t.Fatalf("failed to find zone ID")
	}

	// Delete zone
	if err := repo.DeleteZone(zoneId); err != nil {
		t.Errorf("failed to delete zone: %v", err)
	}

	// Verify deletion
	zonesAfterDelete, err := repo.GetZones()
	if err != nil {
		t.Errorf("failed to get zones after deletion: %v", err)
	}

	if zoneSliceContains(zonesAfterDelete, "Building B") {
		t.Errorf("zone 'Building B' should have been deleted, but got %v", zonesAfterDelete)
	}
}

func zoneSliceContains(zones []e.Zone, zoneName string) bool {
	for _, z := range zones {
		if z.Name == zoneName {
			return true
		}
	}
	return false
}
