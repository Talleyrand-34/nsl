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

// DeleteDevice deletes a device from the database by its integer ID
func (r BasicOpsSQLiteRepository) DeleteDevice(id string) error {
	ctx := context.Background()

	// Convert string ID to integer
	intID, err := strconv.Atoi(id)
	if err != nil {
		return fmt.Errorf("DeleteDevice: invalid id '%s': %w", id, err)
	}

	// Call the generated query method with the integer ID
	if err := r.query.DeleteDevice(ctx, int64(intID)); err != nil {
		return fmt.Errorf("DeleteDevice failed: %w", err)
	}
	return nil
}
