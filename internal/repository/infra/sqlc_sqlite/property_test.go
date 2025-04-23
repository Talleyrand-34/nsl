package sqlc_sqlite

import (
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// --- Zonetypes test --- //

func TestProprietary_AddAndGetTypicalOwners(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	owners := []string{"IT Department", "OT Team", "External Vendor", "Facility Management"}
	for _, o := range owners {
		if err := repo.AddProprietary(o); err != nil {
			t.Errorf("failed to add proprietary/owner %q: %v", o, err)
		}
	}

	got := repo.GetProperties()
	for _, want := range owners {
		if !contains(got, want) {
			t.Errorf("expected proprietary/owner %q in list, got %v", want, got)
		}
	}
}

func TestProprietary_AddDuplicateOwner(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	owner := "IT Department"
	if err := repo.AddProprietary(owner); err != nil {
		t.Errorf("failed to add proprietary/owner: %v", err)
	}
	if err := repo.AddProprietary(owner); err == nil {
		t.Errorf("expected error when adding duplicate proprietary/owner, got nil")
	}
}

// func TestProprietary_AddEmptyAndSpecialChars(t *testing.T) {
// 	repo, err := setupTestRepository(t)
// 	if err != nil {
// 		t.Fatalf("failed to setup repository: %v", err)
// 	}
// 	defer repo.Close()
//
// 	if err := repo.AddProprietary(""); err == nil {
// 		t.Errorf("expected error when adding empty proprietary/owner, got nil")
// 	}
//
// 	special := "Vendor#1"
// 	if err := repo.AddProprietary(special); err != nil {
// 		t.Errorf("failed to add proprietary/owner with special chars: %v", err)
// 	}
// 	got := repo.GetProperties()
// 	if !contains(got, special) {
// 		t.Errorf("expected proprietary/owner %q in list, got %v", special, got)
// 	}
// }

func TestProprietary_CaseSensitivity(t *testing.T) {
	repo, err := setupTestRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer repo.Close()

	p1 := "it department"
	p2 := "IT Department"
	if err := repo.AddProprietary(p1); err != nil {
		t.Errorf("failed to add proprietary/owner: %v", err)
	}
	if err := repo.AddProprietary(p2); err != nil {
		t.Errorf("failed to add proprietary/owner with different case: %v", err)
	}
	got := repo.GetProperties()
	if !contains(got, p1) || !contains(got, p2) {
		t.Errorf("expected both %q and %q in list, got %v", p1, p2, got)
	}
}
