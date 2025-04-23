package basicops

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db
)

// AddDeviceClass adds a new brand to the database
func (r SQLiteRepository) AddDeviceClass(devClassName string) error {
	ctx := context.Background()
	execErr := r.query.AddDeviceClass(ctx, devClassName)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetDeviceClasses gets all the brands available
func (r SQLiteRepository) GetDeviceClasses() []string {
	ctx := context.Background()
	brands, execErr := r.query.GetDeviceClasses(ctx)
	if execErr != nil {
		return []string{}
	}
	return brands
}
