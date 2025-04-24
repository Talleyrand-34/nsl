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
