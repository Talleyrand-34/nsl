package sqlite

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

// func TestSQLiteRepositoryGetBrands(t *testing.T) {
// 	// Setup repository with in-memory database
// 	repo, err := setupTestRepository(t)
// 	assert.NoError(t, err)
// 	defer repo.Close()
//
// 	// Add multiple brands
// 	brandNames := []string{"Apple", "Samsung", "Google", "Xiaomi"}
// 	for _, brand := range brandNames {
// 		err = repo.AddBrand(brand)
// 		assert.NoError(t, err)
// 	}
//
// 	// Get all brands
// 	brands := repo.GetBrands()
//
// 	// Check length
// 	assert.Equal(t, len(brandNames), len(brands))
//
// 	// Check content
// 	for _, brand := range brandNames {
// 		assert.Contains(t, brands, brand)
// 	}
//
// 	// Check alphabetical order - only if we have results
// 	if len(brands) >= 4 {
// 		assert.Equal(t, "Apple", brands[0])
// 		assert.Equal(t, "Google", brands[1])
// 		assert.Equal(t, "Samsung", brands[2])
// 		assert.Equal(t, "Xiaomi", brands[3])
// 	}
// }
//
// func TestSQLiteRepositoryGetBrandsEmpty(t *testing.T) {
// 	// Setup repository with in-memory database
// 	repo, err := setupTestRepository(t)
// 	assert.NoError(t, err)
// 	defer repo.Close()
//
// 	// Get brands when no brands exist
// 	brands := repo.GetBrands()
// 	assert.Empty(t, brands)
// }
//
// func TestSQLiteRepositoryAddDuplicateBrand(t *testing.T) {
// 	// Setup repository with in-memory database
// 	repo, err := setupTestRepository(t)
// 	assert.NoError(t, err)
// 	defer repo.Close()
//
// 	// Add a brand
// 	err = repo.AddBrand("DuplicateBrand")
// 	assert.NoError(t, err)
//
// 	// Try to add the same brand again
// 	_ = repo.AddBrand("DuplicateBrand")
// 	// This could fail if your schema enforces uniqueness
//
// 	// Verify we have at least one instance of the brand
// 	brands := repo.GetBrands()
// 	assert.Contains(t, brands, "DuplicateBrand")
// }

// setupTestRepository creates a new SQLiteRepository with an in-memory database
// and initializes the schema
//
//	func setupTestRepository(t *testing.T) (SQLiteRepository, error) {
//		// repo, err := NewSQLiteRepository(":memory:")
//		repo, err := NewSQLiteRepository("/tmp/sql2.db")
//		if err != nil {
//			return SQLiteRepository{}, err
//		}
//
//		return repo, nil
//	}
func setupTestRepository(t *testing.T) (SQLiteRepository, error) {
	// Use in-memory DB for tests, or a temp file
	db, err := sql.Open("sqlite3", ":memory:")
	// db, err := sql.Open("sqlite3", "/tmp/test.db")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	// Create the Brand table for testing
	_, err = db.Exec(`
		CREATE TABLE Brand (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			brand TEXT NOT NULL
		);
	`)
	if err != nil {
		db.Close()
		t.Fatalf("failed to create table: %v", err)
	}

	// Now pass the same db connection to your repository
	// If your repository expects a file path, you need to use a file-based DB
	// For demonstration, let's assume you use :memory: and pass the same connection

	// If your repository expects a file path, you need to close this db and use the file path
	// For file-based:
	// db, err := sql.Open("sqlite3", "/tmp/sql2.db")
	// ... create table as above ...

	// db.Close() // Close the direct connection, repository will open its own

	// Now create the repository (it will check for table existence)
	repo, err := NewSQLiteRepositoryFromDB(db)
	// repo, err := NewSQLiteRepository("/tmp/test.db")
	if err != nil {
		return SQLiteRepository{}, err
	}

	return repo, nil
}
