// sqlite.go contains the public struct with the compositions of db accesses
package sqlcbase

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

import (
	"context"
	"database/sql"
	"nsl-graph/internal/repository/infra/sqlc_sqlite/basicops"

	s "nsl-graph/internal/repository/infra/sqlc_sqlite"

	specops "nsl-graph/internal/repository/infra/sqlc_sqlite/specops"
)

// SQLiteRepository is the struct to access all the db operations
type SQLiteRepository struct {
	basicops basicops.BasicOpsSQLiteRepository
	specops  specops.SpecOpsSQLiteRepository
}

// NewSQLiteRepositoryFromDB creates a db connection from a db
func NewSQLiteRepositoryFromDB(db *sql.DB) (SQLiteRepository, error) {
	ctx := context.Background()
	// create tables
	ddl := s.Ddl
	// Whether the table does not exists and it is created or it exists yet
	db.ExecContext(ctx, ddl)
	basicops, err := basicops.NewSQLiteRepositoryFromDB(db)
	if err != nil {
		return SQLiteRepository{}, err
	}
	specops, err := specops.NewSQLiteRepositoryFromDB(db)
	if err != nil {
		return SQLiteRepository{}, err
	}

	return SQLiteRepository{basicops: basicops, specops: specops}, nil
}

// NewSQLiteRepository creates a db connection from a file
func NewSQLiteRepository(filePath string) (SQLiteRepository, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return SQLiteRepository{}, err
	}
	return NewSQLiteRepositoryFromDB(db)
}

// Close closes the database connection
func (r SQLiteRepository) Close() error {
	return nil
}
