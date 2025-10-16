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
	"database/sql"
	"fmt"
	"strconv"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

func (r BasicOpsSQLiteRepository) AddZone(
	name string,
	fatherid string,
	father string,
	proprietary string,
	zonename string,
) error {
	ctx := context.Background()

	// Get father ID (handles all cases)
	sfatherid := r.getFatherID(ctx, fatherid, father)

	// Get proprietary ID
	spropid := r.getProprietaryID(ctx, proprietary)

	// Get zone type ID
	szonetypeid := r.getZoneTypeID(ctx, zonename)

	// Prepare struct and insert into DB
	zonestruct := d.AddZoneParams{
		Name:         name,
		Father:       sfatherid,
		LocationType: szonetypeid,
		Proprietary:  spropid,
	}
	if err := r.query.AddZone(ctx, zonestruct); err != nil {
		return fmt.Errorf("AddZone failed: %w", err)
	}
	return nil
}

// deviceclass

func (r BasicOpsSQLiteRepository) GetZones() ([]e.Zone, error) {
	ctx := context.Background()
	zones, execErr := r.query.GetZones(ctx)
	if execErr != nil {
		return []e.Zone{}, nil
	}

	result := make([]e.Zone, 0, len(zones))
	for _, row := range zones {
		var father string
		if row.Father.Valid {
			father = row.Father.String
		} else {
			father = "" // or any default value you want for "no father"
		}

		zone := e.Zone{
			ID:           strconv.FormatInt(row.ID, 10),
			Name:         row.Name,
			Father:       father,
			FatherID:     strconv.FormatInt(row.Fatherid.Int64, 10),
			LocationType: row.LocationType,
			Proprietary:  row.Proprietary,
		}
		result = append(result, zone)
	}
	return result, nil
}

// UpdateZone updates a zone in the database by its ID
func (r BasicOpsSQLiteRepository) UpdateZone(
	zoneId string,
	newZoneName string,
	newFatherZoneId string,
	newZoneTypeId string,
	newProprietaryId string,
) error {
	ctx := context.Background()

	// Convert string ID to int64
	id, err := strconv.ParseInt(zoneId, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid zone ID '%s': %w", zoneId, err)
	}

	// Handle nullable father zone ID
	var fatherZoneId sql.NullInt64
	if newFatherZoneId != "" && newFatherZoneId != "0" {
		fatherId, err := strconv.ParseInt(newFatherZoneId, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid father zone ID '%s': %w", newFatherZoneId, err)
		}
		fatherZoneId = sql.NullInt64{Int64: fatherId, Valid: true}
	}

	// Handle nullable zone type ID
	var zoneTypeId sql.NullInt64
	if newZoneTypeId != "" && newZoneTypeId != "0" {
		ztId, err := strconv.ParseInt(newZoneTypeId, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid zone type ID '%s': %w", newZoneTypeId, err)
		}
		zoneTypeId = sql.NullInt64{Int64: ztId, Valid: true}
	}

	// Handle nullable proprietary ID
	var proprietaryId sql.NullInt64
	if newProprietaryId != "" && newProprietaryId != "0" {
		propId, err := strconv.ParseInt(newProprietaryId, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid proprietary ID '%s': %w", newProprietaryId, err)
		}
		proprietaryId = sql.NullInt64{Int64: propId, Valid: true}
	}

	execErr := r.query.UpdateZone(ctx, d.UpdateZoneParams{
		Name:         newZoneName,
		Father:       fatherZoneId,
		LocationType: zoneTypeId,
		Proprietary:  proprietaryId,
		ID:           id,
	})
	if execErr != nil {
		return fmt.Errorf("UpdateZone failed: %w", execErr)
	}
	return nil
}

// DeleteZone deletes a zone from the database by its integer ID
func (r BasicOpsSQLiteRepository) DeleteZone(id string) error {
	ctx := context.Background()

	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("DeleteZone: invalid id '%s': %w", id, err)
	}

	if err := r.query.DeleteZone(ctx, int64(intID)); err != nil {
		return fmt.Errorf("DeleteZone failed: %w", err)
	}
	return nil
}
