package application_test

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	"nsl-graph/internal/repository/application"
	sqlite "nsl-graph/internal/repository/infra/bare_sqlite"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	assert.NoError(t, err)

	// Create the Brand table
	_, err = db.Exec(`
		CREATE TABLE Brand (
			ID INTEGER PRIMARY KEY AUTOINCREMENT,
			Brand TEXT NOT NULL UNIQUE
		);
	`)
	assert.NoError(t, err)

	return db
}

func TestNetService_AddAndGetBrand(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo, err := sqlite.NewSQLiteRepositoryFromDB(db)
	assert.NoError(t, err)

	service := application.NewNetService(repo)

	// Test AddBrand
	err = service.AddBrand("TestBrand")
	assert.NoError(t, err)

	// Test GetBrands
	brands := service.GetBrands()
	assert.Contains(t, brands, "TestBrand")
}
