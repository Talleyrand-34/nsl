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
func (r BasicOpsSQLiteRepository) AddZoneType(zoneName string) error {
	ctx := context.Background()
	execErr := r.query.AddZoneType(ctx, zoneName)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetZonetypes gets all the zone types available
func (r BasicOpsSQLiteRepository) GetZonetypes() ([]e.ZoneType, error) {
	ctx := context.Background()
	zonetypes, execErr := r.query.GetZoneTypes(ctx)
	if execErr != nil {
		return []e.ZoneType{}, execErr
	}
	result := make([]e.ZoneType, 0, len(zonetypes))
	for _, row := range zonetypes {
		zonetype := e.ZoneType{
			ID:   strconv.FormatInt(row.ID, 10),
			Name: row.LocationType,
		}
		result = append(result, zonetype)
	}
	return result, nil
}

// UpdateZoneType updates a zone type in the database by its ID
func (r BasicOpsSQLiteRepository) UpdateZoneType(zoneTypeId string, newZoneTypeName string) error {
	ctx := context.Background()

	// Convert string ID to int64
	id, err := strconv.ParseInt(zoneTypeId, 10, 64)
	if err != nil {
		return err
	}

	execErr := r.query.UpdateZoneType(ctx, d.UpdateZoneTypeParams{
		LocationType: newZoneTypeName,
		ID:           id,
	})
	if execErr != nil {
		return execErr
	}
	return nil
}

// DeleteZoneType deletes a zone type from the database by its name
func (r BasicOpsSQLiteRepository) DeleteZoneType(zonetype string) error {
	ctx := context.Background()
	if err := r.query.DeleteZoneType(ctx, zonetype); err != nil {
		return fmt.Errorf("DeleteZoneType failed: %w", err)
	}
	return nil
}
