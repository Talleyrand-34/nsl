
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
