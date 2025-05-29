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
func (r BasicOpsSQLiteRepository) AddConnectionType(connectiontypes string) error {
	ctx := context.Background()
	execErr := r.query.AddConnectinType(ctx, connectiontypes)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetBrands gets all the brands available
func (r BasicOpsSQLiteRepository) GetConnectionTypes() ([]e.ConnectionType, error) {
	ctx := context.Background()
	brands, execErr := r.query.GetConnectionTypes(ctx)
	if execErr != nil {
		return []e.ConnectionType{}, execErr
	}

	result := make([]e.ConnectionType, 0, len(brands))
	for _, row := range brands {
		cts := e.ConnectionType{
			Name: row,
		}
		result = append(result, cts)
	}
	return result, nil
}

// UpdateConnectionType updates a connection type in the database by its ID
func (r BasicOpsSQLiteRepository) UpdateConnectionType(connectionTypeId string, newConnectionTypeName string) error {
	ctx := context.Background()

	// Convert string ID to int64
	id, err := strconv.ParseInt(connectionTypeId, 10, 64)
	if err != nil {
		return err
	}

	execErr := r.query.UpdateConnectionType(ctx, d.UpdateConnectionTypeParams{
		ConnectionType: newConnectionTypeName,
		ID:             id,
	})
	if execErr != nil {
		return execErr
	}
	return nil
}
