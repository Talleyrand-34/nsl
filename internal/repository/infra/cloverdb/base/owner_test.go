package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- Owner Tests --- //

func TestOwner_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	owners := []string{"Company A", "Company B", "Company C"}
	for _, p := range owners {
		if err := repo.AddOwner(p); err != nil {
			t.Errorf("failed to add owner %q: %v", p, err)
		}
	}

	got, _ := repo.GetOwners()
	for _, want := range owners {
		if !ownerSliceContains(got, want) {
			t.Errorf("expected owner %q in list, got %v", want, got)
		}
	}
}

func TestOwner_AddDuplicate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	owner := "Company A"
	if err := repo.AddOwner(owner); err != nil {
		t.Errorf("failed to add owner: %v", err)
	}
	if err := repo.AddOwner(owner); err == nil {
		t.Errorf("expected error when adding duplicate owner, got nil")
	}
}

func TestOwner_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	originalName := "Company A"
	if err := repo.AddOwner(originalName); err != nil {
		t.Errorf("failed to add owner %q: %v", originalName, err)
	}

	owners, err := repo.GetOwners()
	if err != nil {
		t.Fatalf("failed to get owners: %v", err)
	}

	var ownerId string
	for _, p := range owners {
		if p.Name == originalName {
			ownerId = p.ID
			break
		}
	}

	if ownerId == "" {
		t.Fatalf("failed to find owner ID for %q", originalName)
	}

	updatedName := "Company A Inc."
	if err := repo.UpdateOwner(ownerId, updatedName); err != nil {
		t.Errorf("failed to update owner: %v", err)
	}

	ownersAfterUpdate, err := repo.GetOwners()
	if err != nil {
		t.Errorf("failed to get owners after update: %v", err)
	}

	if !ownerSliceContains(ownersAfterUpdate, updatedName) {
		t.Errorf("expected updated owner %q in list, got %v", updatedName, ownersAfterUpdate)
	}

	if ownerSliceContains(ownersAfterUpdate, originalName) {
		t.Errorf("original owner %q should not exist after update, got %v", originalName, ownersAfterUpdate)
	}
}

func TestOwner_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	ownerName := "Company B"
	if err := repo.AddOwner(ownerName); err != nil {
		t.Errorf("failed to add owner %q: %v", ownerName, err)
	}

	owners, err := repo.GetOwners()
	if err != nil {
		t.Errorf("failed to get owners: %v", err)
	}

	if !ownerSliceContains(owners, ownerName) {
		t.Errorf("expected owner %q in list before deletion, got %v", ownerName, owners)
	}

	if err := repo.DeleteOwner(ownerName); err != nil {
		t.Errorf("failed to delete owner %q: %v", ownerName, err)
	}

	ownersAfterDelete, err := repo.GetOwners()
	if err != nil {
		t.Errorf("failed to get owners after deletion: %v", err)
	}

	if ownerSliceContains(ownersAfterDelete, ownerName) {
		t.Errorf("owner %q should have been deleted, but got %v", ownerName, ownersAfterDelete)
	}
}

func ownerSliceContains(owners []e.Owner, name string) bool {
	for _, p := range owners {
		if p.Name == name {
			return true
		}
	}
	return false
}
