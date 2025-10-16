package basicops

import (
	"testing"

	e "nsl-graph/internal/repository/entities"
)

// --- Proprietary Tests --- //

func TestProprietary_AddAndGet(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	proprietaries := []string{"Company A", "Company B", "Company C"}
	for _, p := range proprietaries {
		if err := repo.AddProprietary(p); err != nil {
			t.Errorf("failed to add proprietary %q: %v", p, err)
		}
	}

	got, _ := repo.GetProperties()
	for _, want := range proprietaries {
		if !proprietarySliceContains(got, want) {
			t.Errorf("expected proprietary %q in list, got %v", want, got)
		}
	}
}

func TestProprietary_AddDuplicate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	proprietary := "Company A"
	if err := repo.AddProprietary(proprietary); err != nil {
		t.Errorf("failed to add proprietary: %v", err)
	}
	if err := repo.AddProprietary(proprietary); err == nil {
		t.Errorf("expected error when adding duplicate proprietary, got nil")
	}
}

func TestProprietary_CreateAndUpdate(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	originalName := "Company A"
	if err := repo.AddProprietary(originalName); err != nil {
		t.Errorf("failed to add proprietary %q: %v", originalName, err)
	}

	proprietaries, err := repo.GetProperties()
	if err != nil {
		t.Fatalf("failed to get proprietaries: %v", err)
	}

	var proprietaryId string
	for _, p := range proprietaries {
		if p.Name == originalName {
			proprietaryId = p.ID
			break
		}
	}

	if proprietaryId == "" {
		t.Fatalf("failed to find proprietary ID for %q", originalName)
	}

	updatedName := "Company A Inc."
	if err := repo.UpdateProprietary(proprietaryId, updatedName); err != nil {
		t.Errorf("failed to update proprietary: %v", err)
	}

	proprietariesAfterUpdate, err := repo.GetProperties()
	if err != nil {
		t.Errorf("failed to get proprietaries after update: %v", err)
	}

	if !proprietarySliceContains(proprietariesAfterUpdate, updatedName) {
		t.Errorf("expected updated proprietary %q in list, got %v", updatedName, proprietariesAfterUpdate)
	}

	if proprietarySliceContains(proprietariesAfterUpdate, originalName) {
		t.Errorf("original proprietary %q should not exist after update, got %v", originalName, proprietariesAfterUpdate)
	}
}

func TestProprietary_CreateAndDelete(t *testing.T) {
	repo, cleanup, err := setupTestCloverRepository(t)
	if err != nil {
		t.Fatalf("failed to setup repository: %v", err)
	}
	defer cleanup()

	proprietaryName := "Company B"
	if err := repo.AddProprietary(proprietaryName); err != nil {
		t.Errorf("failed to add proprietary %q: %v", proprietaryName, err)
	}

	proprietaries, err := repo.GetProperties()
	if err != nil {
		t.Errorf("failed to get proprietaries: %v", err)
	}

	if !proprietarySliceContains(proprietaries, proprietaryName) {
		t.Errorf("expected proprietary %q in list before deletion, got %v", proprietaryName, proprietaries)
	}

	if err := repo.DeleteProprietary(proprietaryName); err != nil {
		t.Errorf("failed to delete proprietary %q: %v", proprietaryName, err)
	}

	proprietariesAfterDelete, err := repo.GetProperties()
	if err != nil {
		t.Errorf("failed to get proprietaries after deletion: %v", err)
	}

	if proprietarySliceContains(proprietariesAfterDelete, proprietaryName) {
		t.Errorf("proprietary %q should have been deleted, but got %v", proprietaryName, proprietariesAfterDelete)
	}
}

func proprietarySliceContains(proprietaries []e.Proprietary, name string) bool {
	for _, p := range proprietaries {
		if p.Name == name {
			return true
		}
	}
	return false
}
