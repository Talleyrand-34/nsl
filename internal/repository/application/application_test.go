
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
package application_test

import (
	"database/sql"
	"nsl-graph/internal/repository/application"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/assert"

	sqlite "nsl-graph/internal/repository/infra/sqlc_sqlite/base"
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
	assert.Contains(t, string(brands), "TestBrand")
}
