package sqlc_sqlite

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db
)

// deviceclass
func (r SQLiteRepository) AddZoneType(zoneName string) error {
	ctx := context.Background()
	execErr := r.query.AddZoneType(ctx, zoneName)
	if execErr != nil {
		return execErr
	}
	return nil
}

// deviceclass
func (r SQLiteRepository) GetZonetypes() []string {
	ctx := context.Background()
	zonetypes, execErr := r.query.GetZoneTypes(ctx)
	if execErr != nil {
		return []string{}
	}
	return zonetypes
}
