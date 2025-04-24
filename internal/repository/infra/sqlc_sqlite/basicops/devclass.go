package basicops

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
)

// AddDeviceClass adds a new brand to the database
func (r BasicOpsSQLiteRepository) AddDeviceClass(devClassName string) error {
	ctx := context.Background()
	execErr := r.query.AddDeviceClass(ctx, devClassName)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetDeviceClasses gets all the brands available
func (r BasicOpsSQLiteRepository) GetDeviceClasses() []e.DevClass {
	ctx := context.Background()
	devclasses, execErr := r.query.GetDeviceClasses(ctx)
	if execErr != nil {
		return []e.DevClass{}
	}
	result := make([]e.DevClass, 0, len(devclasses))
	for _, row := range devclasses {
		singleDevClass := e.DevClass{
			Name: row,
		}
		result = append(result, singleDevClass)
	}
	return result
}
