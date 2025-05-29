/*
Copyright © 2025 Tecdesoft (rodrigo-gonzalez@tecdesoft.es, t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package basicops

import (
	"context"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
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

// UpdateBrand updates a brand in the database by its ID
func (r BasicOpsSQLiteRepository) UpdateBrand(brandId string, newBrandName string) error {
	ctx := context.Background()

	// Convert string ID to int64
	id, err := strconv.ParseInt(brandId, 10, 64)
	if err != nil {
		return err
	}

	execErr := r.query.UpdateBrand(ctx, d.UpdateBrandParams{
		Brand: newBrandName,
		ID:    id,
	})
	if execErr != nil {
		return execErr
	}
	return nil
}

// DeleteBrand deletes a brand from the database by its name
func (r BasicOpsSQLiteRepository) DeleteBrand(brand string) error {
	ctx := context.Background()
	execErr := r.query.DeleteBrand(ctx, brand)
	if execErr != nil {
		return execErr
	}
	return nil
}
