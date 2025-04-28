
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
package basicops

import (
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// Helper to check if a string is in a slice
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func setupTestRepository(t *testing.T) (BasicOpsSQLiteRepository, error) {
	// Use in-memory DB for tests, or a temp file
	db, err := sql.Open("sqlite3", ":memory:")
	// db, err := sql.Open("sqlite3", "/tmp/test.db")
	// db, err := sql.Open("sqlite3", "file:memdb1?mode=memory&cache=shared")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	repo, err := NewSQLiteRepositoryFromDB(db)
	// repo, err := NewSQLiteRepository("/tmp/test.db")
	if err != nil {
		return BasicOpsSQLiteRepository{}, err
	}

	return repo, nil
}
