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

// deviceclass
func (r BasicOpsSQLiteRepository) AddProprietary(proprietary string) error {
	ctx := context.Background()
	execErr := r.query.AddProprietary(ctx, proprietary)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetProperties gets all the proprietaries available
func (r BasicOpsSQLiteRepository) GetProperties() ([]e.Proprietary, error) {
	ctx := context.Background()
	proprietaries, execErr := r.query.GetProprietaries(ctx)
	if execErr != nil {
		return []e.Proprietary{}, execErr
	}
	result := make([]e.Proprietary, 0, len(proprietaries))
	for _, row := range proprietaries {
		proprietary := e.Proprietary{
			ID:   row.ID,
			Name: row.Proprietary,
		}
		result = append(result, proprietary)
	}
	return result, nil
}

// UpdateProprietary updates a proprietary entry in the database by its ID
func (r BasicOpsSQLiteRepository) UpdateProprietary(proprietaryId string, newProprietaryName string) error {
	ctx := context.Background()

	// Convert string ID to int64
	id, err := strconv.ParseInt(proprietaryId, 10, 64)
	if err != nil {
		return err
	}

	execErr := r.query.UpdateProprietary(ctx, d.UpdateProprietaryParams{
		Proprietary: newProprietaryName,
		ID:          id,
	})
	if execErr != nil {
		return execErr
	}
	return nil
}

// DeleteProprietary deletes a proprietary entry from the database by its name
func (r BasicOpsSQLiteRepository) DeleteProprietary(proprietary string) error {
	ctx := context.Background()
	if err := r.query.DeleteProprietary(ctx, proprietary); err != nil {
		return fmt.Errorf("DeleteProprietary failed: %w", err)
	}
	return nil
}
