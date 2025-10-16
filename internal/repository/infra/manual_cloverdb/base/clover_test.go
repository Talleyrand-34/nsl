package basicops

import (
	"os"
	"path/filepath"
	"testing"

	c "github.com/ostafen/clover/v2"
)

// setupTestCloverRepository creates a test repository using a temporary directory
// Returns the repository, a cleanup function, and an error
func setupTestCloverRepository(t *testing.T) (BasicOpsCloverRepository, func(), error) {

	dbPath := "/tmp/clover"

	// Create a temporary directory for the test database
	// _ = os.Mkdir(dbPath, 0755)
	// if err != nil {
	// 	return BasicOpsCloverRepository{}, nil, err
	// }
	// Remove existing database directory to wipe memory/data
	dbfile := filepath.Join(dbPath, "data.db")
	err := os.Remove(dbfile)
	if err != nil {
		t.Fatalf("failed to clear db directory: %v", err)
	}
	// Open CloverDB in the temporary directory
	db, err := c.Open(dbPath)
	if err != nil {
		return BasicOpsCloverRepository{}, nil, err
	}

	repo, err := NewCloverRepositoryFromDB(db)
	if err != nil {
		db.Close()
		return BasicOpsCloverRepository{}, nil, err
	}

	// Cleanup function
	cleanup := func() {
		repo.Close()
	}

	return repo, cleanup, nil
}
