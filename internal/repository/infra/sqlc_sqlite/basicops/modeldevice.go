
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
	"fmt"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

// GetModels returns all models with resolved brand/class names
func (r BasicOpsSQLiteRepository) GetModels() ([]e.ModelDevice, error) {
	ctx := context.Background()
	models, execErr := r.query.GetModels(ctx)
	if execErr != nil {
		return []e.ModelDevice{}, execErr
	}

	result := make([]e.ModelDevice, 0, len(models))
	for _, row := range models {
		model := e.ModelDevice{
			ID:    strconv.FormatInt(row.ID, 10),
			Model: row.Model,
			Brand: row.Brand,
			Class: row.ClassName,
		}
		result = append(result, model)
	}
	return result, nil
}

// AddModel creates new model entry resolving brand/class names to IDs
func (r BasicOpsSQLiteRepository) AddModel(
	modelName string,
	brandName string,
	className string,
) error {
	ctx := context.Background()

	// Get brand ID from name (required)
	brandID, err := r.query.GetBrandId(ctx, brandName)
	if err != nil {
		return fmt.Errorf("brand not found: %s", brandName)
	}

	// Get class ID from name (required)
	classID, err := r.query.GetClassId(ctx, className)
	if err != nil {
		return fmt.Errorf("class not found: %s", className)
	}

	modelStruct := d.AddModelParams{
		Model:   modelName,
		Brand:   brandID,
		ClassID: classID,
	}

	if err := r.query.AddModel(ctx, modelStruct); err != nil {
		return fmt.Errorf("failed to create model")
	}
	return nil
}
