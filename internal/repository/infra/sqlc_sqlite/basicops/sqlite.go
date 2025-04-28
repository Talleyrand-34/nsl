
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
	"context"
	"database/sql"

	s "nsl-graph/internal/repository/infra/sqlc_sqlite"
	iinfra "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

type BasicOpsSQLiteRepository struct {
	query *iinfra.Queries
}

// NewSQLiteRepositoryFromDB creates a db connection from a db
func NewSQLiteRepositoryFromDB(db *sql.DB) (BasicOpsSQLiteRepository, error) {
	ctx := context.Background()
	// create tables
	ddl := s.Ddl
	db.ExecContext(ctx, ddl)
	queries := iinfra.New(db)

	return BasicOpsSQLiteRepository{query: queries}, nil
}

// NewSQLiteRepository creates a db connection from a file
func NewSQLiteRepository(filePath string) (BasicOpsSQLiteRepository, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return BasicOpsSQLiteRepository{}, err
	}
	return NewSQLiteRepositoryFromDB(db)
}

// Close closes the database connection
func (r BasicOpsSQLiteRepository) Close() error {
	return nil
}
