package sqlc_sqlite

import (
	"context"
	"fmt"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

func (r SQLiteRepository) AddZone(
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

func (r SQLiteRepository) GetZones() []e.Zone {
	ctx := context.Background()
	zones, execErr := r.query.GetZones(ctx)
	if execErr != nil {
		return []e.Zone{}
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
			Id:           int(row.ID),
			Name:         row.Name,
			Father:       father,
			FatherId:     int(row.Fatherid.Int64),
			LocationType: row.LocationType,
			Proprietary:  row.Proprietary,
		}
		result = append(result, zone)
	}
	return result
}
