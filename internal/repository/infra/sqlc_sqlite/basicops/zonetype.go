package basicops

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
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
func (r SQLiteRepository) GetZonetypes() []e.ZoneType {
	ctx := context.Background()
	zonetypes, execErr := r.query.GetZoneTypes(ctx)
	if execErr != nil {
		return []e.ZoneType{}
	}
	result := make([]e.ZoneType, 0, len(zonetypes))
	for _, row := range zonetypes {
		zonetype := e.ZoneType{
			Name: row,
		}
		result = append(result, zonetype)
	}
	return result
}
