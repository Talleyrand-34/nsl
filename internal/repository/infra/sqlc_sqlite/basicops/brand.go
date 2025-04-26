package basicops

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
)

// AddBrand adds a new brand to the database
func (r BasicOpsSQLiteRepository) AddBrand(brand string) error {
	ctx := context.Background()
	execErr := r.query.AddBrand(ctx, brand)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetBrands gets all the brands available
func (r BasicOpsSQLiteRepository) GetBrands() ([]e.Brand, error) {
	ctx := context.Background()
	brands, execErr := r.query.GetBrands(ctx)
	if execErr != nil {
		return []e.Brand{}, execErr
	}

	result := make([]e.Brand, 0, len(brands))
	for _, row := range brands {
		brand := e.Brand{
			Name: row,
		}
		result = append(result, brand)
	}
	return result, nil
}
