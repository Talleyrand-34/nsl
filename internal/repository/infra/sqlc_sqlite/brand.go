package sqlc_sqlite

import (
	"context"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db
)

// AddBrand adds a new brand to the database
func (r SQLiteRepository) AddBrand(brand string) error {
	ctx := context.Background()
	execErr := r.query.AddBrand(ctx, brand)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetBrands gets all the brands available
func (r SQLiteRepository) GetBrands() []string {
	ctx := context.Background()
	brands, execErr := r.query.GetBrands(ctx)
	if execErr != nil {
		return []string{}
	}
	return brands
}
