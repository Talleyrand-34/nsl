package sqlc_sqlite

import (
	"context"
	"database/sql"
	_ "embed"

	iinfra "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

//go:embed schema.sql
var ddl string

type SQLiteRepository struct {
	query *iinfra.Queries
}

func NewSQLiteRepositoryFromDB(db *sql.DB) (SQLiteRepository, error) {
	ctx := context.Background()
	// create tables
	db.ExecContext(ctx, ddl)
	queries := iinfra.New(db)

	return SQLiteRepository{query: queries}, nil
}

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
