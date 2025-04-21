package sqlc_sqlite

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"
)

func TestSQLiteRepositoryAddBrand(t *testing.T) {
	// Setup repository with in-memory database
	repo, err := setupTestRepository(t)
	assert.NoError(t, err)
	defer repo.Close()

	// Test adding a brand
	err = repo.AddBrand("TestBrand")
	assert.NoError(t, err)

	// Verify brand was added
	brands := repo.GetBrands()
	assert.Contains(t, brands, "TestBrand")
}

func setupTestRepository(t *testing.T) (SQLiteRepository, error) {
	// Use in-memory DB for tests, or a temp file
	db, err := sql.Open("sqlite3", ":memory:")
	// db, err := sql.Open("sqlite3", "/tmp/test.db")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	repo, err := NewSQLiteRepositoryFromDB(db)
	// repo, err := NewSQLiteRepository("/tmp/test.db")
	if err != nil {
		return SQLiteRepository{}, err
	}

	return repo, nil
}
