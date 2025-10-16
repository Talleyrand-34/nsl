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
			ID:        strconv.FormatInt(row.ID, 10),
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

// UpdateModelPort updates a model port in the database by its ID
func (r BasicOpsSQLiteRepository) UpdateModelPort(
	modelPortId string,
	newPortName string,
	newPositionX string,
	newPositionY string,
	newModelId string,
) error {
	ctx := context.Background()

	// Convert string ID to int64
	id, err := strconv.ParseInt(modelPortId, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid model port ID '%s': %w", modelPortId, err)
	}

	// Convert position X to int64
	posX, err := strconv.ParseInt(newPositionX, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid position X '%s': %w", newPositionX, err)
	}

	// Convert position Y to int64
	posY, err := strconv.ParseInt(newPositionY, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid position Y '%s': %w", newPositionY, err)
	}

	// Convert model ID to int64
	modelId, err := strconv.ParseInt(newModelId, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid model ID '%s': %w", newModelId, err)
	}

	execErr := r.query.UpdateModelPort(ctx, d.UpdateModelPortParams{
		Name:      newPortName,
		Positionx: posX,
		Positiony: posY,
		ModelID:   modelId,
		ID:        id,
	})
	if execErr != nil {
		return fmt.Errorf("UpdateModelPort failed: %w", execErr)
	}
	return nil
}

// DeleteModelPort deletes a model port from the database by its integer ID
func (r BasicOpsSQLiteRepository) DeleteModelPort(id string) error {
	ctx := context.Background()

	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("DeleteModelPort: invalid id '%s': %w", id, err)
	}

	if err := r.query.DeleteModelPort(ctx, int64(intID)); err != nil {
		return fmt.Errorf("DeleteModelPort failed: %w", err)
	}
	return nil
}
