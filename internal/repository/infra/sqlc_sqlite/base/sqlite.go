package sqlcbase

import (
	"context"
	"database/sql"

	s "nsl-graph/internal/repository/infra/sqlc_sqlite"
	"nsl-graph/internal/repository/infra/sqlc_sqlite/basicops"
	specops "nsl-graph/internal/repository/infra/sqlc_sqlite/specops"
)

// SQLiteRepository is the struct to access all the secialiced bd operations
type SQLiteRepository struct {
	basicops basicops.BasicOpsSQLiteRepository
	specops  specops.SpecOpsSQLiteRepository
}

// NewSQLiteRepositoryFromDB creates a db connection from a db
func NewSQLiteRepositoryFromDB(db *sql.DB) (SQLiteRepository, error) {
	ctx := context.Background()
	// create tables
	ddl := s.Ddl
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
