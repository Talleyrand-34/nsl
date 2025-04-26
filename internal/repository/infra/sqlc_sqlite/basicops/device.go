package basicops

import (
	"context"
	"fmt"

	_ "modernc.org/sqlite" // This imports is the sqlite driver needed to access the db

	e "nsl-graph/internal/repository/entities"
	d "nsl-graph/internal/repository/infra/sqlc_sqlite/internal_sqlc_sqlite"
)

func (r BasicOpsSQLiteRepository) AddDevice(
	label string,
	model string,
	zoneId string,
	zoneName string,
	proprietary string,
) error {
	ctx := context.Background()

	// Get proprietary ID
	spropid := r.getProprietaryID(ctx, proprietary)

	szoneid := r.getZoneID(ctx, zoneId, zoneName)

	smodelid, err := r.getModelID(ctx, model)
	if err != nil {
		return err
	}
	devicestruct := d.AddDeviceParams{
		Label:       label,
		ModelID:     smodelid,
		Proprietary: spropid,
		ZoneID:      szoneid,
	}

	if err := r.query.AddDevice(ctx, devicestruct); err != nil {
		return fmt.Errorf("AddDevice failed: %w", err)
	}
	return nil
}

// deviceclass

func (r BasicOpsSQLiteRepository) GetDevices() ([]e.Device, error) {
	ctx := context.Background()
	devices, execErr := r.query.GetDevices(ctx)
	if execErr != nil {
		return []e.Device{}, execErr
	}

	result := make([]e.Device, 0, len(devices))
	for _, row := range devices {

		zone := e.Device{
			ID:          int(row.ID), // sql.NullInt64 to int
			Name:        row.Label,
			Model:       nullStringToString(row.Model),
			Brand:       nullStringToString(row.Brand),
			ZoneName:    nullStringToString(row.Zonename),
			ZoneID:      int(row.Zoneid.Int64), // sql.NullInt64 to int
			ZoneFather:  nullStringToString(row.Zonefathername),
			Proprietary: nullStringToString(row.Proprietary),
		}

		result = append(result, zone)
	}
	return result, nil
}
