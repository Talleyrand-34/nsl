package basicops

import (
	"context"
	"strconv"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func Test_getFatherID_ByID(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Valid integer string
	id := repo.getFatherID(context.Background(), "42", "")
	if !id.Valid || id.Int64 != 42 {
		t.Errorf("expected valid id=42, got %+v", id)
	}

	// Invalid integer string
	id = repo.getFatherID(context.Background(), "notanumber", "")
	if id.Valid {
		t.Errorf("expected invalid id for non-integer input, got %+v", id)
	}
}

func Test_getFatherID_ByName(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare a zone to resolve by name
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "IT Department", "Physical"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	zones := repo.GetZones()
	var hqID int64
	for _, z := range zones {
		if z.Name == "HQ" {
			hqID = int64(z.ID)
			break
		}
	}
	if hqID == 0 {
		t.Fatalf("could not find HQ zone id")
	}

	id := repo.getFatherID(context.Background(), "", "HQ")
	if !id.Valid || id.Int64 != hqID {
		t.Errorf("expected valid id=%d, got %+v", hqID, id)
	}

	// Non-existent name
	id = repo.getFatherID(context.Background(), "", "NonExistentZone")
	if id.Valid {
		t.Errorf("expected invalid id for non-existent zone name, got %+v", id)
	}
}

func Test_getFatherID_NoneProvided(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	id := repo.getFatherID(context.Background(), "", "")
	if id.Valid {
		t.Errorf("expected invalid id, got %+v", id)
	}
}

func Test_getProprietaryID_ValidAndInvalid(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	id := repo.getProprietaryID(context.Background(), "IT Department")
	if !id.Valid || id.Int64 == 0 {
		t.Errorf("expected valid proprietary id, got %+v", id)
	}

	// Non-existent
	id = repo.getProprietaryID(context.Background(), "NonExistent")
	if id.Valid {
		t.Errorf("expected invalid id for non-existent proprietary, got %+v", id)
	}
}

func Test_getZoneTypeID_ValidAndInvalid(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	id := repo.getZoneTypeID(context.Background(), "Physical")
	if !id.Valid || id.Int64 == 0 {
		t.Errorf("expected valid zone type id, got %+v", id)
	}

	// Non-existent
	id = repo.getZoneTypeID(context.Background(), "NonExistent")
	if id.Valid {
		t.Errorf("expected invalid id for non-existent zone type, got %+v", id)
	}
}

func Test_getZoneID_ByIDAndByName(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare a zone to resolve by name
	if err := repo.AddProprietary("IT Department"); err != nil {
		t.Fatalf("failed to add proprietary: %v", err)
	}
	if err := repo.AddZoneType("Physical"); err != nil {
		t.Fatalf("failed to add zone type: %v", err)
	}
	if err := repo.AddZone("HQ", "", "", "IT Department", "Physical"); err != nil {
		t.Fatalf("failed to add zone: %v", err)
	}
	zones := repo.GetZones()
	var hqID int64
	for _, z := range zones {
		if z.Name == "HQ" {
			hqID = int64(z.ID)
			break
		}
	}
	if hqID == 0 {
		t.Fatalf("could not find HQ zone id")
	}

	// By ID
	id := repo.getZoneID(context.Background(), strconv.FormatInt(hqID, 10), "")
	if !id.Valid || id.Int64 != hqID {
		t.Errorf("expected valid id=%d, got %+v", hqID, id)
	}

	// By name
	id = repo.getZoneID(context.Background(), "", "HQ")
	if !id.Valid || id.Int64 != hqID {
		t.Errorf("expected valid id=%d, got %+v", hqID, id)
	}

	// Invalid ID
	id = repo.getZoneID(context.Background(), "notanumber", "")
	if id.Valid {
		t.Errorf("expected invalid id for non-integer input, got %+v", id)
	}

	// Non-existent name
	id = repo.getZoneID(context.Background(), "", "NonExistentZone")
	if id.Valid {
		t.Errorf("expected invalid id for non-existent zone name, got %+v", id)
	}
}

func Test_getModelID_ValidAndInvalid(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	// Prepare referenced data
	if err := repo.AddBrand("Fortinet"); err != nil {
		t.Fatalf("failed to add brand: %v", err)
	}
	if err := repo.AddDeviceClass("Router"); err != nil {
		t.Fatalf("failed to add device class: %v", err)
	}
	if err := repo.AddModel("F100", "Fortinet", "Router"); err != nil {
		t.Fatalf("failed to add model: %v", err)
	}

	// Valid
	id, err := repo.getModelID(context.Background(), "F100")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if id == 0 {
		t.Errorf("expected valid model id, got %d", id)
	}

	// Non-existent
	id, err = repo.getModelID(context.Background(), "NonExistentModel")
	if err == nil {
		t.Errorf("expected error for non-existent model, got nil")
	}
}
