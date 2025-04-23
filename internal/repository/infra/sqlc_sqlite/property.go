package sqlc_sqlite

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db
)

// deviceclass
func (r SQLiteRepository) AddProprietary(proprietary string) error {
	ctx := context.Background()
	execErr := r.query.AddProprietary(ctx, proprietary)
	if execErr != nil {
		return execErr
	}
	return nil
}

// deviceclass
func (r SQLiteRepository) GetProperties() []string {
	ctx := context.Background()
	proprietaries, execErr := r.query.GetProprietaries(ctx)
	if execErr != nil {
		return []string{}
	}
	return proprietaries
}
