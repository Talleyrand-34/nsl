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

// AddDevicePort adds a new device port to the database
func (r BasicOpsSQLiteRepository) AddDevicePort(deviceid string, modelportid string, macAddress string) error {
	ctx := context.Background()
	sdeviceid, err := strconv.Atoi(deviceid)
	if err != nil {
		return nil
	}
	smodelportid, err := strconv.Atoi(modelportid)
	if err != nil {
		return nil
	}
	checkValidPort := d.CheckDevicePortValidParams{
		ID:   int64(smodelportid),
		ID_2: int64(sdeviceid),
	}
	_, err = r.query.CheckDevicePortValid(ctx, checkValidPort)
	if err != nil {
		return fmt.Errorf("Succesfully rejected non valid port: %v", err)
	}

	// Handle nullable mac_address
	var macAddressNullable sql.NullString
	if macAddress != "" {
		macAddressNullable = sql.NullString{String: macAddress, Valid: true}
	}

	devportstruct := d.AddDevicePortParams{
		DeviceID:    int64(sdeviceid),
		ModelPortID: int64(smodelportid),
		MacAddress:  macAddressNullable,
	}
	execErr := r.query.AddDevicePort(ctx, devportstruct)
	if execErr != nil {
		return execErr
	}
	return nil
}

// GetBrands gets all the brands available
func (r BasicOpsSQLiteRepository) GetDevicePorts() ([]e.DevicePort, error) {
	ctx := context.Background()
	devports, execErr := r.query.GetDevicePorts(ctx)
	if execErr != nil {
		return []e.DevicePort{}, execErr
	}
	result := make([]e.DevicePort, 0, len(devports))
	for _, row := range devports {
		macAddress := ""
		if row.MacAddress.Valid {
			macAddress = row.MacAddress.String
		}
		model := e.DevicePort{
			DeviceID:   strconv.FormatInt(row.DeviceID, 10),
			ModelID:    strconv.FormatInt(row.ModelPortID, 10),
			MacAddress: macAddress,
			DevLabel:   row.Label.String,
			PortName:   row.Name.String,
			Positionx:  int(row.Positionx.Int64),
			Positiony:  int(row.Positiony.Int64),
		}
		result = append(result, model)
	}
	return result, nil
}

// DeleteDevicePort deletes a device port from the database by device and model port IDs
func (r BasicOpsSQLiteRepository) DeleteDevicePort(deviceid string, modelportid string) error {
	ctx := context.Background()

	sdeviceid, err := strconv.Atoi(deviceid)
	if err != nil {
		return fmt.Errorf("DeleteDevicePort: invalid deviceid '%s': %w", deviceid, err)
	}
	smodelportid, err := strconv.Atoi(modelportid)
	if err != nil {
		return fmt.Errorf("DeleteDevicePort: invalid modelportid '%s': %w", modelportid, err)
	}
	devportstruct := d.DeleteDevicePortParams{
		DeviceID:    int64(sdeviceid),
		ModelPortID: int64(smodelportid),
	}

	// Assuming your generated query expects two int64 parameters
	if err := r.query.DeleteDevicePort(ctx, devportstruct); err != nil {
		return fmt.Errorf("DeleteDevicePort failed: %w", err)
	}
	return nil
}
