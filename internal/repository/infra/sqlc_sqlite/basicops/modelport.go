
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
func (r BasicOpsSQLiteRepository) GetModelPorts() ([]e.ModelPort, error) {
	ctx := context.Background()
	models, execErr := r.query.GetModelPorts(ctx)
	if execErr != nil {
		return []e.ModelPort{}, execErr
	}

	result := make([]e.ModelPort, 0, len(models))
	for _, row := range models {
		model := e.ModelPort{
			ID:        int(row.ID),
			Name:      row.Name,
			Positionx: int(row.Positionx),
			Positiony: int(row.Positiony),
			Model:     row.Model.String,
			Brand:     row.Brand.String,
		}
		result = append(result, model)
	}
	return result, nil
}

// AddModel creates new model entry resolving brand/class names to IDs
func (r BasicOpsSQLiteRepository) AddModelPort(
	name string,
	posx string,
	posy string,
	modelName string,
) error {
	ctx := context.Background()

	// Get brand ID from name (required)
	modelID, err := r.query.GetModelId(ctx, modelName)
	if err != nil {
		return fmt.Errorf("model not found: %s", modelName)
	}
	iposx, err := strconv.Atoi(posx)
	if err != nil {
		return nil
	}
	iposy, err := strconv.Atoi(posy)
	if err != nil {
		return nil
	}

	modelStruct := d.AddModelPortParams{
		Name:      name,
		Positionx: int64(iposx),
		Positiony: int64(iposy),
		ModelID:   modelID,
	}

	if err := r.query.AddModelPort(ctx, modelStruct); err != nil {
		return fmt.Errorf("failed to create modelport: %v", err)
	}
	return nil
}
