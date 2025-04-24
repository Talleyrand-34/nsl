package specops

import (
	"context"
	"database/sql"

	s "nsl-graph/internal/repository/infra/sqlc_sqlite"
	iinfra "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

type SpecOpsSQLiteRepository struct {
	query *iinfra.Queries
}

// NewSQLiteRepositoryFromDB creates a db connection from a db
func NewSQLiteRepositoryFromDB(db *sql.DB) (SpecOpsSQLiteRepository, error) {
	ctx := context.Background()
	// create tables
	ddl := s.Ddl
	db.ExecContext(ctx, ddl)
	queries := iinfra.New(db)

	return SpecOpsSQLiteRepository{query: queries}, nil
}

// NewSQLiteRepository creates a db connection from a file
func NewSQLiteRepository(filePath string) (SpecOpsSQLiteRepository, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return SpecOpsSQLiteRepository{}, err
	}
	return NewSQLiteRepositoryFromDB(db)
}

// Close closes the database connection
func (r SpecOpsSQLiteRepository) Close() error {
	return nil
}
