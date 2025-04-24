package basicops

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
)

// deviceclass
func (r BasicOpsSQLiteRepository) AddProprietary(proprietary string) error {
	ctx := context.Background()
	execErr := r.query.AddProprietary(ctx, proprietary)
	if execErr != nil {
		return execErr
	}
	return nil
}

// deviceclass
func (r BasicOpsSQLiteRepository) GetProperties() []e.Proprietary {
	ctx := context.Background()
	proprietaries, execErr := r.query.GetProprietaries(ctx)
	if execErr != nil {
		return []e.Proprietary{}
	}
	result := make([]e.Proprietary, 0, len(proprietaries))
	for _, row := range proprietaries {
		brand := e.Proprietary{
			Name: row,
		}
		result = append(result, brand)
	}
	return result
}
